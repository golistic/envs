// Copyright (c) 2022, Geert JM Vanderkelen

package envs

import "time"

type testEnv struct {
	PtrNumberInt      *int64         `envVar:"PTR_NUMBER"`
	Number            int            `envVar:"NUMBER" default:"999"`
	PtrNumberIntNaked *int8          `envVar:"PTR_NUMBER_naked"`
	UnquotedString    string         `envVar:"STRING"`
	Enabled           bool           `envVar:"ENABLED"`
	PtrString         *string        `envVar:"PTR_STRING"`
	Single            string         `envVar:"SINGLE_QUOTED"`
	Double            string         `envVar:"DOUBLE_QUOTED"`
	Backquoted        string         `envVar:"BACKQUOTED"`
	Empty             string         `envVar:"EMPTY" default:"not empty"`
	Duration          time.Duration  `envVar:"Duration"`
	PtrDuration       *time.Duration `envVar:"PtrDuration"`
	PtrDurationNaked  *time.Duration `envVar:"PtrDuration_naked"`
	InlineComment     string         `envVar:"INLINE_COMMENT"`
	MultilineDouble   string         `envVar:"MULTI_DOUBLE_QUOTED"`
	MultilineSingle   string         `envVar:"MULTI_SINGLE_QUOTED"`
	MultilineBack     string         `envVar:"MULTI_BACKTICKED"`
	Boolean           bool           `envVar:"BOOLEAN"`
	PtrBoolean        *bool          `envVar:"PTR_BOOLEAN"`
	PtrBooleanNaked   *bool          `envVar:"PTR_BOOLEAN_naked"`
}
