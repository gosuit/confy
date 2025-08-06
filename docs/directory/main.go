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

// You can read the directory.
//
// In this case, all files in the directory will be read and the
// data from them will be written to the configuration struct.
//
// You can also create subfolders in the folder specified for reading,
// and confy will read these values as well.
func main() {
	var cfg AppConfig

	err := confy.Read(&cfg, "./config")
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Database)
	fmt.Println(cfg.Logger)
}
