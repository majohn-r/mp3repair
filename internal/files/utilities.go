package files

const backupDirName = "pre-repair-backup"

// per https://docs.microsoft.com/en-us/windows/win32/fileio/naming-a-file
func IsIllegalRuneForFileNames(r rune) bool {
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
