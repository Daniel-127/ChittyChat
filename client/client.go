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

func main() {
	timestamp = 0
	log.Printf("Username: ")
	user, _ = bufio.NewReader(os.Stdin).ReadString('\n')
	user = user[0 : len(user)-1]
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

	stream, err := c.JoinChat(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error when calling JoinChat: %s", err)
	}

	return stream
}

func handleInput(c chat.ChatClient) {
	for {
		log.Print(">> ")
		input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		input = input[0 : len(input)-1]
		timestamp++
		message := &chat.Message{User: user, Timestamp: timestamp, Message: input}
		_, err := c.PostMessage(context.Background(), message)
		if err != nil {
			log.Fatalf("Error when posting message: %s", err)
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
		log.Printf("%s: %s - %d", response.User, response.Message, response.Timestamp)
	}
}
