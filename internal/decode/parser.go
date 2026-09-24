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
	RawPacket gopacket.Packet
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
		RawPacket: packet,
	}, true
}

func (p *Parser) equals(pack1 ParsedPacket, pack2 ParsedPacket) bool { 
	if pack1.DstIP.Equal(pack2.DstIP)&& pack1.DstPort == pack2.DstPort && pack1.SrcPort == pack2.SrcPort && pack1.SrcIP.Equal(pack2.SrcIP) && pack1.Protocol == pack2.Protocol {
		return true
	}
	return false 
}