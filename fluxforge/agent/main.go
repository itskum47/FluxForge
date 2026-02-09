package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Initialize configuration
	cfg := LoadConfig()
	log.Printf("Agent starting. Node ID: %s", cfg.NodeID)

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	// Registration loop with backoff
	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second

	for {
		if ctx.Err() != nil {
			return // Context cancelled, exit
		}

		err := sendRegistration(cfg)
		if err == nil {
			break
		}

		log.Printf("Registration failed: %v. Retrying in %s...", err, backoff)
		
		select {
		case <-time.After(backoff):
			// Continue loop
		case <-ctx.Done():
			return // Exit immediately
		}

		// Exponential backoff with cap
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}

	// Start heartbeat loop
	// Run in a goroutine so we can block on ctx.Done() in main
	go startHeartbeatLoop(ctx, cfg)

	// Start HTTP Server
	executor := NewExecutor(cfg)
	server := NewServer(cfg, executor)
	go func() {
		if err := server.Start(); err != nil {
			log.Printf("HTTP Server failed: %v", err)
			cancel() // Stop agent if server fails
		}
	}()

	// Block until shutdown
	<-ctx.Done()
	log.Println("Agent shutting down.")
}
