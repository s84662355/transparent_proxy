package Queue

import (
	"fmt"
	"sync"
)

// MQueue 是一个泛型队列结构体，用于存储任意类型的数据。
// 它使用链表实现，支持并发安全的入队和出队操作，并且提供了阻塞和非阻塞的出队方式。
type MQueue[T any] struct {
	head      *node[T]     // 队列的头节点指针，指向队列的第一个元素。
	tail      *node[T]     // 队列的尾节点指针，指向队列的最后一个元素。
	status    bool         // 队列的状态，true 表示队列处于打开状态，false 表示队列已关闭。
	count     int64        // 队列中元素的数量。
	recvLock  sync.RWMutex // 读写锁，用于保证并发操作时的线程安全。
	nodePool  sync.Pool    // 节点对象池，用于复用节点，减少内存分配和垃圾回收的开销。
	zeroValue T            // 泛型类型的零值，用于在出队时重置节点的值。
	recvCond  *sync.Cond   // 条件变量，用于在队列为空时阻塞出队操作，直到有新元素入队或队列关闭。
}

// node 是队列中每个节点的结构体，包含一个泛型类型的值和指向下一个节点的指针。
type node[T any] struct {
	value T        // 节点存储的值。
	next  *node[T] // 指向下一个节点的指针。
}

// 新建队列，返回一个空队列
// NewMQueue 函数用于创建一个新的 MQueue 实例，初始化队列的状态、元素数量、条件变量和节点对象池。
func NewMQueue[T any]() *MQueue[T] {
	q := &MQueue[T]{}
	q.status = true                        // 初始化队列状态为打开。
	q.count = 0                            // 初始化队列元素数量为 0。
	q.recvCond = sync.NewCond(&q.recvLock) // 创建条件变量，并关联读写锁。
	q.nodePool = sync.Pool{
		// 当对象池中没有可用节点时，使用 New 函数创建一个新的节点。
		New: func() any {
			return &node[T]{
				value: q.zeroValue, // 初始化节点的值为泛型类型的零值。
				next:  nil,         // 初始化节点的下一个节点指针为 nil。
			}
		},
	}

	return q
}

// Close 方法用于关闭队列，将队列状态设置为 false，并广播通知所有等待的 goroutine。
func (q *MQueue[T]) Close() {
	q.recvLock.Lock()
	defer q.recvLock.Unlock()
	q.status = false       // 设置队列状态为关闭。
	q.recvCond.Broadcast() // 广播通知所有等待的 goroutine，队列状态已改变。
}

// 插入，将给定的值v放在队列的尾部
// Enqueue 方法用于将一个值 v 插入到队列的尾部。
// 如果队列已关闭，返回一个错误。
func (q *MQueue[T]) Enqueue(v T) error {
	q.recvLock.Lock()
	defer q.recvLock.Unlock()

	if !q.status {
		return fmt.Errorf("is close") // 如果队列已关闭，返回错误信息。
	}

	n := q.nodePool.Get().(*node[T]) // 从对象池中获取一个节点。
	n.value = v                      // 设置节点的值为 v。
	n.next = nil                     // 设置节点的下一个节点指针为 nil。

	if q.head == nil {
		q.head = n // 如果队列为空，将头节点和尾节点都指向新节点。
	} else {
		if q.tail == nil {
			q.tail = n           // 如果队列只有一个元素，更新尾节点为新节点。
			q.head.next = q.tail // 将头节点的下一个节点指针指向尾节点。
		} else {
			oldTail := q.tail // 保存旧的尾节点。
			oldTail.next = n  // 将旧尾节点的下一个节点指针指向新节点。
			q.tail = n        // 更新尾节点为新节点。
		}
	}

	q.count++              // 队列元素数量加 1。
	q.recvCond.Broadcast() // 广播通知所有等待的 goroutine，队列中有新元素入队。
	return nil
}

// /不阻塞
// Dequeue 方法是一个非阻塞的出队方法，调用 dequeue 方法进行出队操作。
func (q *MQueue[T]) Dequeue() (t T, ok bool, isClose bool) {
	t, ok, isClose = q.dequeue()
	return
}

// 移除，删除并返回队列头部的值,如果队列为空，则返回nil
// dequeue 方法是一个私有方法，用于执行实际的出队操作。
// 返回出队的值、是否成功出队的标志和队列是否已关闭的标志。
func (q *MQueue[T]) dequeue() (t T, ok bool, isClose bool) {
	q.recvLock.Lock()
	defer q.recvLock.Unlock()
	isClose = !q.status // 获取队列是否已关闭的标志。
	if q.head == nil {
		t = q.zeroValue // 如果队列为空，返回泛型类型的零值。
		return
	} else {

		oldHead := q.head // 保存旧的头节点。
		if oldHead.next == nil {
			q.head = nil // 如果队列只有一个元素，将头节点和尾节点都置为 nil。
		} else {
			q.head = oldHead.next // 更新头节点为旧头节点的下一个节点。
			if q.head == q.tail {
				q.tail = nil // 如果新的头节点是尾节点，将尾节点置为 nil。
			}
		}

		ok = true                   // 标记出队成功。
		q.count--                   // 队列元素数量减 1。
		t = oldHead.value           // 获取旧头节点的值。
		oldHead.value = q.zeroValue // 将旧头节点的值重置为泛型类型的零值。
		oldHead.next = nil          // 将旧头节点的下一个节点指针置为 nil。
		q.nodePool.Put(oldHead)     // 将旧头节点放回对象池，以便复用。

		return
	}
}

// /阻塞    返回值t
// DequeueWait 方法是一个阻塞的出队方法，会一直等待直到有元素出队或队列关闭。
// 返回出队的值、是否成功出队的标志和队列是否已关闭的标志。
func (q *MQueue[T]) DequeueWait() (t T, ok bool, isClose bool) {
	for {
		t, ok, isClose = q.dequeue() // 尝试出队。
		if ok {
			return // 如果出队成功，返回结果。
		}

		if isClose {
			return // 如果队列已关闭，返回结果。
		}

		q.recvLock.Lock()
		if q.status && q.count == 0 {
			q.recvCond.Wait() // 如果队列处于打开状态且为空，阻塞等待。
		}
		q.recvLock.Unlock()
	}
}

// DequeueFunc 方法是一个阻塞的出队方法，会不断出队元素并调用传入的函数 fn 进行处理。
// 直到 fn 函数返回 false 或队列关闭且为空。
// 返回一个错误信息，如果队列关闭且为空，返回相应的错误。
func (q *MQueue[T]) DequeueFunc(fn DequeueFunc[T]) (err error) {
	for {

		t, ok, isClose := q.dequeue() // 尝试出队。
		if ok {
			if !fn(t, isClose) {
				return // 如果 fn 函数返回 false，停止出队并返回。
			}
		} else if isClose {
			return fmt.Errorf("queue is close and empty") // 如果队列关闭且为空，返回错误信息。
		}

		q.recvLock.Lock()

		if q.status && q.count == 0 {
			q.recvCond.Wait() // 如果队列处于打开状态且为空，阻塞等待。
		}

		q.recvLock.Unlock()
	}
}

// Count 方法用于获取队列中元素的数量，使用读锁保证并发安全。
func (q *MQueue[T]) Count() int64 {
	q.recvLock.RLock()
	defer q.recvLock.RUnlock()
	return q.count
}

// Status 方法用于获取队列的状态，使用读锁保证并发安全。
func (q *MQueue[T]) Status() bool {
	q.recvLock.RLock()
	defer q.recvLock.RUnlock()
	return q.status
}
