/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"reflect"
	"testing"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/spf13/afero"

	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

func Test_exportFlagSettings_canWriteConfigurationFile(t *testing.T) {
	tests := map[string]struct {
		efs          *exportSettings
		wantCanWrite bool
		output.WantedRecording
	}{
		"disabled by default": {
			efs: &exportSettings{
				defaultsEnabled: cmdtoolkit.CommandFlag[bool]{Value: false, UserSet: false},
			},
			wantCanWrite: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"Default configuration settings will not be exported.\n" +
					"Why?\n" +
					"As currently configured, exporting default configuration settings is disabled.\n" +
					"What to do:\n" +
					"To enable exporting defaults, use either:\n" +
					"● --defaults or\n" +
					"● --defaults=true\n",
				Log: "level='error'" +
					" --defaults='false'" +
					" user-set='false'" +
					" msg='export defaults disabled'\n",
			},
		},
		"disabled intentionally": {
			efs: &exportSettings{
				defaultsEnabled: cmdtoolkit.CommandFlag[bool]{Value: false, UserSet: true},
			},
			wantCanWrite: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"Default configuration settings will not be exported.\n" +
					"Why?\n" +
					"You explicitly set --defaults false.\n" +
					"What to do:\n" +
					"To enable exporting defaults, use either:\n" +
					"● --defaults or\n" +
					"● --defaults=true\n",
				Log: "level='error'" +
					" --defaults='false'" +
					" user-set='true'" +
					" msg='export defaults disabled'\n",
			},
		},
		"enabled by default": {
			efs: &exportSettings{
				defaultsEnabled: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: false},
			},
			wantCanWrite: true,
		},
		"enabled intentionally": {
			efs: &exportSettings{
				defaultsEnabled: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
			},
			wantCanWrite: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if gotCanWrite := tt.efs.canWriteConfigurationFile(o); gotCanWrite != tt.wantCanWrite {
				t.Errorf("exportFlagSettings.canWriteConfigurationFile() = %v, want %v", gotCanWrite, tt.wantCanWrite)
			}
			o.Report(t, "exportFlagSettings.canWriteConfigurationFile()", tt.WantedRecording)
		})
	}
}

func Test_exportFlagSettings_canOverwriteConfigurationFile(t *testing.T) {
	tests := map[string]struct {
		efs              *exportSettings
		f                string
		wantCanOverwrite bool
		output.WantedRecording
	}{
		"no overwrite by default": {
			efs: &exportSettings{
				overwriteEnabled: cmdtoolkit.CommandFlag[bool]{Value: false, UserSet: false},
			},
			f:                "[file name here]",
			wantCanOverwrite: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The file \"[file name here]\" exists and cannot be overwritten.\n" +
					"Why?\n" +
					"As currently configured, overwriting the file is disabled.\n" +
					"What to do:\n" +
					"To enable overwriting the existing file, use either:\n" +
					"● --overwrite or\n" +
					"● --overwrite=true\n",
				Log: "level='error'" +
					" --overwrite='false'" +
					" fileName='[file name here]'" +
					" user-set='false'" +
					" msg='overwrite is not permitted'\n",
			},
		},
		"no overwrite intentionally": {
			efs: &exportSettings{
				overwriteEnabled: cmdtoolkit.CommandFlag[bool]{Value: false, UserSet: true},
			},
			f:                "[file name here]",
			wantCanOverwrite: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The file \"[file name here]\" exists and cannot be overwritten.\n" +
					"Why?\n" +
					"You explicitly set --overwrite false.\n" +
					"What to do:\n" +
					"To enable overwriting the existing file, use either:\n" +
					"● --overwrite or\n" +
					"● --overwrite=true\n",
				Log: "level='error'" +
					" --overwrite='false'" +
					" fileName='[file name here]'" +
					" user-set='true'" +
					" msg='overwrite is not permitted'\n",
			},
		},
		"overwrite enabled by default": {
			efs: &exportSettings{
				overwriteEnabled: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: false},
			},
			f:                "[file name here]",
			wantCanOverwrite: true,
		},
		"overwrite enabled intentionally": {
			efs: &exportSettings{
				overwriteEnabled: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
			},
			f:                "[file name here]",
			wantCanOverwrite: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			gotCanOverwrite := tt.efs.canOverwriteConfigurationFile(o, tt.f)
			if gotCanOverwrite != tt.wantCanOverwrite {
				t.Errorf(
					"exportFlagSettings.canOverwriteConfigurationFile() = %v, want %v",
					gotCanOverwrite,
					tt.wantCanOverwrite,
				)
			}
			o.Report(t, "exportFlagSettings.canOverwriteConfigurationFile()", tt.WantedRecording)
		})
	}
}

