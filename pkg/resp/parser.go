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

func ParseRESP(input string) (*RESP, error) {
	if input == "" {
		return nil, errors.New("empty input")
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
		return nil, errors.New("unknown RESP type")
	}
}

func parseSimpleString(input string) (*RESP, error) {
	end := strings.Index(input, "\r\n")
	if end == -1 {
		return nil, errors.New("invalid simple string: no CRLF")
	}
	return &RESP{
		Type: SimpleString,
		Data: input[1:end],
	}, nil
}

func parseError(input string) (*RESP, error) {
	end := strings.Index(input, "\r\n")
	if end == -1 {
		return nil, errors.New("invalid error: no CRLF")
	}
	return &RESP{
		Type: Error,
		Data: input[1:end],
	}, nil
}

func parseInteger(input string) (*RESP, error) {
	end := strings.Index(input, "\r\n")
	if end == -1 {
		return nil, errors.New("invalid integer: no CRLF")
	}
	val, err := strconv.ParseInt(input[1:end], 10, 64)
	if err != nil {
		return nil, err
	}
	return &RESP{
		Type: Integer,
		Data: val,
	}, nil
}

func parseBulkString(input string) (*RESP, error) {
	end := strings.Index(input, "\r\n")
	if end == -1 {
		return nil, errors.New("invalid bulk string: no CRLF")
	}
	length, err := strconv.ParseInt(input[1:end], 10, 64)
	if err != nil {
		return nil, err
	}
	if length == -1 {
		return &RESP{Type: BulkString, Data: nil}, nil // Nil bulk string
	}
	startIndex := end + len("\r\n")
	if int64(len(input)) < int64(startIndex)+length+int64(len("\r\n")) {
		return nil, errors.New("invalid bulk string: data too short")
	}
	data := input[startIndex : startIndex+int(length)]
	return &RESP{
		Type: BulkString,
		Data: data,
	}, nil
}

func parseArray(input string) (*RESP, error) {
	arrHeaderEnd := strings.Index(input, "\r\n")
	if arrHeaderEnd == -1 {
		return nil, errors.New("invalid array: no CRLF")
	}
	arrayLength, err := strconv.ParseInt(input[1:arrHeaderEnd], 10, 64)
	if err != nil {
		return nil, errors.New("invalid array length")
	}
	elements := make([]*RESP, 0, arrayLength)
	currentIndex := arrHeaderEnd + len("\r\n") // Start right after the init CRLF

	for i := int64(0); i < arrayLength; i++ {
		if currentIndex >= len(input) {
			return nil, errors.New("incomplete input data")
		}
		nextResp, nextIndex, err := parseNextElement(input, currentIndex)
		if err != nil {
			return nil, err
		}
		elements = append(elements, nextResp)
		currentIndex = nextIndex // Update currentIndex to the end of the last parsed element
	}

	if currentIndex < len(input) {
		remainResp, err := parseArray(input[currentIndex:])
		if err != nil {
			return nil, err
		}
		elements = append(elements, remainResp.Data.([]*RESP)...)
	}

	return &RESP{
		Type: Array,
		Data: elements,
	}, nil
}

// parseNextElement finds and parses the next RESP element in the input string
func parseNextElement(input string, startIndex int) (*RESP, int, error) {
	if startIndex >= len(input) {
		return nil, 0, errors.New("out of bounds when parsing next element")
	}
	nextResp, err := ParseRESP(input[startIndex:])
	if err != nil {
		return nil, startIndex, err
	}
	lengthOfParsed, err := calculateContentLength(nextResp)
	if err != nil {
		return nil, startIndex, err
	}

	nextIndex := startIndex + lengthOfParsed + len("\r\n")
	return nextResp, nextIndex, nil

}

func calculateContentLength(nextResp *RESP) (int, error) {
	var lengthOfParsed = 0
	switch nextResp.Type {
	case SimpleString, Error:
		if str, ok := nextResp.Data.(string); ok {
			lengthOfParsed = len("+") + len(str)
		} else {
			return 0, errors.New("expected string data type")
		}
	case Integer:
		if _, ok := nextResp.Data.(int64); ok {
			lengthOfParsed = len(":") + len(fmt.Sprintf("%d", nextResp.Data.(int64)))
		} else {
			return 0, errors.New("expected integer data type")
		}
	case BulkString:
		lengthSpecifier := strconv.Itoa(len(nextResp.Data.(string)))
		if str, ok := nextResp.Data.(string); ok {
			lengthOfParsed = len("$") + len(lengthSpecifier) + len("\r\n") + len(str)
		} else {
			return 0, errors.New("expected string data type")
		}
	case Array:
		if arr, ok := nextResp.Data.([]*RESP); ok {
			lengthOfParsed = len("*") + len(fmt.Sprintf("%d", cap(arr)))
			for _, element := range arr {
				contentLength, err := calculateContentLength(element)
				if err != nil {
					return 0, err
				}
				lengthOfParsed += contentLength + len("\r\n")
			}
		} else {
			return 0, errors.New("expected array data type")
		}
	}
	return lengthOfParsed, nil
}
