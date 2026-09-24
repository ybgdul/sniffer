package reassemble

import (
	"context"
	"sniffer/internal/decode"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/reassembly"
)

type Engine struct{
	assembler *reassembly.Assembler
}

func NewEngine(out chan<- *HttpRequest) *Engine {
	factory := &httpStreamFactory{out: out}
	streamPool := reassembly.NewStreamPool(factory)
	assembler := reassembly.NewAssembler(streamPool)

	return &Engine{
		assembler: assembler,
	}
}

func (e *Engine) Run(ctx context.Context, in <-chan *decode.ParsedPacket) { 
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for { 
		select { 
		case <-ctx.Done():
			e.assembler.FlushAll()
			return
		case packet, ok := <-in: 
			if !ok { 
				e.assembler.FlushAll()
				return
			}
			if tcpLayer := packet.RawPacket.Layer(layers.LayerTypeTCP); tcpLayer != nil { 
				tcp := tcpLayer.(*layers.TCP)
				ctxInf := &Context{
					CaptureInfo: packet.RawPacket.Metadata().CaptureInfo,
				}
				e.assembler.AssembleWithContext(
					packet.RawPacket.NetworkLayer().NetworkFlow(),
					tcp,
					ctxInf,
				)
			}
		case <-ticker.C: 
			e.assembler.FlushCloseOlderThan(time.Now().Add(-15 * time.Second))
		}
	}
}

type Context struct{ 
	CaptureInfo gopacket.CaptureInfo
}
func (c *Context) GetCaptureInfo() gopacket.CaptureInfo { 
	return c.CaptureInfo
}