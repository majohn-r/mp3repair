package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/x/cmd"
	"gopkg.in/yaml.v3"
)

const (
	coverageFile  = "coverage.out"
	buildDataFile = "buildData.yaml"
)

var (
	build = goyek.Define(goyek.Task{
		Name:  "build",
		Usage: "build the executable",
		Action: func(a *goyek.A) {
			exec, path, flags := readConfig()
			// logged output shows up when running verbose (-v) or on error
			a.Logf("executable: %s\n", exec)
			a.Logf("path: %s\n", path)
			for _, flag := range flags {
				a.Logf("\t%q\n", flag)
			}
			l := fmt.Sprintf("go build -ldflags %q -o %s %s", strings.Join(flags, " "), exec, path)
			o := &bytes.Buffer{}
			cmd.Exec(a, l, options(o)...)
			print(o)
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "clean",
		Usage: "delete build products",
		Action: func(a *goyek.A) {
			exec, _, _ := readConfig()
			os.Remove(filepath.Join("..", coverageFile))
			os.Remove(filepath.Join("..", exec))
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
			if dirs, err := codeDirs(); err == nil {
				o := &bytes.Buffer{}
				for _, f := range dirs {
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

type Config struct {
	// special functions:
	//   application: the application name
	//   path:        the relative path to the main package
	//   timestamp:   flag gets a timestamp value
	Function string
	// flag name
	Flag string
	// value to use
	Value string
}

func readConfig() (exec, path string, flags []string) {
	rawYaml, _ := os.ReadFile(buildDataFile)
	// data := map[string]any{}
	var data []Config
	_ = yaml.Unmarshal(rawYaml, &data)
	for _, value := range data {
		switch value.Function {
		case "application":
			exec = value.Value
		case "path":
			path = value.Value
		case "timestamp":
			flags = append(flags, fmt.Sprintf("-X %s=%s", value.Flag, time.Now().Format(time.RFC3339)))
		default:
			flags = append(flags, fmt.Sprintf("-X %s=%s", value.Flag, value.Value))
		}
	}
	sort.Strings(flags)
	return
}

func allDirs(top string) (dirs []string, err error) {
	var entries []fs.DirEntry
	if entries, err = os.ReadDir(top); err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		f := filepath.Join(top, entry.Name())
		dirs = append(dirs, f)
		var subDirs []string
		if subDirs, err = allDirs(f); err != nil {
			return
		}
		dirs = append(dirs, subDirs...)
	}
	return
}

func codeDirs() (codeDirectories []string, err error) {
	var dirs []string
	if dirs, err = allDirs(".."); err != nil {
		return
	}
	for _, f := range dirs {
		var e []fs.DirEntry
		if e, err = os.ReadDir(f); err != nil {
			return
		}
		if includesCode(e) {
			codeDirectories = append(codeDirectories, f)
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
