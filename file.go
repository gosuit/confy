package confy

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

var (
	validExtensions = []string{".yaml", ".yml", ".json", ".toml", ".env"}
)

func getFileData(path string) (map[string]any, string, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, "", fmt.Errorf("error while '%s' path read: %s", path, err.Error())
	}

	if fi.IsDir() {
		paths, err := getValidFiles(path)
		if err != nil {
			return nil, "", err
		}

		fileData, err := parseMultipleFiles(paths)
		if err != nil {
			return nil, "", err
		}

		return fileData, getMultipleFilesTag(paths), nil
	} else {
		fileData, err := parseFile(path)
		if err != nil {
			return nil, "", err
		}

		return fileData, getFileTag(path), nil
	}
}

func getMultipleFilesData(paths []string) (map[string]any, string, error) {
	files := make([]string, 0)

	for _, path := range paths {
		fi, err := os.Stat(path)
		if err != nil {
			return nil, "", fmt.Errorf("error while '%s' path read: %s", path, err.Error())
		}

		if fi.IsDir() {
			newFiles, err := getValidFiles(path)
			if err != nil {
				return nil, "", err
			}

			files = append(files, newFiles...)
		} else {
			files = append(files, path)
		}
	}

	fileData, err := parseMultipleFiles(files)
	if err != nil {
		return nil, "", err
	}

	return fileData, getMultipleFilesTag(files), nil
}

func parseFile(path string) (map[string]any, error) {
	var data map[string]any
	var err error

	switch ext := strings.ToLower(filepath.Ext(path)); ext {
	case ".yaml", ".yml":
		err = parseYAML(path, &data)
	case ".json":
		err = parseJSON(path, &data)
	case ".toml":
		err = parseTOML(path, &data)
	case ".env":
		err = parseENV(path)
	default:
		return nil, fmt.Errorf("confy doesn`t support '%s' files", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("error while '%s' file parsing: %s", path, err.Error())
	}

	return data, nil
}

func parseMultipleFiles(paths []string) (map[string]any, error) {
	data := make(map[string]any)

	var previous string

	if len(paths) > 0 {
		previous = paths[0]
	}

	for _, path := range paths {
		newData, err := parseFile(path)
		if err != nil {
			return nil, err
		}

		data, err = mergeMaps(data, newData, previous, path, "")
		if err != nil {
			return nil, err
		}

		previous = path
	}

	return data, nil
}

func parseYAML(path string, to *map[string]any) error {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, to)
}

func parseJSON(path string, to *map[string]any) error {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, to)
}

func parseTOML(path string, to *map[string]any) error {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	return toml.Unmarshal(b, to)
}

func parseENV(path string) error {
	return godotenv.Load(path)
}

func mergeMaps(dst, src map[string]any, dstPath, srcPath, commonKey string) (map[string]any, error) {
	for key, val := range src {
		if dstVal, ok := dst[key]; ok {
			if commonKey == "" {
				commonKey = key
			} else {
				commonKey += "." + key
			}

			if dstValMap, ok := dstVal.(map[string]any); ok {
				if valMap, ok := val.(map[string]any); ok {
					newVal, err := mergeMaps(dstValMap, valMap, dstPath, srcPath, commonKey)
					if err != nil {
						return nil, err
					}

					dst[key] = newVal
				} else {
					return nil, fmt.Errorf("conflict while files read: it is impossible to unambiguously determine the value for the '%s' key specified in the '%s' file and in the '%s' file", commonKey, dstPath, srcPath)
				}
			} else {
				return nil, fmt.Errorf("conflict while files read: it is impossible to unambiguously determine the value for the '%s' key specified in the '%s' file and in the '%s' file", commonKey, dstPath, srcPath)
			}
		} else {
			dst[key] = val
		}
	}

	return dst, nil
}

func getValidFiles(path string) ([]string, error) {
	paths := make([]string, 0)

	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))

			if slices.Contains(validExtensions, ext) {
				paths = append(paths, path)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error while '%s' directory read: %s", path, err.Error())
	}

	return paths, nil
}

func getAllPaths(path string) ([]string, error) {
	paths := make([]string, 0)

	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		paths = append(paths, path)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error while '%s' directory read: %s", path, err.Error())
	}

	return paths, nil
}

func getFileTag(path string) string {
	switch ext := strings.ToLower(filepath.Ext(path)); ext {
	case ".yaml", ".yml":
		return yamlTag
	case ".json":
		return jsonTag
	case ".toml":
		return tomlTag
	default:
		return confyTag
	}
}

func getMultipleFilesTag(paths []string) string {
	exts := make([]string, 0)

	for _, path := range paths {
		ext := strings.ToLower(filepath.Ext(path))

		if !slices.Contains(exts, ext) {
			exts = append(exts, ext)
		}
	}

	if len(exts) > 1 {
		return confyTag
	}

	return getFileTag(exts[0])
}
