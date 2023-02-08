package main

import (
	"mp3/internal/commands"
	"testing"

	"github.com/majohn-r/output"
)

func Test_main(t *testing.T) {
	savedExecFunc := execFunc
	savedExitFunc := exitFunc
	defer func() {
		execFunc = savedExecFunc
		exitFunc = savedExitFunc
	}()
	tests := map[string]struct {
		execFunc     func(output.Bus, int, string, string, string, []string) int
		wantDefault  bool
		wantExitCode int
	}{
		"failure": {
			execFunc: func(_ output.Bus, _ int, _ string, _ string, _ string, _ []string) int {
				return 1
			},
			wantDefault:  true,
			wantExitCode: 1,
		},
		"success": {
			execFunc: func(_ output.Bus, _ int, _ string, _ string, _ string, _ []string) int {
				return 0
			},
			wantDefault:  true,
			wantExitCode: 0,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			execFunc = tt.execFunc
			var capturedExitCode int
			exitFunc = func(exitCode int) {
				capturedExitCode = exitCode
			}
			main()
			if got := commands.IsDefault(defaultCommand); got != tt.wantDefault {
				t.Errorf("main() got default command %t want %t", got, tt.wantDefault)
			}
			if capturedExitCode != tt.wantExitCode {
				t.Errorf("main() got exit code %d want %d", capturedExitCode, tt.wantExitCode)
			}
		})
	}
}
