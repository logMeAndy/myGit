package todo

import (
	"flag"
	"fmt"
	"log/slog"
)

func Run(args []string, todoList []ToDoTask) ([]ToDoTask, error) {
	fs := flag.NewFlagSet("todo", flag.ContinueOnError)

	var taskDesc = fs.String("task", "", "Task description e.g. -task=newItemDescription (optional -status=newStatus) (default not started))")
	var status = fs.String("status", "", "New status e.g. not started, completed, started, etc.,")
	var updateIndex = fs.Int("update", -1, "Index of task to update, e.g. update=0 -task=newValue (optional -status=newStatus)")
	var deleteIndex = fs.Int("delete", -1, "Index of task to delete (e.g. delete=0 )")

	fs.Parse(args[1:]) // skip program name

	slog.Debug("args", "deleteIndex", deleteIndex, "updateIndex", updateIndex, "status", status, "taskDesc", taskDesc)
	switch {
	case *updateIndex >= 0:
		slog.Debug("updating index...", slog.Int("updateIndex", *updateIndex))
		if *taskDesc != "" && *updateIndex < len(todoList) {
			if *status != "" {
				todoList[*updateIndex].Description = *taskDesc
				todoList[*updateIndex].Status = *status
			} else {
				todoList[*updateIndex].Description = *taskDesc
			}
			slog.Info("Task updated", "Index", todoList[*updateIndex])
			return todoList, nil
		} else if *status != "" && *updateIndex < len(todoList) {
			todoList[*updateIndex].Status = *status
			slog.Info("Task updated", "Index", todoList[*updateIndex])
			return todoList, nil
		} else {
			slog.Info("Invalid index or missing task description, please check current TODO list or help using -h", "Index", *updateIndex, "Length", len(todoList))
			return todoList, fmt.Errorf("404") //404 not found
		}
	case *taskDesc != "":
		slog.Debug("adding task...", "task", *taskDesc)
		var t ToDoTask
		if *status != "" {
			t = ToDoTask{Description: *taskDesc, Status: *status}
			todoList = append(todoList, t)
		} else {
			t = ToDoTask{Description: *taskDesc, Status: "not started"}
			todoList = append(todoList, t)
		}
		slog.Info("Added new task", "task", t)
		return todoList, nil
	case *deleteIndex >= 0 && *deleteIndex < len(todoList):
		slog.Debug("deleting index...", "task", *deleteIndex)
		if *deleteIndex < len(todoList) {
			deletedTask := todoList[*deleteIndex]
			todoList = append(todoList[:*deleteIndex], todoList[*deleteIndex+1:]...)
			slog.Info("Deleted task", "index", *deleteIndex, "task", deletedTask)
			return todoList, nil
		} else {
			slog.Info("Invalid index or missing task description, please check current TODO list or help using -h", "Index", *deleteIndex, "Length", len(todoList))
			return todoList, fmt.Errorf("404") //404 not found
		}
	default:
		fmt.Println("Current To-Do List :")
		for i, t := range todoList {
			fmt.Printf("%d. %s, %s \n", i, t.Description, t.Status)
		}
	}
	return todoList, nil
}
