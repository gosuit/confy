package confy

import (
	"encoding"
	"errors"
	"os"
	"reflect"
	"strings"
)

type structKey struct {
	Value    string
	Required bool
}

func mergeStruct(data map[string]any, out reflect.Value, fileType string) error {
	for i := range out.NumField() {
		field := out.Field(i)

		env := out.Type().Field(i).Tag.Get(tagEnv)
		envDefault := out.Type().Field(i).Tag.Get(tagEnvDefault)

		separator := out.Type().Field(i).Tag.Get(tagEnvSeparator)
		layout := out.Type().Field(i).Tag.Get(tagEnvLayout)

		if separator == "" {
			separator = defaultSeparator
		}

		key := setupKey(out, i, fileType)
		if key == "" {
			continue
		}

		val, ok := data[key]
		if !ok {
			_, found := validStructs[field.Type()]
			_, text := field.Interface().(encoding.TextUnmarshaler)
			_, pointer := field.Addr().Interface().(encoding.TextUnmarshaler)

			if field.Kind() == reflect.Struct && !found && !text && !pointer {
				if err := mergeStruct(make(map[string]any), field, fileType); err != nil {
					return err
				}

				continue
			}

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
