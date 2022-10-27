// Copyright (c) 2022, Geert JM Vanderkelen

package envs_test

import (
	"fmt"
	"os"
	"path"

	"github.com/golistic/envs"
)

type UserEnv struct {
	Username string `envVar:"USER"`
	HomeDir  string `envVar:"HOME"`
	Avatar   string `envVar:"AVATAR" default:"🐣"`
}

func ExampleOSEnviron() {
	userEnv := &UserEnv{}

	// we have to set the following so test is deterministic
	_ = os.Setenv("USER", "alice")
	_ = os.Setenv("HOME", "/Users/alice")

	if err := envs.OSEnviron(userEnv); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Username: %s\n", userEnv.Username)
		fmt.Printf("HomeDir : %s\n", userEnv.HomeDir)
		fmt.Printf("Avatar  : %s\n", userEnv.Avatar)
	}

	// Output:
	// Username: alice
	// HomeDir : /Users/alice
	// Avatar  : 🐣
}

func ExampleNodeJSDotEnvFromFile() {
	userEnv := &UserEnv{}

	p := path.Join(".", "_test_data", "js.example.env")
	if err := envs.NodeJSDotEnvFromFile(userEnv, p); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Username: %s\n", userEnv.Username)
		fmt.Printf("HomeDir : %s\n", userEnv.HomeDir)
		fmt.Printf("Avatar  : %s\n", userEnv.Avatar)
	}

	// Output:
	// Username: alice
	// HomeDir : /home/alice
	// Avatar  : 🙂️
}
