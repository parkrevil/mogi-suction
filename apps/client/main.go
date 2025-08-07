package main

import (
	"context"
	"fmt"
	"log"
	"mogi-suction/client/packet"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	fmt.Println("Starting packet capture with TCP reassembly (Goroutines)...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 3)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(sigCh)
		close(sigCh)
	}()

	var shutdownOnce sync.Once

	shutdown := func(reason string) {
		shutdownOnce.Do(func() {
			fmt.Printf("\n🛑 Graceful shutdown initiated... (%s)\n", reason)

			cancel()
		})
	}

	go func() {
		select {
		case sig := <-sigCh:
			switch sig {
			case syscall.SIGINT:
				shutdown("SIGINT (Ctrl+C)")
			case syscall.SIGTERM:
				shutdown("SIGTERM")
			default:
				shutdown(fmt.Sprintf("Signal: %v", sig))
			}
		case <-ctx.Done():
			return
		}
	}()

	err := packet.InitPacketSniffer(ctx)
	if err != nil {
		log.Fatal("Failed to initialize packet sniffer:", err)
	}
	defer packet.ClosePacketSniffer()

	packet.StartPacketSniffer()

	<-ctx.Done()

	packet.StopPacketSniffer()
}
