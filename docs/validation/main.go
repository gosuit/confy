package main

import (
	"os"

	"github.com/gosuit/confy"
)

// In confy, validation is implemented using https://github.com/go-playground/validator
type ApiConfig struct {
	Url string `confy:"url"   validate:"url"`
	Key string `env:"API_KEY" validate:"min=10,max=100"`
}

// All functions in confy validate configs.
func main() {
	// Set API_KEY variable
	os.Setenv("API_KEY", "key")

	var cfg ApiConfig

	// It read and validate config.
	//
	// The Url field is valid, but the Key field does not meet the conditions,
	// so an error will be returned.
	err := confy.Read(&cfg, "config.yaml")
	if err != nil {
		panic(err)
	}
}
