package files

import (
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
)

const (
	dirtyFileName = "metadata.dirty"
)

var (
	stateFileInitializationFailureLogged bool
	// a variable so testing can substitute another implementation
	initStateFile = cmdtoolkit.InitStateFile
)

// MarkDirty mark the system dirty
func MarkDirty(o output.Bus) {
	sf := safeStateFile(o)
	defer sf.Close()
	if err := sf.Create(dirtyFileName); err != nil {
		o.Log(output.Warning, "error creating dirty flag", map[string]any{"error": err})
	} else {
		o.Log(output.Info, "metadata dirty file created", map[string]any{"fileName": dirtyFileName})
	}
}

// ClearDirty clears the system dirty state
func ClearDirty(o output.Bus) {
	sf := safeStateFile(o)
	defer sf.Close()
	if err := sf.Remove(dirtyFileName); err != nil {
		o.Log(output.Warning, "error removing dirty flag", map[string]any{"error": err})
	} else {
		o.Log(output.Info, "metadata dirty file deleted", map[string]any{"fileName": dirtyFileName})
	}
}

// Dirty returns whether the system is dirty
func Dirty(o output.Bus) bool {
	sf := safeStateFile(o)
	defer sf.Close()
	return sf.Exists(dirtyFileName)
}

func safeStateFile(o output.Bus) cmdtoolkit.StateFile {
	if sf, err := initStateFile("mp3repair"); err != nil || sf == nil {
		if !stateFileInitializationFailureLogged {
			o.Log(output.Warning, "cannot initialize directory for dirty flag", map[string]any{"error": err})
			stateFileInitializationFailureLogged = true
		}
		return emergencyStateFile{}
	} else {
		return sf
	}
}

type emergencyStateFile struct{}

func (e emergencyStateFile) Read(_ string) ([]byte, error) {
	return []byte{}, nil
}

func (e emergencyStateFile) Write(_ string, _ []byte) error {
	return nil
}

func (e emergencyStateFile) Exists(_ string) bool {
	return false
}

func (e emergencyStateFile) Create(_ string) error {
	return nil
}

func (e emergencyStateFile) Remove(_ string) error {
	return nil
}

func (e emergencyStateFile) Close() {
}
