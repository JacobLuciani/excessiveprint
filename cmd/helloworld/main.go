package main

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	controller "github.com/JacobLuciani/excessiveprint"
	"github.com/mitchellh/mapstructure"
)

func main() {

	i := map[string]interface{}{
		"Field":  "hello",
		"Target": "world",
	}

	b := struct {
		Field  string `mapstructure:"field"`
		Target string `mapstructure:"target"`
	}{}

	mapstructure.Decode(i, &b)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logger := log.New(os.Stdout, "localtest: ", log.LstdFlags)

	cont := controller.NewController(ctx, logger)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		cont.Run()
		defer wg.Done()
	}()

	if err := cont.Await(5 * time.Second); err != nil {
		cancel()
		logger.Fatal(err)
	}

	if err := cont.Process(b); err != nil {
		logger.Fatal(err)
	}

	wg.Wait()
}
