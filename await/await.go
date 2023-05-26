package await

import (
	"context"
	"errors"
	"time"
)

func Await(ready func() bool, timeout time.Duration) error {
	readyChan := make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	go func() {
		for {
			//Nonblocking read allows cleanup after timeout
			select {
			case <-ctx.Done():
				return
			default:
			}
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
	case <-ctx.Done():
		return errors.New("timeout expired")
	}
}
