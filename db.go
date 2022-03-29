package main

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
)

type DBKey = string

// Adapted from https://stackoverflow.com/a/68217701/5008494
type Database struct {
	mu        sync.RWMutex
	container map[DBKey]string
	keys      []DBKey
	keyIndex  map[DBKey]int
}

// Read securely from Database
func (db *Database) Read(key DBKey) (v string, ok bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	v, ok = db.container[key]
	return
}

// Write securely to Database
func (db *Database) Write(key DBKey, value string) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.container[key] = value
	db.keys = append(db.keys, key)
	db.keyIndex[key] = len(db.keys) - 1
	return
}

// Delete securely from Database.
func (db *Database) Delete(key DBKey) {
	db.mu.Lock()
	defer db.mu.Unlock()

	index, ok := db.keyIndex[key]
	if !ok {
		return
	}

	delete(db.keyIndex, key)

	lastIndex := len(db.keys) - 1
	wasLastIndex := index == lastIndex

	// swap last key in place of the deleted one and update its index
	if !wasLastIndex {
		db.keys[index] = db.keys[lastIndex]
		lastKey := db.keys[index]
		db.keyIndex[lastKey] = index
	}
	// remove last element from keys slice
	db.keys = db.keys[:lastIndex]

	delete(db.container, key)
	return
}

func (db *Database) RandomKey() (key DBKey) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	index := rand.Intn(len(db.keys))

	return db.keys[index]
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
func (db *SelectedDatabases) Read(conn net.Conn, key DBKey) (v string, ok bool) {
	d := db.GetDB(conn)
	return d.Read(key)
}

// Write securely to Database
func (db *SelectedDatabases) Write(conn net.Conn, key DBKey, value string) {
	d := db.GetDB(conn)
	d.Write(key, value)
}

// Delete securely from Database
func (db *SelectedDatabases) Delete(conn net.Conn, key DBKey) {
	d := db.GetDB(conn)
	d.Delete(key)
}

// Size returns the number of keys stored in the selected database
func (db *SelectedDatabases) Size(conn net.Conn) int {
	d := db.GetDB(conn)
	d.mu.RLock()
	defer d.mu.RUnlock()

	return len(d.container)
}

func (db *SelectedDatabases) RandomKey(conn net.Conn) DBKey {
	d := db.GetDB(conn)
	return d.RandomKey()
}

func initDB(n int) {
	for n >= 0 {
		databases[fmt.Sprint(n)] = &Database{
			container: make(map[DBKey]string),
			keys:      []DBKey{},
			keyIndex:  make(map[DBKey]int),
		}
		n--
	}
}
