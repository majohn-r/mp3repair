package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/goyek/goyek/v2"
	"github.com/josephspurrier/goversioninfo"
	toolsbuild "github.com/majohn-r/tools-build"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

const (
	applicationDescription = "mp3 repair utility"
	applicationName        = "mp3repair"
	// file generated by the coverage task
	coverageFile = "coverage.out"
	// files generated by the build task
	versionInfoFile = "versionInfo.json"
	resourceFile    = "resource.syso"
	executableName  = applicationName + ".exe"
)

var (
	generatedFiles = []string{
		coverageFile,
		versionInfoFile,
		resourceFile,
		executableName,
	}
	fileSystem = afero.NewOsFs()
	pD         *productData
	build      = goyek.Define(goyek.Task{
		Name:  "build",
		Usage: "build the executable",
		Action: func(a *goyek.A) {
			buildExecutable(a)
		},
	})

	clean = goyek.Define(goyek.Task{
		Name:  "clean",
		Usage: "delete build products",
		Action: func(a *goyek.A) {
			fmt.Println("deleting build products")
			toolsbuild.Clean(generatedFiles)
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "coverage",
		Usage: "run unit tests and produce a coverage report",
		Action: func(a *goyek.A) {
			toolsbuild.GenerateCoverageReport(a, coverageFile)
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "deadcode",
		Usage: "run deadcode analysis",
		Action: func(a *goyek.A) {
			toolsbuild.Deadcode(a)
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "doc",
		Usage: "generate documentation",
		Action: func(a *goyek.A) {
			toolsbuild.GenerateDocumentation(a, []string{"build", ".idea"})
		},
	})

	format = goyek.Define(goyek.Task{
		Name:  "format",
		Usage: "clean up source code formatting",
		Action: func(a *goyek.A) {
			toolsbuild.FormatSelective(a, []string{".idea"})
		},
	})

	lint = goyek.Define(goyek.Task{
		Name:  "lint",
		Usage: "run the linter on source code",
		Action: func(a *goyek.A) {
			toolsbuild.Lint(a)
		},
	})

	nilaway = goyek.Define(goyek.Task{
		Name:  "nilaway",
		Usage: "run nilaway on source code",
		Action: func(a *goyek.A) {
			toolsbuild.NilAway(a)
		},
	})

	updateDependencies = goyek.Define(goyek.Task{
		Name:  "updateDependencies",
		Usage: "update dependencies",
		Action: func(a *goyek.A) {
			toolsbuild.UpdateDependencies(a)
		},
	})

	vulnCheck = goyek.Define(goyek.Task{
		Name:  "vulnCheck",
		Usage: "run vulnerability check on source code",
		Action: func(a *goyek.A) {
			toolsbuild.VulnerabilityCheck(a)
		},
	})

	_ = goyek.Define(goyek.Task{
		Name:  "preCommit",
		Usage: "run all pre-commit tasks",
		Deps: goyek.Deps{
			clean,
			updateDependencies,
			lint,
			nilaway,
			format,
			vulnCheck,
			tests,
			build,
		},
	})

	tests = goyek.Define(goyek.Task{
		Name:  "tests",
		Usage: "run unit tests",
		Action: func(a *goyek.A) {
			toolsbuild.UnitTests(a)
		},
	})
)

type config struct {
	// special functions:
	//   timestamp:   treat as Flag, but the associated value is generated at runtime
	Function string
	// flag name
	Flag string
	// value to use
	Value string
}

func buildExecutable(a *goyek.A) {
	loadProductData()
	// logged output shows up when running verbose (-v) or on error
	a.Logf("configuration: %#v", pD)
	pD.generateVersionInfo()
	if toolsbuild.Generate(a) {
		fmt.Println("building executable")
		toolsbuild.RunCommand(a, fmt.Sprintf("go build -ldflags %q -o %s .", strings.Join(pD.flags, " "), executableName))
	}
}

type productData struct {
	majorLevel      int
	minorLevel      int
	patchLevel      int
	semanticVersion string
	flags           []string
}

func (pD *productData) copyright() string {
	return fmt.Sprintf("Copyright © 2021-%d Marc Johnson", time.Now().Year())
}

func (pD *productData) generateVersionInfo() {
	data := goversioninfo.VersionInfo{}
	data.FixedFileInfo.FileVersion.Major = pD.majorLevel
	data.FixedFileInfo.FileVersion.Minor = pD.minorLevel
	data.FixedFileInfo.FileVersion.Patch = pD.patchLevel
	data.FixedFileInfo.ProductVersion.Major = pD.majorLevel
	data.FixedFileInfo.ProductVersion.Minor = pD.minorLevel
	data.FixedFileInfo.ProductVersion.Patch = pD.patchLevel
	data.FixedFileInfo.FileFlagsMask = "3f"
	data.FixedFileInfo.FileOS = "040004"
	data.FixedFileInfo.FileType = "01"
	data.StringFileInfo.FileDescription = applicationDescription
	data.StringFileInfo.FileVersion = pD.semanticVersion
	data.StringFileInfo.LegalCopyright = pD.copyright()
	data.StringFileInfo.ProductName = applicationName
	data.StringFileInfo.ProductVersion = pD.semanticVersion
	data.VarFileInfo.Translation.LangID = goversioninfo.LngUSEnglish
	data.VarFileInfo.Translation.CharsetID = goversioninfo.CsUnicode
	b, _ := json.Marshal(data)
	fullPath := filepath.Join(toolsbuild.WorkingDir(), versionInfoFile)
	if fileErr := afero.WriteFile(fileSystem, fullPath, b, 0o644); fileErr != nil {
		fmt.Printf("error writing %q! %v\n", versionInfoFile, fileErr)
	}
}

func loadProductData() {
	if pD == nil {
		rawPD := &productData{}
		rawYaml, _ := afero.ReadFile(fileSystem, "buildData.yaml")
		var data []config
		_ = yaml.Unmarshal(rawYaml, &data)
		const flagFormat = "-X %s=%s"
		for _, value := range data {
			switch value.Function {
			case "timestamp":
				rawPD.flags = append(rawPD.flags, fmt.Sprintf(flagFormat, value.Flag, time.Now().Format(time.RFC3339)))
			default:
				rawPD.flags = append(rawPD.flags, fmt.Sprintf(flagFormat, value.Flag, value.Value))
				switch value.Flag {
				case "mp3repair/cmd.version":
					var major int
					var minor int
					var patch int
					if count, scanErr := fmt.Sscanf(value.Value, "%d.%d.%d", &major, &minor, &patch); count == 3 && scanErr == nil {
						rawPD.majorLevel = major
						rawPD.minorLevel = minor
						rawPD.patchLevel = patch
						rawPD.semanticVersion = fmt.Sprintf("v%s", value.Value)
					}
				}
			}
		}
		sort.Strings(rawPD.flags)
		pD = rawPD
	}
}
