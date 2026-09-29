package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sniffer/design"
	"sniffer/internal/capture"
	"sniffer/internal/decode"
	"sniffer/internal/reassemble"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
)

func init() { 
	//haven't thought of it much tbh
}

func main() { 
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	target := "en0"

	capturer, err := capture.NewEngine(target)
	if err != nil { 
		log.Fatalf("Failed to initialize packet capture: %v", err)
	}

	rawPacket := capturer.Sniff(ctx)

	decoder := decode.NewEngine(4)
	parsedChan := decoder.Decode(ctx, rawPacket)

	httpRequests := make(chan *reassemble.HttpRequest, 500)
	assembler := reassemble.NewEngine(httpRequests)
	go assembler.Run(ctx, parsedChan)

	p := tea.NewProgram(design.NewModel(httpRequests), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Bubbletea has failed: %v", err)
	}

}