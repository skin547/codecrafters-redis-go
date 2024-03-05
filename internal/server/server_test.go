package server

import (
	"redis-go/internal/command"
	"testing"
)

// MockCommandHandler is a mock implementation of the CommandHandler interface for testing purposes.
type MockCommandHandler struct {
	executed bool
}

// Execute is a mock implementation of the Execute method for testing purposes.
func (m *MockCommandHandler) Execute(args []string) string {
	m.executed = true
	return "OK"
}

// TestHandleCommand tests the HandleCommand method.
func TestHandleCommand(t *testing.T) {
	// Prepare test data
	commandRegistry := command.NewCommandRegistry()
	mockHandler := &MockCommandHandler{}
	commandRegistry.RegisterCommand("PING", mockHandler)
	server := NewServer(commandRegistry)

	// Test simple string command
	response := server.HandleCommand("PING")
	if response != "OK" {
		t.Errorf("Expected response 'OK', got '%s'", response)
	}
	if !mockHandler.executed {
		t.Error("Expected command handler to be executed")
	}

	// Test array command
	response = server.HandleCommand("*2\r\n$4\r\nPING\r\n")
	if response != "OK" {
		t.Errorf("Expected response 'OK', got '%s'", response)
	}
	if !mockHandler.executed {
		t.Error("Expected command handler to be executed")
	}

	// Test unknown command
	response = server.HandleCommand("INVALID")
	if response != "ERR unknown command" {
		t.Errorf("Expected response 'ERR unknown command', got '%s'", response)
	}
}
