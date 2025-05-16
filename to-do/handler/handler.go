package handler

import (
	"context"
	"encoding/json"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
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

func RunHttpServer(ctx context.Context, wg *sync.WaitGroup, actor chan todo.Request) {
	defer wg.Done()

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/about/", http.StripPrefix("/about/", fs))              //static page
	mux.Handle("/list", WithLoggingAndTrace(http.HandlerFunc(GetList))) //dyanmic page

	//APIs
	mux.Handle("PUT /todo/users/{userID}/{id}", WithLoggingAndTrace(UpdateByID(actor)))
	mux.Handle("DELETE /todo/users/{userID}/{id}", WithLoggingAndTrace(DeleteByID(actor)))
	mux.Handle("GET /todo/users/{userID}", WithLoggingAndTrace(http.HandlerFunc(GetAll(actor))))
	//mux.Handle("POST /todo", WithLoggingAndTrace(Create(actor)))
	mux.Handle("GET /todo/users/{userID}/{id}", WithLoggingAndTrace(FindByID(actor)))

	// e.g. POST /users/{userID}/todo
	mux.Handle("POST /todo/users/{userID}", WithLoggingAndTrace(Create(actor)))

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

func Send(actor chan todo.Request, op, user string, idx int, task todo.ToDoTask) (resp todo.Response) {
	ch := make(chan todo.Response, 1)
	actor <- todo.Request{Op: op, UserID: user, Index: idx, Task: task, ReplyCh: ch}
	return <-ch
}

func Create(actor chan todo.Request) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.SetDefault(LoggerFromContext(r.Context()))
		user := r.PathValue("userID")

		slog.Info("received request to create a todo", "user", user)
		task := requestMessage(r)
		if task == nil {
			slog.Error("Invalid task:", "task", task)
			http.Error(w, `{"error":"could not read request"}`, http.StatusBadRequest)
			return
		}
		//slog.Debug("sending add command", "task", task)
		reply := make(chan todo.Response)
		actor <- todo.Request{Op: "add", UserID: user, Task: *task, ReplyCh: reply}
		res := <-reply
		//slog.Debug("received actor response", "response", res)
		if res.Err != nil {
			slog.Error("Invalid task:", "task", task)
			http.Error(w, `{"error":"could not read request"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
	}

}

func UpdateByID(actor chan todo.Request) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.SetDefault(LoggerFromContext(r.Context()))
		slog.Info("received request to update todo item")
		user := r.PathValue("userID")

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

			//reply := make(chan todo.Response)
			//actor <- todo.Request{Op: "update", UserID: user, Task: *task, Index: idv, ReplyCh: reply}
			//res := <-reply
			//slog.Debug("received actor response", "response", res)

			if Send(actor, "update", user, idv, *task).Err != nil {
				slog.Error("Invalid task:", "index", id)
				http.Error(w, `{"error":"could not read request"}`, http.StatusInternalServerError)
				return
			}

		} else {
			slog.Error("no index parameter", "index", id)
			http.Error(w, `{"error":"Task not found"}`, http.StatusNotFound)
			return
		}

		if err := json.NewEncoder(w).Encode(task); err != nil {
			slog.Error("Failed to encode task response", "error", err)
			http.Error(w, `{"error":"Internal Server Error"}`, http.StatusInternalServerError)
			return
		}
		//	w.Header().Set("Content-Type", "application/json")
		//w.WriteHeader(http.StatusAccepted)
		//json.NewEncoder(w).Encode(`{"error":""}`)
		slog.Info("Update OK")
	}
}

func DeleteByID(actor chan todo.Request) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.SetDefault(LoggerFromContext(r.Context()))
		slog.Info("received request to delete a todo item")
		user := r.PathValue("userID")

		id := r.PathValue("id")
		if id != "" { //create and findAll
			idv, err := strconv.Atoi(id)
			if err != nil {
				slog.Error("invalid index parameter", "raw", id, "error", err)
				http.Error(w, `{"error":"Task not found"}`, http.StatusNotFound)
				return
			}

			reply := make(chan todo.Response)
			actor <- todo.Request{Op: "delete", UserID: user, Index: idv, ReplyCh: reply}
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent) //204
	}
}

func FindByID(actor chan todo.Request) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		slog.SetDefault(LoggerFromContext(r.Context()))
		user := r.PathValue("userID")

		slog.Debug("received request to find todo item", "user", user)
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
		actor <- todo.Request{Op: "get", UserID: user, Index: intIndex, ReplyCh: reply}
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
		user := r.PathValue("userID")

		reply := make(chan todo.Response)
		actor <- todo.Request{Op: "list", UserID: user, ReplyCh: reply}
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
