package main

import (
	"context"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"

	"google.golang.org/grpc"

	desc "github.com/RikiTikiTavee17/course/grpc/pkg/dish_v1"
)

const (
	address = "localhost:50051"
	noteId  = 12
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to cinnect to server %v", err)
	}
	defer conn.Close()

	c := desc.NewNoteV1Client(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.Get(ctx, &desc.GetRequest{Id: noteId})
	if err != nil {
		log.Fatalf("failed to get note %v", err)
	}

	log.Printf("Note info:%v", r.GetNote())
}
