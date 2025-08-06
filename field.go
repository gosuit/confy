package confy

import (
	"os"
	"reflect"
	"strings"
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

func setFieldValue(field reflect.Value, value any) error {

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
