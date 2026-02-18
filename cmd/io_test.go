/*
 * Copyright © 2026 Marc Johnson (marc.johnson27591@gmail.com)
 */

package cmd

import (
	"reflect"
	"testing"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
)

func Test_evaluateIOFlags(t *testing.T) {
	tests := map[string]struct {
		producer cmdtoolkit.FlagProducer
		want     *ioSettings
		want1    bool
		output.WantedRecording
	}{
		"error": {
			producer: testFlagProducer{},
			want:     &ioSettings{},
			want1:    false,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: 'flag \"maxOpenFiles\" does not exist'.\n",
				Log:   "level='error' error='flag \"maxOpenFiles\" does not exist' msg='internal error'\n",
			},
		},
		"good data": {
			producer: testFlagProducer{
				flags: map[string]testFlag{"maxOpenFiles": {value: 25, valueKind: cmdtoolkit.IntType}},
			},
			want:  &ioSettings{openFileLimit: 25},
			want1: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, got1 := evaluateIOFlags(o, tt.producer)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("evaluateIOFlags() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("evaluateIOFlags() got1 = %v, want %v", got1, tt.want1)
			}
			o.Report(t, "evaluateIOFlags()", tt.WantedRecording)
		})
	}
}

func Test_processIOFlags(t *testing.T) {
	tests := map[string]struct {
		values map[string]*cmdtoolkit.CommandFlag[any]
		want   *ioSettings
		want1  bool
		output.WantedRecording
	}{
		"no data": {
			values: map[string]*cmdtoolkit.CommandFlag[any]{},
			want:   &ioSettings{},
			want1:  false,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"maxOpenFiles\" is not found.\n",
				Log:   "level='error' error='flag not found' flag='maxOpenFiles' msg='internal error'\n",
			},
		},
		"default value": {
			values: map[string]*cmdtoolkit.CommandFlag[any]{"maxOpenFiles": {Value: 1000}},
			want:   &ioSettings{openFileLimit: 1000},
			want1:  true,
		},
		"low value": {
			values: map[string]*cmdtoolkit.CommandFlag[any]{"maxOpenFiles": {Value: ioOpenFileMinimum - 1}},
			want:   &ioSettings{openFileLimit: ioOpenFileMinimum},
			want1:  true,
			WantedRecording: output.WantedRecording{
				Log: "level='warning'" +
					" flag='--maxOpenFiles'" +
					" providedValue='0'" +
					" replacedBy='1'" +
					" msg='user-supplied value replaced'\n",
			},
		},
		"high value": {
			values: map[string]*cmdtoolkit.CommandFlag[any]{"maxOpenFiles": {Value: ioOpenFileMaximum + 1}},
			want:   &ioSettings{openFileLimit: ioOpenFileMaximum},
			want1:  true,
			WantedRecording: output.WantedRecording{
				Log: "level='warning'" +
					" flag='--maxOpenFiles'" +
					" providedValue='32768'" +
					" replacedBy='32767'" +
					" msg='user-supplied value replaced'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, got1 := processIOFlags(o, tt.values)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processIOFlags() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("processIOFlags() got1 = %v, want %v", got1, tt.want1)
			}
			o.Report(t, "processIOFlags()", tt.WantedRecording)
		})
	}
}
