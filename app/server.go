package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
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
			conn.Write([]byte(err.Error()))
			conn.Close()
		}
		simple_str := "+PONG\r\n"
		go handle(conn, simple_str)
	}
}

func handle(conn net.Conn, response string) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		fmt.Println(scanner.Err())
	}
	for scanner.Scan() {
		if len(scanner.Text()) == 0 {
			break
		}
	}
	req := scanner.Text()
	fmt.Println("accept a request:", req, " addr:", conn.RemoteAddr())
	conn.Write([]byte(response))
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
