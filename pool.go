package worker_pool

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
)

// Worker pool maintains contexts and waitgroups on the workers
// as well as managing launching and removing workers
type WorkerPool struct {
	size     int
	jobqueue chan Task
	cancel   context.CancelFunc
	ctx      context.Context
	wg       *sync.WaitGroup
}

// Task defines an object that has a Run() function
type Task interface {
	Run() error
}

func NewWorkerPool(size int) WorkerPool {
	sizeQueue := 1
	wp := WorkerPool{
		size:     size,
		jobqueue: make(chan Task, sizeQueue),
		wg:       &sync.WaitGroup{},
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
}

func (wp *WorkerPool) Enqueue(task Task) {
	wp.jobqueue <- task
}

func (wp *WorkerPool) Wait(ctx context.Context) {
	wp.Resize(wp.size)
	defer wp.wg.Wait()

	for {
		// TODO monitor some metric here, probably another passed in function
		// that then tells us if we need to resize, up or down
		// if x { Resize(ctx, newSize) } x = loadavg or cpu usage or some user defined metric perhaps
		//		time.Sleep(time.Millisecond * 500)
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
