package websockets

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

/*
Client Overview:
--------------
The Client handles individual WebSocket connections.
Each connected user has a Client instance that:
1. Manages their WebSocket connection
2. Handles reading messages from the user
3. Handles sending messages to the user
4. Maintains connection health with ping/pong
5. Cleans up on disconnection

Key Features:
- Concurrent message handling (read/write)
- Connection health monitoring
- Automatic cleanup on disconnect
- Buffer management for messages
*/

// Connection management constants
// These values are crucial for production applications
const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period
	// Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// Client represents a connected websocket user
type Client struct {
	hub      *Hub            // Reference to central hub for broadcasting
	conn     *websocket.Conn // Underlying WebSocket connection
	send     chan []byte     // Buffered channel for outbound messages
	room     string         // Current room name
	username string         // User's display name
}

// readPump handles incoming messages from the WebSocket connection
// This is a long-running goroutine that must be started for each client
func (c *Client) readPump() {
	// Cleanup on function exit
	// This ensures resources are freed when connection ends
	defer func() {
		// Notify hub that client is disconnecting
		c.hub.unregister <- c
		// Close the physical connection
		c.conn.Close()
	}()

	// Configure connection constraints
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		// Reset deadline when pong is received
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Main read loop
	for {
		// ReadMessage is a low-level method to read a message
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			// Check if it's an expected closure
			if websocket.IsUnexpectedCloseError(err, 
				websocket.CloseGoingAway, 
				websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break // Exit loop on any error
		}

		// Create message with metadata
		msg := Message{
			Type:     "chat",
			Content:  string(message),  // Change: Convert bytes to string
			RoomName: c.room,
			Username: c.username,
		}

		// Forward message to hub for broadcasting
		c.hub.broadcast <- msg
	}
}

// writePump handles sending messages to the WebSocket connection
// This is a long-running goroutine that must be started for each client
func (c *Client) writePump() {
	// Create ticker for periodic pings
	// This maintains connection health
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			// Set write deadline for each message
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Channel closed by hub
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Get the next writer for the connection
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			// Write the message
			w.Write(message)

			// Close the writer
			w.Close()

		case <-ticker.C:
			// Send periodic ping
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}