const WebSocket = require('ws');
const wss = new WebSocket.Server({ port: 3000 });

// A simple map to keep track of which peers are in which rooms
const rooms = new Map();

wss.on('connection', (ws) => {
    console.log('🟢 New peer connected');

    ws.on('message', (messageAsString) => {
        const message = JSON.parse(messageAsString);

        // 1. Join Room Logic
        if (message.type === 'join-room') {
            const roomId = message.roomId;
            ws.roomId = roomId; // Tag this connection with the room ID
            
            if (!rooms.has(roomId)) rooms.set(roomId, new Set());
            rooms.get(roomId).add(ws);
            
            console.log(`🏠 Peer joined room: ${roomId}`);
        }

        // 2. Relay WebRTC Signals to others in the same room
        if (message.type === 'signal') {
            const room = rooms.get(ws.roomId);
            if (room) {
                room.forEach(client => {
                    // Send to everyone EXCEPT the sender
                    if (client !== ws && client.readyState === WebSocket.OPEN) {
                        client.send(JSON.stringify(message));
                    }
                });
            }
        }
    });

    ws.on('close', () => {
        console.log('🔴 Peer disconnected');
        if (ws.roomId && rooms.has(ws.roomId)) {
            rooms.get(ws.roomId).delete(ws);
        }
    });
});

console.log('🚀 Raw WebSocket Matchmaker running on ws://localhost:3000');