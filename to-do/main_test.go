package main

import (
	"fmt"
	"testing"
	"to-do/todo"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
		action   string
	}{
		{"No Task Added", []string{"cmd", ""}, "", "list"},
		{"Add Task buy apple", []string{"cmd", "-task=buy apple"}, "buy apple", "add"},
		{"Add Task buy cgi", []string{"cmd", "-task=buy cgi"}, "buy cgi", "add"},
		{"Update Task ", []string{"cmd", "-update=0", "-task=updated_task"}, "", "update"},
		{"Delete Task ", []string{"cmd", "-delete=0"}, "", "delete"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todoList, _ := todo.Run(tt.args, []todo.ToDoTask{}) // an empty list
			fmt.Println(todoList)

			if tt.expected == "" {
				if len(todoList) != 0 {
					t.Errorf("expected no tasks, got %d", len(todoList))
				}
				return
			}
			if len(todoList) == 0 { //when todolist is empty
				t.Fatalf("expected at least 1 task, got 0")
				return
			}

			if todoList[0].Description != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, todoList[0].Description)
			}

		})
	}

	var testTasks = []todo.ToDoTask{
		{Description: "Buy groceries", Status: "not started"},
		{Description: "Pay utility bills", Status: "started"},
		{Description: "Submit tax return", Status: "completed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todoList, _ := todo.Run(tt.args, testTasks)

			fmt.Println(todoList)

			if tt.expected == "" { //when no task
				if tt.action == "delete" && len(todoList) == 1 {
					t.Errorf("delete: expected '%s', got '%d'", "2", len(todoList))
				} else if tt.action != "delete" && len(todoList) != 3 {
					t.Errorf("expected no tasks, got %d", len(todoList))
				}
				return
			}
			if len(todoList) == 0 { //when todolist is empty
				t.Fatalf("expected at least 1 task, got 0")
				return
			}

			if todoList[3].Description != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, todoList[3].Description)
			}

		})
	}

}
