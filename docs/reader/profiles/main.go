package main

import (
	"fmt"
	"os"

	"github.com/gosuit/confy"
)

type Config struct {
	Value string `confy:"value"`
}

// As mentioned in the simple example, by default,
// Reader will read the file "./config/local.{json|yaml|yml|toml}"
// or the "./config/local" directory.
//
// This is due to the fact that the Reader reads the value of the ENVIRONMENT variable
// (if it is not set, "local" is set by default).
// This value is the name of the current profile.
//
// Let ENVIRONMENT="prod". Then the following sources will belong to this profile:
//
//	    "./config/prod.{json|yaml|yml|toml}"
//		     or
//	    "./config/prod"
func main() {
	var localCfg Config

	// Default profile name is "local"
	err := confy.NewReader().Read(&localCfg)
	if err != nil {
		panic(err)
	}

	fmt.Println(localCfg)

	// Change profile name
	os.Setenv("ENVIRONMENT", "prod")

	var prodCfg Config

	err = confy.NewReader().Read(&prodCfg)
	if err != nil {
		panic(err)
	}

	fmt.Println(prodCfg)
}
