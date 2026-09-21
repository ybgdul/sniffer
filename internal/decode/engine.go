package decode

import (
	"context"
	"sync"

	"github.com/google/gopacket"
)

type Engine struct{ 
	workerCount int
	parser *Parser
}

func NewEngine(workerCount int) *Engine { 
	return &Engine{
		workerCount: workerCount,
		parser: &Parser{},
	}
}

func (e *Engine) Decode(ctx context.Context, rawPacket <-chan gopacket.Packet) <-chan *ParsedPacket { 
	out := make(chan *ParsedPacket, 1024)
	var wg sync.WaitGroup

	for i := 0; i < e.workerCount; i++ { 
		wg.Add(1)
		go func() { 
			defer wg.Done()
			for { 
				select { 
				case <-ctx.Done(): 
					return 
				case packet, ok := <-rawPacket: 
					if !ok { 
						return
					}
					if parsed, parse := e.parser.Parse(packet); parse { 
						out<- parsed
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}