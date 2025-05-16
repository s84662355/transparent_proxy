package Queue

// DequeueFunc 处理函数类型（需在包外定义时实现）
// 参数：
// - t: 出队的元素值
// - isClose: 队列当前是否已关闭（可能在处理过程中关闭）
// 返回：bool - 是否继续处理（true-继续，false-终止）
type DequeueFunc[T any] func(t T, isClose bool) bool

type Queue[T any] interface {
	Close()
	Enqueue(T) error
	Dequeue() (t T, ok bool, isClose bool)
	DequeueWait() (t T, ok bool, isClose bool)
	DequeueFunc(fn DequeueFunc[T]) (err error)
	Count() int64
	Status() bool
}
