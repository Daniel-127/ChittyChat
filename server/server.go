package main

import (
	"context"
	"log"
	"net"

	"github.com/Daniel-127/ChittyChat/chat"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	chat.UnimplementedChatServer
}

var timestamp int32
var connections = make(map[string]*ClientConnection)

func (s *Server) JoinChat(req *chat.UserRequest, stream chat.Chat_JoinChatServer) error {
	updateTimestamp(req.Timestamp)
	log.Printf("%s joined the server - %d", req.User, timestamp)
	close := make(chan bool)
	connections[req.User] = &ClientConnection{stream: stream, close: make(chan bool)}
	publishMessageToAllClients(&chat.Message{User: "Server", Message: req.User + " joined the chat"})
	select {
	case <-stream.Context().Done(): //Forcefully closed
		req.Timestamp = timestamp
		disconnectFromChat(req)
		return status.Error(codes.Canceled, "Stream was forcefully closed")
	case <-close: //Peacefully closed
		return status.Error(codes.OK, "Stream has been closed")
	}
}

func (s *Server) PostMessage(ctx context.Context, msg *chat.Message) (*chat.Empty, error) {
	updateTimestamp(msg.Timestamp)
	log.Printf("%s: %s - %d", msg.User, msg.Message, timestamp)
	publishMessageToAllClients(msg)
	return &chat.Empty{}, nil
}

func (s *Server) LeaveChat(ctx context.Context, req *chat.UserRequest) (*chat.Empty, error) {
	disconnectFromChat(req)
	connections[req.User].close <- true
	return &chat.Empty{}, nil
}

func disconnectFromChat(req *chat.UserRequest) {
	updateTimestamp(req.Timestamp)
	log.Printf("%s left the chat - %d", req.User, timestamp)
	delete(connections, req.User)
	publishMessageToAllClients(&chat.Message{User: req.User, Message: "Left the chat"})
}

func publishMessageToAllClients(msg *chat.Message) {
	firstPublish := true
	for user, conn := range connections {
		if user != msg.User {
			if firstPublish {
				firstPublish = false
				timestamp++
				msg.Timestamp = timestamp
				log.Printf("Publising message to all users - %d", timestamp)
			}
			err := conn.stream.Send(msg)
			if err != nil {
				log.Fatalf("Failed to send %v", err)
			}
		}
	}
}

func updateTimestamp(incomingTimestamp int32) {
	if incomingTimestamp > timestamp {
		timestamp = incomingTimestamp
	}
	timestamp++
}

func main() {
	timestamp = 0
	// Create listener tcp on port 9080
	list, err := net.Listen("tcp", ":9080")
	if err != nil {
		log.Fatalf("Failed to listen on port 9080: %v", err)
	}
	grpcServer := grpc.NewServer()
	chat.RegisterChatServer(grpcServer, &Server{})
	log.Printf("Chat server is running...")

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to server %v", err)
	}
}

type ClientConnection struct {
	stream chat.Chat_JoinChatServer
	close  chan bool
}
