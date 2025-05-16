package todo

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func RunREPL(actor chan Request, userID string) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Welcome to TODO REPL—type ‘help’ for commands.")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break // EOF or error
		}
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		cmd, args := parts[0], parts[1:]
		switch cmd {
		case "add":
			if len(args) == 0 {
				fmt.Println("Usage: add <description>")
				continue
			}
			desc := strings.Join(args, " ")
			AddItem(actor, userID, desc)
			continue
		case "bye", "quit", "exit":
			fmt.Println("Bye !!")
			return
		case "list":
			getList(actor, userID)
			continue
		case "help":
			fmt.Println("Commands: add, list, update, delete, exit")
		case "update":
			if len(args) < 2 {
				fmt.Println("Usage: update <index> <new description>")
				continue
			}
			idx, err := strconv.Atoi(args[0])
			if err != err {
				fmt.Println("Invalid Index - Usage: update 0 <new description>")
				continue
			}
			newDesc := strings.Join(args[1:], " ")
			UpdateItem(actor, userID, idx, newDesc)
			continue
		case "delete", "remove":
			if len(args) == 0 {
				fmt.Println("Usage: delete <Index>")
				continue
			}
			idx, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Println("Invalid Index - Usage: update 0 <new description>", err)
				continue
			}
			DeleteItem(actor, userID, idx)
			continue
		default:
			fmt.Println("Bad command")
			continue
		}

	}

}

func AddItem(actor chan Request, user string, desc string) {
	reply := make(chan Response)
	actor <- Request{Op: "add", UserID: user, Task: ToDoTask{Description: desc, Status: "Not Started"}, ReplyCh: reply}
	res := <-reply
	if res.Err != nil {
		fmt.Println("failed to process this request: ", res.Err)
		return
	}
	fmt.Printf("New item added, Description : %v, Status: Not Started \n", res.Task.Description)
}

func UpdateItem(actor chan Request, user string, idx int, newDesc string) {
	reply := make(chan Response)
	actor <- Request{Op: "update", UserID: user, Task: ToDoTask{Description: newDesc, Status: "Not Started"}, Index: idx, ReplyCh: reply}
	res := <-reply
	if res.Err != nil {
		fmt.Println("failed to process this request: ", res.Err)
		return
	}
	fmt.Printf("Updated item, New Description : %v \n", res.Task.Description)
}

func DeleteItem(actor chan Request, user string, idx int) {
	reply := make(chan Response)
	actor <- Request{Op: "delete", UserID: user, Index: idx, ReplyCh: reply}
	res := <-reply
	if res.Err != nil {
		fmt.Println("failed to process this request: ", res.Err)
		return
	}
	fmt.Printf("Item Deleted !!  \n")
}

func getList(actor chan Request, user string) {
	reply := make(chan Response)
	actor <- Request{Op: "list", UserID: user, ReplyCh: reply}
	res := <-reply
	if res.Err != nil {
		fmt.Println("failed to process this request: ", res.Err)
		return
	}
	fmt.Printf("TODO List: \n")
	for i, v := range res.Tasks {
		fmt.Printf("%d. %v, %v \n", i, v.Description, v.Status)
	}
}
