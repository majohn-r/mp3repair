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
	addCommandData(exportCommandName, commandData{isDefault: false, init: newExport})
	addDefaultMapping(exportCommandName, map[string]any{
		defaultsFlag:  defaultDefaults,
		overwriteFlag: defaultOverwrite,
	})
}

var defaults = map[string]map[string]any{}

func addDefaultMapping(name string, mapping map[string]any) {
	defaults[name] = mapping
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
	if defaults, ok := evaluateExportDefaults(o, c.SubConfiguration(exportCommandName)); ok {
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

func evaluateExportDefaults(o output.Bus, c *internal.Configuration) (v exportDefaultValues, ok bool) {
	ok = true
	v = exportDefaultValues{}
	var err error
	if v.defaults, err = c.BoolDefault(defaultsFlag, defaultDefaults); err != nil {
		reportBadDefault(o, exportCommandName, err)
		ok = false
	}
	if v.overwrite, err = c.BoolDefault(overwriteFlag, defaultOverwrite); err != nil {
		reportBadDefault(o, exportCommandName, err)
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

func (e *export) runCommand(o output.Bus) bool {
	if !*e.defaults {
		reportNothingToDo(o, exportCommandName, e.logFields())
		return false
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
	// get the search content - it could not be registered the same way that
	// commands pre-register their content, due to circular dependency issues
	s, m := files.SearchDefaults()
	defaults[s] = m
	// ignoring error return, as we're not marshalling structs, where mischief
	// can occur
	b, _ := yaml.Marshal(defaults)
	return b
}

func (e *export) writeDefaults(o output.Bus, b []byte) bool {
	path := internal.ApplicationPath()
	f := filepath.Join(path, internal.DefaultConfigFileName)
	if internal.PlainFileExists(f) {
		return e.overwriteFile(o, f, b)
	}
	return createFile(o, f, b)
}

func (e *export) overwriteFile(o output.Bus, f string, b []byte) (ok bool) {
	if !*e.overwrite {
		o.WriteCanonicalError("The file %q exists; set the %s flag to true if you want it overwritten", f, overwriteFlag)
		o.Log(output.Error, "overwrite is not permitted", map[string]any{
			"-" + overwriteFlag: false,
			"fileName":          f,
		})
	} else {
		backup := f + "-backup"
		if err := os.Rename(f, backup); err != nil {
			o.WriteCanonicalError("The file %q cannot be renamed to %q: %v", f, backup, err)
			o.Log(output.Error, "rename failed", map[string]any{
				"error": err,
				"old":   f,
				"new":   backup,
			})
		} else if createFile(o, f, b) {
			os.Remove(backup)
			ok = true
		}
	}
	return
}

func createFile(o output.Bus, f string, content []byte) bool {
	if err := os.WriteFile(f, content, internal.StdFilePermissions); err != nil {
		reportFileCreationFailure(o, exportCommandName, f, err)
		return false
	}
	return true
}
