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
)

type RESP struct {
	Type RespType
	Data interface{}
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
	var lengthOfParsed int
	switch nextResp.Type {
	case SimpleString, Error:
		if str, ok := nextResp.Data.(string); ok {
			lengthOfParsed = len("+") + len(str) // CRLF
		} else {
			return nil, startIndex, errors.New("expected string data type")
		}
	case Integer:
		if _, ok := nextResp.Data.(int64); ok {
			lengthOfParsed = len(":") + len(fmt.Sprintf("%d", nextResp.Data.(int64)))
		} else {
			return nil, startIndex, errors.New("expected integer data type")
		}
	// TODO: fix dynamic type resp value by calculate end index when parsing
	case BulkString:
		lengthSpecifier := strconv.Itoa(len(nextResp.Data.(string)))
		if str, ok := nextResp.Data.(string); ok {
			lengthOfParsed = len("$") + len(lengthSpecifier) + len("\r\n") + len(str)
		} else {
			return nil, startIndex, errors.New("expected string data type")
		}
	case Array:
		if _, ok := nextResp.Data.([]*RESP); ok {
			lengthOfParsed = len("*") + len(fmt.Sprintf("%d", len(nextResp.Data.([]*RESP))))
		} else {
			return nil, startIndex, errors.New("expected array data type")
		}
	}

	nextIndex := startIndex + lengthOfParsed + len("\r\n")
	return nextResp, nextIndex, nil

}
