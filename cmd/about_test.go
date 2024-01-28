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
	oldBusGetter := cmd.BusGetter
	oldCommandStartLogger := cmd.CommandStartLogger
	oldContentGenerator := cmd.AboutContentGenerator
	defer func() {
		cmd.BusGetter = oldBusGetter
		cmd.CommandStartLogger = oldCommandStartLogger
		cmd.AboutContentGenerator = oldContentGenerator
	}()
	type args struct {
		in0 *cobra.Command
		in1 []string
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"simple": {
			WantedRecording: output.WantedRecording{
				Console: "About content here.\n",
				Log:     "level='info' command='about' msg='executing command'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			cmd.BusGetter = func() output.Bus { return o }
			cmd.CommandStartLogger = func(bus output.Bus, cmdName string, args map[string]any) {
				bus.Log(output.Info, "executing command", map[string]any{"command": "about"})
			}
			cmd.AboutContentGenerator = func(bus output.Bus) {
				bus.WriteCanonicalConsole("about content here")
			}
			cmd.AboutRun(tt.args.in0, tt.args.in1)
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
