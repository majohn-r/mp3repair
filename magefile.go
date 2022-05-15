//go:build mage
// +build mage

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"
)

const (
	executable  string = "mp3.exe"
	versionFile string = "version.txt"
)

// var Default = Build

// Create the executable image
func Build() error {
	versionArgument, version := createVersionArgument()
	fmt.Printf("Creating %s version %s\n", executable, version)
	creationTimestamp := createCreationArgument()
	var flags string
	if len(versionArgument) > 0 {
		flags = fmt.Sprintf("%s %s", versionArgument, creationTimestamp)
	} else {
		flags = creationTimestamp
	}
	cmd := exec.Command("go", "build", "-ldflags", flags, "-o", executable, "./cmd/mp3/")
	unifiedOutput := &bytes.Buffer{}
	cmd.Stderr = unifiedOutput
	cmd.Stdout = unifiedOutput
	err := cmd.Run()
	printOutput(unifiedOutput)
	return err
}

func printOutput(b *bytes.Buffer){
	output := b.String()
	if len(output) > 0 {
		fmt.Println(output)
	}
}

func createCreationArgument() string {
	return fmt.Sprintf("-X main.creation=%s", time.Now().Format(time.RFC3339))
}

func createVersionArgument() (arg string, version string) {
	version = readFirstLine()
	if len(version) > 0 {
		arg = fmt.Sprintf("-X main.version=%s", version)
	} else {
		version = "unknown"
	}
	return
}

func readFirstLine() (line string) {
	if file, err := os.Open(versionFile); err != nil {
		fmt.Fprintf(os.Stderr, "error opening %s: %v", versionFile, err)
	} else {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		if !scanner.Scan() {
			fmt.Fprintf(os.Stderr, "%s is empty!\n", versionFile)
		} else {
			line = scanner.Text()
		}
	}
	return
}

// Delete the executable image
func Clean() error {
	fmt.Printf("Removing %s\n", executable)
	return os.RemoveAll(executable)
}

// Execute all unit tests
func Test() error {
	unifiedOutput := &bytes.Buffer{}
	// go test -cover ./...
	cmd := exec.Command("go", "test", "-cover", "./...")
	cmd.Stderr = unifiedOutput
	cmd.Stdout = unifiedOutput
	fmt.Println("running all unit tests with code coverage")
	// ignore error return
	err := cmd.Run()
	printOutput(unifiedOutput)
	return err
}

// Execute all unit tests and generate a code coverage report
func CoverageReport() error {
	unifiedOutput := &bytes.Buffer{}
	// go test -coverprofile=coverage.out ./...
	cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "./...")
	cmd.Stderr = unifiedOutput
	cmd.Stdout = unifiedOutput
	fmt.Println("generating code coverage data")
	err := cmd.Run()
	printOutput(unifiedOutput)
	if err == nil {
		unifiedOutput = &bytes.Buffer{}
		// go tool cover -html=coverage.out
		cmd = exec.Command("go", "tool", "cover", "-html=coverage.out")
		cmd.Stderr = unifiedOutput
		cmd.Stdout = unifiedOutput
		fmt.Println("generating report from code coverage data")
		// ignore error return
		err = cmd.Run()
		printOutput(unifiedOutput)
	}
	return err
}
