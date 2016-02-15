package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/progrium/duplex/golang"
)

type PayloadWriter struct {
	http.ResponseWriter
	done chan bool
}

func (pw *PayloadWriter) Write(p []byte) (int, error) {
	var msg duplex.Message
	err := json.Unmarshal(p, &msg)
	if err != nil {
		return 0, err
	}
	if !msg.More {
		pw.done <- true
	}
	b, err := json.Marshal(msg.Payload)
	return pw.ResponseWriter.Write(b)
}

func HandleHttp(w http.ResponseWriter, r *http.Request) {
	endpointPath := path.Dir(r.URL.Path)
	method := path.Base(r.URL.Path)
	mu.Lock()
	endpoint, exists := endpoints[endpointPath]
	mu.Unlock()
	if !exists {
		http.NotFound(w, r)
		return
	}
	if r.URL.Query().Get("secret") != endpoint.secret {
		http.Error(w, "Bad secret", http.StatusUnauthorized)
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}
	var payload interface{}
	if len(b) > 0 {
		err = json.Unmarshal(b, &payload)
		if err != nil {
			http.Error(w, "Error decoding JSON", http.StatusBadRequest)
			return
		}
	}
	done := make(chan bool)
	writer := &PayloadWriter{w, done}
	routeKey := endpoint.AddClient(writer)
	msg := &duplex.Message{
		Type:    duplex.TypeRequest,
		Method:  method,
		Payload: payload,
		Id:      1,
		Ext:     map[string]string{"route": routeKey},
	}
	err = endpoint.SendUpstream(msg)
	if err != nil {
		http.Error(w, "Upstream write error", http.StatusServiceUnavailable)
		// remove upstream endpoint?
		return
	}
	if r.URL.Query().Get("async") != "true" {
		<-done
	}
}
