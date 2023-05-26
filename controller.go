package controller

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/JacobLuciani/excessiveprint/await"
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
	return await.Await(func() bool {
		return c.ready
	}, timeout)
}
