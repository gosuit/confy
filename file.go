package confy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

func parseFile(path string, cfg any) error {
	// open the configuration file
	f, err := os.OpenFile(path, os.O_RDONLY|os.O_SYNC, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	// parse the file depending on the file type
	switch ext := strings.ToLower(filepath.Ext(path)); ext {
	case ".yaml", ".yml":
		err = parseYAML(f, cfg)
	case ".json":
		err = parseJSON(f, cfg)
	case ".toml":
		err = parseTOML(f, cfg)
	default:
		return fmt.Errorf("file format '%s' doesn't supported by the parser", ext)
	}
	if err != nil {
		return fmt.Errorf("config file parsing error: %s", err.Error())
	}

	return nil
}

// ParseYAML parses YAML from reader to data structure
func parseYAML(r io.Reader, str any) error {
	var data map[string]any

	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	out := reflect.ValueOf(str)

	if out.Kind() == reflect.Ptr && !out.IsNil() {
		out = out.Elem()
	} else {
		return errors.New("config struct must be pointer and not nil")
	}

	if out.Kind() != reflect.Struct {
		return errors.New("config must be struct")
	}

	if err := mergeStruct(data, out, "yaml"); err != nil {
		return err
	}

	return nil
}

// ParseTOML parses TOML from reader to data structure
func parseTOML(r io.Reader, str any) error {
	var data map[string]any

	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	err = toml.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	out := reflect.ValueOf(str)

	if out.Kind() == reflect.Ptr && !out.IsNil() {
		out = out.Elem()
	} else {
		return errors.New("config struct must be pointer and not nil")
	}

	if out.Kind() != reflect.Struct {
		return errors.New("config must be struct")
	}

	if err := mergeStruct(data, out, "toml"); err != nil {
		return err
	}

	return nil
}

// ParseJSON parses JSON from reader to data structure
func parseJSON(r io.Reader, str any) error {
	var data map[string]any

	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	out := reflect.ValueOf(str)

	if out.Kind() == reflect.Ptr && !out.IsNil() {
		out = out.Elem()
	} else {
		return errors.New("config struct must be pointer and not nil")
	}

	if out.Kind() != reflect.Struct {
		return errors.New("config must be struct")
	}

	if err := mergeStruct(data, out, "json"); err != nil {
		return err
	}

	return nil
}
