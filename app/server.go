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
		go handle(conn)
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

func handle(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	str, err := reader.ReadString('\n')
	switch {
	case err == io.EOF:
		fmt.Println("Read finish")
		return
	case err != nil:
		fmt.Println("Read failed")
	}
	req := strings.Trim(str, "\r\n")
	fmt.Println(req)
	fmt.Println("accept a request:", req, " addr:", conn.RemoteAddr())
	if req == "PING" {
		conn.Write([]byte(toRespSimpleStrings("PONG")))
	} else {
		conn.Write([]byte(toRespBulkStrings(req)))
	}
	fmt.Println("write response")
	fmt.Println("connection close")
}

// func handleConnection(c net.Conn, response string) {
// 	defer c.Close()

// 	scanner := bufio.NewScanner(c)
// 	// Scan first line for the request
// 	if !scanner.Scan() {
// 		fmt.Println(scanner.Err())
// 	}
// 	req := scanner.Text()
// 	for scanner.Scan() {
// 		// Scan until an empty line is seen
// 		if len(scanner.Text()) == 0 {
// 			break
// 		}
// 	}
// 	fmt.Println("req:", req)
// 	if strings.HasPrefix(req, "GET") {
// 		rt := fmt.Sprintf("HTTP/1.1 200 Success\r\n")
// 		rt += fmt.Sprintf("Connection: Close\r\n")
// 		rt += fmt.Sprintf("Content-Type: text/html\r\n\r\n")
// 		rt += fmt.Sprintf("<html><body>Nothing here</body></html>\r\n")
// 		c.Write([]byte(rt))
// 	} else {
// 		rt := fmt.Sprintf("HTTP/1.1 %v Error Occurred\r\n\r\n", 501)
// 		c.Write([]byte(rt))
// 	}
// }
