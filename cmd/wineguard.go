package main

import (
	"log"
	"os"
)

const (
	// TODO: read the following from config
	confListenPort     = 6969
	confServerAddress  = "localhost:6969"
	confClientCertPath = "tls/client.crt"
	confClientKeyPath  = "tls/client.key"
	confServerCertPath = "tls/server.crt"
	confServerKeyPath  = "tls/server.key"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: `%s {server|client} [options...]`\n", os.Args[0])
	}
	switch os.Args[1] {
	case "server":
		RunServer()
	case "client":
		RunClient()
	default:
		log.Fatalf(
			"Unknown subcommand `%s`\nUsage: `%s {server|client} [options...]`\n",
			os.Args[1],
			os.Args[0],
		)
	}
}
