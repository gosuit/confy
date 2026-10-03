package confy

import (
	"fmt"
	"reflect"
)

func processStruct(s reflect.Value, value any, fieldPath string) error {
	if s.Kind() != reflect.Struct {
		return fmt.Errorf("internal error: field '%s' is not a struct, but it is passed as an argument to the processStruct function", fieldPath)
	}

	data, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("error while value parsing: invalid value for '%s' struct field", fieldPath)
	}

	for i := range s.NumField() {
		field := s.Field(i)
		fieldType := s.Type().Field(i)

		metadata, err := getFieldMetadata(fieldType, fieldPath)
		if err != nil {
			return err
		}

		if metadata.key == "-" {
			continue
		}

		if err := processField(field, metadata, data); err != nil {
			return err
		}
	}

	return nil
}

func processField(field reflect.Value, metadata fieldMetadata, data map[string]any) error {
	if !field.CanSet() {
		return nil
	}

	rawValue, exists := data[metadata.key]

	if !exists || rawValue == nil {
		return handleMissing(field, metadata)
	}

	return applyValue(field, metadata, rawValue)
}

func processMap(field reflect.Value, metadata fieldMetadata, value any) error {
	if field.Type().Key().Kind() != reflect.String {
		return fmt.Errorf("error while value parsing: unsupported type. type of '%s' field is a map with non-string key", metadata.fieldPath)
	}

	data, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("error while value parsing: invalid value. the value for the '%s' field must be '%v'", metadata.fieldPath, field.Type())
	}

	newMap := reflect.MakeMapWithSize(field.Type(), len(data))

	for k, v := range data {
		newValue := reflect.New(field.Type().Elem()).Elem()

		if err := applyValue(newValue, metaWithKey(metadata, k), v); err != nil {
			return err
		}

		newKey := reflect.New(field.Type().Key()).Elem()
		newKey.SetString(k)

		newMap.SetMapIndex(newKey, newValue)
	}

	field.Set(newMap)

	return nil
}

func processArray(field reflect.Value, metadata fieldMetadata, value any) error {
	array, ok := value.([]any)
	if !ok {
		return fmt.Errorf("error while value parsing: invalid value. the value for the '%s' field must be '%v'", metadata.fieldPath, field.Type())
	}

	if len(array) > field.Type().Len() {
		return fmt.Errorf("error while value parsing: invalid value. the array value for the '%s' field is longer than %d", metadata.fieldPath, field.Type().Len())
	}

	field.Set(reflect.Zero(field.Type()))

	for i := range array {
		newValue := reflect.New(field.Type().Elem()).Elem()

		if err := applyValue(newValue, metaWithIndex(metadata, i), array[i]); err != nil {
			return err
		}

		field.Index(i).Set(newValue)
	}

	return nil
}

func processSlice(field reflect.Value, metadata fieldMetadata, value any) error {
	slice, ok := value.([]any)
	if !ok {
		return fmt.Errorf("error while value parsing: invalid value. the value for the '%s' field must be '%v'", metadata.fieldPath, field.Type())
	}

	newSlice := reflect.MakeSlice(field.Type(), 0, len(slice))

	for i := range slice {
		newValue := reflect.New(field.Type().Elem()).Elem()

		if err := applyValue(newValue, metaWithIndex(metadata, i), slice[i]); err != nil {
			return err
		}

		newSlice = reflect.Append(newSlice, newValue)
	}

	field.Set(newSlice)

	return nil
}

func applyValue(field reflect.Value, metadata fieldMetadata, value any) error {
	if field.Kind() == reflect.Pointer {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}

		return applyValue(field.Elem(), metadata, value)
	}

	if _, ok := specificParsers[field.Type()]; ok {
		return processValue(field, metadata, value, canExpand(field))
	}

	switch field.Kind() {

	case reflect.Struct:
		return processStruct(field, value, metadata.fieldPath)

	case reflect.Map:
		return processMap(field, metadata, value)

	case reflect.Array:
		return processArray(field, metadata, value)

	case reflect.Slice:
		return processSlice(field, metadata, value)
	}

	return processValue(field, metadata, value, canExpand(field))
}

func handleMissing(field reflect.Value, metadata fieldMetadata) error {
	if metadata.required {
		return fmt.Errorf("error while value parsing: value for '%s' field is required", metadata.fieldPath)
	}

	if metadata.hasDefault {
		return applyDefault(field, metadata)
	}

	return nil
}

func applyDefault(field reflect.Value, metadata fieldMetadata) error {
	if field.Kind() == reflect.Pointer {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}

		return applyDefault(field.Elem(), metadata)
	}

	return processValue(field, metadata, metadata.defaultValue, false)
}
