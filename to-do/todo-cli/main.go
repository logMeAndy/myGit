package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"sync"
	"to-do/handler"
	"to-do/todo"
)

var actor = todo.ReqChan

func main() {
	//initialTasks, _ := todo.LoadFile(todo.TodoFile)
	//go todo.Actor(initialTasks)

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("cannot determine cwd:", err)
	}
	slog.Info("Get Workding Directory ", "DIR", cwd)

	userLists, err := todo.LoadAllTasksInDir(cwd)
	if err != nil {
		log.Fatal("failed loading tasks:", err)
	}
	slog.Info("Loaded files ...", "DATA", userLists)

	go todo.Actor(userLists)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go todo.RunCLI(ctx, &wg, actor)

	handler.WaitForInterrupt()
	cancel()
	slog.Info("Shutting down gracefully...")
	wg.Wait()
	slog.Info("Todo Application stopped.")
}
