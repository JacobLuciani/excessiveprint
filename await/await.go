package await

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func Await(ready func() bool, timeout time.Duration) error {
	return AwaitWithContext(context.Background(), ready, timeout)
}

func AwaitWithContext(baseCtx context.Context, ready func() bool, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(baseCtx, timeout)
	defer cancel()
	readyChan := make(chan struct{})

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
			fmt.Println("still going")
		}
	}()

	select {
	case <-readyChan:
		return nil
	case <-ctx.Done():
		return errors.New("timeout expired")
	}
}

// TODO: make a generic for this
// is this really necessary? just cuts down on some boilerplate I suppose
func AwaitChan(inChan <-chan struct{}, timeout time.Duration) error {
	select {
	case <-inChan:
		return nil
	case <-time.After(timeout):
		return errors.New("timeout expired")
	}
}
