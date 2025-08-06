package main

import (
	"fmt"

	"github.com/gosuit/confy"
)

type Config struct {
	Value string `confy:"value"`
}

// To use confy.Reader, you need to create an object implementing
// the interface using confy.NewReader and call the Reader.Read function
//
// Reader.Read supports all confy features
//
// By default, Reader will read the file "./config/local.{json|yaml|yml|toml}"
// or the "./config/local" directory.
//
// Other examples describe why this happens and how to change this behavior.
func main() {
	var cfg Config

	err := confy.NewReader().Read(&cfg)
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg)
}
