package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/x/cmd"
	"github.com/josephspurrier/goversioninfo"
	"gopkg.in/yaml.v3"
)

const (
	coverageFile    = "coverage.out"
	buildDataFile   = "buildData.yaml"
	versionInfoFile = "versionInfo.json"
	resourceFile    = "resource.syso"
)

type productData struct {
	majorLevel  int
	minorLevel  int
	patchLevel  int
	description string
	name        string
	firstYear   int
}

var (
	build = goyek.Define(goyek.Task{
		Name:  "build",
		Usage: "build the executable",
		Action: func(a *goyek.A) {
			exec, path, flags, jsonInput := readConfig()
			// logged output shows up when running verbose (-v) or on error
			a.Logf("json input: %d %d %d %q %q %d\n", jsonInput.majorLevel, jsonInput.minorLevel, jsonInput.patchLevel, jsonInput.description, jsonInput.name, jsonInput.firstYear)
			a.Logf("executable: %s\n", exec)
			a.Logf("path: %s\n", path)
			for _, flag := range flags {
				a.Logf("\t%q\n", flag)
			}
			generateVersionInfo(path, jsonInput)
			generate(a, path)
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
			exec, path, _, _ := readConfig()
			os.Remove(filepath.Join("..", path, versionInfoFile))
			os.Remove(filepath.Join("..", path, resourceFile))
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
	// application-specific field
	Description string
}

func generate(a *goyek.A, path string) {
	b := &bytes.Buffer{}
	var o []cmd.Option
	o = append(o, cmd.Dir(filepath.Join("..", path)), cmd.Stderr(b), cmd.Stdout(b))
	cmd.Exec(a, "go generate", o...)
	print(b)
}

func readConfig() (exec, path string, flags []string, jsonInput *productData) {
	rawYaml, _ := os.ReadFile(buildDataFile)
	jsonInput = &productData{}
	var data []Config
	_ = yaml.Unmarshal(rawYaml, &data)
	for _, value := range data {
		switch value.Function {
		case "application":
			exec = value.Value
			jsonInput.description = value.Description
		case "path":
			path = value.Value
		case "timestamp":
			flags = append(flags, fmt.Sprintf("-X %s=%s", value.Flag, time.Now().Format(time.RFC3339)))
		default:
			flags = append(flags, fmt.Sprintf("-X %s=%s", value.Flag, value.Value))
			switch value.Flag {
			case "main.firstYear":
				if i, err := strconv.Atoi(value.Value); err == nil {
					jsonInput.firstYear = i
				}
			case "main.version":
				var major int
				var minor int
				var patch int
				if count, err := fmt.Sscanf(value.Value, "%d.%d.%d", &major, &minor, &patch); count == 3 && err == nil {
					jsonInput.majorLevel = major
					jsonInput.minorLevel = minor
					jsonInput.patchLevel = patch
				}
			case "main.appName":
				jsonInput.name = value.Value
			}
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

func generateVersionInfo(path string, jsonInput *productData) {
	data := goversioninfo.VersionInfo{}
	data.FixedFileInfo.FileVersion.Major = jsonInput.majorLevel
	data.FixedFileInfo.FileVersion.Minor = jsonInput.minorLevel
	data.FixedFileInfo.FileVersion.Patch = jsonInput.patchLevel
	data.FixedFileInfo.ProductVersion.Major = jsonInput.majorLevel
	data.FixedFileInfo.ProductVersion.Minor = jsonInput.minorLevel
	data.FixedFileInfo.ProductVersion.Patch = jsonInput.patchLevel
	data.FixedFileInfo.FileFlagsMask = "3f"
	data.FixedFileInfo.FileOS = "040004"
	data.FixedFileInfo.FileType = "01"
	data.StringFileInfo.FileDescription = jsonInput.description
	version := fmt.Sprintf("v%d.%d.%d", jsonInput.majorLevel, jsonInput.minorLevel, jsonInput.patchLevel)
	data.StringFileInfo.FileVersion = version
	currentYear := time.Now().Year()
	if currentYear == jsonInput.firstYear {
		data.StringFileInfo.LegalCopyright = fmt.Sprintf("Copyright © %d Marc Johnson", currentYear)
	} else {
		data.StringFileInfo.LegalCopyright = fmt.Sprintf("Copyright © %d-%d Marc Johnson", jsonInput.firstYear, currentYear)
	}
	data.StringFileInfo.ProductName = jsonInput.name
	data.StringFileInfo.ProductVersion = version
	data.VarFileInfo.Translation.LangID = goversioninfo.LngUSEnglish
	data.VarFileInfo.Translation.CharsetID = goversioninfo.CsUnicode
	b, _ := json.Marshal(data)
	fullPath := filepath.Join("..", path, versionInfoFile)
	file, fileErr := os.Create(fullPath)
	if fileErr == nil {
		defer file.Close()
		file.Write(b)
	} else {
		fmt.Printf("Error writing %q! %v\n", versionInfoFile, fileErr)
	}
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
