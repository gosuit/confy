package confy

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/joho/godotenv"
)

const (
	// DefaultSeparator is a default list and map separator character
	DefaultSeparator = ","
)

// Setter is an interface for a custom value setter.
//
// To implement a custom value setter you need to add a SetValue function to your type that will receive a string raw value:
//
//	type MyField string
//
//	func (f *MyField) SetValue(s string) error {
//		if s == "" {
//			return fmt.Errorf("field value can't be empty")
//		}
//		*f = MyField("my field is: " + s)
//		return nil
//	}
type Setter interface {
	SetValue(string) error
}

// Updater gives an ability to implement custom update function for a field or a whole structure
type Updater interface {
	Update() error
}

// Supported tags
const (
	// TagEnv name of the environment variable or a list of names
	TagEnv = "env"

	// TagEnvLayout value parsing layout (for types like time.Time)
	TagEnvLayout = "env-layout"

	// TagEnvDefault default value
	TagEnvDefault = "env-default"

	// TagEnvSeparator custom list and map separator
	TagEnvSeparator = "env-separator"

	// TagEnvDescription environment variable description
	TagEnvDescription = "env-description"

	// TagEnvUpd flag to mark a field as updatable
	TagEnvUpd = "env-upd"

	// TagEnvRequired flag to mark a field as required
	TagEnvRequired = "env-required"

	// TagEnvPrefix flag to specify prefix for structure fields
	TagEnvPrefix = "env-prefix"
)

// parseENV, in fact, doesn't fill the structure with environment variable values.
// It just parses ENV file and sets all variables to the environment.
// Thus, the structure should be filled at the next steps.
func parseENV(r io.Reader, _ interface{}) error {
	vars, err := godotenv.Parse(r)
	if err != nil {
		return err
	}

	for env, val := range vars {
		if err = os.Setenv(env, val); err != nil {
			return fmt.Errorf("set environment: %w", err)
		}
	}

	return nil
}

// readEnvVars reads environment variables to the provided configuration structure
func readEnvVars(cfg interface{}, update bool) error {
	metaInfo, err := readStructMetadata(cfg)
	if err != nil {
		return err
	}

	if updater, ok := cfg.(Updater); ok {
		if err = updater.Update(); err != nil {
			return err
		}
	}

	for _, meta := range metaInfo {
		// update only updatable fields
		if update && !meta.updatable {
			continue
		}

		var rawValue *string

		for _, env := range meta.envList {
			if value, ok := os.LookupEnv(env); ok {
				rawValue = &value
				break
			}
		}

		var envName string
		if len(meta.envList) > 0 {
			envName = meta.envList[0]
		}

		if rawValue == nil && meta.required && meta.isFieldValueZero() {
			return fmt.Errorf("field %q is required but the value is not provided",
				meta.path+meta.fieldName,
			)
		}

		if rawValue == nil && meta.isFieldValueZero() {
			rawValue = meta.defValue
		}

		if rawValue == nil {
			continue
		}

		if err = parseValue(meta.fieldValue, *rawValue, meta.separator, meta.layout); err != nil {
			return fmt.Errorf("parsing field %q env %q: %v",
				meta.path+meta.fieldName, envName, err,
			)
		}
	}

	return nil
}

// readStructMetadata reads structure metadata (types, tags, etc.)
func readStructMetadata(cfgRoot interface{}) ([]structMeta, error) {
	type cfgNode struct {
		Val    interface{}
		Prefix string
		Path   string
	}

	cfgStack := []cfgNode{{cfgRoot, "", ""}}
	metas := make([]structMeta, 0)

	for i := 0; i < len(cfgStack); i++ {

		s := reflect.ValueOf(cfgStack[i].Val)
		sPrefix := cfgStack[i].Prefix

		// unwrap pointer
		if s.Kind() == reflect.Ptr {
			s = s.Elem()
		}

		// process only structures
		if s.Kind() != reflect.Struct {
			return nil, fmt.Errorf("wrong type %v", s.Kind())
		}
		typeInfo := s.Type()

		// read tags
		for idx := 0; idx < s.NumField(); idx++ {
			fType := typeInfo.Field(idx)

			var (
				defValue  *string
				layout    *string
				separator string
			)

			// process nested structure (except of supported ones)
			if fld := s.Field(idx); fld.Kind() == reflect.Struct {
				//skip unexported
				if !fld.CanInterface() {
					continue
				}
				// add structure to parsing stack
				if _, found := validStructs[fld.Type()]; !found {
					prefix, _ := fType.Tag.Lookup(TagEnvPrefix)
					cfgStack = append(cfgStack, cfgNode{
						Val:    fld.Addr().Interface(),
						Prefix: sPrefix + prefix,
						Path:   fmt.Sprintf("%s%s.", cfgStack[i].Path, fType.Name),
					})
					continue
				}

				// process time.Time
				if l, ok := fType.Tag.Lookup(TagEnvLayout); ok {
					layout = &l
				}
			}

			// check is the field value can be changed
			if !s.Field(idx).CanSet() {
				continue
			}

			if def, ok := fType.Tag.Lookup(TagEnvDefault); ok {
				defValue = &def
			}

			if sep, ok := fType.Tag.Lookup(TagEnvSeparator); ok {
				separator = sep
			} else {
				separator = DefaultSeparator
			}

			_, upd := fType.Tag.Lookup(TagEnvUpd)

			_, required := fType.Tag.Lookup(TagEnvRequired)

			envList := make([]string, 0)

			if envs, ok := fType.Tag.Lookup(TagEnv); ok && len(envs) != 0 {
				envList = strings.Split(envs, DefaultSeparator)
				if sPrefix != "" {
					for i := range envList {
						envList[i] = sPrefix + envList[i]
					}
				}
			}

			metas = append(metas, structMeta{
				envList:     envList,
				fieldName:   s.Type().Field(idx).Name,
				fieldValue:  s.Field(idx),
				defValue:    defValue,
				layout:      layout,
				separator:   separator,
				description: fType.Tag.Get(TagEnvDescription),
				updatable:   upd,
				required:    required,
				path:        cfgStack[i].Path,
			})
		}

	}

	return metas, nil
}
