package handler

import (
	"bytes"
	"net"
	"redis-go/internal/store"
	"testing"
	"time"
)

// TestHandlerRequest tests the handlerRequest function
func TestHandlerRequest(t *testing.T) {
	// Create a mock net.Conn for testing
	mockConn := &mockConn{}

	// Create a mock store for testing
	mockStore := store.NewStore()

	// Create a new Handler instance for testing
	handler := NewHandler(mockConn, mockStore)

	// Define test cases with input strings and expected responses
	testCases := []struct {
		inputString    string
		expectedOutput string
	}{
		{inputString: "*1\r\n$4\r\nPING\r\n", expectedOutput: "+PONG\r\n"},
		{inputString: "*2\r\n$4\r\nECHO\r\n$5\r\nHello\r\n", expectedOutput: "$5\r\nHello\r\n"},
		// Add more test cases here
	}

	// Run the test cases
	for _, tc := range testCases {
		// Reset the mockConn buffer
		mockConn.Reset()

		// Call the handlerRequest method with the input string
		handler.handlerRequest(tc.inputString)

		// Verify the output written to the connection matches the expected response
		actualOutput := mockConn.String()
		if actualOutput != tc.expectedOutput {
			t.Errorf("Input: %s, Expected: %s, Actual: %s", tc.inputString, tc.expectedOutput, actualOutput)
		}
	}
}

// Define a custom type that implements the net.Conn interface for testing
type mockConn struct {
	bytes.Buffer // Embed bytes.Buffer to capture the output
}

// Close implements net.Conn.
func (m *mockConn) Close() error {
	panic("unimplemented")
}

// LocalAddr implements net.Conn.
func (m *mockConn) LocalAddr() net.Addr {
	panic("unimplemented")
}

// Read implements net.Conn.
// Subtle: this method shadows the method (Buffer).Read of mockConn.Buffer.
func (m *mockConn) Read(b []byte) (n int, err error) {
	panic("unimplemented")
}

// RemoteAddr implements net.Conn.
func (m *mockConn) RemoteAddr() net.Addr {
	panic("unimplemented")
}

// SetDeadline implements net.Conn.
func (m *mockConn) SetDeadline(t time.Time) error {
	panic("unimplemented")
}

// SetReadDeadline implements net.Conn.
func (m *mockConn) SetReadDeadline(t time.Time) error {
	panic("unimplemented")
}

// SetWriteDeadline implements net.Conn.
func (m *mockConn) SetWriteDeadline(t time.Time) error {
	panic("unimplemented")
}

// Implement the Write method of the net.Conn interface for the mockConn type
func (m *mockConn) Write(b []byte) (int, error) {
	return m.Buffer.Write(b)
}

// Reset resets the buffer content
func (m *mockConn) Reset() {
	m.Buffer.Reset()
}
