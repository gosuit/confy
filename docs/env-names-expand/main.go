package main

import (
	"fmt"
	"os"

	"github.com/gosuit/confy"
)

type DbConfig struct {
	Host     string `confy:"host"`
	Password string `confy:"password"`
}

// All functions in confy can expand the names of environment variables
// from which to take values.
//
// IMPORTANT: The names of variables installed in the file have priority over
// the "env" and "env-default" tags
func main() {
	// Set variables
	os.Setenv("DB_PASSWORD", "root")

	var cfg DbConfig

	err := confy.Read(&cfg, "config.yaml")
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg)
}
