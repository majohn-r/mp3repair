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
	unifiedOutput := &bytes.Buffer{}
	cmd := exec.Command("go", "test", "-cover", "./...")
	cmd.Stderr = unifiedOutput
	cmd.Stdout = unifiedOutput
	fmt.Println("running all unit tests with code coverage")
	// ignore error return
	_ = cmd.Run()
	fmt.Println(unifiedOutput.String())
}

func CoverageReport(){
	unifiedOutput := &bytes.Buffer{}
	cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "./...")
	cmd.Stderr = unifiedOutput
	cmd.Stdout = unifiedOutput
	fmt.Println("generating code coverage data")
	// ignore error return
	_ = cmd.Run()
	fmt.Println(unifiedOutput.String())
	unifiedOutput = &bytes.Buffer{}
	// go tool cover -html=coverage.out
	cmd = exec.Command("go", "tool", "cover", "-html=coverage.out")
	cmd.Stderr = unifiedOutput
	cmd.Stdout = unifiedOutput
	fmt.Println("generating report from code coverage data")
	// ignore error return
	_ = cmd.Run()
	fmt.Println(unifiedOutput.String())
}