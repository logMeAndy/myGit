package main

import (
	"context"
	"log/slog"
	"sync"
	"to-do/handler"
	"to-do/todo"
)

var actor = todo.ReqChan

func main() {
	initialTasks, _ := todo.LoadFile(todo.TodoFile)
	go todo.Actor(initialTasks)

	ctx, cancel := context.WithCancel(context.Background())
	//	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go handler.RunHttpServer(ctx, &wg, actor)

	handler.WaitForInterrupt()
	cancel()
	slog.Info("Shutting down gracefully...")
	wg.Wait()
	slog.Info("Todo Application stopped.")
}
