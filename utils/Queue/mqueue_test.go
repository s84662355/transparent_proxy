package Queue

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	//"time"
)

// go test -run TestQueue -v
func TestQueue(t *testing.T) {
	// 创建一个存储整数的无锁队列
	var queue Queue[int] = nil

	queue = NewMQueue[int]()
	var wg, wg1 sync.WaitGroup

	var count atomic.Int64

	// 并发入队操作
	numEnqueuers := 1000000
	wg.Add(numEnqueuers)
	go func() {
		for i := 1; i <= numEnqueuers; i++ {
			go func(id int) {
				defer wg.Done()
				queue.Enqueue(1)
				// for i := 0; i < 5; i++ {
				// 	queue.Enqueue(id)
				// }
			}(i)
		}
	}()

	// 并发出队操作
	numDequeuers := 500
	wg1.Add(numDequeuers)
	go func() {
		for i := 0; i < numDequeuers; i++ {
			go func(id int) {
				defer wg1.Done()
				for {
					if value, ok, _ := queue.DequeueWait(); ok {
						//	fmt.Printf("Dequeuer %d removed %d from the queue\n", id, value)
						count.Add(int64(value))
					} else {
						return
					}
				}
			}(i)
		}
	}()

	// 等待所有 goroutine 完成
	wg.Wait()
	queue.Close()
	wg1.Wait()
	fmt.Println(queue.Count(), count.Load())
}

// go test -run TestQueueWait -v
func TestQueueWait(t *testing.T) {
	// 创建一个存储整数的无锁队列
	var queue Queue[int] = nil

	queue = NewMQueue[int]()
	var wg, wg1 sync.WaitGroup

	var count atomic.Int64
	// time.Sleep(1* time.Second)
	// 并发入队操作
	numEnqueuers := 10000
	wg.Add(numEnqueuers)
	go func() {
		for i := 1; i <= numEnqueuers; i++ {
			go func(id int) {
				defer wg.Done()
				// time.Sleep(1* time.Second)
				queue.Enqueue(1)
			}(i)
		}
	}()

	// 并发出队操作
	numDequeuers := 1
	wg1.Add(numDequeuers)
	go func() {
		for i := 0; i < numDequeuers; i++ {
			go func(id int) {
				defer wg1.Done()
				fmt.Printf("开始")
				for {
					if value, ok, isClose := queue.DequeueWait(); ok {
						// fmt.Printf("Dequeuer %d removed %d from the queue\n", id, value)
						count.Add(int64(value))
					} else if isClose {
						return
					}
				}
			}(i)
		}
	}()

	// 等待所有 goroutine 完成
	wg.Wait()
	queue.Close()
	wg1.Wait()
	fmt.Println(queue.Count(), count.Load())
}
