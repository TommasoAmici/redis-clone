package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

func main() {
	network := flag.String("network", "tcp", `The network must be "tcp", "tcp4", "tcp6", "unix" or "unixpacket".`)
	addr := flag.String("address", "127.0.0.1:6379", "Address to listen on")
	dbNum := flag.Int("db-num", 16, "Number of databases to create")
	flag.Parse()

	initDB(*dbNum)

	ln, err := net.Listen(*network, *addr)
	if err != nil {
		log.Fatalln("[ERROR] Failed to start listening on", *addr)
	} else {
		log.Println("[INFO] Listening on", *addr)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("[ERROR]", err)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadString('\n')
		if err != nil || msg == "" {
			return
		}
		if msg[0] == '*' {
			handleURP(reader, conn, msg)
		} else {
			handleInlineCommand(conn, msg)
		}
	}
}

var commandMap = map[string]func(conn net.Conn, args []string) error{
	"dbsize":    DBSize,
	"decr":      IncrDecrGenerator(DirDecr, false),
	"decrby":    IncrDecrGenerator(DirDecr, true),
	"del":       Del,
	"echo":      Echo,
	"exists":    Exists,
	"flushall":  FlushAll,
	"flushdb":   FlushDB,
	"get":       Get,
	"incr":      IncrDecrGenerator(DirIncr, false),
	"incrby":    IncrDecrGenerator(DirIncr, true),
	"move":      Move,
	"ping":      Ping,
	"quit":      Quit,
	"randomkey": RandomKey,
	"select":    Select,
	"set":       Set,
}

func handleCommand(conn net.Conn, command string, args []string) {
	handler, ok := commandMap[command]
	if !ok {
		return
	}
	err := handler(conn, args)
	if err == wrongNumArgsError {
		wrongNumArgsRESP(conn, command)
	}
}

// A client sends the Redis server a RESP Array consisting of only Bulk Strings.
// A Redis server replies to clients, sending any valid RESP data type as a reply.
// So for example a typical interaction could be the following.
// The client sends the command `LLEN mylist` in order to get the length of the list
// stored at key `mylist`. Then the server replies with an Integer reply as in the
// following example (C: is the client, S: the server).
//     C: *2\r\n
//     C: $4\r\n
//     C: LLEN\r\n
//     C: $6\r\n
//     C: mylist\r\n
//     S: :48293\r\n
// As usual, we separate different parts of the protocol with newlines for simplicity,
// but the actual interaction is the client sending
//     *2\r\n$4\r\nLLEN\r\n$6\r\nmylist\r\n.
// https://redis.io/docs/reference/protocol-spec/#send-commands-to-a-redis-server
func handleURP(reader *bufio.Reader, conn net.Conn, msg string) {
	arrayLen, err := strconv.Atoi(strings.TrimSpace(msg[1:]))
	if err != nil {
		log.Println("[ERROR]", err)
		return
	}
	args := []string{}
	for arrayLen > 0 {
		_, err = reader.ReadString('\n')
		if err != nil {
			log.Println("[ERROR]", err)
			return
		}
		arg, err := reader.ReadString('\n')
		if err != nil {
			log.Println("[ERROR]", err)
			return
		}
		args = append(args, strings.TrimSpace(arg))
		arrayLen--
	}
	log.Println("[INFO] unified request protocol received", args)

	command := args[0]
	args = args[1:]
	handleCommand(conn, command, args)
}

// While the Redis protocol is simple to implement, it is not ideal to use in interactive
// sessions, and redis-cli may not always be available. For this reason, Redis also
// accepts commands in the inline command format.
// Basically, you write space-separated arguments in a telnet session. Since no command
// starts with * that is instead used in the unified request protocol, Redis is able to
// detect this condition and parse your command.
// https://redis.io/docs/reference/protocol-spec/#inline-commands
func handleInlineCommand(conn net.Conn, msg string) {
	log.Println("[INFO] inline command received:", msg)

	msg = strings.TrimSpace(msg)
	split := strings.Split(msg, " ")
	command := strings.ToLower(split[0])
	args := split[1:]

	handleCommand(conn, command, args)
}

