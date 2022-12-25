package internal

import "github.com/majohn-r/output"

// ReportInvalidConfigurationData handles errors found when attempting to parse
// a YAML configuration file, both logging the error and notifying the user of
// the error
func ReportInvalidConfigurationData(o output.Bus, s string, e error) {
	o.WriteCanonicalError("The configuration file %q contains an invalid value for %q: %v", DefaultConfigFileName, s, e)
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

// WriteDirectoryCreationError writes a suitable error message to the user when
// a directory cannot be created
func WriteDirectoryCreationError(o output.Bus, d string, e error) {
	o.WriteCanonicalError("The directory %q cannot be created: %v", d, e)
}
