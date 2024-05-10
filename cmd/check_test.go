/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd_test

import (
	"fmt"
	"mp3repair/cmd"
	"mp3repair/internal/files"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"

	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

func TestProcessCheckFlags(t *testing.T) {
	tests := map[string]struct {
		values map[string]*cmd.FlagValue
		want   *cmd.CheckSettings
		want1  bool
		output.WantedRecording
	}{
		"no data": {
			values: map[string]*cmd.FlagValue{},
			want:   cmd.NewCheckSettings(),
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
			values: map[string]*cmd.FlagValue{
				"empty":     cmd.NewFlagValue().WithValue(false),
				"files":     cmd.NewFlagValue().WithValue(false),
				"numbering": cmd.NewFlagValue().WithValue(false),
			},
			want:  cmd.NewCheckSettings(),
			want1: true,
		},
		"overridden": {
			values: map[string]*cmd.FlagValue{
				"empty":     cmd.NewFlagValue().WithValue(true).WithExplicitlySet(true),
				"files":     cmd.NewFlagValue().WithValue(true).WithExplicitlySet(true),
				"numbering": cmd.NewFlagValue().WithValue(true).WithExplicitlySet(true),
			},
			want: cmd.NewCheckSettings().WithEmpty(true).WithEmptyUserSet(
				true).WithFiles(true).WithFilesUserSet(true).WithNumbering(
				true).WithNumberingUserSet(true),
			want1: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, got1 := cmd.ProcessCheckFlags(o, tt.values)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessCheckFlags() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ProcessCheckFlags() got1 = %v, want %v", got1, tt.want1)
			}
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("ProcessCheckFlags() %s", difference)
				}
			}
		})
	}
}

