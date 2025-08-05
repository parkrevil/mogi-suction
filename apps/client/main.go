package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"unicode/utf8"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/layers"
)

func main() {
	fmt.Println("Starting packet capture...")

	// 네트워크 인터페이스 목록 가져오기
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal("Error finding devices:", err)
	}

	if len(devices) == 0 {
		log.Fatal("No network devices found")
	}

	// 첫 번째 인터페이스 사용
	device := devices[0].Name
	fmt.Printf("Using device: %s\n", device)

	// 패킷 캡처 핸들 열기
	handle, err := pcap.OpenLive(device, 1600, true, pcap.BlockForever)
	if err != nil {
		log.Fatal("Error opening device:", err)
	}
	defer handle.Close()

	// BPF 필터 설정 (TCP 포트 16000 - 소스 또는 목적지)
	err = handle.SetBPFFilter("tcp and (port 16000)")
	if err != nil {
		log.Fatal("Error setting BPF filter:", err)
	}

	fmt.Println("Starting packet capture on TCP port 16000...")
	fmt.Println("Filter: tcp and (port 16000)")
	fmt.Println("Press Ctrl+C to stop")

	// 시그널 핸들링
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 패킷 처리 함수
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packetCount := 0

	for {
		select {
		case packet := <-packetSource.Packets():
			packetCount++
			analyzePacket(packet, packetCount)
		case <-sigChan:
			fmt.Printf("\nCaptured %d packets. Stopping...\n", packetCount)
			return
		}
	}
}

func analyzePacket(packet gopacket.Packet, count int) {
	// 이더넷 레이어
	ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethernetLayer != nil {
		ethernet, _ := ethernetLayer.(*layers.Ethernet)
		fmt.Printf("\n[Packet %d] ", count)
		fmt.Printf("Ethernet: %s -> %s\n", ethernet.SrcMAC, ethernet.DstMAC)
	}

	// IP 레이어
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		ip, _ := ipLayer.(*layers.IPv4)
		fmt.Printf("IP: %s -> %s\n", ip.SrcIP, ip.DstIP)
	}

	// TCP 레이어
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		tcp, _ := tcpLayer.(*layers.TCP)

		fmt.Printf("TCP: %s:%d -> %s:%d\n", 
			packet.NetworkLayer().NetworkFlow().Src().String(),
			tcp.SrcPort,
			packet.NetworkLayer().NetworkFlow().Dst().String(),
			tcp.DstPort)
		
		// TCP 플래그 분석
		flags := []string{}
		if tcp.SYN { flags = append(flags, "SYN") }
		if tcp.ACK { flags = append(flags, "ACK") }
		if tcp.FIN { flags = append(flags, "FIN") }
		if tcp.RST { flags = append(flags, "RST") }
		if tcp.PSH { flags = append(flags, "PSH") }
		if tcp.URG { flags = append(flags, "URG") }
		
		fmt.Printf("Flags: %v, Seq: %d, Ack: %d\n", flags, tcp.Seq, tcp.Ack)
		
		// 페이로드 분석
		if len(tcp.Payload) > 0 {
			fmt.Printf("Payload size: %d bytes\n", len(tcp.Payload))
			
			// 페이로드 내용 출력 (처음 100바이트만)
			payload := tcp.Payload
			if len(payload) > 100 {
				payload = payload[:100]
				fmt.Printf("Payload (first 100 bytes): %x\n", payload)
			} else {
				fmt.Printf("Payload: %x\n", payload)
			}
			
			// UTF-8로 해석 시도
			if utf8.Valid(payload) {
				fmt.Printf("Payload (UTF-8): %s\n", string(payload))
			}
		}
	}

	fmt.Println("---")
} 