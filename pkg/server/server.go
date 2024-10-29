package server

import (
    "encoding/json"
    "log"
    "net/http"
    "sync"
    "time"
		"math/rand"
)

/*
	Comments summarizing the code as a whole for easy understanding :

	This is a simple server which creates a new Server instance using the configured time given.
	It responds back with a json {“result”: “pending” or “error” or “completed”} for each request 
	after the time for that request has passed. 

*/

// Config holds the server configuration options.
type Config struct {
	DelaySeconds int // Delay before returning final status.
	ErrorRate    int // Probability of returning "error" instead of "completed".
}

// Response represents the JSON structure returned by the server.
type Response struct {
    Result string `json:"result"`
}

// Server represents the video translation server.
type Server struct {
    startTime     time.Time
    config 				*Config
    status        string
    mu            sync.Mutex
}

// NewServer initializes a new Server instance.
func NewServer(delaySeconds int, errorRate int) (*Server, error) {
	
	// Validate the inputs
	if delaySeconds <= 0 {
			log.Printf("Invalid delay value %d. Using default of 10 seconds.", delaySeconds)
			delaySeconds = 10
	}
	if errorRate < 0 || errorRate > 100 {
			log.Printf("Invalid error rate value %d. Using default of 20%%.", errorRate)
			errorRate = 20
	}

	config := &Config{
			DelaySeconds: delaySeconds,
			ErrorRate:    errorRate,
	}

	// Seed the random number generator for non deterministic random nos.
	rand.Seed(time.Now().UnixNano()) 
	return &Server{
			config:    config,
			startTime: time.Now(),
			status:    "pending",
	}, nil
}


// Start begins listening for HTTP requests on the specified address.
func (s *Server) Start(address string) error {
	http.HandleFunc("/status", s.statusHandler)
	log.Printf("Server is starting on %s with a delay of %d seconds and error rate of %d%%",
			address, s.config.DelaySeconds, s.config.ErrorRate)
	return http.ListenAndServe(address, nil)
}

// statusHandler handles incoming requests to the /status endpoint.
func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Reset the timer and status if the current status is not "pending" 
	// --> Simulating a new job that could have been posted
	if s.status != "pending" {
			s.startTime = time.Now()
			s.status = "pending"
			log.Println("New request received. Resetting timer and status to 'pending'.")
	}

	elapsed := time.Since(s.startTime)
	if s.status == "pending" && elapsed >= time.Duration(s.config.DelaySeconds)*time.Second {
			s.status = s.randomStatus()
	}

	response := Response{Result: s.status}
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
	}

	log.Printf("Handled /status request. Responded with: %s", s.status)
}

// randomStatus determines the final status based on the error rate.
func (s *Server) randomStatus() string {
	if rand.Intn(100) < s.config.ErrorRate {
			return "error"
	}
	return "completed"
}
