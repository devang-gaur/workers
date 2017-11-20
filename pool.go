// Package workers : utility to create worker pools and task allocation pattern.
package workers

import (
	"fmt"
	"sync"
)

var once sync.Once

// pool is a worker group that runs a number of tasks at a configured
// concurrency.
type pool struct {
	taskQueue   chan *Task
	taskChan    chan *Task
	concurrency int // Number of workers
	errChan     chan<- error
}

// Instance : pointer to the singleton instance of the worker pool
var instance *pool

// GetPool initializes a new pool with the given tasks and at the given concurrency.
func GetPool(n int, errorChan chan<- error, wrap <-chan struct{}, done chan struct{}, queueSize int) *pool {

	if instance != nil {
		fmt.Println("Worker pool has already been invoked. Can't change the singleton Instance")
	}

	once.Do(func() {
		instance = &pool{
			taskQueue:   make(chan *Task, queueSize),
			taskChan:    make(chan *Task),
			concurrency: n,
			errChan:     errorChan,
		}

		var workersignals []chan struct{}

		for i := 0; i < instance.concurrency; i++ {
			workersignals = append(workersignals, make(chan struct{}))
		}

		for i := 0; i < instance.concurrency; i++ {
			fmt.Println("Launched a worker")
			go instance.work(workersignals[i])
		}

		go instance.run(wrap, done, workersignals)
	})

	return instance
}

// AssignTasks assigns a slice of tasks to the worker pool.
func (p *pool) AssignTasks(tasks []*Task) *pool {
	fmt.Printf("assigning %d tasks to tasksChannel\n", len(tasks))

	for _, task := range tasks {
		p.taskQueue <- task
	}

	return p
}

// AssignTask assigns a task to the worker pool.
func (p *pool) AssignTask(task *Task) *pool {
	fmt.Println("assigning a task to the Queue")

	p.taskQueue <- task

	return p
}

func (p *pool) wrap(done <-chan struct{}, signal chan<- struct{}, workersignals []chan struct{}) {

	<-done
	close(p.taskChan)

	for _, workersignal := range workersignals {
		<-workersignal
	}

	close(p.errChan)
	close(signal)
}

// run, runs all work within the pool and blocks until it's finished.
func (p *pool) run(wrap <-chan struct{}, done chan<- struct{}, workersignals []chan struct{}) {
	shutdown := make(chan struct{})
	go p.wrap(shutdown, done, workersignals)

L:
	for {
		select {

		case task := <-p.taskQueue:
			p.taskChan <- task

		case <-wrap:
			close(p.taskQueue)
			for task := range p.taskQueue {
				p.taskChan <- task
			}

			close(shutdown)
			break L

		}
	}
}

// The work loop for any single goroutine.
func (p *pool) work(done chan<- struct{}) {
	for task := range p.taskChan {
		task.run()
	}

	fmt.Println("Worker Shutting up...")
	close(done)
}

// Task encapsulates a work item that should go in a work pool.
type Task struct {
	// Err holds an error that occurred during a task. Its result is only
	// meaningful after Run has been called for the task that holds it.
	err error

	f func() error
}

// NewTask initializes a new task based on a given work function.
func NewTask(f func() error) *Task {
	return &Task{f: f}
}

// Run runs a Task and does appropriate accounting via a given sync.WorkGroup.
func (t *Task) run() {
	t.err = t.f()

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			instance.errChan <- err
		}
	}()

	if t.err != nil {
		panic(t.err)
	}
}
