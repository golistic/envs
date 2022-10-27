package envs

import (
	"os"

	"github.com/geertjanvdk/xkit/xutil"
)

type envVarMap map[string]*string

// OSEnviron gets variables from the operating system's environment and
// stores the values in the struct dest.
//
// This function uses Go's os.Environ.
//
// Panics when dest is non-pointer, nil, or not a struct.
func OSEnviron(dest any) error {
	src := envVarMap{}

	for _, s := range os.Environ() {
		for i := 0; i < len(s); i++ {
			if s[i] == '=' {
				src[s[0:i]] = xutil.StringPtr(s[i+1:])
				break
			}
		}
	}

	return reflectMapToStruct(src, dest)
}
