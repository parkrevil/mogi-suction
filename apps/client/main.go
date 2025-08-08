package main

import (
	"context"
	"log"
	"mogi-suction/client/packet"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	quicServerAddr = "localhost:8443"
	connectTimeout = 10 * time.Second
)

func main() {
	log.Println("Starting packet capture with TCP reassembly and QUIC client...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 시그널 처리
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	// QUIC 클라이언트 연결
	quicClient := NewQUICClient(quicServerAddr)
	defer quicClient.Close()

	// QUIC 연결 고루틴
	go func() {
		connectCtx, connectCancel := context.WithTimeout(ctx, connectTimeout)
		defer connectCancel()

		if err := quicClient.Connect(connectCtx); err != nil {
			log.Printf("Failed to connect to QUIC server: %v", err)
			return
		}

		log.Printf("✅ Connected to QUIC server")
		<-ctx.Done()
	}()

	// 패킷 스니퍼 초기화
	if err := packet.InitPacketSniffer(ctx); err != nil {
		log.Fatal("Failed to initialize packet sniffer:", err)
	}
	defer packet.ClosePacketSniffer()

	packet.StartPacketSniffer()

	// 시그널 대기 및 종료
	<-sigCh
	log.Println("🛑 Shutting down...")
	cancel()
	packet.StopPacketSniffer()
}
