package main

import (
	"fmt"

	"github.com/gosuit/confy"
)

type DbConfig struct {
	Host string
}

type LogConfig struct {
	Level string
}

type Config struct {
	Log LogConfig
	Db  DbConfig
}

// This example only applies to reading a profile from a directory.
//
// By default, Reader reads the entire contents of the directory.
//
// You can change this behavior by calling Reader.SetReadAll
// and manually specifying the sources to read using Reader.AddSource
func main() {
	var cfg Config

	err := confy.NewReader().
		SetReadAll(false).
		AddSource("log.yaml").
		Read(&cfg)

	if err != nil {
		panic(err)
	}

	fmt.Println(cfg)
	fmt.Println(cfg.Db.Host == "") // true
}
