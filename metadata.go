package confy

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

const (
	configTag   = "config"
	requiredTag = "required"
	defaultTag  = "default"
	layoutTag   = "time-layout"
)

type fieldMetadata struct {
	key          string
	defaultValue string
	hasDefault   bool
	fieldPath    string
	required     bool
	timeLayout   string
}

func getFieldMetadata(fieldType reflect.StructField, parentPath string) (fieldMetadata, error) {
	fieldPath := getFieldPath(fieldType, parentPath)
	defaultValue, hasDefault, err := getDefault(fieldType, fieldPath)
	if err != nil {
		return fieldMetadata{}, err
	}

	return fieldMetadata{
		key:          getKey(fieldType),
		defaultValue: defaultValue,
		hasDefault:   hasDefault,
		fieldPath:    fieldPath,
		required:     getRequired(fieldType),
		timeLayout:   getTimeLayout(fieldType),
	}, nil
}

func getKey(fieldType reflect.StructField) string {
	key, ok := fieldType.Tag.Lookup(configTag)
	if !ok || key == "" {
		key = strings.ToLower(fieldType.Name)
	}

	return key
}

func getDefault(fieldType reflect.StructField, fieldPath string) (string, bool, error) {
	defaultValue, hasDefault := fieldType.Tag.Lookup(defaultTag)

	if hasDefault {
		switch fieldType.Type.Kind() {
		case reflect.Map, reflect.Array, reflect.Slice:
			return "", false, fmt.Errorf(
				"error while value parsing: default value is not supported for field '%s'",
				fieldPath,
			)

		case reflect.Struct:
			if _, ok := specificParsers[fieldType.Type]; !ok {
				return "", false, fmt.Errorf(
					"error while value parsing: default value is not supported for struct field '%s'",
					fieldPath,
				)
			}
		}
	}

	return defaultValue, hasDefault, nil
}

func getFieldPath(fieldType reflect.StructField, parentPath string) string {
	if parentPath != "" {
		return fmt.Sprintf("%s.%s", parentPath, fieldType.Name)
	}

	return fieldType.Name
}

func getRequired(fieldType reflect.StructField) bool {
	reqStr, ok := fieldType.Tag.Lookup(requiredTag)
	if !ok || reqStr != "true" {
		return false
	}

	return true
}

func getTimeLayout(fieldType reflect.StructField) string {
	layout, ok := fieldType.Tag.Lookup(layoutTag)
	if !ok || layout == "" {
		return time.RFC3339
	}

	return layout
}

func metaWithIndex(meta fieldMetadata, index int) fieldMetadata {
	return fieldMetadata{
		key:          meta.key,
		defaultValue: meta.defaultValue,
		hasDefault:   meta.hasDefault,
		fieldPath:    fmt.Sprintf("%s[%d]", meta.fieldPath, index),
		required:     meta.required,
		timeLayout:   meta.timeLayout,
	}
}

func metaWithKey(meta fieldMetadata, key string) fieldMetadata {
	return fieldMetadata{
		key:          meta.key,
		defaultValue: meta.defaultValue,
		hasDefault:   meta.hasDefault,
		fieldPath:    fmt.Sprintf("%s[%s]", meta.fieldPath, key),
		required:     meta.required,
		timeLayout:   meta.timeLayout,
	}
}
