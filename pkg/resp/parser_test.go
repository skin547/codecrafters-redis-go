package resp

import (
	"reflect"
	"testing"
)

func TestParseSimpleString(t *testing.T) {
	input := "+OK\r\n"
	expected := &RESP{
		Type: SimpleString,
		Data: "OK",
	}
	actual, _, err := ParseRESP(input)
	if err != nil {
		t.Errorf("Error parsing simple string: %v", err)
	}
	if actual.Type != expected.Type {
		t.Errorf("Expected type %v, got %v", expected.Type, actual.Type)
	}
	if actual.Data != expected.Data {
		t.Errorf("Expected data %v, got %v", expected.Data, actual.Data)
	}
}

func TestParseError(t *testing.T) {
	input := "-Error\r\n"
	expected := &RESP{
		Type: Error,
		Data: "Error",
	}
	actual, _, err := ParseRESP(input)
	if err != nil {
		t.Errorf("Error parsing error: %v", err)
	}
	if actual.Type != expected.Type {
		t.Errorf("Expected type %v, got %v", expected.Type, actual.Type)
	}
	if actual.Data != expected.Data {
		t.Errorf("Expected data %v, got %v", expected.Data, actual.Data)
	}
}

func TestParseInteger(t *testing.T) {
	input := ":123\r\n"
	expected := &RESP{
		Type: Integer,
		Data: int64(123),
	}
	actual, _, err := ParseRESP(input)
	if err != nil {
		t.Errorf("Error parsing integer: %v", err)
	}
	if actual.Type != expected.Type {
		t.Errorf("Expected type %v, got %v", expected.Type, actual.Type)
	}
	if actual.Data != expected.Data {
		t.Errorf("Expected data %s, got %s", expected.Data, actual.Data)
	}
}

func TestParseBulkString(t *testing.T) {
	input := "$6\r\nfoobar\r\n"
	expected := &RESP{
		Type: BulkString,
		Data: "foobar",
	}
	actual, _, err := ParseRESP(input)
	if err != nil {
		t.Errorf("Error parsing bulk string: %v", err)
	}
	if actual.Type != expected.Type {
		t.Errorf("Expected type %v, got %v", expected.Type, actual.Type)
	}
	if actual.Data != expected.Data {
		t.Errorf("Expected data %v, got %v", expected.Data, actual.Data)
	}
}

