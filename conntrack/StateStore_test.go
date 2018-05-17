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
}
