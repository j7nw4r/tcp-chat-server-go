package main

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"log"
	"net"
	"sync"
)

type Channels struct {
	mutex          sync.Mutex
	userChannelMap map[uuid.UUID]chan string
}

var channels Channels

func main() {
	log.Printf("tcp_chat_server_go starting...")
	channels.userChannelMap = make(map[uuid.UUID]chan string)

	listener, err := net.Listen("tcp", ":23234")
	if err != nil {
		log.Fatalf("tcp_chat_server_go failed to listen: %v", err)
	}
	defer listener.Close()

	// Listing Loop
	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		conn, connErr := listener.Accept()
		if connErr != nil {
			log.Printf("tcp_chat_server_go failed to accept: %v", connErr)
			continue
		}

		go handleConnection(appCtx, conn)
	}
}

func handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	connID := uuid.New()

	channels.mutex.Lock()
	channels.userChannelMap[connID] = make(chan string)
	channels.mutex.Unlock()

	log.Printf("tcp_chat_server_go handling connection from %v", conn.RemoteAddr())
	wg := sync.WaitGroup{}
	wg.Add(2)
	go handleRead(ctx, &wg, conn, connID)
	go handleWrite(ctx, &wg, conn, connID)
	wg.Wait()
}

func handleRead(ctx context.Context, wg *sync.WaitGroup, conn net.Conn, connID uuid.UUID) {
	defer wg.Done()

	buf := make([]byte, 1024)
	for {
		if ctxErr := ctx.Err(); ctxErr != nil {
			if !errors.Is(ctxErr, context.Canceled) {
				log.Printf("tcp_chat_server_go read error: %v", ctxErr)
			}
			return
		}

		n, readErr := conn.Read(buf)
		if readErr != nil {
			log.Printf("tcp_chat_server_go failed to read: %v", readErr)
			return
		}
		message := string(buf[:n])

		channels.mutex.Lock()
		for userID, channel := range channels.userChannelMap {
			if userID == connID {
				continue
			}
			channel <- message
		}
		channels.mutex.Unlock()
	}
}

func handleWrite(ctx context.Context, wg *sync.WaitGroup, conn net.Conn, connID uuid.UUID) {
	defer wg.Done()
	channels.mutex.Lock()
	connChannel := channels.userChannelMap[connID]
	channels.mutex.Unlock()
	
	for {
		if ctxErr := ctx.Err(); ctxErr != nil {
			if !errors.Is(ctxErr, context.Canceled) {
				log.Printf("tcp_chat_server_go read error: %v", ctxErr)
			}
			return
		}
		message := <-connChannel

		_, err := conn.Write([]byte(message))
		if err != nil {
			log.Printf("tcp_chat_server_go failed to write: %v", message)
		}
	}
}
