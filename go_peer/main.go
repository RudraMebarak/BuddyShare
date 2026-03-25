package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

// SignalMessage defines the JSON structure we send to Node.js
type SignalMessage struct {
	Type   string `json:"type"`             // "join-room" or "signal"
	RoomID string `json:"roomId,omitempty"` // The room to target
	Data   string `json:"data,omitempty"`   // WebRTC connection data (SDP or ICE)
}

func main() {
	roomID := flag.String("room", "test-room", "The ID of the room to join")
	isSender := flag.Bool("sender", false, "Set to true if this peer is sending the file")
	fileName := flag.String("file", "", "The path of the file to send (only for sender)")
	flag.Parse()

	// 1. Setup WebSocket to Node.js Matchmaker
	u := url.URL{Scheme: "ws", Host: "localhost:3000", Path: "/"}
	wsConn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Dial error: ", err)
	}
	defer wsConn.Close()
	log.Println("✅ Connected to Matchmaker")

	// Join the room
	joinMsg := SignalMessage{Type: "join-room", RoomID: *roomID}
	wsConn.WriteJSON(joinMsg)
	log.Printf("🏠 Joined room: %s\n", *roomID)

	// 2. Setup WebRTC PeerConnection
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}}, // Google's free STUN server to find our IP
		},
	}
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		log.Fatal(err)
	}
	defer peerConnection.Close()

	// 3. Handle ICE Candidates (Pathfinding)
	// When Pion finds a network path, send it to the other peer via the Matchmaker
	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		candidateJSON, _ := json.Marshal(c.ToJSON())
		wsConn.WriteJSON(SignalMessage{
			Type:   "signal",
			RoomID: *roomID,
			Data:   string(candidateJSON),
		})
	})

	// 4. Setup the Data Channel (The Pipe)
	if *isSender {
		// Sender CREATES the data channel
		dataChannel, err := peerConnection.CreateDataChannel("file-transfer", nil)
		if err != nil {
			log.Fatal(err)
		}
		dataChannel.OnOpen(func() {
			fmt.Println("🚀 DATA CHANNEL OPEN! Starting file transfer...")

			// 1. Open the file you want to send
			// (You should add a --file flag to your main() function for this!)
			filePath := *fileName
			file, err := os.Open(filePath)
			if err != nil {
				log.Fatal("Could not open file:", err)
			}
			defer file.Close()

			// 2. Create a buffer (chunk size). 16KB is the WebRTC sweet spot.
			bufferSize := 16384 // 16KB
			buffer := make([]byte, bufferSize)

			// 3. Loop through the file and send chunks
			for {
				n, err := file.Read(buffer)
				if err != nil {
					if err.Error() == "EOF" {
						// End of file reached! Tell the receiver we are done.
						dataChannel.SendText("EOF")
						fmt.Println("✅ File sent completely!")
						break
					}
					log.Fatal("Error reading file:", err)
				}

				// 4. THE BACKPRESSURE CHECK
				// If we send faster than the network can handle, the buffer fills up.
				// If it gets over 1MB, we pause for a millisecond to let it drain.
				for dataChannel.BufferedAmount() > 1024*1024 {
					// Wait for network to catch up
					time.Sleep(time.Millisecond)
				}

				// Send the actual chunk of bytes
				err = dataChannel.Send(buffer[:n])
				if err != nil {
					log.Fatal("Error sending data:", err)
				}
			}
		})

		// Create the SDP Offer and send it
		offer, _ := peerConnection.CreateOffer(nil)
		peerConnection.SetLocalDescription(offer)
		offerJSON, _ := json.Marshal(offer)

		wsConn.WriteJSON(SignalMessage{
			Type:   "signal",
			RoomID: *roomID,
			Data:   string(offerJSON),
		})
		log.Println("📤 Sent WebRTC Offer")

	} else {
		// Receiver LISTENS for the data channel
		outFile, err := os.Create("received_file.mp4")
		if err != nil {
			log.Fatal(err)
		}
		peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
			d.OnOpen(func() {
				fmt.Println("🚀 DATA CHANNEL OPEN! Ready to receive files.")
			})
			d.OnMessage(func(msg webrtc.DataChannelMessage) {

				// 3. Check if it's the "End Of File" signal
				if string(msg.Data) == "EOF" {
					fmt.Println("✅ File received completely!")
					outFile.Close() // Safely close the file
					os.Exit(0)      // Exit the program cleanly
					return
				}

				// 4. Otherwise, it's file data. Write it to the disk!
				_, err := outFile.Write(msg.Data)
				if err != nil {
					log.Fatal("Error writing to file:", err)
				}
			})
		})
	}

	// 5. Listen for incoming signals from the Matchmaker
	go func() {
		for {
			var msg SignalMessage
			err := wsConn.ReadJSON(&msg)
			if err != nil {
				log.Println("WebSocket closed:", err)
				return
			}

			if msg.Type == "signal" {
				// We received WebRTC data. Is it an SDP description or an ICE candidate?
				if len(msg.Data) > 0 && msg.Data[2:6] == "type" {
					// It's an SDP Offer/Answer
					var sdp webrtc.SessionDescription
					json.Unmarshal([]byte(msg.Data), &sdp)
					peerConnection.SetRemoteDescription(sdp)

					// If we received an offer, we must send an answer back
					if sdp.Type == webrtc.SDPTypeOffer {
						answer, _ := peerConnection.CreateAnswer(nil)
						peerConnection.SetLocalDescription(answer)
						answerJSON, _ := json.Marshal(answer)
						wsConn.WriteJSON(SignalMessage{
							Type:   "signal",
							RoomID: *roomID,
							Data:   string(answerJSON),
						})
						log.Println("📤 Sent WebRTC Answer")
					}
				} else {
					// It's an ICE Candidate
					var candidate webrtc.ICECandidateInit
					json.Unmarshal([]byte(msg.Data), &candidate)
					peerConnection.AddICECandidate(candidate)
				}
			}
		}
	}()

	// Keep program running
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
}
