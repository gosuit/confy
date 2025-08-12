package confy

import (
	"errors"
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
	requiredTag = "required"

	// Default values
	defaultSeparator = ";"
)

func fillConfig(s any, data map[string]any, dataTag string) error {
	out := reflect.ValueOf(s)

	if out.Kind() == reflect.Pointer && !out.IsNil() {
		out = out.Elem()
	} else {
		return errors.New("config struct must be pointer and not nil")
	}

	if out.Kind() != reflect.Struct {
		return errors.New("config must be struct")
	}

	metadata := make(map[string]string)
	metadata["dataTag"] = dataTag

	return processStruct(out, data, metadata)
}

func processStruct(s reflect.Value, data map[string]any, metadata map[string]string) error {
	if s.Kind() != reflect.Struct {
		return errors.New("internal error")
	}

	for i := range s.NumField() {
		field := s.Field(i)
		fieldStructType := s.Type().Field(i)
		metadata := getFieldMetadata(fieldStructType, metadata)

		if err := processField(field, data, metadata); err != nil {
			return err
		}
	}

	return nil
}

func processField(f reflect.Value, data map[string]any, metadata map[string]string) error {
	if f.Kind() == reflect.Pointer {
		newValue := reflect.New(f.Type().Elem()).Elem()

		if err := processField(newValue, data, metadata); err != nil {
			return err
		}

		f.Set(newValue.Addr())

		return nil
	}

	if f.Kind() == reflect.Struct {
		structData, err := getStructData(data, metadata)
		if err != nil {
			return err
		}

		return processStruct(f, structData, metadata)
	}

	value, err := getFieldValue(f, data, metadata)
	if err != nil {
		return err
	}

	f.Set(value)

	return nil
}

func getFieldMetadata(fieldStructType reflect.StructField, commonMetadata map[string]string) map[string]string {
	metadata := make(map[string]string)

	key, ok := fieldStructType.Tag.Lookup(confyTag)
	if !ok {
		key, ok = fieldStructType.Tag.Lookup(commonMetadata["dataTag"])
		if !ok {
			key = strings.ToLower(fieldStructType.Name)
		}
	}

	metadata["key"] = key

	defaultValue, ok := fieldStructType.Tag.Lookup(defaultTag)
	if !ok {
		defaultValue, ok = fieldStructType.Tag.Lookup(envDefaultTag)
		if ok {
			metadata["defaultValue"] = defaultValue
		}
	} else {
		metadata["defaultValue"] = defaultValue
	}

	env, ok := fieldStructType.Tag.Lookup(envTag)
	if ok {
		metadata["env"] = env
	}

	layout, ok := fieldStructType.Tag.Lookup(layoutTag)
	if !ok {
		layout, ok = fieldStructType.Tag.Lookup(envLayoutTag)
		if ok {
			metadata["layout"] = layout
		}
	} else {
		metadata["layout"] = layout
	}

	separator, ok := fieldStructType.Tag.Lookup(envSeparatorTag)
	if !ok {
		separator = defaultSeparator
	}

	metadata["separator"] = separator

	required, ok := fieldStructType.Tag.Lookup(requiredTag)
	if !ok {
		required, ok = fieldStructType.Tag.Lookup(envRequiredTag)
		if !ok {
			required = "false"
		}
	}

	metadata["required"] = required

	metadata["name"] = fieldStructType.Name

	return nil
}
