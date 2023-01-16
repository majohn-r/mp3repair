package internal

import (
	"fmt"
	"testing"

	"github.com/majohn-r/output"
)

func TestReportInvalidConfigurationData(t *testing.T) {
	const fnName = "ReportInvalidConfigurationData()"
	type args struct {
		s string
		e error
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"typical case": {
			args: args{s: "badSection", e: fmt.Errorf("bad content")},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"badSection\": bad content.\n",
				Log:   "level='error' error='bad content' section='badSection' msg='invalid content in configuration file'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			ReportInvalidConfigurationData(o, tt.args.s, tt.args.e)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s: %s", fnName, issue)
				}
			}
		})
	}
}