func Test_createConfigurationFile(t *testing.T) {
	originalWriteFile := writeFile
	defer func() {
		writeFile = originalWriteFile
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
				Error: "The file \"filename\" cannot be created: 'disc jammed'.\n",
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
			writeFile = tt.writeFile
			if got := createConfigurationFile(o, tt.args.f, tt.args.content); got != tt.want {
				t.Errorf("createConfigurationFile() = %v, want %v", got, tt.want)
			}
			o.Report(t, "createConfigurationFile()", tt.WantedRecording)
		})
	}
}

func Test_exportFlagSettings_overwriteConfigurationFile(t *testing.T) {
	originalWriteFile := writeFile
	originalRename := rename
	originalRemove := remove
	defer func() {
		writeFile = originalWriteFile
		rename = originalRename
		remove = originalRemove
	}()
	type args struct {
		f       string
		payload []byte
	}
	tests := map[string]struct {
		efs *exportSettings
		args
		writeFile  func(string, []byte, fs.FileMode) error
		rename     func(string, string) error
		remove     func(string) error
		wantStatus *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"nothing to do": {
			efs: &exportSettings{
				overwriteEnabled: cmdtoolkit.CommandFlag[bool]{Value: false},
			},
			args:       args{f: "[filename]"},
			wantStatus: cmdtoolkit.NewExitUserError("export"),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The file \"[filename]\" exists and cannot be overwritten.\n" +
					"Why?\n" +
					"As currently configured, overwriting the file is disabled.\n" +
					"What to do:\n" +
					"To enable overwriting the existing file, use either:\n" +
					"● --overwrite or\n" +
					"● --overwrite=true\n",
				Log: "level='error'" +
					" --overwrite='false'" +
					" fileName='[filename]'" +
					" user-set='false'" +
					" msg='overwrite is not permitted'\n",
			},
		},
		"rename fails": {
			efs: &exportSettings{
				overwriteEnabled: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			args: args{f: "[filename]"},
			rename: func(_, _ string) error {
				return fmt.Errorf("sorry, cannot rename that file")
			},
			wantStatus: cmdtoolkit.NewExitSystemError("export"),
			WantedRecording: output.WantedRecording{
				Error: "The file \"[filename]\" cannot be renamed to" +
					" \"[filename]-backup\": 'sorry, cannot rename that file'.\n",
				Log: "level='error'" +
					" error='sorry, cannot rename that file'" +
					" new='[filename]-backup'" +
					" old='[filename]'" +
					" msg='rename failed'\n",
			},
		},
		"rename succeeds, create fails": {
			efs: &exportSettings{
				overwriteEnabled: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			args:   args{f: "[filename]", payload: []byte{}},
			rename: func(_, _ string) error { return nil },
			writeFile: func(_ string, _ []byte, _ fs.FileMode) error {
				return fmt.Errorf("disk is full")
			},
			wantStatus: cmdtoolkit.NewExitSystemError("export"),
			WantedRecording: output.WantedRecording{
				Error: "The file \"[filename]\" cannot be created: 'disk is full'.\n",
				Log: "level='error'" +
					" command='export'" +
					" error='disk is full'" +
					" fileName='[filename]'" +
					" msg='cannot create file'\n",
			},
		},
		"everything succeeds": {
			efs: &exportSettings{
				overwriteEnabled: cmdtoolkit.CommandFlag[bool]{Value: true},
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
			writeFile = tt.writeFile
			rename = tt.rename
			remove = tt.remove
			got := tt.efs.overwriteConfigurationFile(o, tt.args.f, tt.args.payload)
			if !compareExitErrors(got, tt.wantStatus) {
				t.Errorf("exportFlagSettings.overwriteConfigurationFile() got %s want %s", got, tt.wantStatus)
			}
			o.Report(t, "exportFlagSettings.overwriteConfigurationFile()", tt.WantedRecording)
		})
	}
}

func Test_exportFlagSettings_exportDefaultConfiguration(t *testing.T) {
	originalWriteFile := writeFile
	originalRename := rename
	originalRemove := remove
	fileSystem := afero.NewMemMapFs()
	originalFileSystem := cmdtoolkit.AssignFileSystem(fileSystem)
	originalAppPath := cmdtoolkit.SetApplicationPath("appPath")
	defer func() {
		writeFile = originalWriteFile
		rename = originalRename
		remove = originalRemove
		cmdtoolkit.AssignFileSystem(originalFileSystem)
		cmdtoolkit.SetApplicationPath(originalAppPath)
	}()
	tests := map[string]struct {
		preTest    func()
		postTest   func()
		efs        *exportSettings
		writeFile  func(string, []byte, fs.FileMode) error
		rename     func(string, string) error
		remove     func(string) error
		wantStatus *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"not asking to write": {
			preTest:  func() {},
			postTest: func() {},
			efs: &exportSettings{
				overwriteEnabled: cmdtoolkit.CommandFlag[bool]{Value: true},
				defaultsEnabled:  cmdtoolkit.CommandFlag[bool]{Value: false},
			},
			wantStatus: cmdtoolkit.NewExitUserError("export"),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"Default configuration settings will not be exported.\n" +
					"Why?\n" +
					"As currently configured, exporting default configuration settings is" +
					" disabled.\n" +
					"What to do:\n" +
					"To enable exporting defaults, use either:\n" +
					"● --defaults or\n" +
					"● --defaults=true\n",
				Log: "" +
					"level='error'" +
					" --defaults='false'" +
					" user-set='false'" +
					" msg='export defaults disabled'\n",
			},
		},
		"file does not exist but cannot be created": {
			preTest: func() {
				_ = fileSystem.MkdirAll(cmdtoolkit.ApplicationPath(), cmdtoolkit.StdDirPermissions)
			},
			postTest: func() {
				_ = fileSystem.RemoveAll(cmdtoolkit.ApplicationPath())
			},
			efs: &exportSettings{
				overwriteEnabled: cmdtoolkit.CommandFlag[bool]{Value: true},
				defaultsEnabled:  cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			writeFile: func(_ string, _ []byte, _ fs.FileMode) error {
				return fmt.Errorf("cannot write file, sorry")
			},
			wantStatus: cmdtoolkit.NewExitSystemError("export"),
			WantedRecording: output.WantedRecording{
				Error: "The file \"appPath\\\\defaults.yaml\" cannot be created:" +
					" 'cannot write file, sorry'.\n",
				Log: "" +
					"level='error'" +
					" command='export'" +
					" error='cannot write file, sorry'" +
					" fileName='appPath\\defaults.yaml'" +
					" msg='cannot create file'\n",
			},
		},
		"file does not exist": {
			preTest: func() {
				_ = fileSystem.MkdirAll(cmdtoolkit.ApplicationPath(), cmdtoolkit.StdDirPermissions)
			},
			postTest: func() {
				_ = fileSystem.RemoveAll(cmdtoolkit.ApplicationPath())
			},
			efs: &exportSettings{
				overwriteEnabled: cmdtoolkit.CommandFlag[bool]{Value: true},
				defaultsEnabled:  cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			writeFile:  func(_ string, _ []byte, _ fs.FileMode) error { return nil },
			wantStatus: nil,
			WantedRecording: output.WantedRecording{
				Console: "File \"appPath\\\\defaults.yaml\" has been written.\n",
			},
		},
		"file exists": {
			preTest: func() {
				_ = fileSystem.MkdirAll(cmdtoolkit.ApplicationPath(), cmdtoolkit.StdDirPermissions)
				_ = afero.WriteFile(
					fileSystem,
					filepath.Join(cmdtoolkit.ApplicationPath(), "defaults.yaml"),
					[]byte{},
					cmdtoolkit.StdFilePermissions,
				)
			},
			postTest: func() {
				_ = fileSystem.RemoveAll(cmdtoolkit.ApplicationPath())
			},
			efs: &exportSettings{
				overwriteEnabled: cmdtoolkit.CommandFlag[bool]{Value: true},
				defaultsEnabled:  cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			writeFile:  func(_ string, _ []byte, _ fs.FileMode) error { return nil },
			rename:     func(_, _ string) error { return nil },
			remove:     func(_ string) error { return nil },
			wantStatus: nil,
			WantedRecording: output.WantedRecording{
				Console: "File \"appPath\\\\defaults.yaml\" has been written.\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.preTest()
			defer tt.postTest()
			writeFile = tt.writeFile
			remove = tt.remove
			rename = tt.rename
			if got := tt.efs.exportDefaultConfiguration(o); !compareExitErrors(got, tt.wantStatus) {
				t.Errorf("exportFlagSettings.exportDefaultConfiguration() got %s want %s", got, tt.wantStatus)
			}
			o.Report(t, "exportFlagSettings.exportDefaultConfiguration()", tt.WantedRecording)
		})
	}
}

func Test_processExportFlags(t *testing.T) {
	tests := map[string]struct {
		values map[string]*cmdtoolkit.CommandFlag[any]
		want   *exportSettings
		want1  bool
		output.WantedRecording
	}{
		"nothing went right": {
			values: map[string]*cmdtoolkit.CommandFlag[any]{
				exportFlagDefaults:  {Value: "foo"},
				exportFlagOverwrite: {Value: "bar"},
			},
			want:  &exportSettings{},
			want1: false,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"defaults\" is not a boolean" +
					" (foo).\n" +
					"An internal error occurred: flag \"overwrite\" is not a boolean" +
					" (bar).\n",
				Log: "level='error'" +
					" error='flag value is not a boolean'" +
					" flag='defaults'" +
					" value='foo'" +
					" msg='internal error'\n" +
					"level='error'" +
					" error='flag value is not a boolean'" +
					" flag='overwrite'" +
					" value='bar'" +
					" msg='internal error'\n",
			},
		},
		"bad defaults settings": {
			values: map[string]*cmdtoolkit.CommandFlag[any]{
				exportFlagDefaults:  {Value: "foo"},
				exportFlagOverwrite: {Value: true},
			},
			want:  &exportSettings{overwriteEnabled: cmdtoolkit.CommandFlag[bool]{Value: true}},
			want1: false,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"defaults\" is not a boolean" +
					" (foo).\n",
				Log: "level='error'" +
					" error='flag value is not a boolean'" +
					" flag='defaults'" +
					" value='foo'" +
					" msg='internal error'\n",
			},
		},
		"bad overwrites settings": {
			values: map[string]*cmdtoolkit.CommandFlag[any]{
				exportFlagDefaults:  {Value: true},
				exportFlagOverwrite: {Value: 17},
			},
			want:  &exportSettings{defaultsEnabled: cmdtoolkit.CommandFlag[bool]{Value: true}},
			want1: false,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"overwrite\" is not a boolean" +
					" (17).\n",
				Log: "level='error'" +
					" error='flag value is not a boolean'" +
					" flag='overwrite'" +
					" value='17'" +
					" msg='internal error'\n",
			},
		},
		"everything good": {
			values: map[string]*cmdtoolkit.CommandFlag[any]{
				exportFlagDefaults:  {Value: true},
				exportFlagOverwrite: {Value: true},
			},
			want: &exportSettings{
				overwriteEnabled: cmdtoolkit.CommandFlag[bool]{Value: true},
				defaultsEnabled:  cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			want1: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, got1 := processExportFlags(o, tt.values)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processExportFlags() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("processExportFlags() got1 = %v, want %v", got1, tt.want1)
			}
			o.Report(t, "processExportFlags()", tt.WantedRecording)
		})
	}
}

