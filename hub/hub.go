package main

import (
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/progrium/duplex-hub/Godeps/_workspace/src/github.com/progrium/duplex/golang"
	"github.com/progrium/duplex-hub/Godeps/_workspace/src/golang.org/x/net/websocket"
)

var (
	endpoints map[string]*Endpoint
	mu        sync.Mutex
)

func AcceptHandshake(ws *websocket.Conn) error {
	buf := make([]byte, 32)
	_, err := ws.Read(buf) // TODO: timeout
	if err != nil {
		return err
	}
	// TODO: check handshake
	_, err = ws.Write([]byte(duplex.HandshakeAccept))
	if err != nil {
		return err
	}
	return nil
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("x-forwarded-proto") != "https" && os.Getenv("NOTLS") == "" {
		http.Error(w, "TLS required", http.StatusForbidden)
		return
	}
	if r.URL.Query().Get("token") == os.Getenv("TOKEN") {
		HandleBackend(w, r)
	} else {
		switch r.Method {
		case "GET":
			HandleClient(w, r)
		case "POST":
			HandleHttp(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func main() {
	endpoints = make(map[string]*Endpoint)
	http.HandleFunc("/", Handler)
	port := ":" + os.Getenv("PORT")
	log.Println("Duplex Hub started on", port, "...")
	log.Fatal(http.ListenAndServe(port, nil))
}
