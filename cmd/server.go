package main

import (
	"context"
	"crypto/tls"
	"log"
	"net"

	"github.com/quic-go/quic-go"
)

func main() {
	// Local UDP socket. Listen to any IP and port 6969
	udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: 6969})

	if err != nil {
		log.Fatalln("Error starting UDP listener:", err)
	}

	tr := quic.Transport{Conn: udpConn}

	cert, err := tls.LoadX509KeyPair("tls/server.crt", "tls/server.key")

	if err != nil {
		log.Fatalln("Error loading TLS certificate:", err)
	}

	ln, err := tr.Listen(&tls.Config{
		Certificates: []tls.Certificate{cert},
	}, &quic.Config{})

	if err != nil {
		log.Fatalln("Error starting QUIC listener:", err)
	}

	log.Println("QUIC server is listening on", ln.Addr())

	for {
		// Infinite loop to accept new connections
		conn, err := ln.Accept(context.Background())
		log.Println("Accepted new connection from", conn.RemoteAddr())

		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		
		str, err := conn.AcceptStream(conn.Context())

		if err != nil {
			log.Println("Error accepting stream:", err)
			continue
		}
		log.Println("Accepted new stream from", str.StreamID())

		// The stream is handled in a goroutine
		go func() {
			defer str.Close()

			str.Write([]byte("Hello from QUIC server!"))

			buf := make([]byte, 1024)
			n, err := str.Read(buf)

			if err != nil {
				log.Println("Error reading from stream:", err)
				return
			}

			log.Printf("Received %d bytes: %s\n", n, buf[:n])

		}()
	}
}
