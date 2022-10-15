// Copyright (c) 2022, Geert JM Vanderkelen

package envs

import (
	"io"
	"os"
)

// DjangoDotEnv reads environment variables from r and stores them in struct
// dest according to the rules defined by the django-dotenv project
// https://github.com/jpadilla/django-dotenv/blob/master/dotenv.py. The variables
// are stored and available within the dest struct.
func DjangoDotEnv(dest any, r io.Reader) error {
	ds := &dotEnvScanner{
		quotes: map[rune]bool{
			'"':  true,
			'\'': true,
		},
		unsupportedQuotes: map[rune]bool{
			'`': true,
		},
		expandNewlines: map[rune]bool{
			'"': true,
		},
		allowNaked: true,
	}

	return dotEnvToStruct(ds, dest, r)
}

// DjangoDotEnvFromFile reads environment variables from a file with path and stores
// them in struct dest. See DjangoDotEnv() for further details.
func DjangoDotEnvFromFile(dest any, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return &ErrReadingFile{FilePath: path, Err: err}
	}

	return DjangoDotEnv(dest, f)
}
