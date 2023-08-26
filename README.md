envs - Reads configuration from various environments
====================================================

Copyright (c) 2022, 2023, Geert JM Vanderkelen

The Go envs package offers functionality to read environment variables
from the OS or dot-env files.

Overview
--------

Environment Variables can be used, among other things, to configure
applications. This package offers functionality to read such environment
into a Go `struct` with special field tags.

Quick Start
-----------

For example, you want to read variables from the environment where the Go
application runs to identify the user and her home directory. This can be
easily achieved using Go's `os.Getenv`, but using this package, it would
be like this:

```go
package main

import (
	"fmt"

	"github.com/golistic/envs"
)

type UserEnv struct {
	Username string `envVar:"USER"`
	HomeDir  string `envVar:"HOME"`
	Avatar   string `envVar:"AVATAR" default:"🐣"`
}

func main() {
	userEnv := &UserEnv{}
	if err := envs.OSEnviron(userEnv); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Username: %s\n", userEnv.Username)
		fmt.Printf("HomeDir : %s\n", userEnv.HomeDir)
	}
}
```

The above would output something similar too:

```
Username: alice
HomeDir : /Users/alice
Avatar  : 🐣
```

The `AVATAR` environment variables was not available, so the default was used.

Supported Go Types
------------------

We currently support the following types (also pointer values):

* string
* numeric
  - `int`, `int8`, `int16`, `int32`, `int64`
  - empty mean 0 (zero)
* bool
  - `true`, `t`, `1`, `on`, `enabled`
  - `false`, `f`, `0`, `off`, `disabled`
  - empty means `false`
* time.Duration
  - a Go duration as string, for example, `2d5m`
  - empty means `0s`

### Naked Variables

Naked variables are those without value and equal sign, for example:

```
# naked!
TCP_PORT
```

When the destination struct has the field reading `TCP_PORT` as pointer value,
for example, `*int16`, it will be `nil`. However, if type would be `int16`, a
syntax error is shown.


Supported Environments
----------------------

We support the following environments:

* Operating System (OS) environment using Go's `os.Environ`
* `.env` files using rules from
    - the [dotenv][10] project, for NodeJS projects
    - the [djanto-dotenv][11] project, for Django projects

### Operating System (OS) environment

Reading the Operating System (OS) environment is the most common way of getting
the application's configuration.

See [Quick Start](#quick-start) for an example.

### NodeJS projects

Reading an `.env` (dotenv) file from a NodeJS project is done using the rules
defined by the [dotenv][10] package.

Note: variable expansion not yet supported.

Example:

```go
package main

import (
	"fmt"

	"github.com/golistic/envs"
)

type UserEnv struct {
	Username string `envVar:"USER"`
	HomeDir  string `envVar:"HOME"`
	Avatar   string `envVar:"AVATAR" default:"🐣"`
}

func main() {
	userEnv := &UserEnv{}
	if err := envs.NodeJSDotEnvFromFile(userEnv, ".env"); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Username: %s\n", userEnv.Username)
		fmt.Printf("HomeDir : %s\n", userEnv.HomeDir)
		fmt.Printf("Avatar  : %s\n", userEnv.Avatar)
	}
}
```

### Django projects

Reading an `.env` (dotenv) file from a Django project is done using the rules
defined by the [django-dotenv][11] module.

Note: variable expansion not yet supported.

Example code is very similar to the [NodeJS](#nodejs-projects) one, but using
the function `envs.DjangoDotEnvFromFile` instead.


License
-------

Distributed under the MIT license. See LICENSE.txt for more information.


[10]: https://github.com/motdotla/dotenv

[11]: https://github.com/jpadilla/django-dotenv/blob/master/dotenv.py