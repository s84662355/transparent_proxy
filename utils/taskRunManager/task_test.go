package taskRunManager

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

// go test -bench=BenchmarkAtomictask -count=1 -run=none
func BenchmarkAtomictask(t *testing.B) {
	var ff atomic.Uint32
	tttt := NewAtomicTask(100)
	startTime := time.Now()
	for i := 0; i < 10000; i++ {
		tttt.Run(func() {
			ff.Add(1)
			// t.Log(ff.Load())
		})
	}
	tttt.Wait()
	t.Log(time.Since(startTime), ff.Load())
}

// go test -bench=BenchmarkChantask -run=none
func BenchmarkChantask(t *testing.B) {
	var ff atomic.Uint32
	tttt := NewChanTask(100)
	startTime := time.Now()
	for i := 0; i < 5000; i++ {
		tttt.Run(func() {
			ff.Add(1)
		})
	}
	tttt.Wait()
	t.Log(time.Since(startTime), ff.Load())
}

// go test -v -run TestSleepTaskManagert  -tags "dev"
func TestSleepTaskManagert(t *testing.T) {
	taskManager := NewAtomicTask(1)

	for i := 0; i < 500; i++ {
		taskManager.Run(func() {})
	}

	taskManager.Wait()
}

// go test -v   -run TestRunTaskGetSpecifyQuantityResultContext   -tags "dev"
func TestRunTaskGetFirstResult(t *testing.T) {
	urls := []string{
		"https://www.google.com/",
		"https://www.baidu.com/",
		"http://gvisor.dev/",
	}

	resChan := RunTaskGetSpecifyQuantityResultContext[string, string](
		context.Background(),
		urls,
		func(ctx context.Context, url string) (string, bool) {
			resp, err := http.Get(url)
			if err != nil {
				return "", false
			}

			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", false
			}

			return string(body), true
		}, 2, 3)

	for m := range resChan {
		fmt.Sprint(m)
		// t.Log(m)
	}
}

// go test -v -run TestRunTaskGetSpecifyQuantityResultTimeOutContext  -tags "dev"
func TestRunTaskGetSpecifyQuantityResultTimeOutContext(t *testing.T) {
	urls := []string{
		"https://www.cnblogs.com/",
		"https://www.google.com/",

		"https://www.cnblogs.com/",
		"http://gvisor.dev/",
	}

	resChan := RunTaskGetSpecifyQuantityResultTimeOutContext[string, string](
		context.Background(),
		20*time.Second,
		urls,
		func(ctx context.Context, url string) (string, bool) {
			resp, err := http.Get(url)
			if err != nil {
				return "", false
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", false
			}

			return string(body), true
		}, 4, 4)

	for m := range resChan {
		fmt.Sprint(m)
		t.Log(1)
	}
}
