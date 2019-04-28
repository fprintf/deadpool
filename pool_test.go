package worker_pool

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"
)

// TestData used for running unit tests that require
// an interface parameter
type TestData struct {
	message string
	mut     *sync.Mutex
	count   *int
}

func (td TestData) Run() error {
	var err error
	if td.mut != nil {
		td.mut.Lock()
		*td.count++
		td.mut.Unlock()
	}
	return err
}

// BenchData used for benchmarking
type BenchData struct {
	message string
	b       *testing.B
}

var nothingValue = 0

func (bd BenchData) Run() error {
	nothingValue++
	return nil
}

type NullWriter struct{}

func (NullWriter) Write([]byte) (int, error) {
	return 0, nil
}

// Disable logging from out
func init() {
	log.SetOutput(NullWriter{})
}

// TestPool tests the overall functionality as well as
// stopping and starting the pool twice
// ensures all accepted jobs are processed before the pool leaves
func TestPool(t *testing.T) {
	numWorkers := 2
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
	ctx, cancel = context.WithCancel(context.Background())
	// Add a job to the queue
	go func() {
		for i := 0; i < numJobs; i++ {
			wp.Enqueue(TestData{
				message: fmt.Sprintf("this is a test round 2 #%02d", i),
				mut:     mut,
				count:   &count,
			})
			// Also test resize by just running it randomly on each loop
			// this should work fine and not cause any data to be lost
			//wp.Resize(i + 1)
		}
		cancel()
	}()

	wp.Wait(ctx)
	if count != numJobs*2 {
		t.Error("got completed job count:", count, "but expected:", numJobs*2)
	}
}

// Test if the jobqueue suddenly closes on us
func TestCloseJobQueue(t *testing.T) {
	numJobs := 10
	wp := NewWorkerPool(5)
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
		}
		close(wp.jobqueue)
		cancel()
	}()
	wp.Wait(ctx)
	if count != numJobs {
		t.Error("got completed job count:", count, "but expected:", numJobs)
	}
}

type PrintCheck struct {
	message string
}

func (pc PrintCheck) Check(size int) (int, error) {
	fmt.Println(pc.message)
	return size, nil
}

func TestPoolWithMonitor(t *testing.T) {
	numJobs := 10
	numWorkers := 2
	wp := NewWorkerPoolWithMonitor(numWorkers, PrintCheck{"test monitor"}, time.Second*1)
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
		}
		cancel()
	}()
	wp.Wait(ctx)
	ctx, cancel = context.WithCancel(context.Background())
	// Add a job to the queue
	go func() {
		for i := 0; i < numJobs; i++ {
			wp.Enqueue(TestData{
				message: fmt.Sprintf("this is a test round 2 #%02d", i),
				mut:     mut,
				count:   &count,
			})
		}
		cancel()
	}()
	wp.Wait(ctx)
	if count != numJobs*2 {
		t.Error("got completed job count:", count, "but expected:", numJobs*2)
	}
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
}
