package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"to-do/handler"
	"to-do/todo"

	"golang.org/x/time/rate"
)

var actor = todo.ReqChan

func TestHTTPActorConcurrencyOn8080(t *testing.T) {
	t.Log("Starting TestHTTPActorConcurrencyOn8080")
	go todo.Actor(nil)
	t.Log("Actor started")

	mux := http.NewServeMux()
	mux.Handle("PUT /todo/{id}", handler.WithLoggingAndTrace(handler.UpdateByID(actor)))
	mux.Handle("DELETE /todo/{id}", handler.WithLoggingAndTrace(handler.DeleteByID(actor)))
	mux.Handle("GET /todo", handler.WithLoggingAndTrace(http.HandlerFunc(handler.GetAll(actor))))
	mux.Handle("POST /todo", handler.WithLoggingAndTrace(handler.Create(actor)))
	mux.Handle("GET /todo/{id}", handler.WithLoggingAndTrace(handler.FindByID(actor)))

	// Create an unstarted test server
	ts := httptest.NewUnstartedServer(mux)
	//ts := httptest.NewServer(mux)
	//defer ts.Close()

	// Bind it to 127.0.0.1:8080
	t.Log("Attempting to listen on 127.0.0.1:8080")
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Skipf("port 8080 is unavailable, skipping: %v", err)
	}
	t.Log("Listening on 127.0.0.1:8080")
	ts.Listener = listener
	ts.Start()
	defer func() {
		ts.Close()
		t.Log("Test server closed")
	}()

	baseURL := "http://127.0.0.1:8080"
	const n = 20000

	var wg sync.WaitGroup
	wg.Add(n)

	//workerSem := make(chan struct{}, 100)

	rateLimit := rate.NewLimiter(rate.Limit(200), 1)
	ctx := context.Background()

	t.Logf("Firing %d concurrent POST requests", n)
	start := time.Now()
	// Fire n concurrent POSTs
	for i := 0; i < n; i++ {
		i := i

		if err := rateLimit.Wait(ctx); err != nil {
			t.Errorf("Rate limiter wait failed: %v", err)
			continue
		}

		//	workerSem <- struct{}{}

		go func() {
			defer wg.Done()
			//defer func() { <-workerSem }()
			//t.Logf("Goroutine %d sending POST", i)
			task := todo.ToDoTask{Description: fmt.Sprintf("task-%d", i), Status: "not started"}
			payload, _ := json.Marshal(task)
			resp, err := http.Post(baseURL+"/todo", "application/json", bytes.NewReader(payload))
			if err != nil {
				t.Errorf("POST /todo failed for task %d: %v", i, err)
				return
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			//t.Logf("Goroutine %d received status %d", i, resp.StatusCode)
			if resp.StatusCode != http.StatusCreated {
				t.Errorf("expected 201 Created; got %d", resp.StatusCode)
			}
		}()
	}

	t.Log("Waiting for all POST goroutines to finish", time.Since(start))
	wg.Wait()
	t.Log("All POST requests completed", time.Since(start))

	// Verify with GET
	t.Log("Sending GET /todo to verify tasks")
	resp, err := http.Get(baseURL + "/todo")
	if err != nil {
		t.Fatalf("GET /todo failed: %v", err)
	}
	defer resp.Body.Close()
	t.Logf("GET /todo returned status %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK; got %d", resp.StatusCode)
	}

	var tasks []todo.ToDoTask
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	t.Logf("Decoded %d tasks", len(tasks))
	if len(tasks) != n {
		t.Fatalf("expected %d tasks, got %d", n, len(tasks))
	}
}
