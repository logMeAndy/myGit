package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"to-do/handler"
	"to-do/todo"
)

func main() {
	initLogwithTraceID()
	todoList, err := todo.LoadFile(todo.TodoFile)
	if err != nil {
		slog.Error("Failed to load tasks", "error", err)
	}

	todoList, err = todo.Run(os.Args, todoList)
	if err != nil {
		slog.Error("not found", "error", err)
		return
	}

	//fmt.Println(todoList, len(todoList))

	err = todo.SaveFile(todoList, todo.TodoFile)
	if err != nil {
		slog.Error("Failed to saving tasks", "error", err)
		return
	}
	//waitForInterrupt()

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/about/", http.StripPrefix("/about/", fs))

	mux.Handle("/list", handler.WithLoggingAndTrace(http.HandlerFunc(handler.GetList)))

	mux.Handle("POST /todo", handler.WithLoggingAndTrace(http.HandlerFunc(handler.Create)))
	mux.Handle("PUT /todo/{id}", handler.WithLoggingAndTrace(http.HandlerFunc(handler.UpdateByID)))
	mux.Handle("DELETE /todo/{id}", handler.WithLoggingAndTrace(http.HandlerFunc(handler.DeleteByID)))
	mux.Handle("GET /todo/{id}", handler.WithLoggingAndTrace(http.HandlerFunc(handler.FindByID)))
	mux.Handle("GET /todo", handler.WithLoggingAndTrace(http.HandlerFunc(handler.GetAll)))

	slog.Info("Http Server listining on port 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		slog.Error("Error Listining to port 8080", err)
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
	slog.Info("\nReceived signal:", "Signal", sig)
	fmt.Println("Shutting down gracefully...")
}