func TestCheckSettings_HasWorkToDo(t *testing.T) {
	tests := map[string]struct {
		cs   *cmd.CheckSettings
		want bool
		output.WantedRecording
	}{
		"no work, as configured": {
			cs:   cmd.NewCheckSettings(),
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No checks will be executed.\n" +
					"Why?\n" +
					"The flags --empty, --files, and --numbering are all configured false.\n" +
					"What to do:\n" +
					"Either:\n" +
					"[1] Edit the configuration file so that at least one of these flags" +
					" is true, or\n" +
					"[2] explicitly set at least one of these flags true on the command" +
					" line.\n",
			},
		},
		"no work, empty configured that way": {
			cs:   cmd.NewCheckSettings().WithEmptyUserSet(true),
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No checks will be executed.\n" +
					"Why?\n" +
					"In addition to --files and --numbering configured false, you" +
					" explicitly set --empty false.\n" +
					"What to do:\n" +
					"Either:\n" +
					"[1] Edit the configuration file so that at least one of these flags" +
					" is true, or\n" +
					"[2] explicitly set at least one of these flags true on the command" +
					" line.\n",
			},
		},
		"no work, files configured that way": {
			cs:   cmd.NewCheckSettings().WithFilesUserSet(true),
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No checks will be executed.\n" +
					"Why?\n" +
					"In addition to --empty and --numbering configured false, you" +
					" explicitly set --files false.\n" +
					"What to do:\n" +
					"Either:\n" +
					"[1] Edit the configuration file so that at least one of these flags" +
					" is true, or\n" +
					"[2] explicitly set at least one of these flags true on the command" +
					" line.\n",
			},
		},
		"no work, numbering configured that way": {
			cs:   cmd.NewCheckSettings().WithNumberingUserSet(true),
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No checks will be executed.\n" +
					"Why?\n" +
					"In addition to --empty and --files configured false, you explicitly" +
					" set --numbering false.\n" +
					"What to do:\n" +
					"Either:\n" +
					"[1] Edit the configuration file so that at least one of these flags" +
					" is true, or\n" +
					"[2] explicitly set at least one of these flags true on the command" +
					" line.\n",
			},
		},
		"no work, empty and files configured that way": {
			cs:   cmd.NewCheckSettings().WithEmptyUserSet(true).WithFilesUserSet(true),
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No checks will be executed.\n" +
					"Why?\n" +
					"In addition to --numbering configured false, you explicitly set" +
					" --empty and --files false.\n" +
					"What to do:\n" +
					"Either:\n" +
					"[1] Edit the configuration file so that at least one of these flags" +
					" is true, or\n" +
					"[2] explicitly set at least one of these flags true on the command" +
					" line.\n",
			},
		},
		"no work, empty and numbering configured that way": {
			cs:   cmd.NewCheckSettings().WithEmptyUserSet(true).WithNumberingUserSet(true),
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No checks will be executed.\n" +
					"Why?\n" +
					"In addition to --files configured false, you explicitly set --empty" +
					" and --numbering false.\n" +
					"What to do:\n" +
					"Either:\n" +
					"[1] Edit the configuration file so that at least one of these flags" +
					" is true, or\n" +
					"[2] explicitly set at least one of these flags true on the command" +
					" line.\n",
			},
		},
		"no work, numbering and files configured that way": {
			cs:   cmd.NewCheckSettings().WithNumberingUserSet(true).WithFilesUserSet(true),
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No checks will be executed.\n" +
					"Why?\n" +
					"In addition to --empty configured false, you explicitly set --files" +
					" and --numbering false.\n" +
					"What to do:\n" +
					"Either:\n" +
					"[1] Edit the configuration file so that at least one of these flags" +
					" is true, or\n" +
					"[2] explicitly set at least one of these flags true on the command" +
					" line.\n",
			},
		},
		"no work, all flags configured that way": {
			cs: cmd.NewCheckSettings().WithNumberingUserSet(true).WithFilesUserSet(
				true).WithEmptyUserSet(true),
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No checks will be executed.\n" +
					"Why?\n" +
					"You explicitly set --empty, --files, and --numbering false.\n" +
					"What to do:\n" +
					"Either:\n" +
					"[1] Edit the configuration file so that at least one of these flags" +
					" is true, or\n" +
					"[2] explicitly set at least one of these flags true on the command" +
					" line.\n",
			},
		},
		"check empty": {
			cs:   cmd.NewCheckSettings().WithEmpty(true),
			want: true,
		},
		"check files": {
			cs:   cmd.NewCheckSettings().WithFiles(true),
			want: true,
		},
		"check numbering": {
			cs:   cmd.NewCheckSettings().WithNumbering(true),
			want: true,
		},
		"check empty and files": {
			cs:   cmd.NewCheckSettings().WithEmpty(true).WithFiles(true),
			want: true,
		},
		"check empty and numbering": {
			cs:   cmd.NewCheckSettings().WithEmpty(true).WithNumbering(true),
			want: true,
		},
		"check numbering and files": {
			cs:   cmd.NewCheckSettings().WithNumbering(true).WithFiles(true),
			want: true,
		},
		"check everything": {
			cs:   cmd.NewCheckSettings().WithEmpty(true).WithFiles(true).WithNumbering(true),
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if got := tt.cs.HasWorkToDo(o); got != tt.want {
				t.Errorf("CheckSettings.HasWorkToDo() = %v, want %v", got, tt.want)
			}
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("CheckSettings.HasWorkToDo() %s", difference)
				}
			}
		})
	}
}

func TestCheckSettings_PerformEmptyAnalysis(t *testing.T) {
	tests := map[string]struct {
		cs             *cmd.CheckSettings
		checkedArtists []*cmd.ConcernedArtist
		want           bool
	}{
		"do nothing": {cs: cmd.NewCheckSettings().WithEmpty(false)},
		"empty slice": {
			cs:             cmd.NewCheckSettings().WithEmpty(true),
			checkedArtists: nil,
		},
		"full slice, no problems": {
			cs:             cmd.NewCheckSettings().WithEmpty(true),
			checkedArtists: cmd.PrepareConcernedArtists(generateArtists(5, 6, 7)),
		},
		"empty artists": {
			cs:             cmd.NewCheckSettings().WithEmpty(true),
			checkedArtists: cmd.PrepareConcernedArtists(generateArtists(1, 0, 10)),
			want:           true,
		},
		"empty albums": {
			cs:             cmd.NewCheckSettings().WithEmpty(true),
			checkedArtists: cmd.PrepareConcernedArtists(generateArtists(4, 6, 0)),
			want:           true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.cs.PerformEmptyAnalysis(tt.checkedArtists); got != tt.want {
				t.Errorf("CheckSettings.PerformEmptyAnalysis() = %v, want %v", got, tt.want)
			}
			verifiedFound := false
			for _, artist := range tt.checkedArtists {
				if artist.IsConcerned() {
					verifiedFound = true
				}
			}
			if verifiedFound != tt.want {
				t.Errorf("CheckSettings.PerformEmptyAnalysis() verified = %v, want %v",
					verifiedFound, tt.want)
			}
		})
	}
}

