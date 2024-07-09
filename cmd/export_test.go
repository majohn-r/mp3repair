/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd_test

import (
	"fmt"
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"io/fs"
	"mp3repair/cmd"
	"reflect"
	"testing"

	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

func TestExportFlagSettingsCanWriteConfigurationFile(t *testing.T) {
	tests := map[string]struct {
		efs          *cmd.ExportSettings
		wantCanWrite bool
		output.WantedRecording
	}{
		"disabled by default": {
			efs: &cmd.ExportSettings{
				DefaultsEnabled: cmd.CommandFlag[bool]{Value: false, UserSet: false},
			},
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
			efs: &cmd.ExportSettings{
				DefaultsEnabled: cmd.CommandFlag[bool]{Value: false, UserSet: true},
			},
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
			efs: &cmd.ExportSettings{
				DefaultsEnabled: cmd.CommandFlag[bool]{Value: true, UserSet: false},
			},
			wantCanWrite: true,
		},
		"enabled intentionally": {
			efs: &cmd.ExportSettings{
				DefaultsEnabled: cmd.CommandFlag[bool]{Value: true, UserSet: true},
			},
			wantCanWrite: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if gotCanWrite := tt.efs.CanWriteConfigurationFile(o); gotCanWrite != tt.wantCanWrite {
				t.Errorf("ExportFlagSettings.CanWriteConfigurationFile() = %v, want %v", gotCanWrite, tt.wantCanWrite)
			}
			o.Report(t, "ExportFlagSettings.CanWriteConfigurationFile()", tt.WantedRecording)
		})
	}
}

func TestExportFlagSettingsCanOverwriteConfigurationFile(t *testing.T) {
	tests := map[string]struct {
		efs              *cmd.ExportSettings
		f                string
		wantCanOverwrite bool
		output.WantedRecording
	}{
		"no overwrite by default": {
			efs: &cmd.ExportSettings{
				OverwriteEnabled: cmd.CommandFlag[bool]{Value: false, UserSet: false},
			},
			f:                "[file name here]",
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
			efs: &cmd.ExportSettings{
				OverwriteEnabled: cmd.CommandFlag[bool]{Value: false, UserSet: true},
			},
			f:                "[file name here]",
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
			efs: &cmd.ExportSettings{
				OverwriteEnabled: cmd.CommandFlag[bool]{Value: true, UserSet: false},
			},
			f:                "[file name here]",
			wantCanOverwrite: true,
		},
		"overwrite enabled intentionally": {
			efs: &cmd.ExportSettings{
				OverwriteEnabled: cmd.CommandFlag[bool]{Value: true, UserSet: true},
			},
			f:                "[file name here]",
			wantCanOverwrite: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			gotCanOverwrite := tt.efs.CanOverwriteConfigurationFile(o, tt.f)
			if gotCanOverwrite != tt.wantCanOverwrite {
				t.Errorf("ExportFlagSettings.CanOverwriteConfigurationFile() = %v, want %v", gotCanOverwrite, tt.wantCanOverwrite)
			}
			o.Report(t, "ExportFlagSettings.CanOverwriteConfigurationFile()", tt.WantedRecording)
		})
	}
}

func TestCreateConfigurationFile(t *testing.T) {
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
			if got := cmd.CreateConfigurationFile(o, tt.args.f, tt.args.content); got != tt.want {
				t.Errorf("CreateConfigurationFile() = %v, want %v", got, tt.want)
			}
			o.Report(t, "CreateConfigurationFile()", tt.WantedRecording)
		})
	}
}

