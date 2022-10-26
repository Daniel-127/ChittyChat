package main

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"

	"github.com/Daniel-127/ChittyChat/chat"

	"google.golang.org/grpc"
)

var user string
var timestamp int32
var done = make(chan bool)
var reader *bufio.Scanner

func main() {
	timestamp = 0
	log.Print("Username:")
	reader = bufio.NewScanner(os.Stdin)
	reader.Scan()
	user = reader.Text()
	// Creat a virtual RPC Client Connection on port  9080 WithInsecure (because  of http)
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":9080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %s", err)
	}

	// Defer means: When this function returns, call this method (meaing, one main is done, close connection)
	defer conn.Close()

	//  Create new Client from generated gRPC code from proto
	c := chat.NewChatClient(conn)

	stream := joinChat(c)

	go listenOnStream(stream)
	go handleInput(c)

	<-done
}

func updateTimestamp(incomingTimestamp int32) {
	if incomingTimestamp > timestamp {
		timestamp = incomingTimestamp
	}
	timestamp++
}

func joinChat(c chat.ChatClient) chat.Chat_JoinChatClient {
	// Between the curly brackets are nothing, because the .proto file expects no input.
	timestamp++
	req := chat.UserRequest{User: user, Timestamp: timestamp}
	log.Printf("Joining chat - %d", timestamp)
	stream, err := c.JoinChat(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error when calling joining chat: %s", err)
	}

	return stream
}

func handleInput(c chat.ChatClient) {
	for {
		reader.Scan()
		input := reader.Text()
		if len(input) > 0 {
			timestamp++
			message := &chat.Message{User: user, Timestamp: timestamp, Message: input}
			log.Printf("%s: %s - %d", message.User, message.Message, message.Timestamp)
			_, err := c.PostMessage(context.Background(), message)
			if err != nil {
				log.Fatalf("Error when posting message: %s", err)
			}
		}
	}
}

func listenOnStream(stream chat.Chat_JoinChatClient) {
	for {
		response, err := stream.Recv()
		updateTimestamp(response.Timestamp)
		if err == io.EOF {
			log.Printf("Stream has been closed")
			done <- true
			return
		}
		if err != nil {
			log.Fatalf("cannot receive %v", err)
		}
		log.Printf("%s: %s - %d", response.User, response.Message, timestamp)
	}
}
