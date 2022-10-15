// Copyright (c) 2022, Geert JM Vanderkelen

package envs

import (
	"bytes"
	"testing"

	"github.com/geertjanvdk/xkit/xt"
)

// TestDotEnvParse executes common tests for all kinds of .env files.
func TestDotEnvParse(t *testing.T) {
	scanner := &dotEnvScanner{}

	t.Run("spaces before equal sign", func(t *testing.T) {
		r := bytes.NewReader([]byte(`NUMBER  = 123`))
		dest := &testEnv{}
		xt.OK(t, dotEnvToStruct(scanner, dest, r))
		xt.Eq(t, 123, dest.Number)
	})

	t.Run("syntax: number not parsable", func(t *testing.T) {
		r := bytes.NewReader([]byte(`NUMBER=Not a number`))
		dest := &testEnv{}
		err := dotEnvToStruct(scanner, dest, r)
		xt.KO(t, err)
		xt.Eq(t, "line 1: syntax error (number not parsable)", err.Error())
	})

	t.Run("syntax: duration not parsable", func(t *testing.T) {
		r := bytes.NewReader([]byte(`Duration=Not a Duration`))
		dest := &testEnv{}
		err := dotEnvToStruct(scanner, dest, r)
		xt.KO(t, err)
		xt.Eq(t, "line 1: syntax error (not parsable as Go duration string)", err.Error())
	})

	t.Run("syntax: boolean not parsable", func(t *testing.T) {
		r := bytes.NewReader([]byte(`BOOLEAN=Neither true or false`))
		dest := &testEnv{}
		err := dotEnvToStruct(scanner, dest, r)
		xt.KO(t, err)
		xt.Eq(t, "line 1: syntax error (not a valid boolean value)", err.Error())
	})
}
