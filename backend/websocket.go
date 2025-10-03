package main

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/puzpuzpuz/xsync/v4"
)

// Configuration constants
const (
	// WebSocket buffer sizes for optimal performance
	readBufferSize  = 8576    // 8KB read buffer
	writeBufferSize = 1048576 // 1MB write buffer

	// Ping message identifier
	pingMessage = "ping"

	// Pong response message
	pongResponse = `{"message":"pong"}`
)

// Client represents a connected WebSocket client
// It contains the connection and a mutex for thread-safe operations
type Client struct {
	Connection *websocket.Conn
	Mutex      sync.Mutex
}

// ConnectedClients stores all currently connected WebSocket clients
// Uses a thread-safe map with client address as the key
var ConnectedClients = xsync.NewMap[string, *Client]()

// upgrader handles HTTP to WebSocket connection upgrades
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development (should be restricted in production)
		return true
	},
	EnableCompression: true,
	ReadBufferSize:    readBufferSize,
	WriteBufferSize:   writeBufferSize,
}

// HandleWebSocket handles incoming WebSocket connection requests
// It upgrades the HTTP connection to WebSocket and manages the client lifecycle
//
// Parameters:
//   - w: HTTP response writer
//   - r: HTTP request containing the WebSocket upgrade request
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// Handle the WebSocket connection
	handleConnection(conn)
}

// sendMessageToAllClients broadcasts a message to all connected WebSocket clients
// It creates a copy of the client list to avoid holding locks during iteration
//
// Parameters:
//   - message: the message to broadcast to all clients
func sendMessageToAllClients(message []byte) {
	// Create a slice to store client pointers (avoiding mutex copying)
	allClients := []*Client{}

	// Collect all connected clients
	ConnectedClients.Range(func(key string, client *Client) bool {
		allClients = append(allClients, client)
		return true
	})

	// Send message to each client asynchronously
	for _, client := range allClients {
		go func(c *Client) {
			c.Mutex.Lock()
			defer c.Mutex.Unlock()

			// Send the message to this client
			if err := c.Connection.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Failed to send message to client %s: %v", c.Connection.RemoteAddr(), err)
			}
		}(client)
	}
}

// handleConnection manages an individual WebSocket connection
// It handles incoming messages, ping/pong responses, and client lifecycle
//
// Parameters:
//   - conn: the WebSocket connection to manage
func handleConnection(conn *websocket.Conn) {
	// Get the client's remote address for identification
	address := conn.RemoteAddr().String()
	log.Printf("New WebSocket connection from: %s", address)

	// Create a new client instance
	client := &Client{
		Connection: conn,
		Mutex:      sync.Mutex{},
	}

	// Store the client in the connected clients map
	ConnectedClients.Store(address, client)
	log.Printf("Client %s added to connected clients", address)

	// Main message handling loop
	for {
		// Read incoming messages
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from client %s: %v", address, err)
			break
		}

		// Handle ping messages with pong responses
		if strings.Contains(string(message), pingMessage) {
			go func() {
				client.Mutex.Lock()
				defer client.Mutex.Unlock()

				// Send pong response
				if err := client.Connection.WriteMessage(websocket.TextMessage, []byte(pongResponse)); err != nil {
					log.Printf("Failed to send pong to client %s: %v", address, err)
				}
			}()
		}
	}

	// Clean up when connection is closed
	ConnectedClients.Delete(address)
	log.Printf("Client %s disconnected and removed from connected clients", address)
}
