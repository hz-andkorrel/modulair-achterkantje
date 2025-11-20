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

	subscription := redis.Subscribe(context, "events", "hotel.events")
	channel := subscription.Channel()

	fmt.Println("Listening for messages on 'events' and 'hotel.events' channels...")
	for msg := range channel {
		fmt.Printf("📨 [%s] %s\n", msg.Channel, msg.Payload)
	}
}