func TestGenerateMissingNumbers(t *testing.T) {
	type args struct {
		low  int
		high int
	}
	tests := map[string]struct {
		args
		want string
	}{
		"equal":   {args: args{low: 2, high: 2}, want: "2"},
		"inequal": {args: args{low: 2, high: 3}, want: "2-3"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := cmd.GenerateMissingNumbers(tt.args.low,
				tt.args.high); got != tt.want {
				t.Errorf("GenerateMissingNumbers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateNumberingConcerns(t *testing.T) {
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
			if got := cmd.GenerateNumberingConcerns(tt.args.m,
				tt.args.maxTrack); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateNumberingConcerns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckSettings_PerformNumberingAnalysis(t *testing.T) {
	defectiveArtists := []*files.Artist{}
	for r := 0; r < 4; r++ {
		artistName := fmt.Sprintf("my artist %d", r)
		artist := files.NewArtist(artistName, filepath.Join("Music", artistName))
		for k := 0; k < 5; k++ {
			albumName := fmt.Sprintf("my album %d%d", r, k)
			album := files.NewAlbum(albumName, artist, filepath.Join("Music", "my artist",
				albumName))
			for j := 1; j <= 6; j += 2 {
				trackName := fmt.Sprintf("my track %d%d%d", r, k, j)
				track := files.NewTrack(album, fmt.Sprintf("%d %s.mp3", j, trackName),
					trackName, j)
				album.AddTrack(track)
			}
			artist.AddAlbum(album)
		}
		defectiveArtists = append(defectiveArtists, artist)
	}

	tests := map[string]struct {
		cs             *cmd.CheckSettings
		checkedArtists []*cmd.ConcernedArtist
		want           bool
	}{
		"no analysis": {
			cs:             cmd.NewCheckSettings().WithNumbering(false),
			checkedArtists: cmd.PrepareConcernedArtists(generateArtists(5, 6, 7)),
			want:           false,
		},
		"ok analysis": {
			cs:             cmd.NewCheckSettings().WithNumbering(true),
			checkedArtists: cmd.PrepareConcernedArtists(generateArtists(5, 6, 7)),
			want:           false,
		},
		"missing numbers found": {
			cs:             cmd.NewCheckSettings().WithNumbering(true),
			checkedArtists: cmd.PrepareConcernedArtists(defectiveArtists),
			want:           true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.cs.PerformNumberingAnalysis(tt.checkedArtists); got != tt.want {
				t.Errorf("CheckSettings.PerformNumberingAnalysis() = %v, want %v", got,
					tt.want)
			}
			verifiedFound := false
			for _, artist := range tt.checkedArtists {
				if artist.IsConcerned() {
					verifiedFound = true
				}
			}
			if verifiedFound != tt.want {
				t.Errorf("CheckSettings.PerformNumberingAnalysis() verified = %v, want %v",
					verifiedFound, tt.want)
			}
		})
	}
}

func TestRecordFileConcerns(t *testing.T) {
	originalArtists := generateArtists(5, 6, 7)
	tracks := []*files.Track{}
	for _, artist := range originalArtists {
		copiedArtist := artist.Copy()
		for _, album := range artist.Albums() {
			copiedAlbum := album.Copy(copiedArtist, true)
			copiedArtist.AddAlbum(copiedAlbum)
			tracks = append(tracks, copiedAlbum.Tracks()...)
		}
	}
	type args struct {
		checkedArtists []*cmd.ConcernedArtist
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
				checkedArtists: cmd.PrepareConcernedArtists(originalArtists),
				track:          tracks[len(tracks)-1],
				concerns:       []string{"mismatched artist", "mismatched album"},
			},
			wantFoundConcerns: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := cmd.RecordFileConcerns(tt.args.checkedArtists,
				tt.args.track, tt.args.concerns); got != tt.wantFoundConcerns {
				t.Errorf("RecordFileConcerns() = %v, want %v", got, tt.wantFoundConcerns)
			}
			if tt.wantFoundConcerns {
				hasConcerns := false
				for _, cAr := range tt.args.checkedArtists {
					if cAr.IsConcerned() {
						hasConcerns = true
					}
				}
				if !hasConcerns {
					t.Errorf("RecordFileConcerns() true, but no concerns actually recorded")
				}
			}
		})
	}
}

func TestCheckSettings_PerformFileAnalysis(t *testing.T) {
	originalReadMetadata := cmd.ReadMetadata
	defer func() {
		cmd.ReadMetadata = originalReadMetadata
	}()
	cmd.ReadMetadata = func(_ output.Bus, _ []*files.Artist) {}
	type args struct {
		checkedArtists []*cmd.ConcernedArtist
		ss             *cmd.SearchSettings
	}
	tests := map[string]struct {
		cs *cmd.CheckSettings
		args
		want bool
		output.WantedRecording
	}{
		"not permitted to do anything": {
			cs:              cmd.NewCheckSettings().WithFiles(false),
			args:            args{},
			want:            false,
			WantedRecording: output.WantedRecording{},
		},
		"allowed, but nothing to check": {
			cs: cmd.NewCheckSettings().WithFiles(true),
			args: args{
				checkedArtists: []*cmd.ConcernedArtist{},
				ss:             cmd.NewSearchSettings(),
			},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No mp3 files remain after filtering.\n" +
					"Why?\n" +
					"After applying --artistFilter=<nil>, --albumFilter=<nil>, and" +
					" --trackFilter=<nil>, no files remained.\n" +
					"What to do:\n" +
					"Use less restrictive filter settings.\n",
				Log: "level='error' --albumFilter='<nil>' --artistFilter='<nil>'" +
					" --trackFilter='<nil>' msg='no files remain after filtering'\n",
			},
		},
		"work to do": {
			cs: cmd.NewCheckSettings().WithFiles(true),
			args: args{
				checkedArtists: cmd.PrepareConcernedArtists(generateArtists(4, 5, 6)),
				ss: cmd.NewSearchSettings().WithArtistFilter(
					regexp.MustCompile(".*")).WithAlbumFilter(regexp.MustCompile(
					".*")).WithTrackFilter(regexp.MustCompile(".*")),
			},
			want:            true,
			WantedRecording: output.WantedRecording{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if got := tt.cs.PerformFileAnalysis(o, tt.args.checkedArtists,
				tt.args.ss); got != tt.want {
				t.Errorf("CheckSettings.PerformFileAnalysis() = %v, want %v", got, tt.want)
			}
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("CheckSettings.PerformFileAnalysis() %s", difference)
				}
			}
		})
	}
}

