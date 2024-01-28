/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd_test

import (
	"fmt"
	"io/fs"
	"mp3/cmd"
	"reflect"
	"testing"

	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

func TestExportFlagSettingsCanWriteDefaults(t *testing.T) {
	tests := map[string]struct {
		efs          *cmd.ExportFlagSettings
		wantCanWrite bool
		output.WantedRecording
	}{
		"disabled by default": {
			efs:          &cmd.ExportFlagSettings{DefaultsEnabled: false, DefaultsSet: false},
			wantCanWrite: false,
			WantedRecording: output.WantedRecording{
				Error: "Default configuration settings will not be exported.\n" +
					"Why?\n" +
					"As currently configured, exporting default configuration settings is disabled.\n" +
					"What to do:\n" +
					"Use either '--defaults' or '--defaults=true' to enable exporting defaults.\n",
				Log: "level='error' --defaults='false' user-set='false' msg='export defaults disabled'\n",
			},
		},
		"disabled intentionally": {
			efs:          &cmd.ExportFlagSettings{DefaultsEnabled: false, DefaultsSet: true},
			wantCanWrite: false,
			WantedRecording: output.WantedRecording{
				Error: "Default configuration settings will not be exported.\n" +
					"Why?\n" +
					"You explicitly set --defaults false.\n" +
					"What to do:\n" +
					"Use either '--defaults' or '--defaults=true' to enable exporting defaults.\n",
				Log: "level='error' --defaults='false' user-set='true' msg='export defaults disabled'\n",
			},
		},
		"enabled by default": {
			efs:          &cmd.ExportFlagSettings{DefaultsEnabled: true, DefaultsSet: false},
			wantCanWrite: true,
		},
		"enabled intentionally": {
			efs:          &cmd.ExportFlagSettings{DefaultsEnabled: true, DefaultsSet: true},
			wantCanWrite: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if gotCanWrite := tt.efs.CanWriteDefaults(o); gotCanWrite != tt.wantCanWrite {
				t.Errorf("ExportFlagSettings.CanWriteDefaults() = %v, want %v", gotCanWrite, tt.wantCanWrite)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ExportFlagSettings.CanWriteDefaults() %s", issue)
				}
			}
		})
	}
}

func TestExportFlagSettingsCanOverwriteFile(t *testing.T) {
	type args struct {
		f string
	}
	tests := map[string]struct {
		efs *cmd.ExportFlagSettings
		args
		wantCanOverwrite bool
		output.WantedRecording
	}{
		"no overwrite by default": {
			efs:              &cmd.ExportFlagSettings{OverwriteEnabled: false, OverwriteSet: false},
			args:             args{f: "[file name here]"},
			wantCanOverwrite: false,
			WantedRecording: output.WantedRecording{
				Error: "The file \"[file name here]\" exists and cannot be overwritten.\n" +
					"Why?\n" +
					"As currently configured, overwriting the file is disabled.\n" +
					"What to do:\n" +
					"Use either '--overwrite' or '--overwrite=true' to enable overwriting the existing file.\n",
				Log: "level='error' --overwrite='false' fileName='[file name here]' user-set='false' msg='overwrite is not permitted'\n",
			},
		},
		"no overwrite intentionally": {
			efs:              &cmd.ExportFlagSettings{OverwriteEnabled: false, OverwriteSet: true},
			args:             args{f: "[file name here]"},
			wantCanOverwrite: false,
			WantedRecording: output.WantedRecording{
				Error: "The file \"[file name here]\" exists and cannot be overwritten.\n" +
					"Why?\n" +
					"You explicitly set --overwrite false.\n" +
					"What to do:\n" +
					"Use either '--overwrite' or '--overwrite=true' to enable overwriting the existing file.\n",
				Log: "level='error' --overwrite='false' fileName='[file name here]' user-set='true' msg='overwrite is not permitted'\n",
			},
		},
		"overwrite enabled by default": {
			efs:              &cmd.ExportFlagSettings{OverwriteEnabled: true, OverwriteSet: false},
			args:             args{f: "[file name here]"},
			wantCanOverwrite: true,
		},
		"overwrite enabled intentionally": {
			efs:              &cmd.ExportFlagSettings{OverwriteEnabled: true, OverwriteSet: true},
			args:             args{f: "[file name here]"},
			wantCanOverwrite: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if gotCanOverwrite := tt.efs.CanOverwriteFile(o, tt.args.f); gotCanOverwrite != tt.wantCanOverwrite {
				t.Errorf("ExportFlagSettings.CanOverwriteFile() = %v, want %v", gotCanOverwrite, tt.wantCanOverwrite)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ExportFlagSettings.CanOverwriteFile() %s", issue)
				}
			}
		})
	}
}

