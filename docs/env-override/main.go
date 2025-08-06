package main

import (
	"fmt"
	"os"

	"github.com/gosuit/confy"
)

type DatabaseConfig struct {
	Host string `confy:"host"`

	// You can override the value from the file with the value
	// from the environment variable, if it is defined.
	Username string `confy:"username" env:"DB_USERNAME"`

	// If the value is not specified in the file and environment variable,
	// you can use the default value.
	Password string `confy:"password" env:"DB_PASSWORD" env-default:"root"`
}

type ApiConfig struct {
	// If the value is specified in the file but not specified in the environment variable,
	// the default value will not be used.
	Url string `confy:"url" env:"API_URL" env-default:"http://example.com/v2"`

	// You can only use only the value from the environment variable.
	Key string `env:"API_KEY"`
}

type AppConfig struct {
	Db  DatabaseConfig `confy:"database"`
	Api ApiConfig      `confy:"api"`
}

func main() {
	// Set variables
	os.Setenv("API_KEY", "key")
	os.Setenv("DB_USERNAME", "admin")

	var cfg AppConfig

	err := confy.Read(&cfg, "config.yaml")
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg)
}
