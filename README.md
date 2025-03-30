# Confy

Confy is Go library for reading configuration settings from environment variables and YAML or other files. It is based on <a href="https://github.com/ilyakaznacheev/cleanenv">CleanEnv</a>. It provides a simple way to manage application configurations while ensuring that the configurations are valid.

## Installation

```zsh
go get github.com/gosuit/confy
```

## Features
 
- **Files Support**: Load configuration settings from files. Types are supported: 
  - **YAML**
  - **JSON**
  - **TOML**
- **Environment Variables**: Override configuration settings with environment variables.
- **Env Names Expand**: Set the names of environment variables through files to get the values
- **Multiple files**: Load configuration settings from multiple files.
- **Validation**: Validate configuration structures.

## Usage

- [Simple example](#simple-read)
- [Setup env names](#setup-env-names)
- [Environment only](#environment-only)
- [Multiple files](#multiple-files)
- [Config validation](#config-validation)

### Simple example

This is an example of how to read configs. Confy will read the data from the config.yaml and environment variables.

The "confy" tag does not depend on the file type, but if you want, you can also use "yaml", "json" or "toml" tags for the corresponding files.

```yaml
# config.yaml

database:
  host: "localhost"

api:
  url: "http://api"
```

```golang
// main.go

package main

import "github.com/gosuit/confy"

type DatabaseConfig struct {
	Host     string `confy:"host"`
	Password string `env:"DB_PASSWORD" env-default:"root"`
}

type ApiConfig struct {
	Url string `confy:"url"`
	Key string `env:"API_KEY"`
}

type AppConfig struct {
	Db  DatabaseConfig `confy:"database"`
	Api ApiConfig      `confy:"api"`
}

func main() {
	var cfg AppConfig

	err := confy.Get("config.yaml", &cfg)
	if err != nil {
		panic(err)
	}

	// Use config...
}
```

### Setup env names

In addition to the "env" and "env-default" tags, you can specify the name of the environment variable from which to take the value and the default value via a file.

The code below will use the DB_HOST and DB_PASSWORD environment variables and their corresponding default values of "localhost" and "root"

**IMPORTANT**: The names of variables installed in the file have priority over the "env" and "env-default" tags

```yaml
# config.yaml

host:      ${DB_HOST:localhost}
password:  ${DB_PASSWORD:root}
```

```golang
// main.go

package main

import "github.com/gosuit/confy"

type DbConfig struct {
	Host     string `confy:"host"`
	Password string `confy:"password"`
}

func main() {
	var cfg DbConfig

	err := confy.Get("config.yaml", &cfg)
	if err != nil {
		panic(err)
	}

	// Use config...
}
```

### Environment only

You can also use only environment variables.

```golang
// main.go

package main

import "github.com/gosuit/confy"

type DatabaseConfig struct {
	Host     string `env:"DB_HOST"`
	Password string `env:"DB_PASSWORD"`
}

func main() {
	var cfg DatabaseConfig

	err := confy.GetEnv(&cfg)
	if err != nil {
		panic(err)
	}

	// Use config...
}
```

### Multiple files

You can also read config from multiple files.

**IMPORTANT**: For multiple read you can use only "confy" or "yaml" tag.

```yaml
# database.yaml

database:
  host: "localhost"
```

```yaml
# api.yml

api:
  url: "http://api"
```

```golang
// main.go

package main

import "github.com/gosuit/confy"

type DatabaseConfig struct {
	Host     string `confy:"host"`
	Password string `env:"DB_PASSWORD" env-default:"root"`
}

type ApiConfig struct {
	Url string `confy:"url"`
	Key string `env:"API_KEY"`
}

type AppConfig struct {
	Db  DatabaseConfig `confy:"database"`
	Api ApiConfig      `confy:"api"`
}

func main() {
	var cfg AppConfig

	err := confy.GetMany(&cfg, "database.yaml", "api.yml")
	if err != nil {
		panic(err)
	}

	// Use config...
}
```

### Config validation

In addition, you can validate configs. 

In Confy, validation is configured based on <a href="https://github.com/go-playground/validator">validator</a>

```golang
// main.go

package main

import "github.com/gosuit/confy"

type ApiConfig struct {
	Url string `confy:"url"   validate:"url"`
	Key string `env:"API_KEY" validate:"min=10,max=100"`
}

func main() {
	var cfg ApiConfig

	// It read and validate configs
	err := confy.Get("config.yaml", &cfg)
	if err != nil {
		panic(err)
	}

	// Use config...
}
```

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue for any enhancements or bug fixes.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
