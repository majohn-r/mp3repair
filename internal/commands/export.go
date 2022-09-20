package commands

import (
	"flag"
	"mp3/internal"
	"mp3/internal/files"
	"os"
	"path/filepath"

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

func (ex *export) name() string {
	return exportCommandName
}

func newExport(o internal.OutputBus, c *internal.Configuration, fSet *flag.FlagSet) (CommandProcessor, bool) {
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

func newExportCommand(o internal.OutputBus, c *internal.Configuration, fSet *flag.FlagSet) (*export, bool) {
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

func evaluateExportDefaults(o internal.OutputBus, c *internal.Configuration, name string) (defaults exportDefaultValues, ok bool) {
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

func (ex *export) Exec(o internal.OutputBus, args []string) (ok bool) {
	if internal.ProcessArgs(o, ex.f, args) {
		ok = ex.runCommand(o)
	}
	return
}

func (ex *export) logFields() map[string]any {
	return map[string]any{
		fkCommandName:   ex.name(),
		fKDefaultsFlag:  *ex.defaults,
		fKOverwriteFlag: *ex.overwrite,
	}
}

func (ex *export) runCommand(o internal.OutputBus) (ok bool) {
	if !*ex.defaults {
		o.WriteError(internal.USER_SPECIFIED_NO_WORK, ex.name())
		o.LogWriter().Error(internal.LE_NOTHING_TO_DO, ex.logFields())
		return
	}
	return ex.exportDefaults(o)
}

func (ex *export) exportDefaults(o internal.OutputBus) bool {
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

func (ex *export) writeDefaults(o internal.OutputBus, content []byte) (ok bool) {
	if appData, appDataOk := internal.LookupAppData(o); appDataOk {
		path := internal.CreateAppSpecificPath(appData)
		if ensurePathExists(o, path){
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

func ensurePathExists(o internal.OutputBus, path string) (ok bool){
	if internal.DirExists(path) {
		ok = true
	} else {
		if err := internal.Mkdir(path); err != nil {
			o.WriteError(internal.USER_CANNOT_CREATE_DIRECTORY, path, err)
			o.LogWriter().Error(internal.LE_CANNOT_CREATE_DIRECTORY, map[string]any{
				internal.FK_DIRECTORY: path,
				internal.FK_ERROR: err,
			})
		} else {
			ok = true
		}
	}
	return
}

func (ex *export) overwriteFile(o internal.OutputBus, fileName string, content []byte) (ok bool) {
	if !*ex.overwrite {
		o.WriteError(internal.USER_NO_OVERWRITE_ALLOWED, fileName, overwriteFlag)
		o.LogWriter().Error(internal.LE_OVERWRITE_DISABLED, map[string]any{
			fKOverwriteFlag:       false,
			internal.FK_FILE_NAME: fileName,
		})
	} else {
		backupFileName := fileName + "-backup"
		if err := os.Rename(fileName, backupFileName); err != nil {
			o.WriteError(internal.USER_CANNOT_RENAME_FILE, fileName, backupFileName, err)
			o.LogWriter().Error(internal.LE_RENAME_ERROR, map[string]any{
				internal.FK_ERROR: err,
				fKOriginalFile:    fileName,
				fKBackupFile:      backupFileName,
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

func createFile(o internal.OutputBus, fileName string, content []byte) bool {
	if err := os.WriteFile(fileName, content, 0644); err != nil {
		o.WriteError(internal.USER_CANNOT_CREATE_FILE, fileName, err)
		o.LogWriter().Error(internal.LE_CANNOT_CREATE_FILE, map[string]any{
			internal.FK_FILE_NAME: fileName,
			internal.FK_ERROR:     err,
		})
		return false
	}
	return true
}
