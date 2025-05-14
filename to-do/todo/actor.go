package todo

import (
	"fmt"
	"log/slog"
)

type Request struct {
	Op      string
	Index   int
	Task    ToDoTask
	ReplyCh chan Response
}

type Response struct {
	Err   error
	Task  *ToDoTask  // add this if you want to return a single task
	Tasks []ToDoTask // and you can still return the whole slice
}

var ReqChan = make(chan Request, 1000)

// Actor runs in one goroutine and serializes access to `tasks`.
func Actor(initial []ToDoTask) chan Request {
	tasks := append([]ToDoTask(nil), initial...)

	for req := range ReqChan {
		switch req.Op {
		case "get":
			slog.Debug("actor get")
			if req.Index < 0 || req.Index >= len(tasks) {
				req.ReplyCh <- Response{Err: fmt.Errorf("index %d out of range", req.Index)}
			} else {
				req.ReplyCh <- Response{Task: &tasks[req.Index]}
			}
		case "list":
			slog.Debug("actor list")
			req.ReplyCh <- Response{Tasks: append([]ToDoTask(nil), tasks...)}
			//slog.Info("updated todo list", "task", tasks)
		case "add":
			//slog.Debug("adding task...", "task", req.Task)
			if req.Task.Status != "" {
				tasks = append(tasks, req.Task)
			} else {
				t := ToDoTask{Description: req.Task.Description, Status: "not started"}
				tasks = append(tasks, t)
			}
			go func(s []ToDoTask) {
				if err := SaveFile(s, TodoFile); err != nil {
					slog.Error("actor: failed to save tasks", "error", err)
				}
			}(append([]ToDoTask(nil), tasks...))
			req.ReplyCh <- Response{Task: &tasks[len(tasks)-1]}
			//slog.Info("Added new task", "task", tasks)

		case "update":
			slog.Debug("actor update", "task", tasks, "index", req.Index, "len", len(tasks))
			if req.Index < 0 || req.Index >= len(tasks) {
				req.ReplyCh <- Response{Err: fmt.Errorf("index %d out of range", req.Index)}
			} else {
				tasks[req.Index] = req.Task
				req.ReplyCh <- Response{Task: &tasks[req.Index]}
				if err := SaveFile(tasks, TodoFile); err != nil {
					slog.Error("actor: failed to save tasks", "error", err)
					req.ReplyCh <- Response{Err: fmt.Errorf("actor: failed to save tasks -  %d ", err)}
				}
			}
		case "delete":
			slog.Debug("actor delete")
			if req.Index < 0 || req.Index >= len(tasks) {
				req.ReplyCh <- Response{Err: fmt.Errorf("index %d out of range", req.Index)}
			} else {
				tasks = append(tasks[:req.Index], tasks[req.Index+1:]...)
				req.ReplyCh <- Response{Tasks: append([]ToDoTask(nil), tasks...)}
				if err := SaveFile(tasks, TodoFile); err != nil {
					slog.Error("actor: failed to save tasks", "error", err)
					req.ReplyCh <- Response{Err: fmt.Errorf("actor: failed to save tasks -  %d ", err)}
				}
			}
		default:
			slog.Error("unknown op", "op", req.Op)
			req.ReplyCh <- Response{Err: fmt.Errorf("unknown op %q", req.Op)}
		}
	}
	return ReqChan
}
