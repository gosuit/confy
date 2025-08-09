package confy

import (
	"errors"
	"reflect"
)

func prepareStruct(to any) (reflect.Value, error) {
	out := reflect.ValueOf(to)

	if out.Kind() == reflect.Ptr && !out.IsNil() {
		out = out.Elem()
	} else {
		return reflect.Value{}, errors.New("config struct must be pointer and not nil")
	}

	if out.Kind() != reflect.Struct {
		return reflect.Value{}, errors.New("config must be struct")
	}

	return out, nil
}

func processStruct(out reflect.Value, fileData map[string]any, fileTag string) error {
	if out.Kind() != reflect.Struct {
		return errors.New("config must be struct")
	}

	for i := range out.NumField() {
		field := out.Field(i)
		fieldType := out.Type().Field(i)

		value, err := getFieldValue(field, fieldType, fileData, fileTag)
		if err != nil {
			return err
		}

		err = setFieldValue(field, fieldType, value, fileTag)
		if err != nil {
			return err
		}
	}

	return nil
}
