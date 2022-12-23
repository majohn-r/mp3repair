package internal

import (
	"fmt"
	"testing"

	"github.com/majohn-r/output"
)

func TestLogInvalidConfigurationData(t *testing.T) {
	type args struct {
		s string
		e error
	}
	tests := []struct {
		name string
		args
		output.WantedRecording
	}{
		{
			name: "typical case",
			args: args{s: "badSection", e: fmt.Errorf("bad content")},
			WantedRecording: output.WantedRecording{
				Log: "level='error' error='bad content' section='badSection' msg='invalid content in configuration file'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			LogInvalidConfigurationData(o, tt.args.s, tt.args.e)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("LogInvalidConfigurationData: %s", issue)
				}
			}
		})
	}
}
