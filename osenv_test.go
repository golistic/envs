// Copyright (c) 2022, Geert JM Vanderkelen

package envs

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/geertjanvdk/xkit/xt"
)

func TestOSEnviron(t *testing.T) {
	t.Run("string variable default", func(t *testing.T) {
		env := struct {
			Name string `envVar:"NAME" default:"Alice"`
		}{}
		xt.OK(t, OSEnviron(&env))
		xt.Eq(t, "Alice", env.Name)
	})

	t.Run("syntax error: missing closing quote", func(t *testing.T) {
		key := "STRING_fiw2eodc"
		exp := "missing quote at end"

		env := struct {
			Name string `envVar:"STRING_fiw2eodc"`
		}{}

		for _, q := range []string{"'", `"`, "`"} {
			t.Run(q, func(t *testing.T) {
				xt.OK(t, os.Setenv(key, q+exp))
				err := OSEnviron(&env)
				xt.KO(t, err)
				xt.Eq(t, "STRING_fiw2eodc: syntax error (missing closing quote)", err.Error())
			})
		}
	})

	t.Run("string variable set in environment", func(t *testing.T) {
		key := "NAME_diwk3849sk"
		exp := "Bob"
		xt.OK(t, os.Setenv(key, exp))
		env := struct {
			Name string `envVar:"NAME_diwk3849sk" default:"Alice"`
		}{}
		xt.OK(t, OSEnviron(&env))
		xt.Eq(t, exp, env.Name)
	})

	t.Run("numeric variable default", func(t *testing.T) {
		env := struct {
			Number int `envVar:"NUMBER" default:"123"`
		}{}
		xt.OK(t, OSEnviron(&env))
		xt.Eq(t, 123, env.Number)
	})

	t.Run("syntax: number not parsable", func(t *testing.T) {
		env := struct {
			Number int `envVar:"NUMBER" default:"123"`
		}{}
		xt.OK(t, os.Setenv("NUMBER", "Not a number"))
		err := OSEnviron(&env)
		xt.KO(t, err)
		xt.Eq(t, "NUMBER: syntax error (number not parsable)", err.Error())
	})

	t.Run("int variable set in environment", func(t *testing.T) {
		var cases = []struct {
			field string // field name from envNumbers struct
			exp   any
		}{
			{
				field: "Number",
				exp:   546,
			},
			{
				field: "Number64",
				exp:   int64(5466464646464),
			},
			{
				field: "Number8",
				exp:   int8(0),
			},
			{
				field: "Number16",
				exp:   int16(1616),
			},
			{
				field: "Number32",
				exp:   int32(-323232),
			},
			{
				field: "Number64",
				exp:   int64(-64646464646464),
			},
		}

		env := envNumbers{}

		for _, c := range cases {
			envKey := strings.ToUpper(c.field)
			t.Run(fmt.Sprintf("%s value %v", envKey, c.exp), func(t *testing.T) {
				xt.OK(t, os.Setenv(envKey, fmt.Sprintf("%d", c.exp)))
				xt.OK(t, OSEnviron(&env))
				rv := reflect.ValueOf(env)
				rf := rv.FieldByName(c.field)
				if rf.IsZero() {
					xt.Eq(t, c.exp, 0)
				} else {
					xt.Eq(t, fmt.Sprintf("%d", c.exp), fmt.Sprintf("%d", rf.Int()))
				}
				xt.OK(t, os.Unsetenv(envKey))
			})

		}
	})

	t.Run("boolean", func(t *testing.T) {
		var casesFalse = []string{"false", "False", "FALSE", "0", "off", "f", ""}
		var casesTrue = []string{"true", "True", "TRUE", "1", "on", "12345", "t"}

		var testEnv = struct {
			Boolean    bool  `envVar:"BOOL"`
			PtrBoolean *bool `envVar:"PTR_BOOL" default:"true"`
		}{}

		for _, c := range casesFalse {
			t.Run(fmt.Sprintf("%s is false", c), func(t *testing.T) {
				testEnv.Boolean = true
				testEnv.PtrBoolean = &testEnv.Boolean
				xt.OK(t, os.Setenv("BOOL", c))
				xt.OK(t, os.Setenv("PTR_BOOL", c))
				xt.OK(t, OSEnviron(&testEnv))
				xt.Eq(t, false, testEnv.Boolean)
				xt.Eq(t, false, *testEnv.PtrBoolean)
			})
		}

		for _, c := range casesTrue {
			t.Run(fmt.Sprintf("%s is true", c), func(t *testing.T) {
				testEnv.Boolean = false
				testEnv.PtrBoolean = &testEnv.Boolean
				xt.OK(t, os.Setenv("BOOL", c))
				xt.OK(t, os.Setenv("PTR_BOOL", c))
				xt.OK(t, OSEnviron(&testEnv))
				xt.Eq(t, true, testEnv.Boolean)
				xt.Eq(t, true, *testEnv.PtrBoolean)
			})
		}
	})

	t.Run("syntax: boolean not parsable", func(t *testing.T) {
		var env = struct {
			Boolean bool `envVar:"BOOL"`
		}{}
		xt.OK(t, os.Setenv("BOOL", "Neither true or false"))
		err := OSEnviron(&env)
		xt.KO(t, err)
		xt.Eq(t, "BOOL: syntax error (not a valid boolean value)", err.Error())
	})

	t.Run("panic: unsupported type", func(t *testing.T) {
		var env = struct {
			Something io.Writer `envVar:"SOMETHING"`
		}{}
		xt.OK(t, os.Setenv("SOMETHING", "goes into struct field of unsupported type"))
		xt.Panics(t, func() {
			_ = OSEnviron(&env)
		})
	})

	t.Run("panic: destination not valid", func(t *testing.T) {
		t.Run("must be pointer to struct", func(t *testing.T) {
			foo := 1
			xt.Panics(t, func() {
				_ = OSEnviron(&foo)
			})
		})

		t.Run("must be initialized (not nil)", func(t *testing.T) {
			xt.Panics(t, func() {
				var env *testEnv
				_ = OSEnviron(&env)
			})
		})

		t.Run("cannot be nil", func(t *testing.T) {
			xt.Panics(t, func() {
				_ = OSEnviron(nil)
			})
		})
	})
}

type envNumbers struct {
	Number   int   `envVar:"NUMBER" default:"999"`
	Number8  int8  `envVar:"NUMBER8"`
	Number16 int16 `envVar:"NUMBER16"`
	Number32 int32 `envVar:"NUMBER32"`
	Number64 int64 `envVar:"NUMBER64" default:"999"`
}

type envQuoted struct {
	Double   string `envVar:"DOUBLE_QUOTED"`
	BackTick string `envVar:"BACKQUOTED"`
}
