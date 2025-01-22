package main

import (
	"chat-app/websockets"
	"log"

	"github.com/gin-gonic/gin"
)


func main() {
	// Initialize router and hub
	r := gin.Default()
	hub := websockets.NewHub()
	go hub.Run()

	// Set up routes
	r.GET("/ws/:room", websockets.HandleWebSocket(hub))
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Start server
	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}