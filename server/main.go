package main

import (
	"context"
	"fmt"
	"grpc-tutorial/chatpb"
	"io"
	"net"

	// pb "grpc-tutorial/chatpb"
	"log"
	"sync"

	"google.golang.org/grpc"
)

type server struct {
	chatpb.UnimplementedChatServiceServer
	mu sync.Mutex
	// clients map[string][]chatpb.ChatService_JoinRoomServer
	// clients map[chatpb.ChatService_ChatStreamServer]bool
	rooms map[string][]chatpb.ChatService_JoinRoomServer
}

func NewServer() *server {
	return &server{
		rooms: make(map[string][]chatpb.ChatService_JoinRoomServer),
	}
}
func (s *server) ChatStream(stream chatpb.ChatService_ChatStreamServer) error {
	fmt.Println("chat stream working")
	s.mu.Lock()
	defer s.mu.Unlock()
	var currentRoom string
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("msg stream EOF")
			return nil
		}
		if err != nil {
			return err
		}
		// If the room_id is provided, use it to route the message
		if msg.RoomId != "" {
			currentRoom = msg.RoomId
			s.rooms[currentRoom] = append(s.rooms[currentRoom], stream)
		}
		// Broadcast the message to the correct room
		for _, s := range s.rooms[currentRoom] {
			if err := s.Send(msg); err != nil {
				log.Printf("Failed to send message to room %s: %v", currentRoom, err)
			}
		}
	}
}

func (s *server) JoinRoom(req *chatpb.JoinRoomRequest, stream chatpb.ChatService_JoinRoomServer) error {
	s.mu.Lock()
	defer func() {
		fmt.Println("Unlocking JoinRoom")
		s.mu.Unlock()
	}()
	fmt.Println("Locked successfully, JoinRoom...")

	roomID := req.GetRoomId()
	s.rooms[roomID] = append(s.rooms[roomID], stream)
	fmt.Printf("attempting to welcome user... %s \n", roomID)
	welcomeMsg := &chatpb.ChatMessage{
		RoomId:  roomID,
		Message: "A new user has joined the room!",
	}
	s.BroadcastMessage(context.Background(), welcomeMsg)

	<-stream.Context().Done()
	s.removeStreamFromRoom(roomID, stream)

	// for {
	// 	select {
	// 	case <-stream.Context().Done():
	// 		// Client closed the stream
	// 		s.removeStreamFromRoom(roomID, stream)
	// 		return nil
	// 	}
	// }
	return nil
}

// Helper function to remove a stream from a room
func (s *server) removeStreamFromRoom(roomID string, stream chatpb.ChatService_JoinRoomServer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	streams := s.rooms[roomID]
	for i, st := range streams {
		if st == stream {
			// Remove the stream from the list
			s.rooms[roomID] = append(streams[:i], streams[i+1:]...)
			break
		}
	}
	fmt.Printf("Closing room... \n")
}

func (s *server) BroadcastMessage(ctx context.Context, msg *chatpb.ChatMessage) {
	roomID := msg.GetRoomId()
	streams, ok := s.rooms[roomID]
	if !ok {
		log.Printf("Room %s does not exist or has no clients", roomID)
		return
	}

	for _, stream := range streams {
		if err := stream.Send(msg); err != nil {
			log.Printf("Error sending message to client: %v", err)
		}
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	chatpb.RegisterChatServiceServer(grpcServer, NewServer())

	fmt.Println("Chat server started on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// func (s *server) LeaveRoom(req *chatpb.LeaveRoomRequest, stream chatpb.ChatService_LeaveRoomServer) error {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	roomID := req.GetRoomId()

// 	streams, ok := s.rooms[roomID]
// 	if !ok {
// 		return fmt.Errorf("room %s not found", roomID)
// 	}

// 	for i, sr := range streams {
// 		if sr == stream {
// 			s.rooms[roomID] = append(streams[:i], streams[i+1:]...)
// 			break
// 		}
// 	}

// 	if len(s.rooms[roomID]) == 0 {
// 		delete(s.rooms, roomID)
// 	}
// 	if err := stream.Send(&chatpb.ChatMessage{Message: "Left room: " + roomID}); err != nil {
// 		return err
// 	}
// 	return nil
// }
