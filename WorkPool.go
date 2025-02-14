package workerpool

import (
	"fmt"
	"time"
)

type WorkerPool[taskResult any] struct {
	currentWorker int
	maxWorker     int
	resultChan    chan taskResult
	newTaskChan   chan []any
	doneChan      chan bool
}

func (pool *WorkerPool[taskResult]) HasWorker() bool {
	return pool.currentWorker < pool.maxWorker
}
func (pool *WorkerPool[taskResult]) NewWorker(params ...any) {
	pool.newTaskChan <- params
}
func (pool *WorkerPool[taskResult]) SendResult(result taskResult) {
	pool.resultChan <- result
}
func (pool *WorkerPool[taskResult]) WorkDone() {
	pool.doneChan <- true
}
func CreatePool[taskResult any](maxWorker int) *WorkerPool[taskResult] {
	obj := &WorkerPool[taskResult]{
		currentWorker: 0,
		maxWorker:     maxWorker,
		resultChan:    make(chan taskResult),
		newTaskChan:   make(chan []any),
		doneChan:      make(chan bool),
	}
	return obj
}

type FuncAdapter[taskResult any] func(*WorkerPool[taskResult], bool, ...any)

func (pool *WorkerPool[taskResult]) Run(task FuncAdapter[taskResult], funcResultHandler func(taskResult), params ...any) {
	waitForWorkers := func() {
		for {
			select {
			case ps := <-pool.newTaskChan:
				pool.currentWorker++
				go task(pool, true, ps...)
			case result := <-pool.resultChan:
				funcResultHandler(result)
			case <-pool.doneChan:
				pool.currentWorker--
				if pool.currentWorker == 0 {
					return
				}
			}
		}
	}

	start := time.Now()
	pool.currentWorker = 1
	go task(pool, true, params...)
	waitForWorkers()
	fmt.Println(time.Since(start))
}
