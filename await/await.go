package await

import (
	"errors"
	"time"
)

func Await(ready func() bool, timeout time.Duration) error {
	readyChan := make(chan struct{})

	go func() {
		for {
			if ready() {
				readyChan <- struct{}{}
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()

	select {
	case <-readyChan:
		return nil
	case <-time.After(timeout):
		return errors.New("timeout waiting for controller expired")
	}
}