func TestExportFlagSettingsOverwriteConfigurationFile(t *testing.T) {
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
		efs *cmd.ExportSettings
		args
		writeFile  func(string, []byte, fs.FileMode) error
		rename     func(string, string) error
		remove     func(string) error
		wantStatus *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"nothing to do": {
			efs: &cmd.ExportSettings{
				OverwriteEnabled: cmd.CommandFlag[bool]{Value: false},
			},
			args:       args{f: "[filename]"},
			wantStatus: cmdtoolkit.NewExitUserError("export"),
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
			efs: &cmd.ExportSettings{
				OverwriteEnabled: cmd.CommandFlag[bool]{Value: true},
			},
			args: args{f: "[filename]"},
			rename: func(_, _ string) error {
				return fmt.Errorf("sorry, cannot rename that file")
			},
			wantStatus: cmdtoolkit.NewExitSystemError("export"),
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
			efs: &cmd.ExportSettings{
				OverwriteEnabled: cmd.CommandFlag[bool]{Value: true},
			},
			args:   args{f: "[filename]", payload: []byte{}},
			rename: func(_, _ string) error { return nil },
			writeFile: func(_ string, _ []byte, _ fs.FileMode) error {
				return fmt.Errorf("disk is full")
			},
			wantStatus: cmdtoolkit.NewExitSystemError("export"),
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
			efs: &cmd.ExportSettings{
				OverwriteEnabled: cmd.CommandFlag[bool]{Value: true},
			},
			args:       args{f: "[filename]", payload: []byte{}},
			rename:     func(_, _ string) error { return nil },
			writeFile:  func(_ string, _ []byte, _ fs.FileMode) error { return nil },
			remove:     func(_ string) error { return nil },
			wantStatus: nil,
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
			got := tt.efs.OverwriteConfigurationFile(o, tt.args.f, tt.args.payload)
			if !compareExitErrors(got, tt.wantStatus) {
				t.Errorf("ExportFlagSettings.OverwriteConfigurationFile() got %s want %s", got, tt.wantStatus)
			}
			o.Report(t, "ExportFlagSettings.OverwriteConfigurationFile()", tt.WantedRecording)
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
		efs             *cmd.ExportSettings
		writeFile       func(string, []byte, fs.FileMode) error
		plainFileExists func(string) bool
		rename          func(string, string) error
		remove          func(string) error
		applicationPath func() string
		wantStatus      *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"not asking to write": {
			efs: &cmd.ExportSettings{
				OverwriteEnabled: cmd.CommandFlag[bool]{Value: true},
				DefaultsEnabled:  cmd.CommandFlag[bool]{Value: false},
			},
			wantStatus: cmdtoolkit.NewExitUserError("export"),
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
			efs: &cmd.ExportSettings{
				OverwriteEnabled: cmd.CommandFlag[bool]{Value: true},
				DefaultsEnabled:  cmd.CommandFlag[bool]{Value: true},
			},
			writeFile: func(_ string, _ []byte, _ fs.FileMode) error {
				return fmt.Errorf("cannot write file, sorry")
			},
			plainFileExists: func(_ string) bool { return false },
			applicationPath: func() string { return "appPath" },
			wantStatus:      cmdtoolkit.NewExitSystemError("export"),
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
			efs: &cmd.ExportSettings{
				OverwriteEnabled: cmd.CommandFlag[bool]{Value: true},
				DefaultsEnabled:  cmd.CommandFlag[bool]{Value: true},
			},
			writeFile:       func(_ string, _ []byte, _ fs.FileMode) error { return nil },
			plainFileExists: func(_ string) bool { return false },
			applicationPath: func() string { return "appPath" },
			wantStatus:      nil,
			WantedRecording: output.WantedRecording{
				Console: "File \"appPath\\\\defaults.yaml\" has been written.\n",
			},
		},
		"file exists": {
			efs: &cmd.ExportSettings{
				OverwriteEnabled: cmd.CommandFlag[bool]{Value: true},
				DefaultsEnabled:  cmd.CommandFlag[bool]{Value: true},
			},
			writeFile:       func(_ string, _ []byte, _ fs.FileMode) error { return nil },
			plainFileExists: func(_ string) bool { return true },
			rename:          func(_, _ string) error { return nil },
			remove:          func(_ string) error { return nil },
			applicationPath: func() string { return "appPath" },
			wantStatus:      nil,
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
			if got := tt.efs.ExportDefaultConfiguration(o); !compareExitErrors(got, tt.wantStatus) {
				t.Errorf("ExportFlagSettings.ExportDefaultConfiguration() got %s want %s", got, tt.wantStatus)
			}
			o.Report(t, "ExportFlagSettings.ExportDefaultConfiguration()", tt.WantedRecording)
		})
	}
}

