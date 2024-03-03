/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd

import (
	"fmt"
	"path/filepath"

	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	ExportCommand        = "export"
	ExportFlagDefaults   = "defaults"
	exportDefaultsAsFlag = "--" + ExportFlagDefaults
	exportDefaultsCure   = "What to do:\nUse either '" + exportDefaultsAsFlag + "' or '" +
		exportDefaultsAsFlag + "=true' to enable exporting defaults"
	ExportFlagOverwrite   = "overwrite"
	exportOverwriteAsFlag = "--" + ExportFlagOverwrite
	exportOverwriteCure   = "What to do:\nUse either '" + exportOverwriteAsFlag + "' or '" +
		exportOverwriteAsFlag + "=true' to enable overwriting the existing file"
)

// ExportCmd represents the export command
var (
	ExportCmd = &cobra.Command{
		Use: ExportCommand + " [" + exportDefaultsAsFlag + "] [" +
			exportOverwriteAsFlag + "]",
		DisableFlagsInUseLine: true,
		Short:                 "Exports default program configuration data",
		Long: fmt.Sprintf("%q", ExportCommand) +
			` exports default program configuration data to %APPDATA%\mp3\defaults.yaml`,
		Example: ExportCommand + " " + exportDefaultsAsFlag + "\n" +
			"  Write default program configuration data\n" +
			ExportCommand + " " + exportOverwriteAsFlag + "\n" +
			"  Overwrite a pre-existing defaults.yaml file",
		RunE: ExportRun,
	}
	ExportFlags = NewSectionFlags().WithSectionName(ExportCommand).WithFlags(
		map[string]*FlagDetails{
			ExportFlagDefaults: NewFlagDetails().WithAbbreviatedName("d").WithUsage(
				"write default program configuration data").WithExpectedType(
				BoolType).WithDefaultValue(false),
			ExportFlagOverwrite: NewFlagDetails().WithAbbreviatedName("o").WithUsage(
				"overwrite existing file").WithExpectedType(BoolType).WithDefaultValue(false),
		},
	)
	defaultConfigurationSettings = map[string]map[string]any{}
)

func addDefaults(sf *SectionFlags) {
	payload := map[string]any{}
	for flag, details := range sf.Flags() {
		if bounded, ok := details.DefaultValue().(*cmd_toolkit.IntBounds); ok {
			payload[flag] = bounded.Default()
		} else {
			payload[flag] = details.DefaultValue()
		}
	}
	defaultConfigurationSettings[sf.SectionName()] = payload
}

type ExportFlagSettings struct {
	defaultsEnabled  bool
	defaultsSet      bool
	overwriteEnabled bool
	overwriteSet     bool
}

func NewExportFlagSettings() *ExportFlagSettings {
	return &ExportFlagSettings{}
}

func (efs *ExportFlagSettings) WithDefaultsEnabled(b bool) *ExportFlagSettings {
	efs.defaultsEnabled = b
	return efs
}

func (efs *ExportFlagSettings) WithDefaultsSet(b bool) *ExportFlagSettings {
	efs.defaultsSet = b
	return efs
}

func (efs *ExportFlagSettings) WithOverwriteEnabled(b bool) *ExportFlagSettings {
	efs.overwriteEnabled = b
	return efs
}

func (efs *ExportFlagSettings) WithOverwriteSet(b bool) *ExportFlagSettings {
	efs.overwriteSet = b
	return efs
}

func ExportRun(cmd *cobra.Command, _ []string) error {
	o := getBus()
	values, eSlice := ReadFlags(cmd.Flags(), ExportFlags)
	exitError := NewExitProgrammingError(ExportCommand)
	if ProcessFlagErrors(o, eSlice) {
		settings, ok := ProcessExportFlags(o, values)
		if ok {
			LogCommandStart(o, ExportCommand, map[string]any{
				exportDefaultsAsFlag:  settings.defaultsEnabled,
				"defaults-user-set":   settings.defaultsSet,
				exportOverwriteAsFlag: settings.overwriteEnabled,
				"overwrite-user-set":  settings.overwriteSet,
			})
			exitError = settings.ExportDefaultConfiguration(o)
		}
	}
	return ToErrorInterface(exitError)
}

