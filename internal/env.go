package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const appName = "mp3"

// CreateAppSpecificPath creates a path string for an app-related directory
func CreateAppSpecificPath(topDir string) string {
	return filepath.Join(topDir, appName)
}

var (
	AppDataPath   string
	TmpFolder     string
	noTempFolder  error = fmt.Errorf(LOG_NO_TEMP_DIRECTORY)
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
	if AppDataPath, found = os.LookupEnv("APPDATA"); !found {
		errors = append(errors, noAppDataPath)
	}
	return
}

func InterpretEnvVarReferences(s string) string {
	references := findReferences(s)
	if len(references) == 0 {
		return s
	}
	for _, r := range references {
		envVar := ""
		if strings.HasPrefix(r, "$") {
			envVar = r[1:]
		} else {
			envVar = r[1 : len(r)-1]
		}
		s = strings.ReplaceAll(s, r, os.Getenv(envVar))
	}
	return s
}

var (
	unixPattern    = regexp.MustCompile(`[$][a-zA-Z_]{1,}[a-zA-Z0-9_]{0,}`)
	windowsPattern = regexp.MustCompile(`%[a-zA-Z_]{1,}[a-zA-Z0-9_]{0,}%`)
)

type byLength []string

func (a byLength) Len() int {
	return len(a)
}

func (a byLength) Less(i, j int) bool {
	return len(a[i]) > len(a[j])
}

func (a byLength) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func findReferences(s string) []string {
	matches := unixPattern.FindAllString(s, -1)
	if len(matches) > 1 {
		sort.Sort(byLength(matches))
	}
	matches = append(matches, windowsPattern.FindAllString(s, -1)...)
	return matches
}
