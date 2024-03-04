package main

import (
	"fmt"
	"net"
	"os"
	"redis-go/internal/handler"
	"redis-go/internal/store"
)

func main() {
	fmt.Println("Logs from your program will appear here!")
	fmt.Println("Initialize key value store...")
	store := store.NewStore()
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			conn.Close()
		}

		fmt.Println("handle a connection:")
		handler := handler.NewHandler(conn, store)
		go handler.Handle()
	}
}
