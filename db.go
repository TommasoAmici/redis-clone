package main

import (
	"fmt"
	"net"
	"sync"
)

type Database struct {
	mu sync.RWMutex
	v  map[string]string
}

// Read securely from Database
func (db *Database) Read(key string) (v string, ok bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	v, ok = db.v[key]
	return
}

// Write securely to Database
func (db *Database) Write(key string, value string) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.v[key] = value
	return
}

// Delete securely from Database
func (db *Database) Delete(key string) {
	db.mu.Lock()
	defer db.mu.Unlock()

	delete(db.v, key)
	return
}

type DatabaseMap = map[string]*Database

var databases = make(DatabaseMap)

type SelectedDatabases struct {
	mu sync.RWMutex
	v  DatabaseMap
}

// Every connection can select a different database index
var selectedDB = SelectedDatabases{
	v: make(DatabaseMap),
}

func (db *SelectedDatabases) GetDB(conn net.Conn) *Database {
	db.mu.RLock()
	defer db.mu.RUnlock()

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
	return d.Read(key)
}

// Write securely to Database
func (db *SelectedDatabases) Write(conn net.Conn, key string, value string) {
	d := db.GetDB(conn)
	d.Write(key, value)
}

// Delete securely from Database
func (db *SelectedDatabases) Delete(conn net.Conn, key string) {
	d := db.GetDB(conn)
	d.Delete(key)
}

func initDB(n int) {
	for n >= 0 {
		databases[fmt.Sprint(n)] = &Database{v: make(map[string]string)}
		n--
	}
}
