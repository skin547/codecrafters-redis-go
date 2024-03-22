package command

import (
	"errors"
	"strings"
)

const (
	SimpleStringMarker = '+'
	SimpleErrorMarker  = '-'
	IntegerMarker      = ':'
	BulkStringMarker   = '$'
	ArrayMarker        = '*'
)

func createParser(request string) (CommandParser, error) {
	// Split the input string into an array of command and arguments
	parts := strings.Split(strings.TrimSpace(request), "\r\n")
	firstByte := []rune(parts[0])[0]
	switch firstByte {
	// case SimpleStringMarker:
	// 	return SimpleStringMarker, nil
	// case SimpleErrorMarker:
	// 	return SimpleErrorMarker, nil
	// case IntegerMarker:
	// 	return IntegerMarker, nil
	// case BulkStringMarker:
	// 	return BulkStringMarker, nil
	case ArrayMarker:
		return NewArrayParser(request), nil
	default:
		return nil, errors.New("unknown RESP data type")
	}
}

func CreateCommand(str string) Command {
	// Split the input string into an array of command and arguments
	// parts := strings.Split(strings.TrimSpace(str), "\r\n")

	// // Check if the input string is empty or invalid
	// if len(parts) == 0 {
	// 	return NewSimpleErrorCommand("Invalid command")
	// }

	// // Extract command and arguments
	// fmt.Printf("parts: %v\n", parts)
	parser, err := createParser(str)

	if err != nil {
		return NewSimpleErrorCommand(err.Error())
	}

	err = parser.Parse()

	if err != nil {
		return NewSimpleErrorCommand(err.Error())
	}
	return nil
}
