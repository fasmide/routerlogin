package daemon

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/olekukonko/tablewriter"
)

// Daemon accepts connections from a listener and outputs data when they connect
type Daemon struct {
	stores []Store
}

// Accept accepts everything on given listener
func (d *Daemon) Accept(l net.Listener) error {

	for {
		fd, err := l.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}

		go func(c net.Conn) {
			_, err := d.WriteTo(c)
			if err != nil {
				log.Printf("failed writing to connection: %s", err)
			}
			c.Close()
		}(fd)
	}
}

// WriteTo to outputs our output to a writer
func (d *Daemon) WriteTo(w io.Writer) (int64, error) {

	c := Collector{Stores: d.stores}
	err := c.Collect()

	if err != nil {
		return 0, fmt.Errorf("unable to collect data: %s", err)
	}

	t := tablewriter.NewWriter(w)
	t.SetHeader(c.Headers)
	t.SetBorder(false)
	t.AppendBulk(c.Data)
	t.Render()

	return 0, nil
}

// AddStore adds given stores to the daemon
func (d *Daemon) AddStore(s ...Store) {
	d.stores = append(d.stores, s...)
}
