package main

import (
	"fmt"

	"github.com/gosuit/confy"
)

type AppConfig struct {
	Database DatabaseConfig
	Logger   LoggerConfig
}

type DatabaseConfig struct {
	Host     string
	Password string
}

type LoggerConfig struct {
	Level string
	Type  string
}

// To read from multiple files, use confy.ReadMany
func main() {
	var cfg AppConfig

	err := confy.ReadMany(&cfg, "db.yaml", "logger.yaml")
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Database)
	fmt.Println(cfg.Logger)
}
