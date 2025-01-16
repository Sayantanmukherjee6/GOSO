package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	hub := NewHub()
	go hub.Start()

	http.HandleFunc("/", renderStart)
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
