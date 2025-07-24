package main

import (
	"context"
	"crypto/tls"
	"log"
	"net"

	"github.com/quic-go/quic-go"
)

func main() {
	// Remote server to connect to
	addr, err := net.ResolveUDPAddr("udp4", "localhost:6969")

	if err != nil {
		log.Fatalf("Error resolving UDP address: %v", err)
	}

	// Local UDP socket
	// The second argument is nil to make the app listen to any IP and random port
	udpConn, err := net.ListenUDP("udp4", nil)

	if err != nil {
		log.Fatalf("Error starting UDP listener: %v", err)
	}

	tr := quic.Transport{Conn: udpConn}

	cert, err := tls.LoadX509KeyPair("tls/client.crt", "tls/client.key")

	if err != nil {
		log.Fatalf("Error loading TLS certificate: %v", err)
	}

	conn, err := tr.Dial(context.Background(), addr,
		&tls.Config{
			Certificates: []tls.Certificate{cert},
			// This is just for testing, the client needs to verify the server's certificate
			InsecureSkipVerify: true,
		},
		&quic.Config{})

	if err != nil {
		log.Fatalf("Error dialing QUIC connection: %v", err)
	}

	log.Println("QUIC client connected to", conn.RemoteAddr())

	// Bidirectional stream
	str, err := conn.OpenStream()
	defer str.Close()

	if err != nil {
		log.Fatalln("Error opening stream:", err)
	}

	// conn.OpenStream doesn't send anything to the server
	// so the server doesn't know when the stream is open until
	// the client sends data
	str.Write([]byte("Hello from QUIC client!"))

	buf := make([]byte, 1024)
	n, err := str.Read(buf)
	log.Printf("Received %d bytes: %s\n", n, buf[:n])
}
