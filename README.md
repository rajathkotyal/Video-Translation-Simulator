# Video-Translation-Simulator
A video translation client-server library written in Go, with a strong focus on performance and its considerations. 

## About :

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


## Prerequisites

- **Go (Golang):** Ensure you have Go installed. You can download it from [Go's official website](https://golang.org/dl/).
- **cURL or Postman:** For testing HTTP endpoints.

## Installation

1. **Clone the Repository:**

   ```bash
   git clone https://github.com/rajathkotyal/Video-Translation-Simulator.git
   cd Video-Translation-Simulator
   ```

   ```
   make tidy
   ```

2. **Start the Server :**

  ```
  make start-server DELAY=20 ERROR_RATE=25
  ```
  - --delay : Sets the number of seconds it would take for a response to be received as "completed" or "error"
  - --error: Sets the probabilty % of server responding with an "error" instead of "completed"

  Not giving anything would set the delay and error to default values : 10s and 20%

3. **Open a new terminal and run tests :**

  ```
  make test
  ```

4. **Start the client to test out the library using postman / curl:**

  ```
  make client
  ```

  Server runs on localhost:8080
  Client runs on localhost:9090

  endpoint is /status

  example command for postman : 
  ```
  http://localhost:9090/status
  ```

Thank you! 



