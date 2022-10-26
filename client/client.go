package main

import (
	"context"
	"fmt"
	"log"

	t "time"

	"github.com/Daniel-127/ChittyChat/proto"

	"google.golang.org/grpc"
)

func main() {
	// Creat a virtual RPC Client Connection on port  9080 WithInsecure (because  of http)
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":9080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %s", err)
	}

	// Defer means: When this function returns, call this method (meaing, one main is done, close connection)
	defer conn.Close()

	//  Create new Client from generated gRPC code from proto
	c := proto.NewGetCurrentTimeClient(conn)

	for {
		SendGetTimeRequest(c)
		t.Sleep(5 * t.Second)
	}
}

func SendGetTimeRequest(c proto.GetCurrentTimeClient) {
	// Between the curly brackets are nothing, because the .proto file expects no input.
	message := proto.GetTimeRequest{}

	response, err := c.GetTime(context.Background(), &message)
	if err != nil {
		log.Fatalf("Error when calling GetTime: %s", err)
	}

	fmt.Printf("Current time right now: %s\n", response.Reply)
}
