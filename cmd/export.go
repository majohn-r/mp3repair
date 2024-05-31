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
			` exports default program configuration data to %APPDATA%\mp3repair\defaults.yaml`,
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
		bounded, ok := details.DefaultValue().(*cmd_toolkit.IntBounds)
		switch ok {
		case true:
			payload[flag] = bounded.Default()
		case false:
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
		settings, flagsOk := ProcessExportFlags(o, values)
		if flagsOk {
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

func ProcessExportFlags(o output.Bus, values map[string]*FlagValue) (*ExportFlagSettings, bool) {
	var flagErr error
	result := &ExportFlagSettings{}
	flagsOk := true // optimistic
	result.defaultsEnabled, result.defaultsSet, flagErr = GetBool(o, values, ExportFlagDefaults)
	if flagErr != nil {
		flagsOk = false
	}
	result.overwriteEnabled, result.overwriteSet, flagErr = GetBool(o, values, ExportFlagOverwrite)
	if flagErr != nil {
		flagsOk = false
	}
	return result, flagsOk
}

// TODO: Better name: CreateConfigurationFile
func CreateFile(o output.Bus, f string, content []byte) bool {
	if fileErr := WriteFile(f, content, cmd_toolkit.StdFilePermissions); fileErr != nil {
		cmd_toolkit.ReportFileCreationFailure(o, ExportCommand, f, fileErr)
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

func (efs *ExportFlagSettings) ExportDefaultConfiguration(o output.Bus) *ExitError {
	if !efs.CanWriteDefaults(o) {
		return NewExitUserError(ExportCommand)
	}
	// ignoring error return, as we're not marshalling structs, where mischief
	// can occur
	payload, _ := yaml.Marshal(defaultConfigurationSettings)
	f, exists := configFile()
	if exists {
		return efs.OverwriteFile(o, f, payload)
	}
	if !CreateFile(o, f, payload) {
		return NewExitSystemError(ExportCommand)
	}
	return nil
}

// TODO: better name OverwriteConfigurationFile
func (efs *ExportFlagSettings) OverwriteFile(o output.Bus, f string, payload []byte) *ExitError {
	if !efs.CanOverwriteFile(o, f) {
		return NewExitUserError(ExportCommand)
	}
	backup := f + "-backup"
	if fileErr := Rename(f, backup); fileErr != nil {
		o.WriteCanonicalError("The file %q cannot be renamed to %q: %v", f, backup, fileErr)
		o.Log(output.Error, "rename failed", map[string]any{
			"error": fileErr,
			"old":   f,
			"new":   backup,
		})
		return NewExitSystemError(ExportCommand)
	}
	if !CreateFile(o, f, payload) {
		return NewExitSystemError(ExportCommand)
	}
	Remove(backup)
	return nil
}

// TODO: better name: CanOverwriteConfigurationFile
func (efs *ExportFlagSettings) CanOverwriteFile(o output.Bus, f string) bool {
	if !efs.overwriteEnabled {
		o.WriteCanonicalError("The file %q exists and cannot be overwritten", f)
		o.Log(output.Error, "overwrite is not permitted", map[string]any{
			exportOverwriteAsFlag: false,
			"fileName":            f,
			"user-set":            efs.overwriteSet,
		})
		switch {
		case efs.overwriteSet:
			o.WriteCanonicalError("Why?\nYou explicitly set %s false", exportOverwriteAsFlag)
		default:
			o.WriteCanonicalError(
				"Why?\nAs currently configured, overwriting the file is disabled")
		}
		o.WriteCanonicalError(exportOverwriteCure)
		return false
	}
	return true
}

// TODO: better name: CanWriteConfigurationFile
func (efs *ExportFlagSettings) CanWriteDefaults(o output.Bus) bool {
	if !efs.defaultsEnabled {
		o.WriteCanonicalError("default configuration settings will not be exported")
		o.Log(output.Error, "export defaults disabled", map[string]any{
			exportDefaultsAsFlag: false,
			"user-set":           efs.defaultsSet,
		})
		switch {
		case efs.defaultsSet:
			o.WriteCanonicalError("Why?\nYou explicitly set %s false", exportDefaultsAsFlag)
		default:
			o.WriteCanonicalError("Why?\nAs currently configured, exporting default" +
				" configuration settings is disabled")
		}
		o.WriteCanonicalError(exportDefaultsCure)
		return false
	}
	return true
}

func init() {
	RootCmd.AddCommand(ExportCmd)
	addDefaults(ExportFlags)
	o := getBus()
	c := getConfiguration()
	AddFlags(o, c, ExportCmd.Flags(), ExportFlags)
}
