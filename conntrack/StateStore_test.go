package conntrack

import "testing"

func TestStateStore(t *testing.T) {
	// notice you would maybe want to run something like
	// sudo setcap cap_net_admin+ep $(which conntrack)
	// to make tests like these succeed
	s := StateStore{}

	a, err := s.Addresses()
	if err != nil {
		t.Fatalf("statestore addresses failed: %s", err)
	}

	if len(a) == 0 {
		t.Fatalf("statestore addresses had no results")
	}
	t.Logf("Found: %+v", a)

	// lets look at some flows
	flows, err := s.StatesByIP(a[0].String())
	if err != nil {
		t.Fatalf("statestore could not find a address it provided: %s", err)
	}
	if flows[0].State == "" {
		t.Fatalf("first flow found, had no state: %+v", flows[0])
	}
}
