package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// AppName is the name of the application
const AppName = "mp3"

// CreateAppSpecificPath creates a path string for an app-related directory
func CreateAppSpecificPath(topDir string) string {
	return filepath.Join(topDir, AppName)
}

func dereferenceEnvVar(s string) (string, error) {
	refs := findReferences(s)
	if len(refs) == 0 {
		return s, nil
	}
	var missingVars []string
	for _, r := range refs {
		envVar := ""
		if strings.HasPrefix(r, "$") {
			envVar = r[1:]
		} else {
			envVar = r[1 : len(r)-1]
		}
		if value, ok := os.LookupEnv(envVar); !ok {
			missingVars = append(missingVars, envVar)
		} else {
			s = strings.ReplaceAll(s, r, value)
		}
	}
	if len(missingVars) > 0 {
		sort.Strings(missingVars)
		return "", fmt.Errorf("missing environment variables: %v", missingVars)
	}
	return s, nil
}

var (
	unixPattern    = regexp.MustCompile(`[$][a-zA-Z_]+[a-zA-Z0-9_]*`)
	windowsPattern = regexp.MustCompile(`%[a-zA-Z_]+[a-zA-Z0-9_]*%`)
)

type byLength []string

func (bl byLength) Len() int {
	return len(bl)
}

func (bl byLength) Less(i, j int) bool {
	return len(bl[i]) > len(bl[j])
}

func (bl byLength) Swap(i, j int) {
	bl[i], bl[j] = bl[j], bl[i]
}

func findReferences(s string) []string {
	matches := unixPattern.FindAllString(s, -1)
	if len(matches) > 1 {
		sort.Sort(byLength(matches))
	}
	matches = append(matches, windowsPattern.FindAllString(s, -1)...)
	return matches
}
