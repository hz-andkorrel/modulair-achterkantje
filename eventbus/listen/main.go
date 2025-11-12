package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func main() {
	context := context.Background()

	redis := redis.NewClient(&redis.Options{
		Addr: "hub_bus:6379",
		DB:   0,
	})

	subscription := redis.Subscribe(context, "events")
	channel := subscription.Channel()

	fmt.Println("Listening for messages on 'events' channel...")
	for msg := range channel {
		fmt.Printf("Received message: %s\n", msg.Payload)
	}
}