func TestCreateFile(t *testing.T) {
	oldExportFileCreate := cmd.FileWrite
	defer func() {
		cmd.FileWrite = oldExportFileCreate
	}()
	type args struct {
		f       string
		content []byte
	}
	tests := map[string]struct {
		args
		createFunc func(string, []byte, fs.FileMode) error
		want       bool
		output.WantedRecording
	}{
		"good write": {
			args: args{f: "filename", content: []byte{}},
			createFunc: func(_ string, _ []byte, _ fs.FileMode) error {
				return nil
			},
			want: true,
			WantedRecording: output.WantedRecording{
				Console: "File \"filename\" has been written.\n",
			},
		},
		"bad write": {
			args: args{f: "filename", content: []byte{}},
			createFunc: func(_ string, _ []byte, _ fs.FileMode) error {
				return fmt.Errorf("disc jammed")
			},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "The file \"filename\" cannot be created: disc jammed.\n",
				Log:   "level='error' command='export' error='disc jammed' fileName='filename' msg='cannot create file'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			cmd.FileWrite = tt.createFunc
			if got := cmd.CreateFile(o, tt.args.f, tt.args.content); got != tt.want {
				t.Errorf("CreateFile() = %v, want %v", got, tt.want)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("CreateFile() %s", issue)
				}
			}
		})
	}
}

func TestExportFlagSettingsOverwriteFile(t *testing.T) {
	oldExportFileCreate := cmd.FileWrite
	oldExportFileRename := cmd.FileRename
	oldExportFileRemove := cmd.FileRemove
	defer func() {
		cmd.FileWrite = oldExportFileCreate
		cmd.FileRename = oldExportFileRename
		cmd.FileRemove = oldExportFileRemove
	}()
	type args struct {
		f       string
		payload []byte
	}
	tests := map[string]struct {
		efs *cmd.ExportFlagSettings
		args
		createFunc func(string, []byte, fs.FileMode) error
		renameFunc func(string, string) error
		removeFunc func(string) error
		output.WantedRecording
	}{
		"nothing to do": {
			efs:  &cmd.ExportFlagSettings{OverwriteEnabled: false},
			args: args{f: "[filename]"},
			WantedRecording: output.WantedRecording{
				Error: "The file \"[filename]\" exists and cannot be overwritten.\n" +
					"Why?\n" +
					"As currently configured, overwriting the file is disabled.\n" +
					"What to do:\n" +
					"Use either '--overwrite' or '--overwrite=true' to enable overwriting the existing file.\n",
				Log: "level='error' --overwrite='false' fileName='[filename]' user-set='false' msg='overwrite is not permitted'\n",
			},
		},
		"rename fails": {
			efs:        &cmd.ExportFlagSettings{OverwriteEnabled: true},
			args:       args{f: "[filename]"},
			renameFunc: func(_, _ string) error { return fmt.Errorf("sorry, cannot rename that file") },
			WantedRecording: output.WantedRecording{
				Error: "The file \"[filename]\" cannot be renamed to \"[filename]-backup\": sorry, cannot rename that file.\n",
				Log:   "level='error' error='sorry, cannot rename that file' new='[filename]-backup' old='[filename]' msg='rename failed'\n",
			},
		},
		"rename succeeds, create fails": {
			efs:        &cmd.ExportFlagSettings{OverwriteEnabled: true},
			args:       args{f: "[filename]", payload: []byte{}},
			renameFunc: func(_, _ string) error { return nil },
			createFunc: func(_ string, _ []byte, _ fs.FileMode) error { return fmt.Errorf("disk is full") },
			WantedRecording: output.WantedRecording{
				Error: "The file \"[filename]\" cannot be created: disk is full.\n",
				Log:   "level='error' command='export' error='disk is full' fileName='[filename]' msg='cannot create file'\n",
			},
		},
		"everything succeeds": {
			efs:        &cmd.ExportFlagSettings{OverwriteEnabled: true},
			args:       args{f: "[filename]", payload: []byte{}},
			renameFunc: func(_, _ string) error { return nil },
			createFunc: func(_ string, _ []byte, _ fs.FileMode) error { return nil },
			removeFunc: func(_ string) error { return nil },
			WantedRecording: output.WantedRecording{
				Console: "File \"[filename]\" has been written.\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			cmd.FileWrite = tt.createFunc
			cmd.FileRename = tt.renameFunc
			cmd.FileRemove = tt.removeFunc
			tt.efs.OverwriteFile(o, tt.args.f, tt.args.payload)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ExportFlagSettings.OverwriteFile() %s", issue)
				}
			}
		})
	}
}

