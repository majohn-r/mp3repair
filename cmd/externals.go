package cmd

// this file contains variables used to access external functions, allowing test
// code to easily override them
import (
	"mp3/internal/files"
	"os"
	"time"

	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"golang.org/x/sys/windows/svc/mgr"
)

var (
	ApplicationPath         = cmd_toolkit.ApplicationPath
	AppNameGetFunc          = cmd_toolkit.AppName
	BuildDependenciesFunc   = cmd_toolkit.BuildDependencies
	CopyFile                = cmd_toolkit.CopyFile
	DereferenceEnvVarFunc   = cmd_toolkit.DereferenceEnvVar
	DirExists               = cmd_toolkit.DirExists
	AboutContentGenerator   = cmd_toolkit.GenerateAboutContent
	GoVersionFunc           = cmd_toolkit.GoVersion
	AppPathInitFunc         = cmd_toolkit.InitApplicationPath
	InitBuildDataFunc       = cmd_toolkit.InitBuildData
	LogInitFunc             = cmd_toolkit.InitLogging
	CommandStartLogger      = cmd_toolkit.LogCommandStart
	MkDir                   = cmd_toolkit.Mkdir
	PlainFileExists         = cmd_toolkit.PlainFileExists
	ReadConfigFileFunc      = cmd_toolkit.ReadConfigurationFile
	ReadDir                 = cmd_toolkit.ReadDirectory
	AppNameSetFunc          = cmd_toolkit.SetAppName
	FirstYearSetFunc        = cmd_toolkit.SetFirstYear
	FlagIndicatorSetFunc    = cmd_toolkit.SetFlagIndicator
	ClearDirty              = files.ClearDirty
	Dirty                   = files.Dirty
	MarkDirty               = files.MarkDirty
	MetadataReader          = files.ReadMetadata
	ConnectToServiceManager = mgr.Connect
	ExitFunction            = os.Exit
	FileRename              = os.Rename
	FileRemove              = os.Remove
	RemoveAll               = os.RemoveAll
	FileWrite               = os.WriteFile
	NewBusFunc              = output.NewDefaultBus
	DurationCalc            = time.Since
)
