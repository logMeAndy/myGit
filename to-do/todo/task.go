package todo

const TodoFile = "todo.json"

type ToDoTask struct {
	Description string `json:"description"`
	Status      string `json:"status"`
}
