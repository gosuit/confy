package confy

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	specificTypes = []reflect.Type{reflect.TypeOf(time.Time{}), reflect.TypeOf(url.URL{}), reflect.TypeOf(time.Location{}), reflect.TypeOf(time.Duration(0))}
)

func parseValue(f reflect.Value, value any, metadata map[string]string) error {

	switch f.Type() {

	case reflect.TypeOf(time.Time{}):
		return parseTime(f, value, metadata)

	case reflect.TypeOf(url.URL{}):
		return parseUrl(f, value, metadata)

	case reflect.TypeOf(time.Location{}):
		return parseTimeLocation(f, value, metadata)

	case reflect.TypeOf(time.Duration(0)):
		return parseTimeDuration(f, value, metadata)
	}

	switch f.Kind() {

	case reflect.Interface:
		f.Set(reflect.ValueOf(value))

		return nil

	case reflect.String:
		return parseString(f, value, metadata)

	case reflect.Bool:
		return parseBool(f, value, metadata)

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
		return fmt.Errorf("error while value parsing: the '%v' type of the '%s' field is not supported", f.Type(), metadata["name"])

	}
}

func parseTime(f reflect.Value, value any, metadata map[string]string) error {
	layout, ok := metadata["layout"]
	if !ok {
		layout = time.RFC3339
	}

	if stringValue, ok := value.(string); ok {
		timeValue, err := time.Parse(layout, stringValue)
		if err != nil {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be time.Time", metadata["name"])
		}

		f.Set(reflect.ValueOf(timeValue))
	} else {
		return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be string", metadata["name"])
	}

	return nil
}

func parseTimeLocation(f reflect.Value, value any, metadata map[string]string) error {
	if stringValue, ok := value.(string); ok {
		locationValue, err := time.LoadLocation(stringValue)
		if err != nil {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be time.Location", metadata["name"])
		}

		f.Set(reflect.ValueOf(*locationValue))
	} else {
		return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be string", metadata["name"])
	}

	return nil
}

func parseTimeDuration(f reflect.Value, value any, metadata map[string]string) error {
	if stringValue, ok := value.(string); ok {
		durationValue, err := time.ParseDuration(stringValue)
		if err != nil {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be time.Duration", metadata["name"])
		}

		f.Set(reflect.ValueOf(durationValue))
	} else {
		return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be string", metadata["name"])
	}

	return nil
}

func parseUrl(f reflect.Value, value any, metadata map[string]string) error {
	if stringValue, ok := value.(string); ok {
		urlValue, err := url.Parse(stringValue)
		if err != nil {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be URL", metadata["name"])
		}

		f.Set(reflect.ValueOf(*urlValue))
	} else {
		return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be string", metadata["name"])
	}

	return nil
}

func parseString(f reflect.Value, value any, metadata map[string]string) error {
	if stringValue, ok := value.(string); ok {
		f.SetString(stringValue)
	} else {
		return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be string", metadata["name"])
	}

	return nil
}

func parseBool(f reflect.Value, value any, metadata map[string]string) error {
	if boolValue, ok := value.(bool); ok {
		f.SetBool(boolValue)
	} else if stringValue, ok := value.(string); ok && (metadata["isValueEnv"] == "true" || metadata["isValueDefault"] == "true") {
		boolValue, err := strconv.ParseBool(stringValue)
		if err != nil {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be bool", metadata["name"])
		}

		f.SetBool(boolValue)
	} else {
		return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be bool", metadata["name"])
	}

	return nil
}

func parseInt(f reflect.Value, value any, metadata map[string]string) error {
	if intValue, ok := value.(int); ok {
		if !f.OverflowInt(int64(intValue)) {
			f.SetInt(int64(intValue))
		} else {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field is overflowed", metadata["name"])
		}
	} else if stringValue, ok := value.(string); ok && (metadata["isValueEnv"] == "true" || metadata["isValueDefault"] == "true") {
		intValue, err := strconv.ParseInt(stringValue, 10, 64)
		if err != nil {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be int", metadata["name"])
		}

		if !f.OverflowInt(int64(intValue)) {
			f.SetInt(int64(intValue))
		} else {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field is overflowed", metadata["name"])
		}
	} else {
		return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be int", metadata["name"])
	}

	return nil
}

