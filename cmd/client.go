package main

import (
	"context"
	"crypto/tls"
	"log"
	"net"

	"github.com/quic-go/quic-go"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp4", "localhost:6969")

	if err != nil {
		log.Fatalf("Error resolving UDP address: %v", err)
	}

	udpConn, err := net.ListenUDP("udp4", nil)

	if err != nil {
		log.Fatalf("Error starting UDP listener: %v", err)
	}

	tr := quic.Transport{Conn: udpConn}

	conn, err := tr.Dial(context.Background(), addr, &tls.Config{}, &quic.Config{})

	if err != nil {
		log.Fatalf("Error dialing QUIC connection: %v", err)
	}

	log.Println("QUIC client connected to", conn.RemoteAddr())

	str, err := conn.OpenStream()
	defer str.Close()

	if err != nil {
		log.Fatalln("Error opening stream:", err)
	}

	str.Write([]byte("Hello from QUIC client!"))

	var buf []byte
	n, err := str.Read(buf)
	log.Printf("Received %d bytes: %s\n", n, buf[:n])
}

