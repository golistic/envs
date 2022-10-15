package envs

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/geertjanvdk/xkit/xt"
)

func TestNodeJSDotEnviron(t *testing.T) {
	dest := &testEnv{}
	xt.OK(t, NodeJSDotEnvFromFile(dest, "_test_data/js.env"))

	t.Run("numeric", func(t *testing.T) {
		xt.Eq(t, 123, dest.Number)
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
			"BACKQUOTED":    "BACKQUOTED=`I am quoted!`",
		}

		for cn, c := range cases {
			t.Run(cn, func(t *testing.T) {
				r := bytes.NewReader([]byte(c))
				dest := testEnv{}
				xt.OK(t, NodeJSDotEnv(&dest, r))
				have, err := envVarFromStruct(cn, dest)
				xt.OK(t, err)
				xt.Eq(t, "I am quoted!", have)
			})
		}
	})

	t.Run("empty value", func(t *testing.T) {
		xt.Eq(t, "", dest.Empty)
	})

	t.Run("multiple lines", func(t *testing.T) {
		exp := `THIS
IS
A
MULTILINE
STRING`

		for _, v := range []string{"MULTI_DOUBLE_QUOTED", "MULTI_SINGLE_QUOTED", "MULTI_BACKTICKED"} {
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
			"backticket no expand": {
				envVar: "MULTI_BACKTICKED",
				have:   "`newline \\n not expanded`",
				exp:    `newline \n not expanded`,
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
				err := NodeJSDotEnv(&dest, r)
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
		err := NodeJSDotEnv(&dest, r)
		xt.OK(t, err)
		have, err := envVarFromStruct("STRING", dest)
		xt.OK(t, err)
		xt.Eq(t, exp, have)
	})

	t.Run("syntax error: missing equal sign", func(t *testing.T) {
		r := bytes.NewReader([]byte(`NUMBER 123`))
		dest := testEnv{}
		err := NodeJSDotEnv(&dest, r)
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
			"naked variable": {
				env: `NUMBER
STRING=foo`,
				expErr: "line 1: syntax error (naked variable)",
			},
		}

		dest := testEnv{}
		for cn, c := range cases {
			t.Run(cn, func(t *testing.T) {
				r := bytes.NewReader([]byte(c.env))
				err := NodeJSDotEnv(&dest, r)
				xt.KO(t, err)
				xt.Eq(t, c.expErr, err.Error())
			})
		}

	})

	t.Run("syntax error: missing closing quote", func(t *testing.T) {
		var cases = map[string]string{
			"double":   `DOUBLE_QUOTED="I should be closed`,
			"single":   `SINGLE_QUOTED='I should be closed`,
			"backtick": `BACKQUOTED=` + "`" + `I should be closed`,
		}

		for cn, c := range cases {
			t.Run(cn, func(t *testing.T) {
				r := bytes.NewReader([]byte(c))
				dest := envQuoted{}
				err := NodeJSDotEnv(&dest, r)
				xt.KO(t, err)
				xt.Eq(t, "line 1: syntax error (missing closing quote)", err.Error())
			})
		}
	})

	t.Run("not readable file", func(t *testing.T) {
		err := NodeJSDotEnvFromFile(nil, path.Join(os.TempDir(), "829d9klwiwe.env"))
		xt.KO(t, err)
		xt.Assert(t, strings.Contains(err.Error(), "no such file or directory"))
	})
}
