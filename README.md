# Workers

Yet another worker pool library.
However, this library implements a singleton pattern. Only a single worker pool object can reside in the heap.

Error handling mechanism is BYOC (Bring Your Own Channel). The goroutine leak for the error channel is handled by the library.
The client gets the liberty to do whatever else they want with it.

The user gets to tune the number of workers and the buffer size of task queue.

Raising constructive issues and PRs would be appreciated :)

#### DEMO APP

```
package main

import (
	"fmt"
	"github.com/dev-gaur/workers"
	"time"

	"errors"
)

var newtask = workers.NewTask

// size of the worker pool
var poolSize int

func init() {
	poolSize = 10
}

func main() {
	errChan := make(chan error, 10)
	wrap := make(chan struct{})
	var pooldone = make(chan struct{})

	pool := workers.GetPool(poolSize, errChan, wrap, pooldone, 90)
	done := make(chan bool)

	go func() {
		for {
			var c int
			fmt.Scanf("%d", c)

			if c == 0 {
				done <- true
			}
		}
	}()

	go func() {
		for err := range errChan {
			fmt.Println("Error reported :", err)
		}
	}()

	tasks := []*workers.Task{
		newtask(func() error {
			fmt.Println("One")
			return nil
		}),
		newtask(func() error {
			fmt.Println("Two")
			return nil
		}),
		newtask(func() error {
			fmt.Println("Three")
			return nil
		}),
	}

	pool.AssignTasks(tasks)

	time.Sleep(time.Second * 2)

	tasks = []*workers.Task{
		newtask(func() error {
			fmt.Println("Four")
			return nil
		}),
		newtask(func() error {
			fmt.Println("Five")
			return nil
		}),
		newtask(func() error {
			fmt.Println("Six")
			return nil
		}),
	}

	pool.AssignTasks(tasks)

	time.Sleep(time.Second * 2)

	tasks = []*workers.Task{
		newtask(func() error {
			fmt.Println("Seven")
			return nil
		}),
		newtask(func() error {
			fmt.Println("Eight")
			return errors.New("CURSE OF THE EIGHTH")
		}),
		newtask(func() error {
			fmt.Println("Nine")
			return nil
		}),
	}

	pool.AssignTasks(tasks)

	close(wrap)
	<-done
	<-pooldone
	//ENTER 0 into console STDIN to exit
}
```
