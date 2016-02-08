package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/progrium/duplex-hub/Godeps/_workspace/src/github.com/progrium/simplex/golang"
	"github.com/progrium/duplex-hub/Godeps/_workspace/src/golang.org/x/net/websocket"
)

var (
	endpoints map[string]*Endpoint
	mu        sync.Mutex
)

type Endpoint struct {
	sync.Mutex
	path     string
	secret   string
	upstream *websocket.Conn
	counter  int
	clients  map[string]*websocket.Conn
}

func RouteKey(id int) string {
	h := sha1.New()
	h.Write([]byte(os.Getenv("TOKEN")))
	h.Write([]byte(strconv.Itoa(id)))
	return hex.EncodeToString(h.Sum(nil))
}

func AcceptSimplexHandshake(ws *websocket.Conn) error {
	buf := make([]byte, 32)
	_, err := ws.Read(buf) // TODO: timeout
	if err != nil {
		return err
	}
	// TODO: check handshake
	_, err = ws.Write([]byte(simplex.HandshakeAccept))
	if err != nil {
		return err
	}
	return nil
}

func HandleClient(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	endpoint, exists := endpoints[r.URL.Path]
	mu.Unlock()
	if !exists {
		http.NotFound(w, r)
		return
	}
	if r.URL.Query().Get("secret") != endpoint.secret {
		http.Error(w, "Bad secret", http.StatusUnauthorized)
		return
	}
	s := websocket.Server{Handler: func(ws *websocket.Conn) {
		AcceptSimplexHandshake(ws) // TODO handle error
		endpoint.Lock()
		endpoint.counter++
		routeKey := RouteKey(endpoint.counter)
		endpoint.clients[routeKey] = ws
		endpoint.Unlock()
		log.Println("Endpoint", endpoint.path, "client connected", r.RemoteAddr)
		for {
			frameBuf := make([]byte, simplex.MaxFrameSize)
			n, err := ws.Read(frameBuf)
			if err != nil {
				endpoint.Lock()
				delete(endpoint.clients, routeKey)
				endpoint.Unlock()
				break
			}
			var msg simplex.Message
			err = json.Unmarshal(frameBuf[:n], &msg)
			if err != nil {
				// TODO: what happens on decode error
				panic(err)
			}
			var ext map[string]interface{}
			if msg.Ext == nil {
				ext = make(map[string]interface{})
			} else {
				ext = msg.Ext.(map[string]interface{})
			}
			ext["route"] = routeKey
			msg.Ext = ext
			frame, err := json.Marshal(&msg)
			if err != nil {
				// TODO: what happens on encode error
				panic(err)
			}
			if os.Getenv("DEBUG") != "" {
				fmt.Println("<<<", string(frame))
			}
			_, err = endpoint.upstream.Write(frame)
			if err != nil {
				// TODO: handle upstream write error
				// remove upstream endpoint?
				break
			}
		}
	}}
	s.ServeHTTP(w, r)
}

func HandleBackend(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("secret") == "" {
		http.Error(w, "Secret required", http.StatusBadRequest)
		return
	}
	s := websocket.Server{Handler: func(ws *websocket.Conn) {
		endpoint := &Endpoint{
			path:    ws.Request().URL.Path,
			secret:  ws.Request().URL.Query().Get("secret"),
			counter: 0,
			clients: make(map[string]*websocket.Conn),
		}
		AcceptSimplexHandshake(ws) // TODO handle error
		endpoint.upstream = ws
		mu.Lock()
		old, exists := endpoints[endpoint.path]
		if exists {
			old.upstream.Close()
		}
		endpoints[endpoint.path] = endpoint
		mu.Unlock()
		log.Println("Endpoint", endpoint.path, "backend online")
		for {
			frame := make([]byte, simplex.MaxFrameSize)
			n, err := ws.Read(frame)
			if err != nil {
				// TODO: what happens on read error
				log.Println("read err")
				break
			}
			var msg simplex.Message
			err = json.Unmarshal(frame[:n], &msg)
			if err != nil {
				// TODO: what happens on decode error
				panic(err)
			}
			if msg.Ext == nil {
				log.Println("no ext")
				continue
			}
			ext, ok := msg.Ext.(map[string]interface{})
			if !ok {
				log.Println("ext not object")
				continue
			}
			routeKey := ext["route"].(string)
			endpoint.Lock()
			clientConn, ok := endpoint.clients[routeKey]
			endpoint.Unlock()
			if !ok {
				log.Println("routekey not found")
				continue
			}
			if os.Getenv("DEBUG") != "" {
				fmt.Println(">>>", string(frame[:n]))
			}
			_, err = clientConn.Write(frame[:n])
			if err != nil {
				endpoint.Lock()
				delete(endpoint.clients, routeKey)
				endpoint.Unlock()
			}
		}
		log.Println("Endpoint", endpoint.path, "backend offline")
	}}
	s.ServeHTTP(w, r)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("x-forwarded-proto") != "https" && os.Getenv("NOTLS") == "" {
		http.Error(w, "TLS required", http.StatusForbidden)
		return
	}
	if r.URL.Query().Get("token") == os.Getenv("TOKEN") {
		HandleBackend(w, r)
	} else {
		HandleClient(w, r)
	}
}

func main() {
	endpoints = make(map[string]*Endpoint)
	http.HandleFunc("/", Handler)
	port := ":" + os.Getenv("PORT")
	log.Println("Duplex Hub started on", port, "...")
	log.Fatal(http.ListenAndServe(port, nil))
}
