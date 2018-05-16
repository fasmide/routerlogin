package conntrack

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// Flow represents a tracked connection
type Flow struct {

	// Original direction
	Original Direction

	// Reply direction
	Reply Direction

	TTL           int
	State         string // ASSURED, UNREPLIED
	Protocol      string // tcp, udp, imcp....
	ProtocolState string // ESTABLISHED, CLOSE_WAIT etc

	// NAT is not really a conntrack thing, we just check if the original
	// and reply directions match each others ip addresses for convenience
	NAT bool
}

// Direction describes our layer 3 and 4 information, in a given direction
type Direction struct {
	Layer3  Layer3
	Layer4  Layer4
	Counter Counter
}

// Layer3 represents data of the layer 3 OSI stack
type Layer3 struct {
	Source      net.IP
	Destination net.IP
}

// Layer4 represents data of the layer 4 OSI stack
type Layer4 struct {
	SPort uint16
	DPort uint16
}

// Counter represents the accumulated package and byte count for this direction..
type Counter struct {
	Packets uint
	Bytes   uint
}

// ParseFlowLine parses output from conntrack e.g.
// tcp      6 300 ESTABLISHED src=192.168.1.191 dst=192.168.1.1 sport=35786 dport=22 packets=4378 bytes=240025 src=192.168.1.1 dst=192.168.1.191 sport=22 dport=35786 packets=4727 bytes=1455593 [ASSURED] mark=0 use=1
// tcp      6 29 CLOSE_WAIT src=192.168.1.191 dst=52.222.168.153 sport=49746 dport=443 packets=50 bytes=20898 src=52.222.168.153 dst=85.191.222.130 sport=443 dport=49746 packets=48 bytes=13171 [ASSURED] mark=0 use=1
// tcp      6 431884 ESTABLISHED src=192.168.1.244 dst=216.58.213.202 sport=42412 dport=443 packets=18 bytes=2272 src=216.58.213.202 dst=85.191.222.130 sport=443 dport=42412 packets=22 bytes=15245 [ASSURED] mark=0 use=1
// udp      17 156 src=192.168.1.76 dst=209.206.58.5 sport=44017 dport=7351 packets=16330 bytes=2287570 src=209.206.58.5 dst=85.191.222.130 sport=7351 dport=44017 packets=16106 bytes=1205484 [ASSURED] mark=0 use=1
// udp      17 19 src=192.168.1.149 dst=239.255.255.250 sport=45162 dport=1900 packets=3 bytes=1340 [UNREPLIED] src=239.255.255.250 dst=192.168.1.149 sport=1900 dport=45162 packets=0 bytes=0 mark=0 use=1
func ParseFlowLine(s string) (Flow, error) {
	flow := Flow{}

	parts := strings.Fields(s)
	if len(parts) == 0 {
		return flow, fmt.Errorf("Supplied string had no fields: \"%s\"", s)
	}
	// we use this index to jump in our parts
	index := 0

	// part 0 is Protocol
	flow.Protocol = parts[index]
	index++

	// part 1 is protocol in decimal...
	// which we dont use
	index++

	// part 2 is ttl (unless we are parsing events - when there is no ttl)
	if !strings.HasPrefix(parts[index], "src=") {

		ttl, err := strconv.Atoi(parts[index])
		if err != nil {
			return flow, fmt.Errorf("Unable to parse ttl from conntrack: %s could not be parsed as integer: %s", parts[2], err)
		}
		flow.TTL = ttl
		index++
	}

	// if we where talking tcp, we have a special field in here
	// i dont know if other protocols also have this special field so, we
	// are looking for the begining of layer3-4 instead
	if !strings.HasPrefix(parts[index], "src=") {
		flow.ProtocolState = parts[index]
		index++
	}

	// the next parts are layer3-4 info, we have a special function for these
	offset, err := parseLayer3And4(parts[index:], &flow.Original)
	if err != nil {
		return flow, fmt.Errorf("Unable to parse layer 3 and 4 from line %+v: %s", parts, err)
	}
	index = index + offset

	// this part is usually "[UNREPLIED]"
	// but if it has a prefix of src= - move along
	if !strings.HasPrefix(parts[index], "src=") {
		flow.State = parts[index]
		index++
	}

	// then we should get back to our reply layer3-4
	offset, err = parseLayer3And4(parts[index:], &flow.Reply)
	if err != nil {
		return flow, fmt.Errorf("Unable to parse layer 3 and 4 from line: %+v: %s", parts, err)
	}
	index = index + offset

	// If conntrack does not expect the reply to be received by the source - we properly have NAT
	if !flow.Original.Layer3.Source.Equal(flow.Reply.Layer3.Destination) {
		flow.NAT = true
	}

	// are we at the end?
	if index == len(parts) {
		return flow, nil
	}

	// not at the end huh? - we could be assured - lets check
	if parts[index] == "[ASSURED]" {
		flow.State = "ASSURED"
	}

	// That is all

	return flow, nil
}

// we will be parsing slices of parts such as:
// src=192.168.1.149 dst=239.255.255.250 sport=45162 dport=1900 packets=3 bytes=1340
func parseLayer3And4(s []string, d *Direction) (int, error) {
	// We return length so that the call'er is able to figure out how much we parsed
	length := 0

	// our line always starts with src, if we encounter another src, we have gone too far
	var srcPassed bool
	for _, s := range s {
		parts := strings.Split(s, "=")
		switch parts[0] {
		case "src":
			if srcPassed {
				// we have gone too far
				return length, nil
			}

			d.Layer3.Source = net.ParseIP(parts[1])
			length++
			srcPassed = true

		case "dst":
			d.Layer3.Destination = net.ParseIP(parts[1])
			length++
		case "dport":
			dport, err := strconv.ParseUint(parts[1], 10, 16)
			if err != nil {
				return length, fmt.Errorf("Conntrack: Parselayer3-4: Unable to parse dport: %s", err)
			}
			d.Layer4.DPort = uint16(dport)
			length++
		case "sport":
			sport, err := strconv.ParseUint(parts[1], 10, 16)
			if err != nil {
				return length, fmt.Errorf("Conntrack: Parselayer3-4: Unable to parse sport: %s", err)
			}
			d.Layer4.SPort = uint16(sport)
			length++
		case "packets":

			packets, err := strconv.ParseUint(parts[1], 10, 64)
			if err != nil {
				return length, fmt.Errorf("Parselayer3-4: Unable to parse packets: %s", err)
			}
			d.Counter.Packets = uint(packets)
			length++
		case "bytes":
			bytes, err := strconv.ParseUint(parts[1], 10, 64)
			if err != nil {
				return length, fmt.Errorf("Parselayer3-4: Unable to parse bytes: %s", err)
			}
			d.Counter.Bytes = uint(bytes)
			length++
		// type, id and code are only for icmp traffic - lets just skip it for now
		case "type":
			length++
		case "id":
			length++
		case "code":
			length++
		default:
			// we have encountered a field we dont know and we should be pretty sure we have no more fields left...
			return length, nil
		}
	}
	return length, nil
}
