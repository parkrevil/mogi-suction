package packet

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/reassembly"
	"github.com/shirou/gopsutil/v3/mem"
)

var (
	handle    *pcap.Handle
	assembler *reassembly.Assembler
	ctx       context.Context
	wg        sync.WaitGroup
	pcapFile  string // PCAP 파일 경로 (동적으로 설정)
)

// getDefaultPcapPath 기본 PCAP 파일 경로를 반환합니다
func getDefaultPcapPath() string {
	// 현재 실행 파일의 위치를 기준으로 상대 경로 계산
	if wd, err := os.Getwd(); err == nil {
		// apps/client에서 실행되는 경우
		if filepath.Base(wd) == "client" {
			return "../../samples/raid_glasgivnen.pcap"
		}
		// 프로젝트 루트에서 실행되는 경우
		return "samples/raid_glasgivnen.pcap"
	}

	// 절대 경로 시도
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "mogi-suction", "samples", "raid_glasgivnen.pcap")
}

func init() {
	pcapFile = getDefaultPcapPath()
}

// SetPcapFile PCAP 파일 경로를 설정합니다
func SetPcapFile(filePath string) {
	pcapFile = filePath
}

// InitPacketSnifferLive 라이브 네트워크 캡처를 위한 초기화 (기존 기능)
func InitPacketSnifferLive(context context.Context) error {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return err
	}

	if len(devices) == 0 {
		return errors.New("no network devices found")
	}

	handle, err = pcap.OpenLive(devices[0].Name, 65535, true, pcap.BlockForever)
	if err != nil {
		return err
	}

	if err := handle.SetBPFFilter("tcp and port 16000"); err != nil {
		return err
	}

	totalPages, pagesPerConnection, _ := calcOptimalParams()

	streamFactory := &tcpStreamFactory{}
	streamPool := reassembly.NewStreamPool(streamFactory)
	assembler = reassembly.NewAssembler(streamPool)
	assembler.MaxBufferedPagesTotal = totalPages
	assembler.MaxBufferedPagesPerConnection = pagesPerConnection

	ctx = context
	log.Printf("Initialized live packet capture on device: %s", devices[0].Name)
	return nil
}

func InitPacketSniffer(context context.Context) error {
	// 라이브 네트워크 캡처 (주석처리)
	/*
		devices, err := pcap.FindAllDevs()
		if err != nil {
			return err
		}

		if len(devices) == 0 {
			return errors.New("no network devices found")
		}

		handle, err = pcap.OpenLive(devices[0].Name, 65535, true, pcap.BlockForever)
		if err != nil {
			return err
		}

		if err := handle.SetBPFFilter("tcp and port 16000"); err != nil {
			return err
		}
	*/

	// PCAP 파일 존재 확인
	if _, err := os.Stat(pcapFile); os.IsNotExist(err) {
		return fmt.Errorf("pcap file does not exist: %s", pcapFile)
	}

	// PCAP 파일 읽기
	var err error
	handle, err = pcap.OpenOffline(pcapFile)
	if err != nil {
		return fmt.Errorf("failed to open pcap file %s: %w", pcapFile, err)
	}

	log.Printf("Reading PCAP file: %s", pcapFile)

	// BPF 필터 적용 (PCAP 파일에도 동일하게 적용)
	if err := handle.SetBPFFilter("tcp and port 16000"); err != nil {
		return err
	}

	totalPages, pagesPerConnection, _ := calcOptimalParams()

	streamFactory := &tcpStreamFactory{}
	streamPool := reassembly.NewStreamPool(streamFactory)
	assembler = reassembly.NewAssembler(streamPool)
	assembler.MaxBufferedPagesTotal = totalPages
	assembler.MaxBufferedPagesPerConnection = pagesPerConnection

	ctx = context
	return nil
}

func StartPacketSniffer() {
	wg.Add(1)
	go func() {
		defer wg.Done()

		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				assembler.FlushWithOptions(reassembly.FlushOptions{
					T:  time.Now().Add(-6 * time.Second),
					TC: time.Now().Add(-4 * time.Second),
				})
			case packet, ok := <-packetSource.Packets():
				if !ok {
					return
				}

				tcpLayer := packet.Layer(layers.LayerTypeTCP)
				if tcpLayer == nil {
					continue
				}

				tcp, ok := tcpLayer.(*layers.TCP)
				if !ok {
					continue
				}

				assembler.AssembleWithContext(
					packet.NetworkLayer().NetworkFlow(),
					tcp,
					&Context{
						CaptureInfo: packet.Metadata().CaptureInfo,
						TCP:         tcp,
					},
				)
			}
		}
	}()
}

func StopPacketSniffer() {
	assembler.FlushAll()
	wg.Wait()
}

func ClosePacketSniffer() {
	if handle != nil {
		handle.Close()
	}
}

func calcOptimalParams() (totalPages, perConnectionPages, optimalBuffer int) {
	totalPages = 10000
	perConnectionPages = 100
	optimalBuffer = 1000

	vmstat, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("Error getting virtual memory: %v", err)

		return totalPages, perConnectionPages, optimalBuffer
	}

	numCPU := runtime.NumCPU()
	availableGB := float64(vmstat.Available) / (1024 * 1024 * 1024)
	memoryForReassembly := availableGB * 0.1
	pagesPerGB := 250000
	totalPages = int(memoryForReassembly * float64(pagesPerGB))

	if totalPages < 10000 {
		totalPages = 10000
	} else if totalPages > 500000 {
		totalPages = 500000
	}

	perConnectionPages = totalPages / 100

	if perConnectionPages < 100 {
		perConnectionPages = 100
	} else if perConnectionPages > 5000 {
		perConnectionPages = 5000
	}

	baseBufferPerCore := 500
	memoryBuffer := int(availableGB * 200)
	cpuBuffer := numCPU * baseBufferPerCore
	optimalBuffer = cpuBuffer + memoryBuffer

	if optimalBuffer < 1000 {
		optimalBuffer = 1000
	} else if optimalBuffer > 10000 {
		optimalBuffer = 10000
	}

	return totalPages, perConnectionPages, optimalBuffer
}
