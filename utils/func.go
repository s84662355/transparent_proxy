package utils

func DoOrCancel(c <-chan struct{}, do func()) (cancel func()) {
	done := make(chan struct{})

	cancel = func() {
		for range done {
		}
	}

	go func() {
		defer close(done)
		select {
		case <-c:
			do()
		case done <- struct{}{}:
		}
	}()

	return
}
