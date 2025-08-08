package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	go func() {
		if err := startServer(ctx); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	sig := <-sigCh
	log.Printf("Received signal %v, shutting down...", sig)
	cancel()
}
