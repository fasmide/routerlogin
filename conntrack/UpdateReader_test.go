package conntrack

import (
	"net"
	"os"
	"testing"
)

func TestReader(t *testing.T) {
	fd, err := os.Open("conntrack_test_file.txt")
	if err != nil {
		t.Fatalf("unable to open test file: %s", err)
	}

	reader := NewUpdateReader(fd)

	// we should now be able to read the first flow
	flow, err := reader.Read()
	if err != nil {
		t.Fatalf("unable to read flow from conntrack.Reader: %s", err)
	}

	if flow.Type != "NEW" {
		t.Fatalf("first flow of test file was not NEW: %s", flow.Type)
	}
	// skip ahead
	for index := 0; index < 2425; index++ {
		_, err := reader.Read()
		if err != nil {
			t.Fatalf("failed reading flow: %s", err)
		}
	}

	// 2427 reads in, we should find this flow
	// [DESTROY] icmp     1 src=162.243.158.119 dst=85.191.222.130 type=8 code=0 id=8825 packets=5 bytes=230 src=85.191.222.130 dst=162.243.158.119 type=0 code=0 id=8825 packets=5 bytes=230
	flow, err = reader.Read()
	if err != nil {
		t.Fatalf("failed reading flow: %s", err)
	}
	if flow.Flow.NAT != false {
		t.Fatalf("nat was expected to be false %+v", flow)
	}
	if flow.Flow.Protocol != "icmp" {
		t.Fatalf("protocol was expected to be icmp, was %s", flow.Flow.Protocol)
	}
	compareCounter := Counter{Packets: 5, Bytes: 230}
	if flow.Flow.Original.Counter != compareCounter {
		t.Fatalf("original counter packets was expected to be %+v, was %+v", compareCounter, flow.Flow.Original.Counter)
	}
	compareIP := net.ParseIP("162.243.158.119")
	if !flow.Flow.Reply.Layer3.Destination.Equal(compareIP) {
		t.Fatalf(
			"flow reply layer3 destination did not have exptected value of %s: was %s",
			compareIP,
			flow.Flow.Reply.Layer3.Destination,
		)
	}

	if flow.Flow.State != "" {
		t.Fatalf("flow state did not have expexted value of nothing: was %s", flow.Flow.State)
	}

	err = fd.Close()
	if err != nil {
		t.Fatalf("unable to close flow reader: %s", err)
	}
}
