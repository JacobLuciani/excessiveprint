package main

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

type basic struct {
	Field  string `mapstructure:"field"`
	Target string `mapstructure:"target"`
}

func main() {
	// fmt.Println("Hello world")

	i := map[string]interface{}{
		"Field":  "hello",
		"Target": "world",
	}

	b := basic{}

	mapstructure.Decode(i, &b)

	fmt.Println("My basic struct: ", b)

}
