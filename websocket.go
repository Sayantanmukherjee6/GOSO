package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client/user in a specific room
type Client struct {
	conn *websocket.Conn
	room string
	send chan []byte
}

// Hub manages the WebSocket connections and rooms
type Hub struct {
	rooms     map[string]map[*Client]bool
	broadcast chan struct {
		roomID  string
		message []byte
	}
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]map[*Client]bool),
		broadcast: make(chan struct {
			roomID  string
			message []byte
		}),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Start runs the Hub to manage WebSocket events
func (h *Hub) Start() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.rooms[client.room] == nil {
				h.rooms[client.room] = make(map[*Client]bool)
			}
			h.rooms[client.room][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.rooms[client.room][client]; ok {
				delete(h.rooms[client.room], client)
				close(client.send)
				if len(h.rooms[client.room]) == 0 {
					delete(h.rooms, client.room)
				}
			}
			h.mu.Unlock()

		case data := <-h.broadcast:
			h.mu.Lock()
			if clients, ok := h.rooms[data.roomID]; ok {
				for client := range clients {
					select {
					case client.send <- data.message:
					default:
						close(client.send)
						delete(clients, client)
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

// Client methods for WebSocket read
func (c *Client) Read(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		hub.broadcast <- struct {
			roomID  string
			message []byte
		}{roomID: c.room, message: message}
	}
}

// Client methods for WebSocket write
func (c *Client) Write() {
	defer c.conn.Close()
	for message := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("roomID")
	if roomID == "" {
		http.Error(w, "Room ID is required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}

	client := &Client{conn: conn, room: roomID, send: make(chan []byte, 256)}
	hub.register <- client

	go client.Read(hub)
	go client.Write()
}