func TestProcessExportFlags(t *testing.T) {
	tests := map[string]struct {
		values map[string]*cmd.CommandFlag[any]
		want   *cmd.ExportSettings
		want1  bool
		output.WantedRecording
	}{
		"nothing went right": {
			values: map[string]*cmd.CommandFlag[any]{
				cmd.ExportFlagDefaults:  {Value: "foo"},
				cmd.ExportFlagOverwrite: {Value: "bar"},
			},
			want:  &cmd.ExportSettings{},
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
			values: map[string]*cmd.CommandFlag[any]{
				cmd.ExportFlagDefaults:  {Value: "foo"},
				cmd.ExportFlagOverwrite: {Value: true},
			},
			want:  &cmd.ExportSettings{OverwriteEnabled: cmd.CommandFlag[bool]{Value: true}},
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
			values: map[string]*cmd.CommandFlag[any]{
				cmd.ExportFlagDefaults:  {Value: true},
				cmd.ExportFlagOverwrite: {Value: 17},
			},
			want:  &cmd.ExportSettings{DefaultsEnabled: cmd.CommandFlag[bool]{Value: true}},
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
			values: map[string]*cmd.CommandFlag[any]{
				cmd.ExportFlagDefaults:  {Value: true},
				cmd.ExportFlagOverwrite: {Value: true},
			},
			want: &cmd.ExportSettings{
				OverwriteEnabled: cmd.CommandFlag[bool]{Value: true},
				DefaultsEnabled:  cmd.CommandFlag[bool]{Value: true},
			},
			want1: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, got1 := cmd.ProcessExportFlags(o, tt.values)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessExportFlags() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ProcessExportFlags() got1 = %v, want %v", got1, tt.want1)
			}
			o.Report(t, "ProcessExportFlags()", tt.WantedRecording)
		})
	}
}

func TestExportRun(t *testing.T) {
	cmd.InitGlobals()
	originalExportFlags := cmd.ExportFlags
	originalBus := cmd.Bus
	defer func() {
		cmd.ExportFlags = originalExportFlags
		cmd.Bus = originalBus
	}()
	tests := map[string]struct {
		cmd   *cobra.Command
		flags *cmd.SectionFlags
		output.WantedRecording
	}{
		"missing data": {
			cmd: cmd.ExportCmd,
			flags: &cmd.SectionFlags{
				SectionName: cmd.ExportCommand,
				Details: map[string]*cmd.FlagDetails{
					cmd.ExportFlagOverwrite: {
						ExpectedType: cmd.BoolType,
						DefaultValue: 12,
					},
					cmd.ExportFlagDefaults: nil,
				},
			},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: no details for flag \"defaults\".\n",
				Log: "level='error'" +
					" error='no details for flag \"defaults\"'" +
					" msg='internal error'\n",
			},
		},
		"incomplete data": {
			cmd: cmd.ExportCmd,
			flags: &cmd.SectionFlags{
				SectionName: cmd.ExportCommand,
				Details: map[string]*cmd.FlagDetails{
					cmd.ExportFlagOverwrite: {
						ExpectedType: cmd.BoolType,
						DefaultValue: 12,
					},
				},
			},
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
			flags: &cmd.SectionFlags{
				SectionName: cmd.ExportCommand,
				Details: map[string]*cmd.FlagDetails{
					cmd.ExportFlagOverwrite: {
						ExpectedType: cmd.BoolType,
						DefaultValue: false,
					},
					cmd.ExportFlagDefaults: {
						ExpectedType: cmd.BoolType,
						DefaultValue: false,
					},
				},
			},
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
			cmd.ExportFlags = tt.flags
			o := output.NewRecorder()
			cmd.Bus = o // this is what getBus() should return when ExportRun calls it
			_ = cmd.ExportRun(tt.cmd, []string{})
			o.Report(t, "ExportRun()", tt.WantedRecording)
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
					" %APPDATA%\\mp3repair\\defaults.yaml\n" +
					"\n" +
					"Usage:\n" +
					"  mp3repair export [--defaults] [--overwrite]\n" +
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
			_ = command.Help()
			o.Report(t, "export Help()", tt.WantedRecording)
		})
	}
}
