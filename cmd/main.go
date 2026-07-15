package main

import (
	"context"
	"event_proxy/settings"
	"fmt"
	"log"
)

func main() {

	configuration, err := settings.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	if err := configuration.Start(context.TODO()); err != nil {
		fmt.Println("Failed to start app:", err)

	}
	fmt.Println("server running at http://localhost:8090")

	
}
