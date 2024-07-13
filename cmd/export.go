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
	exportCommand        = "export"
	exportFlagDefaults   = "defaults"
	exportDefaultsAsFlag = "--" + exportFlagDefaults
	exportDefaultsCure   = "What to do:\nUse either '" + exportDefaultsAsFlag + "' or '" +
		exportDefaultsAsFlag + "=true' to enable exporting defaults"
	exportFlagOverwrite   = "overwrite"
	exportOverwriteAsFlag = "--" + exportFlagOverwrite
	exportOverwriteCure   = "What to do:\nUse either '" + exportOverwriteAsFlag + "' or '" +
		exportOverwriteAsFlag + "=true' to enable overwriting the existing file"
)

var (
	exportCmd = &cobra.Command{
		Use: exportCommand + " [" + exportDefaultsAsFlag + "] [" +
			exportOverwriteAsFlag + "]",
		DisableFlagsInUseLine: true,
		Short:                 "Exports default program configuration data",
		Long: fmt.Sprintf("%q", exportCommand) +
			` exports default program configuration data to %APPDATA%\mp3repair\defaults.yaml`,
		Example: exportCommand + " " + exportDefaultsAsFlag + "\n" +
			"  Write default program configuration data\n" +
			exportCommand + " " + exportOverwriteAsFlag + "\n" +
			"  Overwrite a pre-existing defaults.yaml file",
		RunE: exportRun,
	}
	exportFlags = &cmdtoolkit.FlagSet{
		Name: exportCommand,
		Details: map[string]*cmdtoolkit.FlagDetails{
			exportFlagDefaults: {
				AbbreviatedName: "d",
				Usage:           "write default program configuration data",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
			exportFlagOverwrite: {
				AbbreviatedName: "o",
				Usage:           "overwrite existing file",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
		},
	}
	defaultConfigurationSettings = map[string]map[string]any{}
)

func addDefaults(sf *cmdtoolkit.FlagSet) {
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
	defaultConfigurationSettings[sf.Name] = payload
}

type exportSettings struct {
	defaultsEnabled  cmdtoolkit.CommandFlag[bool]
	overwriteEnabled cmdtoolkit.CommandFlag[bool]
}

func exportRun(cmd *cobra.Command, _ []string) error {
	o := getBus()
	values, eSlice := cmdtoolkit.ReadFlags(cmd.Flags(), exportFlags)
	exitError := cmdtoolkit.NewExitProgrammingError(exportCommand)
	if cmdtoolkit.ProcessFlagErrors(o, eSlice) {
		settings, flagsOk := processExportFlags(o, values)
		if flagsOk {
			logCommandStart(o, exportCommand, map[string]any{
				exportDefaultsAsFlag:  settings.defaultsEnabled.Value,
				"defaults-user-set":   settings.defaultsEnabled.UserSet,
				exportOverwriteAsFlag: settings.overwriteEnabled.Value,
				"overwrite-user-set":  settings.overwriteEnabled.UserSet,
			})
			exitError = settings.exportDefaultConfiguration(o)
		}
	}
	return cmdtoolkit.ToErrorInterface(exitError)
}

func processExportFlags(o output.Bus, values map[string]*cmdtoolkit.CommandFlag[any]) (*exportSettings, bool) {
	var flagErr error
	result := &exportSettings{}
	flagsOk := true // optimistic
	result.defaultsEnabled, flagErr = cmdtoolkit.GetBool(o, values, exportFlagDefaults)
	if flagErr != nil {
		flagsOk = false
	}
	result.overwriteEnabled, flagErr = cmdtoolkit.GetBool(o, values, exportFlagOverwrite)
	if flagErr != nil {
		flagsOk = false
	}
	return result, flagsOk
}

func createConfigurationFile(o output.Bus, f string, content []byte) bool {
	if fileErr := writeFile(f, content, cmdtoolkit.StdFilePermissions); fileErr != nil {
		cmdtoolkit.ReportFileCreationFailure(o, exportCommand, f, fileErr)
		return false
	}
	o.WriteCanonicalConsole("File %q has been written", f)
	return true
}

func configFile() (path string, exists bool) {
	path = filepath.Join(applicationPath(), cmdtoolkit.DefaultConfigFileName())
	exists = plainFileExists(path)
	return
}

func (es *exportSettings) exportDefaultConfiguration(o output.Bus) *cmdtoolkit.ExitError {
	if !es.canWriteConfigurationFile(o) {
		return cmdtoolkit.NewExitUserError(exportCommand)
	}
	// ignoring error return, as we're not marshalling structs, where mischief
	// can occur
	payload, _ := yaml.Marshal(defaultConfigurationSettings)
	f, exists := configFile()
	if exists {
		return es.overwriteConfigurationFile(o, f, payload)
	}
	if !createConfigurationFile(o, f, payload) {
		return cmdtoolkit.NewExitSystemError(exportCommand)
	}
	return nil
}

func (es *exportSettings) overwriteConfigurationFile(o output.Bus, f string, payload []byte) *cmdtoolkit.ExitError {
	if !es.canOverwriteConfigurationFile(o, f) {
		return cmdtoolkit.NewExitUserError(exportCommand)
	}
	backup := f + "-backup"
	if fileErr := rename(f, backup); fileErr != nil {
		o.WriteCanonicalError("The file %q cannot be renamed to %q: %v", f, backup, fileErr)
		o.Log(output.Error, "rename failed", map[string]any{
			"error": fileErr,
			"old":   f,
			"new":   backup,
		})
		return cmdtoolkit.NewExitSystemError(exportCommand)
	}
	if !createConfigurationFile(o, f, payload) {
		return cmdtoolkit.NewExitSystemError(exportCommand)
	}
	_ = remove(backup)
	return nil
}

func (es *exportSettings) canOverwriteConfigurationFile(o output.Bus, f string) bool {
	if !es.overwriteEnabled.Value {
		o.WriteCanonicalError("The file %q exists and cannot be overwritten", f)
		o.Log(output.Error, "overwrite is not permitted", map[string]any{
			exportOverwriteAsFlag: false,
			"fileName":            f,
			"user-set":            es.overwriteEnabled.UserSet,
		})
		switch {
		case es.overwriteEnabled.UserSet:
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

func (es *exportSettings) canWriteConfigurationFile(o output.Bus) bool {
	if !es.defaultsEnabled.Value {
		o.WriteCanonicalError("default configuration settings will not be exported")
		o.Log(output.Error, "export defaults disabled", map[string]any{
			exportDefaultsAsFlag: false,
			"user-set":           es.defaultsEnabled.UserSet,
		})
		switch {
		case es.defaultsEnabled.UserSet:
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
	rootCmd.AddCommand(exportCmd)
	addDefaults(exportFlags)
	o := getBus()
	c := getConfiguration()
	cmdtoolkit.AddFlags(o, c, exportCmd.Flags(), exportFlags)
}
