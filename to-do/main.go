package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"to-do/handler"
	"to-do/todo"
)

var actor = todo.ReqChan

func main() {

	initialTasks, _ := todo.LoadFile(todo.TodoFile)
	go todo.Actor(initialTasks)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go runCLI(ctx, &wg)
	wg.Add(1)
	go runHttpServer(ctx, &wg)

	waitForInterrupt()
	cancel()
	slog.Info("Shutting down gracefully...")
	wg.Wait()
	slog.Info("Todo Application stopped.")
}

func runCLI(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	initLogwithTraceID()

	err := todo.Run(os.Args, actor)
	if err != nil {
		slog.Error("not found", "error", err)
		return
	}
	slog.Debug("runCLI goroutine waiting for context done")
	<-ctx.Done()
	slog.Debug("runCLI goroutine context done now completed")

}

func runHttpServer(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/about/", http.StripPrefix("/about/", fs))                              //static page
	mux.Handle("/list", handler.WithLoggingAndTrace(http.HandlerFunc(handler.GetList))) //dyanmic page

	//APIs
	mux.Handle("PUT /todo/{id}", handler.WithLoggingAndTrace(handler.UpdateByID(actor)))
	mux.Handle("DELETE /todo/{id}", handler.WithLoggingAndTrace(handler.DeleteByID(actor)))
	mux.Handle("GET /todo", handler.WithLoggingAndTrace(http.HandlerFunc(handler.GetAll(actor))))
	mux.Handle("POST /todo", handler.WithLoggingAndTrace(handler.Create(actor)))
	mux.Handle("GET /todo/{id}", handler.WithLoggingAndTrace(handler.FindByID(actor)))

	server := &http.Server{Addr: ":8080", Handler: mux}
	go func() {
		slog.Info("Http Server listining on port :8080")
		if err := server.ListenAndServe(); err != nil {
			slog.Error("Error Listining to port 8080", "server", err)
		}
		slog.Info("HTTPServer ListenAndServe goroutine stopped")
	}()

	slog.Debug("runHTTPServer goroutine waiting for context done")
	<-ctx.Done()

	// Initiate graceful shutdown
	slog.Info("HTTP Server shutting down...")
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server Shutdown error", "error", err)
	} else {
		slog.Info("HTTP Server gracefully stopped")
	}
}

func initLogwithTraceID() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	traceID := fmt.Sprintf("trace-%d", time.Now().UnixNano())
	ctx := context.WithValue(context.Background(), "traceID", traceID)
	slog.SetDefault(logger.With("traceID", ctx.Value("traceID")))
}

func waitForInterrupt() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	sig := <-sigChan
	slog.Info("Received Interrupt Signal:", "Signal", sig)
}
