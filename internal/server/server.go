package server

import (
	"fmt"
	"redis-go/internal/command"
	"redis-go/internal/util"
	"strings"
)

type Server struct {
	commandRegistry *command.CommandRegistry
}

func NewServer(commandRegistry *command.CommandRegistry) *Server {
	return &Server{
		commandRegistry: commandRegistry,
	}
}

func (s *Server) HandleCommand(commandStr string) string {
	parts := strings.Fields(commandStr)
	if len(parts) == 0 {
		return util.ToRespErrorBulkStrings("Invalid command")
	}

	// not well-implemented, see reference: https://redis.io/docs/reference/protocol-spec
	if strings.HasPrefix(parts[0], "*") {
		if len(parts) < 2 {
			return util.ToRespErrorBulkStrings("Invalid command")
		}
		commandName := strings.ToUpper(parts[2])
		handler := s.commandRegistry.Commands[commandName]
		if handler == nil {
			return "ERR unknown command"
		}
		args := parts[2:]
		return handler.Execute(args)
	}

	// Single-bulk command
	fmt.Println("single-bulk command")

	commandName := strings.ToUpper(parts[0])
	handler := s.commandRegistry.Commands[commandName]
	if handler == nil {
		return "ERR unknown command"
	}
	args := parts[1:]
	return handler.Execute(args)
}
