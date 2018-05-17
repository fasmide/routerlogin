package conntrack

import (
	"bufio"
	"io"
)

// reader embedds io.Reader and reads FlowUpdates from it
type reader struct {
	io.Reader
	buffer *bufio.Reader
}

// NewReader returns a new reader
func NewReader(in io.Reader) *reader {
	return &reader{Reader: in, buffer: bufio.NewReader(in)}
}

// Read reads the next Flow from its embedded reader
func (r *reader) Read() (*Flow, error) {
	line, err := r.buffer.ReadString('\n')
	if err != nil {
		return nil, err
	}

	f, err := ParseFlowLine(line)
	if err != nil {
		return nil, err
	}

	return &f, nil
}