func TestParseInvalid(t *testing.T) {
	input := "foobar\r\n"
	_, _, err := ParseRESP(input)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestParseInvalidBulkString(t *testing.T) {
	input := "$6\r\nfoobar"
	_, _, err := ParseRESP(input)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestParseInvalidInteger(t *testing.T) {
	input := ":123"
	_, _, err := ParseRESP(input)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestParseInvalidSimpleString(t *testing.T) {
	input := "+OK"
	_, _, err := ParseRESP(input)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestParseInvalidError(t *testing.T) {
	input := "-Error"
	_, _, err := ParseRESP(input)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestParseInvalidType(t *testing.T) {
	input := "foobar"
	_, _, err := ParseRESP(input)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestParseArrayOfBulkStrings(t *testing.T) {
	input := "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"
	expected := &RESP{
		Type: Array,
		Data: []*RESP{
			{Type: BulkString, Data: "foo"},
			{Type: BulkString, Data: "bar"},
		},
	}
	result, _, err := parseArray(input)
	if err != nil {
		t.Errorf("parseArray returned an error: %v", err)
	} else {
		for _, element := range result.Data.([]*RESP) {
			t.Logf("element: %+v", element) // print each element of result
		}
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("parseArray returned %+v, expected %+v", result, expected)
	}
}

func TestParseArrayOfSimpleString(t *testing.T) {
	input := "*2\r\n+foo\r\n+bar\r\n"
	expected := &RESP{
		Type: Array,
		Data: []*RESP{
			{Type: SimpleString, Data: "foo"},
			{Type: SimpleString, Data: "bar"},
		},
	}
	result, _, err := parseArray(input)
	// print each element of result
	for _, element := range result.Data.([]*RESP) {
		t.Logf("element: %+v", element) // print each element of result
	}
	if err != nil {
		t.Errorf("parseArray returned an error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("parseArray returned %+v, expected %+v", result, expected)
	}
}

func TestParseArrayWithIntegers(t *testing.T) {
	// Input representing an array of two integers
	input := "*2\r\n:123\r\n:456\r\n"
	expected := &RESP{
		Type: Array,
		Data: []*RESP{
			{Type: Integer, Data: int64(123)},
			{Type: Integer, Data: int64(456)},
		},
	}

	// Call the ParseRESP function or the specific array parsing function
	got, _, err := ParseRESP(input)
	if err != nil {
		t.Fatalf("ParseRESP failed with error: %v", err)
	}

	// Compare the parsed output to the expected output
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ParseRESP returned incorrect result, got %+v, want %+v", got, expected)
	}
}

func TestParseInvalidArray(t *testing.T) {
	input := "*3\r\n+OK\r\n+OK\r\n"
	_, _, err := ParseRESP(input)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestParseInvalidTypeArray(t *testing.T) {
	input := "*3"
	_, _, err := ParseRESP(input)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestParseInvalidTypeBulkString(t *testing.T) {
	input := "$6"
	_, _, err := ParseRESP(input)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestParseArrayWithIntegerAndString(t *testing.T) {
	input := "*2\r\n:100\r\n+OK\r\n"
	expected := &RESP{
		Type: Array,
		Data: []*RESP{
			{Type: Integer, Data: int64(100)},
			{Type: SimpleString, Data: "OK"},
		},
	}

	got, _, err := ParseRESP(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if got.Type != expected.Type {
		t.Fatalf("Expected type %v, got %v", expected.Type, got.Type)
	}

	// Assert and compare each element in the data slice
	if len(got.Data.([]*RESP)) != len(expected.Data.([]*RESP)) {
		t.Fatalf("Expected data length %d, got %d", len(expected.Data.([]*RESP)), len(got.Data.([]*RESP)))
	}

	for i, expResp := range expected.Data.([]*RESP) {
		switch expResp.Data.(type) {
		case int64:
			if gotData, ok := got.Data.([]*RESP)[i].Data.(int64); ok {
				if gotData != expResp.Data.(int64) {
					t.Errorf("Expected integer data %d, got %d", expResp.Data.(int64), gotData)
				}
			} else {
				t.Errorf("Expected data type int64, got %T", got.Data.([]*RESP)[i].Data)
			}
		case string:
			if gotData, ok := got.Data.([]*RESP)[i].Data.(string); ok {
				if gotData != expResp.Data.(string) {
					t.Errorf("Expected string data %s, got %s", expResp.Data.(string), gotData)
				}
			} else {
				t.Errorf("Expected data type string, got %T", got.Data.([]*RESP)[i].Data)
			}
		default:
			t.Errorf("Unhandled type in expected data")
		}
	}
}

func TestParseArrayWithBulkStringAndError(t *testing.T) {
	input := "*2\r\n$10\r\nfooooooooo\r\n-Error\r\n"
	expected := &RESP{
		Type: Array,
		Data: []*RESP{
			{Type: BulkString, Data: "fooooooooo"},
			{Type: Error, Data: "Error"},
		},
	}

	got, _, err := ParseRESP(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Got = %+s, want %+s", got.Data, expected.Data)
	}
}

func TestParseNestedArray(t *testing.T) {
	input := "*2\r\n*1\r\n+OK\r\n:123\r\n"
	expected := &RESP{
		Type: Array,
		Data: []*RESP{
			{
				Type: Array,
				Data: []*RESP{
					{Type: SimpleString, Data: "OK"},
				},
			},
			{Type: Integer, Data: int64(123)},
		},
	}

	got, _, err := ParseRESP(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Got = %+v, want %+v", got, expected)
	}
}

func TestIncompleteArrayElement(t *testing.T) {
	input := "*1\r\n$3\r\nfo"
	_, _, err := ParseRESP(input)
	if err == nil {
		t.Error("Expected error, got none")
	}
}

func TestIncorrectTypeInArray(t *testing.T) {
	input := "*1\r\n?3\r\nfoo\r\n"
	_, _, err := ParseRESP(input)
	if err == nil {
		t.Error("Expected error for incorrect type, got none")
	}
}

func TestSerialize(t *testing.T) {
	tests := []struct {
		name     string
		resp     *RESP
		expected string
	}{
		{
			name: "SimpleString",
			resp: &RESP{
				Type: SimpleString,
				Data: "OK",
			},
			expected: "+OK\r\n",
		},
		{
			name: "Error",
			resp: &RESP{
				Type: Error,
				Data: "Error",
			},
			expected: "-Error\r\n",
		},
		{
			name: "Integer",
			resp: &RESP{
				Type: Integer,
				Data: int64(123),
			},
			expected: ":123\r\n",
		},
		{
			name: "BulkString",
			resp: &RESP{
				Type: BulkString,
				Data: "foo",
			},
			expected: "$3\r\nfoo\r\n",
		},
		{
			name: "Array",
			resp: &RESP{
				Type: Array,
				Data: []*RESP{
					{Type: BulkString, Data: "SET"},
					{Type: BulkString, Data: "foo"},
					{Type: BulkString, Data: "123"},
				},
			},
			expected: "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\n123\r\n",
		},
		{
			name: "Nested Array",
			resp: &RESP{
				Type: Array,
				Data: []*RESP{
					{
						Type: Array,
						Data: []*RESP{
							{Type: SimpleString, Data: "OK"},
						},
					},
					{Type: Integer, Data: int64(123)},
				},
			},
			expected: "*2\r\n*1\r\n+OK\r\n:123\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.resp.Serialize()
			if actual != tt.expected {
				t.Errorf("Serialize() = %q, expected %q", actual, tt.expected)
			}
		})
	}
}
