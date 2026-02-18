/*
Copyright © 2026 Marc Johnson (marc.johnson27591@gmail.com)
*/
package files

import (
	"errors"
	"reflect"
	"testing"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
)

func Test_emergencyStateFile_Close(t *testing.T) {
	// does no harm
	emergencyStateFile{}.Close()
}

func Test_emergencyStateFile_Remove(t *testing.T) {
	tests := map[string]struct {
		in0     string
		wantErr bool
	}{
		"test": {in0: "", wantErr: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := emergencyStateFile{}
			if err := e.Remove(tt.in0); (err != nil) != tt.wantErr {
				t.Errorf("Remove() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_emergencyStateFile_Create(t *testing.T) {
	tests := map[string]struct {
		in0     string
		wantErr bool
	}{
		"test": {in0: "", wantErr: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := emergencyStateFile{}
			if err := e.Create(tt.in0); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_emergencyStateFile_Exists(t *testing.T) {
	tests := map[string]struct {
		in0  string
		want bool
	}{
		"test": {in0: "", want: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := emergencyStateFile{}
			if got := e.Exists(tt.in0); got != tt.want {
				t.Errorf("Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_emergencyStateFile_Write(t *testing.T) {
	type args struct {
		in0 string
		in1 []byte
	}
	tests := map[string]struct {
		args    args
		wantErr bool
	}{
		"test": {args: args{in0: "", in1: nil}, wantErr: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := emergencyStateFile{}
			if err := e.Write(tt.args.in0, tt.args.in1); (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_emergencyStateFile_Read(t *testing.T) {
	tests := map[string]struct {
		in0     string
		want    []byte
		wantErr bool
	}{
		"test": {in0: "", want: []byte{}, wantErr: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := emergencyStateFile{}
			got, err := e.Read(tt.in0)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Read() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type internalSF struct {
	exists      bool
	readData    []byte
	readError   error
	writeError  error
	createError error
	removeError error
}

func (i *internalSF) Read(_ string) ([]byte, error) {
	return i.readData, i.readError
}

func (i *internalSF) Write(_ string, _ []byte) error {
	return i.writeError
}

func (i *internalSF) Exists(_ string) bool {
	return i.exists
}

func (i *internalSF) Create(_ string) error {
	return i.createError
}

func (i *internalSF) Remove(_ string) error {
	return i.removeError
}

func (i *internalSF) Close() {
}

func Test_safeStateFile(t *testing.T) {
	originalInitStateFile := initStateFile
	originalStateFileInitializationFailureLogged := stateFileInitializationFailureLogged
	defer func() {
		initStateFile = originalInitStateFile
		stateFileInitializationFailureLogged = originalStateFileInitializationFailureLogged
	}()
	tests := map[string]struct {
		logged bool
		init   func(string) (cmdtoolkit.StateFile, error)
		want   cmdtoolkit.StateFile
		output.WantedRecording
	}{
		"first error on init": {
			logged: false,
			init: func(string) (cmdtoolkit.StateFile, error) {
				return nil, errors.New("first error")
			},
			want: emergencyStateFile{},
			WantedRecording: output.WantedRecording{
				Log: "level='warning' " +
					"error='first error' " +
					"msg='cannot initialize directory for dirty flag'\n",
			},
		},
		"second error on init": {
			logged: true,
			init: func(string) (cmdtoolkit.StateFile, error) {
				return nil, errors.New("first error")
			},
			want:            emergencyStateFile{},
			WantedRecording: output.WantedRecording{},
		},
		"first nil on init": {
			logged: false,
			init: func(string) (cmdtoolkit.StateFile, error) {
				return nil, nil
			},
			want: emergencyStateFile{},
			WantedRecording: output.WantedRecording{
				Log: "level='warning' " +
					"error='<nil>' " +
					"msg='cannot initialize directory for dirty flag'\n",
			},
		},
		"second nil on init": {
			logged: true,
			init: func(string) (cmdtoolkit.StateFile, error) {
				return nil, nil
			},
			want:            emergencyStateFile{},
			WantedRecording: output.WantedRecording{},
		},
		"success": {
			logged: false,
			init: func(string) (cmdtoolkit.StateFile, error) {
				return &internalSF{}, nil
			},
			want:            &internalSF{},
			WantedRecording: output.WantedRecording{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			stateFileInitializationFailureLogged = tt.logged
			initStateFile = tt.init
			if got := safeStateFile(o); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("safeStateFile() = %v, want %v", got, tt.want)
			}
			o.Report(t, name, tt.WantedRecording)
		})
	}
}

func TestDirty(t *testing.T) {
	originalInitStateFile := initStateFile
	defer func() {
		initStateFile = originalInitStateFile
	}()
	tests := map[string]struct {
		init func(string) (cmdtoolkit.StateFile, error)
		want bool
		output.WantedRecording
	}{
		"exists": {
			init: func(string) (cmdtoolkit.StateFile, error) {
				return &internalSF{exists: true}, nil
			},
			want: true,
		},
		"doesn't exist": {
			init: func(string) (cmdtoolkit.StateFile, error) {
				return &internalSF{exists: false}, nil
			},
			want: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			initStateFile = tt.init
			o := output.NewRecorder()
			if got := Dirty(o); got != tt.want {
				t.Errorf("Dirty() = %v, want %v", got, tt.want)
			}
			o.Report(t, name, tt.WantedRecording)
		})
	}
}

func TestClearDirty(t *testing.T) {
	originalInitStateFile := initStateFile
	defer func() {
		initStateFile = originalInitStateFile
	}()
	tests := map[string]struct {
		init func(string) (cmdtoolkit.StateFile, error)
		output.WantedRecording
	}{
		"error": {
			init: func(string) (cmdtoolkit.StateFile, error) {
				return &internalSF{removeError: errors.New("remove error")}, nil
			},
			WantedRecording: output.WantedRecording{
				Log: "level='warning' " +
					"error='remove error' " +
					"msg='error removing dirty flag'\n",
			},
		},
		"ok": {
			init: func(string) (cmdtoolkit.StateFile, error) {
				return &internalSF{removeError: nil}, nil
			},
			WantedRecording: output.WantedRecording{
				Log: "level='info' " +
					"fileName='metadata.dirty' " +
					"msg='metadata dirty file deleted'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			initStateFile = tt.init
			ClearDirty(o)
			o.Report(t, name, tt.WantedRecording)
		})
	}
}

func TestMarkDirty(t *testing.T) {
	originalInitStateFile := initStateFile
	defer func() {
		initStateFile = originalInitStateFile
	}()
	tests := map[string]struct {
		init func(string) (cmdtoolkit.StateFile, error)
		output.WantedRecording
	}{
		"error": {
			init: func(string) (cmdtoolkit.StateFile, error) {
				return &internalSF{createError: errors.New("create error")}, nil
			},
			WantedRecording: output.WantedRecording{
				Log: "level='warning' " +
					"error='create error' " +
					"msg='error creating dirty flag'\n",
			},
		},
		"success": {
			init: func(string) (cmdtoolkit.StateFile, error) {
				return &internalSF{createError: nil}, nil
			},
			WantedRecording: output.WantedRecording{
				Log: "level='info' " +
					"fileName='metadata.dirty' " +
					"msg='metadata dirty file created'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			initStateFile = tt.init
			o := output.NewRecorder()
			MarkDirty(o)
			o.Report(t, name, tt.WantedRecording)
		})
	}
}
