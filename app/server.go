package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"redis-go/internal/store"
	"redis-go/pkg/resp"
	Resp "redis-go/pkg/resp"
	"strconv"
	"strings"
	"time"
)

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

type Replica struct {
	conn net.Conn
}

var replicas []Replica

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

func sendCommand(conn net.Conn, req *Resp.RESP) string {
	_, err := conn.Write([]byte(req.Serialize()))
	if err != nil {
		fmt.Println("send command failed, err:", err.Error())
	}
	data := make([]byte, 1024)
	n, err := conn.Read(data)
	if err != nil {
		fmt.Println("read connection failed, err:", err.Error())
	}
	res := string(data[:n])
	fmt.Println("response: ", res)
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

func handshakeToMaster() net.Conn {
	masterAddress := config.replica.masterHost + ":" + config.replica.masterPort
	fmt.Println("handshake with master: " + masterAddress)
	conn := connectToMaster(masterAddress)
	sendCommand(conn, &Resp.RESP{Type: Resp.Array, Data: []*Resp.RESP{{Type: Resp.SimpleString, Data: "PING"}}})
	sendCommand(conn, &Resp.RESP{Type: Resp.Array, Data: []*Resp.RESP{{Type: Resp.BulkString, Data: "REPLCONF"}, {Type: Resp.BulkString, Data: "listening-port"}, {Type: Resp.BulkString, Data: config.port}}})
	sendCommand(conn, &Resp.RESP{Type: Resp.Array, Data: []*Resp.RESP{{Type: Resp.BulkString, Data: "REPLCONF"}, {Type: Resp.BulkString, Data: "capa"}, {Type: Resp.BulkString, Data: "psync2"}}})
	sendCommand(conn, &Resp.RESP{Type: Resp.Array, Data: []*Resp.RESP{{Type: Resp.BulkString, Data: "PSYNC"}, {Type: Resp.BulkString, Data: "?"}, {Type: Resp.BulkString, Data: "-1"}}})
	return conn
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	portPtr := flag.Int("port", 6379, "Port number")
	var replicaConfig ReplicaConfig
	flag.Func("replicaof", "Replica of <master_host> <master_port>", func(flagValue string) error {
		fmt.Println("flagValue: " + flagValue)
		if flagValue == "" {
			return nil
		}
		flagValues := strings.Split(flagValue, " ")
		replicaConfig.masterHost = flagValues[0]
		replicaConfig.masterPort = flagValues[1]
		config.role = "slave"
		return nil
	})
	config.replica = &replicaConfig
	if config.role == "master" {
		config.replica.offset = 0
		config.replica.replicationId = generateRandomString(replicaIdLen)
		replicas = []Replica{}
	}
	flag.Parse()
	port := *portPtr
	config.port = strconv.Itoa(port)
	address := fmt.Sprintf("0.0.0.0:%d", port)
	fmt.Println("Listening on " + address)

	fmt.Println("Replica of " + config.replica.masterHost + ":" + config.replica.masterPort + " role: " + config.role + " port: " + config.port)

	store := store.NewStore()
	if config.role == "slave" {
		masterConn := handshakeToMaster()
		// create a standalone go routine to listen command from master
		go func() {
			for {
				handle(masterConn, store, false)
			}
		}()
	}
	l, err := net.Listen("tcp", address)

	if err != nil {
		panic(fmt.Sprintf("Failed to bind to port %d", port))
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			conn.Close()
			continue
		}
		fmt.Println("handle a connection:")
		go handle(conn, store, true)
	}
}

func handle(conn net.Conn, store *store.Store, isMaster bool) {
	fmt.Println("accept a request, addr:", conn.RemoteAddr().String())
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
		fmt.Println("Read: ", string(p[:n]))
		req, err := Resp.ParseRESP(str)
		if err != nil {
			fmt.Printf("Parse RESP failed, input: %s\nerr: %s\n", str, err.Error())
			break
		}
		fmt.Printf("req.Type: %v, req.Data: %v\n", req.Type, req.Data)
		res := handleCommand(req, conn, store)
		if !isMaster {
			break
		}
		if res.Type == Resp.Array {
			for _, element := range res.Data.([]*Resp.RESP) {
				conn.Write([]byte(element.Serialize()))
				if element.Type == Resp.RDB {
					isMaster = false
				}
			}
		} else {
			conn.Write([]byte(res.Serialize()))
		}
	}
	if isMaster {
		conn.Close()
	}
}

