package todo

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
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

func SaveFile(tasks []ToDoTask, todoFile string) error {
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
