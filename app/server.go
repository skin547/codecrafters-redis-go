package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

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
		handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()
	for {
		reader := bufio.NewReader(conn)
		str, err := reader.ReadString('\n')
		if err == io.EOF {
			fmt.Println("Read finish")
			break
		}
		if err != nil {
			fmt.Println("Read failed")
			break
		}
		req := strings.Trim(str, "\r\n")
		fmt.Println("accept a request:", req, " addr:", conn.RemoteAddr())
		if req == "PING" {
			conn.Write([]byte(toRespSimpleStrings("PONG")))
		} else if req == "*2" {
			conn.Write([]byte(toRespSimpleStrings("PONG")))
		} else {
			conn.Write([]byte(toRespBulkStrings(req)))
		}
	}
}

func toRespSimpleStrings(str string) string {
	return terminated("+" + str)
}

func terminated(str string) string {
	return str + "\r\n"
}

func toRespBulkStrings(str string) string {
	fmt.Println("toRespBulkStrings")
	if str == "" {
		return terminated("$0" + terminated(""))
	}
	length := len(str)
	lenStr := strconv.Itoa(length)
	fmt.Println(lenStr)
	res := terminated("$" + terminated(lenStr) + str)
	fmt.Println(res)
	return res
}
