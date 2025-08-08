package main

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"os"
	filepathpkg "path/filepath"
	"time"

	quic "github.com/quic-go/quic-go"
)

const (
	serverPort = 8443
	bufferSize = 32 << 20
)

func startQUICServer(ctx context.Context) error {
	tlsConf, err := loadTLSConfig(resolveDevTLSPaths())
	if err != nil {
		return err
	}

	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{Port: serverPort})
	if err != nil {
		return err
	}

	_ = udpConn.SetReadBuffer(bufferSize)
	_ = udpConn.SetWriteBuffer(bufferSize)

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
			conn, err := listener.Accept(context.Background())
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				continue
			}
			go handleConnection(ctx, conn)
		}
	}
}

func resolveDevTLSPaths() (string, string) {
	wd, _ := os.Getwd()
	root := filepathpkg.Clean(filepathpkg.Join(wd, "..", ".."))

	return filepathpkg.Join(root, "dev-cert.pem"), filepathpkg.Join(root, "dev-key.pem")
}

func handleConnection(ctx context.Context, conn *quic.Conn) {
	defer conn.CloseWithError(0, "server shutdown")
	log.Printf("Client connected: %s -> %s", conn.RemoteAddr(), conn.LocalAddr())

	stream, err := conn.AcceptStream(context.Background())
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
				log.Printf("Received %d bytes (total: %d)", n, total)
			}
			if err != nil {
				break
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
