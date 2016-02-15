package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/progrium/duplex-hub/Godeps/_workspace/src/golang.org/x/net/websocket"
)

type Endpoint struct {
	sync.Mutex
	path     string
	secret   string
	upstream *websocket.Conn
	counter  int
	clients  map[string]io.Writer
}

func (e *Endpoint) AddClient(conn io.Writer) string {
	e.Lock()
	defer e.Unlock()
	e.counter++
	h := sha1.New()
	h.Write([]byte(os.Getenv("TOKEN")))
	h.Write([]byte(strconv.Itoa(e.counter)))
	routeKey := hex.EncodeToString(h.Sum(nil))
	e.clients[routeKey] = conn
	return routeKey
}

func (e *Endpoint) DropClient(routeKey string) {
	e.Lock()
	defer e.Unlock()
	delete(e.clients, routeKey)
}

func (e *Endpoint) SendUpstream(msg interface{}) error {
	frame, err := json.Marshal(&msg)
	if err != nil {
		// TODO: what happens on upstream encode error
		panic(err)
	}
	if os.Getenv("DEBUG") != "" {
		fmt.Println("<<<", string(frame))
	}
	_, err = e.upstream.Write(frame)
	return err
}
