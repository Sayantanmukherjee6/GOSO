package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Client struct for each WebSocket connection
type Client struct {
	conn *websocket.Conn
	room string
	send chan []byte
}

// Hub to manage clients
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

// Create a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Start the Hub
func (h *Hub) Start() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

// Read messages from the WebSocket
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
		hub.broadcast <- message
	}
}

// Write messages to the WebSocket
func (c *Client) Write() {
	defer c.conn.Close()
	for message := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
}

func handleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("roomID")

	// roomID := r.URL.Query().Get("rid")
	// fmt.Println(roomID)
	// if roomID == "" {
	// 	http.Error(w, "Room ID is required", http.StatusBadRequest)
	// 	return
	// }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	client := &Client{conn: conn, room: roomID, send: make(chan []byte, 256)}
	hub.register <- client

	go client.Read(hub)
	go client.Write()
}

func main() {
	hub := NewHub()
	go hub.Start()

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		renderStart(w, "index.html")
	})
	http.HandleFunc("/start-chat", initProcess)
	http.HandleFunc("/join", renderJoin)
	http.HandleFunc("/join-chat", initJoin)
	http.HandleFunc("/room", renderRoom)
	// WebSocket endpoint
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(hub, w, r)
	})

	log.Println("server is listening on :8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		fmt.Println(err)
	}
}

// Generate a random ID
func generateRandomID() string {
	bytes := make([]byte, 16) // 16 bytes = 128 bits
	if _, err := rand.Read(bytes); err != nil {
		log.Fatal("Error generating random ID:", err)
	}
	return hex.EncodeToString(bytes)
}

func renderStart(w http.ResponseWriter, tmpl string) {
	// Parse the template
	tmplPath := "templates/" + tmpl
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the template
	err = t.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}

func initProcess(w http.ResponseWriter, r *http.Request) {
	// Ensure it's a POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the "username" field from the form data
	userName := r.FormValue("uname")

	if len(userName) != 0 {
		roomId := generateRandomID()
		http.Redirect(w, r, "/room?rid="+roomId+"&uname="+userName, http.StatusSeeOther)

	} else {
		http.Error(w, "Username can't be empty", http.StatusBadRequest)
	}
}

func renderJoin(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	roomId := r.URL.Query().Get("rid")
	// Parse the HTML template from the "templates" folder
	tmpl, err := template.ParseFiles("templates/join.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	//Pass query parameters to the template
	data := struct {
		RoomId string
	}{
		RoomId: roomId,
	}

	// Render the template with the data
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func initJoin(w http.ResponseWriter, r *http.Request) {
	// Ensure it's a POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the "username" field from the form data
	userName := r.FormValue("uname")
	roomId := r.FormValue("rid")

	if len(userName) != 0 {
		http.Redirect(w, r, "/room?rid="+roomId+"&uname="+userName, http.StatusSeeOther)

	} else {
		http.Error(w, "Username can't be empty", http.StatusBadRequest)
	}
}

func renderRoom(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	roomId := r.URL.Query().Get("rid")
	userName := r.URL.Query().Get("uname")

	if len(userName) == 0 {
		http.Redirect(w, r, "/join?rid="+roomId, http.StatusSeeOther)
	}

	if roomId == "" {
		http.Error(w, "No Room", http.StatusBadRequest)
		return
	}
	// Parse the HTML template from the "templates" folder
	tmpl, err := template.ParseFiles("templates/room.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	// Pass query parameters to the template
	data := struct {
		Roomlink string
		RoomId   string
		Username string
	}{
		Roomlink: r.Host + r.URL.Path + "?rid=" + roomId,
		RoomId:   roomId,
		Username: userName,
	}

	// Render the template with the data
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

// func renderRoom(w http.ResponseWriter, r *http.Request) {
// 	// Parse query parameters
// 	roomId := "abcd"
// 	userName := r.URL.Query().Get("uname")
// 	fmt.Println(userName)

// 	if len(userName) != 0 {
// 		// Parse the HTML template from the "templates" folder
// 		tmpl, err := template.ParseFiles("templates/room.html")
// 		if err != nil {
// 			http.Error(w, "Error loading template", http.StatusInternalServerError)
// 			return
// 		}

// 		// Pass query parameters to the template
// 		data := struct {
// 			RoomId   string
// 			Username string
// 		}{
// 			RoomId:   roomId,
// 			Username: userName,
// 		}

// 		// Render the template with the data
// 		err = tmpl.Execute(w, data)
// 		if err != nil {
// 			http.Error(w, "Error rendering template", http.StatusInternalServerError)
// 			return
// 		}
// 	} else {
// 		http.Error(w, "Username can't be empty", http.StatusBadRequest)
// 	}
// }
