package main

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"os"
	filepathpkg "path/filepath"
	"sync/atomic"
	"time"

	quic "github.com/quic-go/quic-go"
)

const (
	serverPort     = 8443
	bufferSize     = 32 << 20
	maxConnections = 5000 // 최대 연결 수 제한
)

func startQUICServer(ctx context.Context) error {
	tlsConf, err := loadTLSConfig(resolveDevTLSPaths())
	if err != nil {
		return err
	}

	// 연결 수 제한을 위한 카운터
	var activeConnections int32

	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{Port: serverPort})
	if err != nil {
		return err
	}

	if err := udpConn.SetReadBuffer(bufferSize); err != nil {
		log.Printf("Failed to set read buffer: %v", err)
	}
	if err := udpConn.SetWriteBuffer(bufferSize); err != nil {
		log.Printf("Failed to set write buffer: %v", err)
	}

	listener, err := quic.Listen(udpConn, tlsConf, &quic.Config{
		MaxIdleTimeout:  60 * time.Second,
		KeepAlivePeriod: 30 * time.Second,
		EnableDatagrams: false,
	})
	if err != nil {
		udpConn.Close()
		return err
	}
	defer func() {
		listener.Close()
		udpConn.Close()
	}()

	log.Printf("Server listening on udp:%d (buffer: %d MB)", serverPort, bufferSize/(1<<20))

	for {
		select {
		case <-ctx.Done():
			log.Println("Server shutdown requested")
			return nil
		default:
			// Accept 타임아웃 설정으로 블로킹 방지
			acceptCtx, acceptCancel := context.WithTimeout(ctx, 100*time.Millisecond)
			conn, err := listener.Accept(acceptCtx)
			acceptCancel()

			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				continue
			}
			// 연결 수 제한 확인
			if atomic.LoadInt32(&activeConnections) >= maxConnections {
				log.Printf("Maximum connections reached (%d), rejecting connection", maxConnections)
				conn.CloseWithError(0, "server overloaded")
				continue
			}

			atomic.AddInt32(&activeConnections, 1)
			go func() {
				defer atomic.AddInt32(&activeConnections, -1)
				handleConnection(ctx, conn)
			}()
		}
	}
}

func resolveDevTLSPaths() (string, string) {
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to get working directory: %v", err)
		// 기본값 반환
		return "dev-cert.pem", "dev-key.pem"
	}
	root := filepathpkg.Clean(filepathpkg.Join(wd, "..", ".."))

	return filepathpkg.Join(root, "dev-cert.pem"), filepathpkg.Join(root, "dev-key.pem")
}

func handleConnection(ctx context.Context, conn *quic.Conn) {
	defer conn.CloseWithError(0, "server shutdown")
	log.Printf("Client connected: %s -> %s", conn.RemoteAddr(), conn.LocalAddr())

	// AcceptStream 타임아웃 설정으로 블로킹 방지
	acceptCtx, acceptCancel := context.WithTimeout(ctx, 100*time.Millisecond)
	stream, err := conn.AcceptStream(acceptCtx)
	acceptCancel()

	if err != nil {
		log.Printf("Failed to accept stream: %v", err)
		return
	}
	defer stream.Close()

	var buf [64 * 1024]byte
	var total int64

	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := stream.Read(buf[:])
			if n > 0 {
				total += int64(n)
			}
			if err != nil {
				return
			}
		}
	}
}

func loadTLSConfig(certPath, keyPath string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
		NextProtos:   []string{"mogi-suction-quic"},
	}, nil
}
