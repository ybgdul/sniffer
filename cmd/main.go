package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"sniffer/internal/capture"
	"sniffer/internal/decode"
)

func init() { 

}

func main() { 
	ctx, stop := signal.NotifyContext(context.Background())
	defer stop()

	target := "en0"

	capturer, err := capture.NewEngine(target)
	if err != nil { 
		log.Fatalf("Failed to initialize packet capture: %w", err)
	}

	rawPacket := capturer.Sniff(ctx)

	decoder := decode.NewEngine(4)
	parsedChan := decoder.Decode(ctx, rawPacket)

	fmt.Printf("Started sniffing... sniff-sniff...\n")
	for parsed := range parsedChan { 
		fmt.Printf("[%s] %s:%d -> %s:%d (%d bytes)\n",
			parsed.Protocol,
			parsed.SrcIP, parsed.SrcPort,
			parsed.DstIP, parsed.DstPort,
			parsed.Length,
		)
	}
}