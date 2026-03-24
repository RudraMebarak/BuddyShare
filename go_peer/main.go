package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
)

// SignalMessage defines the JSON structure we send to Node.js
type SignalMessage struct {
	Type   string `json:"type"`
	RoomID string `json:"roomId,omitempty"`
	Data   string `json:"data,omitempty"` // We will put WebRTC SDPs here later
}

func main() {
	// Let the user specify a room name via command line flag
	roomID := flag.String("room", "test-room", "The ID of the room to join")
	flag.Parse()

	log.Printf("Starting Peer Client. Attempting to join room: %s", *roomID)

	// 1. Define the Matchmaker URL
	u := url.URL{Scheme: "ws", Host: "localhost:3000", Path: "/"}
	
	// 2. Connect to the WebSocket Server
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Dial error (Is Node.js running?): ", err)
	}
	defer conn.Close()
	log.Println("✅ Connected to Matchmaker")

	// 3. Create our "join-room" message
	joinMsg := SignalMessage{
		Type:   "join-room",
		RoomID: *roomID,
	}

	// Convert Go struct to JSON bytes and send it
	msgBytes, _ := json.Marshal(joinMsg)
	err = conn.WriteMessage(websocket.TextMessage, msgBytes)
	if err != nil {
		log.Fatal("Failed to send join-room message:", err)
	}
	log.Println("🏠 Join room request sent.")

	// 4. Keep the program running until you press Ctrl+C
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	log.Println("Shutting down...")
}