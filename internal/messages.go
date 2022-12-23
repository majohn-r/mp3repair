package internal

import "github.com/majohn-r/output"

// LogInvalidConfigurationData logs errors found when attempting to parse a YAML
// configuration file
func LogInvalidConfigurationData(o output.Bus, s string, e error) {
	o.Log(output.Error, "invalid content in configuration file", map[string]any{
		"section": s,
		"error":   e,
	})
}

// LogUnreadableDirectory logs errors when a directory cannot be read
func LogUnreadableDirectory(o output.Bus, s string, e error) {
	o.Log(output.Error, "cannot read directory", map[string]any{
		"directory": s,
		"error":     e,
	})
}

// LogFileDeletionFailure logs errors when a file cannot be deleted
func LogFileDeletionFailure(o output.Bus, s string, e error) {
	o.Log(output.Error, "cannot delete file", map[string]any{
		"fileName": s,
		"error":    e,
	})
}

// for output to user
const (
	UserAmbiguousChoices         = "There are multiple %s fields for %q, and there is no unambiguously preferred choice; candidates are %v"
	UserCannotCreateDirectory    = "The directory %q cannot be created: %v"
	UserCannotCreateFile         = "The file %q cannot be created: %v"
	UserCannotDeleteFile         = "The file %q cannot be deleted: %v"
	UserCannotQueryService       = "The status for the service %q cannot be obtained: %v"
	UserConfigurationFileInvalid = "The configuration file %q contains an invalid value for %q: %v"
	UserID3v1TagError            = "An error occurred when trying to read ID3V1 tag information for track %q on album %q by artist %q: %q"
	UserID3v2TagError            = "An error occurred when trying to read ID3V2 tag information for track %q on album %q by artist %q: %q"
	UserNoMusicFilesFound        = "No music files could be found using the specified parameters"
	UserSpecifiedNoWork          = "You disabled all functionality for the command %q"
)

// these constants are errors to be used
const (
	ErrorDoesNotBeginWithDigit = "first character is not a digit"
	ErrorEditUnnecessary       = "no edit required"
)
