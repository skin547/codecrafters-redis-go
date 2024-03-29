package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

type Store struct {
	db  map[string]string
	exp map[string]int64
}

func NewStore() *Store {
	store := &Store{db: map[string]string{}, exp: map[string]int64{}}
	return store
}

func (k Store) Get(key string) (string, bool) {
	if exp, exist := k.exp[key]; exist {
		now := time.Now().UnixNano() / int64(time.Millisecond)
		if exp < now {
			delete(k.exp, key)
			delete(k.db, key)
			return "", false
		}
	}
	val, ok := k.db[key]
	return val, ok
}

func (k Store) Set(key string, value string) string {
	k.db[key] = value
	return "OK"
}

type ReplicaConfig struct {
	masterHost    string
	masterPort    string
	offset        int
	replicationId string
}

func (k Store) SetPx(key string, value string, exp int64) string {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	k.db[key] = value
	k.exp[key] = now + exp
	return "OK"
}

type Config struct {
	role    string
	replica *ReplicaConfig
}

var config = Config{role: "master"}
var replicaIdLen = 40

func generateRandomString(l int) string {
	charSet := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, l)
	for i := range s {
		s[i] = charSet[rand.Intn(len(charSet))]
	}
	return string(s)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func handshakeToMaster() {
	masterAddress := config.replica.masterHost + ":" + config.replica.masterPort
	fmt.Println("handshake to master at ", masterAddress)
	conn, err := net.Dial("tcp", masterAddress)
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()
	_, err = conn.Write([]byte(toRespArrays([]string{"PING"})))
	if err != nil {
		panic(err)
	}
	data := make([]byte, 512)
	n, err := conn.Read(data)
	if err != nil {
		panic(err.Error())
	}
	res := string(data[:n])
	fmt.Println("handshake response: ", res)
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	portPtr := flag.Int("port", 6379, "Port number")
	var replicaConfig ReplicaConfig
	flag.Func("replicaof", "Replica of <master_host> <master_port>", func(flagValue string) error {
		fmt.Println("flagValue" + flagValue)
		if flagValue == "" {
			return nil
		}
		replicaConfig.masterHost = flagValue
		if flag.NArg() != 0 {
			replicaConfig.masterPort = flag.Arg(0)
		}
		config.role = "slave"
		return nil
	})
	config.replica = &replicaConfig
	if config.role == "master" {
		config.replica.offset = 0
		config.replica.replicationId = generateRandomString(replicaIdLen)
	}
	flag.Parse()
	port := *portPtr
	address := fmt.Sprintf("0.0.0.0:%d", port)
	fmt.Println("Listening on " + address)

	fmt.Println("Replica of " + config.replica.masterHost + ":" + config.replica.masterPort)
	l, err := net.Listen("tcp", address)
	if config.role == "slave" {
		handshakeToMaster()
	}

	if err != nil {
		panic(fmt.Sprintf("Failed to bind to port %d", port))
	}
	defer l.Close()

	fmt.Println("Initialize key value store...")
	store := NewStore()
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
					with_opts := resp_arr_len >= 5
					if with_opts {
						opt := strings.ToUpper(arr[8])
						param, err := strconv.ParseInt(arr[10], 0, 64)
						if err != nil {
							conn.Write([]byte(toRespSimpleStrings("ERR wrong expire time")))
						}
						if opt == "PX" {
							store.SetPx(args, value, param)
							conn.Write([]byte(toRespSimpleStrings("OK")))
						}
					} else {
						store.Set(args, value)
						conn.Write([]byte(toRespSimpleStrings("OK")))
					}
					fmt.Println(store)
				}
			case "GET":
				if value, exist := store.Get(args); exist {
					conn.Write([]byte(toRespSimpleStrings(value)))
				} else {
					conn.Write([]byte(toRespErrorBulkStrings()))
				}
			case "INFO":
				conn.Write([]byte(toRespBulkStrings("role:" + config.role + "\r\n" + "master_replid:" + config.replica.replicationId + "\r\n" + "master_repl_offset:" + strconv.Itoa(config.replica.offset) + "\r\n")))
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

func toRespArrays(arr []string) string {
	res := fmt.Sprintf("*%d\r\n", len(arr))
	for _, element := range arr {
		res += toRespBulkStrings(element)
	}
	return res
}