func parseUint(f reflect.Value, value any, metadata map[string]string) error {
	if uintValue, ok := value.(int); ok {
		if !f.OverflowUint(uint64(uintValue)) {
			f.SetUint(uint64(uintValue))
		} else {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field is overflowed", metadata["name"])
		}
	} else if stringValue, ok := value.(string); ok && (metadata["isValueEnv"] == "true" || metadata["isValueDefault"] == "true") {
		uintValue, err := strconv.ParseUint(stringValue, 10, 64)
		if err != nil {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be uint", metadata["name"])
		}

		if !f.OverflowUint(uint64(uintValue)) {
			f.SetUint(uint64(uintValue))
		} else {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field is overflowed", metadata["name"])
		}
	} else {
		return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be uint", metadata["name"])
	}

	return nil
}

func parseFloat(f reflect.Value, value any, metadata map[string]string) error {
	if floatValue, ok := value.(float64); ok {
		if !f.OverflowFloat(floatValue) {
			f.SetFloat(floatValue)
		} else {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field is overflowed", metadata["name"])
		}
	} else if stringValue, ok := value.(string); ok && (metadata["isValueEnv"] == "true" || metadata["isValueDefault"] == "true") {
		floatValue, err := strconv.ParseFloat(stringValue, 64)
		if err != nil {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be float", metadata["name"])
		}

		if !f.OverflowFloat(floatValue) {
			f.SetFloat(floatValue)
		} else {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field is overflowed", metadata["name"])
		}
	} else {
		return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be float", metadata["name"])
	}

	return nil
}

func parseMap(f reflect.Value, value any, metadata map[string]string) error {
	if f.Type().Key().Kind() != reflect.String {
		return fmt.Errorf("error while value parsing: unsuppored type. type of '%s' field is a map with non-string key", metadata["name"])
	}

	var data map[string]any

	if mapValue, ok := value.(map[string]any); ok {
		data = mapValue
	} else if stringValue, ok := value.(string); ok && (metadata["isValueEnv"] == "true" || metadata["isValueDefault"] == "true") {
		items := strings.Split(stringValue, metadata["separator"])
		result := make(map[string]any)

		for _, item := range items {
			pair := strings.SplitN(item, ":", 2)

			if len(pair) != 2 {
				return fmt.Errorf("error while value parsing: invalid map value from environment for '%s' field", metadata["name"])
			}

			result[pair[0]] = pair[1]
		}

		data = result
	} else {
		return fmt.Errorf("error while value parsing: invalid value. the value for the '%s' field must be '%v'", metadata["name"], f.Type())
	}

	newMap := reflect.MakeMap(f.Type())

	for k, v := range data {
		newValue := reflect.New(f.Type().Elem()).Elem()

		if err := parseValue(newValue, v, metadata); err != nil {
			return err
		}

		newKey := reflect.ValueOf(k)

		newMap.SetMapIndex(newKey, newValue)
	}

	f.Set(newMap)

	return nil
}

func parseArray(f reflect.Value, value any, metadata map[string]string) error {
	var array []any

	if arrayValue, ok := value.([]any); ok {
		array = arrayValue
	} else if stringValue, ok := value.(string); ok && (metadata["isValueEnv"] == "true" || metadata["isValueDefault"] == "true") {
		stringArray := strings.Split(stringValue, metadata["separator"])
		var result []any

		for _, v := range stringArray {
			result = append(result, v)
		}

		array = result
	} else {
		return fmt.Errorf("error while value parsing: invalid value. the value for the '%s' field must be '%v'", metadata["name"], f.Type())
	}

	if len(array) > f.Type().Len() {
		return fmt.Errorf("error while value parsing: invalid value. the array value for the '%s' field is longer then %d", metadata["name"], f.Type().Len())
	}

	for i := range array {
		newValue := reflect.New(f.Type().Elem()).Elem()

		if err := parseValue(newValue, array[i], metadata); err != nil {
			return err
		}

		f.Index(i).Set(newValue)
	}

	return nil
}

func parseSlice(f reflect.Value, value any, metadata map[string]string) error {
	var slice []any

	if sliceValue, ok := value.([]any); ok {
		slice = sliceValue
	} else if stringValue, ok := value.(string); ok && (metadata["isValueEnv"] == "true" || metadata["isValueDefault"] == "true") {
		stringArray := strings.Split(stringValue, metadata["separator"])
		var result []any

		for _, v := range stringArray {
			result = append(result, v)
		}

		slice = result
	} else {
		return fmt.Errorf("error while value parsing: invalid value. the value for the '%s' field must be '%v'", metadata["name"], f.Type())
	}

	for i := range slice {
		newValue := reflect.New(f.Type().Elem()).Elem()

		if err := parseValue(newValue, slice[i], metadata); err != nil {
			return err
		}

		f.Set(reflect.Append(f, newValue))
	}

	return nil
}
