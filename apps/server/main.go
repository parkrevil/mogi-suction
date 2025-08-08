package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("Starting QUIC server...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		if err := startQUICServer(ctx); err != nil {
			log.Printf("Server error: %v", err)
			// 서버 에러 시 적절한 종료 처리
			return
		}
	}()

	<-sigCh
	log.Println("🛑 Shutting down...")
	cancel()

	<-serverDone
	log.Println("✅ Server shutdown complete")
}
