package main

import (
	"hotelhub/broker/infrastructure"
	"hotelhub/broker/services"
)

// The entry point of the broker service start by loading the configuration from the .env file.
// The JWT service is initialized for handling authentication.
// The router is set up with all the necessary routes and middleware, and the server is started.
func main() {
	configuration := services.NewConfiguration()
	jwtService := services.NewJwtService(configuration)
	repositoryStrategy := services.NewPostgresRepositoryStrategy()

	router := infrastructure.NewRouter(configuration, jwtService, repositoryStrategy)
	router.Run()
}
