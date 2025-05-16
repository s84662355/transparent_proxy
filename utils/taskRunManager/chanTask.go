package taskRunManager

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type ChanTask struct {
	wg    sync.WaitGroup
	tChan chan struct{}
}

func NewChanTask(count int) *ChanTask {
	return &ChanTask{tChan: make(chan struct{}, count)}
}

func (c *ChanTask) Run(t Task) {
	c.wg.Add(1)
	c.tChan <- struct{}{}
	go func() {
		defer func() {
			<-c.tChan
			c.wg.Done()
		}()
		t()
	}()
}

func (c *ChanTask) Wait() {
	c.wg.Wait()
}

type taskFunc[T, P any] func(ctx context.Context, p P) (T, bool)

func RunTaskGetSpecifyQuantityResultContext[T, P any](
	pCtx context.Context,
	ps []P,
	fn taskFunc[T, P],
	expectResultCount int,
	poolCount int,
) <-chan T {
	resChan := make(chan T, expectResultCount)
	go func() {
		closeChan := sync.OnceFunc(func() {
			close(resChan)
		})
		defer closeChan()

		ctx, cancel := context.WithCancel(pCtx)
		defer cancel()

		var resultCount, enterCount atomic.Int32

		resultCount.Add(int32(expectResultCount))
		enterCount.Add(int32(expectResultCount))
		chanTask := NewChanTask(poolCount)
		for _, pp := range ps {
			p := pp
			chanTask.Run(func() {
				v, ok := fn(ctx, p)
				if ok {
					select {
					case <-ctx.Done():
						return
					default:
						if resultCount.Add(-1) < 0 {
							cancel()
							return
						}
						select {
						case resChan <- v:
							if enterCount.Add(-1) == 0 {
								cancel()
								closeChan()
								return
							}
						default:
						}
					}
				}
			})
		}
		chanTask.Wait()
	}()
	return resChan
}

func RunTaskGetSpecifyQuantityResultTimeOutContext[T, P any](
	ctx context.Context,
	timeOut time.Duration,
	ps []P,
	fn taskFunc[T, P],
	expectResultCount int,
	poolCount int,
) <-chan T {
	resChan := make(chan T, expectResultCount)

	go func() {
		defer close(resChan)
		timeOutctx, cancel := context.WithTimeout(ctx, timeOut)
		defer cancel()

		rescChan := RunTaskGetSpecifyQuantityResultContext(
			timeOutctx,
			ps,
			fn,
			expectResultCount,
			poolCount,
		)

		for {
			select {
			case <-timeOutctx.Done():
				return
			case res, ok := <-rescChan:
				if ok {
					select {
					case resChan <- res:
					default:
					}
				} else {
					return
				}

			}
		}
	}()

	return resChan
}
