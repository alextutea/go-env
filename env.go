package env

import (
	"encoding/json"
	"fmt"
	"github.com/alextutea/go-env/errs"
	"github.com/alextutea/go-tags"
	"github.com/pkg/errors"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const (
	tagOptionRequired = "required"
	tagOptionDefault  = "default"
	TagName           = "env"
	envKeyConnector   = "_"
)

func Unmarshal(i interface{}, filePaths ...string) error {
	return UnmarshalMap(environToMap(os.Environ()), i, filePaths...)
}

func UnmarshalFile(path string, i interface{}) error {
	return UnmarshalMap(environToMap(os.Environ()), i, path)
}

func UnmarshalMap(envMap map[string]string, i interface{}, paths ...string) error {
	v := reflect.ValueOf(i)
	t := v.Type()

	if t.Kind() != reflect.Ptr || v.IsNil() || t.Elem().Kind() != reflect.Struct {
		return errs.NewTargetNotPtrToStructError(t)
	}
	v = v.Elem()
	t = t.Elem()

	cfgMap, err := parseCfgFiles(paths)
	if err != nil {
		return errors.Wrap(err, "parsing cfg files")
	}

	entryMap := parseStruct(v, envKeyConnector)

	err = prioUnmarshal(entryMap, envMap, cfgMap)
	if err != nil {
		return errors.Wrap(err, "unmarshalling aggregated values")
	}

	return nil
}

func prioUnmarshal(entryMap map[string]envEntry, envMap, cfgMap map[string]string) error {
	prioMap := make(map[envEntry]bool)

	for k, entry := range entryMap {
		if envVal, ok := envMap[k]; ok {
			err := setVal(entry.Value, envVal)
			if err != nil {
				return errors.Wrapf(err, "setting value of env var %s", k)
			}
			prioMap[entry] = true
			continue
		}

		if cfgVal, ok := cfgMap[k]; ok && !prioMap[entry] {
			err := setVal(entry.Value, cfgVal)
			if err != nil {
				return errors.Wrapf(err, "setting value from config file of env var %s", k)
			}
			prioMap[entry] = false
		}
	}

	for k, entry := range entryMap {
		if _, ok := prioMap[entry]; ok {
			continue
		}
		if entry.IsRequired {
			return errs.NewRequiredKeyNotPresentError(k)
		}
		if entry.Default == "" {
			continue
		}
		err := setVal(entry.Value, entry.Default)
		if err != nil {
			return errors.Wrapf(err, "setting default value of env var %s", k)
		}
		prioMap[entry] = true
	}

	return nil
}

func setVal(v reflect.Value, str string) error {
	switch v.Type().Kind() {
	case reflect.Bool:
		b, err := strconv.ParseBool(str)
		if err != nil {
			return errors.Wrap(err, "casting value as bool")
		}
		v.SetBool(b)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return errors.Wrap(err, "casting value as int")
		}
		v.SetInt(i)
		return nil
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return errors.Wrap(err, "casting value as float")
		}
		v.SetFloat(f)
		return nil
	case reflect.String:
		v.SetString(str)
		return nil
	}
	return errs.NewUnsupportedFieldTypeError(v.Type())
}

type envEntry struct {
	Value      reflect.Value
	IsRequired bool
	Default    string
}

func parseStruct(v reflect.Value, connector string) map[string]envEntry {
	t := v.Type()
	n := t.NumField()

	result := make(map[string]envEntry)

	for i := 0; i < n; i++ {
		f := t.Field(i)
		fv := v.Field(i)

		tag := tags.ParseTag(f, TagName)
		isRequired := tag.Options[tagOptionRequired] == "true"
		defaultVal := tag.Options[tagOptionDefault]

		for _, k := range tag.Keys {
			switch f.Type.Kind() {
			case reflect.Struct:
				m := parseStruct(fv, envKeyConnector)
				for subK, subV := range m {
					result[k+connector+subK] = subV
				}
			case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float64, reflect.Float32, reflect.Bool:
				result[k] = envEntry{fv, isRequired, defaultVal}
			}
		}
	}

	return result
}

func parseCfgFiles(paths []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, path := range paths {
		m, err := parseCfgFile(path)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing %s", path)
		}
		for k, v := range m {
			result[k] = v
		}
	}
	return result, nil
}

func parseCfgFile(path string) (map[string]string, error) {
	nestedMap, err := readCfgFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "reading env config data from file")
	}
	flatMap := flattenMap(nestedMap, envKeyConnector)
	return flatMap, nil
}

func flattenMap(nestedMap map[string]interface{}, connector string) map[string]string {
	flatMap := make(map[string]string)

	for k, v := range nestedMap {
		if m, ok := v.(map[string]interface{}); ok {
			flatSubmap := flattenMap(m, connector)
			for subK, subV := range flatSubmap {
				flatMap[k+connector+subK] = subV
			}
			continue
		}
		if _, ok := v.([]interface{}); ok {
			continue
		}
		if b, ok := v.(bool); ok {
			flatMap[k] = strconv.FormatBool(b)
			continue
		}
		if i, ok := v.(int64); ok {
			flatMap[k] = strconv.FormatInt(i, 10)
			continue
		}
		if f, ok := v.(float64); ok {
			flatMap[k] = fmt.Sprintf("%g", f)
			continue
		}
		if s, ok := v.(string); ok {
			flatMap[k] = s
			continue
		}
	}

	return flatMap
}

func readCfgFile(path string) (map[string]interface{}, error) {
	jsonBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "opening %s", path)
	}

	m := make(map[string]interface{})

	err = json.Unmarshal(jsonBytes, &m)
	if err != nil {
		return nil, errors.Wrapf(err, "unmarshalling json data from %s", path)
	}

	return m, nil
}

func environToMap(environ []string) map[string]string {
	m := make(map[string]string)
	for _, v := range environ {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) < 2 {
			continue
		}
		if parts[0] == "" {
			continue
		}
		m[parts[0]] = parts[1]
	}
	return m
}
