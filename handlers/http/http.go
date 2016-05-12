package http

import (
	"bytes"
	"fmt"
	"github.com/go-playground/log"
	httpclient "net/http"
	"net/url"
)

// Formatter is the function used to format the HTTP entry
type Formatter func(e *log.Entry) string

// HTTP is an instance of the http logger
type HTTP struct {
	buffer             uint // channel buffer
	remoteHost         string
	formatter          Formatter
	hasCustomFormatter bool
	contentEncoding    string
	httpClient         httpclient.Client
	numWorkers         uint
}

func init() {

}

func New(bufferSize uint, remoteHost string) (*HTTP, error) {

	h := &HTTP{
		buffer:             0,
		remoteHost:         "http://localhost:8888/",
		contentEncoding:    "application/x-www-form-urlencoded",
		hasCustomFormatter: false,
		numWorkers:         1,
	}

	h.httpClient = httpclient.Client{}

	h.buffer = bufferSize

	if _, err := url.Parse(remoteHost); err != nil {
		return nil, err
	}
	h.remoteHost = remoteHost
	h.formatter = func(e *log.Entry) string {
		return fmt.Sprintf("%s", e.Message)
	}

	return h, nil
}

func (h *HTTP) SetBuffer(buff uint) {
	h.buffer = buff
}

func (h *HTTP) SetContentEncoding(encoding string) {
	h.contentEncoding = encoding
}

func (h *HTTP) SetRemoteHost(host string) {
	if _, err := url.Parse(host); err != nil {
		h.remoteHost = "http://localhost:8888/"
	} else {
		h.remoteHost = host
	}
}

func (h *HTTP) SetFormatter(f Formatter) {
	h.formatter = f
	h.hasCustomFormatter = true
}

func (h *HTTP) SetNumWorkers(num uint) {
	if num >= 1 {
		h.numWorkers = num
	}
}

// Run starts the logger consuming on the returned channed
func (h *HTTP) Run() chan<- *log.Entry {
	ch := make(chan *log.Entry, h.buffer)
	for i := 0; i <= int(h.numWorkers); i++ {
		go h.handleLog(ch)
	}
	return ch
}

func (h *HTTP) handleLog(entries <-chan *log.Entry) {
	var e *log.Entry
	var payload string

	for e = range entries {

		payload = h.formatter(e)
		b := bytes.NewBufferString(payload)

		// Issue POST request to send off data
		req, err := httpclient.NewRequest("POST", h.remoteHost, b)
		if err != nil {
			log.Info(fmt.Sprintf("[Error] Could not initialize new request: %v\n", err))
		}
		req.Header.Add("Content-Type", h.contentEncoding)
		resp, err := h.httpClient.Do(req)
		if err != nil {
			log.Info(fmt.Sprintf("[Error] Could not post data to %s: %v\n", h.remoteHost, err))
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 299 {
			log.Info(fmt.Sprintf("[Error] Received HTTP %d during POST request to %s\n", resp.StatusCode, h.remoteHost))
		}
		e.WG.Done()
	}
}
