package confy

import (
	"errors"
	"reflect"
)

func Load(path string, to any) error {
	data, err := loadFileData(path)
	if err != nil {
		return err
	}

	if err := loadConfigToStruct(data, to); err != nil {
		return err
	}

	return nil
}

func loadConfigToStruct(fileData map[string]any, to any) error {
	out := reflect.ValueOf(to)

	if out.Kind() != reflect.Pointer || out.IsNil() {
		return errors.New("the 'to' argument must be a non-nil pointer to struct")
	}

	out = out.Elem()

	if out.Kind() != reflect.Struct {
		return errors.New("the passed pointer does not point to the struct")
	}

	return processStruct(out, fileData, "")
}
