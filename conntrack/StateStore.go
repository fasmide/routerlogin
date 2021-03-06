package conntrack

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

// StateStore stores information about the current conntrack state
type StateStore struct {
	db map[string][]*Flow

	lock         sync.Mutex
	lastPopulate time.Time
}

// ensure updates the database if needed
func (s *StateStore) ensure() error {
	if time.Now().Sub(s.lastPopulate) > time.Second*5 {
		s.db = make(map[string][]*Flow)
		return s.populate()
	}

	return nil
}

// populate populates the database which is expected to be empty
func (s *StateStore) populate() error {
	command := exec.Command("conntrack", "-L")

	input, err := command.StdoutPipe()
	if err != nil {
		return err
	}
	defer input.Close()

	// errorReader is read if someting fails down the line
	errorReader, err := command.StderrPipe()
	if err != nil {
		return err
	}
	defer errorReader.Close()

	var stderr []byte

	go func() {
		stderr, _ = ioutil.ReadAll(errorReader)
	}()

	// our flow reader
	r := NewReader(input)

	err = command.Start()
	if err != nil {
		return err
	}

	for {
		var flow *Flow
		flow, err = r.Read()

		// stop on error
		if err != nil {
			break // and let whatever comes next handle the error
		}

		// we dont need knowledge about non-natted flows
		if !flow.NAT {
			continue
		}

		// we always use the original direction source as our index
		index := flow.Original.Layer3.Source.String()

		// append flow if flow slice exists
		if _, exists := s.db[index]; exists {
			s.db[index] = append(s.db[index], flow)
			continue
		}

		// Insert new flow slice
		s.db[index] = []*Flow{flow}
	}
	// if the error is not nil and also is not an EOF error - we have a problem
	if err != io.EOF && err != nil {
		return err
	}

	// wait for command to exit and check for non status 0 codes
	err = command.Wait()
	if err != nil {
		return fmt.Errorf("conntrack error: %s: %s", err, stderr)
	}

	log.Printf("conntrack.StateStore: updated store with %d entrys", len(s.db))
	s.lastPopulate = time.Now()

	return nil
}

// Addresses returns a sorted slice of ip addresses found
func (s *StateStore) Addresses() ([]net.IP, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// our cache should be to date
	err := s.ensure()
	if err != nil {
		return nil, err
	}

	res := make([]net.IP, len(s.db))
	i := 0
	for ip := range s.db {
		res[i] = net.ParseIP(ip)
		i++
	}
	return res, nil
}

// Data will return stuff about an ip address that we find interesting
func (s *StateStore) Data(ip string) (map[string]string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	err := s.ensure()
	if err != nil {
		return nil, err
	}

	return map[string]string{"nFlows": strconv.Itoa(len(s.db[ip]))}, nil
}

// StatesByIP returns all flows from a given ip
func (s *StateStore) StatesByIP(ip string) ([]*Flow, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// ensure the database is up to date
	err := s.ensure()
	if err != nil {
		return nil, err
	}

	if flow, found := s.db[ip]; found {
		return flow, nil
	}

	return nil, fmt.Errorf("no flows found")
}