func TestCheckSettings_MaybeReportCleanResults(t *testing.T) {
	type args struct {
		emptyConcerns     bool
		numberingConcerns bool
		fileConcerns      bool
	}
	tests := map[string]struct {
		cs *cmd.CheckSettings
		args
		output.WantedRecording
	}{
		"no concerns found because nothing was checked": {
			cs:              cmd.NewCheckSettings(),
			args:            args{},
			WantedRecording: output.WantedRecording{},
		},
		"all concerns found, everything was checked": {
			cs: cmd.NewCheckSettings().WithEmpty(true).WithNumbering(true).WithFiles(true),
			args: args{
				emptyConcerns:     true,
				numberingConcerns: true,
				fileConcerns:      true},
			WantedRecording: output.WantedRecording{},
		},
		"no concerns found, everything was checked": {
			cs:   cmd.NewCheckSettings().WithEmpty(true).WithNumbering(true).WithFiles(true),
			args: args{emptyConcerns: false, numberingConcerns: false, fileConcerns: false},
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
			tt.cs.MaybeReportCleanResults(o, tt.args.emptyConcerns, tt.args.numberingConcerns,
				tt.args.fileConcerns)
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("CheckSettings.MaybeReportCleanResults() %s", difference)
				}
			}
		})
	}
}

