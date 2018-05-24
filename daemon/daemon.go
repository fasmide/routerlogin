package daemon

import (
	"fmt"
	"log"
	"net"
	"sync"
)

// Store is the interface we expect from other packages store's
type Store interface {
	Addresses() ([]net.IP, error)
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

		go func() {
			addresses := d.addresses()
			for _, addr := range addresses {
				fd.Write([]byte(fmt.Sprintf("Address: %s\n", addr)))
			}
			fd.Write([]byte("Bye\n"))
			fd.Close()
		}()
	}
}

// addresses returns the distinct set of ip addresses from all stores
func (d *Daemon) addresses() []string {
	// lets try one of these new and shiny concurrent maps
	var m sync.Map
	var wg sync.WaitGroup

	for _, store := range d.stores {
		wg.Add(1)
		go func() {
			ips, err := store.Addresses()
			if err != nil {
				log.Printf("unable to get addresses from %T: %s", store, err)
				return
			}
			for _, ip := range ips {
				m.LoadOrStore(ip.String(), struct{}{})
			}
			wg.Done()
		}()
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
