package reassemble

import (
	"bufio"
	"io"
	"net/http"
	"time"

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
	reqReader       *io.PipeReader
	reqWriter       *io.PipeWriter
	respReader       *io.PipeReader
	respWriter       *io.PipeWriter
	out     chan<- *HttpTranscation
	clientIP string
}

type HttpTranscation struct {
	Method 	   string 
	URL 	   string 
	Host 	   string 
	ReqHeader  http.Header
	ReqBody    string 

	StatusCode int
	StatusText string 
	RespHeader http.Header
	RespBody   string

	ClientIP   string 
	Duration   time.Duration
	Timestamp  time.Time
}

func (s *httpStream) Accept(tcp *layers.TCP, ci gopacket.CaptureInfo, dir reassembly.TCPFlowDirection, nextSeq reassembly.Sequence, start *bool, ac reassembly.AssemblerContext) bool {
	return true
}

func (s *httpStream) ReassembledSG(sg reassembly.ScatterGather, ac reassembly.AssemblerContext) {
	length, _ := sg.Lengths()
	if length == 0 { 
		return
	}
	data := sg.Fetch(length)
	direction, _, _, _ := sg.Info()

	if len(data) > 0 { 
		switch direction {
		case reassembly.TCPDirServerToClient: 
			s.respWriter.Write(data)
		case reassembly.TCPDirClientToServer: 
			s.reqWriter.Write(data)
		}
	}
}

func (s *httpStream) ReassemblyComplete(ac reassembly.AssemblerContext) bool {
	s.reqWriter.Close()
	s.respWriter.Close()
	return true
}

type httpStreamFactory struct {
	out chan<- *HttpTranscation
}

func (h *httpStreamFactory) New(netFlow gopacket.Flow, tcpFlow gopacket.Flow, tcp *layers.TCP, ac reassembly.AssemblerContext) reassembly.Stream {
	reqRead, reqWrite := io.Pipe()
	respRead, respWrite := io.Pipe()
	stream := &httpStream{
		netFlow: netFlow,
		tcpFlow: tcpFlow,
		reqReader: reqRead,
		reqWriter: reqWrite,
		respReader: respRead,
		respWriter: respWrite,
		out:     h.out,
		clientIP: netFlow.Src().String(),
	}
	go stream.parse()
	return stream
}

func (s *httpStream) parse() {
	reqBuf := bufio.NewReader(s.reqReader)
	respBuf := bufio.NewReader(s.respReader)
	for {
		startTime := time.Now()

		req, err := http.ReadRequest(reqBuf)
		if err != nil { 
			return
		}

		host := req.Host
		if host == "" { 
			host = req.Header.Get("Host")
		}

		reqBodyStr := readBody(req.Body)
		req.Body.Close()

		resp, err := http.ReadResponse(respBuf, req)
		var statusCode int 
		var statusText string 
		var respBody string
		var respHeader http.Header
		if err == nil { 
			statusCode = resp.StatusCode
			statusText = resp.Status
			respHeader = resp.Header
			respBody = readBody(resp.Body)
			resp.Body.Close()
		}

		duration := time.Since(startTime)

		s.out <- &HttpTranscation{
			Method: req.Method,
			URL: req.URL.String(),
			Host: host,
			ReqHeader: req.Header,
			ReqBody: reqBodyStr,
			StatusCode: statusCode,
			StatusText: statusText,
			RespHeader: respHeader,
			RespBody: respBody,
			ClientIP: s.clientIP,
			Duration: duration,
			Timestamp: startTime,
		}
	}
}

func readBody(r io.ReadCloser) string {
	if r == nil { 
		return ""
	}
	bytes, err := io.ReadAll(io.LimitReader(r,64 * 1024))
	if err != nil {
		return ""
	}
	return string(bytes)

}
