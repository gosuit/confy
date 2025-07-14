package confy

import (
	"os"
	"path"
)

type Reader interface {
	WithRootPath(path string) Reader
	WithEnvVariableName(name string) Reader
	Read(cfg any) error
}

type reader struct {
	rootPath   string
	envVarName string
}

func NewReader() Reader {
	return &reader{
		rootPath:   "./config",
		envVarName: "ENVIRONMENT",
	}
}

func (r *reader) WithRootPath(path string) Reader {
	return &reader{
		rootPath:   path,
		envVarName: r.envVarName,
	}
}

func (r *reader) WithEnvVariableName(name string) Reader {
	return &reader{
		rootPath:   r.rootPath,
		envVarName: name,
	}
}

func (r *reader) Read(cfg any) error {
	dir, ok := os.LookupEnv(r.envVarName)
	if !ok {
		dir = "local"
	}

	path := path.Join(r.rootPath, dir)

	return Get(path, cfg)
}
