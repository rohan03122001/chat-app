package websockets

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

/*
WebSocket Handler Overview:
-------------------------
This file handles new WebSocket connections and their initial setup.
Main responsibilities:
1. Validate incoming connection requests
2. Upgrade HTTP connections to WebSocket
3. Create new client instances
4. Register clients with the hub

Connection Flow:
1. Client connects to /ws/:room?username=xxx
2. Validate room name and username
3. Upgrade to WebSocket connection
4. Create new client
5. Start message handling
*/

// upgrader converts HTTP connections to WebSocket connections
var upgrader = websocket.Upgrader{
	// Buffer sizes affect memory usage and performance
	ReadBufferSize:  1024,  // Adjust based on expected message sizes
	WriteBufferSize: 1024,

	// CheckOrigin prevents unauthorized cross-origin requests
	// WARNING: Current implementation allows all origins - NOT SAFE FOR PRODUCTION
	CheckOrigin: func(r *http.Request) bool {
		return true // Development only - accepts all origins
	},
}

// HandleWebSocket creates a WebSocket handler function for Gin
// This is where new WebSocket connections are established
func HandleWebSocket(h *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Step 1: Extract and validate connection parameters
		room := c.Param("room")
		username := c.Query("username")

		// Validate required fields
		if room == "" || username == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "room and username are required"})
			return
		}

		// Step 2: Upgrade HTTP connection to WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v", err)
			return
		}

		// Step 3: Create new client instance
		client := &Client{
			hub:      h,
			conn:     conn,
			send:     make(chan []byte, 256), // Buffer size affects memory usage
			room:     room,
			username: username,
		}

		// Step 4: Register client with hub
		// This also triggers the "user joined" notification
		h.register <- client

		// Create and broadcast join message
		joinMessage := Message{
			Type:     "user_joined",
			Content:  username + " joined the room",
			RoomName: room,
			Username: username,
		}
		h.broadcast <- joinMessage

		// Step 5: Start client read/write pumps
		// These goroutines handle the ongoing communication
		go client.writePump() // Handles sending messages to the client
		go client.readPump()  // Handles receiving messages from the client
	}
}