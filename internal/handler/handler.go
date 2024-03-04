package handler

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"redis-go/internal/store"
	"redis-go/internal/util"
	"strings"
)

// Handler handles an incoming connection
type Handler struct {
	conn  net.Conn
	store *store.Store
}

func NewHandler(conn net.Conn, store *store.Store) *Handler {
	return &Handler{
		conn:  conn,
		store: store,
	}
}

func (h *Handler) Handle() {
	defer h.conn.Close()

	fmt.Println("accept a request, addr:", h.conn.RemoteAddr())

	reader := bufio.NewReader(h.conn)

	for {
		p := make([]byte, 512)
		n, err := reader.Read(p)
		if err == io.EOF {
			fmt.Println("Read finished")
			break
		}
		if err != nil {
			fmt.Println("Read failed:", err)
		}

		str := string(p[:n])

		h.handlerRequest(str)
	}
}

func (h *Handler) handlerRequest(str string) {
	// Split the input string into an array of command and arguments
	parts := strings.Split(strings.TrimSpace(str), "\r\n")

	// Check if the input string is empty or invalid
	if len(parts) == 0 {
		h.conn.Write([]byte(util.ToRespErrorBulkStrings("Invalid command")))
		return
	}

	// Extract command and arguments
	command := strings.ToUpper(parts[0])
	var args []string
	if len(parts) > 1 {
		args = parts[1:]
	}

	// Execute the command based on its type
	switch command {
	case "PING":
		h.handlePingCommand(args)
	case "ECHO":
		h.handleEchoCommand(args)
	case "SET":
		h.handleSetCommand(args)
	case "GET":
		h.handleGetCommand(args)
	default:
		h.conn.Write([]byte(util.ToRespErrorBulkStrings("Unknown command")))
	}
}

func (h *Handler) handlePingCommand(args []string) {
	response := "+PONG\r\n"
	h.conn.Write([]byte(response))
}

func (h *Handler) handleEchoCommand(args []string) {
	if len(args) == 0 {
		h.conn.Write([]byte(util.ToRespErrorBulkStrings("No argument provided")))
		return
	}
	response := util.ToRespBulkStrings(args[0])
	h.conn.Write([]byte(response))
}

func (h *Handler) handleSetCommand(args []string) {
	if len(args) != 2 {
		h.conn.Write([]byte(util.ToRespErrorBulkStrings("Wrong number of arguments")))
		return
	}
	key := args[0]
	value := args[1]
	// Store the key-value pair in the store
	h.store.Set(key, value)
	response := "+OK\r\n"
	h.conn.Write([]byte(response))
}

func (h *Handler) handleGetCommand(args []string) {
	if len(args) != 1 {
		h.conn.Write([]byte(util.ToRespErrorBulkStrings("Wrong number of arguments")))
		return
	}
	key := args[0]
	// Retrieve the value from the store
	value, exist := h.store.Get(key)
	if exist {
		response := util.ToRespBulkStrings(value)
		h.conn.Write([]byte(response))
	} else {
		h.conn.Write([]byte(util.ToRespErrorBulkStrings("Key not found")))
	}
}
