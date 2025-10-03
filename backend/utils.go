package main

import (
	"bytes"
	"fmt"

	bin "github.com/gagliardetto/binary"
)

// DecodeBorsh decodes binary data using the Borsh serialization format
// into the provided destination interface
//
// Parameters:
//   - dst: pointer to the destination struct that will hold the decoded data
//   - data: raw binary data to decode
//
// Returns:
//   - error: any error that occurred during decoding
func DecodeBorsh(dst interface{}, data []byte) error {
	return bin.NewBorshDecoder(data).Decode(dst)
}

// DecodeBase64 is a generic function that decodes base64-encoded data
// and extracts events based on a discriminator pattern
//
// The function performs the following steps:
// 1. Validates that the data is long enough to contain the discriminator
// 2. Checks if the discriminator matches the expected pattern
// 3. Extracts the payload data (everything after the discriminator)
// 4. Decodes the payload using Borsh deserialization
//
// Type Parameters:
//   - T: the type of event to decode (must be a struct)
//
// Parameters:
//   - data: base64-decoded binary data containing the event
//   - discriminator: byte sequence that identifies the event type
//
// Returns:
//   - *T: pointer to the decoded event struct
//   - error: any error that occurred during validation or decoding
func DecodeBase64[T any](data []byte, discriminator []byte) (*T, error) {
	// Validate that data is long enough to contain the discriminator
	if len(data) < len(discriminator) {
		return nil, fmt.Errorf("data too short for discriminator: expected %d bytes, got %d",
			len(discriminator), len(data))
	}

	// Check if the discriminator matches the expected pattern
	if !bytes.Equal(data[:len(discriminator)], discriminator) {
		return nil, fmt.Errorf("invalid discriminator: expected %v, got %v",
			discriminator, data[:len(discriminator)])
	}

	// Extract the payload data (everything after the discriminator)
	payload := data[len(discriminator):]

	// Create a new instance of the target type
	var event T

	// Decode the payload using Borsh deserialization
	err := DecodeBorsh(&event, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Borsh payload: %w", err)
	}

	return &event, nil
}
