package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	fmt.Println("Hello World from Client!")
	
	// 간단한 HTTP 요청으로 서버 테스트
	go func() {
		time.Sleep(1 * time.Second)
		
		resp, err := http.Get("http://localhost:8080")
		if err != nil {
			log.Printf("Error connecting to server: %v", err)
			return
		}
		defer resp.Body.Close()
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response: %v", err)
			return
		}
		
		log.Printf("Server response: %s", string(body))
	}()
	
	// 5초 대기 후 종료
	time.Sleep(5 * time.Second)
	fmt.Println("Client finished!")
} 