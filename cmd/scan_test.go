/*
Copyright © 2026 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd

import (
	"fmt"
	"mp3repair/internal/files"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"

	"github.com/adrg/xdg"
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

func Test_processScanFlags(t *testing.T) {
	tests := map[string]struct {
		values map[string]*cmdtoolkit.CommandFlag[any]
		want   *scanSettings
		want1  bool
		output.WantedRecording
	}{
		"no data": {
			values: map[string]*cmdtoolkit.CommandFlag[any]{},
			want:   &scanSettings{},
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
			want:  &scanSettings{},
			want1: true,
		},
		"overridden": {
			values: map[string]*cmdtoolkit.CommandFlag[any]{
				"empty":     {Value: true, UserSet: true},
				"files":     {Value: true, UserSet: true},
				"numbering": {Value: true, UserSet: true},
			},
			want: &scanSettings{
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
			got, got1 := processScanFlags(o, tt.values)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processScanFlags() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("processScanFlags() got1 = %v, want %v", got1, tt.want1)
			}
			o.Report(t, "processScanFlags()", tt.WantedRecording)
		})
	}
}

func Test_scanSettings_hasWorkToDo(t *testing.T) {
	tests := map[string]struct {
		scanSet *scanSettings
		want    bool
		output.WantedRecording
	}{
		"no work, as configured": {
			scanSet: &scanSettings{},
			want:    false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No scans will be performed.\n" +
					"Why?\n" +
					"The flags --empty, --files, and --numbering are all configured false.\n" +
					"What to do:\n" +
					"Either:\n" +
					" 1. Edit the configuration file so that at least one of these flags is true, or\n" +
					" 2. Explicitly set at least one of these flags true on the command line.\n",
			},
		},
		"no work, empty configured that way": {
			scanSet: &scanSettings{empty: cmdtoolkit.CommandFlag[bool]{UserSet: true}},
			want:    false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No scans will be performed.\n" +
					"Why?\n" +
					"In addition to --files and --numbering configured false, you explicitly set --empty false.\n" +
					"What to do:\n" +
					"Either:\n" +
					" 1. Edit the configuration file so that at least one of these flags is true, or\n" +
					" 2. Explicitly set at least one of these flags true on the command line.\n",
			},
		},
		"no work, files configured that way": {
			scanSet: &scanSettings{files: cmdtoolkit.CommandFlag[bool]{UserSet: true}},
			want:    false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No scans will be performed.\n" +
					"Why?\n" +
					"In addition to --empty and --numbering configured false, you explicitly set --files false.\n" +
					"What to do:\n" +
					"Either:\n" +
					" 1. Edit the configuration file so that at least one of these flags is true, or\n" +
					" 2. Explicitly set at least one of these flags true on the command line.\n",
			},
		},
		"no work, numbering configured that way": {
			scanSet: &scanSettings{numbering: cmdtoolkit.CommandFlag[bool]{UserSet: true}},
			want:    false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No scans will be performed.\n" +
					"Why?\n" +
					"In addition to --empty and --files configured false, you explicitly set --numbering false.\n" +
					"What to do:\n" +
					"Either:\n" +
					" 1. Edit the configuration file so that at least one of these flags is true, or\n" +
					" 2. Explicitly set at least one of these flags true on the command line.\n",
			},
		},
		"no work, empty and files configured that way": {
			scanSet: &scanSettings{
				empty: cmdtoolkit.CommandFlag[bool]{UserSet: true},
				files: cmdtoolkit.CommandFlag[bool]{UserSet: true},
			},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No scans will be performed.\n" +
					"Why?\n" +
					"In addition to --numbering configured false, you explicitly set --empty and --files false.\n" +
					"What to do:\n" +
					"Either:\n" +
					" 1. Edit the configuration file so that at least one of these flags is true, or\n" +
					" 2. Explicitly set at least one of these flags true on the command line.\n",
			},
		},
		"no work, empty and numbering configured that way": {
			scanSet: &scanSettings{
				empty:     cmdtoolkit.CommandFlag[bool]{UserSet: true},
				numbering: cmdtoolkit.CommandFlag[bool]{UserSet: true},
			},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No scans will be performed.\n" +
					"Why?\n" +
					"In addition to --files configured false, you explicitly set --empty and --numbering false.\n" +
					"What to do:\n" +
					"Either:\n" +
					" 1. Edit the configuration file so that at least one of these flags is true, or\n" +
					" 2. Explicitly set at least one of these flags true on the command line.\n",
			},
		},
		"no work, numbering and files configured that way": {
			scanSet: &scanSettings{
				numbering: cmdtoolkit.CommandFlag[bool]{UserSet: true},
				files:     cmdtoolkit.CommandFlag[bool]{UserSet: true},
			},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No scans will be performed.\n" +
					"Why?\n" +
					"In addition to --empty configured false, you explicitly set --files and --numbering false.\n" +
					"What to do:\n" +
					"Either:\n" +
					" 1. Edit the configuration file so that at least one of these flags is true, or\n" +
					" 2. Explicitly set at least one of these flags true on the command line.\n",
			},
		},
		"no work, all flags configured that way": {
			scanSet: &scanSettings{
				numbering: cmdtoolkit.CommandFlag[bool]{UserSet: true},
				files:     cmdtoolkit.CommandFlag[bool]{UserSet: true},
				empty:     cmdtoolkit.CommandFlag[bool]{UserSet: true},
			},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No scans will be performed.\n" +
					"Why?\n" +
					"You explicitly set --empty, --files, and --numbering false.\n" +
					"What to do:\n" +
					"Either:\n" +
					" 1. Edit the configuration file so that at least one of these flags is true, or\n" +
					" 2. Explicitly set at least one of these flags true on the command line.\n",
			},
		},
		"scan empty": {
			scanSet: &scanSettings{empty: cmdtoolkit.CommandFlag[bool]{Value: true}},
			want:    true,
		},
		"scan files": {
			scanSet: &scanSettings{files: cmdtoolkit.CommandFlag[bool]{Value: true}},
			want:    true,
		},
		"scan numbering": {
			scanSet: &scanSettings{numbering: cmdtoolkit.CommandFlag[bool]{Value: true}},
			want:    true,
		},
		"scan empty and files": {
			scanSet: &scanSettings{
				empty: cmdtoolkit.CommandFlag[bool]{Value: true},
				files: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			want: true,
		},
		"scan empty and numbering": {
			scanSet: &scanSettings{
				empty:     cmdtoolkit.CommandFlag[bool]{Value: true},
				numbering: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			want: true,
		},
		"scan numbering and files": {
			scanSet: &scanSettings{
				numbering: cmdtoolkit.CommandFlag[bool]{Value: true},
				files:     cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			want: true,
		},
		"scan everything": {
			scanSet: &scanSettings{
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
			if got := tt.scanSet.hasWorkToDo(o); got != tt.want {
				t.Errorf("scanSettings.hasWorkToDo() = %v, want %v", got, tt.want)
			}
			o.Report(t, "scanSettings.hasWorkToDo()", tt.WantedRecording)
		})
	}
}

func Test_scanSettings_performEmptyAnalysis(t *testing.T) {
	tests := map[string]struct {
		scanSet        *scanSettings
		scannedArtists []*concernedArtist
		want           bool
	}{
		"do nothing": {scanSet: &scanSettings{empty: cmdtoolkit.CommandFlag[bool]{Value: false}}},
		"empty slice": {
			scanSet:        &scanSettings{empty: cmdtoolkit.CommandFlag[bool]{Value: true}},
			scannedArtists: nil,
		},
		"full slice, no problems": {
			scanSet:        &scanSettings{empty: cmdtoolkit.CommandFlag[bool]{Value: true}},
			scannedArtists: createConcernedArtists(generateArtists(5, 6, 7, nil)),
		},
		"empty artists": {
			scanSet:        &scanSettings{empty: cmdtoolkit.CommandFlag[bool]{Value: true}},
			scannedArtists: createConcernedArtists(generateArtists(1, 0, 10, nil)),
			want:           true,
		},
		"empty albums": {
			scanSet:        &scanSettings{empty: cmdtoolkit.CommandFlag[bool]{Value: true}},
			scannedArtists: createConcernedArtists(generateArtists(4, 6, 0, nil)),
			want:           true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.scanSet.performEmptyAnalysis(tt.scannedArtists); got != tt.want {
				t.Errorf("scanSettings.performEmptyAnalysis() = %v, want %v", got, tt.want)
			}
			verifiedFound := false
			for _, artist := range tt.scannedArtists {
				if artist.isConcerned() {
					verifiedFound = true
				}
			}
			if verifiedFound != tt.want {
				t.Errorf("scanSettings.performEmptyAnalysis() verified = %v, want %v",
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

func Test_scankSettings_performNumberingAnalysis(t *testing.T) {
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
		scanSet        *scanSettings
		scannedArtists []*concernedArtist
		want           bool
	}{
		"no analysis": {
			scanSet:        &scanSettings{numbering: cmdtoolkit.CommandFlag[bool]{Value: false}},
			scannedArtists: createConcernedArtists(generateArtists(5, 6, 7, nil)),
			want:           false,
		},
		"ok analysis": {
			scanSet:        &scanSettings{numbering: cmdtoolkit.CommandFlag[bool]{Value: true}},
			scannedArtists: createConcernedArtists(generateArtists(5, 6, 7, nil)),
			want:           false,
		},
		"missing numbers found": {
			scanSet:        &scanSettings{numbering: cmdtoolkit.CommandFlag[bool]{Value: true}},
			scannedArtists: createConcernedArtists(defectiveArtists),
			want:           true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.scanSet.performNumberingAnalysis(tt.scannedArtists); got != tt.want {
				t.Errorf("scanSettings.performNumberingAnalysis() = %v, want %v", got,
					tt.want)
			}
			verifiedFound := false
			for _, artist := range tt.scannedArtists {
				if artist.isConcerned() {
					verifiedFound = true
				}
			}
			if verifiedFound != tt.want {
				t.Errorf("scanSettings.performNumberingAnalysis() verified = %v, want %v",
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
		scannedArtists []*concernedArtist
		track          *files.Track
		concerns       []string
	}
	tests := map[string]struct {
		args
		wantFoundConcerns bool
	}{
		"no concerns": {
			args:              args{scannedArtists: nil, track: nil, concerns: nil},
			wantFoundConcerns: false,
		},
		"concerns": {
			args: args{
				scannedArtists: createConcernedArtists(originalArtists),
				track:          tracks[len(tracks)-1],
				concerns:       []string{"mismatched artist", "mismatched album"},
			},
			wantFoundConcerns: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := recordTrackFileConcerns(tt.args.scannedArtists, tt.args.track, tt.args.concerns)
			if got != tt.wantFoundConcerns {
				t.Errorf("recordTrackFileConcerns() = %v, want %v", got, tt.wantFoundConcerns)
			}
			if tt.wantFoundConcerns {
				hasConcerns := false
				for _, cAr := range tt.args.scannedArtists {
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

func Test_scanSettings_performFileAnalysis(t *testing.T) {
	originalReadMetadata := readMetadata
	defer func() {
		readMetadata = originalReadMetadata
	}()
	readMetadata = func(_ output.Bus, _ []*files.Artist, _ int) {}
	type args struct {
		scannedArtists []*concernedArtist
		ss             *searchSettings
		ios            *ioSettings
	}
	tests := map[string]struct {
		scanSet *scanSettings
		args
		want bool
		output.WantedRecording
	}{
		"not permitted to do anything": {
			scanSet:         &scanSettings{files: cmdtoolkit.CommandFlag[bool]{Value: false}},
			args:            args{},
			want:            false,
			WantedRecording: output.WantedRecording{},
		},
		"allowed, but nothing to scan": {
			scanSet: &scanSettings{files: cmdtoolkit.CommandFlag[bool]{Value: true}},
			args: args{
				scannedArtists: []*concernedArtist{},
				ss:             &searchSettings{},
				ios:            &ioSettings{},
			},
			want:            false,
			WantedRecording: output.WantedRecording{},
		},
		"work to do": {
			scanSet: &scanSettings{files: cmdtoolkit.CommandFlag[bool]{Value: true}},
			args: args{
				scannedArtists: createConcernedArtists(generateArtists(4, 5, 6, nil)),
				ss: &searchSettings{
					artistFilter: regexp.MustCompile(".*"),
					albumFilter:  regexp.MustCompile(".*"),
					trackFilter:  regexp.MustCompile(".*"),
				},
				ios: &ioSettings{openFileLimit: 100},
			},
			want:            true,
			WantedRecording: output.WantedRecording{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got := tt.scanSet.performFileAnalysis(o, tt.args.scannedArtists, tt.args.ss, tt.args.ios)
			if got != tt.want {
				t.Errorf("scanSettings.performFileAnalysis() = %v, want %v", got, tt.want)
			}
			o.Report(t, "scanSettings.performFileAnalysis()", tt.WantedRecording)
		})
	}
}

func Test_scanSettings_maybeReportCleanResults(t *testing.T) {
	tests := map[string]struct {
		scanSet  *scanSettings
		requests scanReportRequests
		output.WantedRecording
	}{
		"no concerns found because nothing was scanned": {
			scanSet:         &scanSettings{},
			requests:        scanReportRequests{},
			WantedRecording: output.WantedRecording{},
		},
		"all concerns found, everything was scanned": {
			scanSet: &scanSettings{
				empty:     cmdtoolkit.CommandFlag[bool]{Value: true},
				numbering: cmdtoolkit.CommandFlag[bool]{Value: true},
				files:     cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			requests: scanReportRequests{
				reportEmptyScanResults:     true,
				reportFilesScanResults:     true,
				reportNumberingScanResults: true,
			},
			WantedRecording: output.WantedRecording{},
		},
		"no concerns found, everything was scanned": {
			scanSet: &scanSettings{
				empty:     cmdtoolkit.CommandFlag[bool]{Value: true},
				numbering: cmdtoolkit.CommandFlag[bool]{Value: true},
				files:     cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			requests: scanReportRequests{
				reportEmptyScanResults:     false,
				reportFilesScanResults:     false,
				reportNumberingScanResults: false,
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
			tt.scanSet.maybeReportCleanResults(o, tt.requests)
			o.Report(t, "scanSettings.maybeReportCleanResults()", tt.WantedRecording)
		})
	}
}

func Test_scanSettings_performScans(t *testing.T) {
	originalReadMetadata := readMetadata
	defer func() {
		readMetadata = originalReadMetadata
	}()
	readMetadata = func(_ output.Bus, _ []*files.Artist, _ int) {}
	type args struct {
		artists []*files.Artist
		ss      *searchSettings
		ios     *ioSettings
	}
	tests := map[string]struct {
		scanSet *scanSettings
		args
		wantStatus *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"no artists": {
			scanSet:         nil,
			args:            args{artists: nil, ss: nil, ios: nil},
			wantStatus:      cmdtoolkit.NewExitUserError("scan"),
			WantedRecording: output.WantedRecording{},
		},
		"artists to scan, scan everything": {
			scanSet: &scanSettings{
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
				ios: &ioSettings{openFileLimit: 100},
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
			got := tt.scanSet.performScans(o, tt.args.artists, tt.args.ss, tt.args.ios)
			if !compareExitErrors(got, tt.wantStatus) {
				t.Errorf("scanSettings.performScans() got %s want %s", got, tt.wantStatus)
			}
			o.Report(t, "scanSettings.performScans()", tt.WantedRecording)
		})
	}
}

func Test_scanSettings_maybeDoWork(t *testing.T) {
	tests := map[string]struct {
		scanSet    *scanSettings
		ss         *searchSettings
		ios        *ioSettings
		wantStatus *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"nothing to do": {
			scanSet:    &scanSettings{},
			ss:         nil,
			ios:        nil,
			wantStatus: cmdtoolkit.NewExitUserError("scan"),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No scans will be performed.\n" +
					"Why?\n" +
					"The flags --empty, --files, and --numbering are all configured false.\n" +
					"What to do:\n" +
					"Either:\n" +
					" 1. Edit the configuration file so that at least one of these flags is true, or\n" +
					" 2. Explicitly set at least one of these flags true on the command line.\n",
			},
		},
		"try a little work": {
			scanSet: &scanSettings{empty: cmdtoolkit.CommandFlag[bool]{Value: true}},
			ss: &searchSettings{
				artistFilter:   regexp.MustCompile(".*"),
				albumFilter:    regexp.MustCompile(".*"),
				trackFilter:    regexp.MustCompile(".*"),
				fileExtensions: []string{".mp3"},
				musicDir:       filepath.Join(".", "no dir"),
			},
			ios:        &ioSettings{openFileLimit: 100},
			wantStatus: cmdtoolkit.NewExitUserError("scan"),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The directory \"no dir\" cannot be read: '*fs.PathError: open no dir: The system" +
					" cannot find the file specified.'.\n" +
					"No mp3 files could be found using the specified parameters.\n" +
					"Why?\n" +
					"There were no directories found in \"no dir\".\n" +
					"What to do:\n" +
					"Set XDG_MUSIC_DIR to the path of a directory that contains artist directories.\n",
				Log: "" +
					"level='error'" +
					" directory='no dir'" +
					" error='open no dir: The system cannot find the file specified.'" +
					" msg='cannot read directory'\n" +
					"level='error'" +
					" $XDG_MUSIC_DIR='no dir'" +
					" msg='cannot find any artist directories'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if got := tt.scanSet.maybeDoWork(o, tt.ss, tt.ios); !compareExitErrors(got, tt.wantStatus) {
				t.Errorf("scanSettings.maybeDoWork() got %s want %s", got, tt.wantStatus)
			}
			o.Report(t, "scanSettings.maybeDoWork()", tt.WantedRecording)
		})
	}
}

func Test_scanRun(t *testing.T) {
	initGlobals()
	originalBus := bus
	originalSearchFlags := searchFlags
	originalMusicDir := xdg.UserDirs.Music
	defer func() {
		bus = originalBus
		searchFlags = originalSearchFlags
		xdg.UserDirs.Music = originalMusicDir
	}()
	searchFlags = safeSearchFlags
	xdg.UserDirs.Music = "."
	scanFlags := &cmdtoolkit.FlagSet{
		Name: scanCommand,
		Details: map[string]*cmdtoolkit.FlagDetails{
			scanEmpty: {
				AbbreviatedName: scanEmptyAbbr,
				Usage:           "report empty album and artist directories",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
			scanFiles: {
				AbbreviatedName: scanFilesAbbr,
				Usage:           "report metadata/file inconsistencies",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
			scanNumbering: {
				AbbreviatedName: scanNumberingAbbr,
				Usage:           "report missing track numbers and duplicated track numbering",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
		},
	}
	command := &cobra.Command{}
	cmdtoolkit.AddFlags(output.NewNilBus(), cmdtoolkit.EmptyConfiguration(), command.Flags(),
		scanFlags, searchFlags, ioFlags)
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
					"No scans will be performed.\n" +
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
			_ = scanRun(tt.args.cmd, tt.args.in1)
			o.Report(t, "scanRun()", tt.WantedRecording)
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

func Test_scan_Help(t *testing.T) {
	originalSearchFlags := searchFlags
	originalMusicDir := xdg.UserDirs.Music
	defer func() {
		searchFlags = originalSearchFlags
		xdg.UserDirs.Music = originalMusicDir
	}()
	searchFlags = safeSearchFlags
	xdg.UserDirs.Music = "."
	commandUnderTest := cloneCommand(scanCmd)
	cmdtoolkit.AddFlags(output.NewNilBus(), cmdtoolkit.EmptyConfiguration(),
		commandUnderTest.Flags(), scanFlags, searchFlags, ioFlags)
	tests := map[string]struct {
		output.WantedRecording
	}{
		"good": {
			WantedRecording: output.WantedRecording{
				Console: "" +
					"\"scan\" inspects mp3 files and their containing directories and " +
					"reports any problems detected\n" +
					"\n" +
					"Usage:\n" +
					"  scan [--empty] [--files] [--numbering] [--albumFilter regex] [--artistFilter regex] " +
					"[--trackFilter regex] [--extensions extensions] [--maxOpenFiles count]\n" +
					"\n" +
					"Examples:\n" +
					"scan --empty\n" +
					"  reports empty artist and album directories\n" +
					"scan --files\n" +
					"  reads each mp3 file's metadata and reports any inconsistencies found\n" +
					"scan --numbering\n" +
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
					"      --maxOpenFiles int      the maximum number of files that can be read simultaneously " +
					"(at least 1, at most 32767, default 1000) (default 1000)\n" +
					"  -n, --numbering             report missing track " +
					"numbers and duplicated track numbering (default false)\n" +
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
			o.Report(t, "scan Help()", tt.WantedRecording)
		})
	}
}

func Test_scan_Usage(t *testing.T) {
	originalSearchFlags := searchFlags
	originalMusicDir := xdg.UserDirs.Music
	defer func() {
		searchFlags = originalSearchFlags
		xdg.UserDirs.Music = originalMusicDir
	}()
	searchFlags = safeSearchFlags
	xdg.UserDirs.Music = "."
	commandUnderTest := cloneCommand(scanCmd)
	cmdtoolkit.AddFlags(output.NewNilBus(), cmdtoolkit.EmptyConfiguration(),
		commandUnderTest.Flags(), scanFlags, searchFlags, ioFlags)
	tests := map[string]struct {
		output.WantedRecording
	}{
		"good": {
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Usage:\n" +
					"  scan [--empty] [--files] [--numbering] [--albumFilter regex] [--artistFilter regex] " +
					"[--trackFilter regex] [--extensions extensions] [--maxOpenFiles count]\n" +
					"\n" +
					"Examples:\n" +
					"scan --empty\n" +
					"  reports empty artist and album directories\n" +
					"scan --files\n" +
					"  reads each mp3 file's metadata and reports any inconsistencies" +
					" found\n" +
					"scan --numbering\n" +
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
					"      --maxOpenFiles int      the maximum number of files that can be read simultaneously " +
					"(at least 1, at most 32767, default 1000) (default 1000)\n" +
					"  -n, --numbering             " +
					"report missing track numbers and duplicated track numbering (default false)\n" +
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
			o.Report(t, "scan Usage()", tt.WantedRecording)
		})
	}
}
