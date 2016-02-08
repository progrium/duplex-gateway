package main

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/progrium/duplex-hub/Godeps/_workspace/src/github.com/progrium/simplex/golang"
	"github.com/progrium/duplex-hub/Godeps/_workspace/src/golang.org/x/net/websocket"
)

func _connect(rpc *simplex.RPC, path string, backend bool) (*simplex.Peer, error) {
	var baseUrl, token, secret, url string
	if os.Getenv("HUB_URL") != "" {
		baseUrl = os.Getenv("HUB_URL")
	} else {
		baseUrl = "ws://localhost:8080"
	}
	if os.Getenv("SECRET") != "" {
		secret = os.Getenv("SECRET")
	} else {
		secret = "test"
	}
	if backend {
		if os.Getenv("TOKEN") != "" {
			token = os.Getenv("TOKEN")
		} else {
			token = "dev"
		}
		url = fmt.Sprintf("%s%s?token=%s&secret=%s",
			baseUrl, path, token, secret)
	} else {
		url = fmt.Sprintf("%s%s?secret=%s",
			baseUrl, path, secret)
	}
	ws, err := websocket.Dial(url, "", "http://example.com")
	if err != nil {
		return nil, err
	}
	return rpc.Handshake(ws)
}

func ConnectBackend(rpc *simplex.RPC, path string) (*simplex.Peer, error) {
	return _connect(rpc, path, true)
}

func ConnectClient(rpc *simplex.RPC, path string) (*simplex.Peer, error) {
	return _connect(rpc, path, false)
}

func TestClientToBackendRoundtrip(t *testing.T) {
	backendRpc := simplex.NewRPC(simplex.NewJSONCodec())
	backendRpc.Register("echo", func(ch *simplex.Channel) error {
		var obj interface{}
		if _, err := ch.Recv(&obj); err != nil {
			return err
		}
		return ch.Send(obj, false)
	})

	backend, err := ConnectBackend(backendRpc, "/test")
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()

	clientRpc := simplex.NewRPC(simplex.NewJSONCodec())
	client, err := ConnectClient(clientRpc, "/test")
	if err != nil {
		t.Fatal(err)
	}

	var reply map[string]interface{}
	err = client.Call("echo", map[string]string{"foo": "bar"}, &reply)
	if err != nil {
		t.Fatal(err)
	}
	if reply["foo"] != "bar" {
		t.Fatal("Unexpected reply:", reply)
	}
}

func TestMultipleClients(t *testing.T) {
	backendRpc := simplex.NewRPC(simplex.NewJSONCodec())
	backendRpc.Register("echo", func(ch *simplex.Channel) error {
		var obj interface{}
		if _, err := ch.Recv(&obj); err != nil {
			return err
		}
		return ch.Send(obj, false)
	})

	backend, err := ConnectBackend(backendRpc, "/test")
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		clientRpc1 := simplex.NewRPC(simplex.NewJSONCodec())
		client1, err := ConnectClient(clientRpc1, "/test")
		if err != nil {
			t.Fatal(err)
		}
		var reply map[string]interface{}
		err = client1.Call("echo", map[string]string{"from": "client1"}, &reply)
		if err != nil {
			t.Fatal(err)
		}
		if reply["from"] != "client1" {
			t.Fatal("Unexpected reply:", reply)
		}
		client1.Close()
		wg.Done()
	}()
	go func() {
		clientRpc2 := simplex.NewRPC(simplex.NewJSONCodec())
		client2, err := ConnectClient(clientRpc2, "/test")
		if err != nil {
			t.Fatal(err)
		}
		var reply map[string]interface{}
		err = client2.Call("echo", map[string]string{"from": "client2"}, &reply)
		if err != nil {
			t.Fatal(err)
		}
		if reply["from"] != "client2" {
			t.Fatal("Unexpected reply:", reply)
		}
		client2.Close()
		wg.Done()
	}()
	wg.Wait()
}
