package decode

import (
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type ParsedPacket struct{ 
	SrcIP net.IP
	DstIP net.IP
	SrcPort uint16
	DstPort uint16
	Protocol string
	Length int
}

type Parser struct{ 

}

func NewParser() *Parser{
	return &Parser{}
}

func(p *Parser) Parse(packet gopacket.Packet) (*ParsedPacket, bool) {
	var srcIP, dstIP net.IP

	if ipv4 := packet.Layer(layers.LayerTypeIPv4); ipv4 != nil {
		ip, _ := ipv4.(*layers.IPv4)
		srcIP = ip.SrcIP
		dstIP = ip.DstIP
	} else if ipv6 := packet.Layer(layers.LayerTypeIPv6); ipv6 != nil {
		ip, _ := ipv6.(*layers.IPv4)
		srcIP = ip.SrcIP
		dstIP = ip.DstIP
	} else { 
		return nil, false
	}

	var srcPort, dstPort uint16
	var protocol string 

	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil { 
		tcp, _ := tcpLayer.(*layers.TCP)
		srcPort = uint16(tcp.SrcPort)
		dstPort = uint16(tcp.DstPort)
		protocol = "TCP"
	} else if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil { 
		udp, _ := udpLayer.(*layers.UDP)
		srcPort = uint16(udp.SrcPort)
		dstPort = uint16(udp.DstPort)
		protocol = "UDP"
	} else { 
		return nil, false
	}

	return &ParsedPacket{
		SrcIP: srcIP,
		DstIP: dstIP,
		SrcPort: srcPort,
		DstPort: dstPort,
		Protocol: protocol,
		Length: len(packet.Data()),
	}, true
}