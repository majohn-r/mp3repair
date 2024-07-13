/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd

import (
	"fmt"
	"mp3repair/internal/files"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"testing"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

func Test_processListFlags(t *testing.T) {
	tests := map[string]struct {
		values map[string]*cmdtoolkit.CommandFlag[any]
		want   *listSettings
		want1  bool
		output.WantedRecording
	}{
		"no data": {
			values: map[string]*cmdtoolkit.CommandFlag[any]{},
			want:   &listSettings{},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"albums\" is not found.\n" +
					"An internal error occurred: flag \"annotate\" is not found.\n" +
					"An internal error occurred: flag \"artists\" is not found.\n" +
					"An internal error occurred: flag \"details\" is not found.\n" +
					"An internal error occurred: flag \"diagnostic\" is not found.\n" +
					"An internal error occurred: flag \"byNumber\" is not found.\n" +
					"An internal error occurred: flag \"byTitle\" is not found.\n" +
					"An internal error occurred: flag \"tracks\" is not found.\n",
				Log: "level='error'" +
					" error='flag not found'" +
					" flag='albums'" +
					" msg='internal error'\n" +
					"level='error'" +
					" error='flag not found'" +
					" flag='annotate'" +
					" msg='internal error'\n" +
					"level='error'" +
					" error='flag not found'" +
					" flag='artists'" +
					" msg='internal error'\n" +
					"level='error'" +
					" error='flag not found'" +
					" flag='details'" +
					" msg='internal error'\n" +
					"level='error'" +
					" error='flag not found'" +
					" flag='diagnostic'" +
					" msg='internal error'\n" +
					"level='error'" +
					" error='flag not found'" +
					" flag='byNumber'" +
					" msg='internal error'\n" +
					"level='error'" +
					" error='flag not found'" +
					" flag='byTitle'" +
					" msg='internal error'\n" +
					"level='error'" +
					" error='flag not found'" +
					" flag='tracks'" +
					" msg='internal error'\n",
			},
		},
		"configured": {
			values: map[string]*cmdtoolkit.CommandFlag[any]{
				"albums":     {Value: true},
				"annotate":   {Value: true},
				"artists":    {Value: true},
				"details":    {Value: true},
				"diagnostic": {Value: true},
				"byNumber":   {Value: true},
				"byTitle":    {Value: true},
				"tracks":     {Value: true},
			},
			want: &listSettings{
				albums:       cmdtoolkit.CommandFlag[bool]{Value: true},
				annotate:     cmdtoolkit.CommandFlag[bool]{Value: true},
				artists:      cmdtoolkit.CommandFlag[bool]{Value: true},
				details:      cmdtoolkit.CommandFlag[bool]{Value: true},
				diagnostic:   cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByTitle:  cmdtoolkit.CommandFlag[bool]{Value: true},
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			want1: true,
		},
		"user set": {
			values: map[string]*cmdtoolkit.CommandFlag[any]{
				"albums":     {Value: false, UserSet: true},
				"annotate":   {Value: false},
				"artists":    {Value: false, UserSet: true},
				"details":    {Value: false},
				"diagnostic": {Value: false},
				"byNumber":   {Value: false, UserSet: true},
				"byTitle":    {Value: false, UserSet: true},
				"tracks":     {Value: false, UserSet: true},
			},
			want: &listSettings{
				albums:       cmdtoolkit.CommandFlag[bool]{UserSet: true},
				artists:      cmdtoolkit.CommandFlag[bool]{UserSet: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{UserSet: true},
				sortByTitle:  cmdtoolkit.CommandFlag[bool]{UserSet: true},
				tracks:       cmdtoolkit.CommandFlag[bool]{UserSet: true},
			},
			want1: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, got1 := processListFlags(o, tt.values)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processListFlags() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("processListFlags() got1 = %v, want %v", got1, tt.want1)
			}
			o.Report(t, "processListFlags()", tt.WantedRecording)
		})
	}
}

