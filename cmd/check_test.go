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
	"testing"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

func Test_processCheckFlags(t *testing.T) {
	tests := map[string]struct {
		values map[string]*cmdtoolkit.CommandFlag[any]
		want   *checkSettings
		want1  bool
		output.WantedRecording
	}{
		"no data": {
			values: map[string]*cmdtoolkit.CommandFlag[any]{},
			want:   &checkSettings{},
			want1:  false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"An internal error occurred: flag \"empty\" is not found.\n" +
					"An internal error occurred: flag \"files\" is not found.\n" +
					"An internal error occurred: flag \"numbering\" is not found.\n",
				Log: "" +
					"level='error'" +
					" error='flag not found'" +
					" flag='empty'" +
					" msg='internal error'\n" +
					"level='error'" +
					" error='flag not found'" +
					" flag='files'" +
					" msg='internal error'\n" +
					"level='error'" +
					" error='flag not found'" +
					" flag='numbering'" +
					" msg='internal error'\n",
			},
		},
		"out of the box": {
			values: map[string]*cmdtoolkit.CommandFlag[any]{
				"empty":     {Value: false},
				"files":     {Value: false},
				"numbering": {Value: false},
			},
			want:  &checkSettings{},
			want1: true,
		},
		"overridden": {
			values: map[string]*cmdtoolkit.CommandFlag[any]{
				"empty":     {Value: true, UserSet: true},
				"files":     {Value: true, UserSet: true},
				"numbering": {Value: true, UserSet: true},
			},
			want: &checkSettings{
				empty:     cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				files:     cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
				numbering: cmdtoolkit.CommandFlag[bool]{Value: true, UserSet: true},
			},
			want1: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, got1 := processCheckFlags(o, tt.values)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processCheckFlags() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("processCheckFlags() got1 = %v, want %v", got1, tt.want1)
			}
			o.Report(t, "processCheckFlags()", tt.WantedRecording)
		})
	}
}