// Ping returns PONG if no argument is provided, otherwise return a copy of the argument as a bulk.
// This command is often used to test if a connection is still alive, or to measure latency.
// https://redis.io/commands/ping/
func Ping(conn net.Conn, args []string) error {
	if len(args) == 0 {
		simpleStringRESP(conn, "PONG")
	} else if len(args) == 1 {
		bulkStringRESP(conn, args[0])
	} else {
		return wrongNumArgsError
	}
	return nil
}

// Echo `message` returns `message`.
// https://redis.io/commands/echo/
func Echo(conn net.Conn, args []string) error {
	if len(args) != 1 {
		return wrongNumArgsError
	}
	bulkStringRESP(conn, args[0])
	return nil
}

// Set `key` to hold the string value. If `key` already holds a value, it is overwritten,
// regardless of its type. Any previous time to live associated with the `key` is
// discarded on successful `SET` operation.
// https://redis.io/commands/set/
func Set(conn net.Conn, args []string) error {
	if len(args) != 2 {
		return wrongNumArgsError
	}
	selectedDB.Write(conn, args[0], args[1])
	okRESP(conn)
	return nil
}

// Get the value of `key`. If the `key`` does not exist the special value `nil` is returned.
// An error is returned if the value stored at `key` is not a string, because `GET` only
// handles string values.
// https://redis.io/commands/get/
func Get(conn net.Conn, args []string) error {
	if len(args) != 1 {
		return wrongNumArgsError
	}
	val, ok := selectedDB.Read(conn, args[0])
	if ok {
		bulkStringRESP(conn, val)
	} else {
		nullBulkRESP(conn)
	}
	return nil
}

// Exists returns a value if `key` exists.
// The user should be aware that if the same existing `key` is mentioned in the arguments
// multiple times, it will be counted multiple times. So if `somekey` exists, `EXIST somekey somekey` will return 2.
// https://redis.io/commands/exists/
func Exists(conn net.Conn, args []string) error {
	if len(args) == 0 {
		return wrongNumArgsError
	}
	count := 0
	for _, arg := range args {
		if v, _ := selectedDB.Read(conn, arg); v != "" {
			count++
		}
	}
	intRESP(conn, count)
	return nil

}

// Del removes the specified keys. A key is ignored if it does not exist.
// Returns Integer reply: The number of keys that were removed.
// https://redis.io/commands/del/
func Del(conn net.Conn, args []string) error {
	if len(args) == 0 {
		return wrongNumArgsError
	}
	count := 0
	for _, arg := range args {
		if v, _ := selectedDB.Read(conn, arg); v != "" {
			selectedDB.Delete(conn, arg)
			count++
		}
	}
	intRESP(conn, count)
	return nil
}

// Select the Redis logical database having the specified zero-based numeric index.
// New connections always use the database 0. https://redis.io/commands/select/
func Select(conn net.Conn, args []string) error {
	if len(args) != 1 {
		return wrongNumArgsError
	}
	selectedDB.mu.Lock()
	selectedDB.v[conn.RemoteAddr().String()] = databases[args[0]]
	selectedDB.mu.Unlock()
	okRESP(conn)
	return nil
}

// Move `key` from the currently selected database (see `SELECT`) to the specified
// destination database. When `key` already exists in the destination database, or it
// does not exist in the source database, it does nothing.
// It is possible to use `MOVE` as a locking primitive because of this.
// https://redis.io/commands/move/
func Move(conn net.Conn, args []string) error {
	if len(args) != 2 {
		return wrongNumArgsError
	}

	key := args[0]
	dbIdx := args[1]
	value, ok := selectedDB.Read(conn, key)
	if !ok {
		intRESP(conn, 0)
		return nil
	}
	newDB, ok := databases[dbIdx]
	if !ok {
		errRESP(conn, "ERR DB index is out of range")
		return nil
	}
	_, ok = newDB.Read(key)
	if ok {
		intRESP(conn, 0)
		return nil
	}
	go newDB.Write(key, value)
	go selectedDB.Delete(conn, key)
	intRESP(conn, 1)
	return nil
}

