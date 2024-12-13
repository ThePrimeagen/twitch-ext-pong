package main

import (
    "log"
    "net/http"
    "sync"

    "github.com/gorilla/websocket"
    "golang.org/x/exp/slog"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow all connections for now 🦍
    },
}

type Server struct {
    // Mutex to protect connections
    sync.RWMutex
    // Connections store
    connections map[*websocket.Conn]bool
}

func NewServer() *Server {
    return &Server{
        connections: make(map[*websocket.Conn]bool),
    }
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        slog.Error("Failed to upgrade connection", "error", err)
        return
    }

    // Add connection to our map
    s.Lock()
    s.connections[conn] = true
    s.Unlock()

    slog.Info("New connection established", "addr", conn.RemoteAddr())

    // Remove connection when function returns
    defer func() {
        s.Lock()
        delete(s.connections, conn)
        s.Unlock()
        conn.Close()
        slog.Info("Connection closed", "addr", conn.RemoteAddr())
    }()

    // Keep connection alive
    for {
        // Read message (required to detect disconnection)
        if _, _, err := conn.ReadMessage(); err != nil {
            break
        }
    }
}

func main() {
    server := NewServer()

    http.HandleFunc("/ws", server.handleWS)

    slog.Info("🦍 STRONK SERVER STARTING ON PORT 42069 🦍")
    if err := http.ListenAndServe(":42069", nil); err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
