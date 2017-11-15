package workers

import (
	"errors"
	"fmt"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
)

var errs []error
var errChan = make(chan error)
var wrap = make(chan struct{})
var done = make(chan struct{})

var mypool = GetPool(5, errChan, wrap, done, 2)

func init() {

	go func() {
		var err error
		for err = range errChan {
			errs = append(errs, err)
		}
	}()
}

func TestEmptyPool(t *testing.T) {

	assert.Equal(t, 0, len(errs))
	assert.Equal(t, []error(nil), errs)
}

func TestWithWork(t *testing.T) {
	tasks := []*Task{
		NewTask(func() error { return nil }),
		NewTask(func() error { return nil }),
		NewTask(func() error { return nil }),
	}
	mypool.AssignTasks(tasks)

	assert.Equal(t, []error(nil), errs)
	assert.Equal(t, 0, len(errs))
}

func TestWithError(t *testing.T) {
	tasks := []*Task{
		NewTask(func() error { return fmt.Errorf("error") }),
		NewTask(func() error { return nil }),
		NewTask(func() error { return nil }),
	}

	mypool.AssignTasks(tasks)

	time.Sleep(time.Second)

	fmt.Println(errs)

	assert.Equal(t, errors.New("error"), errs[0])
	assert.Equal(t, 1, len(errs))

	close(wrap)

	<-done
}
