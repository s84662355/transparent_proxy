package taskRunManager

import (
	"sync"
	"sync/atomic"
)

const forCount = 2

type AtomicTask struct {
	l           sync.RWMutex
	atomicCount atomic.Int32
	count       int32

	c *Condition
}

func NewAtomicTask(count int) *AtomicTask {
	atom := &AtomicTask{
		count: int32(count),
	}
	atom.c = NewCondition()
	return atom
}

func (c *AtomicTask) Run(t Task) {
	for {
		count := c.atomicCount.Load()
		if count < c.count && c.atomicCount.CompareAndSwap(count, count+1) {
			go func() {
				defer c.c.Release()
				defer c.atomicCount.Add(-1)
				t()
			}()
			return
		}

		c.c.Block(func() bool {
			return c.atomicCount.Load() == c.count
		})
	}
}

func (c *AtomicTask) Wait() {
	for {
		c.c.Block(func() bool {
			return c.atomicCount.Load() != 0
		})
		if c.atomicCount.Load() == 0 {
			return
		}
	}
}

type Condition struct {
	cond *sync.Cond
}

func NewCondition() *Condition {
	s := &Condition{}
	s.cond = sync.NewCond(&sync.Mutex{})
	return s
}

func (s *Condition) Block(f func() bool) {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()
	if !f() {
		return
	}
	s.cond.Wait()
}

func (s *Condition) Release() {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()
	s.cond.Broadcast()
}
