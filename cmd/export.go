package cmd

import (
	"fmt"
	"path/filepath"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
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
	ExportFlags = &SectionFlags{
		SectionName: ExportCommand,
		Details: map[string]*FlagDetails{
			ExportFlagDefaults: {
				AbbreviatedName: "d",
				Usage:           "write default program configuration data",
				ExpectedType:    BoolType,
				DefaultValue:    false,
			},
			ExportFlagOverwrite: {
				AbbreviatedName: "o",
				Usage:           "overwrite existing file",
				ExpectedType:    BoolType,
				DefaultValue:    false,
			},
		},
	}
	defaultConfigurationSettings = map[string]map[string]any{}
)

func addDefaults(sf *SectionFlags) {
	payload := map[string]any{}
	for flag, details := range sf.Details {
		bounded, ok := details.DefaultValue.(*cmdtoolkit.IntBounds)
		switch ok {
		case true:
			payload[flag] = bounded.DefaultValue
		case false:
			payload[flag] = details.DefaultValue
		}
	}
	defaultConfigurationSettings[sf.SectionName] = payload
}

type ExportSettings struct {
	DefaultsEnabled  CommandFlag[bool]
	OverwriteEnabled CommandFlag[bool]
}

func ExportRun(cmd *cobra.Command, _ []string) error {
	o := getBus()
	values, eSlice := ReadFlags(cmd.Flags(), ExportFlags)
	exitError := cmdtoolkit.NewExitProgrammingError(ExportCommand)
	if ProcessFlagErrors(o, eSlice) {
		settings, flagsOk := ProcessExportFlags(o, values)
		if flagsOk {
			LogCommandStart(o, ExportCommand, map[string]any{
				exportDefaultsAsFlag:  settings.DefaultsEnabled.Value,
				"defaults-user-set":   settings.DefaultsEnabled.UserSet,
				exportOverwriteAsFlag: settings.OverwriteEnabled.Value,
				"overwrite-user-set":  settings.OverwriteEnabled.UserSet,
			})
			exitError = settings.ExportDefaultConfiguration(o)
		}
	}
	return cmdtoolkit.ToErrorInterface(exitError)
}

func ProcessExportFlags(o output.Bus, values map[string]*CommandFlag[any]) (*ExportSettings, bool) {
	var flagErr error
	result := &ExportSettings{}
	flagsOk := true // optimistic
	result.DefaultsEnabled, flagErr = GetBool(o, values, ExportFlagDefaults)
	if flagErr != nil {
		flagsOk = false
	}
	result.OverwriteEnabled, flagErr = GetBool(o, values, ExportFlagOverwrite)
	if flagErr != nil {
		flagsOk = false
	}
	return result, flagsOk
}

func CreateConfigurationFile(o output.Bus, f string, content []byte) bool {
	if fileErr := WriteFile(f, content, cmdtoolkit.StdFilePermissions); fileErr != nil {
		cmdtoolkit.ReportFileCreationFailure(o, ExportCommand, f, fileErr)
		return false
	}
	o.WriteCanonicalConsole("File %q has been written", f)
	return true
}

func configFile() (path string, exists bool) {
	path = filepath.Join(ApplicationPath(), cmdtoolkit.DefaultConfigFileName())
	exists = PlainFileExists(path)
	return
}

func (es *ExportSettings) ExportDefaultConfiguration(o output.Bus) *cmdtoolkit.ExitError {
	if !es.CanWriteConfigurationFile(o) {
		return cmdtoolkit.NewExitUserError(ExportCommand)
	}
	// ignoring error return, as we're not marshalling structs, where mischief
	// can occur
	payload, _ := yaml.Marshal(defaultConfigurationSettings)
	f, exists := configFile()
	if exists {
		return es.OverwriteConfigurationFile(o, f, payload)
	}
	if !CreateConfigurationFile(o, f, payload) {
		return cmdtoolkit.NewExitSystemError(ExportCommand)
	}
	return nil
}

func (es *ExportSettings) OverwriteConfigurationFile(o output.Bus, f string, payload []byte) *cmdtoolkit.ExitError {
	if !es.CanOverwriteConfigurationFile(o, f) {
		return cmdtoolkit.NewExitUserError(ExportCommand)
	}
	backup := f + "-backup"
	if fileErr := Rename(f, backup); fileErr != nil {
		o.WriteCanonicalError("The file %q cannot be renamed to %q: %v", f, backup, fileErr)
		o.Log(output.Error, "rename failed", map[string]any{
			"error": fileErr,
			"old":   f,
			"new":   backup,
		})
		return cmdtoolkit.NewExitSystemError(ExportCommand)
	}
	if !CreateConfigurationFile(o, f, payload) {
		return cmdtoolkit.NewExitSystemError(ExportCommand)
	}
	_ = Remove(backup)
	return nil
}

func (es *ExportSettings) CanOverwriteConfigurationFile(o output.Bus, f string) bool {
	if !es.OverwriteEnabled.Value {
		o.WriteCanonicalError("The file %q exists and cannot be overwritten", f)
		o.Log(output.Error, "overwrite is not permitted", map[string]any{
			exportOverwriteAsFlag: false,
			"fileName":            f,
			"user-set":            es.OverwriteEnabled.UserSet,
		})
		switch {
		case es.OverwriteEnabled.UserSet:
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

func (es *ExportSettings) CanWriteConfigurationFile(o output.Bus) bool {
	if !es.DefaultsEnabled.Value {
		o.WriteCanonicalError("default configuration settings will not be exported")
		o.Log(output.Error, "export defaults disabled", map[string]any{
			exportDefaultsAsFlag: false,
			"user-set":           es.DefaultsEnabled.UserSet,
		})
		switch {
		case es.DefaultsEnabled.UserSet:
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
