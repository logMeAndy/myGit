package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// ToDoTask matches what the server expects.
type ToDoTask struct {
	Description string `json:"description"`
	Status      string `json:"status"`
}

func main() {
	baseURL := "http://localhost:8080"
	totalRequests := 20000
	ratePerSec := 300

	// Create HTTP client with a timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Ticker fires every 1/ratePerSec seconds
	interval := time.Second / time.Duration(ratePerSec)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var wg sync.WaitGroup
	ctx := context.Background()

	start := time.Now()
	for i := 0; i < totalRequests; i++ {
		<-ticker.C
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			sendTask(ctx, client, baseURL, i)
		}(i)
	}

	// Wait for all requests to complete before exiting
	wg.Wait()
	log.Println("All requests done -->", time.Since(start))
}

func sendTask(ctx context.Context, client *http.Client, baseURL string, i int) {
	// Build the payload
	task := ToDoTask{
		Description: fmt.Sprintf("task-%d", i),
		Status:      "not started",
	}
	payload, err := json.Marshal(task)
	if err != nil {
		log.Printf("ERROR marshaling task %d: %v", i, err)
		return
	}

	// Create the POST request (so you can attach ctx)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/todo", bytes.NewReader(payload))
	if err != nil {
		log.Printf("ERROR creating request for task %d: %v", i, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Send it
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("ERROR POST /todo failed for task %d: %v", i, err)
		return
	}
	defer resp.Body.Close()

	// Drain the body so connections can be reused
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusCreated {
		log.Printf("ERROR unexpected status for task %d: got %d", i, resp.StatusCode)
	} /* else {
		log.Print(i)
	} */
}
