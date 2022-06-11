package files

import "path/filepath"

const backupDirName = "pre-repair-backup"

// CreateBackupPath creates a backup directory path based on the provided top
// directory.
func CreateBackupPath(topDir string) string {
	return filepath.Join(topDir, backupDirName)
}

// per https://docs.microsoft.com/en-us/windows/win32/fileio/naming-a-file
func isIllegalRuneForFileNames(r rune) bool {
	if r >= 0 && r <= 31 {
		return true
	}
	switch r {
	case '<', '>', ':', '"', '/', '\\', '|', '?', '*':
		return true
	default:
		return false
	}
}