func TestExportFlagSettingsExportDefaultConfiguration(t *testing.T) {
	oldFileWrite := cmd.FileWrite
	oldFileRename := cmd.FileRename
	oldFileRemove := cmd.FileRemove
	oldPlainFileExists := cmd.PlainFileExists
	oldApplicationPath := cmd.ApplicationPath
	defer func() {
		cmd.FileWrite = oldFileWrite
		cmd.FileRename = oldFileRename
		cmd.FileRemove = oldFileRemove
		cmd.PlainFileExists = oldPlainFileExists
		cmd.ApplicationPath = oldApplicationPath
	}()
	tests := map[string]struct {
		efs         *cmd.ExportFlagSettings
		createFunc  func(string, []byte, fs.FileMode) error
		existsFunc  func(string) bool
		renameFunc  func(string, string) error
		removeFunc  func(string) error
		appPathFunc func() string
		output.WantedRecording
	}{
		// only going to test happy flows - other tests will catch unhappy
		// flows, e.g., cannot create the file
		"file does not exist": {
			efs:         &cmd.ExportFlagSettings{OverwriteEnabled: true, DefaultsEnabled: true},
			createFunc:  func(_ string, _ []byte, _ fs.FileMode) error { return nil },
			existsFunc:  func(_ string) bool { return false },
			appPathFunc: func() string { return "appPath" },
			WantedRecording: output.WantedRecording{
				Console: "File \"appPath\\\\defaults.yaml\" has been written.\n",
			},
		},
		"file exists": {
			efs:         &cmd.ExportFlagSettings{OverwriteEnabled: true, DefaultsEnabled: true},
			createFunc:  func(_ string, _ []byte, _ fs.FileMode) error { return nil },
			existsFunc:  func(_ string) bool { return true },
			renameFunc:  func(_, _ string) error { return nil },
			removeFunc:  func(_ string) error { return nil },
			appPathFunc: func() string { return "appPath" },
			WantedRecording: output.WantedRecording{
				Console: "File \"appPath\\\\defaults.yaml\" has been written.\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			cmd.FileWrite = tt.createFunc
			cmd.PlainFileExists = tt.existsFunc
			cmd.FileRemove = tt.removeFunc
			cmd.FileRename = tt.renameFunc
			cmd.ApplicationPath = tt.appPathFunc
			tt.efs.ExportDefaultConfiguration(o)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ExportFlagSettings.ExportDefaultConfiguration() %s", issue)
				}
			}
		})
	}
}

