package dnsmasq

import "testing"

func TestIPLookup(t *testing.T) {
	store := Store{Path: "dnsmasq_test.leases"}

	lease, err := store.LeaseByIP("192.168.1.132")
	if err != nil {
		t.Fatalf("failed looking up ip address: %s", err)
	}

	if lease.Mac != "00:aa:bb:cc:dd:ee" {
		t.Fatalf("lease lookup found wrong entry: %+v", lease)
	}

	populated := store.lastPopulate

	// When looking up a lease again, lastPopulate should not have changed
	_, _ = store.LeaseByIP("192.168.1.132")
	if store.lastPopulate != populated {
		t.Fatalf("store did not cache results")
	}
}

func TestWrongPath(t *testing.T) {
	store := Store{Path: "blarh.txt"}

	_, err := store.LeaseByIP("10.0.0.1")
	if err == nil {
		t.Fatalf("found lease when using non existing leases path")
	}
}

func TestMalformedLeasesFile(t *testing.T) {
	store := Store{Path: "/etc/passwd"}

	_, err := store.LeaseByIP("127.0.0.1")
	if err == nil {
		t.Fatalf("found lease when using /etc/passwd as leases path")
	}
}
