# 🚀 P2P File Drop (Go + WebRTC)

A lightning-fast, command-line peer-to-peer file sharing utility. Built as a rapid prototype to explore WebRTC Data Channels, NAT Traversal, and Go concurrency.

This project allows two computers to transfer files directly to each other over a local network or the internet without routing the actual file data through a central cloud server.

## 🏗 Architecture

This project consists of two distinct components:

1. **The Matchmaker (Node.js/WebSocket):** A lightweight signaling server. It acts as a switchboard, allowing peers to discover each other in "rooms" and exchange WebRTC connection data (SDP offers/answers and ICE candidates). It **never** touches the file data.
2. **The Engine (Go/Pion WebRTC):** The core CLI application. It handles NAT traversal (finding its own IP via STUN), establishes a direct peer-to-peer WebRTC Data Channel, reads files from the disk in chunks, and streams them safely using network backpressure.

## 🛠 Prerequisites

* [Go 1.21+](https://go.dev/doc/install)
* [Node.js](https://nodejs.org/) & npm

## 📁 Project Structure
p2p-project/
│
├── matchmaker/            # Node.js Signaling Server
│   ├── package.json
│   └── server.js          # Raw WebSocket server
│
└── go-client/             # Go WebRTC Engine
    ├── go.mod
    ├── go.sum
    └── main.go            # Core P2P and file chunking logic
    
🚀 Setup & Installation
1. Start the Matchmaker
Open a terminal, navigate to the matchmaker directory, and start the WebSocket server:

Bash
cd matchmaker
npm install ws
node server.js
(The server runs on ws://localhost:3000 by default).

2. Prepare the Go Client
Open a new terminal and navigate to the go-client directory:

Bash
cd go-client
go mod tidy
💻 Usage
To test the transfer, you need two terminal windows running the Go client (representing two different computers).

Scenario A: Testing on the Same Machine (Localhost)
Make sure your Go main.go file is pointing to ws://localhost:3000.

Terminal 1 (The Receiver):
Join a room and wait for the file.

Bash
go run main.go --room=secret-drop
Terminal 2 (The Sender):
Join the same room, trigger the WebRTC offer, and specify the file to send.

Bash
go run main.go --room=secret-drop --sender=true --file=./my-video.mp4
Scenario B: Testing Across a Local Wi-Fi Network
To send a file to a friend sitting on the same Wi-Fi network:

Find the Local IP Address of the computer running the Node.js Matchmaker (e.g., 192.168.1.50).

On both computers, update the WebSocket URL in main.go to point to that IP:
u := url.URL{Scheme: "ws", Host: "192.168.1.50:3000", Path: "/"}

Run the Receiver command on one laptop, and the Sender command on the other!

🗺 Future Roadmap
[x] Establish WebRTC Data Channel integration via Pion

[x] Build Node.js WebSocket signaling bridge

[ ] UI Polish: Add a terminal progress bar to track incoming file chunks.

[ ] TURN Server: Add TURN server credentials to support strict corporate/university firewalls.

[ ] Multi-Peer: Upgrade the logic to allow one sender to broadcast a file to multiple receivers in the same room.

Built while studying Go concurrency and WebRTC internals.
