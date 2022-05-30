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

type Store struct {
	db map[string]string
}

func NewStore() *Store {
	store := &Store{db: map[string]string{}}
	return store
}

func (k Store) Get(key string) (string, bool) {
	val, ok := k.db[key]
	return val, ok
}

func (k Store) Set(key string, value string) string {
	k.db[key] = value
	return "OK"
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	fmt.Println("Initialize key value store...")
	store := NewStore()
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
		go handle(conn, store)
	}
}

func handle(conn net.Conn, store *Store) {
	fmt.Println("accept a request, addr:", conn.RemoteAddr())
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
		str := string(p[:n])
		first := str[0:1]
		var arr []string
		resp_arr_len, err := strconv.ParseInt(str[1:2], 10, 0)
		if err != nil {
			fmt.Println("err", err)
		}
		if first == "*" {
			arr = strings.Split(str[1:], "\r\n")
			for index, element := range arr {
				fmt.Print(index, ":", element, ", ")
			}
			fmt.Println()
		}
		command := strings.ToUpper(arr[2])
		fmt.Println("arr:", arr, "command:", command, "resp_arr_len:", resp_arr_len)
		with_args := resp_arr_len >= 2
		fmt.Println("withArgs", with_args)
		if with_args {
			args := arr[4]
			switch command {
			case "PING":
				conn.Write([]byte(toRespBulkStrings(args)))
			case "ECHO":
				conn.Write([]byte(toRespBulkStrings(args)))
			case "SET":
				if resp_arr_len < 3 {
					conn.Write([]byte(toRespSimpleStrings("ERR wrong number of arguments for command")))
				} else {
					value := arr[6]
					store.Set(args, value)
					conn.Write([]byte(toRespSimpleStrings("OK")))
					fmt.Println(store)
				}
			case "GET":
				value, exist := store.Get(args)
				fmt.Println("value", value)
				if exist {
					conn.Write([]byte(toRespSimpleStrings(value)))
				} else {
					conn.Write([]byte(toRespErrorBulkStrings()))
				}
			default:
				conn.Write([]byte(toRespSimpleStrings("ERR wrong command " + command)))
			}
		} else {
			switch command {
			case "PING":
				conn.Write([]byte(toRespSimpleStrings("PONG")))
			case "ECHO":
				conn.Write([]byte(toRespSimpleStrings("ERR wrong number of arguments for command")))
			default:
				conn.Write([]byte(toRespSimpleStrings("ERR wrong command " + command)))
			}
		}
	}
}

func toRespSimpleStrings(str string) string {
	return terminated("+" + str)
}

func terminated(str string) string {
	return str + "\r\n"
}

func toRespErrorBulkStrings() string {
	return terminated("$-1")
}

func toRespBulkStrings(str string) string {
	if str == "" {
		return terminated("$0" + terminated(""))
	}
	length := len(str)
	lenStr := strconv.Itoa(length)
	res := terminated("$" + terminated(lenStr) + str)
	fmt.Println("len:", lenStr, " res:", res)
	return res
}

func toRespArray(str string) string {
	return terminated("+" + str)
}
