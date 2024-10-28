package cmd

import (
	"fmt"
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

const (
	exportCommand         = "export"
	exportFlagDefaults    = "defaults"
	exportDefaultsAsFlag  = "--" + exportFlagDefaults
	exportFlagOverwrite   = "overwrite"
	exportOverwriteAsFlag = "--" + exportFlagOverwrite
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
)

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
	o.ConsolePrintf("File %q has been written.\n", f)
	return true
}

func (es *exportSettings) exportDefaultConfiguration(o output.Bus) *cmdtoolkit.ExitError {
	if !es.canWriteConfigurationFile(o) {
		return cmdtoolkit.NewExitUserError(exportCommand)
	}
	// ignoring error return, as we're not marshalling structs, where mischief
	// can occur
	f, exists := cmdtoolkit.DefaultConfigFileStatus()
	payload := cmdtoolkit.WritableDefaults()
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
		o.ErrorPrintf("The file %q cannot be renamed to %q: %v.\n", f, backup, cmdtoolkit.ErrorToString(fileErr))
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
		o.ErrorPrintf("The file %q exists and cannot be overwritten.\n", f)
		o.Log(output.Error, "overwrite is not permitted", map[string]any{
			exportOverwriteAsFlag: false,
			"fileName":            f,
			"user-set":            es.overwriteEnabled.UserSet,
		})
		o.ErrorPrintln("Why?")
		switch {
		case es.overwriteEnabled.UserSet:
			o.ErrorPrintf("You explicitly set %s false.\n", exportOverwriteAsFlag)
		default:
			o.ErrorPrintln("As currently configured, overwriting the file is disabled.")
		}
		o.ErrorPrintln("What to do:")
		o.ErrorPrintln("To enable overwriting the existing file, use either:")
		o.BeginErrorList(false)
		o.ErrorPrintln(exportOverwriteAsFlag + " or")
		o.ErrorPrintln(exportOverwriteAsFlag + "=true")
		o.EndErrorList()
		return false
	}
	return true
}

func (es *exportSettings) canWriteConfigurationFile(o output.Bus) bool {
	if !es.defaultsEnabled.Value {
		o.ErrorPrintln("Default configuration settings will not be exported.")
		o.Log(output.Error, "export defaults disabled", map[string]any{
			exportDefaultsAsFlag: false,
			"user-set":           es.defaultsEnabled.UserSet,
		})
		o.ErrorPrintln("Why?")
		switch {
		case es.defaultsEnabled.UserSet:
			o.ErrorPrintf("You explicitly set %s false.\n", exportDefaultsAsFlag)
		default:
			o.ErrorPrintln("As currently configured, exporting default configuration settings is disabled.")
		}
		o.ErrorPrintln("What to do:")
		o.ErrorPrintln("To enable exporting defaults, use either:")
		o.BeginErrorList(false)
		o.ErrorPrintln(exportDefaultsAsFlag + " or")
		o.ErrorPrintln(exportDefaultsAsFlag + "=true")
		o.EndErrorList()
		return false
	}
	return true
}

func init() {
	rootCmd.AddCommand(exportCmd)
	cmdtoolkit.AddDefaults(exportFlags)
	cmdtoolkit.AddFlags(getBus(), getConfiguration(), exportCmd.Flags(), exportFlags)
}
