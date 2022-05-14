//go:build mage
// +build mage

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"
	// "github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
)

const (
	executable  string = "mp3.exe"
	versionFile string = "version.txt"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

// Create the executable image
func Build() error {
	// mg.Deps(InstallDeps)
	versionArgument, version := createVersionArgument()
	fmt.Printf("Creating %s version %s\n", executable, version)
	flags := fmt.Sprintf("%s %s", versionArgument, createCreationArgument())
	cmd := exec.Command("go", "build", "-ldflags", flags, "-o", executable, "./cmd/mp3/")
	return cmd.Run()
}

func createCreationArgument() string {
	return fmt.Sprintf("-X main.creation=%s", time.Now().Format(time.RFC3339))
}

func createVersionArgument() (arg string, version string) {
	version = "unknown"
	if content, err := ioutil.ReadFile(versionFile); err != nil {
		fmt.Printf("could not open %s: %v\n", versionFile, err)
	} else {
		// this nicely handles cases where the file contains newlines or extra lines
		var major, minor, patch int
		sContent := string(content)
		if _, err := fmt.Sscanf(sContent, "%d.%d.%d", &major, &minor, &patch); err != nil {
			fmt.Fprintf(os.Stderr, "error reading %s (content = %q): %v\n", versionFile, sContent, err)
		}
		version = fmt.Sprintf("%d.%d.%d", major, minor, patch)
	}
	arg = fmt.Sprintf("-X main.version=%s", version)
	return
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

// Execute all unit tests
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

// Execute all unit tests and generate a code coverage report
func CoverageReport() {
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
