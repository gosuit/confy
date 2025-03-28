package confy

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

const (
	// defaultSeparator is a default list and map separator character
	defaultSeparator = ","
)

// Supported tags
const (
	// tagEnv name of the environment variable or a list of names
	tagEnv = "env"

	// tagEnvLayout value parsing layout (for types like time.Time)
	tagEnvLayout = "env-layout"

	// tagEnvDefault default value
	tagEnvDefault = "env-default"

	// tagEnvSeparator custom list and map separator
	tagEnvSeparator = "env-separator"

	// tagEnvDescription environment variable description
	tagEnvDescription = "env-description"

	// tagEnvUpd flag to mark a field as updatable
	tagEnvUpd = "env-upd"

	// tagEnvRequired flag to mark a field as required
	tagEnvRequired = "env-required"

	// tagEnvPrefix flag to specify prefix for structure fields
	tagEnvPrefix = "env-prefix"
)

// readEnvVars reads environment variables to the provided configuration structure
func readEnvVars(cfg any, update bool) error {
	metaInfo, err := readStructMetadata(cfg)
	if err != nil {
		return err
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
func readStructMetadata(cfgRoot any) ([]structMeta, error) {
	type cfgNode struct {
		Val    any
		Prefix string
		Path   string
	}

	cfgStack := []cfgNode{{cfgRoot, "", ""}}
	metas := make([]structMeta, 0)

	for i := range cfgStack {

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
		for idx := range s.NumField() {
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
					prefix, _ := fType.Tag.Lookup(tagEnvPrefix)
					cfgStack = append(cfgStack, cfgNode{
						Val:    fld.Addr().Interface(),
						Prefix: sPrefix + prefix,
						Path:   fmt.Sprintf("%s%s.", cfgStack[i].Path, fType.Name),
					})
					continue
				}

				// process time.Time
				if l, ok := fType.Tag.Lookup(tagEnvLayout); ok {
					layout = &l
				}
			}

			// check is the field value can be changed
			if !s.Field(idx).CanSet() {
				continue
			}

			if def, ok := fType.Tag.Lookup(tagEnvDefault); ok {
				defValue = &def
			}

			if sep, ok := fType.Tag.Lookup(tagEnvSeparator); ok {
				separator = sep
			} else {
				separator = defaultSeparator
			}

			_, upd := fType.Tag.Lookup(tagEnvUpd)

			_, required := fType.Tag.Lookup(tagEnvRequired)

			envList := make([]string, 0)

			if envs, ok := fType.Tag.Lookup(tagEnv); ok && len(envs) != 0 {
				envList = strings.Split(envs, defaultSeparator)
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
				description: fType.Tag.Get(tagEnvDescription),
				updatable:   upd,
				required:    required,
				path:        cfgStack[i].Path,
			})
		}

	}

	return metas, nil
}
