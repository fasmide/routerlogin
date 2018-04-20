package dnsmasq

import (
	"os"
	"sync"
	"time"
)

// Store exposes an API to lookup dnsmasq leases by different means
type Store struct {
	Path string

	lock         sync.Mutex
	lastPopulate time.Time
	db           map[string]Entry
}

func (s *Store) populate() error {
	fd, err := os.Open(s.Path)
	if err != nil {
		return err
	}

	defer fd.Close()

	data, err := Parse(fd)
	if err != nil {
		return err
	}

	for _, e := range data {
		s.db[e.IP] = e
	}

	s.lastPopulate = time.Now()

	return nil
}

func (s *Store) ensure() error {
	if time.Now().Sub(s.lastPopulate) > time.Second*5 {
		s.db = make(map[string]Entry)
		return s.populate()
	}

	return nil
}

// LeaseByIP returns a single dnsmasq lease found by ip address
func (s *Store) LeaseByIP(ip string) (Entry, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// ensure we have recent data
	err := s.ensure()
	if err != nil {
		return Entry{}, err
	}

	return s.db[ip], nil

}
