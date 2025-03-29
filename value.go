package confy

import (
	"encoding"
	"errors"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func setupValue(field reflect.Value, val any, env, envDefault, fileType, separator, layout string, last bool) error {
	var expanded bool
	if sval, ok := val.(string); ok {
		env, envDefault, expanded = updateTags(sval, env, envDefault)
	}

	if env != "" {
		eval, ok := os.LookupEnv(env)
		if ok {
			return parseValue(field, eval, separator, &layout)
		} else if expanded {
			return parseValue(field, envDefault, separator, &layout)
		}
	}

	if structParser, found := validStructs[field.Type()]; found {
		if sval, ok := val.(string); ok {
			return structParser(&field, sval, &layout)
		} else {
			return errors.New("invalid input data.")
		}
	}

	if sval, ok := val.(string); ok {
		if field.CanInterface() {
			if ct, ok := field.Interface().(encoding.TextUnmarshaler); ok {
				return ct.UnmarshalText([]byte(sval))
			} else if ctp, ok := field.Addr().Interface().(encoding.TextUnmarshaler); ok {
				return ctp.UnmarshalText([]byte(sval))
			}
		}
	}

	switch field.Kind() {

	case reflect.Interface:
		field.Set(reflect.ValueOf(val))

	case reflect.Map:
		if field.Type().Key().Kind() != reflect.String {
			return errors.New("Map key must be string")
		}

		newMap := reflect.MakeMap(field.Type())

		if mval, ok := val.(map[string]any); ok {
			for k, v := range mval {
				newVal := reflect.New(field.Type().Elem()).Elem()

				if err := setupValue(newVal, v, "", "", fileType, separator, layout, last); err != nil {
					return err
				}

				newKey := reflect.ValueOf(k)

				newMap.SetMapIndex(newKey, newVal)
			}
		} else {
			return errors.New("Value for map must be map")
		}

		field.Set(newMap)

	case reflect.Array:
		if aval, ok := val.([]any); ok {
			if len(aval) > field.Type().Len() {
				return errors.New("Input array is longer then expected")
			}

			for i := range aval {
				newVal := reflect.New(field.Type().Elem()).Elem()

				if err := setupValue(newVal, aval[i], "", "", fileType, separator, layout, last); err != nil {
					return err
				}

				field.Index(i).Set(newVal)
			}
		} else {
			return errors.New("Value for array must be array")
		}

	case reflect.Slice:
		if sval, ok := val.([]any); ok {
			for i := range sval {
				newVal := reflect.New(field.Type().Elem()).Elem()

				if err := setupValue(newVal, sval[i], "", "", fileType, separator, layout, last); err != nil {
					return err
				}

				field.Set(reflect.Append(field, newVal))
			}
		} else {
			return errors.New("Value for slice must be slice")
		}

	case reflect.Struct:
		if mval, ok := val.(map[string]any); ok {
			if err := mergeStruct(mval, field, "", last); err != nil {
				return err
			}
		} else {
			return errors.New("Value for struct must be struct")
		}

	case reflect.Ptr:
		newField := reflect.New(field.Type().Elem()).Elem()

		if err := setupValue(newField, val, "", "", fileType, separator, layout, last); err != nil {
			return err
		}

		field.Set(newField.Addr())

	case reflect.String:
		if sval, ok := val.(string); ok {
			field.SetString(sval)
		} else {
			return errors.New("Value for string must be string")
		}

	case reflect.Bool:
		if bval, ok := val.(bool); ok {
			field.SetBool(bval)
		} else {
			return errors.New("Value for bool must be bool")
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		if ival, ok := val.(int); ok {
			if !field.OverflowInt(int64(ival)) {
				field.SetInt(int64(ival))
			} else {
				return errors.New("Value for int is overflowed")
			}
		} else {
			return errors.New("Value for int must be int")
		}

	case reflect.Int64:
		sval, ok := val.(string)
		if field.Type() == reflect.TypeOf(time.Duration(0)) && ok {
			d, err := time.ParseDuration(sval)
			if err != nil {
				return err
			}
			field.SetInt(int64(d))
		} else if ival, ok := val.(int); ok {
			if !field.OverflowInt(int64(ival)) {
				field.SetInt(int64(ival))
			} else {
				return errors.New("Value for int is overflowed")
			}
		} else {
			return errors.New("Value for int must be int")
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if uval, ok := val.(int); ok {
			if !field.OverflowUint(uint64(uval)) {
				field.SetUint(uint64(uval))
			} else {
				return errors.New("Value for uint is overflowed")
			}
		} else {
			return errors.New("Value for uint must be uint")
		}

	case reflect.Float32, reflect.Float64:
		if fval, ok := val.(float64); ok {
			if !field.OverflowFloat(fval) {
				field.SetFloat(fval)
			} else {
				return errors.New("Value for float is overflowed")
			}
		} else {
			return errors.New("Value for float must be float")
		}

	default:
		return errors.New("Unsupported value type")
	}

	return nil
}

// parseValue parses value into the corresponding field.
// In case of maps and slices it uses provided separator to split raw value string
func parseValue(field reflect.Value, value, sep string, layout *string) error {
	valueType := field.Type()

	// look for supported struct parser
	// parsing of struct must be done before checking the implementation `encoding.TextUnmarshaler`
	// standard struct types already have the implementation `encoding.TextUnmarshaler` (for example `time.Time`)
	if structParser, found := validStructs[valueType]; found {
		return structParser(&field, value, layout)
	}

	if field.CanInterface() {
		if ct, ok := field.Interface().(encoding.TextUnmarshaler); ok {
			return ct.UnmarshalText([]byte(value))
		} else if ctp, ok := field.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return ctp.UnmarshalText([]byte(value))
		}
	}

	switch valueType.Kind() {

	// parse string value
	case reflect.String:
		field.SetString(value)

	// parse boolean value
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(b)

	// parse integer
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		number, err := strconv.ParseInt(value, 0, valueType.Bits())
		if err != nil {
			return err
		}
		field.SetInt(number)

	case reflect.Int64:
		if valueType == reflect.TypeOf(time.Duration(0)) {
			// try to parse time
			d, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			field.SetInt(int64(d))
		} else {
			// parse regular integer
			number, err := strconv.ParseInt(value, 0, valueType.Bits())
			if err != nil {
				return err
			}
			field.SetInt(number)
		}

	// parse unsigned integer value
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		number, err := strconv.ParseUint(value, 0, valueType.Bits())
		if err != nil {
			return err
		}
		field.SetUint(number)

	// parse floating point value
	case reflect.Float32, reflect.Float64:
		number, err := strconv.ParseFloat(value, valueType.Bits())
		if err != nil {
			return err
		}
		field.SetFloat(number)

	// parse sliced value
	case reflect.Slice:
		sliceValue, err := parseSlice(valueType, value, sep, layout)
		if err != nil {
			return err
		}

		field.Set(*sliceValue)

	// parse mapped value
	case reflect.Map:
		mapValue, err := parseMap(valueType, value, sep, layout)
		if err != nil {
			return err
		}

		field.Set(*mapValue)

	default:
		return fmt.Errorf("unsupported type %s.%s", valueType.PkgPath(), valueType.Name())
	}

	return nil
}

// parseSlice parses value into a slice of given type
func parseSlice(valueType reflect.Type, value string, sep string, layout *string) (*reflect.Value, error) {
	sliceValue := reflect.MakeSlice(valueType, 0, 0)
	if valueType.Elem().Kind() == reflect.Uint8 {
		sliceValue = reflect.ValueOf([]byte(value))
	} else if len(strings.TrimSpace(value)) != 0 {
		values := strings.Split(value, sep)
		sliceValue = reflect.MakeSlice(valueType, len(values), len(values))

		for i, val := range values {
			if err := parseValue(sliceValue.Index(i), val, sep, layout); err != nil {
				return nil, err
			}
		}
	}
	return &sliceValue, nil
}

// parseMap parses value into a map of given type
func parseMap(valueType reflect.Type, value string, sep string, layout *string) (*reflect.Value, error) {
	mapValue := reflect.MakeMap(valueType)
	if len(strings.TrimSpace(value)) != 0 {
		pairs := strings.SplitSeq(value, sep)
		for pair := range pairs {
			kvPair := strings.SplitN(pair, ":", 2)
			if len(kvPair) != 2 {
				return nil, fmt.Errorf("invalid map item: %q", pair)
			}
			k := reflect.New(valueType.Key()).Elem()
			err := parseValue(k, kvPair[0], sep, layout)
			if err != nil {
				return nil, err
			}
			v := reflect.New(valueType.Elem()).Elem()
			err = parseValue(v, kvPair[1], sep, layout)
			if err != nil {
				return nil, err
			}
			mapValue.SetMapIndex(k, v)
		}
	}
	return &mapValue, nil
}

// structMeta is a structure metadata entity
type structMeta struct {
	envList     []string
	fieldName   string
	fieldValue  reflect.Value
	defValue    *string
	layout      *string
	separator   string
	description string
	updatable   bool
	required    bool
	path        string
}

// isFieldValueZero determines if fieldValue empty or not
func (sm *structMeta) isFieldValueZero() bool {
	return sm.fieldValue.IsZero()
}

// parseFunc custom value parser function
type parseFunc func(*reflect.Value, string, *string) error

// Any specific supported struct can be added here
var validStructs = map[reflect.Type]parseFunc{

	reflect.TypeOf(time.Time{}): func(field *reflect.Value, value string, layout *string) error {
		var l string
		if layout != nil {
			l = *layout
		} else {
			l = time.RFC3339
		}
		val, err := time.Parse(l, value)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(val))
		return nil
	},

	reflect.TypeOf(url.URL{}): func(field *reflect.Value, value string, _ *string) error {
		val, err := url.Parse(value)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(*val))
		return nil
	},

	reflect.TypeOf(&time.Location{}): func(field *reflect.Value, value string, _ *string) error {
		loc, err := time.LoadLocation(value)
		if err != nil {
			return err
		}

		field.Set(reflect.ValueOf(loc))
		return nil
	},
}
