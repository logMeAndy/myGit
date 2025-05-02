package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
	"to-do/todo"
)

const todoFile = "todo.json"

func main() {
	initLogwithTraceID()

	todoList, err := todo.LoadFile(todoFile)
	if err != nil {
		slog.Error("Failed to load tasks", "error", err)
	}

	todoList = run(os.Args, todoList)

	//fmt.Println(todoList, len(todoList))

	err = todo.SaveFile(todoList, todoFile)
	if err != nil {
		slog.Error("Failed to saving tasks", "error", err)
		return
	}
	waitForInterrupt()
}

func run(args []string, todoList []todo.ToDoTask) []todo.ToDoTask {
	fs := flag.NewFlagSet("todo", flag.ContinueOnError)

	var taskDesc = fs.String("task", "", "Task description e.g. -task=newItemDescription (optional -status=newStatus) (default not started))")
	var status = fs.String("status", "", "New status e.g. not started, completed, started, etc.,")
	var updateIndex = fs.Int("update", -1, "Index of task to update, e.g. update=0 -task=newValue (optional -status=newStatus)")
	var deleteIndex = fs.Int("delete", -1, "Index of task to delete (e.g. delete=0 )")

	fs.Parse(args[1:]) // skip program name

	slog.Debug("args", "deleteIndex", deleteIndex, "updateIndex", updateIndex, "status", status, "taskDesc", taskDesc)
	switch {
	case *updateIndex >= 0:
		slog.Debug("updating index...", slog.Int("updateIndex", *updateIndex))
		if *taskDesc != "" && *updateIndex < len(todoList) {
			if *status != "" {
				todoList[*updateIndex].Description = *taskDesc
				todoList[*updateIndex].Status = *status
			} else {
				todoList[*updateIndex].Description = *taskDesc
			}
			slog.Info("Task updated", "Index", todoList[*updateIndex])
		} else if *status != "" && *updateIndex < len(todoList) {
			todoList[*updateIndex].Status = *status
			slog.Info("Task updated", "Index", todoList[*updateIndex])
		} else {
			slog.Info("Invalid index or missing task description, please check current TODO list or help using -h", "Index", *updateIndex, "Length", len(todoList))
		}
	case *taskDesc != "":
		slog.Debug("adding task...", "task", *taskDesc)
		var t todo.ToDoTask
		if *status != "" {
			t = todo.ToDoTask{Description: *taskDesc, Status: *status}
			todoList = append(todoList, t)
		} else {
			t = todo.ToDoTask{Description: *taskDesc, Status: "not started"}
			todoList = append(todoList, t)
		}
		slog.Info("Added new task", "task", t)
	case *deleteIndex >= 0 && *deleteIndex < len(todoList):
		slog.Debug("deleting index...", "task", *deleteIndex)
		if *deleteIndex < len(todoList) {
			deletedTask := todoList[*deleteIndex]
			todoList = append(todoList[:*deleteIndex], todoList[*deleteIndex+1:]...)
			slog.Info("Deleted task", "index", *deleteIndex, "task", deletedTask)
		} else {
			slog.Info("Invalid index or missing task description, please check current TODO list or help using -h", "Index", *deleteIndex, "Length", len(todoList))
		}
	default:
		fmt.Println("Current To-Do List :")
		for i, t := range todoList {
			fmt.Printf("%d. %s, %s \n", i, t.Description, t.Status)
		}
	}
	return todoList
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
