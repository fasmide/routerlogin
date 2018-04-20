package dnsmasq

import (
	"os"
	"strings"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	fd, err := os.Open("dnsmasq_test.leases")
	if err != nil {
		t.Logf("could not open test.leases: %s", err)
		t.FailNow()
	}

	data, err := Parse(fd)
	if err != nil {
		t.Logf("failed to parse test leases: %s", err)
		t.FailNow()
	}

	if data[2].IP != "192.168.1.111" {
		t.Fail()
	}

	fd.Close()
}

func TestInvalidDateUnmarshal(t *testing.T) {
	reader := strings.NewReader("blarp asd asd asd asd asd asd")

	_, err := Parse(reader)
	if err == nil {
		t.Logf("Parse failed to fail on invalid date data")
		t.FailNow()
	}
}

func TestInvalidFormatUnmarshal(t *testing.T) {
	reader := strings.NewReader("a b c")

	_, err := Parse(reader)
	if err == nil {
		t.Logf("Parse failed to fail on invalid input")
		t.FailNow()
	}
}
