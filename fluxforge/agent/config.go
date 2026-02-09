package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Config holds the agent configuration and identity.
type Config struct {
	NodeID   string
	Hostname string
	OS       string
	Arch     string
	Version  string
	ServerURL string
	Port      int
	Address   string // Reachable address
}

// LoadConfig initializes the agent configuration.
// It loads or generates the NodeID and sets up system metadata.
func LoadConfig() *Config {
	nodeID, err := getOrCreateNodeID()
	if err != nil {
		log.Fatalf("Failed to initialize Node ID: %v", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("Warning: could not get hostname: %v", err)
		hostname = "unknown"
	}

	return &Config{
		NodeID:    nodeID,
		Hostname:  hostname,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Version:   "0.1.0",
		ServerURL: "http://localhost:8080",
		Port:      8081,
		Address:   hostname, // Using hostname as reachable address for Phase 2
	}
}

// getOrCreateNodeID retrieves the existing Node ID or generates a new one.
// It persists the ID to ~/.fluxforge/node_id.
func getOrCreateNodeID() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".fluxforge")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory %s: %w", configDir, err)
	}

	nodeIDPath := filepath.Join(configDir, "node_id")

	// Try reading existing ID
	data, err := os.ReadFile(nodeIDPath)
	if err == nil {
		id := strings.TrimSpace(string(data))
		if id != "" {
			return id, nil
		}
	}

	// Generate new ID
	newID := generateUUID()
	if err := os.WriteFile(nodeIDPath, []byte(newID), 0600); err != nil {
		return "", fmt.Errorf("failed to save node ID to %s: %w", nodeIDPath, err)
	}

	return newID, nil
}

// generateUUID generates a random UUID-like string.
func generateUUID() string {
	b := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		log.Fatalf("Failed to generate random UUID: %v", err)
	}
	// Variant bits; see section 4.1.1
	b[8] = b[8]&0x3f | 0x80
	// Version 4 (random)
	b[6] = b[6]&0x0f | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
