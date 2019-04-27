package worker_pool

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

type TestData struct {
	message string
	mut     *sync.Mutex
	count   *int
}

func (td TestData) Run() error {
	var err error
	if td.mut != nil {
		td.mut.Lock()
		*td.count += 1
		td.mut.Unlock()
	}
	return err
}

type BenchData struct {
	message string
	b       *testing.B
}

var nothing_value = 0

func (bd BenchData) Run() error {
	nothing_value += 1
	return nil
}

func TestPool(t *testing.T) {
	numWorkers := 10
	numJobs := 10
	wp := NewWorkerPool(numWorkers)
	ctx, cancel := context.WithCancel(context.Background())
	mut := &sync.Mutex{}
	count := 0
	// Add a job to the queue
	go func() {
		for i := 0; i < numJobs; i++ {
			wp.Enqueue(TestData{
				message: fmt.Sprintf("this is a test #%02d", i),
				mut:     mut,
				count:   &count,
			})
			// Also test resize by just running it randomly on each loop
			// this should work fine and not cause any data to be lost
			wp.Resize(i + 1)
		}
		cancel()
	}()

	wp.Wait(ctx)
	// All go routines should have exited and jobs completed by now
	if count != numJobs {
		t.Error("got completed job count:", count, "but expected:", numJobs)
	}
}

// Test if the jobqueue suddenly closes on us
func TestCloseJobQueue(t *testing.T) {
	wp := NewWorkerPool(1)
	ctx, cancel := context.WithCancel(context.Background())
	// Add a job to the queue
	go func() {
		wp.Enqueue(TestData{message: "this is a test"})
		wp.Enqueue(TestData{message: "this is a test2"})
		wp.Enqueue(TestData{message: "this is a test3"})
		wp.Enqueue(TestData{message: "this is a test4"})
		close(wp.jobqueue)
		cancel()
	}()
	wp.Wait(ctx)
}

func BenchmarkPool(b *testing.B) {
	numWorkers := 10
	numJobs := b.N
	wp := NewWorkerPool(numWorkers)
	ctx, cancel := context.WithCancel(context.Background())
	// Add a job to the queue
	go func() {
		defer cancel()
		for i := 0; i < numJobs; i++ {
			wp.Enqueue(BenchData{
				message: fmt.Sprintf("this is a test #%02d", i),
				b:       b,
			})
		}
	}()

	wp.Wait(ctx)
	b.Log("Processed", numJobs, "jobs")
}
