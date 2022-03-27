package main

import (
	"bufio"
	"flag"
	"fmt"
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
}
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
		conn.Write([]byte("PONG\n"))
	} else if len(args) == 1 {
		conn.Write([]byte(args[0] + "\n"))
	} else {
		wrongNumberArgs(conn, "ping")
	}
}

func wrongNumberArgs(conn net.Conn, name string) {
	conn.Write([]byte("ERR wrong number of arguments for '" + name + "' command\n"))
}
