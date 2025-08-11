package main

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
	"github.com/songgao/water"
)

func RunServer() {
	iface, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		log.Fatalf("Could not create tunnel: (%T) %v\n", err, err)
	}
	// ideally closed only when program dies
	defer iface.Close()

	log.Println("Created tunnel at", iface.Name())
	// Local UDP socket. Listen to any IP and port 6969
	udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: confListenPort})

	if err != nil {
		log.Fatalf("Error starting UDP listener: (%T) %v\n", err, err)
	}

	tr := quic.Transport{Conn: udpConn}

	cert, err := tls.LoadX509KeyPair(confServerCertPath, confServerKeyPath)

	if err != nil {
		log.Fatalf("Error loading TLS certificate: (%T) %v\n", err, err)
	}

	ln, err := tr.Listen(&tls.Config{
		Certificates: []tls.Certificate{cert},
	}, &quic.Config{})

	if err != nil {
		log.Fatalf("Error starting QUIC listener: (%T) %v\n", err, err)
	}

	log.Println("QUIC server is listening on", ln.Addr())

	for {
		// Infinite loop to accept new connections
		conn, err := ln.Accept(context.Background())
		if err != nil {
			log.Printf("Error accepting connection: (%T) %v\n", err, err)
			continue
		}
		log.Println("Accepted new connection from", conn.RemoteAddr())

		str, err := conn.AcceptStream(conn.Context())

		if err != nil {
			log.Printf("Error accepting stream: (%T) %v\n", err, err)
			continue
		}
		log.Println("Accepted new stream from", str.StreamID())

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			// incoming packets on `str` (stream of QUIC connection) get piped into `iface` (dev tun)
			_, err := io.Copy(iface, str)
			if err != nil {
				log.Printf("Error in receive stream: (%T) %v\n", err, err)
			}

		}()
		go func() {
			defer wg.Done()
			// incoming packets on `iface` get piped into `str`
			io.Copy(str, iface)
			if err != nil {
				log.Printf("Error in send stream: (%T) %v\n", err, err)
			}
		}()

		wg.Wait()
		str.Close()
	}
}
