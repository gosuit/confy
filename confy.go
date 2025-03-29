package confy

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
)

// Get reads config from file and override values with environment variables.
func Get(path string, cfg any) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		paths := make([]string, 0)

		err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
			if !info.IsDir() {
				paths = append(paths, path)
			}

			return nil
		})
		if err != nil {
			return err
		}

		return GetMany(cfg, paths...)
	} else {
		err := parseFile(path, cfg, true)
		if err != nil {
			return err
		}

		validate := validator.New()

		err = validate.Struct(cfg)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetMany reads config from multiple files and override values with environment variables.
func GetMany(cfg any, files ...string) error {
	for i, path := range files {
		last := false

		if i == len(files)-1 {
			last = true
		}

		err := parseFile(path, cfg, last)
		if err != nil {
			return err
		}

		validate := validator.New()

		err = validate.Struct(cfg)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetEnv reads environment variables into the structure.
func GetEnv(cfg any) error {
	err := readEnvVars(cfg, false)
	if err != nil {
		return err
	}

	validate := validator.New()

	err = validate.Struct(cfg)
	if err != nil {
		return err
	}

	return nil
}
