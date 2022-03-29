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
	value, set := os.LookupEnv("TMP")
	savedStates = append(savedStates, envState{varName: "TMP", varValue: value, varSet: set})
	value, set = os.LookupEnv("TEMP")
	savedStates = append(savedStates, envState{varName: "TEMP", varValue: value, varSet: set})
	value, set = os.LookupEnv("HOMEPATH")
	savedStates = append(savedStates, envState{varName: "HOMEPATH", varValue: value, varSet: set})
	var savedTmpFolder = TmpFolder
	var savedHomePath = HomePath
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
	}()
	tests := []struct {
		name          string
		envs          []envState
		wantTmpFolder string
		wantHomePath  string
		wantErrors    []error
	}{
		{
			name: "expected use case",
			envs: []envState{
				{varName: "TMP", varValue: "/tmp", varSet: true},
				{varName: "TEMP", varValue: "/tmp2", varSet: true},
				{varName: "HOMEPATH", varValue: "/users/myUser", varSet: true},
			},
			wantTmpFolder: "/tmp", wantHomePath: "/users/myUser", wantErrors: nil,
		},
		{
			name: "missing TMP",
			envs: []envState{
				{varName: "TMP", varValue: "", varSet: false},
				{varName: "TEMP", varValue: "/tmp2", varSet: true},
				{varName: "HOMEPATH", varValue: "/users/myUser", varSet: true},
			},
			wantTmpFolder: "/tmp2", wantHomePath: "/users/myUser", wantErrors: nil,
		},
		{
			name: "missing TMP and TEMP",
			envs: []envState{
				{varName: "TMP", varValue: "", varSet: false},
				{varName: "TEMP", varValue: "", varSet: false},
				{varName: "HOMEPATH", varValue: "/users/myUser", varSet: true},
			},
			wantTmpFolder: "", wantHomePath: "/users/myUser", wantErrors: []error{noTempFolder},
		},
		{
			name: "missing homepath",
			envs: []envState{
				{varName: "TMP", varValue: "/tmp", varSet: true},
				{varName: "TEMP", varValue: "/tmp2", varSet: true},
				{varName: "HOMEPATH", varValue: "", varSet: false},
			},
			wantTmpFolder: "/tmp", wantHomePath: "", wantErrors: []error{noHomePath},
		},
		{
			name: "missing both vars",
			envs: []envState{
				{varName: "TMP", varValue: "", varSet: false},
				{varName: "TEMP", varValue: "", varSet: false},
				{varName: "HOMEPATH", varValue: "", varSet: false},
			},
			wantTmpFolder: "", wantHomePath: "", wantErrors: []error{noTempFolder, noHomePath},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// clear initial state
			TmpFolder = ""
			HomePath = ""
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
