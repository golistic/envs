// Copyright (c) 2022, Geert JM Vanderkelen

package envs

import (
	"io"
	"os"
)

// NodeJSDotEnv reads environment variables from a file typically called `.env`
// according to the rules defined by the NPM package https://www.npmjs.com/package/dotenv.
func NodeJSDotEnv(dest any, r io.Reader) error {
	ds := &dotEnvScanner{
		quotes: map[rune]bool{
			'"':  true,
			'`':  true,
			'\'': true,
		},
		expandNewlines: map[rune]bool{
			'"': true,
		},
	}

	return dotEnvToStruct(ds, dest, r)
}

// NodeJSDotEnvFromFile reads environment variables from a file with path and stores
// them in struct dest. See NodeJSDotEnv() for further details.
func NodeJSDotEnvFromFile(dest any, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return &ErrReadingFile{FilePath: path, Err: err}
	}

	return NodeJSDotEnv(dest, f)
}
