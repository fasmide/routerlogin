package dnsmasq

import (
	"fmt"
	"net"
	"os"
	"sort"
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

// byExpiry is used for sorting
type byExpiry []Entry

func (a byExpiry) Len() int           { return len(a) }
func (a byExpiry) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byExpiry) Less(i, j int) bool { return a[i].Expiry.Before(a[j].Expiry) }

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
func (s *Store) LeaseByIP(ip string) (*Entry, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// ensure we have recent data
	err := s.ensure()
	if err != nil {
		return nil, err
	}

	if entry, found := s.db[ip]; found {
		return &entry, nil
	}

	return nil, fmt.Errorf("no Entry with ip %s", ip)
}

// Addresses returns all net.IP addresses discovered by this store
func (s *Store) Addresses() ([]net.IP, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	err := s.ensure()
	if err != nil {
		return nil, err
	}

	// we need to sort all entries first
	sorted := make(byExpiry, len(s.db))
	i := 0
	for _, e := range s.db {
		sorted[i] = e
		i++
	}

	// actural sorting
	sort.Sort(sorted)

	res := make([]net.IP, len(s.db))
	for i, e := range sorted {
		res[i] = net.ParseIP(e.IP)
	}
	return res, nil

}

// Data returns interesting data about a ip address
func (s *Store) Data(ip string) (map[string]string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	err := s.ensure()
	if err != nil {
		return nil, err
	}
	hostname := "n/a"
	hostname = s.db[ip].Hostname
	return map[string]string{"hostname": hostname}, nil
}
