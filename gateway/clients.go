package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/progrium/duplex-hub/Godeps/_workspace/src/github.com/progrium/duplex/golang"
	"github.com/progrium/duplex-hub/Godeps/_workspace/src/golang.org/x/net/websocket"
)

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
		AcceptHandshake(ws) // TODO handle error
		routeKey := endpoint.AddClient(ws)
		log.Println("Endpoint", endpoint.path, "client connected", r.RemoteAddr)
		for {
			frameBuf := make([]byte, duplex.MaxFrameSize)
			n, err := ws.Read(frameBuf)
			if err != nil {
				endpoint.DropClient(routeKey)
				break
			}
			var msg duplex.Message
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
			err = endpoint.SendUpstream(msg)
			if err != nil {
				// TODO: handle upstream write error
				// remove upstream endpoint?
				break
			}
		}
	}}
	s.ServeHTTP(w, r)
}
