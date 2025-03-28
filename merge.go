package confy

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"time"
)

type structKey struct {
	Value    string
	Required bool
}

func mergeStruct(data map[string]any, out reflect.Value, fileType string) error {
	for i := range out.NumField() {
		field := out.Field(i)

		env := out.Type().Field(i).Tag.Get(TagEnv)
		envDefault := out.Type().Field(i).Tag.Get(TagEnvDefault)
		separator := out.Type().Field(i).Tag.Get(TagEnvSeparator)
		layout := out.Type().Field(i).Tag.Get(TagEnvLayout)

		key := setupKey(out, i, fileType)
		if key == "" {
			continue
		}

		val, ok := data[key]
		if !ok {
			eval, ok := os.LookupEnv(env)
			if ok {
				if err := parseValue(field, eval, separator, &layout); err != nil {
					return err
				}
			} else if envDefault != "" {
				if err := parseValue(field, envDefault, separator, &layout); err != nil {
					return err
				}
			} else {
				required := out.Type().Field(i).Tag.Get("required")

				if required == "true" {
					return errors.New("invalid input")
				}
			}
		} else {
			if err := setupValue(field, val, env, envDefault, fileType, separator, layout); err != nil {
				return err
			}
		}
	}

	return nil
}

func setupValue(field reflect.Value, val any, env, envDefault, fileType, separator, layout string) error {
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

				if err := setupValue(newVal, v, "", "", fileType, separator, layout); err != nil {
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

				if err := setupValue(newVal, aval[i], "", "", fileType, separator, layout); err != nil {
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

				if err := setupValue(newVal, sval[i], "", "", fileType, separator, layout); err != nil {
					return err
				}

				field.Set(reflect.Append(field, newVal))
			}
		} else {
			return errors.New("Value for slice must be slice")
		}

	case reflect.Struct:
		if mval, ok := val.(map[string]any); ok {
			if err := mergeStruct(mval, field, ""); err != nil {
				return err
			}
		} else {
			return errors.New("Value for struct must be struct")
		}

	case reflect.Ptr:
		newField := reflect.New(field.Type().Elem()).Elem()

		if err := setupValue(newField, val, "", "", fileType, separator, layout); err != nil {
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

func updateTags(val, env, envDefault string) (string, string, bool) {
	var expended bool

	if len(val) > 3 && val[0] == '$' && val[1] == '{' && val[len(val)-1] == '}' {
		expended = true
		parts := strings.Split(val[2:len(val)-1], ":")

		if len(parts) > 1 {
			env = parts[0]
			envDefault = strings.Join(parts[1:], ":")
		} else {
			env = parts[0]
		}
	}

	return env, envDefault, expended
}

func setupKey(out reflect.Value, index int, fileType string) string {
	tag, ok := out.Type().Field(index).Tag.Lookup("confy")
	if !ok {
		tag = out.Type().Field(index).Tag.Get(fileType)
	}

	if tag != "-" {
		if tag == "" {
			return strings.ToLower(out.Type().Field(index).Name)
		} else {
			return tag
		}
	} else {
		return ""
	}
}
