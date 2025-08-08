package main

import (
	"context"
	"log"
	"time"
)

func testQUICClient() {
	// QUIC 클라이언트 생성
	client := NewQUICClient("localhost:8443")

	// 서버에 연결
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// 테스트 데이터 전송
	testData := []byte("Hello from QUIC client!")

	if err := client.SendData(testData); err != nil {
		log.Fatalf("Failed to send data: %v", err)
	}

	log.Printf("Data sent successfully")
	time.Sleep(1 * time.Second) // 서버 처리 시간 대기
}
