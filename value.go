package confy

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

func getStructData(data map[string]any, metadata map[string]string) (map[string]any, error) {
	structData, ok := data[metadata["key"]]
	if !ok {
		return make(map[string]any), nil
	}

	if mapStructData, ok := structData.(map[string]any); ok {
		return mapStructData, nil
	} else {
		return nil, fmt.Errorf("error while value parsing: invalid value for '%s' struct field", metadata["name"])
	}
}

func getFieldValue(f reflect.Value, data map[string]any, metadata map[string]string) (reflect.Value, error) {
	value, fileOk, expanded := getFieldFileValue(data, metadata)

	value, envOk := overrideValueWithEnv(value, metadata)

	if envOk || expanded {
		metadata["isValueEnv"] = "true"
	} else {
		metadata["isValueEnv"] = "false"
	}

	if !(fileOk || envOk) {
		var defaultOk bool

		value, defaultOk = getFieldDefaultValue(metadata)
		if !defaultOk {
			if isValueRequired(metadata) {
				return reflect.Value{}, fmt.Errorf("error while value parsing: value for '%s' field is required", metadata["name"])
			} else {
				result := reflect.New(f.Type().Elem()).Elem()

				return result, nil
			}
		}
	}

	return parseValue(f, value, metadata)
}

func getFieldFileValue(data map[string]any, metadata map[string]string) (any, bool, bool) {
	var expanded bool

	value, ok := data[metadata["key"]]
	if ok {
		var envOk bool

		value, expanded, envOk = expandValue(value)

		if expanded {
			ok = envOk
		}
	}

	return value, ok, expanded
}

func expandValue(value any) (any, bool, bool) {
	expanded := false
	envOk := false

	if strVal, ok := value.(string); ok {
		if len(strVal) > 3 && strVal[0] == '$' && strVal[1] == '{' && strVal[len(strVal)-1] == '}' {
			parts := strings.Split(strVal[2:len(strVal)-1], ":")
			expanded = true

			value, ok = os.LookupEnv(parts[0])
			if !ok {
				if len(parts) > 1 {
					value = parts[1]
					envOk = true
				}
			} else {
				envOk = true
			}
		}
	}

	return value, expanded, envOk
}

func overrideValueWithEnv(value any, metadata map[string]string) (any, bool) {
	varName, ok := metadata["env"]
	if ok {
		envValue, ok := os.LookupEnv(varName)
		if ok {
			return envValue, true
		} else {
			return value, false
		}
	} else {
		return value, false
	}
}

func getFieldDefaultValue(metadata map[string]string) (any, bool) {
	defaultValue, ok := metadata["defaultValue"]
	if !ok {
		return nil, false
	}

	return defaultValue, true
}

func isValueRequired(metadata map[string]string) bool {
	return metadata["required"] == "true"
}
