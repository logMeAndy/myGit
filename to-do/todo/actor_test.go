package todo

import (
	"fmt"
	"testing"
	"time"
)

var actor = ReqChan
var start = time.Now()

func TestActorConcurrentAdded(t *testing.T) {

	done := make(chan bool)

	initialTasks, _ := LoadFile(TodoFile)

	go Actor(initialTasks)
	const n = 10000
	// launch n subtests in parallel, each adding one task

	//go func() {
	for i := 0; i < n; i++ {
		i := i // capture loop variable
		//t.Logf("[Subtest %02d] start", i)
		t.Run(fmt.Sprintf("AddTask-%02d", i), func(t *testing.T) {
			//t.Parallel()
			//t.Logf("[Subtest %02d] start", i)

			reply := make(chan Response)
			task := ToDoTask{Description: fmt.Sprintf("task-%d", i), Status: "not started"}
			//t.Logf("[Subtest %02d] sending Request{Op: \"add\", Task: %+v}", i, task)
			actor <- Request{Op: "add", Task: task, ReplyCh: reply}
			res := <-reply
			//t.Logf("[Subtest %02d] received Response{Err: %v, Task: %+v}", i, res.Err, res.Task)
			//t.Logf("[Subtest %02d] received Response{Err: %v}", i, res.Err)
			//t.Logf("[Subtest %02d] sending Request{Op: \"add\", Task: %+v}", i, task)
			if res.Err != nil {
				t.Errorf("add failed for task %d: %v", i, res.Err)
			}

		})
		t.Run(fmt.Sprintf("getTask-%02d", i), func(t *testing.T) {
			reply1 := make(chan Response)
			actor <- Request{Op: "get", Index: i, ReplyCh: reply1}
			res := <-reply1

			if res.Err != nil {
				t.Errorf("add failed for task %d: %v", i, res.Err)
			}
		})

	}
	close(done)

	t.Logf("[Time taken ----------------------------------------------------------------------------- %02d] ", time.Since(start).Milliseconds())
	//}()
	<-done
}

func TestActorVerify(t *testing.T) {
	t.Logf("[Time taken ----------------------------------------------------------------------------- %02d] ", time.Since(start).Milliseconds())

	reply := make(chan Response)
	actor <- Request{Op: "list", ReplyCh: reply}
	res := <-reply
	t.Logf("[VerifyCount] received Response.Tasks length=%d", len(res.Tasks))
	n := 10000
	if len(res.Tasks) != n {
		t.Fatalf("expected %d tasks, got %d", n, len(res.Tasks))
	}

}

/*

func TestActorConcurrentUpdated(t *testing.T) {

	done := make(chan bool)

	initialTasks, _ := LoadFile(TodoFile)

	go Actor(initialTasks) // spins up the actor goroutine
	const n = 100
	// launch n subtests in parallel, each adding one task

	go func() {
		for i := 0; i < n; i++ {
			i := i // capture loop variable
			t.Logf("[Subtest %02d] start", i)
			t.Run(fmt.Sprintf("AddTask-%02d", i), func(t *testing.T) {
				t.Parallel()
				//t.Logf("[Subtest %02d] start", i)

				reply := make(chan Response)
				task := ToDoTask{Description: fmt.Sprintf("task-%d", i), Status: "not started"}
				//t.Logf("[Subtest %02d] sending Request{Op: \"add\", Task: %+v}", i, task)
				actor <- Request{Op: "add", Task: task, ReplyCh: reply}
				res := <-reply
				//t.Logf("[Subtest %02d] received Response{Err: %v, Task: %+v}", i, res.Err, res.Task)
				t.Logf("[Subtest %02d] received Response{Err: %v}", i, res.Err)
				if res.Err != nil {
					t.Errorf("add failed for task %d: %v", i, res.Err)
				}
			})

		}
		close(done)
	}()
	<-done
}

*/
