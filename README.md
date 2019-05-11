# deadpool

A simple workerpool in go-lang that allows live resizing and a custom monitoring metric
that can be used to adjust the size of the pool as well as manual resizing

## Features

* Set initial size of pool
* Resize pool at any time without losing jobs safely
* Supports contexts
* Simple and easy to understand
* Task interface super easy to use
* Optional monitor function to run at a specified interval

## Example

```go
package mypackage

import(
	"fmt"

	deadpool "github.com/fprintf/deadpool"
)

type Job struct {
	name string
}

func (j Job) Run() error {
	// do something with j.name or whatever
	fmt.Println("running job", j.name)
}

func main() {
	// A new worker pool with 5 workers
	wp := deadpool.NewWorkerPool(5)
	ctx, cancel := context.WithCancel(context.Background())

	// Note this will block/deadlock if not run in a go routine
	go func() {
		// Cancel runs when we're done sending jobs
		// This will cause all jobs to be processed and then exit
		defer cancel()
		wp.Enqueue(Job{"job1"})
		wp.Enqueue(Job{"job2"})
	}()

	// Actually start and wait for the workers now
	wp.Wait(ctx)
}

```

## Example using a Monitor

```go
package mypackage

import(
	"fmt"
	"time"

	"github.com/fprintf/deadpool"
)

type Job struct {
	name string
}

func (j Job) Run() error {
	// do something with j.name or whatever
	fmt.Println("running job", j.name)
}

// The monitor type we wish to use
// note this can contain any arbitrary information
// and the object simply has to implement the interface
// (a Check() function as described below)
type Monitor struct {
	title string
}

// Implement the interface required for the monitor object
// This is a function called Check() which takes the current
// worker pool size and returns the same or new size (smaller or larger)
// plus an optional error indicating if the pool should be resized
func (m Monitor) Check(size int) (int, error) {
	fmt.Println(m.title, "check running, current pool size:", size)
	return size, nil
}

func main() {
	// A new worker pool with 5 workers
	monitor := Monitor{"simple report"}
	wp := deadpool.NewWorkerPoolWithMonitor(5, monitor, 5 * time.Second)
	ctx, cancel := context.WithCancel(context.Background())

	// Note this will block/deadlock if not run in a go routine
	go func() {
		// Cancel runs when we're done sending jobs
		// This will cause all jobs to be processed and then exit
		defer cancel()
		wp.Enqueue(Job{"job1"})
		wp.Enqueue(Job{"job2"})
	}()

	// Actually start and wait for the workers now
	wp.Wait(ctx)
}
```
