package controller

import (
	"context"
	"errors"
	"log"
	"time"
)

type controller struct {
	ctx    context.Context
	logger *log.Logger

	ready bool

	processChan chan interface{}
}

func NewController(ctx context.Context, logger *log.Logger) *controller {
	return &controller{
		ctx:         ctx,
		ready:       false,
		logger:      logger,
		processChan: make(chan interface{}),
	}
}

func (c *controller) Run() {
	c.ready = true

	for c.ready {
		select {
		case val := <-c.processChan:
			c.print(val)
		case <-c.ctx.Done():
			c.ready = false
		}
	}
}

func (c *controller) print(val interface{}) {
	c.logger.Println(val)
}

func (c *controller) Process(val interface{}) error {
	if !c.ready {
		return errors.New("controller not available")
	}

	c.processChan <- val
	return nil
}

func (c *controller) Await(timeout time.Duration) error {
	readyChan := make(chan struct{})

	go func() {
		for {
			if c.ready {
				c.logger.Println("controller now ready")
				readyChan <- struct{}{}
				return
			}
			c.logger.Println("controller not ready")
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
