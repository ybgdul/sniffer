package reassemble

import (
	"bufio"
	"io"
	"net/http"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/reassembly"
)

type HttpRequest struct {
	Method   string
	URL      string
	Host     string
	Protocol string
	IP       string

	Header 	 http.Header
	Body 	 string 
	Type     string
}

type httpStream struct {
	netFlow gopacket.Flow
	tcpFlow gopacket.Flow
	r       *io.PipeReader
	w       *io.PipeWriter
	out     chan<- *HttpRequest
}

func (s *httpStream) Accept(tcp *layers.TCP, ci gopacket.CaptureInfo, dir reassembly.TCPFlowDirection, nextSeq reassembly.Sequence, start *bool, ac reassembly.AssemblerContext) bool {
	return true
}

func (s *httpStream) ReassembledSG(sg reassembly.ScatterGather, ac reassembly.AssemblerContext) {
	length, _ := sg.Lengths()
	if length == 0 { 
		return
	}

	bytes := sg.Fetch(length)

	if len(bytes) > 0 { 
		s.w.Write(bytes)
	}
}

func (s *httpStream) ReassemblyComplete(ac reassembly.AssemblerContext) bool {
	s.w.Close()
	return true
}

type httpStreamFactory struct {
	out chan<- *HttpRequest
}

func (h *httpStreamFactory) New(netFlow gopacket.Flow, tcpFlow gopacket.Flow, tcp *layers.TCP, ac reassembly.AssemblerContext) reassembly.Stream {
	pr, pw := io.Pipe()
	stream := &httpStream{
		netFlow: netFlow,
		tcpFlow: tcpFlow,
		r:       pr,
		w:       pw,
		out:     h.out,
	}
	go stream.parse()
	return stream
}

func (s *httpStream) parse() {
	buf := bufio.NewReader(s.r)
	for {
		req, err := http.ReadRequest(buf)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return
		} else if err != nil {
			continue
		}

		var bodyStr string
		if req.Body != nil { 
			bytes, err := io.ReadAll(io.LimitReader(req.Body, 64*1024)) 
			if err == nil && len(bytes) > 0{ 
				bodyStr = string(bytes)
			}
			req.Body.Close()
		}

		s.out <- &HttpRequest{
			Method:   req.Method,
			URL:      req.URL.String(),
			Host:     req.Host,
			Protocol: req.Proto,
			IP:       s.netFlow.Src().String(),
			Header:   req.Header,
			Body:     bodyStr,
			Type: 	  req.Header.Get("Content-Type"),
		}
		req.Body.Close()
	}
}
