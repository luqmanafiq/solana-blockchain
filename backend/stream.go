package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

// Configuration constants
const (
	// WebSocket URL for Helius RPC endpoint
	websocketURL = "wss://mainnet.helius-rpc.com/?api-key=0f803376-0189-4d72-95f6-a5f41cef157d"

	// PumpFun program address on Solana mainnet
	pumpFunProgram = "TSLvdd1pWpHVjahSpsvCXUbgwsL3JAcvokwaKt1eokM"

	// Reconnection delay when connection fails
	reconnectDelay = 2 * time.Second

	// Magic string to identify relevant logs
	logIdentifier = "G3KpTd7r"

	// Program data prefix in log messages
	programDataPrefix = "Program data: "
)

// RawEvent represents the decoded event data from Solana program logs
type RawEvent struct {
	Name   string           // Token name
	Symbol string           // Token symbol
	Uri    string           // Token metadata URI
	Mint   solana.PublicKey // Token mint address
}

// CreateEvent represents the formatted event data sent to clients
type CreateEvent struct {
	Name   string `json:"name"`   // Token name
	Symbol string `json:"symbol"` // Token symbol
	Uri    string `json:"uri"`    // Token metadata URI
	Mint   string `json:"mint"`   // Token mint address as string
}

// creationDiscriminator is the byte sequence that identifies creation events
var creationDiscriminator = []byte{27, 114, 169, 77, 222, 235, 99, 118}

// listenToNewPairs establishes a WebSocket connection to listen for new token pair creations
// on the PumpFun program. It handles reconnections automatically and processes incoming
// program logs to extract creation events.
func listenToNewPairs() {
	fmt.Println("Starting to listen for new token pairs...")

	for {
		if err := connectAndListen(); err != nil {
			fmt.Printf("Connection error: %v\n", err)
			fmt.Printf("Reconnecting in %v...\n", reconnectDelay)
			time.Sleep(reconnectDelay)
		}
	}
}

// connectAndListen establishes a WebSocket connection and listens for program logs
func connectAndListen() error {
	// Establish WebSocket connection
	socket, err := ws.Connect(context.Background(), websocketURL)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}
	defer socket.Close()

	fmt.Println("Successfully connected to WebSocket")

	// Subscribe to logs mentioning the PumpFun program
	sub, err := socket.LogsSubscribeMentions(
		solana.MPK(pumpFunProgram),
		rpc.CommitmentProcessed,
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to logs: %w", err)
	}

	fmt.Println("Subscribed to PumpFun program logs")

	// Listen for incoming messages
	return listenForMessages(sub)
}

// listenForMessages processes incoming WebSocket messages and extracts creation events
func listenForMessages(sub *ws.LogSubscription) error {
	for {
		message, err := sub.Recv(context.Background())
		if err != nil {
			return fmt.Errorf("error receiving message: %w", err)
		}

		// Process each log in the message
		for _, log := range message.Value.Logs {
			if err := processLog(log); err != nil {
				// Log error but continue processing other logs
				fmt.Printf("Error processing log: %v\n", err)
			}
		}
	}
}

// processLog processes a single log entry and extracts creation events
func processLog(log string) error {
	// Check if log contains the identifier for relevant events
	if !strings.Contains(log, logIdentifier) {
		return nil // Not a relevant log, skip
	}

	// Extract program data from log
	data, err := extractProgramData(log)
	if err != nil {
		return fmt.Errorf("failed to extract program data: %w", err)
	}

	// Decode base64 data
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return fmt.Errorf("failed to decode base64 data: %w", err)
	}

	// Check if this is a creation event
	if !bytes.HasPrefix(decoded, creationDiscriminator) {
		return nil // Not a creation event, skip
	}

	// Decode the event data
	event, err := DecodeBase64[RawEvent](decoded, creationDiscriminator)
	if err != nil {
		return fmt.Errorf("failed to decode event: %w", err)
	}

	// Create formatted event for clients
	createEvent := CreateEvent{
		Name:   event.Name,
		Symbol: event.Symbol,
		Uri:    event.Uri,
		Mint:   event.Mint.String(),
	}

	// Marshal to JSON and send to clients
	marshalled, err := json.Marshal(createEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	fmt.Printf("New token created: %s\n", string(marshalled))

	// Send to all connected clients asynchronously
	go sendMessageToAllClients(marshalled)

	return nil
}

// extractProgramData extracts the program data portion from a log message
func extractProgramData(log string) (string, error) {
	parts := strings.Split(log, programDataPrefix)
	if len(parts) < 2 {
		return "", fmt.Errorf("log does not contain program data")
	}
	return parts[1], nil
}
