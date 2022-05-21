package main

import (
	"fmt"
	"net"
)

func main() {
	fmt.Println("echo")
	echo("PING")
	echo("HELLO WORLD")
	// for {
	// 	go echo()
	// 	time.Sleep(200 * time.Millisecond)
	// }
}

func echo(command string) {
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer conn.Close()
	fmt.Println(conn, "GET / HTTP/1.0")
	_, err = conn.Write([]byte(command))
	if err != nil {
		fmt.Println(err)
		return
	}
	data := make([]byte, 512)
	n, err := conn.Read(data)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	res := string(data[:n])
	fmt.Println("response: ", res)
	fmt.Println("connection close")
}
