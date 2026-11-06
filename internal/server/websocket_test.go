package server

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// Use only loopback fake clients, never the application connection handler.
func TestConcurrentClientWrites(t *testing.T) {
	accepted := make(chan *websocket.Conn, 1)
	release := make(chan struct{})
	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		accepted <- conn
		<-release
	}))
	defer srv.Close()
	defer close(release)
	peer, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer peer.Close()
	conn := <-accepted
	defer conn.Close()
	clientsMu.Lock()
	clients[conn] = &sync.Mutex{}
	clientsMu.Unlock()
	defer func() { clientsMu.Lock(); delete(clients, conn); clientsMu.Unlock() }()

	const iterations = 50
	readDone := make(chan error, 1)
	go func() {
		peer.SetReadDeadline(time.Now().Add(10 * time.Second))
		for i := 0; i < iterations*4; i++ {
			if _, _, err := peer.ReadMessage(); err != nil {
				readDone <- err
				return
			}
		}
		readDone <- nil
	}()
	var writers sync.WaitGroup
	for kind := 0; kind < 4; kind++ {
		writers.Add(1)
		go func(kind int) {
			defer writers.Done()
			for i := 0; i < iterations; i++ {
				switch kind {
				case 0:
					sendStatus(conn, "ready", "test")
				case 1:
					broadcastPreparationFailure(fmt.Errorf("fake failure"))
				case 2:
					broadcastReady()
				case 3:
					broadcastRoomUpdate(nil)
				}
			}
		}(kind)
	}
	writers.Wait()
	if err := <-readDone; err != nil {
		t.Fatal(err)
	}
	clientsMu.Lock()
	delete(clients, conn)
	clientsMu.Unlock()
	if err := writeClientMessage(conn, websocket.TextMessage, nil); err != net.ErrClosed {
		t.Fatalf("late write = %v, want closed", err)
	}
	clientsMu.Lock()
	_, exists := clients[conn]
	clientsMu.Unlock()
	if exists {
		t.Fatal("late write recreated client state")
	}
}
