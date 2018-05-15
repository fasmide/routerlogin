package conntrack

import (
	"bufio"
	"io"
	"strings"
)

// reader embedds io.ReadCloser and reads FlowUpdates from it
type reader struct {
	io.ReadCloser
	buffer *bufio.Reader
}

// NewReader returns a new reader
func NewReader(in io.ReadCloser) *reader {
	return &reader{ReadCloser: in, buffer: bufio.NewReader(in)}
}

// Read reads the next FlowUpdate from its embedded reader
func (r *reader) Read() (*FlowUpdate, error) {
	line, err := r.buffer.ReadString('\n')
	if err != nil {
		return nil, err
	}

	// the flow type knows nothing about the first 10 bytes, these
	// indicate if this flow is NEW, if its an UPDATE or a DESTROY'ed flow
	u := FlowUpdate{Type: strings.Trim(line[:10], " []")}
	u.Flow, err = ParseFlowLine(line[10:])

	if err != nil {
		return nil, err
	}

	return &u, nil
}
