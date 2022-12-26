package commands

import (
	"flag"
	"mp3/internal"
	"mp3/internal/files"
	"os"
	"path/filepath"

	"github.com/majohn-r/output"
	"gopkg.in/yaml.v3"
)

func init() {
	addCommandData(exportCommandName, commandData{isDefault: false, initFunction: newExport})
	addDefaultMapping(exportCommandName, map[string]any{
		defaultsFlag:  defaultDefaults,
		overwriteFlag: defaultOverwrite,
	})
}

var defaultMapping = map[string]map[string]any{}

func addDefaultMapping(name string, mapping map[string]any) {
	defaultMapping[name] = mapping
}

type export struct {
	defaults  *bool
	overwrite *bool
	f         *flag.FlagSet
}

func newExport(o output.Bus, c *internal.Configuration, fSet *flag.FlagSet) (CommandProcessor, bool) {
	return newExportCommand(o, c, fSet)
}

const (
	exportCommandName = "export"

	defaultDefaults  = false
	defaultOverwrite = false

	defaultsFlag  = "defaults"
	overwriteFlag = "overwrite"
)

type exportDefaultValues struct {
	defaults  bool
	overwrite bool
}

func newExportCommand(o output.Bus, c *internal.Configuration, fSet *flag.FlagSet) (*export, bool) {
	defaults, defaultsOk := evaluateExportDefaults(o, c.SubConfiguration(exportCommandName), exportCommandName)
	if defaultsOk {
		defaultsUsage := internal.DecorateBoolFlagUsage("write defaults.yaml", defaults.defaults)
		overwriteUsage := internal.DecorateBoolFlagUsage("overwrite file if it exists", defaults.overwrite)
		return &export{
			defaults:  fSet.Bool(defaultsFlag, defaults.defaults, defaultsUsage),
			overwrite: fSet.Bool(overwriteFlag, defaults.overwrite, overwriteUsage),
			f:         fSet,
		}, true
	}
	return nil, false
}

func evaluateExportDefaults(o output.Bus, c *internal.Configuration, name string) (defaults exportDefaultValues, ok bool) {
	ok = true
	defaults = exportDefaultValues{}
	var err error
	defaults.defaults, err = c.BoolDefault(defaultsFlag, defaultDefaults)
	if err != nil {
		reportBadDefault(o, name, err)
		ok = false
	}
	defaults.overwrite, err = c.BoolDefault(overwriteFlag, defaultOverwrite)
	if err != nil {
		reportBadDefault(o, name, err)
		ok = false
	}
	return
}

func (e *export) Exec(o output.Bus, args []string) (ok bool) {
	if internal.ProcessArgs(o, e.f, args) {
		ok = e.runCommand(o)
	}
	return
}

func (e *export) logFields() map[string]any {
	return map[string]any{
		"command":           exportCommandName,
		"-" + defaultsFlag:  *e.defaults,
		"-" + overwriteFlag: *e.overwrite,
	}
}

func (e *export) runCommand(o output.Bus) (ok bool) {
	if !*e.defaults {
		reportNothingToDo(o, exportCommandName, e.logFields())
		return
	}
	return e.exportDefaults(o)
}

func (e *export) exportDefaults(o output.Bus) bool {
	if !*e.defaults {
		return true
	}
	return e.writeDefaults(o, defaultsContent())
}

func defaultsContent() []byte {
	// get the search content - it cannot be registered the same way that
	// commands register their content, due to circular dependency issues
	searchName, searchDefaults := files.SearchDefaults()
	defaultMapping[searchName] = searchDefaults
	// ignoring error return, as we're not marshalling structs, where mischief
	// can occur
	content, _ := yaml.Marshal(defaultMapping)
	return content
}

func (e *export) writeDefaults(o output.Bus, content []byte) (ok bool) {
	if appData, appDataOk := internal.LookupAppData(o); appDataOk {
		path := internal.CreateAppSpecificPath(appData)
		if ensurePathExists(o, path) {
			configFile := filepath.Join(path, internal.DefaultConfigFileName)
			if internal.PlainFileExists(configFile) {
				ok = e.overwriteFile(o, configFile, content)
			} else {
				ok = createFile(o, configFile, content)
			}
		}
	}
	return
}

func ensurePathExists(o output.Bus, path string) (ok bool) {
	if internal.DirExists(path) {
		ok = true
	} else {
		if err := internal.Mkdir(path); err != nil {
			reportDirectoryCreationFailure(o, exportCommandName, path, err)
		} else {
			ok = true
		}
	}
	return
}

func (e *export) overwriteFile(o output.Bus, fileName string, content []byte) (ok bool) {
	if !*e.overwrite {
		o.WriteCanonicalError("The file %q exists; set the %s flag to true if you want it overwritten", fileName, overwriteFlag)
		o.Log(output.Error, "overwrite is not permitted", map[string]any{
			"-" + overwriteFlag: false,
			"fileName":          fileName,
		})
	} else {
		backupFileName := fileName + "-backup"
		if err := os.Rename(fileName, backupFileName); err != nil {
			o.WriteCanonicalError("The file %q cannot be renamed to %q: %v", fileName, backupFileName, err)
			o.Log(output.Error, "rename failed", map[string]any{
				"error":    err,
				"original": fileName,
				"backup":   backupFileName,
			})
		} else if createFile(o, fileName, content) {
			os.Remove(backupFileName)
			ok = true
		}
	}
	return
}

func createFile(o output.Bus, fileName string, content []byte) bool {
	if err := os.WriteFile(fileName, content, 0o644); err != nil {
		reportFileCreationFailure(o, exportCommandName, fileName, err)
		return false
	}
	return true
}
