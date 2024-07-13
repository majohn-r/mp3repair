package cmd

import (
	"fmt"
	"runtime/debug"
	"strings"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

const (
	aboutCommand   = "about"
	aboutStyle     = "style"
	aboutStyleFlag = "--" + aboutStyle
	appName        = "mp3repair" // the name of the application
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
		Short:                 "Provides information about the " + appName + " program",
		Long: fmt.Sprintf("%q", aboutCommand) +
			` provides the following information about the ` + appName + ` program:

• The program version and build timestamp
• Copyright information
• Build information:
  • The version of go used to compile the code
  • A list of dependencies and their versions
• The directory where log files are written
• The full path of the application configuration file and whether it exists
• Whether ` + appName + ` is running with elevated privileges, and, if not, why not`,
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

func aboutRun(cmd *cobra.Command, _ []string) error {
	o := BusGetter()
	values, eSlice := cmdtoolkit.ReadFlags(cmd.Flags(), aboutFlags)
	exitError := cmdtoolkit.NewExitProgrammingError(aboutCommand)
	if cmdtoolkit.ProcessFlagErrors(o, eSlice) {
		flag, err := cmdtoolkit.GetString(o, values, aboutStyle)
		if err == nil {
			style := interpretStyle(flag)
			logCommandStart(o, aboutCommand, map[string]any{aboutStyle: style})
			o.WriteConsole(strings.Join(cmdtoolkit.StyledFlowerBox(acquireAboutData(o), style), "\n"))
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
	cachedGoVersion, cachedBuildDependencies = interpretBuildData(debug.ReadBuildInfo)
	// 9: 1 each for
	// - app name
	// - copyright
	// - build information header
	// - go version
	// - log file location
	// - configuration file status
	// - and up to 3 for elevation status
	lines := make([]string, 0, 9+len(cachedBuildDependencies))
	lines = append(lines,
		cmdtoolkit.DecoratedAppName(appName, version, creation),
		cmdtoolkit.Copyright(o, firstYear, creation, author),
		"Build Information",
		cmdtoolkit.FormatGoVersion(cachedGoVersion))
	lines = append(lines, cmdtoolkit.FormatBuildDependencies(cachedBuildDependencies)...)
	lines = append(lines, fmt.Sprintf("Log files are written to %s", logPath()))
	path, exists := configFile()
	switch {
	case exists:
		lines = append(lines, fmt.Sprintf("Configuration file %s exists", path))
	default:
		lines = append(lines, fmt.Sprintf("Configuration file %s does not yet exist", path))
	}
	elevationData := mp3repairElevationControl.Status(appName)
	lines = append(lines, elevationData[0])
	if len(elevationData) > 1 {
		for _, s := range elevationData[1:] {
			lines = append(lines, fmt.Sprintf(" - %s", s))
		}
	}
	return lines
}

func init() {
	RootCmd.AddCommand(aboutCmd)
	addDefaults(aboutFlags)
	o := getBus()
	c := getConfiguration()
	cmdtoolkit.AddFlags(o, c, aboutCmd.Flags(), aboutFlags)
}
