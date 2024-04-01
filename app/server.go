package main

import (
	"bufio"
	"encoding/base64"
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

func (k Store) SetPx(key string, value string, exp int64) string {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	k.db[key] = value
	k.exp[key] = now + exp
	return "OK"
}

type ReplicaConfig struct {
	masterHost    string
	masterPort    string
	offset        int
	replicationId string
}

type Config struct {
	role    string
	replica *ReplicaConfig
	port    string
}

var config = Config{role: "master"}
var replicaIdLen = 40

var replicas []net.Conn

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

func connectToMaster(address string) net.Conn {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		panic(err.Error())
	}
	return conn
}

func sendCommand(conn net.Conn, msg []string) string {
	_, err := conn.Write([]byte(toRespArrays(msg)))
	if err != nil {
		fmt.Println(err.Error())
	}
	data := make([]byte, 1024)
	n, err := conn.Read(data)
	if err != nil {
		fmt.Println(err.Error())
	}
	res := string(data[:n])
	return res
}

func getRdbFile() []byte {
	contentsInBase64 := "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="
	contents, err := base64.StdEncoding.DecodeString(contentsInBase64)
	if err != nil {
		panic(err.Error())
	}
	return contents
}

func handshakeToMaster() {
	masterAddress := config.replica.masterHost + ":" + config.replica.masterPort
	conn := connectToMaster(masterAddress)
	sendCommand(conn, []string{"PING"})
	sendCommand(conn, []string{"REPLCONF", "listening-port", config.port})
	sendCommand(conn, []string{"REPLCONF", "capa", "psync2"})
	sendCommand(conn, []string{"PSYNC", "?", "-1"})
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
			config.role = "slave"
		}
		return nil
	})
	config.replica = &replicaConfig
	if config.role == "master" {
		config.replica.offset = 0
		config.replica.replicationId = generateRandomString(replicaIdLen)
		replicas = []net.Conn{}
	}
	flag.Parse()
	port := *portPtr
	config.port = strconv.Itoa(port)
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
			continue
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
						go func() {
							for _, replica := range replicas {
								sendCommand(replica, []string{command, args, value, opt, strconv.FormatInt(param, 10)})
							}
						}()
					} else {
						store.Set(args, value)
						conn.Write([]byte(toRespSimpleStrings("OK")))
						go func() {
							for _, replica := range replicas {
								sendCommand(replica, []string{command, args, value})
							}
						}()
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
				conn.Write([]byte(toRespBulkStrings(fmt.Sprintf("role:%s\r\nmaster_replid:%s\r\nmaster_repl_offset:%d\r\n", config.role, config.replica.replicationId, config.replica.offset))))
			case "REPLCONF":
				conn.Write([]byte(toRespSimpleStrings("OK")))
			case "PSYNC":
				conn.Write([]byte(toRespSimpleStrings(fmt.Sprintf("FULLRESYNC %s %d", config.replica.replicationId, config.replica.offset))))
				rdbFile := getRdbFile()
				conn.Write([]byte(toRdbResponse(rdbFile)))
				replicas = append(replicas, conn)
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

func toRdbResponse(rdbFile []byte) string {
	return fmt.Sprintf("$%d\r\n%s", len(rdbFile), string(rdbFile))
}
