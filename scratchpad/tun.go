package scratchpad

import (
	"log"
	"slices"

	"github.com/songgao/water"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

type protocolNum int

const (
	icmpv4 protocolNum = 1
	icmpv6 protocolNum = 58
)

// Simple TUN test
// (in terminal 1:)
//
//	sudo go run ./scratchpad/tun.go
//
// (in terminal 2:)
//
//	sudo ip addr add 10.0.69.0/24 dev tun0
//	sudo ip link set dev tun0 up
//	ping 10.0.69.69
func main() {
	// Here we chose TUN so we are operating at layer 3.
	// i.e. we will be receiving IP packets
	config := water.Config{
		DeviceType: water.TUN,
	}

	iface, err := water.New(config)
	if err != nil {
		log.Fatalln(err)
	}
	defer iface.Close()

	name := iface.Name()
	log.Println("Created tunnel at", name)

	buf := make([]byte, 4096)
	// Read IP packets forever
	for {
		n, err := iface.Read(buf)
		if err != nil {
			log.Fatalln(err)
		}
		if n == 0 {
			continue
		}
		buf = buf[:n]

		// now try parsing IP
		var payload []byte
		var nextHeader protocolNum

		ipVersion := buf[0] >> 4
		// headers are parsed and printed out
		switch ipVersion {
		case 6:
			header, err := ipv6.ParseHeader(buf)
			if err == nil {
				log.Printf("Received IPv6: %+v\n", header)
				nextHeader = protocolNum(header.NextHeader)
				payload = buf[ipv6.HeaderLen:]
			}
		case 4:
			header, err := ipv4.ParseHeader(buf)
			if err == nil {
				log.Printf("Received IPv4: %+v\n", header)
				nextHeader = protocolNum(header.Protocol)
				payload = buf[ipv4.HeaderLen:]
			}
		}

		if payload != nil {
			// implement ICMP parsing so that we can look at pretty messages while we test :)
			switch nextHeader {
			case icmpv4, icmpv6:
				icmpMsg, err := icmp.ParseMessage(int(nextHeader), payload)
				if err == nil {
					log.Printf("=> ICMP message: %+v\n", icmpMsg)
				} else {
					logHexDump("=> (Unknown):", 16, payload)
				}
			default:
				logHexDump("=> (Unknown):", 16, payload)
			}
		} else {
			logHexDump("Received (unknown)", 16, buf)
		}
	}
}

func logHexDump(prefix string, bytesPerLine int, data []byte) {
	lenPrefix := max(4, len(prefix))
	lenPrefix = min(lenPrefix, 24)

	chunks := slices.Collect(slices.Chunk(data, bytesPerLine))
	log.Printf("%.*s: % x\n", lenPrefix, prefix, chunks[0])
	for i, chunk := range chunks[1:] {
		log.Printf("(%0*x): % x\n", lenPrefix-2, bytesPerLine*(i+1), chunk)
	}
}
