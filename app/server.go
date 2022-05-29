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
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()
	for {
		reader := bufio.NewReader(conn)
		p := make([]byte, 512)
		n, err := reader.Read(p)
		if err == io.EOF {
			fmt.Println("Read finish")
			break
		}
		if err != nil {
			fmt.Println("Read failed")
			break
		}
		fmt.Println("accept a request, addr:", conn.RemoteAddr())
		str := string(p[:n])
		first := str[0:1]
		var arr []string
		if first == "*" {
			arr = strings.Split(str[1:], "\r\n")
			for index, element := range arr {
				fmt.Print(index, ":", element, ", ")
			}
		}
		if len(arr) == 4 && strings.EqualFold(arr[2], "PING") {
			command := arr[2]
			switch command {
			case "PING":
				conn.Write([]byte(toRespSimpleStrings("PONG")))
			case "ECHO":
				conn.Write([]byte(toRespBulkStrings(arr[4])))
			default:
				err := "not handle type"
				fmt.Println("err", err)
			}
		} else {
			conn.Write([]byte(toRespBulkStrings(arr[4])))
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

func toRespArray(str string) string {
	return terminated("+" + str)
}