func TestCheckSettings_PerformChecks(t *testing.T) {
	originalReadMetadata := cmd.ReadMetadata
	defer func() {
		cmd.ReadMetadata = originalReadMetadata
	}()
	cmd.ReadMetadata = func(_ output.Bus, _ []*files.Artist) {}
	type args struct {
		artists       []*files.Artist
		artistsLoaded bool
		ss            *cmd.SearchSettings
	}
	tests := map[string]struct {
		cs *cmd.CheckSettings
		args
		wantStatus *cmd.ExitError
		output.WantedRecording
	}{
		"no artists loaded": {
			cs: nil,
			args: args{
				artists:       generateArtists(1, 1, 1),
				artistsLoaded: false,
				ss:            nil,
			},
			wantStatus:      cmd.NewExitUserError("check"),
			WantedRecording: output.WantedRecording{},
		},
		"no artists": {
			cs:              nil,
			args:            args{artists: nil, artistsLoaded: true, ss: nil},
			wantStatus:      cmd.NewExitUserError("check"),
			WantedRecording: output.WantedRecording{},
		},
		"artists to check, check everything": {
			cs: cmd.NewCheckSettings().WithEmpty(true).WithNumbering(true).WithFiles(true),
			args: args{
				artists:       generateArtists(1, 2, 3),
				artistsLoaded: true,
				ss: cmd.NewSearchSettings().WithArtistFilter(
					regexp.MustCompile(".*")).WithAlbumFilter(
					regexp.MustCompile(".*")).WithTrackFilter(regexp.MustCompile(".*")),
			},
			wantStatus: nil,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Artist \"my artist 0\"\n" +
					"* [files] for all albums: for all tracks: differences cannot be determined: metadata has not been read\n" +
					"Empty Folder Analysis: no empty folders found.\n" +
					"Numbering Analysis: no missing or duplicate tracks found.\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if got := tt.cs.PerformChecks(o, tt.args.artists, tt.args.artistsLoaded,
				tt.args.ss); !compareExitErrors(got, tt.wantStatus) {
				t.Errorf("CheckSettings.PerformChecks() got %s want %s", got, tt.wantStatus)
			}
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("CheckSettings.PerformChecks() %s", difference)
				}
			}
		})
	}
}

func TestCheckSettings_MaybeDoWork(t *testing.T) {
	tests := map[string]struct {
		cs         *cmd.CheckSettings
		ss         *cmd.SearchSettings
		wantStatus *cmd.ExitError
		output.WantedRecording
	}{
		"nothing to do": {
			cs:         cmd.NewCheckSettings(),
			ss:         nil,
			wantStatus: cmd.NewExitUserError("check"),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No checks will be executed.\n" +
					"Why?\n" +
					"The flags --empty, --files, and --numbering are all configured" +
					" false.\n" +
					"What to do:\n" +
					"Either:\n" +
					"[1] Edit the configuration file so that at least one of these flags" +
					" is true, or\n" +
					"[2] explicitly set at least one of these flags true on the command" +
					" line.\n",
			},
		},
		"try a little work": {
			cs: cmd.NewCheckSettings().WithEmpty(true),
			ss: cmd.NewSearchSettings().WithTopDirectory(
				filepath.Join(".", "no dir")).WithFileExtensions(
				[]string{".mp3"}).WithAlbumFilter(
				regexp.MustCompile(".*")).WithArtistFilter(
				regexp.MustCompile(".*")).WithTrackFilter(regexp.MustCompile(".*")),
			wantStatus: cmd.NewExitUserError("check"),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The directory \"no dir\" cannot be read: open no dir: The system" +
					" cannot find the file specified.\n" +
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
			if got := tt.cs.MaybeDoWork(o, tt.ss); !compareExitErrors(got, tt.wantStatus) {
				t.Errorf("CheckSettings.MaybeDoWork() got %s want %s", got, tt.wantStatus)
			}
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("CheckSettings.MaybeDoWork() %s", difference)
				}
			}
		})
	}
}

