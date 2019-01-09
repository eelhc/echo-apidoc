package apidoc

import (
	"bufio"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo"
)

type docResponse struct {
	status int
	header http.Header
	body   []byte
}

func newDocResponse() *docResponse {
	return &docResponse{
		header: http.Header{},
		status: http.StatusInternalServerError,
		body:   []byte{},
	}
}

func (resp *docResponse) StatusCode() string {
	return strconv.Itoa(resp.status)
}

func (resp *docResponse) MediaType() string {
	return resp.header.Get("Content-Type")
}

func (resp *docResponse) Body() string {
	if strings.Contains(resp.MediaType(), "application/json") {
		return bytesToJSONPretty(resp.body, defaultIndent)
	}
	return string(resp.body)
}

func (resp *docResponse) Header() string {
	var header string
	for hkey, hval := range resp.header {
		if hkey == "Content-Length" {
			continue
		}
		header += hkey + ": " + strings.Join(hval, ", ") + "\n"
	}
	return header
}

type respWriterWrapper struct {
	echoResp *echo.Response
	docResp  *docResponse
	http.ResponseWriter
}

func newRespWriterWrapper(eresp *echo.Response, dresp *docResponse) *respWriterWrapper {
	return &respWriterWrapper{
		echoResp:       eresp,
		docResp:        dresp,
		ResponseWriter: eresp.Writer,
	}
}

func (wr *respWriterWrapper) WriteHeader(code int) {
	wr.ResponseWriter.WriteHeader(code)
	wr.docResp.status = code
	wr.docResp.header = wr.echoResp.Header()
}

func (wr *respWriterWrapper) Write(b []byte) (int, error) {
	wr.docResp.body = append(wr.docResp.body, b...)
	return wr.ResponseWriter.Write(b)
}

func (wr *respWriterWrapper) Flush() {
	wr.ResponseWriter.(http.Flusher).Flush()
}

func (wr *respWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return wr.ResponseWriter.(http.Hijacker).Hijack()
}

func (wr *respWriterWrapper) CloseNotify() <-chan bool {
	return wr.ResponseWriter.(http.CloseNotifier).CloseNotify()
}
