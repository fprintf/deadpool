# deadpool

A simple workerpool in go-lang that allows live resizing and a custom monitoring metric
that can be used to adjust the size of the pool as well as manual resizing

## Features

* Set initial size of pool
* Resize pool at any time without losing jobs safely
* Supports contexts
* Simple and easy to understand
* Task interface super easy to use

## Example

```go
package mypackage

import(
	"fmt"

	worker_pool "github.com/fprintf/worker_pool"
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
	wp := NewWorkerPool(5)
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
