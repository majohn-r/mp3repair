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

	fKBackupFile    = "backup"
	fKDefaultsFlag  = "-" + defaultsFlag
	fKOriginalFile  = "original"
	fKOverwriteFlag = "-" + overwriteFlag
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

func (ex *export) Exec(o output.Bus, args []string) (ok bool) {
	if internal.ProcessArgs(o, ex.f, args) {
		ok = ex.runCommand(o)
	}
	return
}

func (ex *export) logFields() map[string]any {
	return map[string]any{
		fieldKeyCommandName: exportCommandName,
		fKDefaultsFlag:      *ex.defaults,
		fKOverwriteFlag:     *ex.overwrite,
	}
}

func (ex *export) runCommand(o output.Bus) (ok bool) {
	if !*ex.defaults {
		o.WriteCanonicalError(internal.UserSpecifiedNoWork, exportCommandName)
		o.Log(output.Error, internal.LogErrorNothingToDo, ex.logFields())
		return
	}
	return ex.exportDefaults(o)
}

func (ex *export) exportDefaults(o output.Bus) bool {
	if !*ex.defaults {
		return true
	}
	return ex.writeDefaults(o, getDefaultsContent())
}

func getDefaultsContent() []byte {
	// get the search content - it cannot be registered as the commands register
	// their content, due to circular dependency issues
	searchName, searchDefaults := files.SearchDefaults()
	defaultMapping[searchName] = searchDefaults
	// ignoring error return, as we're not marshalling structs, where mischief
	// can occur
	content, _ := yaml.Marshal(defaultMapping)
	return content
}

func (ex *export) writeDefaults(o output.Bus, content []byte) (ok bool) {
	if appData, appDataOk := internal.LookupAppData(o); appDataOk {
		path := internal.CreateAppSpecificPath(appData)
		if ensurePathExists(o, path) {
			configFile := filepath.Join(path, internal.DefaultConfigFileName)
			if internal.PlainFileExists(configFile) {
				ok = ex.overwriteFile(o, configFile, content)
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
			o.WriteCanonicalError(internal.UserCannotCreateDirectory, path, err)
			o.Log(output.Error, internal.LogErrorCannotCreateDirectory, map[string]any{
				internal.FieldKeyDirectory: path,
				internal.FieldKeyError:     err,
			})
		} else {
			ok = true
		}
	}
	return
}

func (ex *export) overwriteFile(o output.Bus, fileName string, content []byte) (ok bool) {
	if !*ex.overwrite {
		o.WriteCanonicalError(internal.UserNoOverwriteAllowed, fileName, overwriteFlag)
		o.Log(output.Error, internal.LogErrorOverwriteDisabled, map[string]any{
			fKOverwriteFlag:           false,
			internal.FieldKeyFileName: fileName,
		})
	} else {
		backupFileName := fileName + "-backup"
		if err := os.Rename(fileName, backupFileName); err != nil {
			o.WriteCanonicalError(internal.UserCannotRenameFile, fileName, backupFileName, err)
			o.Log(output.Error, internal.LogErrorRenameError, map[string]any{
				internal.FieldKeyError: err,
				fKOriginalFile:         fileName,
				fKBackupFile:           backupFileName,
			})
		} else {
			if createFile(o, fileName, content) {
				os.Remove(backupFileName)
				ok = true
			}
		}
	}
	return
}

func createFile(o output.Bus, fileName string, content []byte) bool {
	if err := os.WriteFile(fileName, content, 0644); err != nil {
		o.WriteCanonicalError(internal.UserCannotCreateFile, fileName, err)
		o.Log(output.Error, internal.LogErrorCannotCreateFile, map[string]any{
			internal.FieldKeyFileName: fileName,
			internal.FieldKeyError:    err,
		})
		return false
	}
	return true
}