func Test_listSettings_hasWorkToDo(t *testing.T) {
	tests := map[string]struct {
		ls   *listSettings
		want bool
		output.WantedRecording
	}{
		"none true, none explicitly set": {
			ls: &listSettings{},
			WantedRecording: output.WantedRecording{
				Error: "No listing will be output.\n" +
					"Why?\n" +
					"The flags --albums, --artists, and --tracks are all configured" +
					" false.\n" +
					"What to do:\n" +
					"Either:\n" +
					"[1] Edit the configuration file so that at least one of these flags" +
					" is true, or\n" +
					"[2] explicitly set at least one of these flags true on the command" +
					" line.\n",
			},
		},
		"none true, tracks explicitly set": {
			ls: &listSettings{tracks: cmdtoolkit.CommandFlag[bool]{UserSet: true}},
			WantedRecording: output.WantedRecording{
				Error: "No listing will be output.\n" +
					"Why?\n" +
					"In addition to --albums and --artists configured false, you" +
					" explicitly set --tracks false.\n" +
					"What to do:\n" +
					"Either:\n" +
					"[1] Edit the configuration file so that at least one of these flags" +
					" is true, or\n" +
					"[2] explicitly set at least one of these flags true on the command" +
					" line.\n",
			},
		},
		"none true, artists explicitly set": {
			ls: &listSettings{artists: cmdtoolkit.CommandFlag[bool]{UserSet: true}},
			WantedRecording: output.WantedRecording{
				Error: "No listing will be output.\n" +
					"Why?\n" +
					"In addition to --albums and --tracks configured false, you" +
					" explicitly set --artists false.\n" +
					"What to do:\n" +
					"Either:\n" +
					"[1] Edit the configuration file so that at least one of these flags" +
					" is true, or\n" +
					"[2] explicitly set at least one of these flags true on the command" +
					" line.\n",
			},
		},
		"none true, artists and tracks explicitly set": {
			ls: &listSettings{
				artists: cmdtoolkit.CommandFlag[bool]{UserSet: true},
				tracks:  cmdtoolkit.CommandFlag[bool]{UserSet: true},
			},
			WantedRecording: output.WantedRecording{
				Error: "No listing will be output.\n" +
					"Why?\n" +
					"In addition to --albums configured false, you explicitly set" +
					" --artists and --tracks false.\n" +
					"What to do:\n" +
					"Either:\n" +
					"[1] Edit the configuration file so that at least one of these flags" +
					" is true, or\n" +
					"[2] explicitly set at least one of these flags true on the command" +
					" line.\n",
			},
		},
		"none true, albums explicitly set": {
			ls: &listSettings{albums: cmdtoolkit.CommandFlag[bool]{UserSet: true}},
			WantedRecording: output.WantedRecording{
				Error: "No listing will be output.\n" +
					"Why?\n" +
					"In addition to --artists and --tracks configured false, you" +
					" explicitly set --albums false.\n" +
					"What to do:\n" +
					"Either:\n" +
					"[1] Edit the configuration file so that at least one of these flags" +
					" is true, or\n" +
					"[2] explicitly set at least one of these flags true on the command" +
					" line.\n",
			},
		},
		"none true, albums and tracks explicitly set": {
			ls: &listSettings{
				albums: cmdtoolkit.CommandFlag[bool]{UserSet: true},
				tracks: cmdtoolkit.CommandFlag[bool]{UserSet: true},
			},
			WantedRecording: output.WantedRecording{
				Error: "No listing will be output.\n" +
					"Why?\n" +
					"In addition to --artists configured false, you explicitly set" +
					" --albums and --tracks false.\n" +
					"What to do:\n" +
					"Either:\n" +
					"[1] Edit the configuration file so that at least one of these flags" +
					" is true, or\n" +
					"[2] explicitly set at least one of these flags true on the command" +
					" line.\n",
			},
		},
		"none true, albums and artists explicitly set": {
			ls: &listSettings{
				albums:  cmdtoolkit.CommandFlag[bool]{UserSet: true},
				artists: cmdtoolkit.CommandFlag[bool]{UserSet: true},
			},
			WantedRecording: output.WantedRecording{
				Error: "No listing will be output.\n" +
					"Why?\n" +
					"In addition to --tracks configured false, you explicitly set" +
					" --albums and --artists false.\n" +
					"What to do:\n" +
					"Either:\n" +
					"[1] Edit the configuration file so that at least one of these flags" +
					" is true, or\n" +
					"[2] explicitly set at least one of these flags true on the command" +
					" line.\n",
			},
		},
		"none true, albums and artists and tracks explicitly set": {
			ls: &listSettings{
				albums:  cmdtoolkit.CommandFlag[bool]{UserSet: true},
				artists: cmdtoolkit.CommandFlag[bool]{UserSet: true},
				tracks:  cmdtoolkit.CommandFlag[bool]{UserSet: true},
			},
			WantedRecording: output.WantedRecording{
				Error: "No listing will be output.\n" +
					"Why?\n" +
					"You explicitly set --albums, --artists, and --tracks false.\n" +
					"What to do:\n" +
					"Either:\n" +
					"[1] Edit the configuration file so that at least one of these flags" +
					" is true, or\n" +
					"[2] explicitly set at least one of these flags true on the command" +
					" line.\n",
			},
		},
		"tracks true": {
			ls:   &listSettings{tracks: cmdtoolkit.CommandFlag[bool]{Value: true}},
			want: true,
		},
		"artists true": {
			ls:   &listSettings{artists: cmdtoolkit.CommandFlag[bool]{Value: true}},
			want: true,
		},
		"artists and tracks true": {
			ls: &listSettings{
				artists: cmdtoolkit.CommandFlag[bool]{Value: true},
				tracks:  cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			want: true,
		},
		"albums true": {
			ls:   &listSettings{albums: cmdtoolkit.CommandFlag[bool]{Value: true}},
			want: true,
		},
		"albums and tracks true": {
			ls: &listSettings{
				albums: cmdtoolkit.CommandFlag[bool]{Value: true},
				tracks: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			want: true,
		},
		"albums and artists true": {
			ls: &listSettings{
				albums:  cmdtoolkit.CommandFlag[bool]{Value: true},
				artists: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			want: true,
		},
		"albums and artists and tracks true": {
			ls: &listSettings{
				albums:  cmdtoolkit.CommandFlag[bool]{Value: true},
				artists: cmdtoolkit.CommandFlag[bool]{Value: true},
				tracks:  cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if got := tt.ls.hasWorkToDo(o); got != tt.want {
				t.Errorf("listSettings.hasWorkToDo() = %v, want %v", got, tt.want)
			}
			o.Report(t, "listSettings.hasWorkToDo()", tt.WantedRecording)
		})
	}
}

func Test_listSettings_tracksSortable(t *testing.T) {
	tests := map[string]struct {
		ls      *listSettings
		want    bool
		lsFinal *listSettings
		output.WantedRecording
	}{
		// https://github.com/majohn-r/mp3repair/issues/170
		"-lrt --byTitle": {
			ls: &listSettings{
				albums:      cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				artists:     cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				tracks:      cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				sortByTitle: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
			},
			want: true,
			lsFinal: &listSettings{
				albums:      cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				artists:     cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				tracks:      cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				sortByTitle: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
			},
		},
		// https://github.com/majohn-r/mp3repair/issues/170
		"-lrt --byNumber": {
			ls: &listSettings{
				albums:       cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				artists:      cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
			},
			want: true,
			lsFinal: &listSettings{
				albums:       cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				artists:      cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
			},
		},
		// https://github.com/majohn-r/mp3repair/issues/170
		"-lt --byTitle": {
			ls: &listSettings{
				albums:      cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				tracks:      cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				sortByTitle: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
			},
			want: true,
			lsFinal: &listSettings{
				albums:      cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				tracks:      cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				sortByTitle: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
			},
		},
		// https://github.com/majohn-r/mp3repair/issues/170
		"-lt --byNumber": {
			ls: &listSettings{
				albums:       cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
			},
			want: true,
			lsFinal: &listSettings{
				albums:       cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
			},
		},
		// https://github.com/majohn-r/mp3repair/issues/170
		"-rt --byTitle": {
			ls: &listSettings{
				artists:     cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				tracks:      cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				sortByTitle: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
			},
			want: true,
			lsFinal: &listSettings{
				artists:     cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				tracks:      cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				sortByTitle: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
			},
		},
		// https://github.com/majohn-r/mp3repair/issues/170
		"-t --byTitle": {
			ls: &listSettings{
				tracks:      cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				sortByTitle: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
			},
			want: true,
			lsFinal: &listSettings{
				tracks:      cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				sortByTitle: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
			},
		},
		"tracks listed, both options set, neither explicitly": {
			ls: &listSettings{
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByTitle:  cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "Track sorting cannot be done.\n" +
					"Why?\n" +
					"The --byNumber and --byTitle flags are both configured true.\n" +
					"What to do:\n" +
					"Either edit the configuration file and use those default values, or" +
					" use appropriate command line values.\n",
			},
		},
		"tracks listed, both options set, by number explicitly": {
			ls: &listSettings{
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				sortByTitle:  cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "Track sorting cannot be done.\n" +
					"Why?\n" +
					"The --byTitle flag is configured true and you explicitly set" +
					" --byNumber true.\n" +
					"What to do:\n" +
					"Either edit the configuration file and use those default values, or" +
					" use appropriate command line values.\n",
			},
		},
		"tracks listed, both options set, by title explicitly": {
			ls: &listSettings{
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByTitle:  cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
			},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "Track sorting cannot be done.\n" +
					"Why?\n" +
					"The --byNumber flag is configured true and you explicitly set" +
					" --byTitle true.\n" +
					"What to do:\n" +
					"Either edit the configuration file and use those default values, or" +
					" use appropriate command line values.\n",
			},
		},
		"tracks listed, both options set, both explicitly": {
			ls: &listSettings{
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				sortByTitle:  cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
			},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "Track sorting cannot be done.\n" +
					"Why?\n" +
					"You explicitly set --byNumber and --byTitle true.\n" +
					"What to do:\n" +
					"Either edit the configuration file and use those default values, or" +
					" use appropriate command line values.\n",
			},
		},
		"tracks listed, no albums, sort by number, neither explicitly": {
			ls: &listSettings{
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "Sorting tracks by number not possible.\n" +
					"Why?\n" +
					"Track numbers are only relevant if albums are also output.\n" +
					"--albums is configured as false, and --byNumber is configured as" +
					" true.\n" +
					"What to do:\n" +
					"Either edit the configuration file or change which flags you set on" +
					" the command line.\n",
			},
		},
		"tracks listed, no albums, sort by number, albums explicitly": {
			ls: &listSettings{
				albums:       cmdtoolkit.CommandFlag[bool]{UserSet: true},
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "Sorting tracks by number not possible.\n" +
					"Why?\n" +
					"Track numbers are only relevant if albums are also output.\n" +
					"You set --albums false and --byNumber is configured as true.\n" +
					"What to do:\n" +
					"Either edit the configuration file or change which flags you set on" +
					" the command line.\n",
			},
		},
		"tracks listed, no albums, sort by number, sort explicitly": {
			ls: &listSettings{
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
			},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "Sorting tracks by number not possible.\n" +
					"Why?\n" +
					"Track numbers are only relevant if albums are also output.\n" +
					"You set --byNumber true and --albums is configured as false.\n" +
					"What to do:\n" +
					"Either edit the configuration file or change which flags you set on" +
					" the command line.\n",
			},
		},
		"tracks listed, no albums, sort by number, both explicitly": {
			ls: &listSettings{
				albums:       cmdtoolkit.CommandFlag[bool]{UserSet: true},
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
			},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "Sorting tracks by number not possible.\n" +
					"Why?\n" +
					"Track numbers are only relevant if albums are also output.\n" +
					"You set --byNumber true and --albums false.\n" +
					"What to do:\n" +
					"Either edit the configuration file or change which flags you set on" +
					" the command line.\n",
			},
		},
		"tracks listed, both sorting options explicitly false": {
			ls: &listSettings{
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{UserSet: true},
				sortByTitle:  cmdtoolkit.CommandFlag[bool]{UserSet: true},
			},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "A listing of tracks is not possible.\n" +
					"Why?\n" +
					"Tracks are enabled, but you set both --byNumber and --byTitle false.\n" +
					"What to do:\n" +
					"Enable one of the sorting flags.\n",
			},
		},
		"tracks listed, no sorting, user said no to number": {
			ls: &listSettings{
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{UserSet: true},
			},
			want: true,
			lsFinal: &listSettings{
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{UserSet: true},
				sortByTitle:  cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			WantedRecording: output.WantedRecording{
				Log: "level='info'" +
					" --albums='false'" +
					" --byTitle='true'" +
					" byNumber='false'" +
					" msg='no track sorting set, providing a sensible value'\n",
			},
		},
		"tracks listed, no sorting, user said no to title": {
			ls: &listSettings{
				tracks:      cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByTitle: cmdtoolkit.CommandFlag[bool]{UserSet: true},
			},
			want: true,
			lsFinal: &listSettings{
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByTitle:  cmdtoolkit.CommandFlag[bool]{UserSet: true},
			},
			WantedRecording: output.WantedRecording{
				Log: "level='info'" +
					" --albums='false'" +
					" --byTitle='false'" +
					" byNumber='true'" +
					" msg='no track sorting set, providing a sensible value'\n",
			},
		},
		"tracks listed, no sorting, albums included": {
			ls: &listSettings{
				albums: cmdtoolkit.CommandFlag[bool]{Value: true},
				tracks: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			want: true,
			lsFinal: &listSettings{
				albums:       cmdtoolkit.CommandFlag[bool]{Value: true},
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			WantedRecording: output.WantedRecording{
				Log: "level='info'" +
					" --albums='true'" +
					" --byTitle='false'" +
					" byNumber='true'" +
					" msg='no track sorting set, providing a sensible value'\n",
			},
		},
		"tracks listed, no sorting, no albums": {
			ls:   &listSettings{tracks: cmdtoolkit.CommandFlag[bool]{Value: true}},
			want: true,
			lsFinal: &listSettings{
				tracks:      cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByTitle: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			WantedRecording: output.WantedRecording{
				Log: "level='info'" +
					" --albums='false'" +
					" --byTitle='true'" +
					" byNumber='false'" +
					" msg='no track sorting set, providing a sensible value'\n",
			},
		},
		"tracks not listed, no sorting explicitly called for": {
			ls: &listSettings{
				sortByNumber: cmdtoolkit.CommandFlag[bool]{UserSet: true},
				sortByTitle:  cmdtoolkit.CommandFlag[bool]{UserSet: true},
			},
			want: true,
			lsFinal: &listSettings{
				sortByNumber: cmdtoolkit.CommandFlag[bool]{UserSet: true},
				sortByTitle:  cmdtoolkit.CommandFlag[bool]{UserSet: true},
			},
		},
		"tracks not listed, sort by number explicitly called for": {
			ls:   &listSettings{sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true}},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "Your sorting preferences are not relevant.\n" +
					"Why?\n" +
					"Tracks are not included in the output, but you explicitly set" +
					" --byNumber or --byTitle true.\n" +
					"What to do:\n" +
					"Either set --tracks true or remove the sorting flags from the" +
					" command line.\n",
			},
		},
		"tracks not listed, sort by title explicitly called for": {
			ls:   &listSettings{sortByTitle: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true}},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "Your sorting preferences are not relevant.\n" +
					"Why?\n" +
					"Tracks are not included in the output, but you explicitly set" +
					" --byNumber or --byTitle true.\nWhat to do:\n" +
					"Either set --tracks true or remove the sorting flags from the" +
					" command line.\n",
			},
		},
		"tracks not listed, sort by number and title explicitly called for": {
			ls: &listSettings{
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				sortByTitle:  cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
			},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "Your sorting preferences are not relevant.\n" +
					"Why?\n" +
					"Tracks are not included in the output, but you explicitly set" +
					" --byNumber or --byTitle true.\n" +
					"What to do:\n" +
					"Either set --tracks true or remove the sorting flags from the" +
					" command line.\n",
			},
		},
		"tracks listed, albums too, just sort by number": {
			ls: &listSettings{
				albums:       cmdtoolkit.CommandFlag[bool]{Value: true},
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			want: true,
			lsFinal: &listSettings{
				albums:       cmdtoolkit.CommandFlag[bool]{Value: true},
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
		},
		"tracks listed, just sort by title": {
			ls: &listSettings{
				tracks:      cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByTitle: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			want: true,
			lsFinal: &listSettings{
				tracks:      cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByTitle: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if got := tt.ls.tracksSortable(o); got != tt.want {
				t.Errorf("listSettings.tracksSortable() = %v, want %v", got, tt.want)
			}
			if tt.want {
				if *tt.ls != *tt.lsFinal {
					t.Errorf("listSettings.tracksSortable() ls = %v, want %v", tt.ls, tt.lsFinal)
				}
			}
			o.Report(t, "listSettings.tracksSortable()", tt.WantedRecording)
		})
	}
}

var (
	sampleTrack = files.TrackMaker{
		Album: files.AlbumMaker{
			Title:  "my album",
			Artist: files.NewArtist("my artist", "music/my artist"),
			Path:   "music/my artist/my album",
		}.NewAlbum(),
		FileName:   "10 track 10.mp3",
		SimpleName: "track 10",
		Number:     10,
	}.NewTrack()
	safeSearchFlags = &cmdtoolkit.FlagSet{
		Name: "search",
		Details: map[string]*cmdtoolkit.FlagDetails{
			searchAlbumFilter: {
				Usage:        "regular expression specifying which albums to select",
				ExpectedType: cmdtoolkit.StringType,
				DefaultValue: ".*",
			},
			searchArtistFilter: {
				Usage:        "regular expression specifying which artists to select",
				ExpectedType: cmdtoolkit.StringType,
				DefaultValue: ".*",
			},
			searchTrackFilter: {
				Usage:        "regular expression specifying which tracks to select",
				ExpectedType: cmdtoolkit.StringType,
				DefaultValue: ".*",
			},
			searchTopDir: {
				Usage:        "top directory specifying where to find mp3 files",
				ExpectedType: cmdtoolkit.StringType,
				DefaultValue: ".",
			},
			searchFileExtensions: {
				Usage:        "comma-delimited list of file extensions used by mp3" + " files",
				ExpectedType: cmdtoolkit.StringType,
				DefaultValue: ".mp3",
			},
		},
	}
)

func Test_showID3V1Diagnostics(t *testing.T) {
	type args struct {
		track *files.Track
		tags  []string
		err   error
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"with error": {
			args: args{
				track: sampleTrack,
				err:   fmt.Errorf("could not read track"),
			},
			WantedRecording: output.WantedRecording{
				Log: "level='error'" +
					" error='could not read track'" +
					" metadata='ID3V1'" +
					" track='music\\my artist\\my album\\10 track 10.mp3'" +
					" msg='metadata read error'\n",
			},
		},
		"without error": {
			args: args{
				track: sampleTrack,
				tags: []string{
					"artist=my artist",
					"album=my album",
					"track=track 10",
					"number=10",
				},
			},
			WantedRecording: output.WantedRecording{
				Console: "" +
					"  ID3V1 artist=my artist\n" +
					"  ID3V1 album=my album\n" +
					"  ID3V1 track=track 10\n" +
					"  ID3V1 number=10\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			o.IncrementTab(2)
			showID3V1Diagnostics(o, tt.args.track, tt.args.tags, tt.args.err)
			o.Report(t, "showID3V1Diagnostics()", tt.WantedRecording)
		})
	}
}

func Test_showID3V2Diagnostics(t *testing.T) {
	type args struct {
		track *files.Track
		info  *files.ID3V2Info
		err   error
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"error": {
			args: args{
				track: sampleTrack,
				err:   fmt.Errorf("no ID3V2 data found"),
			},
			WantedRecording: output.WantedRecording{
				Log: "level='error'" +
					" error='no ID3V2 data found'" +
					" metadata='ID3V2'" +
					" track='music\\my artist\\my album\\10 track 10.mp3'" +
					" msg='metadata read error'\n",
			},
		},
		"empty frames": {
			args: args{
				track: sampleTrack,
				info:  &files.ID3V2Info{Version: 1, Encoding: "UTF-8"},
			},
			WantedRecording: output.WantedRecording{
				Console: "" +
					"  ID3V2 Version: 1\n" +
					"  ID3V2 Encoding: \"UTF-8\"\n",
			},
		},
		"with frames": {
			args: args{
				track: sampleTrack,
				info: &files.ID3V2Info{
					Version:      1,
					Encoding:     "UTF-8",
					FrameStrings: []string{"FRAME1", "FRAME2"},
				},
			},
			WantedRecording: output.WantedRecording{
				Console: "" +
					"  ID3V2 Version: 1\n" +
					"  ID3V2 Encoding: \"UTF-8\"\n" +
					"  ID3V2 FRAME1\n" +
					"  ID3V2 FRAME2\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			o.IncrementTab(2)
			showID3V2Diagnostics(o, tt.args.track, tt.args.info, tt.args.err)
			o.Report(t, "showID3V2Diagnostics()", tt.WantedRecording)
		})
	}
}

func Test_listSettings_listTrackDiagnostics(t *testing.T) {
	tests := map[string]struct {
		ls    *listSettings
		track *files.Track
		output.WantedRecording
	}{
		"permitted": {
			ls:    &listSettings{diagnostic: cmdtoolkit.CommandFlag[bool]{Value: true}},
			track: sampleTrack,
			WantedRecording: output.WantedRecording{
				Log: "level='error'" +
					" error='open music\\my artist\\my album\\10 track 10.mp3: The system" +
					" cannot find the path specified.'" +
					" metadata='ID3V2'" +
					" track='music\\my artist\\my album\\10 track 10.mp3'" +
					" msg='metadata read error'\n" +
					"level='error'" +
					" error='open music\\my artist\\my album\\10 track 10.mp3: The system" +
					" cannot find the path specified.'" +
					" metadata='ID3V1'" +
					" track='music\\my artist\\my album\\10 track 10.mp3'" +
					" msg='metadata read error'\n",
			},
		},
		"not permitted": {
			ls: &listSettings{diagnostic: cmdtoolkit.CommandFlag[bool]{Value: false}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.ls.listTrackDiagnostics(o, tt.track)
			o.Report(t, "listSettings.listTrackDiagnostics()", tt.WantedRecording)
		})
	}
}

func Test_showDetails(t *testing.T) {
	type args struct {
		track        *files.Track
		details      map[string]string
		detailsError error
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"error": {
			args: args{
				track:        sampleTrack,
				detailsError: fmt.Errorf("details service offline"),
			},
			WantedRecording: output.WantedRecording{
				Error: "The details are not available for track \"track 10\" on album" +
					" \"my album\" by artist \"my artist\": \"details service offline\".\n",
				Log: "level='error'" +
					" error='details service offline'" +
					" track='music\\my artist\\my album\\10 track 10.mp3'" +
					" msg='cannot get details'\n",
			},
		},
		"no error, and no details": {args: args{track: sampleTrack}},
		"no error, with details": {
			args: args{
				track: sampleTrack,
				details: map[string]string{
					"composer": "some German",
					"producer": "A True Genius",
				},
			},
			WantedRecording: output.WantedRecording{
				Console: "" +
					"  Details:\n" +
					"    composer = \"some German\"\n" +
					"    producer = \"A True Genius\"\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			o.IncrementTab(2)
			showDetails(o, tt.args.track, tt.args.details, tt.args.detailsError)
			o.Report(t, "showDetails()", tt.WantedRecording)
		})
	}
}

func Test_listSettings_listTrackDetails(t *testing.T) {
	tests := map[string]struct {
		ls    *listSettings
		track *files.Track
		output.WantedRecording
	}{
		"not wanted": {ls: &listSettings{details: cmdtoolkit.CommandFlag[bool]{Value: false}}},
		"wanted": {
			ls:    &listSettings{details: cmdtoolkit.CommandFlag[bool]{Value: true}},
			track: sampleTrack,
			WantedRecording: output.WantedRecording{
				Error: "The details are not available for track \"track 10\" on album" +
					" \"my album\" by artist \"my artist\":" +
					" \"open music\\\\my artist\\\\my album\\\\10 track 10.mp3: The" +
					" system cannot find the path specified.\".\n",
				Log: "level='error'" +
					" error='open music\\my artist\\my album\\10 track 10.mp3: The system" +
					" cannot find the path specified.'" +
					" track='music\\my artist\\my album\\10 track 10.mp3'" +
					" msg='cannot get details'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.ls.listTrackDetails(o, tt.track)
			o.Report(t, "listSettings.listTrackDetails()", tt.WantedRecording)
		})
	}
}

func Test_listSettings_annotateTrackName(t *testing.T) {
	tests := map[string]struct {
		ls    *listSettings
		track *files.Track
		want  string
	}{
		"no annotations": {
			ls:    &listSettings{annotate: cmdtoolkit.CommandFlag[bool]{Value: false}},
			track: sampleTrack,
			want:  "track 10",
		},
		"annotations, albums printed": {
			ls: &listSettings{
				albums:   cmdtoolkit.CommandFlag[bool]{Value: true},
				annotate: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			track: sampleTrack,
			want:  "track 10",
		},
		"annotations, no albums, artists included": {
			ls: &listSettings{
				albums:   cmdtoolkit.CommandFlag[bool]{Value: false},
				artists:  cmdtoolkit.CommandFlag[bool]{Value: true},
				annotate: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			track: sampleTrack,
			want:  "\"track 10\" on \"my album\"",
		},
		"annotations, no albums, no artists": {
			ls: &listSettings{
				albums:   cmdtoolkit.CommandFlag[bool]{Value: false},
				artists:  cmdtoolkit.CommandFlag[bool]{Value: false},
				annotate: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			track: sampleTrack,
			want:  "\"track 10\" on \"my album\" by \"my artist\"",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.ls.annotateTrackName(tt.track); got != tt.want {
				t.Errorf("listSettings.annotateTrackName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func generateTracks(count int) []*files.Track {
	albums := generateAlbums(1, count)
	for _, album := range albums {
		return album.Tracks
	}
	return nil
}

func Test_listSettings_listTracksByName(t *testing.T) {
	tests := map[string]struct {
		ls     *listSettings
		tracks []*files.Track
		tab    uint8
		output.WantedRecording
	}{
		"no tracks": {
			ls:     &listSettings{},
			tracks: nil,
			tab:    2,
		},
		"multiple tracks": {
			ls:     &listSettings{annotate: cmdtoolkit.CommandFlag[bool]{Value: true}},
			tracks: generateTracks(25),
			tab:    0,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"\"my track 001\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 0010\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 0011\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 0012\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 0013\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 0014\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 0015\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 0016\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 0017\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 0018\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 0019\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 002\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 0020\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 0021\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 0022\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 0023\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 0024\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 0025\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 003\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 004\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 005\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 006\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 007\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 008\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 009\" on \"my album 00\" by \"my artist 0\"\n",
			},
		},
		"https://github.com/majohn-r/mp3repair/issues/147": {
			ls: &listSettings{annotate: cmdtoolkit.CommandFlag[bool]{Value: true}},
			tracks: []*files.Track{
				{
					SimpleName: "Old Brown Shoe",
					Album: &files.Album{
						Title:           "Anthology 3 [Disc 2]",
						RecordingArtist: &files.Artist{Name: "The Beatles"},
					},
				},
				{
					SimpleName: "Old Brown Shoe",
					Album: &files.Album{
						Title:           "Live In Japan [Disc 1]",
						RecordingArtist: &files.Artist{Name: "George Harrison & Eric Clapton"},
					},
				},
				{
					SimpleName: "Old Brown Shoe",
					Album: &files.Album{
						Title:           "Past Masters, Vol. 2",
						RecordingArtist: &files.Artist{Name: "The Beatles"},
					},
				},
				{
					SimpleName: "Old Brown Shoe",
					Album: &files.Album{
						Title:           "Songs From The Material World - A Tribute To George Harrison",
						RecordingArtist: &files.Artist{Name: "Various Artists"},
					},
				},
				{
					SimpleName: "Old Brown Shoe (Take 2)",
					Album: &files.Album{
						Title:           "Abbey Road- Sessions [Disc 2]",
						RecordingArtist: &files.Artist{Name: "The Beatles"},
					},
				},
			},
			tab: 0,
			WantedRecording: output.WantedRecording{
				Console: "" +
					`"Old Brown Shoe" on "Anthology 3 [Disc 2]" by "The Beatles"
"Old Brown Shoe" on "Live In Japan [Disc 1]" by "George Harrison & Eric Clapton"
"Old Brown Shoe" on "Past Masters, Vol. 2" by "The Beatles"
"Old Brown Shoe" on "Songs From The Material World - A Tribute To George Harrison" by "Various Artists"
"Old Brown Shoe (Take 2)" on "Abbey Road- Sessions [Disc 2]" by "The Beatles"
`,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			o.IncrementTab(tt.tab)
			tt.ls.listTracksByName(o, tt.tracks)
			o.Report(t, "listSettings.listTracksByName()", tt.WantedRecording)
		})
	}
}

func Test_listSettings_listTracksByNumber(t *testing.T) {
	tests := map[string]struct {
		ls     *listSettings
		tracks []*files.Track
		tab    uint8
		output.WantedRecording
	}{
		"no tracks": {
			ls: &listSettings{},
		},
		"lots of tracks": {
			ls:     &listSettings{},
			tracks: generateTracks(17),
			tab:    2,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"   1. my track 001\n" +
					"   2. my track 002\n" +
					"   3. my track 003\n" +
					"   4. my track 004\n" +
					"   5. my track 005\n" +
					"   6. my track 006\n" +
					"   7. my track 007\n" +
					"   8. my track 008\n" +
					"   9. my track 009\n" +
					"  10. my track 0010\n" +
					"  11. my track 0011\n" +
					"  12. my track 0012\n" +
					"  13. my track 0013\n" +
					"  14. my track 0014\n" +
					"  15. my track 0015\n" +
					"  16. my track 0016\n" +
					"  17. my track 0017\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			o.IncrementTab(tt.tab)
			tt.ls.listTracksByNumber(o, tt.tracks)
			o.Report(t, "listSettings.listTracksByNumber()", tt.WantedRecording)
		})
	}
}

func Test_listSettings_listTracks(t *testing.T) {
	tests := map[string]struct {
		ls     *listSettings
		tracks []*files.Track
		tab    uint8
		output.WantedRecording
	}{
		"no tracks": {
			ls: &listSettings{tracks: cmdtoolkit.CommandFlag[bool]{Value: true}},
		},
		"do not list tracks": {
			ls: &listSettings{
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: false},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			tracks: generateTracks(99),
		},
		"list tracks by number": {
			ls: &listSettings{
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			tracks: generateTracks(25),
			tab:    2,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"   1. my track 001\n" +
					"   2. my track 002\n" +
					"   3. my track 003\n" +
					"   4. my track 004\n" +
					"   5. my track 005\n" +
					"   6. my track 006\n" +
					"   7. my track 007\n" +
					"   8. my track 008\n" +
					"   9. my track 009\n" +
					"  10. my track 0010\n" +
					"  11. my track 0011\n" +
					"  12. my track 0012\n" +
					"  13. my track 0013\n" +
					"  14. my track 0014\n" +
					"  15. my track 0015\n" +
					"  16. my track 0016\n" +
					"  17. my track 0017\n" +
					"  18. my track 0018\n" +
					"  19. my track 0019\n" +
					"  20. my track 0020\n" +
					"  21. my track 0021\n" +
					"  22. my track 0022\n" +
					"  23. my track 0023\n" +
					"  24. my track 0024\n" +
					"  25. my track 0025\n",
			},
		},
		"list tracks by name": {
			ls: &listSettings{
				tracks:      cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByTitle: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			tracks: generateTracks(25),
			tab:    2,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"  my track 001\n" +
					"  my track 0010\n" +
					"  my track 0011\n" +
					"  my track 0012\n" +
					"  my track 0013\n" +
					"  my track 0014\n" +
					"  my track 0015\n" +
					"  my track 0016\n" +
					"  my track 0017\n" +
					"  my track 0018\n" +
					"  my track 0019\n" +
					"  my track 002\n" +
					"  my track 0020\n" +
					"  my track 0021\n" +
					"  my track 0022\n" +
					"  my track 0023\n" +
					"  my track 0024\n" +
					"  my track 0025\n" +
					"  my track 003\n" +
					"  my track 004\n" +
					"  my track 005\n" +
					"  my track 006\n" +
					"  my track 007\n" +
					"  my track 008\n" +
					"  my track 009\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			o.IncrementTab(tt.tab)
			tt.ls.listTracks(o, tt.tracks)
			o.Report(t, "listSettings.listTracks()", tt.WantedRecording)
		})
	}
}

func Test_listSettings_annotateAlbumName(t *testing.T) {
	tests := map[string]struct {
		ls   *listSettings
		want string
	}{
		"no annotation, no artist": {
			ls: &listSettings{
				artists:  cmdtoolkit.CommandFlag[bool]{Value: false},
				annotate: cmdtoolkit.CommandFlag[bool]{Value: false},
			},
			want: "my album",
		},
		"no annotation, with artist": {
			ls: &listSettings{
				artists:  cmdtoolkit.CommandFlag[bool]{Value: true},
				annotate: cmdtoolkit.CommandFlag[bool]{Value: false},
			},
			want: "my album",
		},
		"with annotation, no artist": {
			ls: &listSettings{
				artists:  cmdtoolkit.CommandFlag[bool]{Value: false},
				annotate: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			want: "\"my album\" by \"my artist\"",
		},
		"with annotation, with artist": {
			ls: &listSettings{
				artists:  cmdtoolkit.CommandFlag[bool]{Value: true},
				annotate: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			want: "my album",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			album := files.AlbumMaker{
				Title:  "my album",
				Artist: files.NewArtist("my artist", filepath.Join("Music", "my artist")),
				Path:   filepath.Join("Music", "my artist", "my album"),
			}.NewAlbum()
			if got := tt.ls.annotateAlbumName(album); got != tt.want {
				t.Errorf("listSettings.annotateAlbumName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func generateArtists(artistCount, albumCount, trackCount int) []*files.Artist {
	artists := make([]*files.Artist, 0)
	for r := 0; r < artistCount; r++ {
		artistName := fmt.Sprintf("my artist %d", r)
		artist := files.NewArtist(artistName, filepath.Join("Music", artistName))
		for k := 0; k < albumCount; k++ {
			albumName := fmt.Sprintf("my album %d%d", r, k)
			album := files.AlbumMaker{
				Title:  albumName,
				Artist: artist,
				Path:   filepath.Join("Music", "my artist", albumName),
			}.NewAlbum()
			for j := 1; j <= trackCount; j++ {
				trackName := fmt.Sprintf("my track %d%d%d", r, k, j)
				track := files.TrackMaker{
					Album:      album,
					FileName:   fmt.Sprintf("%d %s.mp3", j, trackName),
					SimpleName: trackName,
					Number:     j,
				}.NewTrack()
				album.AddTrack(track)
			}
			artist.AddAlbum(album)
		}
		artists = append(artists, artist)
	}
	return artists
}

func generateAlbums(albumCount, trackCount int) []*files.Album {
	artists := generateArtists(1, albumCount, trackCount)
	for _, artist := range artists {
		return artist.Albums
	}
	return nil
}

func Test_listSettings_listAlbums(t *testing.T) {
	tests := map[string]struct {
		ls     *listSettings
		albums []*files.Album
		tab    uint8
		output.WantedRecording
	}{
		"no albums": {
			ls:     &listSettings{},
			albums: nil,
			tab:    0,
		},
		"list albums without tracks": {
			ls:     &listSettings{albums: cmdtoolkit.CommandFlag[bool]{Value: true}},
			albums: generateAlbums(3, 3),
			tab:    2,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"  Album: my album 00\n" +
					"  Album: my album 01\n" +
					"  Album: my album 02\n",
			},
		},
		"list tracks only": {
			ls: &listSettings{
				artists:     cmdtoolkit.CommandFlag[bool]{Value: true},
				tracks:      cmdtoolkit.CommandFlag[bool]{Value: true},
				annotate:    cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByTitle: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			albums: generateAlbums(2, 2),
			tab:    2,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"  \"my track 001\" on \"my album 00\"\n" +
					"  \"my track 002\" on \"my album 00\"\n" +
					"  \"my track 011\" on \"my album 01\"\n" +
					"  \"my track 012\" on \"my album 01\"\n",
			},
		},
		"list albums and tracks": {
			ls: &listSettings{
				albums:       cmdtoolkit.CommandFlag[bool]{Value: true},
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true},
				annotate:     cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			albums: generateAlbums(3, 3),
			tab:    0,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Album: \"my album 00\" by \"my artist 0\"\n" +
					"   1. my track 001\n" +
					"   2. my track 002\n" +
					"   3. my track 003\n" +
					"Album: \"my album 01\" by \"my artist 0\"\n" +
					"   1. my track 011\n" +
					"   2. my track 012\n" +
					"   3. my track 013\n" +
					"Album: \"my album 02\" by \"my artist 0\"\n" +
					"   1. my track 021\n" +
					"   2. my track 022\n" +
					"   3. my track 023\n",
			},
		},
		"https://github.com/majohn-r/mp3repair/issues/147": {
			ls: &listSettings{
				albums:   cmdtoolkit.CommandFlag[bool]{Value: true},
				annotate: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			albums: []*files.Album{
				{
					Title:           "Live Rhymin' [Bonus Tracks]",
					RecordingArtist: &files.Artist{Name: "Paul Simon"},
				},
				{
					Title:           "Live In Paris & Toronto [Disc 2]",
					RecordingArtist: &files.Artist{Name: "Loreena McKennitt"},
				},
				{
					Title:           "Live In Paris & Toronto [Disc 1]",
					RecordingArtist: &files.Artist{Name: "Loreena McKennitt"},
				},
				{
					Title:           "Live In Japan [Disc 2]",
					RecordingArtist: &files.Artist{Name: "George Harrison & Eric Clapton"},
				},
				{
					Title:           "Live In Japan [Disc 1]",
					RecordingArtist: &files.Artist{Name: "George Harrison & Eric Clapton"},
				},
				{
					Title:           "Live From New York City, 1967",
					RecordingArtist: &files.Artist{Name: "Simon & Garfunkel"},
				},
				{
					Title:           "Live At The Circle Room",
					RecordingArtist: &files.Artist{Name: "Nat King Cole"},
				},
				{
					Title:           "Live At The BBC [Disc 2]",
					RecordingArtist: &files.Artist{Name: "The Beatles"},
				},
				{
					Title:           "Live At The BBC [Disc 1]",
					RecordingArtist: &files.Artist{Name: "The Beatles"},
				},
				{
					Title: "Live 1975-85 [Disc 3]",
					RecordingArtist: &files.Artist{
						Name: "Bruce Springsteen & The E Street Band",
					},
				},
				{
					Title: "Live 1975-85 [Disc 2]",
					RecordingArtist: &files.Artist{
						Name: "Bruce Springsteen & The E Street Band",
					},
				},
				{
					Title: "Live 1975-85 [Disc 1]",
					RecordingArtist: &files.Artist{
						Name: "Bruce Springsteen & The E Street Band",
					},
				},
				{
					Title:           "Live",
					RecordingArtist: &files.Artist{Name: "Roger Whittaker"},
				},
				{
					Title:           "Live",
					RecordingArtist: &files.Artist{Name: "Blondie"},
				},
				{
					Title:           "Live",
					RecordingArtist: &files.Artist{Name: "Big Bad Voodoo Daddy"},
				},
			},
			tab: 0,
			WantedRecording: output.WantedRecording{
				Console: "" +
					`Album: "Live" by "Big Bad Voodoo Daddy"
Album: "Live" by "Blondie"
Album: "Live" by "Roger Whittaker"
Album: "Live 1975-85 [Disc 1]" by "Bruce Springsteen & The E Street Band"
Album: "Live 1975-85 [Disc 2]" by "Bruce Springsteen & The E Street Band"
Album: "Live 1975-85 [Disc 3]" by "Bruce Springsteen & The E Street Band"
Album: "Live At The BBC [Disc 1]" by "The Beatles"
Album: "Live At The BBC [Disc 2]" by "The Beatles"
Album: "Live At The Circle Room" by "Nat King Cole"
Album: "Live From New York City, 1967" by "Simon & Garfunkel"
Album: "Live In Japan [Disc 1]" by "George Harrison & Eric Clapton"
Album: "Live In Japan [Disc 2]" by "George Harrison & Eric Clapton"
Album: "Live In Paris & Toronto [Disc 1]" by "Loreena McKennitt"
Album: "Live In Paris & Toronto [Disc 2]" by "Loreena McKennitt"
Album: "Live Rhymin' [Bonus Tracks]" by "Paul Simon"
`,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			o.IncrementTab(tt.tab)
			tt.ls.listAlbums(o, tt.albums)
			o.Report(t, "listSettings.listAlbums()", tt.WantedRecording)
		})
	}
}

func Test_listSettings_listFilteredArtists(t *testing.T) {
	tests := map[string]struct {
		ls      *listSettings
		artists []*files.Artist
		output.WantedRecording
	}{
		"no artists": {ls: &listSettings{}},
		"tracks": {
			ls: &listSettings{
				tracks:      cmdtoolkit.CommandFlag[bool]{Value: true},
				annotate:    cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByTitle: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			artists: generateArtists(3, 3, 3),
			WantedRecording: output.WantedRecording{
				Console: "" +
					"\"my track 001\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 002\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 003\" on \"my album 00\" by \"my artist 0\"\n" +
					"\"my track 011\" on \"my album 01\" by \"my artist 0\"\n" +
					"\"my track 012\" on \"my album 01\" by \"my artist 0\"\n" +
					"\"my track 013\" on \"my album 01\" by \"my artist 0\"\n" +
					"\"my track 021\" on \"my album 02\" by \"my artist 0\"\n" +
					"\"my track 022\" on \"my album 02\" by \"my artist 0\"\n" +
					"\"my track 023\" on \"my album 02\" by \"my artist 0\"\n" +
					"\"my track 101\" on \"my album 10\" by \"my artist 1\"\n" +
					"\"my track 102\" on \"my album 10\" by \"my artist 1\"\n" +
					"\"my track 103\" on \"my album 10\" by \"my artist 1\"\n" +
					"\"my track 111\" on \"my album 11\" by \"my artist 1\"\n" +
					"\"my track 112\" on \"my album 11\" by \"my artist 1\"\n" +
					"\"my track 113\" on \"my album 11\" by \"my artist 1\"\n" +
					"\"my track 121\" on \"my album 12\" by \"my artist 1\"\n" +
					"\"my track 122\" on \"my album 12\" by \"my artist 1\"\n" +
					"\"my track 123\" on \"my album 12\" by \"my artist 1\"\n" +
					"\"my track 201\" on \"my album 20\" by \"my artist 2\"\n" +
					"\"my track 202\" on \"my album 20\" by \"my artist 2\"\n" +
					"\"my track 203\" on \"my album 20\" by \"my artist 2\"\n" +
					"\"my track 211\" on \"my album 21\" by \"my artist 2\"\n" +
					"\"my track 212\" on \"my album 21\" by \"my artist 2\"\n" +
					"\"my track 213\" on \"my album 21\" by \"my artist 2\"\n" +
					"\"my track 221\" on \"my album 22\" by \"my artist 2\"\n" +
					"\"my track 222\" on \"my album 22\" by \"my artist 2\"\n" +
					"\"my track 223\" on \"my album 22\" by \"my artist 2\"\n",
			},
		},
		"albums": {
			ls: &listSettings{
				albums:   cmdtoolkit.CommandFlag[bool]{Value: true},
				annotate: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			artists: generateArtists(3, 3, 3),
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Album: \"my album 00\" by \"my artist 0\"\n" +
					"Album: \"my album 01\" by \"my artist 0\"\n" +
					"Album: \"my album 02\" by \"my artist 0\"\n" +
					"Album: \"my album 10\" by \"my artist 1\"\n" +
					"Album: \"my album 11\" by \"my artist 1\"\n" +
					"Album: \"my album 12\" by \"my artist 1\"\n" +
					"Album: \"my album 20\" by \"my artist 2\"\n" +
					"Album: \"my album 21\" by \"my artist 2\"\n" +
					"Album: \"my album 22\" by \"my artist 2\"\n",
			},
		},
		"albums and tracks": {
			ls: &listSettings{
				albums:       cmdtoolkit.CommandFlag[bool]{Value: true},
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true},
				annotate:     cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			artists: generateArtists(3, 3, 3),
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Album: \"my album 00\" by \"my artist 0\"\n" +
					"   1. my track 001\n" +
					"   2. my track 002\n" +
					"   3. my track 003\n" +
					"Album: \"my album 01\" by \"my artist 0\"\n" +
					"   1. my track 011\n" +
					"   2. my track 012\n" +
					"   3. my track 013\n" +
					"Album: \"my album 02\" by \"my artist 0\"\n" +
					"   1. my track 021\n" +
					"   2. my track 022\n" +
					"   3. my track 023\n" +
					"Album: \"my album 10\" by \"my artist 1\"\n" +
					"   1. my track 101\n" +
					"   2. my track 102\n" +
					"   3. my track 103\n" +
					"Album: \"my album 11\" by \"my artist 1\"\n" +
					"   1. my track 111\n" +
					"   2. my track 112\n" +
					"   3. my track 113\n" +
					"Album: \"my album 12\" by \"my artist 1\"\n" +
					"   1. my track 121\n" +
					"   2. my track 122\n" +
					"   3. my track 123\n" +
					"Album: \"my album 20\" by \"my artist 2\"\n" +
					"   1. my track 201\n" +
					"   2. my track 202\n" +
					"   3. my track 203\n" +
					"Album: \"my album 21\" by \"my artist 2\"\n" +
					"   1. my track 211\n" +
					"   2. my track 212\n" +
					"   3. my track 213\n" +
					"Album: \"my album 22\" by \"my artist 2\"\n" +
					"   1. my track 221\n" +
					"   2. my track 222\n" +
					"   3. my track 223\n",
			},
		},
		"artists": {
			ls:      &listSettings{artists: cmdtoolkit.CommandFlag[bool]{Value: true}},
			artists: generateArtists(3, 3, 3),
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Artist: my artist 0\n" +
					"Artist: my artist 1\n" +
					"Artist: my artist 2\n",
			},
		},
		"artists and tracks": {
			ls: &listSettings{
				artists:     cmdtoolkit.CommandFlag[bool]{Value: true},
				tracks:      cmdtoolkit.CommandFlag[bool]{Value: true},
				annotate:    cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByTitle: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			artists: generateArtists(3, 3, 3),
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Artist: my artist 0\n" +
					"  \"my track 001\" on \"my album 00\"\n" +
					"  \"my track 002\" on \"my album 00\"\n" +
					"  \"my track 003\" on \"my album 00\"\n" +
					"  \"my track 011\" on \"my album 01\"\n" +
					"  \"my track 012\" on \"my album 01\"\n" +
					"  \"my track 013\" on \"my album 01\"\n" +
					"  \"my track 021\" on \"my album 02\"\n" +
					"  \"my track 022\" on \"my album 02\"\n" +
					"  \"my track 023\" on \"my album 02\"\n" +
					"Artist: my artist 1\n" +
					"  \"my track 101\" on \"my album 10\"\n" +
					"  \"my track 102\" on \"my album 10\"\n" +
					"  \"my track 103\" on \"my album 10\"\n" +
					"  \"my track 111\" on \"my album 11\"\n" +
					"  \"my track 112\" on \"my album 11\"\n" +
					"  \"my track 113\" on \"my album 11\"\n" +
					"  \"my track 121\" on \"my album 12\"\n" +
					"  \"my track 122\" on \"my album 12\"\n" +
					"  \"my track 123\" on \"my album 12\"\n" +
					"Artist: my artist 2\n" +
					"  \"my track 201\" on \"my album 20\"\n" +
					"  \"my track 202\" on \"my album 20\"\n" +
					"  \"my track 203\" on \"my album 20\"\n" +
					"  \"my track 211\" on \"my album 21\"\n" +
					"  \"my track 212\" on \"my album 21\"\n" +
					"  \"my track 213\" on \"my album 21\"\n" +
					"  \"my track 221\" on \"my album 22\"\n" +
					"  \"my track 222\" on \"my album 22\"\n" +
					"  \"my track 223\" on \"my album 22\"\n",
			},
		},
		"artists and albums": {
			ls: &listSettings{
				albums:  cmdtoolkit.CommandFlag[bool]{Value: true},
				artists: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			artists: generateArtists(3, 3, 3),
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Artist: my artist 0\n" +
					"  Album: my album 00\n" +
					"  Album: my album 01\n" +
					"  Album: my album 02\n" +
					"Artist: my artist 1\n" +
					"  Album: my album 10\n" +
					"  Album: my album 11\n" +
					"  Album: my album 12\n" +
					"Artist: my artist 2\n" +
					"  Album: my album 20\n" +
					"  Album: my album 21\n" +
					"  Album: my album 22\n",
			},
		},
		"everything": {
			ls: &listSettings{
				artists:      cmdtoolkit.CommandFlag[bool]{Value: true},
				albums:       cmdtoolkit.CommandFlag[bool]{Value: true},
				tracks:       cmdtoolkit.CommandFlag[bool]{Value: true},
				sortByNumber: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			artists: generateArtists(3, 3, 3),
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Artist: my artist 0\n" +
					"  Album: my album 00\n" +
					"     1. my track 001\n" +
					"     2. my track 002\n" +
					"     3. my track 003\n" +
					"  Album: my album 01\n" +
					"     1. my track 011\n" +
					"     2. my track 012\n" +
					"     3. my track 013\n" +
					"  Album: my album 02\n" +
					"     1. my track 021\n" +
					"     2. my track 022\n" +
					"     3. my track 023\n" +
					"Artist: my artist 1\n" +
					"  Album: my album 10\n" +
					"     1. my track 101\n" +
					"     2. my track 102\n" +
					"     3. my track 103\n" +
					"  Album: my album 11\n" +
					"     1. my track 111\n" +
					"     2. my track 112\n" +
					"     3. my track 113\n" +
					"  Album: my album 12\n" +
					"     1. my track 121\n" +
					"     2. my track 122\n" +
					"     3. my track 123\n" +
					"Artist: my artist 2\n" +
					"  Album: my album 20\n" +
					"     1. my track 201\n" +
					"     2. my track 202\n" +
					"     3. my track 203\n" +
					"  Album: my album 21\n" +
					"     1. my track 211\n" +
					"     2. my track 212\n" +
					"     3. my track 213\n" +
					"  Album: my album 22\n" +
					"     1. my track 221\n" +
					"     2. my track 222\n" +
					"     3. my track 223\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.ls.listFilteredArtists(o, tt.artists)
			o.Report(t, "listSettings.listFilteredArtists()", tt.WantedRecording)
		})
	}
}

func Test_listRun(t *testing.T) {
	initGlobals()
	originalBus := bus
	originalSearchFlags := searchFlags
	defer func() {
		bus = originalBus
		searchFlags = originalSearchFlags
	}()
	searchFlags = safeSearchFlags

	testListFlags := &cmdtoolkit.FlagSet{
		Name: listCommand,
		Details: map[string]*cmdtoolkit.FlagDetails{
			listAlbums: {
				AbbreviatedName: "l",
				Usage:           "include album names in listing",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
			listArtists: {
				AbbreviatedName: "r",
				Usage:           "include artist names in listing",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    true,
			},
			listTracks: {
				AbbreviatedName: "t",
				Usage:           "include track names in listing",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
			listSortByNumber: {
				Usage:        "sort tracks by track number",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: false,
			},
			listSortByTitle: {
				Usage:        "sort tracks by track title",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: false,
			},
			listAnnotate: {
				Usage:        "annotate listings with album and artist names",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: false,
			},
			listDetails: {
				Usage:        "include details with tracks",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: false,
			},
			listDiagnostic: {
				Usage:        "include diagnostic information with tracks",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: false,
			},
		},
	}
	testCmd := &cobra.Command{}
	cmdtoolkit.AddFlags(output.NewNilBus(), cmdtoolkit.EmptyConfiguration(), testCmd.Flags(),
		testListFlags, searchFlags)

	testListFlags2 := &cmdtoolkit.FlagSet{
		Name: listCommand,
		Details: map[string]*cmdtoolkit.FlagDetails{
			listAlbums: {
				AbbreviatedName: "l",
				Usage:           "include album names in listing",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
			listArtists: {
				AbbreviatedName: "r",
				Usage:           "include artist names in listing",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    true,
			},
			listTracks: {
				AbbreviatedName: "t",
				Usage:           "include track names in listing",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    true,
			},
			listSortByNumber: {
				Usage:        "sort tracks by track number",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: true,
			},
			listSortByTitle: {
				Usage:        "sort tracks by track title",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: true,
			},
			listAnnotate: {
				Usage:        "annotate listings with album and artist names",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: false,
			},
			listDetails: {
				Usage:        "include details with tracks",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: false,
			},
			listDiagnostic: {
				Usage:        "include diagnostic information with tracks",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: false,
			},
		},
	}
	testCmd2 := &cobra.Command{}
	cmdtoolkit.AddFlags(output.NewNilBus(), cmdtoolkit.EmptyConfiguration(), testCmd2.Flags(),
		testListFlags2, searchFlags)

	testListFlags3 := &cmdtoolkit.FlagSet{
		Name: listCommand,
		Details: map[string]*cmdtoolkit.FlagDetails{
			listAlbums: {
				AbbreviatedName: "l",
				Usage:           "include album names in listing",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
			listArtists: {
				AbbreviatedName: "r",
				Usage:           "include artist names in listing",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
			listTracks: {
				AbbreviatedName: "t",
				Usage:           "include track names in listing",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
			listSortByNumber: {
				Usage:        "sort tracks by track number",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: false,
			},
			listSortByTitle: {
				Usage:        "sort tracks by track title",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: false,
			},
			listAnnotate: {
				Usage:        "annotate listings with album and artist names",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: false,
			},
			listDetails: {
				Usage:        "include details with tracks",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: false,
			},
			listDiagnostic: {
				Usage:        "include diagnostic information with tracks",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: false,
			},
		},
	}
	testCmd3 := &cobra.Command{}
	cmdtoolkit.AddFlags(output.NewNilBus(), cmdtoolkit.EmptyConfiguration(), testCmd3.Flags(),
		testListFlags3, searchFlags)

	tests := map[string]struct {
		cmd *cobra.Command
		in1 []string
		output.WantedRecording
	}{
		"typical": {
			cmd: testCmd,
			in1: nil,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No mp3 files could be found using the specified parameters.\n" +
					"Why?\n" +
					"There were no directories found in \".\" (the --topDir value).\n" +
					"What to do:\n" +
					"Set --topDir to the path of a directory that contains artist" +
					" directories.\n",
				Log: "" +
					"level='info'" +
					" --albumFilter='.*'" +
					" --albums='false'" +
					" --annotate='false'" +
					" --artistFilter='.*'" +
					" --artists='true'" +
					" --byNumber='false'" +
					" --byTitle='false'" +
					" --details='false'" +
					" --diagnostic='false'" +
					" --extensions='[.mp3]'" +
					" --topDir='.'" +
					" --trackFilter='.*'" +
					" --tracks='false'" +
					" albums-user-set='false'" +
					" artists-user-set='false'" +
					" byNumber-user-set='false'" +
					" byTitle-user-set='false'" +
					" command='list'" +
					" tracks-user-set='false'" +
					" msg='executing command'\n" +
					"level='error'" +
					" --topDir='.'" +
					" msg='cannot find any artist directories'\n",
			},
		},
		"typical but sorting is screwy": {
			cmd: testCmd2,
			in1: nil,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"Track sorting cannot be done.\n" +
					"Why?\n" +
					"The --byNumber and --byTitle flags are both configured true.\n" +
					"What to do:\n" +
					"Either edit the configuration file and use those default values, or" +
					" use appropriate command line values.\n",
				Log: "" +
					"level='info'" +
					" --albumFilter='.*'" +
					" --albums='false'" +
					" --annotate='false'" +
					" --artistFilter='.*'" +
					" --artists='true'" +
					" --byNumber='true'" +
					" --byTitle='true'" +
					" --details='false'" +
					" --diagnostic='false'" +
					" --extensions='[.mp3]'" +
					" --topDir='.'" +
					" --trackFilter='.*'" +
					" --tracks='true'" +
					" albums-user-set='false'" +
					" artists-user-set='false'" +
					" byNumber-user-set='false'" +
					" byTitle-user-set='false'" +
					" command='list'" +
					" tracks-user-set='false'" +
					" msg='executing command'\n",
			},
		},
		"no work to do": {
			cmd: testCmd3,
			in1: nil,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No listing will be output.\n" +
					"Why?\n" +
					"The flags --albums, --artists, and --tracks are all configured" +
					" false.\n" +
					"What to do:\n" +
					"Either:\n" +
					"[1] Edit the configuration file so that at least one of these flags" +
					" is true, or\n" +
					"[2] explicitly set at least one of these flags true on the command" +
					" line.\n",
				Log: "" +
					"level='info'" +
					" --albumFilter='.*'" +
					" --albums='false'" +
					" --annotate='false'" +
					" --artistFilter='.*'" +
					" --artists='false'" +
					" --byNumber='false'" +
					" --byTitle='false'" +
					" --details='false'" +
					" --diagnostic='false'" +
					" --extensions='[.mp3]'" +
					" --topDir='.'" +
					" --trackFilter='.*'" +
					" --tracks='false'" +
					" albums-user-set='false'" +
					" artists-user-set='false'" +
					" byNumber-user-set='false'" +
					" byTitle-user-set='false'" +
					" command='list'" +
					" tracks-user-set='false'" +
					" msg='executing command'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			bus = o // cook getBus()
			_ = listRun(tt.cmd, tt.in1)
			o.Report(t, "listRun()", tt.WantedRecording)
		})
	}
}

func compareExitErrors(e1, e2 *cmdtoolkit.ExitError) bool {
	if e1 == nil {
		return e2 == nil
	}
	if e2 == nil {
		return false
	}
	return e1.Error() == e2.Error()
}

func Test_listSettings_listArtists(t *testing.T) {
	type args struct {
		allArtists     []*files.Artist
		searchSettings *searchSettings
	}
	tests := map[string]struct {
		ls *listSettings
		args
		wantStatus *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"no data": {
			ls: &listSettings{artists: cmdtoolkit.CommandFlag[bool]{Value: true}},
			args: args{
				allArtists: nil,
				searchSettings: &searchSettings{
					artistFilter: regexp.MustCompile(".*"),
					albumFilter:  regexp.MustCompile(".*"),
					trackFilter:  regexp.MustCompile(".*"),
				},
			},
			wantStatus: cmdtoolkit.NewExitUserError(listCommand),
			// note: no error or log output; that would have been handled by
			// loading artists resulting in no artists
		},
		"with data": {
			ls: &listSettings{artists: cmdtoolkit.CommandFlag[bool]{Value: true}},
			args: args{
				allArtists: generateArtists(3, 4, 5),
				searchSettings: &searchSettings{
					artistFilter: regexp.MustCompile(".*"),
					albumFilter:  regexp.MustCompile(".*"),
					trackFilter:  regexp.MustCompile(".*"),
				},
			},
			wantStatus: nil,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Artist: my artist 0\n" +
					"Artist: my artist 1\n" +
					"Artist: my artist 2\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got := tt.ls.listArtists(o, tt.args.allArtists, tt.args.searchSettings)
			if !compareExitErrors(got, tt.wantStatus) {
				t.Errorf("listSettings.listArtists() got %s want %s", got, tt.wantStatus)
			}
			o.Report(t, "listSettings.listArtists()", tt.WantedRecording)
		})
	}
}

func Test_list_Help(t *testing.T) {
	originalSearchFlags := searchFlags
	defer func() {
		searchFlags = originalSearchFlags
	}()
	searchFlags = safeSearchFlags
	commandUnderTest := cloneCommand(listCmd)
	cmdtoolkit.AddFlags(output.NewNilBus(), cmdtoolkit.EmptyConfiguration(),
		commandUnderTest.Flags(), listFlags, searchFlags)
	tests := map[string]struct {
		output.WantedRecording
	}{
		"good": {
			WantedRecording: output.WantedRecording{
				Console: "" +
					"\"list\" lists mp3 files and containing album and artist" +
					" directories\n" +
					"\n" +
					"Usage:\n" +
					"  list [--albums] [--artists] [--tracks] [--annotate] [--details]" +
					" [--diagnostic] [--byNumber | --byTitle] [--albumFilter regex]" +
					" [--artistFilter regex] [--trackFilter regex] [--topDir dir]" +
					" [--extensions extensions]\n" +
					"\n" +
					"Examples:\n" +
					"list --annotate\n" +
					"  Annotate tracks with album and artist data and albums with artist" +
					" data\n" +
					"list --details\n" +
					"  Include detailed information, if available, for each track. This" +
					" includes composer,\n" +
					"  conductor, key, lyricist, orchestra/band, and subtitle\n" +
					"list --albums\n" +
					"  Include the album names in the output\n" +
					"list --artists\n" +
					"  Include the artist names in the output\n" +
					"list --tracks\n" +
					"  Include the track names in the output\n" +
					"list --byTitle\n" +
					"  Sort tracks by name, ignoring track numbers\n" +
					"list --byNumber\n" +
					"  Sort tracks by track number\n" +
					"\n" +
					"Flags:\n" +
					"      --albumFilter string    " +
					"regular expression specifying which albums to select (default \".*\")\n" +
					"  -l, --albums                " +
					"include album names in listing (default false)\n" +
					"      --annotate              " +
					"annotate listings with album and artist names (default false)\n" +
					"      --artistFilter string   " +
					"regular expression specifying which artists to select (default \".*\")\n" +
					"  -r, --artists               " +
					"include artist names in listing (default false)\n" +
					"      --byNumber              " +
					"sort tracks by track number (default false)\n" +
					"      --byTitle               " +
					"sort tracks by track title (default false)\n" +
					"      --details               " +
					"include details with tracks (default false)\n" +
					"      --diagnostic            " +
					"include diagnostic information with tracks (default false)\n" +
					"      --extensions string     " +
					"comma-delimited list of file extensions used by mp3 files (default \".mp3\")\n" +
					"      --topDir string         " +
					"top directory specifying where to find mp3 files (default \".\")\n" +
					"      --trackFilter string    " +
					"regular expression specifying which tracks to select (default \".*\")\n" +
					"  -t, --tracks                " +
					"include track names in listing (default false)\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			command := commandUnderTest
			enableCommandRecording(o, command)
			_ = command.Help()
			o.Report(t, "list Help()", tt.WantedRecording)
		})
	}
}

func Test_trackSlice_sort(t *testing.T) {
	tests := map[string]struct {
		ts   []*files.Track
		want []*files.Track
	}{
		"https://github.com/majohn-r/mp3repair/issues/147": {
			ts: []*files.Track{
				{
					SimpleName: "b",
					Album: &files.Album{
						Title:           "b",
						RecordingArtist: &files.Artist{Name: "c"},
					},
				},
				{
					SimpleName: "b",
					Album: &files.Album{
						Title:           "a",
						RecordingArtist: &files.Artist{Name: "c"},
					},
				},
				{
					SimpleName: "b",
					Album: &files.Album{
						Title:           "b",
						RecordingArtist: &files.Artist{Name: "a"},
					},
				},
				{
					SimpleName: "a",
					Album: &files.Album{
						Title:           "b",
						RecordingArtist: &files.Artist{Name: "c"},
					},
				},
				{
					SimpleName: "a",
					Album: &files.Album{
						Title:           "a",
						RecordingArtist: &files.Artist{Name: "c"},
					},
				},
				{
					SimpleName: "a",
					Album: &files.Album{
						Title:           "b",
						RecordingArtist: &files.Artist{Name: "a"},
					},
				},
			},
			want: []*files.Track{
				{
					SimpleName: "a",
					Album: &files.Album{
						Title:           "a",
						RecordingArtist: &files.Artist{Name: "c"},
					},
				},
				{
					SimpleName: "a",
					Album: &files.Album{
						Title:           "b",
						RecordingArtist: &files.Artist{Name: "a"},
					},
				},
				{
					SimpleName: "a",
					Album: &files.Album{
						Title:           "b",
						RecordingArtist: &files.Artist{Name: "c"},
					},
				},
				{
					SimpleName: "b",
					Album: &files.Album{
						Title:           "a",
						RecordingArtist: &files.Artist{Name: "c"},
					},
				},
				{
					SimpleName: "b",
					Album: &files.Album{
						Title:           "b",
						RecordingArtist: &files.Artist{Name: "a"},
					},
				},
				{
					SimpleName: "b",
					Album: &files.Album{
						Title:           "b",
						RecordingArtist: &files.Artist{Name: "c"},
					},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			sort.Sort(trackSlice(tt.ts))
			if !reflect.DeepEqual(tt.ts, tt.want) {
				t.Errorf("trackSlice.sort = %v, want %v", tt.ts, tt.want)
			}
		})
	}
}

func Test_albumSlice_sort(t *testing.T) {
	tests := map[string]struct {
		ts   []*files.Album
		want []*files.Album
	}{
		"https://github.com/majohn-r/mp3repair/issues/147": {
			ts: []*files.Album{
				{
					Title:           "b",
					RecordingArtist: &files.Artist{Name: "c"},
				},
				{
					Title:           "a",
					RecordingArtist: &files.Artist{Name: "c"},
				},
				{
					Title:           "b",
					RecordingArtist: &files.Artist{Name: "a"},
				},
			},
			want: []*files.Album{
				{
					Title:           "a",
					RecordingArtist: &files.Artist{Name: "c"},
				},
				{
					Title:           "b",
					RecordingArtist: &files.Artist{Name: "a"},
				},
				{
					Title:           "b",
					RecordingArtist: &files.Artist{Name: "c"},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			sort.Sort(albumSlice(tt.ts))
			if !reflect.DeepEqual(tt.ts, tt.want) {
				t.Errorf("albumSlice.sort = %v, want %v", tt.ts, tt.want)
			}
		})
	}
}
