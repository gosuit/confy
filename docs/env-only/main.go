package main

import (
	"fmt"
	"os"

	"github.com/gosuit/confy"
)

type DatabaseConfig struct {
	Host     string `env:"DB_HOST"`
	Password string `env:"DB_PASSWORD"`
}

// If you want to use values only from environment variables,
// then you can use confy.ReadEnv.
func main() {
	// Set variables.
	os.Setenv("DB_HOST", "0.0.0.0")
	os.Setenv("DB_PASSWORD", "root")

	var cfg DatabaseConfig

	err := confy.ReadEnv(&cfg)
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg)
}
