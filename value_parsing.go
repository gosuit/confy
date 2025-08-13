package confy

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func parseValue(f reflect.Value, value any, metadata map[string]string) (reflect.Value, error) {

	switch f.Type() {

	case reflect.TypeOf(time.Time{}):
		return parseTimeTime(value, metadata)

	case reflect.TypeOf(url.URL{}):
		return parseUrlUrl(value, metadata)

	case reflect.TypeOf(time.Location{}):
		return parseTimeLocation(value, metadata)

	case reflect.TypeOf(time.Duration(0)):
		return parseTimeDuration(value, metadata)
	}

	switch f.Kind() {

	case reflect.Interface:
		return reflect.ValueOf(value), nil

	case reflect.String:
		return parseString(value, metadata)

	case reflect.Bool:
		return parseBool(value, metadata)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return parseInt(f, value, metadata)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return parseUint(f, value, metadata)

	case reflect.Float32, reflect.Float64:
		return parseFloat(f, value, metadata)

	case reflect.Map:
		return parseMap(f, value, metadata)

	case reflect.Array:
		return parseArray(f, value, metadata)

	case reflect.Slice:
		return parseSlice(f, value, metadata)

	default:
		return reflect.Value{}, fmt.Errorf("error while value parsing: the '%v' type of the '%s' field is not supported", f.Type(), metadata["name"])

	}
}

func parseTimeTime(value any, metadata map[string]string) (reflect.Value, error) {
	layout, ok := metadata["layout"]
	if !ok {
		layout = time.RFC3339
	}

	if stringValue, ok := value.(string); ok {
		timeValue, err := time.Parse(layout, stringValue)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be time.Time", metadata["name"])
		}

		return reflect.ValueOf(timeValue), nil
	} else {
		return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be string", metadata["name"])
	}
}

func parseTimeLocation(value any, metadata map[string]string) (reflect.Value, error) {
	if stringValue, ok := value.(string); ok {
		locationValue, err := time.LoadLocation(stringValue)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be time.Location", metadata["name"])
		}

		return reflect.ValueOf(*locationValue), nil
	} else {
		return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be string", metadata["name"])
	}
}

func parseTimeDuration(value any, metadata map[string]string) (reflect.Value, error) {
	if stringValue, ok := value.(string); ok {
		durationValue, err := time.ParseDuration(stringValue)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be time.Duration", metadata["name"])
		}

		return reflect.ValueOf(durationValue), nil
	} else {
		return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be string", metadata["name"])
	}
}

func parseUrlUrl(value any, metadata map[string]string) (reflect.Value, error) {
	if stringValue, ok := value.(string); ok {
		urlValue, err := url.Parse(stringValue)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be URL", metadata["name"])
		}

		return reflect.ValueOf(*urlValue), nil
	} else {
		return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be string", metadata["name"])
	}
}

func parseString(value any, metadata map[string]string) (reflect.Value, error) {
	if stringValue, ok := value.(string); ok {
		return reflect.ValueOf(stringValue), nil
	} else {
		return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be string", metadata["name"])
	}
}

func parseBool(value any, metadata map[string]string) (reflect.Value, error) {
	if boolValue, ok := value.(bool); ok {
		return reflect.ValueOf(boolValue), nil
	} else if stringValue, ok := value.(string); ok && metadata["isValueEnv"] == "true" {
		boolValue, err := strconv.ParseBool(stringValue)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be bool", metadata["name"])
		}

		return reflect.ValueOf(boolValue), nil
	} else {
		return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be bool", metadata["name"])
	}
}

func parseInt(f reflect.Value, value any, metadata map[string]string) (reflect.Value, error) {
	if intValue, ok := value.(int); ok {
		if !f.OverflowInt(int64(intValue)) {
			return reflect.ValueOf(intValue), nil
		} else {
			return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field is overflowed", metadata["name"])
		}
	} else if stringValue, ok := value.(string); ok && metadata["isValueEnv"] == "true" {
		intValue, err := strconv.ParseInt(stringValue, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be int", metadata["name"])
		}

		if !f.OverflowInt(int64(intValue)) {
			return reflect.ValueOf(intValue), nil
		} else {
			return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field is overflowed", metadata["name"])
		}
	} else {
		return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be int", metadata["name"])
	}
}