func ProcessExportFlags(o output.Bus, values map[string]*FlagValue) (*ExportFlagSettings,
	bool) {
	var err error
	result := &ExportFlagSettings{}
	ok := true // optimistic
	result.defaultsEnabled, result.defaultsSet, err = GetBool(o, values, ExportFlagDefaults)
	if err != nil {
		ok = false
	}
	result.overwriteEnabled, result.overwriteSet, err = GetBool(o, values,
		ExportFlagOverwrite)
	if err != nil {
		ok = false
	}
	return result, ok
}

func CreateFile(o output.Bus, f string, content []byte) bool {
	if err := WriteFile(f, content, cmd_toolkit.StdFilePermissions); err != nil {
		cmd_toolkit.ReportFileCreationFailure(o, ExportCommand, f, err)
		return false
	}
	o.WriteCanonicalConsole("File %q has been written", f)
	return true
}

func configFile() (path string, exists bool) {
	path = filepath.Join(ApplicationPath(), cmd_toolkit.DefaultConfigFileName())
	exists = PlainFileExists(path)
	return
}

func (efs *ExportFlagSettings) ExportDefaultConfiguration(o output.Bus) (err *ExitError) {
	err = NewExitUserError(ExportCommand)
	if efs.CanWriteDefaults(o) {
		// ignoring error return, as we're not marshalling structs, where mischief
		// can occur
		payload, _ := yaml.Marshal(defaultConfigurationSettings)
		if f, exists := configFile(); exists {
			err = efs.OverwriteFile(o, f, payload)
		} else {
			if CreateFile(o, f, payload) {
				err = nil
			} else {
				err = NewExitSystemError(ExportCommand)
			}
		}
	}
	return
}

func (efs *ExportFlagSettings) OverwriteFile(o output.Bus, f string, payload []byte) (e *ExitError) {
	e = NewExitUserError(ExportCommand)
	if efs.CanOverwriteFile(o, f) {
		e = NewExitSystemError(ExportCommand)
		backup := f + "-backup"
		if err := Rename(f, backup); err != nil {
			o.WriteCanonicalError("The file %q cannot be renamed to %q: %v", f, backup, err)
			o.Log(output.Error, "rename failed", map[string]any{
				"error": err,
				"old":   f,
				"new":   backup,
			})
		} else if CreateFile(o, f, payload) {
			Remove(backup)
			e = nil
		}
	}
	return
}

func (efs *ExportFlagSettings) CanOverwriteFile(o output.Bus, f string) (canOverwrite bool) {
	if !efs.overwriteEnabled {
		o.WriteCanonicalError("The file %q exists and cannot be overwritten", f)
		o.Log(output.Error, "overwrite is not permitted", map[string]any{
			exportOverwriteAsFlag: false,
			"fileName":            f,
			"user-set":            efs.overwriteSet,
		})
		if efs.overwriteSet {
			o.WriteCanonicalError("Why?\nYou explicitly set %s false", exportOverwriteAsFlag)
		} else {
			o.WriteCanonicalError(
				"Why?\nAs currently configured, overwriting the file is disabled")
		}
		o.WriteCanonicalError(exportOverwriteCure)
	} else {
		canOverwrite = true
	}
	return
}

func (efs *ExportFlagSettings) CanWriteDefaults(o output.Bus) (canWrite bool) {
	if !efs.defaultsEnabled {
		o.WriteCanonicalError("default configuration settings will not be exported")
		o.Log(output.Error, "export defaults disabled", map[string]any{
			exportDefaultsAsFlag: false,
			"user-set":           efs.defaultsSet,
		})
		if efs.defaultsSet {
			o.WriteCanonicalError("Why?\nYou explicitly set %s false", exportDefaultsAsFlag)
		} else {
			o.WriteCanonicalError("Why?\nAs currently configured, exporting default" +
				" configuration settings is disabled")
		}
		o.WriteCanonicalError(exportDefaultsCure)
	} else {
		canWrite = true
	}
	return
}

func init() {
	RootCmd.AddCommand(ExportCmd)
	addDefaults(ExportFlags)
	o := getBus()
	c := getConfiguration()
	AddFlags(o, c, ExportCmd.Flags(), ExportFlags)
}
