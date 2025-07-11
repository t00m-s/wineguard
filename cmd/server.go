package main

import (
	"context"
	"crypto/tls"
	"log"
	"net"

	"github.com/quic-go/quic-go"
)

func main() {
	udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: 6969})

	if err != nil {
		log.Fatalln("Error starting UDP listener:", err)
	}

	tr := quic.Transport{Conn: udpConn}

	ln, err := tr.Listen(&tls.Config{}, &quic.Config{})

	if err != nil {
		log.Fatalln("Error starting QUIC listener:", err)
	}

	log.Println("QUIC server is listening on", ln.Addr())

	for {
		conn, err := ln.Accept(context.Background())
		log.Println("Accepted new connection from", conn.RemoteAddr())

		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		go func() {
			str, err := conn.AcceptStream(context.Background())
			defer str.Close()

			if err != nil {
				log.Println("Error accepting stream:", err)
				return
			}

			str.Write([]byte("Hello from QUIC server!"))

			var buf []byte
			n, err := str.Read(buf)

			if err != nil {
				log.Println("Error reading from stream:", err)
				return
			}

			log.Printf("Received %d bytes: %s\n", n, buf[:n])

		}()
	}
}
