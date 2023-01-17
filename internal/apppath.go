package internal

import (
	"os"

	"github.com/majohn-r/output"
)

var applicationPath string

const appDataVar = "APPDATA"

// InitApplicationPath ensures that the application path exists
func InitApplicationPath(o output.Bus) (initialized bool) {
	if dir, ok := appDataValue(o); ok {
		applicationPath = CreateAppSpecificPath(dir)
		if DirExists(applicationPath) {
			initialized = true
		} else {
			if err := Mkdir(applicationPath); err == nil {
				initialized = true
			} else {
				WriteDirectoryCreationError(o, applicationPath, err)
				o.Log(output.Error, "cannot create directory", map[string]any{
					"directory": applicationPath,
					"error":     err,
				})
			}
		}
	}
	return
}

// ApplicationPath returns the path to application-specific data (%APPDATA%\mp3)
func ApplicationPath() string {
	return applicationPath
}

// SetApplicationPathForTesting is used strictly for testing to set applicationPath to a known value
func SetApplicationPathForTesting(s string) (previous string) {
	previous = applicationPath
	applicationPath = s
	return
}

func appDataValue(o output.Bus) (string, bool) {
	if value, ok := os.LookupEnv(appDataVar); ok {
		return value, ok
	}
	o.Log(output.Info, "not set", map[string]any{"environmentVariable": appDataVar})
	return "", false
}
