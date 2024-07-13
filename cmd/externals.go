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
	applicationPath       = cmdtoolkit.ApplicationPath
	copyFile              = cmdtoolkit.CopyFile
	dereferenceEnvVar     = cmdtoolkit.DereferenceEnvVar
	dirExists             = cmdtoolkit.DirExists
	initApplicationPath   = cmdtoolkit.InitApplicationPath
	initLogging           = cmdtoolkit.InitLogging
	interpretBuildData    = cmdtoolkit.InterpretBuildData
	logCommandStart       = cmdtoolkit.LogCommandStart
	logPath               = cmdtoolkit.LogPath
	mkdir                 = cmdtoolkit.Mkdir
	plainFileExists       = cmdtoolkit.PlainFileExists
	processIsElevated     = cmdtoolkit.ProcessIsElevated
	readConfigurationFile = cmdtoolkit.ReadConfigurationFile
	readDirectory         = cmdtoolkit.ReadDirectory
	clearDirty            = files.ClearDirty
	dirty                 = files.Dirty
	markDirty             = files.MarkDirty
	readMetadata          = files.ReadMetadata
	connect               = mgr.Connect
	Exit                  = os.Exit
	rename                = os.Rename
	remove                = os.Remove
	removeAll             = os.RemoveAll
	writeFile             = os.WriteFile
	newDefaultBus         = output.NewDefaultBus
	since                 = time.Since
)
