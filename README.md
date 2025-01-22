# Go WebSocket Chat

A room-based chat application built with Go and WebSockets. Based on the hub pattern for efficient real-time communication.

[Read the blog i have written for this app](https://rohan-bhujbal.hashnode.dev/websockets-explained-golang-a-practical-guide-to-real-time-apps)

## Quick Start

```bash
# Get the code
git clone https://github.com/yourusername/go-websocket-chat
cd go-websocket-chat

# Install deps
go mod download

# Run server
go run main.go
```

## Test It Out

Install wscat:
```bash
npm install -g wscat
```

Open multiple terminals and connect:
```bash
# Terminal 1
wscat -c ws://localhost:8080/ws/room1

# Terminal 2 
wscat -c ws://localhost:8080/ws/room1

# Terminal 3
wscat -c ws://localhost:8080/ws/room2
```

Start chatting! Messages only go to users in the same room.

## Structure

```
├── main.go           # Server setup
├── websockets/
│   ├── hub.go       # Connection manager  
│   ├── client.go    # Client handler
│   └── websocket.go # WS upgrader
```

## Troubleshooting

- **Can't connect**: Check server is running on 8080
- **Messages not working**: Verify same room name
- **wscat not found**: Run `npm install -g wscat`

## Built With

- [Gin](https://gin-gonic.com/)
- [Gorilla WebSocket](https://github.com/gorilla/websocket)

## License

MIT