func Test_checkSettings_hasWorkToDo(t *testing.T) {
	tests := map[string]struct {
		cs   *checkSettings
		want bool
		output.WantedRecording
	}{
		"no work, as configured": {
			cs:   &checkSettings{},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No checks will be executed.\n" +
					"Why?\n" +
					"The flags --empty, --files, and --numbering are all configured false.\n" +
					"What to do:\n" +
					"Either:\n" +
					" 1. Edit the configuration file so that at least one of these flags is true, or\n" +
					" 2. Explicitly set at least one of these flags true on the command line.\n",
			},
		},
		"no work, empty configured that way": {
			cs:   &checkSettings{empty: cmdtoolkit.CommandFlag[bool]{UserSet: true}},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No checks will be executed.\n" +
					"Why?\n" +
					"In addition to --files and --numbering configured false, you explicitly set --empty false.\n" +
					"What to do:\n" +
					"Either:\n" +
					" 1. Edit the configuration file so that at least one of these flags is true, or\n" +
					" 2. Explicitly set at least one of these flags true on the command line.\n",
			},
		},
		"no work, files configured that way": {
			cs:   &checkSettings{files: cmdtoolkit.CommandFlag[bool]{UserSet: true}},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No checks will be executed.\n" +
					"Why?\n" +
					"In addition to --empty and --numbering configured false, you explicitly set --files false.\n" +
					"What to do:\n" +
					"Either:\n" +
					" 1. Edit the configuration file so that at least one of these flags is true, or\n" +
					" 2. Explicitly set at least one of these flags true on the command line.\n",
			},
		},
		"no work, numbering configured that way": {
			cs:   &checkSettings{numbering: cmdtoolkit.CommandFlag[bool]{UserSet: true}},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No checks will be executed.\n" +
					"Why?\n" +
					"In addition to --empty and --files configured false, you explicitly set --numbering false.\n" +
					"What to do:\n" +
					"Either:\n" +
					" 1. Edit the configuration file so that at least one of these flags is true, or\n" +
					" 2. Explicitly set at least one of these flags true on the command line.\n",
			},
		},
		"no work, empty and files configured that way": {
			cs: &checkSettings{
				empty: cmdtoolkit.CommandFlag[bool]{UserSet: true},
				files: cmdtoolkit.CommandFlag[bool]{UserSet: true},
			},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No checks will be executed.\n" +
					"Why?\n" +
					"In addition to --numbering configured false, you explicitly set --empty and --files false.\n" +
					"What to do:\n" +
					"Either:\n" +
					" 1. Edit the configuration file so that at least one of these flags is true, or\n" +
					" 2. Explicitly set at least one of these flags true on the command line.\n",
			},
		},
		"no work, empty and numbering configured that way": {
			cs: &checkSettings{
				empty:     cmdtoolkit.CommandFlag[bool]{UserSet: true},
				numbering: cmdtoolkit.CommandFlag[bool]{UserSet: true},
			},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No checks will be executed.\n" +
					"Why?\n" +
					"In addition to --files configured false, you explicitly set --empty and --numbering false.\n" +
					"What to do:\n" +
					"Either:\n" +
					" 1. Edit the configuration file so that at least one of these flags is true, or\n" +
					" 2. Explicitly set at least one of these flags true on the command line.\n",
			},
		},
		"no work, numbering and files configured that way": {
			cs: &checkSettings{
				numbering: cmdtoolkit.CommandFlag[bool]{UserSet: true},
				files:     cmdtoolkit.CommandFlag[bool]{UserSet: true},
			},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No checks will be executed.\n" +
					"Why?\n" +
					"In addition to --empty configured false, you explicitly set --files and --numbering false.\n" +
					"What to do:\n" +
					"Either:\n" +
					" 1. Edit the configuration file so that at least one of these flags is true, or\n" +
					" 2. Explicitly set at least one of these flags true on the command line.\n",
			},
		},
		"no work, all flags configured that way": {
			cs: &checkSettings{
				numbering: cmdtoolkit.CommandFlag[bool]{UserSet: true},
				files:     cmdtoolkit.CommandFlag[bool]{UserSet: true},
				empty:     cmdtoolkit.CommandFlag[bool]{UserSet: true},
			},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No checks will be executed.\n" +
					"Why?\n" +
					"You explicitly set --empty, --files, and --numbering false.\n" +
					"What to do:\n" +
					"Either:\n" +
					" 1. Edit the configuration file so that at least one of these flags is true, or\n" +
					" 2. Explicitly set at least one of these flags true on the command line.\n",
			},
		},
		"check empty": {
			cs:   &checkSettings{empty: cmdtoolkit.CommandFlag[bool]{Value: true}},
			want: true,
		},
		"check files": {
			cs:   &checkSettings{files: cmdtoolkit.CommandFlag[bool]{Value: true}},
			want: true,
		},
		"check numbering": {
			cs:   &checkSettings{numbering: cmdtoolkit.CommandFlag[bool]{Value: true}},
			want: true,
		},
		"check empty and files": {
			cs: &checkSettings{
				empty: cmdtoolkit.CommandFlag[bool]{Value: true},
				files: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			want: true,
		},
		"check empty and numbering": {
			cs: &checkSettings{
				empty:     cmdtoolkit.CommandFlag[bool]{Value: true},
				numbering: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			want: true,
		},
		"check numbering and files": {
			cs: &checkSettings{
				numbering: cmdtoolkit.CommandFlag[bool]{Value: true},
				files:     cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			want: true,
		},
		"check everything": {
			cs: &checkSettings{
				empty:     cmdtoolkit.CommandFlag[bool]{Value: true},
				files:     cmdtoolkit.CommandFlag[bool]{Value: true},
				numbering: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if got := tt.cs.hasWorkToDo(o); got != tt.want {
				t.Errorf("checkSettings.hasWorkToDo() = %v, want %v", got, tt.want)
			}
			o.Report(t, "checkSettings.hasWorkToDo()", tt.WantedRecording)
		})
	}
}

func Test_checkSettings_performEmptyAnalysis(t *testing.T) {
	tests := map[string]struct {
		cs             *checkSettings
		checkedArtists []*concernedArtist
		want           bool
	}{
		"do nothing": {cs: &checkSettings{empty: cmdtoolkit.CommandFlag[bool]{Value: false}}},
		"empty slice": {
			cs:             &checkSettings{empty: cmdtoolkit.CommandFlag[bool]{Value: true}},
			checkedArtists: nil,
		},
		"full slice, no problems": {
			cs:             &checkSettings{empty: cmdtoolkit.CommandFlag[bool]{Value: true}},
			checkedArtists: createConcernedArtists(generateArtists(5, 6, 7, nil)),
		},
		"empty artists": {
			cs:             &checkSettings{empty: cmdtoolkit.CommandFlag[bool]{Value: true}},
			checkedArtists: createConcernedArtists(generateArtists(1, 0, 10, nil)),
			want:           true,
		},
		"empty albums": {
			cs:             &checkSettings{empty: cmdtoolkit.CommandFlag[bool]{Value: true}},
			checkedArtists: createConcernedArtists(generateArtists(4, 6, 0, nil)),
			want:           true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.cs.performEmptyAnalysis(tt.checkedArtists); got != tt.want {
				t.Errorf("checkSettings.performEmptyAnalysis() = %v, want %v", got, tt.want)
			}
			verifiedFound := false
			for _, artist := range tt.checkedArtists {
				if artist.isConcerned() {
					verifiedFound = true
				}
			}
			if verifiedFound != tt.want {
				t.Errorf("checkSettings.performEmptyAnalysis() verified = %v, want %v",
					verifiedFound, tt.want)
			}
		})
	}
}

