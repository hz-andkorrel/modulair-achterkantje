package main

import (
	"hotelhub/broker/handlers"
	"hotelhub/broker/listener"
	"hotelhub/broker/repository"
	"hotelhub/broker/services"
)

// The entry point of the broker service start by loading the configuration from the .env file.
// The JWT service is initialized for handling authentication.
// The router is set up with all the necessary routes and middleware, and the server is started.
func main() {
	configuration := services.NewConfiguration()
	jwtService := services.NewJwtService(configuration)
	database := services.NewPostgres(configuration)
	repositoryStrategy := repository.NewPostgresRepositoryStrategy(database)

	eventlistener := listener.NewEventBusListener(repositoryStrategy)
	go eventlistener.Start()

	router := handlers.NewRouter(configuration, jwtService, repositoryStrategy)
	router.Run()
}
