package main

import (
    "net/http"
    "os"
    "sync"
    "time"

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
    // Add connection count for metrics
    connectionCount int
}

func NewServer() *Server {
    return &Server{
        connections: make(map[*websocket.Conn]bool),
    }
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
    // Log incoming connection attempt
    slog.Info("Incoming WebSocket connection attempt",
        "remote_addr", r.RemoteAddr,
        "user_agent", r.UserAgent(),
        "timestamp", time.Now().Format(time.RFC3339))

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        slog.Error("Failed to upgrade connection",
            "error", err,
            "remote_addr", r.RemoteAddr,
            "timestamp", time.Now().Format(time.RFC3339))
        return
    }

    // Add connection to our map
    s.Lock()
    s.connections[conn] = true
    s.connectionCount++
    currentCount := s.connectionCount
    s.Unlock()

    slog.Info("New connection established",
        "addr", conn.RemoteAddr(),
        "total_connections", currentCount,
        "timestamp", time.Now().Format(time.RFC3339))

    // Remove connection when function returns
    defer func() {
        s.Lock()
        delete(s.connections, conn)
        s.connectionCount--
        currentCount := s.connectionCount
        s.Unlock()
        conn.Close()
        slog.Info("Connection closed",
            "addr", conn.RemoteAddr(),
            "remaining_connections", currentCount,
            "timestamp", time.Now().Format(time.RFC3339))
    }()

    // Keep connection alive
    for {
        // Read message (required to detect disconnection)
        if _, _, err := conn.ReadMessage(); err != nil {
            slog.Debug("Connection read error",
                "error", err,
                "addr", conn.RemoteAddr(),
                "timestamp", time.Now().Format(time.RFC3339))
            break
        }
    }
}

func main() {
    // Setup JSON logger with timestamp
    logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelDebug,
        AddSource: true,
    })
    logger := slog.New(logHandler)
    slog.SetDefault(logger)

    server := NewServer()

    // Log server configuration
    slog.Info("🦍 STRONK SERVER CONFIGURATION 🦍",
        "port", 42069,
        "timestamp", time.Now().Format(time.RFC3339),
        "version", "1.0.0",
        "log_level", "debug")

    // Serve static files from /app/src directory
    fs := http.FileServer(http.Dir("/app/src"))
    http.Handle("/", http.StripPrefix("/", fs))

    // Handle WebSocket connections
    http.HandleFunc("/ws", server.handleWS)

    slog.Info("🦍 STRONK SERVER STARTING ON PORT 42069 🦍",
        "timestamp", time.Now().Format(time.RFC3339))
    if err := http.ListenAndServe(":42069", nil); err != nil {
        slog.Error("Server failed to start",
            "error", err,
            "timestamp", time.Now().Format(time.RFC3339))
        os.Exit(1)
    }
}
