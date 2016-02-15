package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/progrium/duplex-hub/Godeps/_workspace/src/github.com/progrium/duplex/golang"
	"github.com/progrium/duplex-hub/Godeps/_workspace/src/golang.org/x/net/websocket"
)

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
			clients: make(map[string]io.Writer),
		}
		AcceptHandshake(ws) // TODO handle error
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
			frame := make([]byte, duplex.MaxFrameSize)
			n, err := ws.Read(frame)
			if err != nil {
				// TODO: what happens on read error
				log.Println("read err:", err)
				break
			}
			var msg duplex.Message
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
				endpoint.DropClient(routeKey)
			}
		}
		log.Println("Endpoint", endpoint.path, "backend offline")
	}}
	s.ServeHTTP(w, r)
}