// RandomKey returns a random key from the currently selected database.
// This function relies on the fact that Go iterates randomly over maps https://go.dev/doc/go1#iteration.
// https://redis.io/commands/randomkey/
func RandomKey(conn net.Conn, args []string) error {
	if len(args) != 0 {
		return wrongNumArgsError
	}

	bulkStringRESP(conn, selectedDB.RandomKey(conn))
	return nil
}

const (
	DirIncr = iota
	DirDecr
)

// Increments or decrements the number stored at key by one or by the value provided.
// If the key does not exist, it is set to 0 before performing the operation.
// An error is returned if the key contains a value of the wrong type or contains a
// string that can not be represented as integer. This operation is limited to 64 bit signed integers.
// Note: this is a string operation because Redis does not have a dedicated integer type.
// The string stored at the key is interpreted as a base-10 64 bit signed integer to
// execute the operation.
// Redis stores integers in their integer representation, so for string values that
// actually hold an integer, there is no overhead for storing the string representation
// of the integer.
// https://redis.io/commands/incr/
// https://redis.io/commands/decr/
// https://redis.io/commands/incrby/
// https://redis.io/commands/decrby/
func IncrDecrGenerator(dir int, by bool) func(conn net.Conn, args []string) error {
	var sum func(a, b int) int

	if dir == DirDecr {
		sum = func(a, b int) int {
			return a - b
		}
	} else {
		sum = func(a, b int) int {
			return a + b
		}
	}

	// DECRBY and INCRBY accept two arguments
	numArgs := 1
	if by {
		numArgs = 2
	}

	return func(conn net.Conn, args []string) error {
		if len(args) != numArgs {
			return wrongNumArgsError
		}

		key := args[0]

		val, err := selectedDB.ReadInt(conn, key)
		if err != nil {
			if err == KeyDoesNotExist {
				var v int
				if by {
					v, err = strconv.Atoi(args[1])
					if err != nil {
						valueIsNotIntRESP(conn)
						return nil
					}
				} else {
					v = 1
				}
				selectedDB.Write(conn, key, fmt.Sprint(v))
				intRESP(conn, v)
				return nil
			} else {
				valueIsNotIntRESP(conn)
				return nil
			}
		}

		var v int
		if by {
			changeBy, err := strconv.Atoi(args[1])
			if err != nil {
				valueIsNotIntRESP(conn)
				return nil
			}
			v = sum(val, changeBy)
		} else {
			v = sum(val, 1)
		}
		selectedDB.Write(conn, key, fmt.Sprint(v))
		intRESP(conn, v)
		return nil
	}
}

// DBSize returns the number of keys in the currently-selected database.
// https://redis.io/commands/dbsize/
func DBSize(conn net.Conn, args []string) error {
	if len(args) != 0 {
		wrongNumArgsRESP(conn, "dbsize")
		return wrongNumArgsError
	}
	intRESP(conn, selectedDB.Size(conn))
	return nil
}

func FlushDB(conn net.Conn, args []string) error {
	if len(args) != 0 {
		return wrongNumArgsError
	}
	selectedDB.Flush(conn)
	okRESP(conn)
	return nil
}

// FlushAll delete all the keys of all the existing databases, not just
// the currently selected one.
// https://redis.io/commands/flushall/
func FlushAll(conn net.Conn, args []string) error {
	if len(args) != 0 {
		return wrongNumArgsError
	}
	for _, d := range databases {
		d.Flush()
	}
	okRESP(conn)
	return nil
}

// Quit closes the connection. https://redis.io/commands/quit/
func Quit(conn net.Conn, args []string) error {
	selectedDB.mu.Lock()
	delete(selectedDB.v, conn.RemoteAddr().String())
	selectedDB.mu.Unlock()
	okRESP(conn)
	conn.Close()
	return nil
}
