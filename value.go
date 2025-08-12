package confy

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func getStructData(data map[string]any, metadata map[string]string) (map[string]any, error) {
	key, ok := metadata["key"]
	if !ok {
		return nil, errors.New("internal error")
	}

	structData, ok := data[key]
	if !ok {
		return make(map[string]any), nil
	}

	if mapStructData, ok := structData.(map[string]any); ok {
		return mapStructData, nil
	} else {
		return nil, fmt.Errorf("invalid value for struct: %v", structData)
	}
}

func getFieldValue(f reflect.Value, data map[string]any, metadata map[string]string) (reflect.Value, error) {
	value, fileOk, err := getFieldFileValue(data, metadata)
	if err != nil {
		return reflect.Value{}, err
	}

	value, envOk := overrideValueWithEnv(value, metadata)
	if envOk {
		metadata["isValueEnv"] = "true"
	} else {
		metadata["isValueEnv"] = "false"
	}

	if !(fileOk || envOk) {
		var defaultOk bool

		value, defaultOk = getFieldDefaultValue(metadata)
		if !defaultOk {
			required, err := isValueRequired(metadata)
			if err != nil {
				return reflect.Value{}, err
			}

			if required {
				name, ok := metadata["name"]
				if !ok {
					return reflect.Value{}, errors.New("internal error")
				}

				return reflect.Value{}, fmt.Errorf("field %s is required", name)
			} else {
				result := reflect.New(f.Type().Elem()).Elem()

				return result, nil
			}
		}
	}

	return parseValue(f, value, metadata)
}

