package main

import (
	"fmt"
	"log"
	"net"
	"sync"

	pb "grpc-tuto-2/chatpb"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedChatServiceServer
	mu      sync.Mutex
	clients map[pb.ChatService_ChatStreamServer]bool
}

func (s *server) ChatStream(stream pb.ChatService_ChatStreamServer) error {
	s.mu.Lock()
	s.clients[stream] = true
	s.mu.Unlock()

	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			break
		}
		log.Printf("Received message from %s: %s", msg.User, msg.Message)

		// Broadcast the message to all clients
		s.broadcastMessage(msg)
	}

	s.mu.Lock()
	delete(s.clients, stream)
	s.mu.Unlock()

	return nil
}

func (s *server) broadcastMessage(msg *pb.ChatMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for client := range s.clients {
		err := client.Send(msg)
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterChatServiceServer(grpcServer, &server{
		clients: make(map[pb.ChatService_ChatStreamServer]bool),
	})

	fmt.Println("Chat server started on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
