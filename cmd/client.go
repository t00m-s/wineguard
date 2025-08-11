package main

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"net"
	"slices"
	"sync"

	"github.com/quic-go/quic-go"
	"github.com/songgao/water"
)

func RunClient() {
	iface, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		log.Fatalf("Could not create tunnel: (%T) %v\n", err, err)
	}
	defer iface.Close()

	log.Println("Created tunnel at", iface.Name())
	// Remote server to connect to
	addr, err := net.ResolveUDPAddr("udp4", confServerAddress)

	if err != nil {
		log.Fatalf("Error resolving UDP address: (%T) %v\n", err, err)
	}

	// Local UDP socket
	// The second argument is nil to make the app listen to any IP and random port
	udpConn, err := net.ListenUDP("udp4", nil)

	if err != nil {
		log.Fatalf("Error starting UDP listener: (%T) %v\n", err, err)
	}

	tr := quic.Transport{Conn: udpConn}

	cert, err := tls.LoadX509KeyPair(confClientCertPath, confClientKeyPath)

	if err != nil {
		log.Fatalf("Error loading TLS certificate: (%T) %v\n", err, err)
	}

	conn, err := tr.Dial(context.Background(), addr,
		&tls.Config{
			Certificates: []tls.Certificate{cert},
			// This is just for testing, the client needs to verify the server's certificate
			InsecureSkipVerify: true,
		},
		&quic.Config{})

	if err != nil {
		log.Fatalf("Error dialing QUIC connection: (%T) %v\n", err, err)
	}

	log.Println("QUIC client connected to", conn.RemoteAddr())

	// Bidirectional stream
	str, err := conn.OpenStream()
	if err != nil {
		log.Fatalf("Error opening stream: (%T) %v\n", err, err)
	}
	defer str.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		// incoming packets on `str` (stream of QUIC connection) get piped into `iface` (dev tun)
		writer := io.MultiWriter(iface, logHexWriter{"Recv"})
		_, err := io.Copy(writer, str)
		if err != nil {
			log.Printf("Error in receive stream: (%T) %v\n", err, err)
		}
	}()
	go func() {
		defer wg.Done()
		// incoming packets on `iface` get piped into `str`
		writer := io.MultiWriter(str, logHexWriter{"Send"})
		_, err := io.Copy(writer, iface)
		if err != nil {
			log.Printf("Error in send stream: (%T) %v\n", err, err)
		}
	}()

	wg.Wait()
	log.Println("Connection closed. Bye")
}

type logHexWriter struct {
	Prefix string
}

func (w logHexWriter) Write(buf []byte) (n int, err error) {
	logHexDump(w.Prefix, 16, buf)
	return len(buf), nil
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
