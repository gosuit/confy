package confy

import (
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"time"
)

type valueParser func(field reflect.Value, metadata fieldMetadata, value any) error

var specificParsers = map[reflect.Type]valueParser{
	reflect.TypeFor[time.Time]():     parseTime,
	reflect.TypeFor[time.Location](): parseTimeLocation,
	reflect.TypeFor[time.Duration](): parseTimeDuration,
	reflect.TypeFor[url.URL]():       parseUrl,
}

func processValue(field reflect.Value, metadata fieldMetadata, value any, allowExpand bool) error {
	if value == nil {
		field.Set(reflect.Zero(field.Type()))
		return nil
	}

	if allowExpand {
		v, err := expandValue(value, metadata)
		if err != nil {
			return err
		}
		value = v
	}

	if parser, ok := specificParsers[field.Type()]; ok {
		return parser(field, metadata, value)
	}

	switch field.Kind() {

	case reflect.Interface:
		return setInterface(field, metadata, value)

	case reflect.String:
		return parseString(field, metadata, value)

	case reflect.Bool:
		return parseBool(field, metadata, value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return parseInt(field, metadata, value)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return parseUint(field, metadata, value)

	case reflect.Float32, reflect.Float64:
		return parseFloat(field, metadata, value)

	default:
		return fmt.Errorf("error while value parsing: the '%v' type of the '%s' field is not supported", field.Type(), metadata.fieldPath)
	}
}

func parseTime(field reflect.Value, metadata fieldMetadata, value any) error {
	if stringValue, ok := value.(string); ok {
		timeValue, err := time.Parse(metadata.timeLayout, stringValue)
		if err != nil {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be time.Time", metadata.fieldPath)
		}

		field.Set(reflect.ValueOf(timeValue))
	} else {
		return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be string", metadata.fieldPath)
	}

	return nil
}

func parseTimeLocation(field reflect.Value, metadata fieldMetadata, value any) error {
	if stringValue, ok := value.(string); ok {
		locationValue, err := time.LoadLocation(stringValue)
		if err != nil {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be time.Location", metadata.fieldPath)
		}

		field.Set(reflect.ValueOf(*locationValue))
	} else {
		return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be string", metadata.fieldPath)
	}

	return nil
}

func parseTimeDuration(field reflect.Value, metadata fieldMetadata, value any) error {
	if stringValue, ok := value.(string); ok {
		durationValue, err := time.ParseDuration(stringValue)
		if err != nil {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be time.Duration", metadata.fieldPath)
		}

		field.Set(reflect.ValueOf(durationValue))
	} else {
		return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be string", metadata.fieldPath)
	}

	return nil
}

func parseUrl(field reflect.Value, metadata fieldMetadata, value any) error {
	if stringValue, ok := value.(string); ok {
		urlValue, err := url.Parse(stringValue)
		if err != nil {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be URL", metadata.fieldPath)
		}

		field.Set(reflect.ValueOf(*urlValue))
	} else {
		return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be string", metadata.fieldPath)
	}

	return nil
}

func setInterface(field reflect.Value, metadata fieldMetadata, value any) error {
	v := reflect.ValueOf(value)

	if !v.Type().AssignableTo(field.Type()) {
		return fmt.Errorf("error while value parsing: '%s' field cannot accept value of type '%v'", metadata.fieldPath, v.Type())
	}

	field.Set(v)

	return nil
}

func parseString(field reflect.Value, metadata fieldMetadata, value any) error {
	if stringValue, ok := value.(string); ok {
		field.SetString(stringValue)
	} else {
		return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be string", metadata.fieldPath)
	}

	return nil
}

func parseBool(field reflect.Value, metadata fieldMetadata, value any) error {
	if boolValue, ok := value.(bool); ok {
		field.SetBool(boolValue)
	} else if stringValue, ok := value.(string); ok {
		boolValue, err := strconv.ParseBool(stringValue)
		if err != nil {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be bool", metadata.fieldPath)
		}

		field.SetBool(boolValue)
	} else {
		return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be bool", metadata.fieldPath)
	}

	return nil
}

func parseInt(field reflect.Value, metadata fieldMetadata, value any) error {
	if intValue, ok := value.(int); ok {
		if !field.OverflowInt(int64(intValue)) {
			field.SetInt(int64(intValue))
		} else {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field is overflowed", metadata.fieldPath)
		}
	} else if stringValue, ok := value.(string); ok {
		intValue, err := strconv.ParseInt(stringValue, 10, 64)
		if err != nil {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be int", metadata.fieldPath)
		}

		if !field.OverflowInt(intValue) {
			field.SetInt(intValue)
		} else {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field is overflowed", metadata.fieldPath)
		}
	} else {
		return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be int", metadata.fieldPath)
	}

	return nil
}

func parseUint(field reflect.Value, metadata fieldMetadata, value any) error {
	if intValue, ok := value.(int); ok {
		if intValue < 0 {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field can not be negative", metadata.fieldPath)
		}

		if !field.OverflowUint(uint64(intValue)) {
			field.SetUint(uint64(intValue))
		} else {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field is overflowed", metadata.fieldPath)
		}
	} else if stringValue, ok := value.(string); ok {
		uintValue, err := strconv.ParseUint(stringValue, 10, 64)
		if err != nil {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be uint", metadata.fieldPath)
		}

		if !field.OverflowUint(uintValue) {
			field.SetUint(uintValue)
		} else {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field is overflowed", metadata.fieldPath)
		}
	} else {
		return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be uint", metadata.fieldPath)
	}

	return nil
}

func parseFloat(field reflect.Value, metadata fieldMetadata, value any) error {
	if floatValue, ok := value.(float64); ok {
		if !field.OverflowFloat(floatValue) {
			field.SetFloat(floatValue)
		} else {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field is overflowed", metadata.fieldPath)
		}
	} else if stringValue, ok := value.(string); ok {
		floatValue, err := strconv.ParseFloat(stringValue, 64)
		if err != nil {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be float", metadata.fieldPath)
		}

		if !field.OverflowFloat(floatValue) {
			field.SetFloat(floatValue)
		} else {
			return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field is overflowed", metadata.fieldPath)
		}
	} else {
		return fmt.Errorf("error while value parsing: invalid value. the value for '%s' field must be float", metadata.fieldPath)
	}

	return nil
}

func canExpand(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		return false

	case reflect.Struct:
		if _, ok := specificParsers[val.Type()]; !ok {
			return false
		}
	}

	return true
}

func expandValue(value any, metadata fieldMetadata) (any, error) {
	s, ok := value.(string)
	if !ok {
		return value, nil
	}

	if len(s) < 4 || s[0] != '$' || s[1] != '{' || s[len(s)-1] != '}' {
		return value, nil
	}

	key := s[2 : len(s)-1]

	val, ok := os.LookupEnv(key)
	if !ok {
		return nil, fmt.Errorf(
			"error while value parsing: environment variable '%s' for '%s' field is not set",
			key, metadata.fieldPath,
		)
	}

	return val, nil
}
