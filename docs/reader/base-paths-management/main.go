package main

import (
	"fmt"
	"os"

	"github.com/gosuit/confy"
)

type Config struct {
	Value string `confy:"value"`
}

// By default, Reader searches for data sources in the "./config" directory.
// But you can change this behavior by calling the Reader.SetRootPath function.
//
// By default, Reader reads the ENVIRONMENT variable to find out the name of the current profile.
// However, you can change the name of the variable that will contain the profile name
// by calling Reader.SetEnvVariableName.
func main() {
	// Set profile name
	os.Setenv("ENV", "dev")

	var cfg Config

	err := confy.NewReader().
		SetRootPath("./configuration").
		SetEnvVariableName("ENV").
		Read(&cfg)

	if err != nil {
		panic(err)
	}

	fmt.Println(cfg)
}
