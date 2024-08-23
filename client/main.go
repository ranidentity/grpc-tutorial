package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"time"

	pb "grpc-tuto-2/chatpb"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewChatServiceClient(conn)

	stream, err := client.ChatStream(context.Background())
	if err != nil {
		log.Fatalf("Error creating stream: %v", err)
	}

	go func() {
		for {
			in, err := stream.Recv()
			if err != nil {
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

		msg := &pb.ChatMessage{
			User:      "ClientName", // Replace with actual user name
			Message:   text,
			Timestamp: time.Now().Unix(),
		}

		if err := stream.Send(msg); err != nil {
			log.Fatalf("Error sending message: %v", err)
		}
	}
}
