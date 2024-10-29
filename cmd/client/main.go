package main

import (
    "log"
    "net/http"
    "os"

    "Video-Translation-Simulator/pkg/client"
)

func main() {
    logger := log.New(os.Stdout, "INFO: ", log.LstdFlags)
    c := client.NewClient("http://localhost:8080", logger)

    // Set up the HTTP server.
    http.HandleFunc("/status", c.HandleStatusRequest)

    serverAddress := ":9090"
    logger.Printf("Client server is starting on %s", serverAddress)
    if err := http.ListenAndServe(serverAddress, nil); err != nil {
        logger.Fatalf("Client server failed to start: %v", err)
    }
}
