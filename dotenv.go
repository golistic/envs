// Copyright (c) 2022, Geert JM Vanderkelen

package envs

import (
	"bufio"
	"io"
	"regexp"
	"text/scanner"
)

var reUnescapeNewLines = regexp.MustCompile(`(\\r)?\\n`)

type dotEnvScanner struct {
	src     *bufio.Reader
	ch      rune
	vars    envVarMap
	line    int
	lastErr error

	allowNaked        bool // variables without value and =-sign
	quotes            map[rune]bool
	unsupportedQuotes map[rune]bool
	expandNewlines    map[rune]bool
}

func (ds *dotEnvScanner) next() bool {
	r, _, err := ds.src.ReadRune()
	if err != nil {
		if err != io.EOF {
			ds.lastErr = err
		} else {
			ds.ch = scanner.EOF
		}
		return false
	}
	ds.ch = r

	if ds.ch == '\n' {
		ds.line++
	}

	return true
}

// parse will take a reader r and parse the variables with their values,
// storing them in a map.
// This is not going to win any performance contest, but usually the dot-env files
// are small and read once (or when changed).
func (ds *dotEnvScanner) parse(r io.Reader) error {
	ds.src = bufio.NewReader(r)
	ds.line = 1
	ds.vars = envVarMap{}

	for ds.next() {
		switch ds.ch {
		case ' ', '\r', '\n', '\t':
			continue
		case '#':
			ds.consumeRestLine()
			continue
		default:
			variable, naked, err := ds.handleName()
			if err != nil {
				return err
			}

			if !naked {
				value, err := ds.handleValue()
				if err != nil {
					return err
				}
				ds.vars[variable] = &value
			} else {
				ds.vars[variable] = nil
			}
		}

		if ds.lastErr != nil {
			return ds.lastErr
		}
	}

	return nil
}

func (ds *dotEnvScanner) consumeRestLine() {
	for ds.next() {
		if ds.ch == '\n' {
			break
		}
	}
}

func (ds *dotEnvScanner) handleName() (string, bool, error) {
	variable := string(ds.ch)

next:
	for ds.next() {
		switch ds.ch {
		case '\r', '\n':
			if !ds.allowNaked {
				return "", false, &ErrSyntax{Line: ds.line - 1, Reason: "naked variable"}
			}
			return variable, true, nil
		case ' ':
			for ds.next() {
				if ds.ch == '=' {
					break
				}
			}
			if ds.ch != '=' {
				return "", false, &ErrSyntax{Line: ds.line, Reason: "invalid variable name"}
			}
			break next
		case '=':
			break next
		default:
			variable += string(ds.ch)
		}
	}

	return variable, false, nil
}

func (ds *dotEnvScanner) handleValue() (string, error) {
	var value string
next:
	for ds.next() {
		switch {
		case ds.quotes[ds.ch]:
			q := ds.ch
			var err error
			if value, err = ds.handleQuotedValue(); err != nil {
				ds.lastErr = err
			} else if ds.expandNewlines[q] {
				value = reUnescapeNewLines.ReplaceAllString(value, "\n")
			}
			break next
		case ds.ch == '#':
			ds.consumeRestLine()
			break next
		case ds.ch == '\n':
			break next
		case ds.unsupportedQuotes[ds.ch]:
			return "", &ErrSyntax{Line: ds.line, Reason: "unsupported quote"}
		default:
			value += string(ds.ch)
		}
	}

	return value, nil
}

func (ds *dotEnvScanner) handleQuotedValue() (string, error) {
	quote := ds.ch
	value := string(ds.ch)
	for ds.next() {
		value += string(ds.ch)
		if ds.ch == quote {
			return value, nil
		}
	}

	return "", &ErrSyntax{Line: ds.line, Reason: "missing closing quote"}
}

func dotEnvToStruct(s *dotEnvScanner, dest any, r io.Reader) error {
	if err := s.parse(r); err != nil {
		return err
	}

	if err := reflectMapToStruct(s.vars, dest); err != nil {
		if e, ok := err.(*ErrSyntax); ok {
			e.Line = s.line
			return e
		}
		return err
	}
	return nil
}
