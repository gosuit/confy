package confy

import (
	"os"

	"github.com/go-playground/validator/v10"
)

// Read reads config from file or directory and override values with environment variables.
func Read(to any, from string) error {
	fi, err := os.Stat(from)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		paths, err := getFilesFromDir(from)
		if err != nil {
			return err
		}

		return ReadMany(to, paths...)
	} else {
		err := parseFile(from, to)
		if err != nil {
			return err
		}

		validate := validator.New()

		return validate.Struct(to)
	}
}

// ReadMany reads config from multiple files and override values with environment variables.
func ReadMany(to any, from ...string) error {
	err := parseMultiple(from, to)
	if err != nil {
		return err
	}

	validate := validator.New()

	err = validate.Struct(to)
	if err != nil {
		return err
	}

	return nil
}

// ReadEnv reads environment variables into the structure.
func ReadEnv(to any) error {
	err := readEnvVars(to, false)
	if err != nil {
		return err
	}

	validate := validator.New()

	err = validate.Struct(to)
	if err != nil {
		return err
	}

	return nil
}
