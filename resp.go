package main

import (
	"fmt"
	"net"
)

const (
	RESP_STRING = '+'
	RESP_INT    = ':'
	RESP_ERROR  = '-'
	RESP_BULK   = '$'
)

// This type is just a CRLF-terminated string that represents an integer, prefixed by a
// ':' byte. For example, ":0\r\n" and ":1000\r\n" are integer replies.
// https://redis.io/docs/reference/protocol-spec/#resp-integers
func intRESP(conn net.Conn, n int) {
	conn.Write([]byte(fmt.Sprintf("%c%d\r\n", RESP_INT, n)))
}

// Simple Strings are encoded as follows: a plus character, followed by a string that
// cannot contain a CR or LF character (no newlines are allowed), and terminated by CRLF (that is "\r\n").
// For example:
//     "+OK\r\n"
// https://redis.io/docs/reference/protocol-spec/#resp-simple-strings
func simpleStringRESP(conn net.Conn, s string) {
	conn.Write([]byte(fmt.Sprintf("%c%s\r\n", RESP_STRING, s)))
}

func okRESP(conn net.Conn) {
	simpleStringRESP(conn, "OK")
}

// Bulk Strings are used in order to represent a single binary-safe string up to 512 MB in length.
// Bulk Strings are encoded in the following way:
//     - A '$' byte followed by the number of bytes composing the string (a prefixed length), terminated by CRLF.
//     - The actual string data.
//     - A final CRLF.
// So the string "hello" is encoded as follows:
//     "$6\r\nhello\r\n"
// https://redis.io/docs/reference/protocol-spec/#resp-bulk-strings
func bulkStringRESP(conn net.Conn, s string) {
	conn.Write([]byte(fmt.Sprintf("%c%d\r\n%s\r\n", RESP_BULK, len(s), s)))
}

// RESP Bulk Strings can also be used in order to signal non-existence of a value using
// a special format to represent a Null value. In this format, the length is -1, and
// there is no data. Null is represented as:
//     "$-1\r\n"
// This is called a Null Bulk String.
func nullBulkRESP(conn net.Conn) {
	conn.Write([]byte(fmt.Sprintf("%c-1\r\n", RESP_BULK)))
}

// RESP has a specific data type for errors. They are similar to RESP Simple Strings,
// but the first character is a minus ‘-’ character instead of a plus. The real
// difference between Simple Strings and Errors in RESP is that clients treat errors
// as exceptions, and the string that composes the Error type is the error message itself.
// https://redis.io/docs/reference/protocol-spec/#resp-errors
func errRESP(conn net.Conn, msg string) {
	conn.Write([]byte(fmt.Sprintf("%c%s\r\n", RESP_ERROR, msg)))
}

func wrongNumArgsRESP(conn net.Conn, name string) {
	errRESP(conn, "ERR wrong number of arguments for '"+name+"' command")
}
