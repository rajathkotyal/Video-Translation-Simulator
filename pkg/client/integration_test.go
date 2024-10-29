package client

import (
    "encoding/json"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "testing"
    "time"
)

func TestClientHandleStatusRequest(t *testing.T) {
    logger := log.New(os.Stdout, "TestLog: ", log.LstdFlags)
    c := NewClient("http://localhost:8080", logger)

    // First we initialize a test client server
    server := http.Server{
        Addr:    ":9090",
        Handler: http.HandlerFunc(c.HandleStatusRequest),
    }

    // Start the server in a seperate goroutine to keep it open to requests without blockers
    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatalf("ListenAndServe(): %v", err)
        }
    }()
    defer server.Close()

    client := &http.Client{}

    // Simulating user requests. User may click repeatedly in the beginning
    // Initial rapid requests.
    for i := 0; i < 100; i++ {
        resp, err := client.Get("http://localhost:9090/status")
        if err != nil {
            t.Fatalf("Request failed: %v", err)
        }
        resp.Body.Close()
        time.Sleep(10 * time.Millisecond)
    }

    // Slowing down requests. They slow down later on
    for i := 0; i < 10; i++ {
        resp, err := client.Get("http://localhost:9090/status")
        if err != nil {
            t.Fatalf("Request failed: %v", err)
        }
        resp.Body.Close()
        time.Sleep(2 * time.Second)
    }
}

func TestClientHandleErrors(t *testing.T) {
    logger := log.New(os.Stdout, "TestLog: ", log.LstdFlags)
    c := NewClient("http://localhost:8080", logger)

    // First we initialize a test client server
    server := http.Server{
        Addr:    ":9090",
        Handler: http.HandlerFunc(c.HandleStatusRequest),
    }

    // Start the server in a seperate goroutine to keep it open to requests without blockers.
    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatalf("ListenAndServe(): %v", err)
        }
    }()
    defer server.Close()

    // Simulate user requests with potential errors.
    // Set timeout higher than server delay. Hardcoded for simplicity and simulation.
    client := &http.Client{Timeout: 25 * time.Second} 

    for i := 0; i < 5; i++ {
        resp, err := client.Get("http://localhost:9090/status")
        if err != nil {
            t.Fatalf("Request failed: %v", err)
        }
        defer resp.Body.Close()

        body, _ := ioutil.ReadAll(resp.Body)
        var result map[string]string
        json.Unmarshal(body, &result)

        logger.Printf("Attempt %d: Status: %s, Result: %s", i+1, resp.Status, result["result"])

        if result["result"] == "error" {
            logger.Println("Received 'error' response, sending another request")
            continue
        }

        if result["result"] == "completed" {
            break
        }

        time.Sleep(5 * time.Second)
    }

}