func parseUint(f reflect.Value, value any, metadata map[string]string) (reflect.Value, error) {
	if uintValue, ok := value.(int); ok {
		if !f.OverflowUint(uint64(uintValue)) {
			return reflect.ValueOf(uintValue), nil
		} else {
			return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field is overflowed", metadata["name"])
		}
	} else if stringValue, ok := value.(string); ok && metadata["isValueEnv"] == "true" {
		uintValue, err := strconv.ParseUint(stringValue, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be uint", metadata["name"])
		}

		if !f.OverflowUint(uint64(uintValue)) {
			return reflect.ValueOf(uintValue), nil
		} else {
			return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field is overflowed", metadata["name"])
		}
	} else {
		return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be uint", metadata["name"])
	}
}

func parseFloat(f reflect.Value, value any, metadata map[string]string) (reflect.Value, error) {
	if floatValue, ok := value.(float64); ok {
		if !f.OverflowFloat(floatValue) {
			return reflect.ValueOf(floatValue), nil
		} else {
			return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field is overflowed", metadata["name"])
		}
	} else if stringValue, ok := value.(string); ok && metadata["isValueEnv"] == "true" {
		floatValue, err := strconv.ParseFloat(stringValue, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be float", metadata["name"])
		}

		if !f.OverflowFloat(floatValue) {
			return reflect.ValueOf(floatValue), nil
		} else {
			return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field is overflowed", metadata["name"])
		}
	} else {
		return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be float", metadata["name"])
	}
}

func parseMap(f reflect.Value, value any, metadata map[string]string) (reflect.Value, error) {
	if f.Type().Key().Kind() != reflect.String {
		return reflect.Value{}, fmt.Errorf("error while value parsing: unsuppored type. type of '%s' field is a map with non-string key", metadata["name"])
	}

	var data map[string]any

	if mapValue, ok := value.(map[string]any); ok {
		data = mapValue
	} else if stringValue, ok := value.(string); ok && metadata["isValueEnv"] == "true" {
		items := strings.Split(stringValue, metadata["separator"])
		result := make(map[string]any)

		for _, item := range items {
			pair := strings.SplitN(item, ":", 2)

			if len(pair) != 2 {
				return reflect.Value{}, fmt.Errorf("error while value parsing: invalid map value from environment for '%s' field", metadata["name"])
			}

			result[pair[0]] = pair[1]
		}

		data = result
	} else {
		return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for the '%s' field must be '%v'", metadata["name"], f.Type())
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
}

func parseArray(f reflect.Value, value any, metadata map[string]string) (reflect.Value, error) {
	newArray := reflect.New(f.Type())

	var array []any

	if arrayValue, ok := value.([]any); ok {
		array = arrayValue
	} else if stringValue, ok := value.(string); ok && metadata["isValueEnv"] == "true" {
		stringArray := strings.Split(stringValue, metadata["separator"])
		var result []any

		for _, v := range stringArray {
			result = append(result, v)
		}

		array = result
	} else {
		return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for the '%s' field must be '%v'", metadata["name"], f.Type())
	}

	if len(array) > f.Type().Len() {
		return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the array value for the '%s' field is longer then %d", metadata["name"], f.Type().Len())
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
}

func parseSlice(f reflect.Value, value any, metadata map[string]string) (reflect.Value, error) {
	newSlice := reflect.New(f.Type())

	var slice []any

	if sliceValue, ok := value.([]any); ok {
		slice = sliceValue
	} else if stringValue, ok := value.(string); ok && metadata["isValueEnv"] == "true" {
		stringArray := strings.Split(stringValue, metadata["separator"])
		var result []any

		for _, v := range stringArray {
			result = append(result, v)
		}

		slice = result
	} else {
		return reflect.Value{}, fmt.Errorf("error while value parsing: invalid value. the value for the '%s' field must be '%v'", metadata["name"], f.Type())
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
}
