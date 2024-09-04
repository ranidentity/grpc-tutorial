package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"grpc-tutorial/chatpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// conn, err := grpc.NewClient("localhost:50051")
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()

	client := chatpb.NewChatServiceClient(conn)

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter room: ")
	scanner.Scan()
	roomID := scanner.Text()

	joinRoom(client, roomID)
}

func joinRoom(client chatpb.ChatServiceClient, roomID string) {
	req := &chatpb.JoinRoomRequest{
		RoomId: roomID,
	}

	stream, err := client.JoinRoom(context.Background(), req)
	if err != nil {
		log.Fatalf("Error joining room: %v", err)
	}
	// Mutex to ensure synchronized access to the stream
	var mu sync.Mutex

	// Goroutine to receive messages from the server
	go func() {
		for {
			in, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					log.Println("Stream closed by server.")
					return
				}
				log.Fatalf("Error receiving message: %v", err)
			}
			log.Printf("%s: %s", in.User, in.Message)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		msg := &chatpb.ChatMessage{
			RoomId:    roomID,
			User:      "ClientName", // Replace with actual user name
			Message:   text,
			Timestamp: time.Now().Unix(),
		}
		mu.Lock()
		// Assuming you have a method to send a message in the room (like ChatStream)
		if err := stream.SendMsg(msg); err != nil {
			mu.Unlock()
			if err == io.EOF {
				log.Println("Stream closed by server. Cannot send message.")
				return
			}
			log.Fatalf("Error sending message: %v", err)
		}
		mu.Unlock()
	}

	// Close the stream after sending all messages
	mu.Lock()
	if err := stream.CloseSend(); err != nil {
		mu.Unlock()
		log.Fatalf("Failed to close send stream: %v", err)
	}
	mu.Unlock()

	// Handle scanner errors
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading from input: %v", err)
	}
}
