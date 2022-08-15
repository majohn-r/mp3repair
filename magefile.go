//go:build mage
// +build mage

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	executable  = "mp3.exe"
	versionFile = "version.txt"
)

// var Default = Build

// Create the executable image
func Build() (err error) {
	versionArgument, version := createVersionArgument()
	fmt.Printf("Creating %s version %s\n", executable, version)
	var args []string
	if len(versionArgument) > 0 {
		args = append(args, versionArgument)
	}
	args = append(args, createCreationArgument())
	flags := strings.Join(args, " ")
	cmd := exec.Command("go", "build", "-ldflags", flags, "-o", executable, "./cmd/mp3/")
	unifiedOutput := &bytes.Buffer{}
	cmd.Stderr = unifiedOutput
	cmd.Stdout = unifiedOutput
	err = cmd.Run()
	printOutput(unifiedOutput)
	return
}

func printOutput(b *bytes.Buffer) {
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
func Test() (err error) {
	unifiedOutput := &bytes.Buffer{}
	// go test -cover ./...
	cmd := exec.Command("go", "test", "-cover", "./...")
	cmd.Stderr = unifiedOutput
	cmd.Stdout = unifiedOutput
	fmt.Println("running all unit tests with code coverage")
	err = cmd.Run()
	printOutput(unifiedOutput)
	return
}

// Execute all unit tests and generate a code coverage report
func CoverageReport() (err error) {
	// go test -coverprofile=coverage.out ./...
	cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "./...")
	fmt.Println("generating code coverage data")
	err = cmd.Run()
	if err == nil {
		// go tool cover -html=coverage.out
		cmd = exec.Command("go", "tool", "cover", "-html=coverage.out")
		fmt.Println("generating report from code coverage data")
		// ignore error return
		err = cmd.Run()
	}
	return err
}

// Generate go doc output
func Doc() (err error) {
	var folders []string
	if folders, err = getCodeFolders(); err == nil {
		unifiedOutput := &bytes.Buffer{}
		for _, folder := range folders {
			// go doc -all .\\{folder}
			if !strings.HasPrefix(folder, ".") {
				folder = ".\\" + folder
			}
			cmd := exec.Command("go", "doc", "-all", folder)
			cmd.Stderr = unifiedOutput
			cmd.Stdout = unifiedOutput
			if err = cmd.Run(); err != nil {
				break
			}
		}
		printOutput(unifiedOutput)
	}
	return
}

func getCodeFolders() (folders []string, err error) {
	var candidates []string
	if candidates, err = getAllFolders("."); err != nil {
		return
	}
	for _, folder := range candidates {
		var entries []fs.DirEntry
		if entries, err = os.ReadDir(folder); err != nil {
			return
		}
		if includesCode(entries) {
			folders = append(folders, folder)
		}
	}
	return
}

func includesCode(entries []fs.DirEntry) (ok bool) {
	for _, entry := range entries {
		if isCode(entry) {
			ok = true
			return
		}
	}
	return
}

func isCode(entry fs.DirEntry) (ok bool) {
	if entry.Type().IsRegular() {
		name := entry.Name()
		ok = strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") && !strings.HasPrefix(name, "testing")
	}
	return
}

func getAllFolders(topDir string) (folders []string, err error) {
	var dirs []fs.DirEntry
	if dirs, err = os.ReadDir(topDir); err != nil {
		return
	}
	for _, d := range dirs {
		if d.IsDir() {
			folder := filepath.Join(topDir, d.Name())
			folders = append(folders, folder)
			var subfolders []string
			if subfolders, err = getAllFolders(folder); err != nil {
				return
			}
			folders = append(folders, subfolders...)
		}
	}
	return
}
