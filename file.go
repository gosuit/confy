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

var (
	validExt = []string{".yaml", ".yml", ".json", ".toml"}
)

func parseFile(path string, cfg any) error {
	// open the configuration file
	f, err := os.OpenFile(path, os.O_RDONLY|os.O_SYNC, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	var data map[string]any
	var fileType string

	// parse the file depending on the file type
	switch ext := strings.ToLower(filepath.Ext(path)); ext {
	case ".yaml", ".yml":
		fileType = "yaml"
		err = parseYAML(f, &data)
	case ".json":
		fileType = "json"
		err = parseJSON(f, &data)
	case ".toml":
		fileType = "toml"
		err = parseTOML(f, &data)
	default:
		return fmt.Errorf("file format '%s' doesn't supported by the parser", ext)
	}
	if err != nil {
		return fmt.Errorf("config file parsing error: %s", err.Error())
	}

	out := reflect.ValueOf(cfg)

	if out.Kind() == reflect.Ptr && !out.IsNil() {
		out = out.Elem()
	} else {
		return errors.New("config struct must be pointer and not nil")
	}

	if out.Kind() != reflect.Struct {
		return errors.New("config must be struct")
	}

	return mergeStruct(data, out, fileType)
}

func parseMultiple(paths []string, cfg any) error {
	data := make(map[string]any)

	for _, path := range paths {
		// open the configuration file
		f, err := os.OpenFile(path, os.O_RDONLY|os.O_SYNC, 0)
		if err != nil {
			return err
		}
		defer f.Close()

		var newData map[string]any

		// parse the file depending on the file type
		switch ext := strings.ToLower(filepath.Ext(path)); ext {
		case ".yaml", ".yml":
			err = parseYAML(f, &newData)
		case ".json":
			err = parseJSON(f, &newData)
		case ".toml":
			err = parseTOML(f, &newData)
		default:
			return fmt.Errorf("file format '%s' doesn't supported by the parser", ext)
		}
		if err != nil {
			return fmt.Errorf("config file parsing error: %s", err.Error())
		}

		data = mergeMaps(data, newData)
	}

	out := reflect.ValueOf(cfg)

	if out.Kind() == reflect.Ptr && !out.IsNil() {
		out = out.Elem()
	} else {
		return errors.New("config struct must be pointer and not nil")
	}

	if out.Kind() != reflect.Struct {
		return errors.New("config must be struct")
	}

	return mergeStruct(data, out, "yaml")
}

func mergeMaps(dst, src map[string]any) map[string]any {
	for key, val := range src {
		if dstVal, present := dst[key]; present {
			dst[key] = mergeMaps(dstVal.(map[string]any), val.(map[string]any))
		} else {
			dst[key] = val
		}
	}
	return dst
}

// ParseYAML parses YAML from reader to data structure
func parseYAML(r io.Reader, data *map[string]any) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, data)
}

// ParseTOML parses TOML from reader to data structure
func parseTOML(r io.Reader, data *map[string]any) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	return toml.Unmarshal(b, data)
}

// ParseJSON parses JSON from reader to data structure
func parseJSON(r io.Reader, data *map[string]any) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, data)
}
