package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
)

// Configuration constants
const (
	// Server port to listen on
	serverPort = ":8080"

	// WebSocket endpoint path
	websocketEndpoint = "/connect"
)

// main is the entry point of the application
// It starts the Solana event listener in a goroutine and then starts the HTTP server
func main() {
	fmt.Println("Starting Nova Frontend Trial Task...")

	// Start the Solana event listener in background
	go listenToNewPairs()

	// Start the HTTP server (this will block until server stops)
	startServer()
}

// startServer initializes and starts the HTTP server with WebSocket support
// It sets up routing and handles graceful shutdown
func startServer() {
	// Create a new router with strict slash handling
	handler := mux.NewRouter().StrictSlash(true)

	// Register the WebSocket handler
	handler.HandleFunc(websocketEndpoint, HandleWebSocket)

	// Create HTTP server configuration
	server := &http.Server{
		Addr:    serverPort,
		Handler: handler,
	}

	fmt.Printf("Server starting on port %s\n", serverPort)
	fmt.Printf("WebSocket endpoint available at %s%s\n", serverPort, websocketEndpoint)

	// Start the server in a goroutine to allow for graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Server error: %v\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	waitForShutdown(server)
}

// waitForShutdown waits for OS signals and gracefully shuts down the server
func waitForShutdown(server *http.Server) {
	// Create a channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)

	// Register signals to listen for
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for signal
	sig := <-sigChan
	fmt.Printf("\nReceived signal %v, shutting down gracefully...\n", sig)

	// Attempt graceful shutdown
	if err := server.Shutdown(nil); err != nil {
		log.Printf("Error during server shutdown: %v\n", err)
	}

	fmt.Println("Server stopped")
}