// TODO: make errors more usefull
func parseValue(f reflect.Value, value any, metadata map[string]string) (reflect.Value, error) {
	var isValueEnv bool

	isValueEnvString, ok := metadata["isValueEnv"]
	if !ok {
		return reflect.Value{}, errors.New("internal error")
	} else {
		isValueEnv = isValueEnvString == "true"
	}

	switch f.Type() {

	case reflect.TypeOf(time.Time{}):
		layout, ok := metadata["layout"]
		if !ok {
			layout = time.RFC3339
		}

		if stringValue, ok := value.(string); ok {
			timeValue, err := time.Parse(layout, stringValue)
			if err != nil {
				return reflect.Value{}, err
			}

			return reflect.ValueOf(timeValue), nil
		} else {
			return reflect.Value{}, errors.New("value for time.Time must be string")
		}

	case reflect.TypeOf(url.URL{}):
		if stringValue, ok := value.(string); ok {
			urlValue, err := url.Parse(stringValue)
			if err != nil {
				return reflect.Value{}, err
			}

			return reflect.ValueOf(*urlValue), nil
		} else {
			return reflect.Value{}, errors.New("value for url.URL must be string")
		}

	case reflect.TypeOf(time.Location{}):
		if stringValue, ok := value.(string); ok {
			locationValue, err := time.LoadLocation(stringValue)
			if err != nil {
				return reflect.Value{}, err
			}

			return reflect.ValueOf(*locationValue), nil
		} else {
			return reflect.Value{}, errors.New("value for time.Location must be string")
		}

	}

	switch f.Kind() {

	case reflect.Interface:
		return reflect.ValueOf(value), nil

	case reflect.String:
		if stringValue, ok := value.(string); ok {
			return reflect.ValueOf(stringValue), nil
		} else {
			return reflect.Value{}, errors.New("value for string must be string")
		}

	case reflect.Bool:
		if boolValue, ok := value.(bool); ok {
			return reflect.ValueOf(boolValue), nil
		} else if stringValue, ok := value.(string); ok && isValueEnv {
			boolValue, err := strconv.ParseBool(stringValue)
			if err != nil {
				return reflect.Value{}, err
			}

			return reflect.ValueOf(boolValue), nil
		} else {
			return reflect.Value{}, errors.New("value for bool must be bool")
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		if intValue, ok := value.(int); ok {
			if !f.OverflowInt(int64(intValue)) {
				return reflect.ValueOf(intValue), nil
			} else {
				return reflect.Value{}, errors.New("value for int is overflowed")
			}
		} else if stringValue, ok := value.(string); ok && isValueEnv {
			intValue, err := strconv.ParseInt(stringValue, 10, 64)
			if err != nil {
				return reflect.Value{}, err
			}

			if !f.OverflowInt(int64(intValue)) {
				return reflect.ValueOf(intValue), nil
			} else {
				return reflect.Value{}, errors.New("value for int is overflowed")
			}
		} else {
			return reflect.Value{}, errors.New("value for int must be int")
		}

	case reflect.Int64:
		stringValue, ok := value.(string)
		if f.Type() == reflect.TypeOf(time.Duration(0)) && ok {
			d, err := time.ParseDuration(stringValue)
			if err != nil {
				return reflect.Value{}, err
			}

			return reflect.ValueOf(d), nil
		} else if ok && isValueEnv {
			intValue, err := strconv.ParseInt(stringValue, 10, 64)
			if err != nil {
				return reflect.Value{}, err
			}

			if !f.OverflowInt(int64(intValue)) {
				return reflect.ValueOf(intValue), nil
			} else {
				return reflect.Value{}, errors.New("value for int is overflowed")
			}
		} else if intValue, ok := value.(int); ok {
			if !f.OverflowInt(int64(intValue)) {
				return reflect.ValueOf(intValue), nil
			} else {
				return reflect.Value{}, errors.New("value for int is overflowed")
			}
		} else {
			return reflect.Value{}, errors.New("value for int must be int")
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if uintValue, ok := value.(int); ok {
			if !f.OverflowUint(uint64(uintValue)) {
				return reflect.ValueOf(uintValue), nil
			} else {
				return reflect.Value{}, errors.New("value for uint is overflowed")
			}
		} else if stringValue, ok := value.(string); ok && isValueEnv {
			uintValue, err := strconv.ParseUint(stringValue, 10, 64)
			if err != nil {
				return reflect.Value{}, err
			}

			if !f.OverflowUint(uint64(uintValue)) {
				return reflect.ValueOf(uintValue), nil
			} else {
				return reflect.Value{}, errors.New("value for uint is overflowed")
			}
		} else {
			return reflect.Value{}, errors.New("value for uint must be uint")
		}

	case reflect.Float32, reflect.Float64:
		if floatValue, ok := value.(float64); ok {
			if !f.OverflowFloat(floatValue) {
				return reflect.ValueOf(floatValue), nil
			} else {
				return reflect.Value{}, errors.New("value for float is overflowed")
			}
		} else if stringValue, ok := value.(string); ok && isValueEnv {
			floatValue, err := strconv.ParseFloat(stringValue, 64)
			if err != nil {
				return reflect.Value{}, err
			}

			if !f.OverflowFloat(floatValue) {
				return reflect.ValueOf(floatValue), nil
			} else {
				return reflect.Value{}, errors.New("value for float is overflowed")
			}
		} else {
			return reflect.Value{}, errors.New("value for float must be float")
		}

	case reflect.Map:
		if f.Type().Key().Kind() != reflect.String {
			return reflect.Value{}, errors.New("map key must be string")
		}

		var data map[string]any

		if mapValue, ok := value.(map[string]any); ok {
			data = mapValue
		} else if stringValue, ok := value.(string); ok && isValueEnv {
			var err error

			data, err = parseMap(stringValue, metadata)
			if err != nil {
				return reflect.Value{}, err
			}
		} else {
			return reflect.Value{}, errors.New("value for map must be map")
		}

		newMap := reflect.MakeMap(f.Type())

		for k, v := range data {
			newValue := reflect.New(f.Type().Elem()).Elem()

			val, err := parseValue(newValue, v, metadata)
			if err != nil {
				return reflect.Value{}, err
			}

			newValue.Set(val)

			newKey := reflect.ValueOf(k)

			newMap.SetMapIndex(newKey, newValue)
		}

		return newMap, nil

	case reflect.Array:
		newArray := reflect.New(f.Type())

		var array []any

		if arrayValue, ok := value.([]any); ok {
			array = arrayValue
		} else if stringValue, ok := value.(string); ok && isValueEnv {
			var err error

			array, err = parseArray(stringValue, metadata)
			if err != nil {
				return reflect.Value{}, err
			}
		} else {
			return reflect.Value{}, errors.New("value for array must be array")
		}

		if len(array) > f.Type().Len() {
			return reflect.Value{}, errors.New("input array is longer then expected")
		}

		for i := range array {
			newValue := reflect.New(f.Type().Elem()).Elem()

			val, err := parseValue(newValue, array[i], metadata)
			if err != nil {
				return reflect.Value{}, err
			}

			newValue.Set(val)

			newArray.Index(i).Set(newValue)
		}

		return newArray, nil

	case reflect.Slice:
		newSlice := reflect.New(f.Type())

		var slice []any

		if sliceValue, ok := value.([]any); ok {
			slice = sliceValue
		} else if stringValue, ok := value.(string); ok && isValueEnv {
			var err error

			slice, err = parseArray(stringValue, metadata)
			if err != nil {
				return reflect.Value{}, err
			}
		} else {
			return reflect.Value{}, errors.New("value for array must be array")
		}

		for i := range slice {
			newValue := reflect.New(f.Type().Elem()).Elem()

			val, err := parseValue(newValue, slice[i], metadata)
			if err != nil {
				return reflect.Value{}, err
			}

			newValue.Set(val)

			newSlice.Set(reflect.Append(newValue, newValue))
		}

		return newSlice, nil

	default:
		return reflect.Value{}, fmt.Errorf("type '%s' is not supported", f.Type())

	}
}

func getFieldFileValue(data map[string]any, metadata map[string]string) (any, bool, error) {
	key, ok := metadata["key"]
	if !ok {
		return reflect.Value{}, false, errors.New("internal error")
	}

	value, ok := data[key]
	if ok {
		value = expandValue(value)
	}

	return value, ok, nil
}

func expandValue(value any) any {
	if strVal, ok := value.(string); ok {
		if len(strVal) > 3 && strVal[0] == '$' && strVal[1] == '{' && strVal[len(strVal)-1] == '}' {
			parts := strings.Split(strVal[2:len(strVal)-1], ":")

			if len(parts) > 1 {
				value, ok = os.LookupEnv(parts[0])
				if !ok {
					value = parts[1]
				}
			} else {
				value = os.Getenv(parts[0])
			}
		}
	}

	return value
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

func isValueRequired(metadata map[string]string) (bool, error) {
	required, ok := metadata["required"]
	if !ok {
		return false, errors.New("internal error")
	}

	return required == "true", nil
}

func parseMap(value string, metadata map[string]string) (map[string]any, error) {
	separator, ok := metadata["separator"]
	if !ok {
		return nil, errors.New("internal error")
	}

	items := strings.Split(value, separator)
	result := make(map[string]any)

	for _, item := range items {
		pair := strings.SplitN(item, ":", 2)

		if len(pair) != 2 {
			return nil, errors.New("invalid item in map")
		}

		result[pair[0]] = pair[1]
	}

	return result, nil
}

func parseArray(value string, metadata map[string]string) ([]any, error) {
	separator, ok := metadata["separator"]
	if !ok {
		return nil, errors.New("internal error")
	}

	stringArray := strings.Split(value, separator)
	var result []any

	for _, v := range stringArray {
		result = append(result, v)
	}

	return result, nil
}
