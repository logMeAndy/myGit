package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"to-do/todo"
)

var actor = todo.ReqChan

func main() {
	//initialTasks, _ := todo.LoadFile(todo.TodoFile)
	//go todo.Actor(initialTasks)

	user := flag.String("user", "", "User ID (required for repl session)")
	flag.Parse()
	if *user == "" {
		fmt.Fprintln(os.Stderr, "ERROR: you must provide -user")
		flag.Usage()
		os.Exit(1)
	}

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

	todo.RunREPL(actor, *user)

}
