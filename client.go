package main

import (
	"fmt"
	"net"
)

func main() {
	fmt.Println("echo")
	echo()
	// for {
	// 	go echo()
	// 	time.Sleep(200 * time.Millisecond)
	// }
}

func echo() {
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer conn.Close()
	// scanner := bufio.NewScanner(conn)
	// for scanner.Scan() {
	// 	if len(scanner.Text()) == 0 {
	// 		break
	// 	}
	// }
	// scanner.Text()
	fmt.Println(conn, "GET / HTTP/1.0")
	req := "PONG\r\n"
	_, err = conn.Write([]byte(req))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("connection close")
	data := make([]byte, 512)
	n, err := conn.Read(data)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	res := string(data[:n])
	fmt.Println(res)
}
