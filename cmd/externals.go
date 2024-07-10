package cmd

// this file contains variables used to access external functions, allowing test
// code to easily override them
import (
	"mp3repair/internal/files"
	"os"
	"time"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"golang.org/x/sys/windows/svc/mgr"
)

var (
	ApplicationPath       = cmdtoolkit.ApplicationPath
	CopyFile              = cmdtoolkit.CopyFile
	DereferenceEnvVar     = cmdtoolkit.DereferenceEnvVar
	DirExists             = cmdtoolkit.DirExists
	InitApplicationPath   = cmdtoolkit.InitApplicationPath
	InitLogging           = cmdtoolkit.InitLogging
	InterpretBuildData    = cmdtoolkit.InterpretBuildData
	LogCommandStart       = cmdtoolkit.LogCommandStart
	LogPath               = cmdtoolkit.LogPath
	Mkdir                 = cmdtoolkit.Mkdir
	PlainFileExists       = cmdtoolkit.PlainFileExists
	ReadConfigurationFile = cmdtoolkit.ReadConfigurationFile
	ReadDirectory         = cmdtoolkit.ReadDirectory
	ClearDirty            = files.ClearDirty
	Dirty                 = files.Dirty
	MarkDirty             = files.MarkDirty
	ReadMetadata          = files.ReadMetadata
	Connect               = mgr.Connect
	Exit                  = os.Exit
	Rename                = os.Rename
	Remove                = os.Remove
	RemoveAll             = os.RemoveAll
	WriteFile             = os.WriteFile
	NewDefaultBus         = output.NewDefaultBus
	Since                 = time.Since
)
