package confy

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"time"
)

const (
	// File tags
	confyTag = "confy"
	yamlTag  = "yaml"
	jsonTag  = "json"
	tomlTag  = "toml"

	// Env tags
	envTag          = "env"
	envDefaultTag   = "env-default"
	envLayoutTag    = "env-layout"
	envSeparatorTag = "env-separator"
	envRequiredTag  = "env-required"

	// Common tags
	defaultTag  = "default"
	layoutTag   = "layout"
	validateTag = "validate"

	// Default values
	defaultRootPath   = "config"
	defaultEnvVarName = "ENVIRONMENT"
	defaultSeparator  = ";"
)

func setFieldValue(field reflect.Value, value any, fileTag string) error {
	switch field.Kind() {

	case reflect.Interface:
		field.Set(reflect.ValueOf(value))

	case reflect.Map:
		if field.Type().Key().Kind() != reflect.String {
			return errors.New("map key must be string")
		}

		if mapValue, ok := value.(map[string]any); ok {
			newMap := reflect.MakeMap(field.Type())

			for k, v := range mapValue {
				newValue := reflect.New(field.Type().Elem()).Elem()

				if err := setFieldValue(newValue, v, fileTag); err != nil {
					return err
				}

				newKey := reflect.ValueOf(k)

				newMap.SetMapIndex(newKey, newValue)
			}

			field.Set(newMap)
		} else {
			return errors.New("value for map must be map")
		}

	case reflect.Array:
		if arrayValue, ok := value.([]any); ok {
			if len(arrayValue) > field.Type().Len() {
				return errors.New("input array is longer then expected")
			}

			for i := range arrayValue {
				newValue := reflect.New(field.Type().Elem()).Elem()

				if err := setFieldValue(newValue, arrayValue[i], fileTag); err != nil {
					return err
				}

				field.Index(i).Set(newValue)
			}
		} else {
			return errors.New("value for array must be array")
		}

	case reflect.Slice:
		if sliceValue, ok := value.([]any); ok {
			for i := range sliceValue {
				newValue := reflect.New(field.Type().Elem()).Elem()

				if err := setFieldValue(newValue, sliceValue[i], fileTag); err != nil {
					return err
				}

				field.Set(reflect.Append(field, newValue))
			}
		} else {
			return errors.New("value for slice must be slice")
		}

	case reflect.Struct:
		if mapValue, ok := value.(map[string]any); ok {
			if err := processStruct(field, mapValue, fileTag); err != nil {
				return err
			}
		} else {
			return errors.New("value for struct must be struct")
		}

	case reflect.Ptr:
		newField := reflect.New(field.Type().Elem()).Elem()

		if err := setFieldValue(newField, value, fileTag); err != nil {
			return err
		}

		field.Set(newField.Addr())

	case reflect.String:
		if stringValue, ok := value.(string); ok {
			field.SetString(stringValue)
		} else {
			return errors.New("value for string must be string")
		}

	case reflect.Bool:
		if boolValue, ok := value.(bool); ok {
			field.SetBool(boolValue)
		} else {
			return errors.New("value for bool must be bool")
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		if intValue, ok := value.(int); ok {
			if !field.OverflowInt(int64(intValue)) {
				field.SetInt(int64(intValue))
			} else {
				return errors.New("value for int is overflowed")
			}
		} else {
			return errors.New("value for int must be int")
		}

	case reflect.Int64:
		stringValue, ok := value.(string)
		if field.Type() == reflect.TypeOf(time.Duration(0)) && ok {
			d, err := time.ParseDuration(stringValue)
			if err != nil {
				return err
			}
			field.SetInt(int64(d))
		} else if intValue, ok := value.(int); ok {
			if !field.OverflowInt(int64(intValue)) {
				field.SetInt(int64(intValue))
			} else {
				return errors.New("value for int is overflowed")
			}
		} else {
			return errors.New("value for int must be int")
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if uintValue, ok := value.(int); ok {
			if !field.OverflowUint(uint64(uintValue)) {
				field.SetUint(uint64(uintValue))
			} else {
				return errors.New("value for uint is overflowed")
			}
		} else {
			return errors.New("value for uint must be uint")
		}

	case reflect.Float32, reflect.Float64:
		if floatValue, ok := value.(float64); ok {
			if !field.OverflowFloat(floatValue) {
				field.SetFloat(floatValue)
			} else {
				return errors.New("value for float is overflowed")
			}
		} else {
			return errors.New("value for float must be float")
		}

	default:
		return errors.New("unsupported value type")

	}

	return nil
}

func getFieldValue(fieldType reflect.StructField, fileData map[string]any, fileTag string) any {
	value, fileOk := getFieldFileValue(fieldType, fileData, fileTag)

	value, envOk := overrideValueWithEnv(value, fieldType)

	if !(fileOk || envOk) {
		value = getFieldDefaultValue(fieldType)
	}

	return value
}

func getFieldFileValue(fieldType reflect.StructField, fileData map[string]any, fileTag string) (any, bool) {
	tag := getFieldFileTag(fieldType, fileTag)

	value, ok := fileData[tag]
	if ok {
		value = expandValue(value)
	}

	return value, ok
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

func getFieldFileTag(fieldType reflect.StructField, fileTag string) string {
	tag, ok := fieldType.Tag.Lookup(confyTag)
	if !ok {
		tag, ok = fieldType.Tag.Lookup(fileTag)
		if !ok {
			tag = strings.ToLower(fieldType.Name)
		}
	}

	return tag
}

func overrideValueWithEnv(value any, fieldType reflect.StructField) (any, bool) {
	tag, ok := fieldType.Tag.Lookup(envTag)
	if ok {
		envValue, ok := os.LookupEnv(tag)
		if ok {
			return envValue, true
		} else {
			return value, false
		}
	} else {
		return value, false
	}
}

func getFieldDefaultValue(fieldType reflect.StructField) any {
	value, ok := fieldType.Tag.Lookup(defaultTag)
	if !ok {
		value = fieldType.Tag.Get(envDefaultTag)
	}

	return value
}
