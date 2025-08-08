package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	quic "github.com/quic-go/quic-go"
)

// TODO 연결 끊김 시 재연결 로직 추가
// TODO 연결 끊김 시 sniffer 종료 로직 추가
// TODO 재연결 성공 시 sniffer 재시작 로직 추가
// TODO 연결 끊긴 상태에서 sniffer에서 데이터가 온다면 무시하고 받은 데이터 비우고(버퍼 비우기) sniffer flush

type QUICClient struct {
	conn    *quic.Conn
	addr    string
	tlsConf *tls.Config
}

func NewQUICClient(addr string) *QUICClient {
	return &QUICClient{
		addr: addr,
		tlsConf: &tls.Config{
			InsecureSkipVerify: true, // 개발용 자가서명 인증서 허용
			NextProtos:         []string{"mogi-suction-quic"},
		},
	}
}

func (c *QUICClient) Connect(ctx context.Context) error {
	conn, err := quic.DialAddr(ctx, c.addr, c.tlsConf, &quic.Config{
		MaxIdleTimeout:  60 * time.Second,
		KeepAlivePeriod: 30 * time.Second,
		EnableDatagrams: false,
	})
	if err != nil {
		return err
	}

	c.conn = conn
	log.Printf("Connected to server: %s", c.addr)
	return nil
}

func (c *QUICClient) SendData(data []byte) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	// 스트림 생성 시 타임아웃 설정
	streamCtx, streamCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer streamCancel()

	stream, err := c.conn.OpenStreamSync(streamCtx)
	if err != nil {
		return err
	}
	defer stream.Close()

	_, err = stream.Write(data)
	if err != nil {
		return err
	}

	log.Printf("Sent %d bytes to server", len(data))
	return nil
}

func (c *QUICClient) Close() error {
	if c.conn != nil {
		return c.conn.CloseWithError(0, "client shutdown")
	}
	return nil
}

func (c *QUICClient) IsConnected() bool {
	return c.conn != nil
}
