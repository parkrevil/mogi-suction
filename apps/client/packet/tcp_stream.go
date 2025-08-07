package packet

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/reassembly"
)

type tcpStreamFactory struct{}

type tcpStream struct {
	net, transport gopacket.Flow
}

type Context struct {
	CaptureInfo gopacket.CaptureInfo
	TCP         *layers.TCP
}

func (c *Context) GetCaptureInfo() gopacket.CaptureInfo {
	return c.CaptureInfo
}

func (t *tcpStreamFactory) New(net, transport gopacket.Flow, tcp *layers.TCP, ac reassembly.AssemblerContext) reassembly.Stream {
	return &tcpStream{
		net:       net,
		transport: transport,
	}
}

func (t *tcpStream) Accept(tcp *layers.TCP, ci gopacket.CaptureInfo, dir reassembly.TCPFlowDirection, nextSeq reassembly.Sequence, start *bool, ac reassembly.AssemblerContext) bool {
	return true
}

func (t *tcpStream) ReassembledSG(sg reassembly.ScatterGather, ac reassembly.AssemblerContext) {
	if ac == nil {
		return
	}

	if context, ok := ac.(*Context); !ok || context.TCP == nil || !context.TCP.PSH {
		return
	}

	_, _, skip, _ := sg.Info()

	if skip {
		return
	}

	length, _ := sg.Lengths()

	analyzePayload(sg.Fetch(length))
}

func (t *tcpStream) ReassemblyComplete(ac reassembly.AssemblerContext) bool {
	return false
}
