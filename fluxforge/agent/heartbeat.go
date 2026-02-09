package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	heartbeatInterval = 5 * time.Second
)

// sendRegistration registers the agent with the control plane.
// It returns an error if registration fails.
func sendRegistration(cfg *Config) error {
	payload := map[string]interface{}{
		"node_id":  cfg.NodeID,
		"hostname": cfg.Hostname,
		"os":       cfg.OS,
		"arch":     cfg.Arch,
		"version":  cfg.Version,
		"address":  cfg.Address,
		"port":     cfg.Port, // IMPORTANT: int, not string
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal registration payload: %w", err)
	}

	resp, err := http.Post(
		cfg.ServerURL+"/agent/register",
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return fmt.Errorf("registration request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registration failed with status code: %d", resp.StatusCode)
	}

	log.Printf("Successfully registered agent: %s", cfg.NodeID)
	return nil
}

// sendHeartbeat sends a heartbeat to the control plane.
func sendHeartbeat(cfg *Config) {
	payload := map[string]interface{}{
		"node_id":   cfg.NodeID,
		"timestamp": time.Now().Unix(),
		"status":    "alive",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling heartbeat: %v", err)
		return
	}

	resp, err := http.Post(
		cfg.ServerURL+"/agent/heartbeat",
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		log.Printf("Error sending heartbeat: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Heartbeat failed with status: %d", resp.StatusCode)
	} else {
		log.Printf("Heartbeat sent successfully")
	}
}

// startHeartbeatLoop starts the heartbeat loop.
// It runs until the context is cancelled.
func startHeartbeatLoop(ctx context.Context, cfg *Config) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sendHeartbeat(cfg)
		case <-ctx.Done():
			log.Println("Heartbeat loop stopping...")
			return
		}
	}
}
