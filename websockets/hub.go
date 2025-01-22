package websockets

import (
	"encoding/json"
	"log"
	"strings"
)

/*
Hub Overview:
------------
The Hub manages all active WebSocket connections and handles message routing.
It's the central component that:
1. Tracks active clients in each chat room
2. Handles new client connections
3. Manages client disconnections
4. Routes messages between clients in the same room
5. Maintains the list of online users per room

Key Features:
- Room-based chat (multiple chat rooms)
- User join/leave notifications
- Online user tracking per room
- Message broadcasting to room members
*/

// Message defines the structure of all communications in the chat system
type Message struct {
	Type     string `json:"type"`     // Message types: chat, user_joined, user_left, online_users
	Content  string `json:"content"`   // The message content
	RoomName string `json:"room"`     // The room this message belongs to
	Username string `json:"username"`  // The sender's username
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	clients    map[*Client]bool                // All connected clients
	rooms      map[string]map[*Client]bool     // Room-based client groups
	broadcast  chan Message                    // Channel for inbound messages
	register   chan *Client                    // Channel for client registration
	unregister chan *Client                    // Channel for client disconnection
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)
		case client := <-h.unregister:
			h.handleUnregister(client)
		case message := <-h.broadcast:
			h.handleBroadcast(message)
		}
	}
}

func (h *Hub) handleRegister(client *Client) {
	// Create room if needed
	if _, exists := h.rooms[client.room]; !exists {
		h.rooms[client.room] = make(map[*Client]bool)
	}
	
	// Add client to room and global list
	h.rooms[client.room][client] = true
	h.clients[client] = true

	// Send online users list
	h.broadcastRoomUsers(client.room)
}

func (h *Hub) handleUnregister(client *Client) {
	if _, exists := h.clients[client]; !exists {
		return
	}

	// Remove client
	delete(h.clients, client)
	delete(h.rooms[client.room], client)

	// Notify room and update user list
	h.handleBroadcast(Message{
		Type:     "user_left",
		Content:  client.username + " left the room",
		RoomName: client.room,
		Username: client.username,
	})
	h.broadcastRoomUsers(client.room)

	// Clean up empty room
	if len(h.rooms[client.room]) == 0 {
		delete(h.rooms, client.room)
	}
}

func (h *Hub) broadcastRoomUsers(room string) {
	users := []string{}
	if roomClients, exists := h.rooms[room]; exists {
		for client := range roomClients {
			users = append(users, client.username)
		}
	}

	h.handleBroadcast(Message{
		Type:     "online_users",
		Content:  strings.Join(users, ","),
		RoomName: room,
	})
}

func (h *Hub) handleBroadcast(msg Message) {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	// Send to all clients in the room
	if roomClients, exists := h.rooms[msg.RoomName]; exists {
		for client := range roomClients {
			select {
			case client.send <- jsonMsg:
				// Message sent successfully
			default:
				// Client's buffer is full, remove them
				close(client.send)
				delete(h.clients, client)
				delete(h.rooms[msg.RoomName], client)
			}
		}
	}
}

