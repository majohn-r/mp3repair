package internal

import (
	"errors"
	"os"
	"testing"
)

func TestLookupEnvVars(t *testing.T) {
	fnName := "LookupEnvVars()"
	type envState struct {
		varName  string
		varValue string
		varSet   bool
	}
	var savedStates []envState
	for _, name := range []string{"TMP", "TEMP", "HOMEPATH", "APPDATA"} {
		value, set := os.LookupEnv(name)
		savedStates = append(savedStates, envState{varName: name, varValue: value, varSet: set})
	}
	var savedTmpFolder = TmpFolder
	var savedHomePath = HomePath
	var savedAppDataPath = AppDataPath
	defer func() {
		for _, ss := range savedStates {
			if ss.varSet {
				os.Setenv(ss.varName, ss.varValue)
			} else {
				os.Unsetenv(ss.varName)
			}
		}
		TmpFolder = savedTmpFolder
		HomePath = savedHomePath
		AppDataPath = savedAppDataPath
	}()
	tests := []struct {
		name            string
		envs            []envState
		wantTmpFolder   string
		wantHomePath    string
		wantAppDataPath string
		wantErrors      []error
	}{
		{
			name: "expected use case",
			envs: []envState{
				{varName: "TMP", varValue: "/tmp", varSet: true},
				{varName: "TEMP", varValue: "/tmp2", varSet: true},
				{varName: "HOMEPATH", varValue: "/users/myUser", varSet: true},
				{varName: "APPDATA", varValue: "/users/myUser/AppData/Roaming", varSet: true},
			},
			wantTmpFolder:   "/tmp",
			wantHomePath:    "/users/myUser",
			wantAppDataPath: "/users/myUser/AppData/Roaming",
		},
		{
			name: "missing TMP",
			envs: []envState{
				{varName: "TMP"},
				{varName: "TEMP", varValue: "/tmp2", varSet: true},
				{varName: "HOMEPATH", varValue: "/users/myUser", varSet: true},
				{varName: "APPDATA", varValue: "/users/myUser/AppData/Roaming", varSet: true},
			},
			wantTmpFolder:   "/tmp2",
			wantHomePath:    "/users/myUser",
			wantAppDataPath: "/users/myUser/AppData/Roaming",
		},
		{
			name: "missing TMP and TEMP",
			envs: []envState{
				{varName: "TMP"},
				{varName: "TEMP"},
				{varName: "HOMEPATH", varValue: "/users/myUser", varSet: true},
				{varName: "APPDATA", varValue: "/users/myUser/AppData/Roaming", varSet: true},
			},
			wantHomePath:    "/users/myUser",
			wantAppDataPath: "/users/myUser/AppData/Roaming",
			wantErrors:      []error{noTempFolder},
		},
		{
			name: "missing homepath",
			envs: []envState{
				{varName: "TMP", varValue: "/tmp", varSet: true},
				{varName: "TEMP", varValue: "/tmp2", varSet: true},
				{varName: "HOMEPATH"},
				{varName: "APPDATA", varValue: "/users/myUser/AppData/Roaming", varSet: true},
			},
			wantTmpFolder:   "/tmp",
			wantAppDataPath: "/users/myUser/AppData/Roaming",
			wantErrors:      []error{noHomePath},
		},
		{
			name: "missing appDataPath",
			envs: []envState{
				{varName: "TMP", varValue: "/tmp", varSet: true},
				{varName: "TEMP", varValue: "/tmp2", varSet: true},
				{varName: "HOMEPATH", varValue: "/users/myUser", varSet: true},
				{varName: "APPDATA"},
			},
			wantTmpFolder: "/tmp",
			wantHomePath:  "/users/myUser",
			wantErrors:    []error{noAppDataPath},
		},
		{
			name: "missing all vars",
			envs: []envState{
				{varName: "TMP"},
				{varName: "TEMP"},
				{varName: "HOMEPATH"},
				{varName: "APPDATA"},
			},
			wantErrors: []error{noTempFolder, noHomePath, noAppDataPath},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// clear initial state
			TmpFolder = ""
			HomePath = ""
			AppDataPath = ""
			for _, env := range tt.envs {
				if env.varSet {
					os.Setenv(env.varName, env.varValue)
				} else {
					os.Unsetenv(env.varName)
				}
			}
			if gotErrors := LookupEnvVars(); !equalErrorSlices(gotErrors, tt.wantErrors) {
				t.Errorf("%s errors = %v, want %v", fnName, gotErrors, tt.wantErrors)
			}
			if TmpFolder != tt.wantTmpFolder {
				t.Errorf("%s TmpFolder = %v, want %v", fnName, TmpFolder, tt.wantTmpFolder)
			}
			if HomePath != tt.wantHomePath {
				t.Errorf("%s HomePath = %v, want %v", fnName, HomePath, tt.wantHomePath)
			}
			if AppDataPath != tt.wantAppDataPath {
				t.Errorf("%s AppDataPath = %v, want %v", fnName, AppDataPath, tt.wantAppDataPath)
			}
		})
	}
}

func equalErrorSlices(got []error, want []error) bool {
	if len(got) != len(want) {
		return false
	}
	for k := 0; k < len(got); k++ {
		if !errors.Is(got[k], want[k]) {
			return false
		}
	}
	return true
}
