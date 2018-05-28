package daemon

import (
	"io"
	"io/ioutil"
	"net"
	"regexp"
	"sync"
	"testing"
	"time"
)

type Teststore1 struct{}

func (t *Teststore1) Data(i net.IP) (map[string]string, error) {
	return make(map[string]string), nil
}
func (t *Teststore1) Addresses() ([]net.IP, error) {
	return []net.IP{
		net.ParseIP("127.0.0.1"),
	}, nil
}

type Teststore2 struct{}

func (t *Teststore2) Data(i net.IP) (map[string]string, error) {
	return make(map[string]string), nil
}
func (t *Teststore2) Addresses() ([]net.IP, error) {
	return []net.IP{
		net.ParseIP("127.0.0.1"),
		net.ParseIP("127.0.0.2"),
	}, nil
}

type MockListener struct {
	buf      *MockConn
	accepted bool
}

func (m *MockListener) Accept() (net.Conn, error) {
	if m.accepted {
		var wg sync.WaitGroup
		wg.Add(1)
		wg.Wait()
	}
	m.accepted = true
	return m.buf, nil
}

func (m *MockListener) Addr() net.Addr {
	return &net.TCPAddr{}
}

func (m *MockListener) Close() error {
	return nil
}

type MockConn struct {
	io.WriteCloser
}

func (m *MockConn) Read(_ []byte) (int, error) {
	return 0, nil
}

func (m *MockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{}
}

func (m *MockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{}
}

func (m *MockConn) SetDeadline(_ time.Time) error {
	return nil
}

func (m *MockConn) SetReadDeadline(_ time.Time) error {
	return nil
}

func (m *MockConn) SetWriteDeadline(_ time.Time) error {
	return nil
}

func TestAccept(t *testing.T) {
	daemon := Daemon{}

	daemon.AddStore(&Teststore1{})
	r, w := io.Pipe()

	l := MockListener{buf: &MockConn{WriteCloser: w}}

	go daemon.Accept(&l)

	data, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatalf("Could not read response from daemon: %s", err)
	}
	// We should be able to find the ip addresses
	// from the test store
	re := regexp.MustCompile("127.0.0.1")

	if !re.MatchString(string(data)) {
		t.Fatalf("could not find expected data in output from daemon: %s", data)
	}
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
