# Workers

Yet another worker pool library.
However, this library implements a singleton pattern. Only a single worker pool object can reside in the heap.

Error handling mechanism is BYOC (Bring Your Own Channel). The goroutine leak for the error channel is handled by the library.
The client gets the liberty to do whatever else they want with it.

The user gets to tune the number of workers and the buffer size of task queue.

Raising constructive issues and PRs would be appreciated :)

#### DEMO APP

```golang
package main

import (
	"fmt"
	"time"

	"errors"

	"github.com/dev-gaur/workers"
)

var newtask = workers.NewTask

var poolSize, queueSize int

func init() {
	poolSize = 4
	queueSize = 2
}

func main() {
	errChan := make(chan error, 10)
	wrap := make(chan struct{})
	pooldone := make(chan struct{})

	fmt.Println("ENTER 0 into console STDIN to exit")

	/*
	 * Invoking and fetching a refernce to the workerpool instance
	 * poolSize : determines the number of workers you want the pool to have
	 * errChan : the channel on which the worker pool broadcasts error caused by any task. To see usage, check the goroutine on line 51
	 * wrap : channel used by the client to signal pool to shutdown
	 * pooldone : channel used by the worker pool to confirm shut down after recieving wrap signal
	 * queueSize : length of the task queue
	 */
	pool := workers.GetPool(poolSize, errChan, wrap, pooldone, queueSize)

	done := make(chan bool)

	// Goroutine to listen to the STDIN. The program shuts down when a '0' is read from STDIN
	go func() {
		for {
			var c int
			fmt.Scanf("%d", c)

			if c == 0 {
				done <- true
			}
		}
	}()

	// goroutine to listen on the error chan
	go func() {
		for err := range errChan {
			fmt.Println("Error reported :", err)
		}
	}()

	//Adding 3 tasks to the worker pool
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

	//Adding 3 tasks to the worker pool
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

	//Adding 3 tasks to the worker pool. One of them has error.
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

	task := newtask(func() error {
		return randomfunc("gopher inside!")
	})

	// Adding the last task.
	pool.AssignTask(task)

	// This blocks till the user signals to quit by inputting '0' to STDIN
	<-done
	// signalling the worker pool to shutdown
	close(wrap)
	// This will block tills the shutdown completes and worker pool signals back
	<-pooldone

}

func randomfunc(str string) error {
	fmt.Println("random function invoked with argument :", str)

	return errors.New("Error from random function")
}
```
