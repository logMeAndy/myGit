package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"to-do/todo"
)

func FindByID(w http.ResponseWriter, r *http.Request) {
	slog.SetDefault(LoggerFromContext(r.Context()))
	slog.Debug("received request to find todo item")
	todoList, err := todo.LoadFile(todo.TodoFile)
	if err != nil {
		slog.Error("Failed to load tasks", "error", err)
		http.Error(w, `{"error":"Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	idv, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		slog.Error("invalid index parameter", "raw", r.PathValue("id"), "error", err)
		http.Error(w, `{"error":"invalid index parameter"}`, http.StatusNotFound)
		return
	}

	// now check bounds: valid indices are 0 through len(todoList)-1
	if idv < 0 || idv >= len(todoList) {
		slog.Error("index out of range", "index", idv, "length", len(todoList))
		http.Error(w, `{"error":"index out of range"}`, http.StatusNotFound)
		return
	}

	todoItem := todoList[idv]
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(todoItem)
	slog.Info("returned one task", "index", idv, "task", todoItem)
}

func Create(w http.ResponseWriter, r *http.Request) {
	slog.SetDefault(LoggerFromContext(r.Context()))
	slog.Info("received request to create a todo")
	task := requestMessage(w, r)
	if task == nil {
		slog.Error("Invalid task:", "task", task)
		return
	}

	tdesc := "-task=" + task.Description
	tstatus := "-status=" + task.Status
	args := []string{"cmd", tdesc, tstatus}
	process(args, "")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func basicCheck() {

}

func process(args []string, id string) error {
	todoList, err := todo.LoadFile(todo.TodoFile)
	if err != nil {
		slog.Error("Failed to load tasks", "error", err)
		return err
	}
	slog.Debug("User Request Path Value", "length", len(todoList), "PathValue", id)

	idv, err := strconv.Atoi(id)
	if err != nil {
		slog.Error("invalid index parameter", "raw", id, "error", err)
		return errors.New("invalid index parameter")
	}

	// now check bounds: valid indices are 0 through len(todoList)-1
	if idv < 0 || idv >= len(todoList) {
		slog.Error("index out of range", "index", idv, "length", len(todoList))
		return fmt.Errorf("index %d out of range", idv)
	}

	//	todoList[len(todoList)-1]

	todoList, err = todo.Run(args, todoList)
	if err != nil {
		slog.Error("Run failed", "error", err)
		return err
	}

	err = todo.SaveFile(todoList, todo.TodoFile)
	if err != nil {
		slog.Error("Failed to saving tasks", "error", err)
		return err
	}
	return nil
}

func UpdateByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := LoggerFromContext(ctx)
	slog.SetDefault(logger)
	slog.Info("received request to update todo item")
	task := requestMessage(w, r)
	if task == nil {
		slog.Error("Invalid task:", "task", task)
		http.Error(w, `{"error":"Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	tupt := "-update=" + r.PathValue("id")
	tdesc := "-task=" + task.Description
	tstatus := "-status=" + task.Status
	args := []string{"cmd", tupt, tdesc, tstatus}
	if err := process(args, r.PathValue("id")); err != nil {
		http.Error(w, `{"error":"Task not found"}`, http.StatusNotFound)
		slog.Warn("Index not found", "error", err)
		return
	}
	if err := json.NewEncoder(w).Encode(task); err != nil {
		slog.Error("Failed to encode task response", "error", err)
		http.Error(w, `{"error":"Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	//	w.Header().Set("Content-Type", "application/json")
	//	w.WriteHeader(http.StatusAccepted)
	//json.NewEncoder(w).Encode(`{"error":""}`)
	slog.Info("Update OK")
}
func DeleteByID(w http.ResponseWriter, r *http.Request) {
	slog.SetDefault(LoggerFromContext(r.Context()))

	slog.Info("received request to delete a todo item")

	tdel := "-delete=" + r.PathValue("id")
	args := []string{"cmd", tdel}
	if process(args, r.PathValue("id")) != nil {
		http.Error(w, `{"error":"Task not found"}`, http.StatusNotFound)
		slog.Warn("Task not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent) //204
}

func GetAll(w http.ResponseWriter, r *http.Request) {
	slog.SetDefault(LoggerFromContext(r.Context()))
	slog.Info("received request to fetch all todo list")

	todoList, err := todo.LoadFile(todo.TodoFile)
	if err != nil {
		slog.Error("Failed to load tasks", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(todoList)
}

func GetList(w http.ResponseWriter, r *http.Request) {
	slog.SetDefault(LoggerFromContext(r.Context()))
	slog.Info("received request to list all todo items")

	todoList, err := todo.LoadFile(todo.TodoFile)
	if err != nil {
		slog.Error("Failed to load tasks", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tmpl, err := template.ParseFiles("dynamic/list.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		slog.Error("template parse failed", "error", err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, todoList); err != nil {
		slog.Error("template execute failed", "error", err)
	}
}

func requestMessage(w http.ResponseWriter, r *http.Request) *todo.ToDoTask {
	slog.Debug("request received in handler.requestMessage")
	var task todo.ToDoTask
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("could not read request", "Body length", len(body))
		http.Error(w, `{"error":"could not read request"}`, http.StatusBadRequest)
		return nil
	}
	if json.Unmarshal(body, &task) != nil {
		slog.Error("Invalid JSON", "Body length", len(body))
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return nil
	}
	slog.Debug("Unmarshal data ", "task list", task)
	return &task
}

/*
func JSONError(w http.ResponseWriter, ctx context.Context, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
//	getTraceID

	traceKey := "traceID"
	traceID, _ := ctx.Value(traceKey).(string)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":   msg,
		"traceID": traceID,
	})
}

*/
