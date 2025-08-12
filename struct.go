package confy

import (
	"errors"
	"fmt"
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

func fillConfig(cfg any, data map[string]any, dataTag string) error {
	out := reflect.ValueOf(cfg)

	if out.Kind() == reflect.Pointer && !out.IsNil() {
		out = out.Elem()
	} else {
		return errors.New("the 'to' argument must be a non-nil pointer to struct")
	}

	if out.Kind() != reflect.Struct {
		return errors.New("the passed pointer does not point to the struct")
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

	if value.Type() != f.Type() {
		return errors.New("internal error")
	}

	f.Set(value)

	return nil
}

func getFieldMetadata(fieldStructType reflect.StructField, commonMetadata map[string]string) map[string]string {
	metadata := make(map[string]string)

	// Set required metadata
	metadata["key"] = getMetadataKey(fieldStructType, commonMetadata)
	metadata["name"] = getMetadataName(fieldStructType, commonMetadata)
	metadata["required"] = getMetadataRequired(fieldStructType)
	metadata["separator"] = getMetadataSeparator(fieldStructType)

	// Set non-required metadata
	env, ok := getMetadataEnv(fieldStructType)
	if ok {
		metadata["env"] = env
	}

	defaultValue, ok := getMetadataDefaultValue(fieldStructType)
	if ok {
		metadata["defaultValue"] = defaultValue
	}

	layout, ok := getMetadataLayout(fieldStructType)
	if ok {
		metadata["layout"] = layout
	}

	return metadata
}

func getMetadataKey(fieldStructType reflect.StructField, metadata map[string]string) string {
	key, ok := fieldStructType.Tag.Lookup(confyTag)
	if !ok {
		key, ok = fieldStructType.Tag.Lookup(metadata["dataTag"])
		if !ok {
			key = strings.ToLower(fieldStructType.Name)
		}
	}

	return key
}

func getMetadataName(fieldStructType reflect.StructField, metadata map[string]string) string {
	name, ok := metadata["name"]
	if ok {
		return fmt.Sprintf("%s.%s", name, fieldStructType.Name)
	}

	return fieldStructType.Name
}

func getMetadataRequired(fieldStructType reflect.StructField) string {
	required, ok := fieldStructType.Tag.Lookup(requiredTag)
	if !ok {
		required, ok = fieldStructType.Tag.Lookup(envRequiredTag)
		if !ok {
			required = "false"
		}
	}

	return required
}

func getMetadataSeparator(fieldStructType reflect.StructField) string {
	separator, ok := fieldStructType.Tag.Lookup(envSeparatorTag)
	if !ok {
		separator = defaultSeparator
	}

	return separator
}

func getMetadataEnv(fieldStructType reflect.StructField) (string, bool) {
	return fieldStructType.Tag.Lookup(envTag)
}

func getMetadataDefaultValue(fieldStructType reflect.StructField) (string, bool) {
	defaultValue, ok := fieldStructType.Tag.Lookup(defaultTag)
	if !ok {
		defaultValue, ok = fieldStructType.Tag.Lookup(envDefaultTag)
		if ok {
			return defaultValue, true
		}
	} else {
		return defaultValue, true
	}

	return "", false
}

func getMetadataLayout(fieldStructType reflect.StructField) (string, bool) {
	layout, ok := fieldStructType.Tag.Lookup(layoutTag)
	if !ok {
		return fieldStructType.Tag.Lookup(envLayoutTag)
	} else {
		return layout, true
	}
}
