package main

import (
	"fmt"

	"github.com/gosuit/confy"
)

// Using struct tag, you must specify a field from the file that contains the value.
// However, if you do not specify the field name using the struct tag,
// then the field name in the struct will be used.
//
// The "confy" tag does not depend on the file type, but if you want,
// you can also use "yaml", "json" or "toml" tags for the corresponding files.
type Config struct {
	// The following simple types are available for reading:
	//
	//    1) string
	//    2) bool
	//    3) all integer types
	//    4) all unsigned integer types
	//    5) float types
	//    6) pointer
	//
	Host string `confy:"host"`

	// The following composite types are available for reading:
	//
	//     1) struct
	// 	   2) map
	//     3) array and slice
	//     4) interface
	//
	Logger Logger `confy:"log"`
}

type Logger struct {
	Level string `confy:"level"`
}

// confy.Read reads the specified file and writes the value to the passed struct.
//
// yamlCfg, ymlCfg, jsonCfg, tomlCfg will contain the same values.
func main() {
	// Read YAML file
	var yamlCfg Config

	err := confy.Read(&yamlCfg, "config.yaml")
	if err != nil {
		panic(err)
	}

	// Read YML file
	var ymlCfg Config

	err = confy.Read(&ymlCfg, "config.yml")
	if err != nil {
		panic(err)
	}

	// Read JSON file
	var jsonCfg Config

	err = confy.Read(&jsonCfg, "config.json")
	if err != nil {
		panic(err)
	}

	// Read TOML file
	var tomlCfg Config

	err = confy.Read(&tomlCfg, "config.toml")
	if err != nil {
		panic(err)
	}

	fmt.Println(yamlCfg)
	fmt.Println(ymlCfg)
	fmt.Println(jsonCfg)
	fmt.Println(tomlCfg)
}