func TestCheckRun(t *testing.T) {
	cmd.InitGlobals()
	originalBus := cmd.Bus
	originalSearchFlags := cmd.SearchFlags
	defer func() {
		cmd.Bus = originalBus
		cmd.SearchFlags = originalSearchFlags
	}()
	cmd.SearchFlags = safeSearchFlags
	checkFlags := cmd.NewSectionFlags().WithSectionName(cmd.CheckCommand).WithFlags(
		map[string]*cmd.FlagDetails{
			cmd.CheckEmpty: cmd.NewFlagDetails().WithAbbreviatedName(
				cmd.CheckEmptyAbbr).WithUsage(
				"report empty album and artist directories").WithExpectedType(
				cmd.BoolType).WithDefaultValue(false),
			cmd.CheckFiles: cmd.NewFlagDetails().WithAbbreviatedName(
				cmd.CheckFilesAbbr).WithUsage(
				"report metadata/file inconsistencies").WithExpectedType(
				cmd.BoolType).WithDefaultValue(false),
			cmd.CheckNumbering: cmd.NewFlagDetails().WithAbbreviatedName(
				cmd.CheckNumberingAbbr).WithUsage(
				"report missing track numbers and duplicated track" +
					" numbering").WithExpectedType(cmd.BoolType).WithDefaultValue(false),
		},
	)
	command := &cobra.Command{}
	cmd.AddFlags(output.NewNilBus(), cmd_toolkit.EmptyConfiguration(), command.Flags(),
		checkFlags, cmd.SearchFlags)
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
					"The flags --empty, --files, and --numbering are all configured" +
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
					" --artistFilter='.*'" +
					" --empty='false'" +
					" --extensions='[.mp3]'" +
					" --files='false'" +
					" --numbering='false'" +
					" --topDir='.'" +
					" --trackFilter='.*'" +
					" command='check'" +
					" empty-user-set='false'" +
					" files-user-set='false'" +
					" numbering-user-set='false'" +
					" msg='executing command'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			cmd.Bus = o // cook getBus()
			cmd.CheckRun(tt.args.cmd, tt.args.in1)
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("CheckRun() %s", difference)
				}
			}
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

func TestCheckHelp(t *testing.T) {
	originalSearchFlags := cmd.SearchFlags
	defer func() {
		cmd.SearchFlags = originalSearchFlags
	}()
	cmd.SearchFlags = safeSearchFlags
	commandUnderTest := cloneCommand(cmd.CheckCmd)
	cmd.AddFlags(output.NewNilBus(), cmd_toolkit.EmptyConfiguration(),
		commandUnderTest.Flags(), cmd.CheckFlags, cmd.SearchFlags)
	tests := map[string]struct {
		output.WantedRecording
	}{
		"good": {
			WantedRecording: output.WantedRecording{
				Console: "" +
					"\"check\" inspects mp3 files and their containing directories and reports any problems detected\n" +
					"\n" +
					"Usage:\n" +
					"  check [--empty] [--files] [--numbering] [--albumFilter regex] [--artistFilter regex] [--trackFilter regex] [--topDir dir] [--extensions extensions]\n" +
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
					"      --albumFilter string    regular expression specifying which albums to select (default \".*\")\n" +
					"      --artistFilter string   regular expression specifying which artists to select (default \".*\")\n" +
					"  -e, --empty                 report empty album and artist directories (default false)\n" +
					"      --extensions string     comma-delimited list of file extensions used by mp3 files (default \".mp3\")\n" +
					"  -f, --files                 report metadata/file inconsistencies (default false)\n" +
					"  -n, --numbering             report missing track numbers and duplicated track numbering (default false)\n" +
					"      --topDir string         top directory specifying where to find mp3 files (default \".\")\n" +
					"      --trackFilter string    regular expression specifying which tracks to select (default \".*\")\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			command := commandUnderTest
			enableCommandRecording(o, command)
			command.Help()
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("check Help() %s", difference)
				}
			}
		})
	}
}

func TestCheckUsage(t *testing.T) {
	originalSearchFlags := cmd.SearchFlags
	defer func() {
		cmd.SearchFlags = originalSearchFlags
	}()
	cmd.SearchFlags = safeSearchFlags
	commandUnderTest := cloneCommand(cmd.CheckCmd)
	cmd.AddFlags(output.NewNilBus(), cmd_toolkit.EmptyConfiguration(),
		commandUnderTest.Flags(), cmd.CheckFlags, cmd.SearchFlags)
	tests := map[string]struct {
		output.WantedRecording
	}{
		"good": {
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Usage:\n" +
					"  check [--empty] [--files] [--numbering] [--albumFilter regex] [--artistFilter regex] [--trackFilter regex] [--topDir dir] [--extensions extensions]\n" +
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
			command.Usage()
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("check Usage() %s", difference)
				}
			}
		})
	}
}
