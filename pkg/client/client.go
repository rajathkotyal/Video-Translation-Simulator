package client

import (
    "context"
    "encoding/json"
    "errors"
    "log"
    "math/rand"
    "net/http"
    "sync"
    "time"
)

/*
	Comments summarizing the code as a whole for easy understanding :

	This is a client library for handling requests from the user and passing it over to the server.
    To handle load, 
    - It uses exponential backoff to reduce load on the server by spacing the requests.
    - It adds jitter to prevent Thundering Herd problems and synchronized reties : 
        i.e. multiple users may request for the job at the same time leading to requests to the 
        server at the same exact time, which would overload it.
    - It maintains state of the job through response from previous requests, such that,
      even when the user keeps trying to fetch status, the service doesnt hang and sends the previous response itself
    - It uses mutex locks to make sure shared variables are accessed and modified properly by concurrent requests
    - Adaptive Retry : A retry mechanism based on the previously observed delay from server, but since in this simulation
            we have a fixed amount of delay, we wouldnt need this.

    How it helps the users : 
    - Responsive Interaction: Users receive immediate responses to their requests, enhancing the user experience.
    - Reduced Waiting Time: The client handles the polling logic, so users don't need to wait for long-running server processes.
    
    How it helps the 3rd party dev using this library : 
    - Simplified Client-Side Logic: Developers interact with a straightforward API/REST without worrying about the underlying polling mechanics.

    Stretch Goal implementaions (Not necessary for this simulation, Sample code is present at end of file) :
    A token bucket based rate limiter :
        - Limit the number of requests to prevent DDOS attacks and reduce load at client side itself.
        - Since we already have a custom rate limiter that would only send requests based on the number of times its
            been received, we wouldnt need it in this simulation.
    A Request Queue : 
        - Explicitly prioritize the requests as they come in (This is handled through go routines by default, but we
            may need a request queue to do some processing explicitly)
    
    
*/

// Client represents the client library to interact with the server.
type Client struct {
    BaseURL       string
    Logger        *log.Logger
    httpClient    *http.Client
    mu            sync.Mutex
    status        string
    attempt       int
    delay         time.Duration
    maxDelay      time.Duration
    maxRetries    int
    lastRequest   time.Time
    nextRequest   time.Time
    initialDelay  time.Duration
    pending       bool
    timeout       time.Duration
}

// NewClient initializes a new Client with default settings.
func NewClient(baseURL string, logger *log.Logger) *Client {
    return &Client{
        BaseURL:      baseURL,
        Logger:       logger,
        httpClient:   &http.Client{},
        initialDelay: 500 * time.Millisecond,
        maxDelay:     10 * time.Second,
        maxRetries:   20,
        status:       "",
        attempt:      0,
        delay:        0,
        lastRequest:  time.Time{},
        nextRequest:  time.Time{},
        pending:      false,
        timeout:      5 * time.Second,
    }
}

// HandleStatusRequest handles incoming /status HTTP requests.
func (c *Client) HandleStatusRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    c.mu.Lock()
    defer c.mu.Unlock()

    // Check if we need to initialize a new polling sequence.
    if !c.pending {
        c.Logger.Println("Starting new polling sequence")
        c.pending = true
        c.status = "pending"
        c.attempt = 0
        c.delay = c.initialDelay
        c.lastRequest = time.Time{}
        c.nextRequest = time.Now()
    }

    now := time.Now()
    if now.Before(c.nextRequest) {
        // Not yet time to make the next request.
        c.Logger.Printf("Next request to server in %v", c.nextRequest.Sub(now))
        // Return last known status.
        c.respondWithStatus(w, c.status)
        return
    }

    // Make request to  server.
    c.attempt++
    status, err := c.RetrieveStatus(ctx)
    if err != nil {
        c.Logger.Printf("Attempt %d: Error fetching status: %v", c.attempt, err)
        if c.attempt >= c.maxRetries {
            c.Logger.Printf("Max retries reached")
            c.respondWithError(w, "Max retries reached")
            c.pending = false
            return
        }
    } else {
        c.Logger.Printf("Attempt %d: Received status: %s", c.attempt, status)
        c.status = status
        if status == "pending" {
            // Update delay and next request time.
            c.delay = c.nextDelay(c.delay)
            c.nextRequest = time.Now().Add(c.delay)
            c.Logger.Printf("Next attempt in %v", c.delay)
        } else {
            // Final status received.
            c.pending = false
        }
    }

    c.lastRequest = time.Now()
    c.respondWithStatus(w, c.status)
}

func (c *Client) respondWithStatus(w http.ResponseWriter, status string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"result": status})
}

func (c *Client) respondWithError(w http.ResponseWriter, message string) {
    http.Error(w, message, http.StatusInternalServerError)
}

// nextDelay calculates the next delay with exponential backoff and jitter.
func (c *Client) nextDelay(currentDelay time.Duration) time.Duration {
    if currentDelay == 0 {
        currentDelay = c.initialDelay
    } else {
        currentDelay *= 2
    }
    if currentDelay > c.maxDelay {
        currentDelay = c.maxDelay
    }
    // Add jitter.
    jitter := time.Duration(rand.Int63n(int64(currentDelay / 2)))
    totalDelay := currentDelay/2 + jitter
    c.Logger.Printf("Exponential backoff: current delay %v, jitter added %v, total delay %v", currentDelay/2, jitter, totalDelay)
    return totalDelay
}

// RetrieveStatus makes an HTTP GET request to the /status endpoint.
func (c *Client) RetrieveStatus(ctx context.Context) (string, error) {
    req, err := http.NewRequest("GET", c.BaseURL+"/status", nil)
    if err != nil {
        return "", err
    }

    ctx, cancel := context.WithTimeout(ctx, c.timeout)
    defer cancel()
    req = req.WithContext(ctx)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", errors.New("received non-200 response from server")
    }

    var response struct {
        Result string `json:"result"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return "", err
    }

    return response.Result, nil
}


// TODO : We can add an adaptiveRetry and request queue as well. But since our server load isnt variying 
// and only 1 user is making requests in this simulation, we wouldnt need it.

// func (c *Client) adaptiveRetryDelay() time.Duration {
//     newDelay := time.Duration(float64(c.delay) * c.retryMultiplier)
//     if newDelay > c.maxDelay {
//         newDelay = c.maxDelay
//     }
//     return newDelay
// }

// A very basic Request queue with Rate Limiter example. Again, not necessary for this simulation I feel
/*
rateLimiter:     rate.NewLimiter(rate.Every(100*time.Millisecond), 1), // 10 requests per second
requestQueue:    make(chan struct{}, 100)
if err := c.rateLimiter.Wait(ctx); err != nil {
    c.respondWithError(w, "Rate limit exceeded")
    return
}

// We add request to q here through a channel, and any new request is in waiting state
select {
case c.requestQueue <- struct{}{}:
default:
    c.respondWithError(w, "Request queue full")
    return
}
defer func() { <-c.requestQueue }()

*/
