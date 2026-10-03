package confy

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func loadFileData(path string) (map[string]any, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("error while read '%s': %w", path, err)
	}

	if fi.IsDir() {
		return nil, errors.New("config file must not be directory")
	}

	return parseFile(path)
}

func parseFile(path string) (map[string]any, error) {
	switch ext := strings.ToLower(filepath.Ext(path)); ext {
	case ".yml", ".yaml":
		return parseYAML(path)
	default:
		return nil, fmt.Errorf("confy does not support '%s' files", ext)
	}
}

func parseYAML(path string) (map[string]any, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("error while read '%s': %w", path, err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("error while read bytes from '%s': %w", path, err)
	}

	var data map[string]any

	if err := yaml.Unmarshal(b, &data); err != nil {
		return nil, fmt.Errorf("error while parse '%s': %w", path, err)
	}

	return data, nil
}