func TestProcessExportFlags(t *testing.T) {
	type args struct {
		values map[string]*cmd.FlagValue
	}
	tests := map[string]struct {
		args
		want  *cmd.ExportFlagSettings
		want1 bool
		output.WantedRecording
	}{
		"nothing went right": {
			args: args{values: map[string]*cmd.FlagValue{
				cmd.ExportFlagDefaults:  {Value: "foo"},
				cmd.ExportFlagOverwrite: {Value: "bar"},
			}},
			want:  &cmd.ExportFlagSettings{},
			want1: false,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"defaults\" is not boolean (foo).\n" +
					"An internal error occurred: flag \"overwrite\" is not boolean (bar).\n",
				Log: "level='error' error='flag value not boolean' flag='defaults' value='foo' msg='internal error'\n" +
					"level='error' error='flag value not boolean' flag='overwrite' value='bar' msg='internal error'\n",
			},
		},
		"bad defaults settings": {
			args: args{values: map[string]*cmd.FlagValue{
				cmd.ExportFlagDefaults:  {Value: "foo"},
				cmd.ExportFlagOverwrite: {Value: true},
			}},
			want:  &cmd.ExportFlagSettings{OverwriteEnabled: true},
			want1: false,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"defaults\" is not boolean (foo).\n",
				Log:   "level='error' error='flag value not boolean' flag='defaults' value='foo' msg='internal error'\n",
			},
		},
		"bad overwrites settings": {
			args: args{values: map[string]*cmd.FlagValue{
				cmd.ExportFlagDefaults:  {Value: true},
				cmd.ExportFlagOverwrite: {Value: 17},
			}},
			want:  &cmd.ExportFlagSettings{DefaultsEnabled: true},
			want1: false,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"overwrite\" is not boolean (17).\n",
				Log:   "level='error' error='flag value not boolean' flag='overwrite' value='17' msg='internal error'\n",
			},
		},
		"everything good": {
			args: args{values: map[string]*cmd.FlagValue{
				cmd.ExportFlagDefaults:  {Value: true},
				cmd.ExportFlagOverwrite: {Value: true},
			}},
			want:  &cmd.ExportFlagSettings{DefaultsEnabled: true, OverwriteEnabled: true},
			want1: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, got1 := cmd.ProcessExportFlags(o, tt.args.values)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessExportFlags() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ProcessExportFlags() got1 = %v, want %v", got1, tt.want1)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ProcessExportFlags() %s", issue)
				}
			}
		})
	}
}

func TestExportRun(t *testing.T) {
	cmd.InitGlobals()
	oldExitFunction := cmd.ExitFunction
	oldExportFlags := cmd.ExportFlags
	oldBus := cmd.Bus
	defer func() {
		cmd.ExitFunction = oldExitFunction
		cmd.ExportFlags = oldExportFlags
		cmd.Bus = oldBus
	}()
	cmd.ExitFunction = func(int) {}
	tests := map[string]struct {
		cmd   *cobra.Command
		flags cmd.SectionFlags
		output.WantedRecording
	}{
		"missing data": {
			cmd: cmd.ExportCmd,
			flags: cmd.SectionFlags{
				SectionName: cmd.ExportCommand,
				Flags: map[string]*cmd.FlagDetails{
					cmd.ExportFlagOverwrite: {ExpectedType: cmd.BoolType, DefaultValue: 12},
					cmd.ExportFlagDefaults:  nil,
				},
			},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: no details for flag \"defaults\".\n",
				Log:   "level='error' error='no details for flag \"defaults\"' msg='internal error'\n",
			},
		},
		"incomplete data": {
			cmd: cmd.ExportCmd,
			flags: cmd.SectionFlags{
				SectionName: cmd.ExportCommand,
				Flags: map[string]*cmd.FlagDetails{
					cmd.ExportFlagOverwrite: {ExpectedType: cmd.BoolType, DefaultValue: 12},
				},
			},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"defaults\" is not found.\n",
				Log:   "level='error' error='flag not found' flag='defaults' msg='internal error'\n",
			},
		},
		"valid data": {
			cmd: cmd.ExportCmd,
			flags: cmd.SectionFlags{
				SectionName: cmd.ExportCommand,
				Flags: map[string]*cmd.FlagDetails{
					cmd.ExportFlagOverwrite: {ExpectedType: cmd.BoolType, DefaultValue: false},
					cmd.ExportFlagDefaults:  {ExpectedType: cmd.BoolType, DefaultValue: false},
				},
			},
			WantedRecording: output.WantedRecording{
				Error: "Default configuration settings will not be exported.\n" +
					"Why?\n" +
					"As currently configured, exporting default configuration settings is disabled.\n" +
					"What to do:\n" +
					"Use either '--defaults' or '--defaults=true' to enable exporting defaults.\n",
				Log: "level='info' --defaults='false' --overwrite='false' command='export' defaults-user-set='false' overwrite-user-set='false' msg='executing command'\n" +
					"level='error' --defaults='false' user-set='false' msg='export defaults disabled'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmd.ExportFlags = tt.flags
			o := output.NewRecorder()
			cmd.Bus = o // this is what getBus() should return when ExportRun calls it
			cmd.ExportRun(tt.cmd, []string{})
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ExportRun() %s", issue)
				}
			}
		})
	}
}
