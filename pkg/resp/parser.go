package resp

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type RespType int

const (
	SimpleString RespType = iota
	Error
	Integer
	BulkString
	Array
	RDB
	NullBulkString
)

type RESP struct {
	Type RespType
	Data interface{}
}

func (r *RESP) Serialize() string {
	switch r.Type {
	case SimpleString:
		return fmt.Sprintf("+%s\r\n", r.Data)
	case Error:
		return fmt.Sprintf("-%s\r\n", r.Data)
	case Integer:
		return fmt.Sprintf(":%d\r\n", r.Data)
	case BulkString:
		return fmt.Sprintf("$%d\r\n%s\r\n", len(r.Data.(string)), r.Data)
	case Array:
		res := fmt.Sprintf("*%d\r\n", len(r.Data.([]*RESP)))
		for _, element := range r.Data.([]*RESP) {
			res += element.Serialize()
		}
		return res
	case RDB:
		return fmt.Sprintf("$%d\r\n%s", len(r.Data.([]byte)), string(r.Data.([]byte)))
	case NullBulkString:
		return "$-1\r\n"
	}
	panic("unknown RESP type")
}

type RESPParser struct {
	rawRequest   string
	CurrentIndex int
}

func NewRESPParser(input string) *RESPParser {
	return &RESPParser{
		rawRequest:   input,
		CurrentIndex: 0,
	}
}

func (p *RESPParser) HasNext() bool {
	hasNext := p.CurrentIndex < len(p.rawRequest)
	if hasNext {
		// print left data
		fmt.Printf("left data: %s\n", p.rawRequest[p.CurrentIndex:])
	}
	fmt.Printf("hasNext: %v\n", hasNext)
	return hasNext
}

func (p *RESPParser) ParseNext() (*RESP, error) {
	if p.CurrentIndex >= len(p.rawRequest) {
		return nil, errors.New("no more data to parse")
	}
	resp, nextIndex, err := ParseRESP(p.rawRequest[p.CurrentIndex:])
	if err != nil {
		return nil, err
	}
	p.CurrentIndex += nextIndex
	return resp, nil
}

func ParseRESP(input string) (*RESP, int, error) {
	if input == "" {
		return nil, 0, errors.New("empty input")
	}
	switch input[0] {
	case '+':
		return parseSimpleString(input)
	case '-':
		return parseError(input)
	case ':':
		return parseInteger(input)
	case '$':
		return parseBulkString(input)
	case '*':
		return parseArray(input)
	default:
		return nil, 0, fmt.Errorf("unknown RESP type: %s", string(input))
	}
}

func parseSimpleString(input string) (*RESP, int, error) {
	end := strings.Index(input, "\r\n")
	if end == -1 {
		return nil, 0, errors.New("invalid simple string: no CRLF")
	}
	return &RESP{
		Type: SimpleString,
		Data: input[1:end],
	}, end + len("\r\n"), nil
}

func parseError(input string) (*RESP, int, error) {
	end := strings.Index(input, "\r\n")
	if end == -1 {
		return nil, 0, errors.New("invalid error: no CRLF")
	}
	return &RESP{
		Type: Error,
		Data: input[1:end],
	}, end + len("\r\n"), nil
}

func parseInteger(input string) (*RESP, int, error) {
	end := strings.Index(input, "\r\n")
	if end == -1 {
		return nil, 0, errors.New("invalid integer: no CRLF")
	}
	val, err := strconv.ParseInt(input[1:end], 10, 64)
	if err != nil {
		return nil, 0, err
	}
	return &RESP{
		Type: Integer,
		Data: val,
	}, end + len("\r\n"), nil
}

func parseBulkString(input string) (*RESP, int, error) {
	end := strings.Index(input, "\r\n")
	if end == -1 {
		return nil, 0, errors.New("invalid bulk string: no CRLF")
	}
	length, err := strconv.ParseInt(input[1:end], 10, 64)
	if err != nil {
		return nil, 0, err
	}
	if length == -1 {
		return &RESP{Type: BulkString, Data: nil}, 0, nil // Nil bulk string
	}
	startIndex := end + len("\r\n")
	if int64(len(input)) < int64(startIndex)+length+int64(len("\r\n")) {
		return nil, 0, errors.New("invalid bulk string: data too short")
	}
	data := input[startIndex : startIndex+int(length)]
	return &RESP{
		Type: BulkString,
		Data: data,
	}, startIndex + int(length) + len("\r\n"), nil
}

func parseArray(input string) (*RESP, int, error) {
	arrHeaderEnd := strings.Index(input, "\r\n")
	if arrHeaderEnd == -1 {
		return nil, 0, errors.New("invalid array: no CRLF")
	}
	arrayLength, err := strconv.ParseInt(input[1:arrHeaderEnd], 10, 64)
	if err != nil {
		return nil, 0, errors.New("invalid array length")
	}
	elements := make([]*RESP, 0, arrayLength)
	currentIndex := arrHeaderEnd + len("\r\n") // Start right after the init CRLF

	for i := int64(0); i < arrayLength; i++ {
		if currentIndex >= len(input) {
			return nil, 0, errors.New("incomplete input data")
		}
		// print the next index
		nextResp, nextIndex, err := parseNextElement(input, currentIndex)
		if err != nil {
			return nil, 0, err
		}
		elements = append(elements, nextResp)
		currentIndex += nextIndex // Update currentIndex to the end of the last parsed element
	}

	return &RESP{
		Type: Array,
		Data: elements,
	}, currentIndex, nil
}

// parseNextElement finds and parses the next RESP element in the input string
func parseNextElement(input string, startIndex int) (*RESP, int, error) {
	if startIndex >= len(input) {
		return nil, 0, errors.New("out of bounds when parsing next element")
	}
	nextResp, nextIndex, err := ParseRESP(input[startIndex:])
	if err != nil {
		return nil, startIndex, err
	}

	return nextResp, nextIndex, nil
}
