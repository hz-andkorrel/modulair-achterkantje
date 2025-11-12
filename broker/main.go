package main

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

func main() {
	context := context.Background()

	redis := redis.NewClient(&redis.Options{
		Addr: "hub_bus:6379",
		DB:   0,
	})

	err := redis.Publish(context, "events", "hello").Err()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to publish message: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Println("Message pushed successfully")
	}
}