func Test_numberGap_generateMissingTrackNumbers(t *testing.T) {
	tests := map[string]struct {
		gap  numberGap
		want string
	}{
		"equal":    {gap: numberGap{value1: 2, value2: 2}, want: "2"},
		"unequal":  {gap: numberGap{value1: 2, value2: 3}, want: "2-3"},
		"unequal2": {gap: numberGap{value1: 3, value2: 2}, want: "2-3"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.gap.generateMissingTrackNumbers(); got != tt.want {
				t.Errorf("numberGap.generateMissingTrackNumbers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_generateNumberingConcerns(t *testing.T) {
	type args struct {
		m        map[int][]string
		maxTrack int
	}
	tests := map[string]struct {
		args
		want []string
	}{
		"empty": {
			args: args{m: nil, maxTrack: 0},
			want: []string{},
		},
		"clean": {
			args: args{
				m: map[int][]string{
					1: {"track 1"},
					2: {"track 2"},
					3: {"track 3"},
					4: {"track 4"},
					5: {"track 5"},
				},
				maxTrack: 5,
			},
			want: []string{},
		},
		"problematic": {
			args: args{
				m: map[int][]string{
					3:  {"track 3"},
					5:  {"track 4", "track 5", "some other track"},
					8:  {"track 8"},
					9:  {},
					10: {"track 10"},
					19: {"track 19"},
				},
				maxTrack: 20,
			},
			want: []string{
				"multiple tracks identified as track 5: \"some other track\", \"track 4\"" +
					" and \"track 5\"",
				"missing tracks identified: 1-2, 4, 6-7, 9, 11-18, 20",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := generateNumberingConcerns(tt.args.m,
				tt.args.maxTrack); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateNumberingConcerns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkSettings_performNumberingAnalysis(t *testing.T) {
	var defectiveArtists []*files.Artist
	for r := 0; r < 4; r++ {
		artistName := fmt.Sprintf("my artist %d", r)
		artist := files.NewArtist(artistName, filepath.Join("Music", artistName))
		for k := 0; k < 5; k++ {
			albumName := fmt.Sprintf("my album %d%d", r, k)
			album := files.AlbumMaker{
				Title:     albumName,
				Artist:    artist,
				Directory: filepath.Join("Music", "my artist", albumName),
			}.NewAlbum(true)
			for j := 1; j <= 6; j += 2 {
				trackName := fmt.Sprintf("my track %d%d%d", r, k, j)
				files.TrackMaker{
					Album:      album,
					FileName:   fmt.Sprintf("%d %s.mp3", j, trackName),
					SimpleName: trackName,
					Number:     j,
				}.NewTrack(true)
			}
		}
		defectiveArtists = append(defectiveArtists, artist)
	}

	tests := map[string]struct {
		cs             *checkSettings
		checkedArtists []*concernedArtist
		want           bool
	}{
		"no analysis": {
			cs:             &checkSettings{numbering: cmdtoolkit.CommandFlag[bool]{Value: false}},
			checkedArtists: createConcernedArtists(generateArtists(5, 6, 7, nil)),
			want:           false,
		},
		"ok analysis": {
			cs:             &checkSettings{numbering: cmdtoolkit.CommandFlag[bool]{Value: true}},
			checkedArtists: createConcernedArtists(generateArtists(5, 6, 7, nil)),
			want:           false,
		},
		"missing numbers found": {
			cs:             &checkSettings{numbering: cmdtoolkit.CommandFlag[bool]{Value: true}},
			checkedArtists: createConcernedArtists(defectiveArtists),
			want:           true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.cs.performNumberingAnalysis(tt.checkedArtists); got != tt.want {
				t.Errorf("checkSettings.performNumberingAnalysis() = %v, want %v", got,
					tt.want)
			}
			verifiedFound := false
			for _, artist := range tt.checkedArtists {
				if artist.isConcerned() {
					verifiedFound = true
				}
			}
			if verifiedFound != tt.want {
				t.Errorf("checkSettings.performNumberingAnalysis() verified = %v, want %v",
					verifiedFound, tt.want)
			}
		})
	}
}

func Test_recordTrackFileConcerns(t *testing.T) {
	originalArtists := generateArtists(5, 6, 7, nil)
	tracks := make([]*files.Track, 0)
	for _, artist := range originalArtists {
		copiedArtist := artist.Copy()
		for _, album := range artist.Albums() {
			copiedAlbum := album.Copy(copiedArtist, true, true)
			tracks = append(tracks, copiedAlbum.Tracks()...)
		}
	}
	type args struct {
		checkedArtists []*concernedArtist
		track          *files.Track
		concerns       []string
	}
	tests := map[string]struct {
		args
		wantFoundConcerns bool
	}{
		"no concerns": {
			args:              args{checkedArtists: nil, track: nil, concerns: nil},
			wantFoundConcerns: false,
		},
		"concerns": {
			args: args{
				checkedArtists: createConcernedArtists(originalArtists),
				track:          tracks[len(tracks)-1],
				concerns:       []string{"mismatched artist", "mismatched album"},
			},
			wantFoundConcerns: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := recordTrackFileConcerns(tt.args.checkedArtists, tt.args.track, tt.args.concerns)
			if got != tt.wantFoundConcerns {
				t.Errorf("recordTrackFileConcerns() = %v, want %v", got, tt.wantFoundConcerns)
			}
			if tt.wantFoundConcerns {
				hasConcerns := false
				for _, cAr := range tt.args.checkedArtists {
					if cAr.isConcerned() {
						hasConcerns = true
					}
				}
				if !hasConcerns {
					t.Errorf("recordTrackFileConcerns() true, but no concerns actually recorded")
				}
			}
		})
	}
}

func Test_checkSettings_performFileAnalysis(t *testing.T) {
	originalReadMetadata := readMetadata
	defer func() {
		readMetadata = originalReadMetadata
	}()
	readMetadata = func(_ output.Bus, _ []*files.Artist) {}
	type args struct {
		checkedArtists []*concernedArtist
		ss             *searchSettings
	}
	tests := map[string]struct {
		cs *checkSettings
		args
		want bool
		output.WantedRecording
	}{
		"not permitted to do anything": {
			cs:              &checkSettings{files: cmdtoolkit.CommandFlag[bool]{Value: false}},
			args:            args{},
			want:            false,
			WantedRecording: output.WantedRecording{},
		},
		"allowed, but nothing to check": {
			cs: &checkSettings{files: cmdtoolkit.CommandFlag[bool]{Value: true}},
			args: args{
				checkedArtists: []*concernedArtist{},
				ss:             &searchSettings{},
			},
			want:            false,
			WantedRecording: output.WantedRecording{},
		},
		"work to do": {
			cs: &checkSettings{files: cmdtoolkit.CommandFlag[bool]{Value: true}},
			args: args{
				checkedArtists: createConcernedArtists(generateArtists(4, 5, 6, nil)),
				ss: &searchSettings{
					artistFilter: regexp.MustCompile(".*"),
					albumFilter:  regexp.MustCompile(".*"),
					trackFilter:  regexp.MustCompile(".*"),
				},
			},
			want:            true,
			WantedRecording: output.WantedRecording{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got := tt.cs.performFileAnalysis(o, tt.args.checkedArtists, tt.args.ss)
			if got != tt.want {
				t.Errorf("checkSettings.performFileAnalysis() = %v, want %v", got, tt.want)
			}
			o.Report(t, "checkSettings.performFileAnalysis()", tt.WantedRecording)
		})
	}
}

func Test_checkSettings_maybeReportCleanResults(t *testing.T) {
	tests := map[string]struct {
		cs       *checkSettings
		requests checkReportRequests
		output.WantedRecording
	}{
		"no concerns found because nothing was checked": {
			cs:              &checkSettings{},
			requests:        checkReportRequests{},
			WantedRecording: output.WantedRecording{},
		},
		"all concerns found, everything was checked": {
			cs: &checkSettings{
				empty:     cmdtoolkit.CommandFlag[bool]{Value: true},
				numbering: cmdtoolkit.CommandFlag[bool]{Value: true},
				files:     cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			requests: checkReportRequests{
				reportEmptyCheckResults:     true,
				reportFilesCheckResults:     true,
				reportNumberingCheckResults: true,
			},
			WantedRecording: output.WantedRecording{},
		},
		"no concerns found, everything was checked": {
			cs: &checkSettings{
				empty:     cmdtoolkit.CommandFlag[bool]{Value: true},
				numbering: cmdtoolkit.CommandFlag[bool]{Value: true},
				files:     cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			requests: checkReportRequests{
				reportEmptyCheckResults:     false,
				reportFilesCheckResults:     false,
				reportNumberingCheckResults: false,
			},
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Empty Folder Analysis: no empty folders found.\n" +
					"Numbering Analysis: no missing or duplicate tracks found.\n" +
					"File Analysis: no inconsistencies found.\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.cs.maybeReportCleanResults(o, tt.requests)
			o.Report(t, "checkSettings.maybeReportCleanResults()", tt.WantedRecording)
		})
	}
}

func Test_checkSettings_performChecks(t *testing.T) {
	originalReadMetadata := readMetadata
	defer func() {
		readMetadata = originalReadMetadata
	}()
	readMetadata = func(_ output.Bus, _ []*files.Artist) {}
	type args struct {
		artists []*files.Artist
		ss      *searchSettings
	}
	tests := map[string]struct {
		cs *checkSettings
		args
		wantStatus *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"no artists": {
			cs:              nil,
			args:            args{artists: nil, ss: nil},
			wantStatus:      cmdtoolkit.NewExitUserError("check"),
			WantedRecording: output.WantedRecording{},
		},
		"artists to check, check everything": {
			cs: &checkSettings{
				empty:     cmdtoolkit.CommandFlag[bool]{Value: true},
				numbering: cmdtoolkit.CommandFlag[bool]{Value: true},
				files:     cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			args: args{
				artists: generateArtists(1, 2, 3, nil),
				ss: &searchSettings{
					artistFilter: regexp.MustCompile(".*"),
					albumFilter:  regexp.MustCompile(".*"),
					trackFilter:  regexp.MustCompile(".*"),
				},
			},
			wantStatus: nil,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Artist \"my artist 0\"\n" +
					"* [files] for all albums: for all tracks: " +
					"differences cannot be determined: metadata has not been read\n" +
					"Empty Folder Analysis: no empty folders found.\n" +
					"Numbering Analysis: no missing or duplicate tracks found.\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got := tt.cs.performChecks(o, tt.args.artists, tt.args.ss)
			if !compareExitErrors(got, tt.wantStatus) {
				t.Errorf("checkSettings.performChecks() got %s want %s", got, tt.wantStatus)
			}
			o.Report(t, "checkSettings.performChecks()", tt.WantedRecording)
		})
	}
}

func Test_checkSettings_maybeDoWork(t *testing.T) {
	tests := map[string]struct {
		cs         *checkSettings
		ss         *searchSettings
		wantStatus *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"nothing to do": {
			cs:         &checkSettings{},
			ss:         nil,
			wantStatus: cmdtoolkit.NewExitUserError("check"),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No checks will be executed.\n" +
					"Why?\n" +
					"The flags --empty, --files, and --numbering are all configured false.\n" +
					"What to do:\n" +
					"Either:\n" +
					" 1. Edit the configuration file so that at least one of these flags is true, or\n" +
					" 2. Explicitly set at least one of these flags true on the command line.\n",
			},
		},
		"try a little work": {
			cs: &checkSettings{empty: cmdtoolkit.CommandFlag[bool]{Value: true}},
			ss: &searchSettings{
				artistFilter:   regexp.MustCompile(".*"),
				albumFilter:    regexp.MustCompile(".*"),
				trackFilter:    regexp.MustCompile(".*"),
				fileExtensions: []string{".mp3"},
				topDirectory:   filepath.Join(".", "no dir"),
			},
			wantStatus: cmdtoolkit.NewExitUserError("check"),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The directory \"no dir\" cannot be read: '*fs.PathError: open no dir: The system" +
					" cannot find the file specified.'.\n" +
					"No mp3 files could be found using the specified parameters.\n" +
					"Why?\n" +
					"There were no directories found in \"no dir\" (the --topDir value).\n" +
					"What to do:\n" +
					"Set --topDir to the path of a directory that contains artist" +
					" directories.\n",
				Log: "" +
					"level='error'" +
					" directory='no dir'" +
					" error='open no dir: The system cannot find the file specified.'" +
					" msg='cannot read directory'\n" +
					"level='error'" +
					" --topDir='no dir'" +
					" msg='cannot find any artist directories'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if got := tt.cs.maybeDoWork(o, tt.ss); !compareExitErrors(got, tt.wantStatus) {
				t.Errorf("checkSettings.maybeDoWork() got %s want %s", got, tt.wantStatus)
			}
			o.Report(t, "checkSettings.maybeDoWork()", tt.WantedRecording)
		})
	}
}

func Test_checkRun(t *testing.T) {
	initGlobals()
	originalBus := bus
	originalSearchFlags := searchFlags
	defer func() {
		bus = originalBus
		searchFlags = originalSearchFlags
	}()
	searchFlags = safeSearchFlags
	checkFlags := &cmdtoolkit.FlagSet{
		Name: checkCommand,
		Details: map[string]*cmdtoolkit.FlagDetails{
			checkEmpty: {
				AbbreviatedName: checkEmptyAbbr,
				Usage:           "report empty album and artist directories",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
			checkFiles: {
				AbbreviatedName: checkFilesAbbr,
				Usage:           "report metadata/file inconsistencies",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
			checkNumbering: {
				AbbreviatedName: checkNumberingAbbr,
				Usage:           "report missing track numbers and duplicated track numbering",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
		},
	}
	command := &cobra.Command{}
	cmdtoolkit.AddFlags(output.NewNilBus(), cmdtoolkit.EmptyConfiguration(), command.Flags(),
		checkFlags, searchFlags)
	type args struct {
		cmd *cobra.Command
		in1 []string
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"default case": {
			args: args{cmd: command},
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No checks will be executed.\n" +
					"Why?\n" +
					"The flags --empty, --files, and --numbering are all configured false.\n" +
					"What to do:\n" +
					"Either:\n" +
					" 1. Edit the configuration file so that at least one of these flags is true, or\n" +
					" 2. Explicitly set at least one of these flags true on the command line.\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			bus = o // cook getBus()
			_ = checkRun(tt.args.cmd, tt.args.in1)
			o.Report(t, "checkRun()", tt.WantedRecording)
		})
	}
}

func cloneCommand(original *cobra.Command) *cobra.Command {
	clone := &cobra.Command{
		Use:                   original.Use,
		DisableFlagsInUseLine: original.DisableFlagsInUseLine,
		Short:                 original.Short,
		Long:                  original.Long,
		Example:               original.Example,
		Run:                   original.Run,
		RunE:                  original.RunE,
	}
	return clone
}

func Test_check_Help(t *testing.T) {
	originalSearchFlags := searchFlags
	defer func() {
		searchFlags = originalSearchFlags
	}()
	searchFlags = safeSearchFlags
	commandUnderTest := cloneCommand(checkCmd)
	cmdtoolkit.AddFlags(output.NewNilBus(), cmdtoolkit.EmptyConfiguration(),
		commandUnderTest.Flags(), checkFlags, searchFlags)
	tests := map[string]struct {
		output.WantedRecording
	}{
		"good": {
			WantedRecording: output.WantedRecording{
				Console: "" +
					"\"check\" inspects mp3 files and their containing directories and " +
					"reports any problems detected\n" +
					"\n" +
					"Usage:\n" +
					"  check [--empty] [--files] [--numbering] [--albumFilter regex] [--artistFilter regex] " +
					"[--trackFilter regex] [--topDir dir] [--extensions extensions]\n" +
					"\n" +
					"Examples:\n" +
					"check --empty\n" +
					"  reports empty artist and album directories\n" +
					"check --files\n" +
					"  reads each mp3 file's metadata and reports any inconsistencies found\n" +
					"check --numbering\n" +
					"  reports errors in the track numbers of mp3 files\n" +
					"\n" +
					"Flags:\n" +
					"      --albumFilter string    regular expression specifying which albums to " +
					"select (default \".*\")\n" +
					"      --artistFilter string   regular expression specifying which " +
					"artists to select (default \".*\")\n" +
					"  -e, --empty                 report empty album and artist directories (default false)\n" +
					"      --extensions string     comma-delimited list of file " +
					"extensions used by mp3 files (default \".mp3\")\n" +
					"  -f, --files                 report metadata/file inconsistencies (default false)\n" +
					"  -n, --numbering             report missing track " +
					"numbers and duplicated track numbering (default false)\n" +
					"      --topDir string         top directory specifying where to find mp3 files (default \".\")\n" +
					"      --trackFilter string    regular expression " +
					"specifying which tracks to select (default \".*\")\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			command := commandUnderTest
			enableCommandRecording(o, command)
			_ = command.Help()
			o.Report(t, "check Help()", tt.WantedRecording)
		})
	}
}

func Test_check_Usage(t *testing.T) {
	originalSearchFlags := searchFlags
	defer func() {
		searchFlags = originalSearchFlags
	}()
	searchFlags = safeSearchFlags
	commandUnderTest := cloneCommand(checkCmd)
	cmdtoolkit.AddFlags(output.NewNilBus(), cmdtoolkit.EmptyConfiguration(),
		commandUnderTest.Flags(), checkFlags, searchFlags)
	tests := map[string]struct {
		output.WantedRecording
	}{
		"good": {
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Usage:\n" +
					"  check [--empty] [--files] [--numbering] [--albumFilter regex] [--artistFilter regex] " +
					"[--trackFilter regex] [--topDir dir] [--extensions extensions]\n" +
					"\n" +
					"Examples:\n" +
					"check --empty\n" +
					"  reports empty artist and album directories\n" +
					"check --files\n" +
					"  reads each mp3 file's metadata and reports any inconsistencies" +
					" found\n" +
					"check --numbering\n" +
					"  reports errors in the track numbers of mp3 files\n" +
					"\n" +
					"Flags:\n" +
					"      --albumFilter string    " +
					"regular expression specifying which albums to select (default \".*\")\n" +
					"      --artistFilter string   " +
					"regular expression specifying which artists to select (default \".*\")\n" +
					"  -e, --empty                 " +
					"report empty album and artist directories (default false)\n" +
					"      --extensions string     " +
					"comma-delimited list of file extensions used by mp3 files (default \".mp3\")\n" +
					"  -f, --files                 " +
					"report metadata/file inconsistencies (default false)\n" +
					"  -n, --numbering             " +
					"report missing track numbers and duplicated track numbering (default false)\n" +
					"      --topDir string         " +
					"top directory specifying where to find mp3 files (default \".\")\n" +
					"      --trackFilter string    " +
					"regular expression specifying which tracks to select (default \".*\")\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			command := commandUnderTest
			enableCommandRecording(o, command)
			_ = command.Usage()
			o.Report(t, "check Usage()", tt.WantedRecording)
		})
	}
}
