package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/x/cmd"
)

const (
	coverageFile = "coverage.out"
	executable   = "mp3.exe"
	versionFile  = "version.txt"
)

var (
	build = goyek.Define(goyek.Task{
		Name:  "build",
		Usage: "build the executable",
		Action: func(a *goyek.A) {
			versionArgument, version := createVersionArgument()
			fmt.Printf("Creating %s version %s\n", executable, version)
			var args []string
			if versionArgument != "" {
				args = append(args, versionArgument)
			}
			args = append(args, createCreationArgument())
			cmdLine := fmt.Sprintf("go build -ldflags %q -o %s ./cmd/mp3/", strings.Join(args, " "), executable)
			unifiedOutput := &bytes.Buffer{}
			cmd.Exec(a, cmdLine, makeOptions(unifiedOutput)...)
			printOutput(unifiedOutput)
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "clean",
		Usage: "delete build products",
		Action: func(a *goyek.A) {
			os.Remove(filepath.Join("..", coverageFile))
			os.Remove(filepath.Join("..", executable))
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "coverage",
		Usage: "run unit tests and produce a coverage report",
		Action: func(a *goyek.A) {
			o := makeOptions(nil)
			cmdLine := fmt.Sprintf("go test -coverprofile=%s ./...", coverageFile)
			if cmd.Exec(a, cmdLine, o...) {
				cmdLine = fmt.Sprintf("go tool cover -html=%s", coverageFile)
				cmd.Exec(a, cmdLine, o...)
			}
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "doc",
		Usage: "generate documentation",
		Action: func(a *goyek.A) {
			if folders, err := getCodeFolders(); err == nil {
				unifiedOutput := &bytes.Buffer{}
				for _, folder := range folders {
					folder = folder[3:]
					if folder == "build" {
						continue
					}
					cmdLine := fmt.Sprintf("go doc -all ./%s", strings.ReplaceAll(folder, "\\", "/"))
					if !cmd.Exec(a, cmdLine, makeOptions(unifiedOutput)...) {
						break
					}
				}
				printOutput(unifiedOutput)
			}
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "format",
		Usage: "clean up source code formatting",
		Action: func(a *goyek.A) {
			unifiedOutput := &bytes.Buffer{}
			cmd.Exec(a, "gofmt -e -l -s -w .", makeOptions(unifiedOutput)...)
			printOutput(unifiedOutput)
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "lint",
		Usage: "run the linter on source code",
		Action: func(a *goyek.A) {
			unifiedOutput := &bytes.Buffer{}
			cmd.Exec(a, "gocritic check -enableAll ./...", makeOptions(unifiedOutput)...)
			printOutput(unifiedOutput)
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "tests",
		Usage: "run unit tests",
		Action: func(a *goyek.A) {
			unifiedOutput := &bytes.Buffer{}
			cmd.Exec(a, "go test -cover ./...", makeOptions(unifiedOutput)...)
			printOutput(unifiedOutput)
		},
	})
)

func createCreationArgument() string {
	return fmt.Sprintf("-X main.creation=%s", time.Now().Format(time.RFC3339))
}

func createVersionArgument() (arg, version string) {
	version = readFirstLine()
	if version != "" {
		arg = fmt.Sprintf("-X main.version=%s", version)
	} else {
		version = "unknown"
	}
	return
}

func getAllFolders(topDir string) (folders []string, err error) {
	var dirs []fs.DirEntry
	if dirs, err = os.ReadDir(topDir); err != nil {
		return
	}
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		folder := filepath.Join(topDir, d.Name())
		folders = append(folders, folder)
		var subfolders []string
		if subfolders, err = getAllFolders(folder); err != nil {
			return
		}
		folders = append(folders, subfolders...)
	}
	return
}

func getCodeFolders() (folders []string, err error) {
	var candidates []string
	if candidates, err = getAllFolders(".."); err != nil {
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
func makeOptions(b *bytes.Buffer) []cmd.Option {
	var outputOptions []cmd.Option
	outputOptions = append(outputOptions, cmd.Dir(".."))
	if b != nil {
		outputOptions = append(outputOptions, cmd.Stderr(b), cmd.Stdout(b))
	}
	return outputOptions
}

func printOutput(b *bytes.Buffer) {
	output := b.String()
	if output != "" {
		fmt.Println(output)
	}
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
