# 🚀 Go-WebRTC P2P File Share

A blazing-fast, command-line peer-to-peer file sharing utility. Built as a 6-hour MVP sprint to explore WebRTC Data Channels and Go concurrency.

This project allows two computers to transfer files directly to each other over the internet without routing the file data through a central server.

## 🏗 Architecture

This project consists of two main components:
1.  **The Matchmaker (Node.js Signaling Server):** A lightweight Socket.io server that allows peers to discover each other and exchange WebRTC connection data (SDP offers/answers and ICE candidates).
2.  **The Engine (Go WebRTC Client):** A CLI tool built with [Pion WebRTC](https://github.com/pion/webrtc) that handles reading/writing files, chunking data, and streaming it directly to the connected peer via WebRTC Data Channels.

## ✨ Features

* **True P2P Transfer:** Files go directly from Peer A to Peer B.
* **Memory Efficient:** Uses chunking and backpressure to stream massive files without crashing your RAM.
* **NAT Traversal:** Utilizes Google's public STUN servers to connect across different local networks.
* **Cross-Platform:** The Go client compiles to a single, dependency-free binary for Windows, Mac, or Linux.

## 🛠 Prerequisites

* [Go 1.26+](https://go.dev/doc/install)
* [Node.js](https://nodejs.org/) & npm

## 🚀 Quick Start

### 1. Start the Signaling Server
Open a terminal and navigate to the signaling server directory:
```bash
cd signaling-server
npm install
node server.js