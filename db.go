package main

import (
	"fmt"
	"net"
	"sync"
)

type Database struct {
	mu sync.Mutex
	v  map[string]string
}

type DatabaseMap = map[string]*Database

var databases = make(DatabaseMap)

type SelectedDatabases struct {
	mu sync.Mutex
	v  DatabaseMap
}

// Every connection can select a different database index
var selectedDB = SelectedDatabases{
	v: make(DatabaseMap),
}

func (db *SelectedDatabases) GetDB(conn net.Conn) *Database {
	db.mu.Lock()
	defer db.mu.Unlock()

	d, ok := db.v[conn.RemoteAddr().String()]
	if ok {
		return d
	}
	db.v[conn.RemoteAddr().String()] = databases["0"]
	return databases["0"]
}

// Read securely from Database
func (db *SelectedDatabases) Read(conn net.Conn, key string) (v string, ok bool) {
	d := db.GetDB(conn)
	d.mu.Lock()
	defer d.mu.Unlock()

	v, ok = d.v[key]
	return
}

// Write securely to Database
func (db *SelectedDatabases) Write(conn net.Conn, key string, value string) {
	d := db.GetDB(conn)
	d.mu.Lock()
	defer d.mu.Unlock()

	d.v[key] = value
	return
}

// Delete securely from Database
func (db *SelectedDatabases) Delete(conn net.Conn, key string) {
	d := db.GetDB(conn)
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.v, key)
	return
}

func initDB(n int) {
	for n >= 0 {
		databases[fmt.Sprint(n)] = &Database{v: make(map[string]string)}
		n--
	}
}
