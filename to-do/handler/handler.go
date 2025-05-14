package handler

import (
	"encoding/json"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"
	"to-do/todo"
)

func requestMessage(r *http.Request) *todo.ToDoTask {
	slog.Debug("request received in handler.requestMessage")
	var task todo.ToDoTask
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("could not read request", "Body length", len(body))
		return nil
	}
	slog.Debug("raw request body", "body", string(body))
	if json.Unmarshal(body, &task) != nil {
		slog.Error("Invalid JSON", "Body length", len(body))
		return nil
	}
	slog.Debug("Unmarshal parsed task data ", "task list", task)
	return &task
}

func Create(actor chan todo.Request) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.SetDefault(LoggerFromContext(r.Context()))
		//slog.Info("received request to create a todo")
		task := requestMessage(r)
		if task == nil {
			slog.Error("Invalid task:", "task", task)
			http.Error(w, `{"error":"could not read request"}`, http.StatusBadRequest)
			return
		}
		//slog.Debug("sending add command", "task", task)
		reply := make(chan todo.Response)
		actor <- todo.Request{Op: "add", Task: *task, ReplyCh: reply}
		res := <-reply
		//slog.Debug("received actor response", "response", res)
		if res.Err != nil {
			slog.Error("Invalid task:", "task", task)
			http.Error(w, `{"error":"could not read request"}`, http.StatusInternalServerError)
			return
		}
		//SaveSnapshot(actor)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
	}

}

/*
func SaveSnapshot(actor chan todo.Request) {
	slog.Debug("requesting full list to be saved")
	listCh := make(chan todo.Response)
	actor <- todo.Request{Op: "list", ReplyCh: listCh}
	listRes := <-listCh
	slog.Debug("received full list", "count", len(listRes.Tasks))
	//todo.SaveFile((<-listCh).Tasks, todo.TodoFile)
	if err := todo.SaveFile(listRes.Tasks, todo.TodoFile); err != nil {
		slog.Error("failed to save tasks", "error", err)
	}

} */

func UpdateByID(actor chan todo.Request) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.SetDefault(LoggerFromContext(r.Context()))
		slog.Info("received request to update todo item")
		task := requestMessage(r)
		if task == nil {
			slog.Error("Invalid task:", "task", task)
			http.Error(w, `{"error":"could not read request"}`, http.StatusBadRequest)
			return
		}

		id := r.PathValue("id")
		if id != "" { //create and findAll
			idv, err := strconv.Atoi(id)
			if err != nil {
				slog.Error("invalid index parameter", "raw", id, "error", err)
				http.Error(w, `{"error":"Task not found"}`, http.StatusNotFound)
				return
			}

			reply := make(chan todo.Response)
			actor <- todo.Request{Op: "update", Task: *task, Index: idv, ReplyCh: reply}
			res := <-reply
			slog.Debug("received actor response", "response", res)
			if res.Err != nil {
				slog.Error("Invalid task:", "index", id)
				http.Error(w, `{"error":"could not read request"}`, http.StatusInternalServerError)
				return
			}

		} else {
			slog.Error("no index parameter", "index", id)
			http.Error(w, `{"error":"Task not found"}`, http.StatusNotFound)
			return
		}

		//SaveSnapshot(actor)

		if err := json.NewEncoder(w).Encode(task); err != nil {
			slog.Error("Failed to encode task response", "error", err)
			http.Error(w, `{"error":"Internal Server Error"}`, http.StatusInternalServerError)
			return
		}
		//	w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		//json.NewEncoder(w).Encode(`{"error":""}`)
		slog.Info("Update OK")
	}
}

func DeleteByID(actor chan todo.Request) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.SetDefault(LoggerFromContext(r.Context()))
		slog.Info("received request to delete a todo item")

		id := r.PathValue("id")
		if id != "" { //create and findAll
			idv, err := strconv.Atoi(id)
			if err != nil {
				slog.Error("invalid index parameter", "raw", id, "error", err)
				http.Error(w, `{"error":"Task not found"}`, http.StatusNotFound)
				return
			}

			reply := make(chan todo.Response)
			actor <- todo.Request{Op: "delete", Index: idv, ReplyCh: reply}
			res := <-reply
			slog.Debug("received actor response", "response", res)
			if res.Err != nil {
				slog.Error("Invalid task:", "index", id)
				http.Error(w, `{"error":"could not read request"}`, http.StatusInternalServerError)
				return
			}
		} else {
			slog.Error("no index parameter", "index", id)
			http.Error(w, `{"error":"Task not found"}`, http.StatusNotFound)
			return
		}

		//SaveSnapshot(actor)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent) //204
	}
}

func FindByID(actor chan todo.Request) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		slog.SetDefault(LoggerFromContext(r.Context()))
		slog.Debug("received request to find todo item")
		var intIndex int
		id := r.PathValue("id")
		if id != "" {
			idv, err := strconv.Atoi(id)
			intIndex = idv
			if err != nil {
				slog.Error("invalid index parameter", "raw", id, "error", err)
				http.Error(w, `{"error":"Task not found"}`, http.StatusNotFound)
				return
			}
		} else {
			slog.Error("no index parameter", "index", id)
			http.Error(w, `{"error":"Task not found"}`, http.StatusNotFound)
			return
		}
		reply := make(chan todo.Response)
		actor <- todo.Request{Op: "get", Index: intIndex, ReplyCh: reply}
		res := <-reply
		slog.Debug("received actor response", "response", res.Task)
		if res.Err != nil {
			slog.Error("Invalid task:", "index", id)
			http.Error(w, `{"error":"could not read request"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res.Task)
		slog.Info("returned one task", "index", id, "task", res.Task)
		slog.Debug("Request timings", "method", r.Method, "Url", r.RequestURI, "time", time.Since(start).Milliseconds())
	}
}

func GetAll(actor chan todo.Request) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.SetDefault(LoggerFromContext(r.Context()))
		//slog.Info("received request to fetch all todo list")

		reply := make(chan todo.Response)
		actor <- todo.Request{Op: "list", ReplyCh: reply}
		res := <-reply
		//slog.Debug("received actor response", "response", res.Tasks)
		if res.Err != nil {
			slog.Error("Internal error:")
			http.Error(w, `{"error":"could not read request"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res.Tasks)
	}
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
