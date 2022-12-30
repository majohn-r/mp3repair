package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/majohn-r/output"
)

func TestInitApplicationPath(t *testing.T) {
	const fnName = "InitApplicationPath()"
	savedAppData := SaveEnvVarForTesting("APPDATA")
	here := SecureAbsolutePathForTesting(".")
	preExistingPath := "test"
	if err := Mkdir(preExistingPath); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, preExistingPath, err)
	}
	preExistingMp3 := filepath.Join(preExistingPath, "mp3")
	if err := Mkdir(preExistingMp3); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, preExistingMp3, err)
	}
	unfortunatePath := "test2"
	if err := Mkdir(unfortunatePath); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, unfortunatePath, err)
	}
	if err := CreateFileForTesting(unfortunatePath, "mp3"); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, filepath.Join(unfortunatePath, "mp3"), err)
	}
	defer func() {
		DestroyDirectoryForTesting(fnName, "mp3")
		DestroyDirectoryForTesting(fnName, preExistingPath)
		DestroyDirectoryForTesting(fnName, unfortunatePath)
		savedAppData.RestoreForTesting()
	}()
	tests := []struct {
		name            string
		appData         *SavedEnvVar
		wantInitialized bool
		wantPath        string
		output.WantedRecording
	}{
		{
			name:            "happy path",
			appData:         &SavedEnvVar{Name: "APPDATA", Value: here, Set: true},
			wantInitialized: true,
			wantPath:        filepath.Join(here, "mp3"),
		},
		{
			name:    "missing appdata setting",
			appData: &SavedEnvVar{Name: "APPDATA"},
			WantedRecording: output.WantedRecording{
				Log: "level='info' environmentVariable='APPDATA' msg='not set'\n",
			},
		},
		{
			name:            "path exists",
			appData:         &SavedEnvVar{Name: "APPDATA", Value: preExistingPath, Set: true},
			wantInitialized: true,
			wantPath:        preExistingMp3,
		},
		{
			name:    "file blocks path creation",
			appData: &SavedEnvVar{Name: "APPDATA", Value: unfortunatePath, Set: true},
			WantedRecording: output.WantedRecording{
				Error: "The directory \"test2\\\\mp3\" cannot be created: file exists and is not a directory.\n",
				Log:   "level='error' directory='test2\\mp3' error='file exists and is not a directory' msg='cannot create directory'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.appData.RestoreForTesting()
			o := output.NewRecorder()
			if gotInitialized := InitApplicationPath(o); gotInitialized != tt.wantInitialized {
				t.Errorf("%s = %v, want %v", fnName, gotInitialized, tt.wantInitialized)
			}
			if tt.wantInitialized {
				if gotPath := ApplicationPath(); gotPath != tt.wantPath {
					t.Errorf("%s = %s, want %s", fnName, gotPath, tt.wantPath)
				}
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func TestSetApplicationPathForTesting(t *testing.T) {
	old := ApplicationPath()
	defer SetApplicationPathForTesting(old)
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args
		wantPrevious string
	}{
		{
			name:         "trivial test",
			args:         args{s: old + "1"},
			wantPrevious: old,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotPrevious := SetApplicationPathForTesting(tt.args.s); gotPrevious != tt.wantPrevious {
				t.Errorf("SetApplicationPathForTesting() = %v, want %v", gotPrevious, tt.wantPrevious)
			}
		})
	}
}

func Test_lookupAppData(t *testing.T) {
	fnName := "lookupAppData()"
	savedState := SaveEnvVarForTesting(appDataVar)
	os.Setenv(appDataVar, SecureAbsolutePathForTesting("."))
	defer func() {
		savedState.RestoreForTesting()
	}()
	tests := []struct {
		name  string
		state *SavedEnvVar
		want  string
		want1 bool
		output.WantedRecording
	}{
		{
			name:            "value is set",
			state:           &SavedEnvVar{Name: appDataVar, Value: "appData!", Set: true},
			want:            "appData!",
			want1:           true,
			WantedRecording: output.WantedRecording{},
		},
		{
			name:  "value is not set",
			state: &SavedEnvVar{Name: appDataVar},
			WantedRecording: output.WantedRecording{
				Log: "level='info' environmentVariable='APPDATA' msg='not set'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.state.RestoreForTesting()
			o := output.NewRecorder()
			got, got1 := lookupAppData(o)
			if got != tt.want {
				t.Errorf("%s got = %q, want %q", fnName, got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("%s got1 = %v, want %v", fnName, got1, tt.want1)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}
