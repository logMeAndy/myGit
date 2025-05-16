package todo

const TodoFile = "todo.json"

type ToDoTask struct {
	//ID          int    `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"`
}
