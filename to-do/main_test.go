package main

import (
	"testing"
	"to-do/todo"
)

//var actor = todo.ReqChan

func TestRun(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
		action   string
	}{
		{"Add Task buy apple", []string{"cmd", "-task=buy apple"}, "buy apple", "add"},
		{"Add Task buy cgi", []string{"cmd", "-task=buy cgi"}, "buy cgi", "add"},
		{"Update Task ", []string{"cmd", "-update=0", "-task=buy apple updated"}, "", "update"},
		{"No Task Added", []string{"cmd", ""}, "", "list"},
		{"Update Task ", []string{"cmd", "-update=0", "-status=completed"}, "", "update"},
		{"No Task Added", []string{"cmd", ""}, "", "list"},
		{"Delete Task ", []string{"cmd", "-delete=0"}, "", "delete"},
		{"No Task Added", []string{"cmd", ""}, "", "list"},
	}
	go todo.Actor(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log("cmd args ", tt.args)
			err := todo.Run(tt.args, actor) //
			//fmt.Println(todoList)
			if err != nil {
				t.Errorf("expected no error, got %d", err)
				return
			}

		})
	}

}
