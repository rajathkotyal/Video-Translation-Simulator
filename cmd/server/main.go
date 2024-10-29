package main

import (
		"flag"
    "log"
    "Video-Translation-Simulator/pkg/server"
)

/*
	Main Function is used as the entry point to start the server
	it is kept minimal since we are assuming a simple server with a configurable response time
	You can configure the response time using the file config.json at the root directory of this project

	Please make sure port 8080 is not already bound to another process. 
*/

func main() {
	delay := flag.Int("delay", 10, "Delay before returning final status (in seconds)")
	errorRate := flag.Int("error", 20, "Probability of returning 'error' instead of 'completed' (0-100)")

	// Parse the flags
	flag.Parse()

	// Initialize and start the server with the parsed values
	srv, err := server.NewServer(*delay, *errorRate)
	if err != nil {
			log.Fatalf("Failed to initialize server: %v", err)
	}
	if err := srv.Start(":8080"); err != nil {
			log.Fatalf("Server failed to start: %v", err)
	}
}
