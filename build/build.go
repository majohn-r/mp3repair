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
)

var (
	build = goyek.Define(goyek.Task{
		Name:  "build",
		Usage: "build the executable",
		Action: func(a *goyek.A) {
			vArg, v := versionArgument()
			fmt.Printf("Creating %s version %s\n", executable, v)
			var args []string
			if vArg != "" {
				args = append(args, vArg)
			}
			args = append(args, creationArgument())
			l := fmt.Sprintf("go build -ldflags %q -o %s ./cmd/mp3/", strings.Join(args, " "), executable)
			o := &bytes.Buffer{}
			cmd.Exec(a, l, options(o)...)
			print(o)
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
			o := options(nil)
			l := fmt.Sprintf("go test -coverprofile=%s ./...", coverageFile)
			if cmd.Exec(a, l, o...) {
				l = fmt.Sprintf("go tool cover -html=%s", coverageFile)
				cmd.Exec(a, l, o...)
			}
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "doc",
		Usage: "generate documentation",
		Action: func(a *goyek.A) {
			if folders, err := codeFolders(); err == nil {
				o := &bytes.Buffer{}
				for _, f := range folders {
					f = f[3:]
					if f == "build" {
						continue
					}
					l := fmt.Sprintf("go doc -all ./%s", strings.ReplaceAll(f, "\\", "/"))
					if !cmd.Exec(a, l, options(o)...) {
						break
					}
				}
				print(o)
			}
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "format",
		Usage: "clean up source code formatting",
		Action: func(a *goyek.A) {
			o := &bytes.Buffer{}
			cmd.Exec(a, "gofmt -e -l -s -w .", options(o)...)
			print(o)
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "lint",
		Usage: "run the linter on source code",
		Action: func(a *goyek.A) {
			o := &bytes.Buffer{}
			cmd.Exec(a, "gocritic check -enableAll ./...", options(o)...)
			print(o)
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "tests",
		Usage: "run unit tests",
		Action: func(a *goyek.A) {
			o := &bytes.Buffer{}
			cmd.Exec(a, "go test -cover ./...", options(o)...)
			print(o)
		},
	})
)

func creationArgument() string {
	return fmt.Sprintf("-X main.creation=%s", time.Now().Format(time.RFC3339))
}

func versionArgument() (vArg, v string) {
	v = firstLine("version.txt")
	if v != "" {
		vArg = fmt.Sprintf("-X main.version=%s", v)
	} else {
		v = "unknown"
	}
	return
}

func allFolders(top string) (folders []string, err error) {
	var dirs []fs.DirEntry
	if dirs, err = os.ReadDir(top); err != nil {
		return
	}
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		f := filepath.Join(top, d.Name())
		folders = append(folders, f)
		var subfolders []string
		if subfolders, err = allFolders(f); err != nil {
			return
		}
		folders = append(folders, subfolders...)
	}
	return
}

func codeFolders() (folders []string, err error) {
	var candidates []string
	if candidates, err = allFolders(".."); err != nil {
		return
	}
	for _, f := range candidates {
		var e []fs.DirEntry
		if e, err = os.ReadDir(f); err != nil {
			return
		}
		if includesCode(e) {
			folders = append(folders, f)
		}
	}
	return
}

func includesCode(entries []fs.DirEntry) (ok bool) {
	for _, e := range entries {
		if isCode(e) {
			ok = true
			return
		}
	}
	return
}

func isCode(entry fs.DirEntry) (ok bool) {
	if entry.Type().IsRegular() {
		n := entry.Name()
		ok = strings.HasSuffix(n, ".go") && !strings.HasSuffix(n, "_test.go") && !strings.HasPrefix(n, "testing")
	}
	return
}
func options(b *bytes.Buffer) (o []cmd.Option) {
	o = append(o, cmd.Dir(".."))
	if b != nil {
		o = append(o, cmd.Stderr(b), cmd.Stdout(b))
	}
	return o
}

func print(b *bytes.Buffer) {
	s := b.String()
	if s != "" {
		fmt.Println(s)
	}
}

func firstLine(file string) (line string) {
	if f, err := os.Open(file); err != nil {
		fmt.Fprintf(os.Stderr, "error opening %s: %v", file, err)
	} else {
		defer f.Close()
		s := bufio.NewScanner(f)
		if !s.Scan() {
			fmt.Fprintf(os.Stderr, "%s is empty!\n", file)
		} else {
			line = s.Text()
		}
	}
	return
}
