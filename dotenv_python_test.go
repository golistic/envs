package envs

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/golistic/xgo/xt"
)

func TestPythonDotEnv(t *testing.T) {
	dest := &testEnv{}
	xt.OK(t, DjangoDotEnvFromFile(dest, "_test_data/py.env"))

	t.Run("numeric", func(t *testing.T) {
		xt.Eq(t, int64(123), dest.Number)
	})

	t.Run("numeric variable not set; naked variable", func(t *testing.T) {
		xt.Eq(t, nil, dest.PtrNumberIntNaked)
	})

	t.Run("numeric variable zero", func(t *testing.T) {
		xt.Eq(t, int64(0), *dest.PtrNumberInt)
	})

	t.Run("unquoted strings are trimmed of whitespaces", func(t *testing.T) {
		xt.Eq(t, "My String", dest.UnquotedString)
	})

	t.Run("inline comments are ignored", func(t *testing.T) {
		xt.Eq(t, "Are ignored", dest.InlineComment)
	})

	t.Run("quoted value", func(t *testing.T) {
		var cases = map[string]string{
			"DOUBLE_QUOTED": `DOUBLE_QUOTED="I am quoted!"`,
			"SINGLE_QUOTED": `SINGLE_QUOTED='I am quoted!'`,
		}

		for cn, c := range cases {
			t.Run(cn, func(t *testing.T) {
				r := bytes.NewReader([]byte(c))
				dest := testEnv{}
				xt.OK(t, DjangoDotEnv(&dest, r))
				have, err := envVarFromStruct(cn, dest)
				xt.OK(t, err)
				xt.Eq(t, "I am quoted!", have)
			})
		}
	})

	t.Run("syntax error: unsupported backtick", func(t *testing.T) {
		r := bytes.NewReader([]byte(fmt.Sprintf("BACKTICK=`backquotes not supported`")))
		dest := testEnv{}
		err := DjangoDotEnv(&dest, r)
		xt.KO(t, err)
		xt.Eq(t, "line 1: syntax error (unsupported quote)", err.Error())
	})

	t.Run("empty value", func(t *testing.T) {
		xt.Eq(t, "", dest.Empty)
	})

	t.Run("boolean empty value", func(t *testing.T) {
		xt.Eq(t, false, *dest.PtrBoolean)
	})

	t.Run("boolean variable not set", func(t *testing.T) {
		xt.Eq(t, nil, dest.PtrBooleanNaked)
	})

	t.Run("boolean variable not set", func(t *testing.T) {
		xt.Eq(t, nil, dest.PtrBooleanNaked)
	})

	t.Run("time.Duration variable not set; naked variable", func(t *testing.T) {
		xt.Eq(t, nil, dest.PtrDurationNaked)
	})

	t.Run("time.Duration variable zero", func(t *testing.T) {
		xt.Eq(t, 0, dest.PtrDuration.Seconds())
	})

	t.Run("multiple lines", func(t *testing.T) {
		exp := `THIS
IS
A
MULTILINE
STRING`

		for _, v := range []string{"MULTI_DOUBLE_QUOTED", "MULTI_SINGLE_QUOTED"} {
			t.Run(v, func(t *testing.T) {
				have, err := envVarFromStruct(v, dest)
				xt.OK(t, err)
				xt.Eq(t, exp, have)
			})
		}
	})

	t.Run("expanded newlines", func(t *testing.T) {
		var cases = map[string]struct {
			envVar string
			have   string
			exp    string
		}{
			"single quote no expand": {
				envVar: "MULTI_SINGLE_QUOTED",
				have:   `'newline \n not expanded'`,
				exp:    `newline \n not expanded`,
			},
			"double quote expand": {
				envVar: "MULTI_DOUBLE_QUOTED",
				have:   `"newline\nexpanded"`,
				exp:    "newline\nexpanded",
			},
			"unescaped newlines": {
				envVar: "MULTI_DOUBLE_QUOTED",
				have: `MULTILINE="newline
expanded"`,
				exp: "newline\nexpanded",
			},
			"Windows newlines expand": {
				envVar: "MULTI_DOUBLE_QUOTED",
				have:   `"newline\r\nexpanded"`,
				exp:    "newline\nexpanded",
			},
		}

		for cn, c := range cases {
			t.Run(cn, func(t *testing.T) {
				r := bytes.NewReader([]byte(fmt.Sprintf("%s=%s", c.envVar, c.have)))
				dest := testEnv{}
				err := DjangoDotEnv(&dest, r)
				xt.OK(t, err)
				have, err := envVarFromStruct(c.envVar, dest)
				xt.OK(t, err)
				xt.Eq(t, c.exp, have)
			})
		}
	})

	t.Run("unquoted string does not expanded new lines", func(t *testing.T) {
		exp := `not\nexpanded`
		r := bytes.NewReader([]byte(`STRING=` + exp))
		dest := testEnv{}
		err := DjangoDotEnv(&dest, r)
		xt.OK(t, err)
		have, err := envVarFromStruct("STRING", dest)
		xt.OK(t, err)
		xt.Eq(t, exp, have)
	})

	t.Run("syntax error: missing equal sign", func(t *testing.T) {
		r := bytes.NewReader([]byte(`NUMBER 123`))
		dest := testEnv{}
		err := DjangoDotEnv(&dest, r)
		xt.KO(t, err)
		xt.Eq(t, "line 1: syntax error (invalid variable name)", err.Error())
	})

	t.Run("syntax error: variable name", func(t *testing.T) {
		var cases = map[string]struct {
			env    string
			expErr string
		}{
			"missing equal with spaces": {
				env:    `NUMBER 123`,
				expErr: "line 1: syntax error (invalid variable name)",
			},
		}

		dest := testEnv{}
		for cn, c := range cases {
			t.Run(cn, func(t *testing.T) {
				r := bytes.NewReader([]byte(c.env))
				err := DjangoDotEnv(&dest, r)
				xt.KO(t, err)
				xt.Eq(t, c.expErr, err.Error())
			})
		}
	})

	t.Run("syntax error: missing closing quote", func(t *testing.T) {
		var cases = map[string]string{
			"double": `DOUBLE_QUOTED="I should be closed`,
			"single": `SINGLE_QUOTED='I should be closed`,
		}

		for cn, c := range cases {
			t.Run(cn, func(t *testing.T) {
				r := bytes.NewReader([]byte(c))
				dest := envQuoted{}
				err := DjangoDotEnv(&dest, r)
				xt.KO(t, err)
				xt.Eq(t, "line 1: syntax error (missing closing quote)", err.Error())
			})
		}
	})

	t.Run("not readable file", func(t *testing.T) {
		err := DjangoDotEnvFromFile(nil, path.Join(os.TempDir(), "vnmjkef8qjfi.env"))
		xt.KO(t, err)
		xt.Assert(t, strings.Contains(err.Error(), "no such file or directory"))
	})
}
