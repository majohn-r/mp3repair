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
			efs: cmd.NewExportFlagSettings().WithDefaultsEnabled(
				false).WithDefaultsSet(false),
			wantCanWrite: false,
			WantedRecording: output.WantedRecording{
				Error: "Default configuration settings will not be exported.\n" +
					"Why?\n" +
					"As currently configured, exporting default configuration settings is" +
					" disabled.\n" +
					"What to do:\n" +
					"Use either '--defaults' or '--defaults=true' to enable exporting" +
					" defaults.\n",
				Log: "level='error'" +
					" --defaults='false'" +
					" user-set='false'" +
					" msg='export defaults disabled'\n",
			},
		},
		"disabled intentionally": {
			efs: cmd.NewExportFlagSettings().WithDefaultsEnabled(
				false).WithDefaultsSet(true),
			wantCanWrite: false,
			WantedRecording: output.WantedRecording{
				Error: "Default configuration settings will not be exported.\n" +
					"Why?\n" +
					"You explicitly set --defaults false.\n" +
					"What to do:\n" +
					"Use either '--defaults' or '--defaults=true' to enable exporting" +
					" defaults.\n",
				Log: "level='error'" +
					" --defaults='false'" +
					" user-set='true'" +
					" msg='export defaults disabled'\n",
			},
		},
		"enabled by default": {
			efs: cmd.NewExportFlagSettings().WithDefaultsEnabled(
				true).WithDefaultsSet(false),
			wantCanWrite: true,
		},
		"enabled intentionally": {
			efs: cmd.NewExportFlagSettings().WithDefaultsEnabled(
				true).WithDefaultsSet(true),
			wantCanWrite: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if gotCanWrite := tt.efs.CanWriteDefaults(o); gotCanWrite != tt.wantCanWrite {
				t.Errorf("ExportFlagSettings.CanWriteDefaults() = %v, want %v",
					gotCanWrite, tt.wantCanWrite)
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
			efs: cmd.NewExportFlagSettings().WithOverwriteEnabled(
				false).WithOverwriteSet(false),
			args:             args{f: "[file name here]"},
			wantCanOverwrite: false,
			WantedRecording: output.WantedRecording{
				Error: "The file \"[file name here]\" exists and cannot be overwritten.\n" +
					"Why?\n" +
					"As currently configured, overwriting the file is disabled.\n" +
					"What to do:\n" +
					"Use either '--overwrite' or '--overwrite=true' to enable overwriting" +
					" the existing file.\n",
				Log: "level='error'" +
					" --overwrite='false'" +
					" fileName='[file name here]'" +
					" user-set='false'" +
					" msg='overwrite is not permitted'\n",
			},
		},
		"no overwrite intentionally": {
			efs: cmd.NewExportFlagSettings().WithOverwriteEnabled(
				false).WithOverwriteSet(true),
			args:             args{f: "[file name here]"},
			wantCanOverwrite: false,
			WantedRecording: output.WantedRecording{
				Error: "The file \"[file name here]\" exists and cannot be overwritten.\n" +
					"Why?\n" +
					"You explicitly set --overwrite false.\n" +
					"What to do:\n" +
					"Use either '--overwrite' or '--overwrite=true' to enable overwriting" +
					" the existing file.\n",
				Log: "level='error'" +
					" --overwrite='false'" +
					" fileName='[file name here]'" +
					" user-set='true'" +
					" msg='overwrite is not permitted'\n",
			},
		},
		"overwrite enabled by default": {
			efs: cmd.NewExportFlagSettings().WithOverwriteEnabled(
				true).WithOverwriteSet(false),
			args:             args{f: "[file name here]"},
			wantCanOverwrite: true,
		},
		"overwrite enabled intentionally": {
			efs: cmd.NewExportFlagSettings().WithOverwriteEnabled(
				true).WithOverwriteSet(true),
			args:             args{f: "[file name here]"},
			wantCanOverwrite: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if gotCanOverwrite := tt.efs.CanOverwriteFile(o,
				tt.args.f); gotCanOverwrite != tt.wantCanOverwrite {
				t.Errorf("ExportFlagSettings.CanOverwriteFile() = %v, want %v",
					gotCanOverwrite, tt.wantCanOverwrite)
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
	originalWriteFile := cmd.WriteFile
	defer func() {
		cmd.WriteFile = originalWriteFile
	}()
	type args struct {
		f       string
		content []byte
	}
	tests := map[string]struct {
		args
		writeFile func(string, []byte, fs.FileMode) error
		want      bool
		output.WantedRecording
	}{
		"good write": {
			args: args{f: "filename", content: []byte{}},
			writeFile: func(_ string, _ []byte, _ fs.FileMode) error {
				return nil
			},
			want: true,
			WantedRecording: output.WantedRecording{
				Console: "File \"filename\" has been written.\n",
			},
		},
		"bad write": {
			args: args{f: "filename", content: []byte{}},
			writeFile: func(_ string, _ []byte, _ fs.FileMode) error {
				return fmt.Errorf("disc jammed")
			},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "The file \"filename\" cannot be created: disc jammed.\n",
				Log: "level='error'" +
					" command='export'" +
					" error='disc jammed'" +
					" fileName='filename'" +
					" msg='cannot create file'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			cmd.WriteFile = tt.writeFile
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
	originalWriteFile := cmd.WriteFile
	originalRename := cmd.Rename
	originalRemove := cmd.Remove
	defer func() {
		cmd.WriteFile = originalWriteFile
		cmd.Rename = originalRename
		cmd.Remove = originalRemove
	}()
	type args struct {
		f       string
		payload []byte
	}
	tests := map[string]struct {
		efs *cmd.ExportFlagSettings
		args
		writeFile  func(string, []byte, fs.FileMode) error
		rename     func(string, string) error
		remove     func(string) error
		wantStatus int
		output.WantedRecording
	}{
		"nothing to do": {
			efs:        cmd.NewExportFlagSettings().WithOverwriteEnabled(false),
			args:       args{f: "[filename]"},
			wantStatus: cmd.UserError,
			WantedRecording: output.WantedRecording{
				Error: "The file \"[filename]\" exists and cannot be overwritten.\n" +
					"Why?\n" +
					"As currently configured, overwriting the file is disabled.\n" +
					"What to do:\n" +
					"Use either '--overwrite' or '--overwrite=true' to enable overwriting" +
					" the existing file.\n",
				Log: "level='error'" +
					" --overwrite='false'" +
					" fileName='[filename]'" +
					" user-set='false'" +
					" msg='overwrite is not permitted'\n",
			},
		},
		"rename fails": {
			efs:  cmd.NewExportFlagSettings().WithOverwriteEnabled(true),
			args: args{f: "[filename]"},
			rename: func(_, _ string) error {
				return fmt.Errorf("sorry, cannot rename that file")
			},
			wantStatus: cmd.SystemError,
			WantedRecording: output.WantedRecording{
				Error: "The file \"[filename]\" cannot be renamed to" +
					" \"[filename]-backup\": sorry, cannot rename that file.\n",
				Log: "level='error'" +
					" error='sorry, cannot rename that file'" +
					" new='[filename]-backup'" +
					" old='[filename]'" +
					" msg='rename failed'\n",
			},
		},
		"rename succeeds, create fails": {
			efs:    cmd.NewExportFlagSettings().WithOverwriteEnabled(true),
			args:   args{f: "[filename]", payload: []byte{}},
			rename: func(_, _ string) error { return nil },
			writeFile: func(_ string, _ []byte, _ fs.FileMode) error {
				return fmt.Errorf("disk is full")
			},
			wantStatus: cmd.SystemError,
			WantedRecording: output.WantedRecording{
				Error: "The file \"[filename]\" cannot be created: disk is full.\n",
				Log: "level='error'" +
					" command='export'" +
					" error='disk is full'" +
					" fileName='[filename]'" +
					" msg='cannot create file'\n",
			},
		},
		"everything succeeds": {
			efs:        cmd.NewExportFlagSettings().WithOverwriteEnabled(true),
			args:       args{f: "[filename]", payload: []byte{}},
			rename:     func(_, _ string) error { return nil },
			writeFile:  func(_ string, _ []byte, _ fs.FileMode) error { return nil },
			remove:     func(_ string) error { return nil },
			wantStatus: cmd.Success,
			WantedRecording: output.WantedRecording{
				Console: "File \"[filename]\" has been written.\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			cmd.WriteFile = tt.writeFile
			cmd.Rename = tt.rename
			cmd.Remove = tt.remove
			if got := tt.efs.OverwriteFile(o, tt.args.f,
				tt.args.payload); got != tt.wantStatus {
				t.Errorf("ExportFlagSettings.OverwriteFile() got %d want %d", got,
					tt.wantStatus)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ExportFlagSettings.OverwriteFile() %s", issue)
				}
			}
		})
	}
}

func TestExportFlagSettingsExportDefaultConfiguration(t *testing.T) {
	originalWriteFile := cmd.WriteFile
	originalRename := cmd.Rename
	originalRemove := cmd.Remove
	originalPlainFileExists := cmd.PlainFileExists
	originalApplicationPath := cmd.ApplicationPath
	defer func() {
		cmd.WriteFile = originalWriteFile
		cmd.Rename = originalRename
		cmd.Remove = originalRemove
		cmd.PlainFileExists = originalPlainFileExists
		cmd.ApplicationPath = originalApplicationPath
	}()
	tests := map[string]struct {
		efs             *cmd.ExportFlagSettings
		writeFile       func(string, []byte, fs.FileMode) error
		plainFileExists func(string) bool
		rename          func(string, string) error
		remove          func(string) error
		applicationPath func() string
		wantStatus      int
		output.WantedRecording
	}{
		"not asking to write": {
			efs: cmd.NewExportFlagSettings().WithOverwriteEnabled(
				true).WithDefaultsEnabled(false),
			wantStatus: cmd.UserError,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"Default configuration settings will not be exported.\n" +
					"Why?\n" +
					"As currently configured, exporting default configuration settings is" +
					" disabled.\n" +
					"What to do:\n" +
					"Use either '--defaults' or '--defaults=true' to enable exporting" +
					" defaults.\n",
				Log: "" +
					"level='error'" +
					" --defaults='false'" +
					" user-set='false'" +
					" msg='export defaults disabled'\n",
			},
		},
		"file does not exist but cannot be created": {
			efs: cmd.NewExportFlagSettings().WithOverwriteEnabled(
				true).WithDefaultsEnabled(true),
			writeFile: func(_ string, _ []byte, _ fs.FileMode) error {
				return fmt.Errorf("cannot write file, sorry")
			},
			plainFileExists: func(_ string) bool { return false },
			applicationPath: func() string { return "appPath" },
			wantStatus:      cmd.SystemError,
			WantedRecording: output.WantedRecording{
				Error: "The file \"appPath\\\\defaults.yaml\" cannot be created:" +
					" cannot write file, sorry.\n",
				Log: "" +
					"level='error'" +
					" command='export'" +
					" error='cannot write file, sorry'" +
					" fileName='appPath\\defaults.yaml'" +
					" msg='cannot create file'\n",
			},
		},
		"file does not exist": {
			efs: cmd.NewExportFlagSettings().WithOverwriteEnabled(
				true).WithDefaultsEnabled(true),
			writeFile:       func(_ string, _ []byte, _ fs.FileMode) error { return nil },
			plainFileExists: func(_ string) bool { return false },
			applicationPath: func() string { return "appPath" },
			wantStatus:      cmd.Success,
			WantedRecording: output.WantedRecording{
				Console: "File \"appPath\\\\defaults.yaml\" has been written.\n",
			},
		},
		"file exists": {
			efs: cmd.NewExportFlagSettings().WithOverwriteEnabled(
				true).WithDefaultsEnabled(true),
			writeFile:       func(_ string, _ []byte, _ fs.FileMode) error { return nil },
			plainFileExists: func(_ string) bool { return true },
			rename:          func(_, _ string) error { return nil },
			remove:          func(_ string) error { return nil },
			applicationPath: func() string { return "appPath" },
			wantStatus:      cmd.Success,
			WantedRecording: output.WantedRecording{
				Console: "File \"appPath\\\\defaults.yaml\" has been written.\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			cmd.WriteFile = tt.writeFile
			cmd.PlainFileExists = tt.plainFileExists
			cmd.Remove = tt.remove
			cmd.Rename = tt.rename
			cmd.ApplicationPath = tt.applicationPath
			if got := tt.efs.ExportDefaultConfiguration(o); got != tt.wantStatus {
				t.Errorf("ExportFlagSettings.ExportDefaultConfiguration() got %d want %d",
					got, tt.wantStatus)
			}
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
				cmd.ExportFlagDefaults:  cmd.NewFlagValue().WithValue("foo"),
				cmd.ExportFlagOverwrite: cmd.NewFlagValue().WithValue("bar"),
			}},
			want:  &cmd.ExportFlagSettings{},
			want1: false,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"defaults\" is not boolean" +
					" (foo).\n" +
					"An internal error occurred: flag \"overwrite\" is not boolean" +
					" (bar).\n",
				Log: "level='error'" +
					" error='flag value not boolean'" +
					" flag='defaults'" +
					" value='foo'" +
					" msg='internal error'\n" +
					"level='error'" +
					" error='flag value not boolean'" +
					" flag='overwrite'" +
					" value='bar'" +
					" msg='internal error'\n",
			},
		},
		"bad defaults settings": {
			args: args{values: map[string]*cmd.FlagValue{
				cmd.ExportFlagDefaults:  cmd.NewFlagValue().WithValue("foo"),
				cmd.ExportFlagOverwrite: cmd.NewFlagValue().WithValue(true),
			}},
			want:  cmd.NewExportFlagSettings().WithOverwriteEnabled(true),
			want1: false,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"defaults\" is not boolean" +
					" (foo).\n",
				Log: "level='error'" +
					" error='flag value not boolean'" +
					" flag='defaults'" +
					" value='foo'" +
					" msg='internal error'\n",
			},
		},
		"bad overwrites settings": {
			args: args{values: map[string]*cmd.FlagValue{
				cmd.ExportFlagDefaults:  cmd.NewFlagValue().WithValue(true),
				cmd.ExportFlagOverwrite: cmd.NewFlagValue().WithValue(17),
			}},
			want:  cmd.NewExportFlagSettings().WithDefaultsEnabled(true),
			want1: false,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"overwrite\" is not boolean" +
					" (17).\n",
				Log: "level='error'" +
					" error='flag value not boolean'" +
					" flag='overwrite'" +
					" value='17'" +
					" msg='internal error'\n",
			},
		},
		"everything good": {
			args: args{values: map[string]*cmd.FlagValue{
				cmd.ExportFlagDefaults:  cmd.NewFlagValue().WithValue(true),
				cmd.ExportFlagOverwrite: cmd.NewFlagValue().WithValue(true),
			}},
			want: cmd.NewExportFlagSettings().WithDefaultsEnabled(
				true).WithOverwriteEnabled(true),
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
	originalExit := cmd.Exit
	originalExportFlags := cmd.ExportFlags
	originalBus := cmd.Bus
	defer func() {
		cmd.Exit = originalExit
		cmd.ExportFlags = originalExportFlags
		cmd.Bus = originalBus
	}()
	var exitCode int
	var exitCalled bool
	cmd.Exit = func(code int) {
		exitCode = code
		exitCalled = true
	}
	tests := map[string]struct {
		cmd            *cobra.Command
		flags          *cmd.SectionFlags
		wantExitCode   int
		wantExitCalled bool
		output.WantedRecording
	}{
		"missing data": {
			cmd: cmd.ExportCmd,
			flags: cmd.NewSectionFlags().WithSectionName(cmd.ExportCommand).WithFlags(
				map[string]*cmd.FlagDetails{
					cmd.ExportFlagOverwrite: cmd.NewFlagDetails().WithExpectedType(
						cmd.BoolType).WithDefaultValue(12),
					cmd.ExportFlagDefaults: nil,
				},
			),
			wantExitCode:   cmd.ProgramError,
			wantExitCalled: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: no details for flag \"defaults\".\n",
				Log: "level='error'" +
					" error='no details for flag \"defaults\"'" +
					" msg='internal error'\n",
			},
		},
		"incomplete data": {
			cmd: cmd.ExportCmd,
			flags: cmd.NewSectionFlags().WithSectionName(cmd.ExportCommand).WithFlags(
				map[string]*cmd.FlagDetails{
					cmd.ExportFlagOverwrite: cmd.NewFlagDetails().WithExpectedType(
						cmd.BoolType).WithDefaultValue(12),
				},
			),
			wantExitCode:   cmd.ProgramError,
			wantExitCalled: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"defaults\" is not found.\n",
				Log: "level='error'" +
					" error='flag not found'" +
					" flag='defaults'" +
					" msg='internal error'\n",
			},
		},
		"valid data": {
			cmd: cmd.ExportCmd,
			flags: cmd.NewSectionFlags().WithSectionName(cmd.ExportCommand).WithFlags(
				map[string]*cmd.FlagDetails{
					cmd.ExportFlagOverwrite: cmd.NewFlagDetails().WithExpectedType(
						cmd.BoolType).WithDefaultValue(false),
					cmd.ExportFlagDefaults: cmd.NewFlagDetails().WithExpectedType(
						cmd.BoolType).WithDefaultValue(false),
				},
			),
			wantExitCode:   cmd.UserError,
			wantExitCalled: true,
			WantedRecording: output.WantedRecording{
				Error: "Default configuration settings will not be exported.\n" +
					"Why?\n" +
					"As currently configured, exporting default configuration settings is" +
					" disabled.\n" +
					"What to do:\n" +
					"Use either '--defaults' or '--defaults=true' to enable exporting" +
					" defaults.\n",
				Log: "level='info'" +
					" --defaults='false'" +
					" --overwrite='false'" +
					" command='export'" +
					" defaults-user-set='false'" +
					" overwrite-user-set='false'" +
					" msg='executing command'\n" +
					"level='error'" +
					" --defaults='false'" +
					" user-set='false'" +
					" msg='export defaults disabled'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			exitCode = -1
			exitCalled = false
			cmd.ExportFlags = tt.flags
			o := output.NewRecorder()
			cmd.Bus = o // this is what getBus() should return when ExportRun calls it
			cmd.ExportRun(tt.cmd, []string{})
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ExportRun() %s", issue)
				}
			}
			if got := exitCode; got != tt.wantExitCode {
				t.Errorf("ExportRun() got %d want %d", got, tt.wantExitCode)
			}
			if got := exitCalled; got != tt.wantExitCalled {
				t.Errorf("ExportRun() got %t want %t", got, tt.wantExitCalled)
			}
		})
	}
}

func TestExportHelp(t *testing.T) {
	tests := map[string]struct {
		output.WantedRecording
	}{
		"good": {
			WantedRecording: output.WantedRecording{
				Console: "" +
					"\"export\" exports default program configuration data to" +
					" %APPDATA%\\mp3\\defaults.yaml\n" +
					"\n" +
					"Usage:\n" +
					"  mp3 export [--defaults] [--overwrite]\n" +
					"\n" +
					"Examples:\n" +
					"export --defaults\n" +
					"  Write default program configuration data\n" +
					"export --overwrite\n" +
					"  Overwrite a pre-existing defaults.yaml file\n" +
					"\n" +
					"Flags:\n" +
					"  -d, --defaults    write default program configuration data" +
					" (default false)\n" +
					"  -o, --overwrite   overwrite existing file (default false)\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			command := cmd.ExportCmd
			enableCommandRecording(o, command)
			command.Help()
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("export Help() %s", issue)
				}
			}
		})
	}
}
