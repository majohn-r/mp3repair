package cmd

// this file contains variables used to access external functions, allowing test
// code to easily override them
import (
	"fmt"
	"mp3repair/internal/files"
	"os"
	"time"

	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/mattn/go-isatty"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

var (
	ApplicationPath        = cmd_toolkit.ApplicationPath
	AppName                = cmd_toolkit.AppName
	BuildDependencies      = cmd_toolkit.BuildDependencies
	CopyFile               = cmd_toolkit.CopyFile
	DereferenceEnvVar      = cmd_toolkit.DereferenceEnvVar
	DirExists              = cmd_toolkit.DirExists
	GenerateAboutContent   = cmd_toolkit.GenerateAboutContent
	GoVersion              = cmd_toolkit.GoVersion
	InitApplicationPath    = cmd_toolkit.InitApplicationPath
	InitBuildData          = cmd_toolkit.InitBuildData
	InitLogging            = cmd_toolkit.InitLogging
	InterpretBuildData     = cmd_toolkit.InterpretBuildData
	LogCommandStart        = cmd_toolkit.LogCommandStart
	LogPath                = cmd_toolkit.LogPath
	Mkdir                  = cmd_toolkit.Mkdir
	PlainFileExists        = cmd_toolkit.PlainFileExists
	ReadConfigurationFile  = cmd_toolkit.ReadConfigurationFile
	ReadDirectory          = cmd_toolkit.ReadDirectory
	SetAppName             = cmd_toolkit.SetAppName
	SetFirstYear           = cmd_toolkit.SetFirstYear
	SetFlagIndicator       = cmd_toolkit.SetFlagIndicator
	ClearDirty             = files.ClearDirty
	Dirty                  = files.Dirty
	MarkDirty              = files.MarkDirty
	ReadMetadata           = files.ReadMetadata
	Scanf                  = fmt.Scanf
	IsCygwinTerminal       = isatty.IsCygwinTerminal
	IsTerminal             = isatty.IsTerminal
	Connect                = mgr.Connect
	Exit                   = os.Exit
	LookupEnv              = os.LookupEnv
	Rename                 = os.Rename // TODO: replace with afero
	Remove                 = os.Remove // TODO: replace with afero
	RemoveAll              = os.RemoveAll // TODO: replace with afero
	WriteFile              = os.WriteFile // TODO: replace with afero
	NewDefaultBus          = output.NewDefaultBus
	Since                  = time.Since
	GetCurrentProcessToken = windows.GetCurrentProcessToken
	ShellExecute           = windows.ShellExecute
	IsElevated             = windows.Token.IsElevated
)
