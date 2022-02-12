//go:build mage
// +build mage

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	// "github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
)

const (
	executable string = "mp3.exe"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

// Create the executable image
func Build() error {
	// mg.Deps(InstallDeps)
	fmt.Printf("Creating %s\n", executable)
	cmd := exec.Command("go", "build", "-o", executable, "./cmd/mp3/")
	return cmd.Run()
}

// A custom install step if you need your bin someplace other than go/bin
// func Install() error {
// 	mg.Deps(Build)
// 	fmt.Println("Installing...")
// 	return os.Rename("./MyApp", "/usr/bin/MyApp")
// }

// Manage your deps, or running package managers.
// func InstallDeps() (e error) {
// 	fmt.Println("Installing Deps...")
// 	cmd := exec.Command("go", "get", "github.com/stretchr/piglatin")
// 	return cmd.Run()
// }

// Delete the executable image
func Clean() {
	fmt.Printf("Removing %s\n", executable)
	os.RemoveAll(executable)
}

func Test() {
	stdout := &bytes.Buffer{}
    stderr := &bytes.Buffer{}
	cmd := exec.Command("go", "test", "-cover", "./...")
    cmd.Stderr = stderr
    cmd.Stdout = stdout
	fmt.Println("running all unit tests with code coverage")
	if err := cmd.Run(); err != nil {
		fmt.Println(stderr.String())
	} else {
		fmt.Println(stdout.String())
	}
}