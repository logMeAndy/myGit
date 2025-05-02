package todo

import (
	"testing"
)

func TestSaveAndLoadTasks(t *testing.T) {
	tmpFile := "test_todo.json"
	// The file does not existing
	//	loaded, err := LoadFile(tmpFile)
	//	if err != nil {
	//		t.Fatalf("LoadFile() expected error = %v, got : %v", "nil", err)
	//	}

	var tasks []ToDoTask
	tasks = append(tasks, ToDoTask{Description: "Test", Status: "started"})

	err := SaveFile(tasks, tmpFile)
	if err != nil {
		t.Errorf("SaveFile() expected error = %v, got : %v", "nil", err)
	}

	loaded, err := LoadFile(tmpFile)
	if err != nil {
		t.Fatalf("LoadFile() expected error = %v, got : %v", "nil", err)
	}

	if loaded[0].Description != "Test" {
		t.Errorf("expected 'Test', got '%s'", loaded[0].Description)
	}

}
