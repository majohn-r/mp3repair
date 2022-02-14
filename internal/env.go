package internal

import (
	"fmt"
	"os"
)

var (
	HomePath     string
	TmpFolder    string
	noTempFolder error = fmt.Errorf("no temporary folder defined, checked TMP and TEMP")
	noHomePath   error = fmt.Errorf("no home path defined, checked HOMEPATH")
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
	return
}
