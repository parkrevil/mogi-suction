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

func startServer(ctx context.Context) error {
	tlsConf, err := loadTLSConfig(resolveDevTLSPaths())
	if err != nil {
		return err
	}

	port := 8443
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{Port: port})
	if err != nil {
		return err
	}

	const bufferSize = 32 << 20
	_ = udpConn.SetReadBuffer(bufferSize)
	_ = udpConn.SetWriteBuffer(bufferSize)

	listener, err := quic.Listen(udpConn, tlsConf, &quic.Config{
		MaxIdleTimeout:  60 * time.Second,
		KeepAlivePeriod: 30 * time.Second,
		EnableDatagrams: false,
	})
	if err != nil {
		return err
	}

	log.Printf("Server listening on udp%s", addr)

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			return nil
		}
		go handleConnection(conn)
	}
}

func resolveDevTLSPaths() (string, string) {
	wd, _ := os.Getwd()
	root := filepathpkg.Clean(filepathpkg.Join(wd, "..", ".."))

	return filepathpkg.Join(root, "dev-cert.pem"), filepathpkg.Join(root, "dev-key.pem")
}

func handleConnection(conn *quic.Conn) {
	defer conn.CloseWithError(0, "server shutdown")
	log.Printf("Client connected: %s -> %s", conn.RemoteAddr(), conn.LocalAddr())

	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			return
		}

		go func(s *quic.Stream) {
			defer s.Close()
			var buf [64 * 1024]byte
			var total int64

			for {
				n, err := s.Read(buf[:])
				if n > 0 {
					total += int64(n)
				}
				if err != nil {
					break
				}
			}
		}(stream)
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
