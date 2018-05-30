package daemon

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/olekukonko/tablewriter"
)

// Store is the interface we expect from other packages store's
type Store interface {
	Addresses() ([]net.IP, error)
	Data(string) (map[string]string, error)
}

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

	t := tablewriter.NewWriter(w)

	addresses := d.addresses()

	data := make([][]string, 0)
	header := make([]string, 2)
	fieldOrder := make(map[string]int)

	// we want hostname and ip first
	fieldOrder["hostname"] = 0
	fieldOrder["ip"] = 1

	for _, addr := range addresses {
		addrData, err := d.data(addr)
		if err != nil {
			return 0, fmt.Errorf("could not get address data: %s", err)
		}

		line := make([]string, len(addrData)+1)
		// loop this ip's data and put it into our table
		for key, value := range addrData {

			var o int
			var exists bool
			if o, exists = fieldOrder[key]; !exists {
				// assign next fieldOrder
				o = len(fieldOrder)
				fieldOrder[key] = o
				header = append(header, "")
			}
			header[o] = key
			line[o] = value
		}
		data = append(data, line)

	}
	t.SetHeader(header)
	t.SetBorder(false)
	t.AppendBulk(data)
	t.Render()

	return 0, nil
}

// Data will collect data from all stores and combine them
// the map consists of fieldName -> value
func (d *Daemon) data(ip string) (map[string]string, error) {
	line := make(map[string]string)
	line["ip"] = ip

	for _, store := range d.stores {
		sData, err := store.Data(ip)
		if err != nil {
			return nil, fmt.Errorf("unable to retive data from %T: %s", store, err)
		}

		// put this data into our line
		for key, value := range sData {
			// we should not overwrite keys
			if _, exists := line[key]; exists {
				return nil, fmt.Errorf("duplicate key found: %s", key)
			}
			line[key] = value
		}
	}
	return line, nil
}

// addresses returns the distinct set of ip addresses from all stores
func (d *Daemon) addresses() []string {
	// lets try one of these new and shiny concurrent maps
	var m sync.Map
	var wg sync.WaitGroup

	for _, store := range d.stores {
		wg.Add(1)
		go func(store Store) {
			ips, err := store.Addresses()
			if err != nil {
				log.Printf("unable to get addresses from %T: %s", store, err)
				return
			}
			for _, ip := range ips {
				m.LoadOrStore(ip.String(), struct{}{})
			}
			wg.Done()
		}(store)
	}

	wg.Wait()

	// on a simple home lan, lets expect there to be less then 256 in most cases
	ips := make([]string, 0, 256)

	// m should now hold distinct ip addresses
	m.Range(func(key, value interface{}) bool {

		// we are pretty confident this map holds strings
		ip, _ := key.(string)
		ips = append(ips, ip)
		return true
	})

	return ips
}

// AddStore adds given stores to the daemon
func (d *Daemon) AddStore(s ...Store) {
	d.stores = append(d.stores, s...)
}
