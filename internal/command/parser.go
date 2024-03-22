package command

import (
	"errors"
	"strconv"
	"strings"
)

type CommandParser interface {
	Parse() error
}

func determineDataType(str string) (rune, error) {
	firstByte := []rune(str)[0]
	switch firstByte {
	case SimpleStringMarker:
		return SimpleStringMarker, nil
	case SimpleErrorMarker:
		return SimpleErrorMarker, nil
	case IntegerMarker:
		return IntegerMarker, nil
	case BulkStringMarker:
		return BulkStringMarker, nil
	case ArrayMarker:
		return ArrayMarker, nil
	default:
		return 0, errors.New("unknown RESP data type")
	}
}

// ArrayParser parse array
type ArrayParser struct {
	request string
	// Add a field to store the parsed result, if required
	Elements []interface{}
}

func NewArrayParser(request string) *ArrayParser {
	return &ArrayParser{
		request: request,
	}
}

// ArrayParser parse RESP array datatype request
func (p *ArrayParser) Parse() error {
	if !strings.HasPrefix(p.request, "*") {
		return errors.New("invalid array format")
	}

	// Split the request into lines
	lines := strings.Split(p.request, "\r\n")

	// Extract the number of elements
	numElements, err := strconv.Atoi(lines[0][1:])
	if err != nil {
		return err
	}

	if numElements < 0 {
		return errors.New("invalid number of elements in array")
	}

	p.Elements = make([]interface{}, numElements)

	// Parse each element
	elementIndex := 0
	for i := 1; i < len(lines); i++ {
		line := lines[i]

		// Skip empty lines
		if line == "" {
			continue
		}

		// Determine the type of the element
		dataType := line[0]
		switch dataType {
		case SimpleStringMarker:
			p.Elements[elementIndex] = line[1:] // Skip the '+' marker
		case SimpleErrorMarker:
			p.Elements[elementIndex] = line[1:] // Skip the '-' marker
		case IntegerMarker:
			integerValue, err := strconv.Atoi(line[1:]) // Skip the ':' marker
			if err != nil {
				return err
			}
			p.Elements[elementIndex] = integerValue

		case BulkStringMarker:
			// Extract the length of the bulk string
			_, err := strconv.Atoi(line[1:])
			if err != nil {
				return err
			}

			// Extract the bulk string
			if i+1 < len(lines) {
				p.Elements[elementIndex] = lines[i+1]
				i++ // Skip the next line
			} else {
				return errors.New("incomplete bulk string")
			}
		case ArrayMarker:
			// Parse nested array recursively
			nestedArray := &ArrayParser{request: strings.Join(lines[i:], "\r\n")}
			if err := nestedArray.Parse(); err != nil {
				return err
			}
			p.Elements[elementIndex] = nestedArray.Elements
			i += len(nestedArray.Elements) // Skip the lines parsed for the nested array
		default:
			return errors.New("unknown data type")
		}
		elementIndex++
	}

	if elementIndex != numElements {
		return errors.New("actual number of elements does not match the declared number")
	}

	return nil
}
