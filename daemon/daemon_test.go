package daemon

import (
	"net"
	"testing"
)

type Teststore1 struct{}
type Teststore2 struct{}

func (t *Teststore1) Addresses() ([]net.IP, error) {
	return []net.IP{
		net.ParseIP("127.0.0.1"),
	}, nil
}

func (t *Teststore2) Addresses() ([]net.IP, error) {
	return []net.IP{
		net.ParseIP("127.0.0.1"),
		net.ParseIP("127.0.0.2"),
	}, nil
}

func TestDistinctAddresses(t *testing.T) {
	daemon := Daemon{}

	daemon.AddStore(&Teststore1{})
	daemon.AddStore(&Teststore2{})

	addresses := daemon.addresses()

	// length of addresses should be exactly 2
	if len(addresses) != 2 {
		t.Fatalf("distinct addresses does not seem distinct: %+v", addresses)
	}
}
