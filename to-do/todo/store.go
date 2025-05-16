package todo

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func LoadFile(todoFile string) ([]ToDoTask, error) {
	var tasks []ToDoTask

	file, err := os.Open(todoFile)
	if err != nil {
		if os.IsNotExist(err) {
			return tasks, nil
		}
		slog.Error("Failed to open file", "error", err)
		return nil, err
	}

	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &tasks)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func LoadAllTasksInDir(dir string) (map[string][]ToDoTask, error) {
	lists := make(map[string][]ToDoTask)

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, "_"+TodoFile) {
			return nil
		}
		// extract userID from filename: <user>_todo.json
		user := strings.TrimSuffix(name, "_"+TodoFile)
		slog.Info("file path", "path", path)

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}
		var tasks []ToDoTask
		if err := json.Unmarshal(data, &tasks); err != nil {
			return fmt.Errorf("invalid JSON in %s: %w", path, err)
		}
		lists[user] = tasks
		return nil
	})
	if err != nil {
		return nil, err
	}
	return lists, nil
}

func SaveFile(tasks []ToDoTask, todoFile string) error {
	var mu sync.RWMutex
	mu.Lock()
	defer mu.Unlock()
	bytes, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		slog.Error("Failed to marshall file ", "error", err)
		return err
	}

	err = os.WriteFile(todoFile, bytes, 0644)
	if err != nil {
		slog.Error("Failed to write file ", "error", err)
		return err
	}

	return nil
}
