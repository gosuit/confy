package confy

import (
	"os"

	"github.com/go-playground/validator/v10"
)

// Get reads config from file or directory and override values with environment variables.
func Get(path string, cfg any) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		paths, err := getFilesFromDir(path)
		if err != nil {
			return err
		}

		return GetMany(cfg, paths...)
	} else {
		err := parseFile(path, cfg)
		if err != nil {
			return err
		}

		validate := validator.New()

		return validate.Struct(cfg)
	}
}

// GetMany reads config from multiple files and override values with environment variables.
func GetMany(cfg any, files ...string) error {
	err := parseMultiple(files, cfg)
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
