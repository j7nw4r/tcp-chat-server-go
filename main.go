package main

import (
	"log"
	"net"
)

var userChannelMap = make(map[string]chan string)

func main() {
	log.Printf("tcp_chat_server_go starting...")

	listener, err := net.Listen("tcp", ":23234")
	if err != nil {
		log.Fatalf("tcp_chat_server_go failed to listen: %v", err)
	}
	defer listener.Close()

	// Listing Loop
	for {
		conn, connErr := listener.Accept()
		if connErr != nil {
			log.Printf("tcp_chat_server_go failed to accept: %v", connErr)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	log.Printf("tcp_chat_server_go handling connection from %v", conn.RemoteAddr())
	conn.Write([]byte("testing"))

	defer conn.Close()
}
