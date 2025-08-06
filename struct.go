package confy

import (
	"errors"
	"reflect"
)

func processStruct(out reflect.Value, fileData map[string]any, fileTag string) error {
	if out.Kind() != reflect.Struct {
		return errors.New("config must be struct")
	}

	for i := range out.NumField() {
		field := out.Field(i)
		fieldType := out.Type().Field(i)

		value := getFieldValue(fieldType, fileData, fileTag)

		err := setFieldValue(field, value)
		if err != nil {
			return err
		}
	}

	return nil
}
