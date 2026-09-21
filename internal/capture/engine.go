package capture

import (
	"context"
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

type Engine struct{ 
	handle *pcap.Handle
}

func NewEngine(deviceName string) (*Engine, error){ 
	handle, err := pcap.OpenLive(deviceName, 1600, true, pcap.BlockForever)
	if err != nil { 
		return nil, fmt.Errorf("opening a device: %w", err)
	}
	bpfErr := handle.SetBPFFilter("tcp and port 80")
	if bpfErr != nil { 
		return nil, fmt.Errorf("setting a bpf filter: %w", err)
	}
	return &Engine{handle: handle}, nil
}

func (e *Engine) Sniff(ctx context.Context) (<-chan gopacket.Packet) { 
	packetChan := make(chan gopacket.Packet, 1024)

	packetSource := gopacket.NewPacketSource(e.handle, e.handle.LinkType())

	go func() { 
		defer close(packetChan)
		defer e.handle.Close()

		for {
			select { 
			case <-ctx.Done(): 
				return 
			default: 
				packet, err := packetSource.NextPacket()
				if err != nil { 
					continue
				}
				packetChan<- packet
			}
		}
	}()

	return packetChan
}