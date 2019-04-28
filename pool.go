package worker_pool

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

// Monitor interface specifies a Check function
// which is used to determine if the worker pool needs
// to be resized
type Monitor interface {
	Check(int) (int, error)
}

// WorkerPool maintains contexts and waitgroups on the workers
// as well as managing launching and removing workers
type WorkerPool struct {
	size     int
	jobqueue chan Task
	cancel   context.CancelFunc
	ctx      context.Context
	wg       *sync.WaitGroup
	monitor  Monitor
	checkDur time.Duration
}

// Task defines an object that has a Run() function
type Task interface {
	Run() error
}

// NewWorkerPool creates and returns a
// new WokerPool object properly initialized
func NewWorkerPool(size int) WorkerPool {
	sizeQueue := 0
	wp := WorkerPool{
		size:     size,
		jobqueue: make(chan Task, sizeQueue),
		wg:       &sync.WaitGroup{},
	}
	return wp
}

// NewWorkerPoolWithMonitor creates and returns a WorkerPool
// object that also checks a monitor every checkDur duration
func NewWorkerPoolWithMonitor(size int, monitor Monitor, checkDur time.Duration) WorkerPool {
	sizeQueue := 0
	wp := WorkerPool{
		size:     size,
		jobqueue: make(chan Task, sizeQueue),
		wg:       &sync.WaitGroup{},
		monitor:  monitor,
		checkDur: checkDur,
	}
	return wp
}

func (wp *WorkerPool) addWorker() {
	wp.wg.Add(1)
	go func() {
		defer wp.wg.Done()
		id := fmt.Sprintf("%d", rand.Int63())[0:4]
		log.Printf("Worker %s starting", id)
		for {
			select {
			case entry, ok := <-wp.jobqueue:
				if !ok {
					/* TODO Log error from reading closed jobqueue */
					log.Printf("Worker %s exiting read on closed channel", id)
					return
				}
				err := entry.Run()
				if err != nil {
					/* TODO Log error or perhaps send it to another error channel? */
				}
			case <-wp.ctx.Done():
				/* Our context is closed, exit */
				err := wp.ctx.Err()
				if err == nil {
					err = errors.New("context has no error..?")
				}
				log.Printf("Worker %s exiting %s", id, err)
				return
			}
		}
	}()
}

// Resize the workerpool by telling existing workers to shutdown
// and launching a new set of workers (not necessarily in that order)
func (wp *WorkerPool) Resize(size int) {
	wp.stopWorkers()
	wp.ctx, wp.cancel = context.WithCancel(context.Background())
	for count := 0; count < size; count++ {
		wp.addWorker()
	}
	wp.size = size
}

// Enqueue adds a new task to the WokerPool job queue
func (wp *WorkerPool) Enqueue(task Task) {
	wp.jobqueue <- task
}

// Wait launches the workers and waits for it's context to be canceled
// or some other issue (such as the jobqueue closing unexpectedly) or for
// all workers to exit
func (wp *WorkerPool) Wait(ctx context.Context) {
	wp.Resize(wp.size) // Initial launch of workers
	defer wp.wg.Wait() // wait for workers to exit properly

	// Begin monitor loop to check for workers
	if wp.monitor != nil {
		go func(ctx context.Context) {
			monitor := wp.monitor
			for {
				select {
				case <-time.After(wp.checkDur):
					size, err := monitor.Check(wp.size)
					if err != nil {
						log.Println("Error in monitor check:", err)
						continue
					}
					if size != wp.size {
						wp.Resize(size)
					}
				case <-ctx.Done():
					return
				}
			}
		}(ctx)
	}

	// Wait for us to be told we're done
	for {
		select {
		case <-ctx.Done():
			wp.stopWorkers()
			return
		}
	}
}

func (wp *WorkerPool) stopWorkers() {
	if wp.cancel != nil {
		wp.cancel()
	}
}
