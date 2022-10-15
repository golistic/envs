// Copyright (c) 2022, Geert JM Vanderkelen

package envs

import "fmt"

type ErrSyntax struct {
	Line   int
	EnvVar string
	Reason string
}

func (err *ErrSyntax) Error() string {
	if err.Line > 0 {
		return fmt.Sprintf("line %d: syntax error (%s)", err.Line, err.Reason)
	}
	return fmt.Sprintf("%s: syntax error (%s)", err.EnvVar, err.Reason)
}

type ErrReadingFile struct {
	FilePath string
	Err      error
}

func (err *ErrReadingFile) Error() string {
	return fmt.Sprintf("error reading %s (%s)", err.FilePath, err.Err)
}
