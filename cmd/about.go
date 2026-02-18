/*
Copyright © 2026 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd

import (
	"fmt"
	"strings"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

const (
	aboutCommand   = "about"
	aboutStyle     = "style"
	aboutStyleFlag = "--" + aboutStyle
	author         = "Marc Johnson"
	firstYear      = 2021 // the year when development of this application began
)

var (
	// version is the application's semantic version and its value is injected by the build
	version string
	// creation is the application's build timestamp in RFC3339 format (2006-01-02T15:04:05Z07:00)
	// and its value is injected by the build
	creation string
	aboutCmd = &cobra.Command{
		Use:                   aboutCommand + " [" + aboutStyleFlag + " name]",
		DisableFlagsInUseLine: true,
		Short:                 genAboutShortHelp(cmdtoolkit.AppName()),
		Long:                  getAboutLongHelp(cmdtoolkit.AppName()),
		Example: aboutCommand + " " + aboutStyleFlag + " name\n" +
			"  Write 'about' information in a box of the named style.\n" +
			"  Valid names are:\n" +
			"  ● ascii\n    " + strings.Join(
			cmdtoolkit.StyledFlowerBox([]string{"output ..."}, cmdtoolkit.ASCIIFlowerBox)[0:3],
			"\n    ",
		) +
			"\n  ● rounded (default)\n    " + strings.Join(
			cmdtoolkit.StyledFlowerBox([]string{"output ..."}, cmdtoolkit.CurvedFlowerBox)[0:3],
			"\n    ",
		) +
			"\n  ● light\n    " + strings.Join(
			cmdtoolkit.StyledFlowerBox([]string{"output ..."}, cmdtoolkit.LightLinedFlowerBox)[0:3],
			"\n    ",
		) +
			"\n  ● heavy\n    " + strings.Join(
			cmdtoolkit.StyledFlowerBox([]string{"output ..."}, cmdtoolkit.HeavyLinedFlowerBox)[0:3],
			"\n    ",
		) +
			"\n  ● double\n    " + strings.Join(
			cmdtoolkit.StyledFlowerBox([]string{"output ..."}, cmdtoolkit.DoubleLinedFlowerBox)[0:3],
			"\n    ",
		),
		RunE: aboutRun,
	}
	aboutFlags = &cmdtoolkit.FlagSet{
		Name: aboutCommand,
		Details: map[string]*cmdtoolkit.FlagDetails{
			aboutStyle: {
				Usage:        "specify the output border style",
				ExpectedType: cmdtoolkit.StringType,
				DefaultValue: "rounded",
			},
		},
	}
	cachedGoVersion           string
	cachedBuildDependencies   []string
	mp3repairElevationControl = cmdtoolkit.NewElevationControlWithEnvVar(
		ElevatedPrivilegesPermissionVar,
		DefaultElevatedPrivilegesPermission,
	)
)

func genAboutShortHelp(appName string) string {
	return "Provides information about the " + appName + " program"
}

func getAboutLongHelp(appName string) string {
	return fmt.Sprintf("%q", aboutCommand) +
		" provides the following information about the " + appName + " program:\n\n" +
		"• The program version and build timestamp\n" +
		"• Copyright information\n" +
		"• Build information:\n" +
		"  • The version of go used to compile the code\n" +
		"  • A list of dependencies and their versions\n" +
		"• The directory where log files are written\n" +
		"• The full path of the application configuration file and whether it exists\n" +
		"• Whether " + appName + " is running with elevated privileges, and, if not, why not"
}

func aboutRun(cmd *cobra.Command, _ []string) error {
	o := busGetter()
	values, eSlice := cmdtoolkit.ReadFlags(cmd.Flags(), aboutFlags)
	exitError := cmdtoolkit.NewExitProgrammingError(aboutCommand)
	if cmdtoolkit.ProcessFlagErrors(o, eSlice) {
		flag, err := cmdtoolkit.GetString(o, values, aboutStyle)
		if err == nil {
			o.ConsolePrintf(strings.Join(
				cmdtoolkit.StyledFlowerBox(acquireAboutData(o), interpretStyle(flag)),
				"\n",
			))
			exitError = nil
		}
	}
	return cmdtoolkit.ToErrorInterface(exitError)
}

func interpretStyle(flag cmdtoolkit.CommandFlag[string]) cmdtoolkit.FlowerBoxStyle {
	var style cmdtoolkit.FlowerBoxStyle
	switch strings.ToLower(flag.Value) {
	case "ascii":
		style = cmdtoolkit.ASCIIFlowerBox
	case "light":
		style = cmdtoolkit.LightLinedFlowerBox
	case "heavy":
		style = cmdtoolkit.HeavyLinedFlowerBox
	case "double":
		style = cmdtoolkit.DoubleLinedFlowerBox
	default:
		style = cmdtoolkit.CurvedFlowerBox
	}
	return style
}

func acquireAboutData(o output.Bus) []string {
	// 6: 1 each for
	// - app name
	// - copyright
	// - build information header
	// - go version
	// - log file location
	// - configuration file status
	elevationData := mp3repairElevationControl.Status(applicationName)
	lines := make([]string, 0, 6+len(cachedBuildDependencies)+len(elevationData))
	lines = append(lines,
		cmdtoolkit.DecoratedAppName(applicationName, version, creation),
		cmdtoolkit.Copyright(o, firstYear, creation, author),
		"Build Information",
		cmdtoolkit.FormatGoVersion(cachedGoVersion))
	lines = append(lines, cmdtoolkit.FormatBuildDependencies(cachedBuildDependencies)...)
	lines = append(lines, fmt.Sprintf("Log files are written to %s", logPath()))
	path, exists := cmdtoolkit.DefaultConfigFileStatus()
	switch {
	case exists:
		lines = append(lines, fmt.Sprintf("Configuration file %s exists", path))
	default:
		lines = append(lines, fmt.Sprintf("Configuration file %s does not yet exist", path))
	}
	lines = append(lines, elevationData[0])
	if len(elevationData) > 1 {
		for _, s := range elevationData[1:] {
			lines = append(lines, fmt.Sprintf(" - %s", s))
		}
	}
	return lines
}

func init() {
	rootCmd.AddCommand(aboutCmd)
	cmdtoolkit.AddDefaults(aboutFlags)
	cmdtoolkit.AddFlags(getBus(), getConfiguration(), aboutCmd.Flags(), aboutFlags)
}
