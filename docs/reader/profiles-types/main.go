package main

import (
	"fmt"
	"os"

	"github.com/gosuit/confy"
)

type Config struct {
	Value string `confy:"value"`
}

// You can specify the data source using either a file or a directory.
//
// The file name should look like this: profile name + extension (.yaml, .yml, .json, .toml).
// There can be only one file for each profile (if there are more,
// for example "local.yaml" and "local.json", then a panic will be thrown)
//
// The directory name must be equal to the profile name.
// Any folder structure can be set in the directory.
// The file names can be any, but must have extensions (.yaml, .yml, .json, .toml).
// Files with other extensions will be ignored.
//
// A directory and a file cannot be used for the same profile at the same time.
// Otherwise, a panic will be thrown.
func main() {
	var localCfg Config

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
