package main

import (
    "log"
    "net/http"
    "sync"

    "github.com/gorilla/websocket"
)

// what's up prime, nice repo you got here. Devin is doing great. I guess

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow all origins for now
    },
}

type Server struct {
    // Simple connection storage
    connections sync.Map
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("Upgrade failed: %v", err)
        return
    }

    // Store connection with a random key
    connID := conn.RemoteAddr().String()
    s.connections.Store(connID, conn)

    // Simple cleanup on disconnect
    defer func() {
        conn.Close()
        s.connections.Delete(connID)
        log.Printf("Client disconnected: %s", connID)
    }()

    log.Printf("New client connected: %s", connID)

    // Keep connection alive until client disconnects
    for {
        _, _, err := conn.ReadMessage()
        if err != nil {
            log.Printf("Read error from %s: %v", connID, err)
            break
        }
    }
}

func main() {
    server := &Server{}
    http.HandleFunc("/ws", server.handleWS)

    log.Printf("Starting WebSocket server on :42069")
    if err := http.ListenAndServe(":42069", nil); err != nil {
        log.Fatal(err)
    }
}