func Test_exportRun(t *testing.T) {
	initGlobals()
	originalExportFlags := exportFlags
	originalBus := bus
	defer func() {
		exportFlags = originalExportFlags
		bus = originalBus
	}()
	tests := map[string]struct {
		cmd   *cobra.Command
		flags *cmdtoolkit.FlagSet
		output.WantedRecording
	}{
		"missing data": {
			cmd: exportCmd,
			flags: &cmdtoolkit.FlagSet{
				Name: exportCommand,
				Details: map[string]*cmdtoolkit.FlagDetails{
					exportFlagOverwrite: {
						ExpectedType: cmdtoolkit.BoolType,
						DefaultValue: 12,
					},
					exportFlagDefaults: nil,
				},
			},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: 'no details for flag \"defaults\"'.\n",
				Log: "level='error'" +
					" error='no details for flag \"defaults\"'" +
					" msg='internal error'\n",
			},
		},
		"incomplete data": {
			cmd: exportCmd,
			flags: &cmdtoolkit.FlagSet{
				Name: exportCommand,
				Details: map[string]*cmdtoolkit.FlagDetails{
					exportFlagOverwrite: {
						ExpectedType: cmdtoolkit.BoolType,
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
			cmd: exportCmd,
			flags: &cmdtoolkit.FlagSet{
				Name: exportCommand,
				Details: map[string]*cmdtoolkit.FlagDetails{
					exportFlagOverwrite: {
						ExpectedType: cmdtoolkit.BoolType,
						DefaultValue: false,
					},
					exportFlagDefaults: {
						ExpectedType: cmdtoolkit.BoolType,
						DefaultValue: false,
					},
				},
			},
			WantedRecording: output.WantedRecording{
				Error: "" +
					"Default configuration settings will not be exported.\n" +
					"Why?\n" +
					"As currently configured, exporting default configuration settings is disabled.\n" +
					"What to do:\n" +
					"To enable exporting defaults, use either:\n" +
					"● --defaults or\n" +
					"● --defaults=true\n",
				Log: "" +
					"level='error'" +
					" --defaults='false'" +
					" user-set='false'" +
					" msg='export defaults disabled'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			exportFlags = tt.flags
			o := output.NewRecorder()
			bus = o // this is what getBus() should return when exportRun calls it
			_ = exportRun(tt.cmd, []string{})
			o.Report(t, "exportRun()", tt.WantedRecording)
		})
	}
}

func Test_export_Help(t *testing.T) {
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
			command := exportCmd
			enableCommandRecording(o, command)
			_ = command.Help()
			o.Report(t, "export Help()", tt.WantedRecording)
		})
	}
}
