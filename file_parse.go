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
	"olympos.io/encoding/edn"
)

func parseFile(path string, cfg interface{}) error {
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
	case ".edn":
		err = parseEDN(f, cfg)
	case ".env":
		err = parseENV(f, cfg)
	default:
		return fmt.Errorf("file format '%s' doesn't supported by the parser", ext)
	}
	if err != nil {
		return fmt.Errorf("config file parsing error: %s", err.Error())
	}
	return nil
}

// ParseYAML parses YAML from reader to data structure
func parseYAML(r io.Reader, str interface{}) error {
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

	if err := mergeStruct(data, out); err != nil {
		return err
	}

	return nil
}

// ParseTOML parses TOML from reader to data structure
func parseTOML(r io.Reader, str interface{}) error {
	_, err := toml.NewDecoder(r).Decode(str)
	return err
}

// ParseJSON parses JSON from reader to data structure
func parseJSON(r io.Reader, str interface{}) error {
	return json.NewDecoder(r).Decode(str)
}

// parseEDN parses EDN from reader to data structure
func parseEDN(r io.Reader, str interface{}) error {
	return edn.NewDecoder(r).Decode(str)
}
