package todo

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

func InitLogwithTraceID() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	traceID := fmt.Sprintf("trace-%d", time.Now().UnixNano())
	ctx := context.WithValue(context.Background(), "traceID", traceID)
	slog.SetDefault(logger.With("traceID", ctx.Value("traceID")))
}

func RunCLI(ctx context.Context, wg *sync.WaitGroup, actor chan Request) {
	defer wg.Done()
	InitLogwithTraceID()

	err := Run(os.Args, actor)
	if err != nil {
		slog.Error("not found", "error", err)
		return
	}
	slog.Debug("runCLI goroutine waiting for context done")
	<-ctx.Done()
	slog.Debug("runCLI goroutine context done now completed")

}

func Run(args []string, actor chan Request) error {
	fs := flag.NewFlagSet("todo", flag.ContinueOnError)

	var taskDesc = fs.String("task", "", "Task description e.g. -task=newItemDescription (optional -status=newStatus) (default not started))")
	var status = fs.String("status", "", "New status e.g. not started, completed, started, etc.,")
	var updateIndex = fs.Int("update", -1, "Index of task to update, e.g. update=0 -task=newValue (optional -status=newStatus)")
	var deleteIndex = fs.Int("delete", -1, "Index of task to delete (e.g. delete=0 )")

	fs.Parse(args[1:]) // skip program name

	slog.Debug("args", "deleteIndex", deleteIndex, "updateIndex", updateIndex, "status", status, "taskDesc", taskDesc)
	switch {
	case *updateIndex >= 0:
		var t ToDoTask
		if *status != "" {
			t = ToDoTask{Description: *taskDesc, Status: *status}
		} else {
			t = ToDoTask{Description: *taskDesc, Status: "not started"}
		}
		reply := make(chan Response)
		actor <- Request{Op: "update", Task: t, Index: *updateIndex, ReplyCh: reply}
		res := <-reply
		slog.Debug("received actor response", "response", res)
		if res.Err != nil {
			slog.Error("Invalid task:", "index", *updateIndex)
			return fmt.Errorf("404") //404 not found
		}
		slog.Info("Task updated", "Index", *updateIndex)
		return nil

	case *taskDesc != "":
		slog.Debug("adding task...", "task", *taskDesc)
		var t ToDoTask
		if *status != "" {
			t = ToDoTask{Description: *taskDesc, Status: *status}
		} else {
			t = ToDoTask{Description: *taskDesc, Status: "not started"}
		}
		reply := make(chan Response)
		actor <- Request{Op: "add", Task: t, ReplyCh: reply}
		res := <-reply
		slog.Debug("received actor response", "response", res)
		if res.Err != nil {
			slog.Error("Invalid task:", "index", *updateIndex)
			return fmt.Errorf("404") //404 not found
		}
		slog.Info("Added new task", "task", t)
		return nil
	case *deleteIndex >= 0:
		slog.Debug("deleting index...", "task", *deleteIndex)
		reply := make(chan Response)
		actor <- Request{Op: "delete", Index: *deleteIndex, ReplyCh: reply}
		res := <-reply
		slog.Debug("received actor response", "response", res)
		if res.Err != nil {
			slog.Error("Invalid task:", "index", *updateIndex)
			return fmt.Errorf("404") //404 not found
		}

	default:
		reply := make(chan Response)
		actor <- Request{Op: "list", ReplyCh: reply}
		res := <-reply
		slog.Info("received actor response", "response", res.Tasks)
	}
	return nil
}
