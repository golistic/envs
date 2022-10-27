// Copyright (c) 2022, Geert JM Vanderkelen

package envs

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	tagEnvVar  = "envVar"
	tagDefault = "default"
)

var trues = map[string]struct{}{
	"t":       {},
	"true":    {},
	"1":       {},
	"on":      {},
	"enabled": {},
}

var falses = map[string]struct{}{
	"f":        {},
	"false":    {},
	"0":        {},
	"off":      {},
	"disabled": {},
}

// reflectMapToStruct will go through src and set each field of dest based
// on the tag `envVar`.
//
// Values are trimmed of any surrounding spaces before they are unquoted.
//
// If src does not contain the field's envVar-tag, it will use
// the value of the default-tag. If not, the empty value is considered.
func reflectMapToStruct(src envVarMap, dest any) error {
	rv := reflect.Indirect(reflect.ValueOf(dest))
	rt := rv.Type()
	if rt.Kind() != reflect.Struct {
		panic(fmt.Sprintf("dest must be non-nil struct (was %s)", rt.String()))
	}

	for i := 0; i < rt.NumField(); i++ {
		rtf := rt.Field(i)

		envVar := rtf.Tag.Get(tagEnvVar)
		if envVar == "" {
			continue
		}

		envVarValue, have := src[envVar]
		d := rtf.Tag.Get(tagDefault)
		if !have && d != "" {
			envVarValue = &d
		}

		if envVarValue != nil {
			*envVarValue = strings.TrimSpace(*envVarValue)
		}

		switch t := rv.Field(i).Interface().(type) {
		case time.Duration, *time.Duration:
			if err := handleTimeDuration(envVar, rtf, rv.Field(i), envVarValue); err != nil {
				return err
			}
		case string, *string:
			if err := handleString(envVar, rtf, rv.Field(i), envVarValue); err != nil {
				return err
			}
		case bool, *bool:
			if err := handleBoolean(envVar, rtf, rv.Field(i), envVarValue); err != nil {
				return err
			}
		case int, int8, int16, int32, int64, *int, *int8, *int16, *int32, *int64:
			err := handleNumeric(envVar, rtf, rv.Field(i), envVarValue)
			if err != nil {
				return err
			}
		default:
			panic(fmt.Sprintf("unsupported type '%T' for field %s", t, rtf.Name))
		}
	}

	return nil
}

func handleString(name string, field reflect.StructField, fieldValue reflect.Value, value *string) error {
	if value == nil {
		if field.Type.Kind() == reflect.Pointer {
			var v *string
			fieldValue.Set(reflect.ValueOf(v))
		} else {
			fieldValue.Set(reflect.Zero(fieldValue.Type()))
		}
		return nil
	}
	v := *value

	if v != "" {
		switch v[0] {
		case '"', '`', '\'':
			if v[0] != v[len(v)-1] {
				return &ErrSyntax{
					EnvVar: name,
					Reason: "missing closing quote",
				}
			}
			v = v[1 : len(v)-1]
		}
	}

	if field.Type.Kind() == reflect.Pointer {
		fieldValue.Set(reflect.ValueOf(v))
	} else {
		fieldValue.SetString(v)
	}
	return nil
}

// handleTimeDuration takes struct field and its fieldValue and parse the value
func handleTimeDuration(name string, field reflect.StructField, fieldValue reflect.Value, value *string) error {
	if field.Type.Kind() == reflect.Pointer {
		if value == nil {
			var v *time.Duration
			fieldValue.Set(reflect.ValueOf(v))
			return nil
		}
	}

	if value == nil || *value == "" {
		v := time.Duration(0)
		if field.Type.Kind() == reflect.Pointer {
			fieldValue.Set(reflect.ValueOf(&v))
		} else {
			fieldValue.Set(reflect.ValueOf(v))
		}
		return nil
	}

	res, err := time.ParseDuration(*value)
	if err != nil {
		return &ErrSyntax{
			EnvVar: name,
			Reason: "not parsable as Go duration string",
		}
	}

	fieldValue.Set(reflect.ValueOf(res))
	return nil
}

func handleNumeric(name string, field reflect.StructField, value reflect.Value, s *string) error {
	if field.Type.Kind() == reflect.Pointer {
		if s == nil {
			value.Set(reflect.Zero(value.Type()))
			return nil
		}
	}

	var n int64

	if s == nil || *s == "" {
		n = 0
	} else {
		var err error
		n, err = strconv.ParseInt(*s, 10, 64)
		if err != nil {
			return &ErrSyntax{
				EnvVar: name,
				Reason: "number not parsable",
			}
		}
	}

	if value.Kind() == reflect.Pointer {
		value.Set(reflect.ValueOf(&n))
	} else {
		value.SetInt(n)
	}
	return nil
}

func handleBoolean(name string, field reflect.StructField, value reflect.Value, s *string) error {
	if field.Type.Kind() == reflect.Pointer {
		if s == nil {
			value.Set(reflect.Zero(value.Type()))
			return nil
		}
	}

	b := false

	if s == nil || *s == "" {
		b = false
	} else {
		v := strings.ToLower(*s)
		if _, ok := trues[v]; ok {
			b = true
		} else if _, ok := falses[v]; ok {
			// explicit so we can error on syntax
			b = false
		} else if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			if n > 0 {
				b = true
			}
		} else {
			return &ErrSyntax{EnvVar: name, Reason: "not a valid boolean value"}
		}
	}

	if value.Kind() == reflect.Pointer {
		value.Set(reflect.ValueOf(&b))
	} else {
		value.SetBool(b)
	}
	return nil
}

func envVarFromStruct(name string, s any) (any, error) {
	rv := reflect.Indirect(reflect.ValueOf(s))
	rt := rv.Type()
	if rt.Kind() != reflect.Struct {
		panic(fmt.Sprintf("dest must be non-nil struct (was %s)", rt.Kind()))
	}

	for i := 0; i < rt.NumField(); i++ {
		rtf := rt.Field(i)

		envVar := rtf.Tag.Get("envVar")

		if envVar == "" {
			continue
		}

		if envVar == name {
			return rv.Field(i).Interface(), nil
		}
	}
	return nil, fmt.Errorf("envVar %s not available in struct", name)
}
