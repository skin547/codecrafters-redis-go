package command

import (
	"reflect"
	"testing"
)

// Your existing constants and determineDataType function would be included here

func TestParseArray(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    []interface{}
		wantErr bool
	}{
		{
			name:    "simple array",
			data:    "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n",
			want:    []interface{}{"foo", "bar"},
			wantErr: false,
		},
		{
			name:    "array with different types",
			data:    "*3\r\n:1\r\n$3\r\nbaz\r\n+OK\r\n",
			want:    []interface{}{1, "baz", "OK"},
			wantErr: false,
		},
		{
			name:    "empty array",
			data:    "*0\r\n",
			want:    []interface{}{},
			wantErr: false,
		},
		{
			name:    "invalid array format",
			data:    "*2\r\n$3\r\nfoo\r\nbar\r\n",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewArrayParser(tt.data)
			err := parser.Parse()
			got := parser.Elements
			if (err != nil) != tt.wantErr {
				t.Errorf("parseArray() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("test case fail: %v", tt.name)
				t.Errorf("parseArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

// parseArray, parseSimpleString, parseError, parseInteger, parseBulkString would be defined here.
