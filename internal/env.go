package internal

import (
	"fmt"
	"os"
)

var (
	AppDataPath   string
	HomePath      string
	TmpFolder     string
	noTempFolder  error = fmt.Errorf(LOG_NO_TEMP_DIRECTORY)
	noHomePath    error = fmt.Errorf(LOG_NO_HOME_PATH)
	noAppDataPath error = fmt.Errorf(LOG_NO_APP_DATA_PATH)
)

func LookupEnvVars() (errors []error) {
	var found bool
	// get temporary folder
	if TmpFolder, found = os.LookupEnv("TMP"); !found {
		if TmpFolder, found = os.LookupEnv("TEMP"); !found {
			errors = append(errors, noTempFolder)
		}
	}
	// get homepath
	if HomePath, found = os.LookupEnv("HOMEPATH"); !found {
		errors = append(errors, noHomePath)
	}
	if AppDataPath, found = os.LookupEnv("APPDATA"); !found {
		errors = append(errors, noAppDataPath)
	}
	return
}
