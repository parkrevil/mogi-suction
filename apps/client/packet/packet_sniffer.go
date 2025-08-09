package packet

import (
	"context"
	"errors"
	"log"
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
)

func InitPacketSniffer(context context.Context) error {
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
