package confy

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

const (
	defaultRootPath   = "config"
	defaultEnvVarName = "ENVIRONMENT"
)

type Reader interface {
	SetRootPath(path string) Reader
	SetEnvVariableName(name string) Reader
	SetReadAll(readAll bool) Reader
	AddSource(source string) Reader
	Read(to any) error
}

type reader struct {
	rootPath   string
	envVarName string
	readAll    bool
	sources    []string
}

func NewReader() Reader {
	return &reader{
		rootPath:   defaultRootPath,
		envVarName: defaultEnvVarName,
		readAll:    true,
		sources:    make([]string, 0),
	}
}

func (r *reader) SetRootPath(path string) Reader {
	r.rootPath = path

	return r
}

func (r *reader) SetEnvVariableName(name string) Reader {
	r.envVarName = name

	return r
}

func (r *reader) SetReadAll(readAll bool) Reader {
	r.readAll = readAll

	return r
}

func (r *reader) AddSource(source string) Reader {
	if r.readAll {
		panic("you can`t add source for reader when ReadAll = true")
	}

	r.sources = append(r.sources, source)

	return r
}

func (r *reader) Read(to any) error {
	env, ok := os.LookupEnv(r.envVarName)
	if !ok {
		env = "local"
	}

	paths, err := getAllPaths(r.rootPath)
	if err != nil {
		return err
	}

	dirSource := filepath.Join(r.rootPath, env)
	yamlSource := filepath.Join(r.rootPath, env+".yaml")
	ymlSource := filepath.Join(r.rootPath, env+".yml")
	jsonSource := filepath.Join(r.rootPath, env+".json")
	tomlSource := filepath.Join(r.rootPath, env+".toml")

	dirSourceExists := slices.Contains(paths, dirSource)
	fileSourceExists := slices.Contains(paths, yamlSource) ||
		slices.Contains(paths, ymlSource) ||
		slices.Contains(paths, jsonSource) ||
		slices.Contains(paths, tomlSource)

	if dirSourceExists && fileSourceExists {
		panic("confy: you can't use directory source and file source at the same time")
	} else if !dirSourceExists && !fileSourceExists {
		panic("confy: not a single source was found")
	} else if fileSourceExists {
		filePath := ""
		fileFound := false

		if slices.Contains(paths, yamlSource) {
			fileFound = true
			filePath = yamlSource
		}

		if slices.Contains(paths, ymlSource) {
			if fileFound {
				panic("confy: there can only be one file source")
			}

			fileFound = true
			filePath = ymlSource
		}

		if slices.Contains(paths, jsonSource) {
			if fileFound {
				panic("confy: there can only be one file source")
			}

			fileFound = true
			filePath = jsonSource
		}

		if slices.Contains(paths, tomlSource) {
			if fileFound {
				panic("confy: there can only be one file source")
			}

			filePath = tomlSource
		}

		return Read(to, filePath)
	} else {
		if r.readAll {
			return Read(to, dirSource)
		} else {
			//TODO: add special sources for env support
			toRead := r.sources

			for i := range toRead {
				path := filepath.Join(dirSource, toRead[i])

				if !slices.Contains(paths, path) {
					panic(fmt.Sprintf("confy: source %s wasn`t found", path))
				}

				toRead[i] = path
			}

			return ReadMany(to, toRead...)
		}
	}
}
