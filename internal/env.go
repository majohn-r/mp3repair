package internal

import (
	"fmt"
	"os"
)

var (
	HomePath  string
	TmpFolder string
)

func LookupEnvVars() []error {
	var errors []error
	var found bool
	// get temporary folder
	if TmpFolder, found = os.LookupEnv("TMP"); !found {
		if TmpFolder, found = os.LookupEnv("TEMP"); !found {
			errors = append(errors, fmt.Errorf("no temporary folder defined, checked TMP and TEMP"))
		}
	}
	// get homepath
	if HomePath, found = os.LookupEnv("HOMEPATH"); !found {
		errors = append(errors, fmt.Errorf("no home path defined, checked HOMEPATH"))
	}
	return errors
}