package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"strings"
)

func main() {
	network := flag.String("network", "tcp", `The network must be "tcp", "tcp4", "tcp6", "unix" or "unixpacket".`)
	addr := flag.String("address", "127.0.0.1:6379", "Address to listen on")
	flag.Parse()

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

var handlers = map[string]func(conn net.Conn, message string, args []string){
	"ping":   Ping,
	"set":    Set,
	"get":    Get,
	"del":    Del,
	"exists": Exists,
	"quit":   Quit,
}

var db = map[string]string{}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadString('\n')
		if err != nil || msg == "" {
			return
		}
		log.Println("[INFO] Message Received:", msg)

		msg = strings.TrimSpace(msg)
		split := strings.Split(msg, " ")
		command := strings.ToLower(split[0])
		args := split[1:]

		handler, ok := handlers[command]
		if !ok {
			return
		}
		handler(conn, msg, args)
	}
}

// Ping returns PONG if no argument is provided, otherwise return a copy of the argument as a bulk.
// This command is often used to test if a connection is still alive, or to measure latency.
// https://redis.io/commands/ping/
func Ping(conn net.Conn, msg string, args []string) {
	if len(args) == 0 {
		simpleStringRESP(conn, "PONG")
	} else if len(args) == 1 {
		simpleStringRESP(conn, args[0])
	} else {
		wrongNumArgsRESP(conn, "ping")
	}
}

// Set `key` to hold the string value. If `key` already holds a value, it is overwritten,
// regardless of its type. Any previous time to live associated with the `key` is
// discarded on successful `SET` operation.
// https://redis.io/commands/set/
func Set(conn net.Conn, msg string, args []string) {
	if len(args) != 2 {
		wrongNumArgsRESP(conn, "set")
	} else {
		db[args[0]] = args[1]
		okRESP(conn)
	}
}

// Get the value of `key`. If the `key`` does not exist the special value `nil` is returned.
// An error is returned if the value stored at `key` is not a string, because `GET` only
// handles string values.
// https://redis.io/commands/get/
func Get(conn net.Conn, msg string, args []string) {
	if len(args) != 1 {
		wrongNumArgsRESP(conn, "get")
	} else {
		val, ok := db[args[0]]
		if ok {
			bulkStringRESP(conn, val)
		} else {
			nullBulkRESP(conn)
		}
	}
}

// Exists returns a value if `key` exists.
// The user should be aware that if the same existing `key` is mentioned in the arguments
// multiple times, it will be counted multiple times. So if `somekey` exists, `EXIST somekey somekey` will return 2.
// https://redis.io/commands/exists/
func Exists(conn net.Conn, msg string, args []string) {
	if len(args) == 0 {
		wrongNumArgsRESP(conn, "exists")
	} else {
		count := 0
		for _, arg := range args {
			if db[arg] != "" {
				count++
			}
		}
		intRESP(conn, count)
	}
}

// Del removes the specified keys. A key is ignored if it does not exist.
// Returns Integer reply: The number of keys that were removed.
// https://redis.io/commands/del/
func Del(conn net.Conn, msg string, args []string) {
	if len(args) == 0 {
		wrongNumArgsRESP(conn, "del")
	} else {
		count := 0
		for _, arg := range args {
			if db[arg] != "" {
				delete(db, arg)
				count++
			}
		}
		intRESP(conn, count)
	}
}

// Quit closes the connection. https://redis.io/commands/quit/
func Quit(conn net.Conn, msg string, args []string) {
	okRESP(conn)
	conn.Close()
}
