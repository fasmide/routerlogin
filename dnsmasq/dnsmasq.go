package dnsmasq

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"time"
)

// Entry represents an dnsmasq lease
type Entry struct {
	Expiry   time.Time
	Mac      string
	IP       string
	Hostname string
	ClientID string
}

// Parse parses an dnsmasq.leases file
func Parse(in io.Reader) ([]Entry, error) {
	s := bufio.NewScanner(in)

	list := make([]Entry, 0)

	for s.Scan() {
		var t string
		var e Entry

		_, err := fmt.Sscanf(s.Text(), "%s %s %s %s %s", &t, &e.Mac, &e.IP, &e.Hostname, &e.ClientID)
		if err != nil {
			return nil, err
		}

		tUnix, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return nil, err
		}
		e.Expiry = time.Unix(tUnix, 0)
		list = append(list, e)

	}

	return list, nil
}
