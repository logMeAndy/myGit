package main

import (
	"to-do/todo"
)

var actor = todo.ReqChan

func main() {
	initialTasks, _ := todo.LoadFile(todo.TodoFile)
	go todo.Actor(initialTasks)
	todo.RunREPL(actor)
}
