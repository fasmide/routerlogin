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
	Data(string) (map[string]string, error)
}

type Collector struct {
	// Stores are the stores we will be reading from
	Stores []Store

	// Data and Headers can be read, when Collect have finished with a non error return value
	Data    [][]string
	Headers []string
}

func (c *Collector) Collect() error {

	addresses := c.addresses()

	c.Data = make([][]string, 0)
	c.Headers = make([]string, 2)
	fieldOrder := make(map[string]int)

	// we want hostname and ip first
	fieldOrder["hostname"] = 0
	fieldOrder["ip"] = 1

	for _, addr := range addresses {
		addrData, err := c.data(addr)
		if err != nil {
			return fmt.Errorf("could not get address data: %s", err)
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
				c.Headers = append(c.Headers, "")
			}
			c.Headers[o] = key
			line[o] = value
		}
		c.Data = append(c.Data, line)

	}
	return nil
}

// Data will collect data from all stores and combine them
// the map consists of fieldName -> value
func (c *Collector) data(ip string) (map[string]string, error) {
	line := make(map[string]string)
	line["ip"] = ip

	for _, store := range c.Stores {
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
func (c *Collector) addresses() []string {
	// lets try one of these new and shiny concurrent maps
	var m sync.Map
	var wg sync.WaitGroup

	for _, store := range c.Stores {
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