func handleCommand(req *Resp.RESP, conn net.Conn, store *store.Store) *Resp.RESP {
	data := req.Data.([]*resp.RESP)
	if data[0].Type != resp.BulkString {
		return &Resp.RESP{
			Type: Resp.SimpleString,
			Data: "ERR wrong command",
		}
	}
	command := strings.ToUpper(data[0].Data.(string))
	fmt.Printf("command: %s\n", command)
	switch command {
	case "PING":
		if len(data) < 2 {
			return &Resp.RESP{
				Type: Resp.SimpleString,
				Data: "PONG",
			}
		}
		args := strings.ToUpper(data[1].Data.(string))
		fmt.Printf("command: %s, args: %s\n", command, args)
		return &Resp.RESP{
			Type: Resp.BulkString,
			Data: args,
		}
	case "ECHO":
		if len(data) < 2 {
			return &Resp.RESP{
				Type: Resp.SimpleString,
				Data: "ERR wrong number of arguments for command",
			}
		}
		args := strings.ToUpper(data[1].Data.(string))
		fmt.Printf("command: %s, args: %s\n", command, args)
		return &Resp.RESP{
			Type: Resp.BulkString,
			Data: args,
		}
	case "SET":
		if len(data) < 3 {
			return &Resp.RESP{
				Type: Resp.SimpleString,
				Data: "ERR wrong number of arguments for command",
			}
		}
		key := data[1].Data.(string)
		value := data[2].Data.(string)
		with_opts := len(data) > 3
		if with_opts {
			opt := strings.ToUpper(data[3].Data.(string))
			param, err := strconv.ParseInt(data[4].Data.(string), 0, 64)
			if err != nil {
				fmt.Println("Error parsing expire time: ", err.Error())
				return &Resp.RESP{
					Type: Resp.SimpleString,
					Data: "ERR wrong expire time",
				}
			}
			var res *Resp.RESP
			if opt == "PX" {
				store.SetPx(key, value, param)
				res = &Resp.RESP{
					Type: Resp.SimpleString,
					Data: "OK",
				}
			}
			for _, replica := range replicas {
				go func(replica Replica) {
					fmt.Printf("send command to replica: %s\n", replica.conn.RemoteAddr().String())
					sendCommand(replica.conn, req)
				}(replica)
			}
			return res
		}
		store.Set(key, value)
		for _, replica := range replicas {
			go func(replica Replica) {
				fmt.Printf("send command to replica: %s\n", replica.conn.RemoteAddr().String())
				sendCommand(replica.conn, req)
			}(replica)
		}
		return &Resp.RESP{
			Type: Resp.SimpleString,
			Data: "OK",
		}
	case "GET":
		if len(data) < 2 {
			return &Resp.RESP{
				Type: Resp.SimpleString,
				Data: "ERR wrong number of arguments for command",
			}
		}
		key := data[1].Data.(string)
		if value, exist := store.Get(key); exist {
			return &Resp.RESP{
				Type: Resp.SimpleString,
				Data: value,
			}
		}
		return &Resp.RESP{
			Type: Resp.NullBulkString,
			Data: nil,
		}
	case "INFO":
		return &Resp.RESP{
			Type: Resp.BulkString,
			Data: fmt.Sprintf("role:%s\r\nmaster_replid:%s\r\nmaster_repl_offset:%d\r\n", config.role, config.replica.replicationId, config.replica.offset),
		}
	case "REPLCONF":
		if len(data) < 2 {
			return &Resp.RESP{
				Type: Resp.SimpleString,
				Data: "ERR wrong number of arguments for command",
			}
		}
		args := strings.ToUpper(data[1].Data.(string))
		if args == "LISTENING-PORT" {
			replicas = append(replicas, Replica{
				conn: conn,
			})
		}
		return &Resp.RESP{
			Type: Resp.SimpleString,
			Data: "OK",
		}
	case "PSYNC":
		res := Resp.RESP{
			Type: Resp.Array,
			Data: make([]*Resp.RESP, 0),
		}
		res.Data = append(res.Data.([]*Resp.RESP), &Resp.RESP{
			Type: Resp.SimpleString,
			Data: fmt.Sprintf("FULLRESYNC %s %d", config.replica.replicationId, config.replica.offset),
		})
		rdbFile := getRdbFile()

		res.Data = append(res.Data.([]*Resp.RESP), &Resp.RESP{
			Type: Resp.RDB,
			Data: rdbFile,
		})
		return &res
	default:
		return &Resp.RESP{
			Type: Resp.SimpleString,
			Data: "ERR wrong command " + command,
		}
	}
}
