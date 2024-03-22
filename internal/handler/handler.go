package handler

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"redis-go/internal/command"
	"redis-go/internal/store"
	"redis-go/internal/util"
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
	command := command.CreateCommand(str)
	response := command.Execute()
	h.conn.Write([]byte(response))
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
