package confy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

func parseFile(path string) (map[string]any, string, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, "", err
	}

	var data map[string]any
	var fileTag string

	switch ext := strings.ToLower(filepath.Ext(path)); ext {
	case ".yaml", ".yml":
		err = parseYAML(b, &data)
		fileTag = yamlTag
	case ".json":
		err = parseJSON(b, &data)
		fileTag = jsonTag
	case ".toml":
		err = parseTOML(b, &data)
		fileTag = tomlTag
	default:
		return nil, "", fmt.Errorf("file format '%s' doesn't supported by confy", ext)
	}

	if err != nil {
		return nil, "", err
	}

	return data, fileTag, nil
}

func parseMultipleFiles(paths []string) (map[string]any, error) {
	data := make(map[string]any)

	for _, path := range paths {
		newData, _, err := parseFile(path)
		if err != nil {
			return nil, err
		}

		data, err = mergeMaps(data, newData)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

func mergeMaps(dst, src map[string]any) (map[string]any, error) {
	for key, val := range src {
		if dstVal, ok := dst[key]; ok {
			if dstValMap, ok := dstVal.(map[string]any); ok {
				if valMap, ok := val.(map[string]any); ok {
					newVal, err := mergeMaps(dstValMap, valMap)
					if err != nil {
						return nil, err
					}

					dst[key] = newVal
				} else {
					return nil, errors.New("value conflict in different sources")
				}
			} else {
				return nil, errors.New("value conflict in different sources")
			}
		} else {
			dst[key] = val
		}
	}

	return dst, nil
}

func parseYAML(b []byte, data *map[string]any) error {
	return yaml.Unmarshal(b, data)
}

func parseJSON(b []byte, data *map[string]any) error {
	return json.Unmarshal(b, data)
}

func parseTOML(b []byte, data *map[string]any) error {
	return toml.Unmarshal(b, data)
}
