/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd_test

import (
	"mp3/cmd"
	"testing"

	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

func TestAboutRun(t *testing.T) {
	originalBusGetter := cmd.BusGetter
	originalLogCommandStart := cmd.LogCommandStart
	originalGenerateAboutContent := cmd.GenerateAboutContent
	originalExit := cmd.Exit
	defer func() {
		cmd.BusGetter = originalBusGetter
		cmd.LogCommandStart = originalLogCommandStart
		cmd.GenerateAboutContent = originalGenerateAboutContent
		cmd.Exit = originalExit
	}()
	var exitCalled bool
	var exitCode int
	cmd.Exit = func(code int) {
		exitCalled = true
		exitCode = code
	}
	type args struct {
		in0 *cobra.Command
		in1 []string
	}
	tests := map[string]struct {
		args
		wantExitCode   int
		wantExitCalled bool
		output.WantedRecording
	}{
		"simple": {
			WantedRecording: output.WantedRecording{
				Console: "About content here.\n",
				Log:     "level='info' command='about' msg='executing command'\n",
			},
			wantExitCode:   0,
			wantExitCalled: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			exitCalled = false
			exitCode = -1
			o := output.NewRecorder()
			cmd.BusGetter = func() output.Bus { return o }
			cmd.LogCommandStart = func(bus output.Bus, cmdName string, args map[string]any) {
				bus.Log(output.Info, "executing command", map[string]any{"command": "about"})
			}
			cmd.GenerateAboutContent = func(bus output.Bus) {
				bus.WriteCanonicalConsole("about content here")
			}
			cmd.AboutRun(tt.args.in0, tt.args.in1)
			if got := exitCode; got != tt.wantExitCode {
				t.Errorf("AboutRun() got exit code %d want %d", got, tt.wantExitCode)
			}
			if got := exitCalled; got != tt.wantExitCalled {
				t.Errorf("AboutRun() got exit called %t want %t", got, tt.wantExitCalled)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("AboutRun() %s", issue)
				}
			}
		})
	}
}

func TestInitializeAbout(t *testing.T) {
	tests := map[string]struct {
		version  string
		creation string
	}{
		"good": {version: "0.1.1", creation: "2006-01-02T15:04:05Z07:00"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmd.InitializeAbout(tt.version, tt.creation)
		})
	}
}

func enableCommandRecording(o *output.Recorder, command *cobra.Command) {
	command.SetErr(o.ErrorWriter())
	command.SetOutput(o.ConsoleWriter())
}

func TestAboutHelp(t *testing.T) {
	tests := map[string]struct {
		output.WantedRecording
	}{
		"good": {
			WantedRecording: output.WantedRecording{
				Console: "" +
					"\"about\" provides the following information about the mp3 program:\n" +
					"\n" +
					"* The program version\n" +
					"* Copyright information\n" +
					"* Build information:\n" +
					"  * The build timestamp\n" +
					"  * The version of go used to compile the code\n" +
					"  * A list of dependencies and their versions\n" +
					"\n" +
					"Usage:\n" +
					"  mp3 about\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			command := cmd.AboutCmd
			enableCommandRecording(o, command)
			command.Help()
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("about Help() %s", issue)
				}
			}
		})
	}
}
