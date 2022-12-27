package commands

import (
	"flag"
	"mp3/internal"
	"mp3/internal/files"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/majohn-r/output"
)

var (
	fFlag = false
	tFlag = true
)

func Test_analyzeEmptyFolders(t *testing.T) {
	const fnName = "analyzeEmptyFolders()"
	emptyDir := "empty"
	dirtyDir := "dirty"
	goodDir := "good"
	if err := internal.Mkdir(emptyDir); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, emptyDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, emptyDir)
		internal.DestroyDirectoryForTesting(fnName, dirtyDir)
		internal.DestroyDirectoryForTesting(fnName, goodDir)
	}()
	if err := internal.Mkdir(dirtyDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, dirtyDir, err)
	}
	if err := internal.PopulateTopDirForTesting(dirtyDir); err != nil {
		t.Errorf("%s error populating %q: %v", fnName, dirtyDir, err)
	}
	if err := internal.Mkdir(goodDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, goodDir, err)
	}
	if err := internal.Mkdir(filepath.Join(goodDir, "goodArtist")); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, "goodArtist", err)
	}
	if err := internal.Mkdir(filepath.Join(goodDir, "goodArtist", "goodAlbum")); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, "good album", err)
	}
	if err := internal.CreateFileForTestingWithContent(filepath.Join(goodDir, "goodArtist", "goodAlbum"), "01 goodTrack.mp3", []byte("good content")); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, "01 goodTrack.mp3", err)
	}
	goodArtist := files.NewArtist("goodArtist", filepath.Join(goodDir, "goodArtist"))
	goodAlbum := files.NewAlbum("goodAlbum", goodArtist, filepath.Join(goodDir, "goodArtist", "goodAlbum"))
	goodArtist.AddAlbum(goodAlbum)
	goodTrack := files.NewTrack(goodAlbum, "01 goodTrack.mp3", "goodTrack", 1)
	goodAlbum.AddTrack(goodTrack)
	type args struct {
		s *files.Search
	}
	tests := []struct {
		name string
		c    *check
		args
		wantArtists         []*files.Artist
		wantFilteredArtists []*checkedArtist
		wantOk              bool
		output.WantedRecording
	}{
		{name: "no work to do", c: &check{emptyFolders: &fFlag}, args: args{}, wantOk: true},
		{
			name: "empty topDir",
			c:    &check{emptyFolders: &tFlag},
			args: args{s: files.CreateSearchForTesting(emptyDir)},
			WantedRecording: output.WantedRecording{
				Error: "No music files could be found using the specified parameters.\n",
				Log: "level='info' -ext='.mp3' -topDir='empty' msg='reading unfiltered music files'\n" +
					"level='error' -ext='.mp3' -topDir='empty' msg='cannot find any artist directories'\n",
			},
		},
		{
			name:        "folders, no empty folders present",
			c:           &check{emptyFolders: &tFlag},
			args:        args{s: files.CreateSearchForTesting(goodDir)},
			wantArtists: []*files.Artist{goodArtist},
			wantOk:      true,
			WantedRecording: output.WantedRecording{
				Console: "Empty Folder Analysis: no empty folders found.\n",
				Log:     "level='info' -ext='.mp3' -topDir='good' msg='reading unfiltered music files'\n",
			},
		},
		{
			name:        "empty folders present",
			c:           &check{emptyFolders: &tFlag},
			args:        args{s: files.CreateSearchForTesting(dirtyDir)},
			wantArtists: files.CreateAllArtistsForTesting(dirtyDir, true),
			wantOk:      true,
			wantFilteredArtists: []*checkedArtist{
				{
					backing: files.NewArtist("Test Artist 0", "Test Artist 0"),
					albums: []*checkedAlbum{{
						backing: files.NewAlbum("Test Album 999", nil, "Test Artist 0/Test Album 999"),
						issues:  []string{"no tracks found"}}},
				},
				{
					backing: files.NewArtist("Test Artist 1", "Test Artist 1"),
					albums: []*checkedAlbum{{
						backing: files.NewAlbum("Test Album 999", nil, "Test Artist 1/Test Album 999"),
						issues:  []string{"no tracks found"},
					}},
				},
				{
					backing: files.NewArtist("Test Artist 2", "Test Artist 2"),
					albums: []*checkedAlbum{{
						backing: files.NewAlbum("Test Album 999", nil, "Test Artist 2/Test Album 999"),
						issues:  []string{"no tracks found"},
					}},
				},
				{
					backing: files.NewArtist("Test Artist 3", "Test Artist 3"),
					albums: []*checkedAlbum{{
						backing: files.NewAlbum("Test Album 999", nil, "Test Artist 3/Test Album 999"),
						issues:  []string{"no tracks found"},
					}},
				},
				{
					backing: files.NewArtist("Test Artist 4", "Test Artist 4"),
					albums: []*checkedAlbum{{
						backing: files.NewAlbum("Test Album 999", nil, "Test Artist 4/Test Album 999"),
						issues:  []string{"no tracks found"},
					}},
				},
				{
					backing: files.NewArtist("Test Artist 5", "Test Artist 5"),
					albums: []*checkedAlbum{{
						backing: files.NewAlbum("Test Album 999", nil, "Test Artist 5/Test Album 999"),
						issues:  []string{"no tracks found"},
					}},
				},
				{
					backing: files.NewArtist("Test Artist 6", "Test Artist 6"),
					albums: []*checkedAlbum{{
						backing: files.NewAlbum("Test Album 999", nil, "Test Artist 6/Test Album 999"),
						issues:  []string{"no tracks found"},
					}},
				},
				{
					backing: files.NewArtist("Test Artist 7", "Test Artist 7"),
					albums: []*checkedAlbum{{
						backing: files.NewAlbum("Test Album 999", nil, "Test Artist 7/Test Album 999"),
						issues:  []string{"no tracks found"},
					}},
				},
				{
					backing: files.NewArtist("Test Artist 8", "Test Artist 8"),
					albums: []*checkedAlbum{{
						backing: files.NewAlbum("Test Album 999", nil, "Test Artist 8/Test Album 999"),
						issues:  []string{"no tracks found"},
					}},
				},
				{
					backing: files.NewArtist("Test Artist 9", "Test Artist 9"),
					albums: []*checkedAlbum{{
						backing: files.NewAlbum("Test Album 999", nil, "Test Artist 9/Test Album 999"),
						issues:  []string{"no tracks found"},
					}},
				},
				{
					backing: files.NewArtist("Test Artist 999", "Test Artist 999"),
					issues:  []string{"no albums found"},
				},
			},
			WantedRecording: output.WantedRecording{
				Log: "level='info' -ext='.mp3' -topDir='dirty' msg='reading unfiltered music files'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			gotArtists, gotArtistsWithIssues, gotOk := tt.c.analyzeEmptyFolders(o, tt.args.s)
			if !reflect.DeepEqual(gotArtists, tt.wantArtists) {
				t.Errorf("%s = %v, want %v", fnName, gotArtists, tt.wantArtists)
			} else {
				filteredArtists := filterAndSortCheckedArtists(gotArtistsWithIssues)
				if !equalCheckedArtists(filteredArtists, tt.wantFilteredArtists) {
					t.Errorf("%s = %v, want %v", fnName, filteredArtists, tt.wantFilteredArtists)
				}
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s ok = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_filterArtists(t *testing.T) {
	const fnName = "filterArtists()"
	topDir := "filterArtists"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, topDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	if err := internal.PopulateTopDirForTesting(topDir); err != nil {
		t.Errorf("%s error populating %q: %v", fnName, topDir, err)
	}
	searchStruct := files.CreateSearchForTesting(topDir)
	fullArtists, _ := searchStruct.LoadUnfilteredData(output.NewNilBus())
	filteredArtists, _ := searchStruct.LoadData(output.NewNilBus())
	type args struct {
		s       *files.Search
		artists []*files.Artist
	}
	tests := []struct {
		name string
		c    *check
		args
		wantFilteredArtists []*files.Artist
		wantOk              bool
		output.WantedRecording
	}{
		{
			name:   "neither gap analysis nor integrity enabled",
			c:      &check{trackNumberingGaps: &fFlag, integrity: &fFlag},
			args:   args{s: nil, artists: nil},
			wantOk: true,
		},
		{
			name:                "only gap analysis enabled, no artists supplied",
			c:                   &check{trackNumberingGaps: &tFlag, integrity: &fFlag},
			args:                args{s: searchStruct},
			wantFilteredArtists: filteredArtists,
			wantOk:              true,
			WantedRecording: output.WantedRecording{
				Log: "level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='filterArtists' msg='reading filtered music files'\n",
			},
		},
		{
			name:                "only gap analysis enabled, artists supplied",
			c:                   &check{trackNumberingGaps: &tFlag, integrity: &fFlag},
			args:                args{s: searchStruct, artists: fullArtists},
			wantFilteredArtists: filteredArtists,
			wantOk:              true,
			WantedRecording: output.WantedRecording{
				Log: "level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='filterArtists' msg='filtering music files'\n",
			},
		},
		{
			name:                "only integrity check enabled, no artists supplied",
			c:                   &check{trackNumberingGaps: &fFlag, integrity: &tFlag},
			args:                args{s: searchStruct},
			wantFilteredArtists: filteredArtists,
			wantOk:              true,
			WantedRecording: output.WantedRecording{
				Log: "level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='filterArtists' msg='reading filtered music files'\n",
			},
		},
		{
			name:                "only integrity check enabled, artists supplied",
			c:                   &check{trackNumberingGaps: &fFlag, integrity: &tFlag},
			args:                args{s: searchStruct, artists: fullArtists},
			wantFilteredArtists: filteredArtists,
			wantOk:              true,
			WantedRecording: output.WantedRecording{
				Log: "level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='filterArtists' msg='filtering music files'\n",
			},
		},
		{
			name:                "gap analysis and integrity check enabled, no artists supplied",
			c:                   &check{trackNumberingGaps: &tFlag, integrity: &tFlag},
			args:                args{s: searchStruct},
			wantFilteredArtists: filteredArtists,
			wantOk:              true,
			WantedRecording: output.WantedRecording{
				Log: "level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='filterArtists' msg='reading filtered music files'\n",
			},
		},
		{
			name:                "gap analysis and integrity check enabled, artists supplied",
			c:                   &check{trackNumberingGaps: &tFlag, integrity: &tFlag},
			args:                args{s: searchStruct, artists: fullArtists},
			wantFilteredArtists: filteredArtists,
			wantOk:              true,
			WantedRecording: output.WantedRecording{
				Log: "level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='filterArtists' msg='filtering music files'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			gotFilteredArtists, gotOk := tt.c.filterArtists(o, tt.args.s, tt.args.artists)
			if !reflect.DeepEqual(gotFilteredArtists, tt.wantFilteredArtists) {
				t.Errorf("%s = %v, want %v", fnName, gotFilteredArtists, tt.wantFilteredArtists)
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s ok = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_check_analyzeGaps(t *testing.T) {
	const fnName = "check.analyzeGaps()"
	goodArtist := files.NewArtist("My Good Artist", "")
	goodAlbum := files.NewAlbum("An Excellent Album", goodArtist, "")
	goodArtist.AddAlbum(goodAlbum)
	goodAlbum.AddTrack(files.NewTrack(goodAlbum, "", "First Track", 1))
	goodAlbum.AddTrack(files.NewTrack(goodAlbum, "", "Second Track", 2))
	goodAlbum.AddTrack(files.NewTrack(goodAlbum, "", "Third Track", 3))
	badArtist := files.NewArtist("BadArtist", "")
	badAlbum := files.NewAlbum("No Biscuits For You!", badArtist, "")
	badArtist.AddAlbum(badAlbum)
	badAlbum.AddTrack(files.NewTrack(badAlbum, "", "Awful Track", 0))
	badAlbum.AddTrack(files.NewTrack(badAlbum, "", "Nasty Track", 1))
	badAlbum.AddTrack(files.NewTrack(badAlbum, "", "Worse Track", 1))
	badAlbum.AddTrack(files.NewTrack(badAlbum, "", "Bonus Track", 9))
	type args struct {
		artists []*files.Artist
	}
	tests := []struct {
		name string
		c    *check
		args
		wantConflictedArtists []*checkedArtist
		output.WantedRecording
	}{
		{name: "no analysis", c: &check{trackNumberingGaps: &fFlag}, args: args{}},
		{
			name: "no content",
			c:    &check{trackNumberingGaps: &tFlag},
			args: args{},
			WantedRecording: output.WantedRecording{
				Console: "Check Gaps: no gaps found.\n",
			},
		},
		{
			name: "good artist",
			c:    &check{trackNumberingGaps: &tFlag},
			args: args{artists: []*files.Artist{goodArtist}},
			WantedRecording: output.WantedRecording{
				Console: "Check Gaps: no gaps found.\n",
			},
		},
		{
			name: "bad artist",
			c:    &check{trackNumberingGaps: &tFlag},
			args: args{artists: []*files.Artist{badArtist}},
			wantConflictedArtists: []*checkedArtist{
				{
					backing: files.NewArtist("BadArtist", "BadArtist"),
					albums: []*checkedAlbum{
						{
							backing: files.NewAlbum("No Biscuits For You!", badArtist, "BadArtist/No Biscuits for You!"),
							issues: []string{
								"missing track 2",
								"missing track 3",
								"missing track 4",
								"track 0 (\"Awful Track\") is not a valid track number; valid tracks are 1..7",
								"track 1 used by \"Nasty Track\" and \"Worse Track\"",
								"track 9 (\"Bonus Track\") is not a valid track number; valid tracks are 1..7",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			if gotConflictedArtists := filterAndSortCheckedArtists(tt.c.analyzeGaps(o, tt.args.artists)); !equalCheckedArtists(gotConflictedArtists, tt.wantConflictedArtists) {
				t.Errorf("%s = %v, want %v", fnName, gotConflictedArtists, tt.wantConflictedArtists)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_check_analyzeIntegrity(t *testing.T) {
	const fnName = "check.analyzeIntegrity()"
	// create some data to work with
	topDir := "integrity"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, topDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	// keep it simple: one artist, one album, one track
	artistPath := filepath.Join(topDir, "artist")
	if err := internal.Mkdir(artistPath); err != nil {
		t.Errorf("%s error creating artist folder", fnName)
	}
	backingArtist := files.NewArtist("artist", artistPath)
	albumPath := filepath.Join(artistPath, "album")
	if err := internal.Mkdir(albumPath); err != nil {
		t.Errorf("%s error creating album folder", fnName)
	}
	backingAlbum := files.NewAlbum("album", backingArtist, albumPath)
	backingArtist.AddAlbum(backingAlbum)
	if err := internal.CreateFileForTestingWithContent(albumPath, "01 track.mp3", nil); err != nil {
		t.Errorf("%s error creating track", fnName)
	}
	backingTrack := files.NewTrack(backingAlbum, "01 track.mp3", "track", 1)
	backingAlbum.AddTrack(backingTrack)
	s := files.CreateSearchForTesting(topDir)
	a, _ := s.LoadUnfilteredData(output.NewNilBus())
	type args struct {
		artists []*files.Artist
	}
	tests := []struct {
		name string
		c    *check
		args
		wantConflictedArtists []*checkedArtist
		output.WantedRecording
	}{
		{name: "degenerate case", c: &check{integrity: &fFlag}, args: args{}},
		{
			name: "no artists",
			c:    &check{integrity: &tFlag},
			args: args{},
			WantedRecording: output.WantedRecording{
				Console: "Integrity Analysis: no issues found.\n",
				Error:   "Reading track metadata.\n",
			},
		},
		{
			name: "meaningful case",
			c:    &check{integrity: &tFlag},
			args: args{artists: a},
			WantedRecording: output.WantedRecording{
				Error: "Reading track metadata.\n" +
					"An error occurred when trying to read ID3V1 tag information for track \"track\" on album \"album\" by artist \"artist\": \"seek integrity\\\\artist\\\\album\\\\01 track.mp3: An attempt was made to move the file pointer before the beginning of the file.\".\n" +
					"An error occurred when trying to read ID3V2 tag information for track \"track\" on album \"album\" by artist \"artist\": \"zero length\".\n",
				Log: "level='error' error='seek integrity\\artist\\album\\01 track.mp3: An attempt was made to move the file pointer before the beginning of the file.' track='integrity\\artist\\album\\01 track.mp3' msg='id3v1 tag error'\n" +
					"level='error' error='zero length' track='integrity\\artist\\album\\01 track.mp3' msg='id3v2 tag error'\n",
			},
			wantConflictedArtists: []*checkedArtist{
				{
					backing: backingArtist,
					albums: []*checkedAlbum{
						{
							backing: backingAlbum,
							tracks: []*checkedTrack{
								{
									backing: backingTrack,
									issues:  []string{"differences cannot be determined: there was an error reading metadata"},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			gotConflictedArtists := filterAndSortCheckedArtists(tt.c.analyzeIntegrity(o, tt.args.artists))
			if !equalCheckedArtists(gotConflictedArtists, tt.wantConflictedArtists) {
				t.Errorf("%s = %#v, want %#v", fnName, gotConflictedArtists, tt.wantConflictedArtists)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func equalCheckedArtists(got, want []*checkedArtist) bool {
	if len(got) != len(want) {
		return false
	}
	for n, gotCAr := range got {
		if !equalCheckedArtist(gotCAr, want[n]) {
			return false
		}
	}
	return true
}

func equalCheckedArtist(got, want *checkedArtist) bool {
	if !reflect.DeepEqual(got.issues, want.issues) {
		return false
	}
	if got.backing.Name() != want.backing.Name() {
		return false
	}
	return equalCheckedAlbums(got.albums, want.albums)
}

func equalCheckedAlbums(got, want []*checkedAlbum) bool {
	if len(got) != len(want) {
		return false
	}
	for n, gotCAl := range got {
		if !equalCheckedAlbum(gotCAl, want[n]) {
			return false
		}
	}
	return true
}

func equalCheckedAlbum(got, want *checkedAlbum) bool {
	if !reflect.DeepEqual(got.issues, want.issues) {
		return false
	}
	if got.backing.Name() != want.backing.Name() {
		return false
	}
	return equalCheckedTracks(got.tracks, want.tracks)
}

func equalCheckedTracks(got, want []*checkedTrack) bool {
	if len(got) != len(want) {
		return false
	}
	for n, gotCT := range got {
		if !equalCheckedTrack(gotCT, want[n]) {
			return false
		}
	}
	return true
}

func equalCheckedTrack(got, want *checkedTrack) bool {
	if !reflect.DeepEqual(got.issues, want.issues) {
		return false
	}
	if got.backing.Name() != want.backing.Name() {
		return false
	}
	return got.backing.Number() == want.backing.Number()
}

func makeCheckCommand() *check {
	c, _ := newCheckCommand(output.NewNilBus(), internal.EmptyConfiguration(), flag.NewFlagSet("check", flag.ContinueOnError))
	return c
}

func Test_check_Exec(t *testing.T) {
	const fnName = "check.Exec()"
	topDir := "checkExec"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, topDir, err)
	}
	savedHomePath := internal.SaveEnvVarForTesting("HOMEPATH")
	homePath := internal.SavedEnvVar{
		Name:  "HOMEPATH",
		Value: "C:\\Users\\The User",
		Set:   true,
	}
	homePath.RestoreForTesting()
	defer func() {
		savedHomePath.RestoreForTesting()
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	if err := internal.PopulateTopDirForTesting(topDir); err != nil {
		t.Errorf("%s error populating directory %q: %v", fnName, topDir, err)
	}
	type args struct {
		args []string
	}
	tests := []struct {
		name string
		c    *check
		args
		wantOk bool
		output.WantedRecording
	}{
		{
			name: "help",
			c:    makeCheckCommand(),
			args: args{[]string{"--help"}},
			WantedRecording: output.WantedRecording{
				Error: "Usage of check:\n" +
					"  -albumFilter regular expression\n" +
					"    \tregular expression specifying which albums to select (default \".*\")\n" +
					"  -artistFilter regular expression\n" +
					"    \tregular expression specifying which artists to select (default \".*\")\n" +
					"  -empty\n" +
					"    \tcheck for empty artist and album folders (default false)\n" +
					"  -ext extension\n" +
					"    \textension identifying music files (default \".mp3\")\n" +
					"  -gaps\n" +
					"    \tcheck for gaps in track numbers (default false)\n" +
					"  -integrity\n" +
					"    \tcheck for disagreement between the file system and audio file metadata (default true)\n" +
					"  -topDir directory\n" +
					"    \ttop directory specifying where to find music files (default \"C:\\\\Users\\\\The User\\\\Music\")\n",
				Log: "level='error' arguments='[--help]' msg='flag: help requested'\n",
			},
		},
		{
			name: "do nothing",
			c:    makeCheckCommand(),
			args: args{[]string{"-topDir", topDir, "-empty=false", "-gaps=false", "-integrity=false"}},
			WantedRecording: output.WantedRecording{
				Error: "You disabled all functionality for the command \"check\".\n",
				Log:   "level='error' -empty='false' -gaps='false' -integrity='false' command='check' msg='the user disabled all functionality'\n",
			},
		},
		{
			name:   "do something",
			c:      makeCheckCommand(),
			args:   args{[]string{"-topDir", topDir, "-empty=true", "-gaps=false", "-integrity=false"}},
			wantOk: true,
			WantedRecording: output.WantedRecording{
				Console: strings.Join([]string{
					"Test Artist 0",
					"    Test Album 999",
					"      no tracks found",
					"Test Artist 1",
					"    Test Album 999",
					"      no tracks found",
					"Test Artist 2",
					"    Test Album 999",
					"      no tracks found",
					"Test Artist 3",
					"    Test Album 999",
					"      no tracks found",
					"Test Artist 4",
					"    Test Album 999",
					"      no tracks found",
					"Test Artist 5",
					"    Test Album 999",
					"      no tracks found",
					"Test Artist 6",
					"    Test Album 999",
					"      no tracks found",
					"Test Artist 7",
					"    Test Album 999",
					"      no tracks found",
					"Test Artist 8",
					"    Test Album 999",
					"      no tracks found",
					"Test Artist 9",
					"    Test Album 999",
					"      no tracks found",
					"Test Artist 999",
					"  no albums found",
					"",
				}, "\n"),
				Log: "level='info' -empty='true' -gaps='false' -integrity='false' command='check' msg='executing command'\n" +
					"level='info' -ext='.mp3' -topDir='checkExec' msg='reading unfiltered music files'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			if gotOk := tt.c.Exec(o, tt.args.args); gotOk != tt.wantOk {
				t.Errorf("%s ok = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_newCheckCommand(t *testing.T) {
	const fnName = "newCheckCommand()"
	savedAppData := internal.SaveEnvVarForTesting("APPDATA")
	os.Setenv("APPDATA", internal.SecureAbsolutePathForTesting("."))
	defer func() {
		savedAppData.RestoreForTesting()
	}()
	topDir := "loadTest"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, topDir, err)
	}
	if err := internal.PopulateTopDirForTesting(topDir); err != nil {
		t.Errorf("%s error populating %q: %v", fnName, topDir, err)
	}
	if err := internal.CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("%s error creating defaults.yaml: %v", fnName, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
		internal.DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	config, _ := internal.ReadConfigurationFile(output.NewNilBus())
	type args struct {
		c *internal.Configuration
	}
	tests := []struct {
		name string
		args
		wantEmptyFolders         bool
		wantGapsInTrackNumbering bool
		wantIntegrity            bool
		wantOk                   bool
		output.WantedRecording
	}{
		{
			name:                     "ordinary defaults",
			args:                     args{c: internal.EmptyConfiguration()},
			wantEmptyFolders:         false,
			wantGapsInTrackNumbering: false,
			wantIntegrity:            true,
			wantOk:                   true,
		},
		{
			name:                     "overridden defaults",
			args:                     args{c: config},
			wantEmptyFolders:         true,
			wantGapsInTrackNumbering: true,
			wantIntegrity:            false,
			wantOk:                   true,
		},
		{
			name: "bad default empty folder",
			args: args{
				c: internal.CreateConfiguration(output.NewNilBus(), map[string]any{
					"check": map[string]any{
						emptyFolders: "Empty!!",
					},
				}),
			},
			wantOk: false,
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"check\": invalid boolean value \"Empty!!\" for -empty: parse error.\n",
				Log:   "level='error' error='invalid boolean value \"Empty!!\" for -empty: parse error' section='check' msg='invalid content in configuration file'\n",
			},
		},
		{
			name: "bad default gaps",
			args: args{
				c: internal.CreateConfiguration(output.NewNilBus(), map[string]any{
					"check": map[string]any{
						trackNumberingGaps: "No",
					},
				}),
			},
			wantOk: false,
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"check\": invalid boolean value \"No\" for -gaps: parse error.\n",
				Log:   "level='error' error='invalid boolean value \"No\" for -gaps: parse error' section='check' msg='invalid content in configuration file'\n",
			},
		},
		{
			name: "bad default integrity",
			args: args{
				c: internal.CreateConfiguration(output.NewNilBus(), map[string]any{
					"check": map[string]any{
						integrity: "Off",
					},
				}),
			},
			wantOk: false,
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"check\": invalid boolean value \"Off\" for -integrity: parse error.\n",
				Log:   "level='error' error='invalid boolean value \"Off\" for -integrity: parse error' section='check' msg='invalid content in configuration file'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			c, gotOk := newCheckCommand(o, tt.args.c, flag.NewFlagSet("check", flag.ContinueOnError))
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk %t wantOk %t", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
			if c != nil {
				if _, ok := c.sf.ProcessArgs(output.NewNilBus(), []string{
					"-topDir", topDir,
					"-ext", ".mp3",
				}); ok {
					if *c.emptyFolders != tt.wantEmptyFolders {
						t.Errorf("%s %q: got checkEmptyFolders %t want %t", fnName, tt.name, *c.emptyFolders, tt.wantEmptyFolders)
					}
					if *c.trackNumberingGaps != tt.wantGapsInTrackNumbering {
						t.Errorf("%s %q: got checkGapsInTrackNumbering %t want %t", fnName, tt.name, *c.trackNumberingGaps, tt.wantGapsInTrackNumbering)
					}
					if *c.integrity != tt.wantIntegrity {
						t.Errorf("%s %q: got checkIntegrity %t want %t", fnName, tt.name, *c.integrity, tt.wantIntegrity)
					}
				} else {
					t.Errorf("%s %q: error processing arguments", fnName, tt.name)
				}
			}
		})
	}
}

func Test_merge(t *testing.T) {
	const fnName = "merge()"
	type args struct {
		sets [][]*checkedArtist
	}
	tests := []struct {
		name string
		args
		want []*checkedArtist
	}{
		{name: "degenerate case", args: args{}},
		{
			name: "more interesting case",
			args: args{sets: [][]*checkedArtist{
				// set 1
				{
					{
						backing: files.NewArtist("artist1", "artist1"),
						issues:  []string{"bad artist"},
						albums: []*checkedAlbum{
							{
								backing: files.NewAlbum("album1", nil, "artist1/album1"),
								issues:  []string{"skips badly"},
								tracks: []*checkedTrack{
									{
										backing: files.NewTrack(files.NewAlbum("album1", files.NewArtist("artist1", "./artist1"), "./artist1/album1"), "01 track1.mp3", "track1", 1),
										issues:  []string{"inaudible"},
									},
								},
							},
						},
					},
				},
				// set 2
				{
					{
						backing: files.NewArtist("artist1", "artist1"),
						issues:  []string{"really awful artist"},
						albums: []*checkedAlbum{
							{
								backing: files.NewAlbum("album1", nil, "artist1/album1"),
								issues:  []string{"bad cover art"},
								tracks: []*checkedTrack{
									{
										backing: files.NewTrack(files.NewAlbum("album1", files.NewArtist("artist1", "./artist1"), "./artist1/album1"), "01 track1.mp3", "track1", 1),
										issues:  []string{"plays backwards!"},
									},
									{
										backing: files.NewTrack(files.NewAlbum("album1", files.NewArtist("artist1", "./artist1"), "./artist1/album1"), "02 track2.mp3", "track2", 2),
										issues:  []string{"truly insipid"},
									},
								},
							},
							{
								backing: files.NewAlbum("album2", nil, "artist1/album2"),
								issues:  []string{"horrible sequel"},
								tracks: []*checkedTrack{
									{
										backing: files.NewTrack(files.NewAlbum("album2", files.NewArtist("artist1", "./artist1"), "./artist1/album2"), "03 track3.mp3", "track3", 3),
										issues:  []string{"singer is dreadful, band is worse"},
									},
								},
							},
						},
					},
					{
						backing: files.NewArtist("artist2", "artist2"),
						issues:  []string{"tone deaf"},
						albums: []*checkedAlbum{
							{
								backing: files.NewAlbum("album34", nil, "artist2/album34"),
								issues:  []string{"worst album I own"},
								tracks: []*checkedTrack{
									{
										backing: files.NewTrack(files.NewAlbum("album34", files.NewArtist("artist2", "./artist2"), "./artist2/album34"), "40 track40.mp3", "track40", 40),
										issues:  []string{"singer died in mid act and that improved the track"},
									},
								},
							},
						},
					},
				},
			}},
			want: []*checkedArtist{
				{
					backing: files.NewArtist("artist1", "artist1"),
					issues:  []string{"bad artist", "really awful artist"},
					albums: []*checkedAlbum{
						{
							backing: files.NewAlbum("album1", nil, "artist1/album1"),
							issues:  []string{"bad cover art", "skips badly"},
							tracks: []*checkedTrack{
								{
									backing: files.NewTrack(files.NewAlbum("album1", files.NewArtist("artist1", "./artist1"), "./artist1/album1"), "01 track1.mp3", "track1", 1),
									issues:  []string{"inaudible", "plays backwards!"},
								},
								{
									backing: files.NewTrack(files.NewAlbum("album1", files.NewArtist("artist1", "./artist1"), "./artist1/album1"), "02 track2.mp3", "track2", 2),
									issues:  []string{"truly insipid"},
								},
							},
						},
						{
							backing: files.NewAlbum("album2", nil, "artist1/album2"),
							issues:  []string{"horrible sequel"},
							tracks: []*checkedTrack{
								{
									backing: files.NewTrack(files.NewAlbum("album2", files.NewArtist("artist1", "./artist1"), "./artist1/album2"), "03 track3.mp3", "track3", 3),
									issues:  []string{"singer is dreadful, band is worse"},
								},
							},
						},
					},
				},
				{
					backing: files.NewArtist("artist2", "artist2"),
					issues:  []string{"tone deaf"},
					albums: []*checkedAlbum{
						{
							backing: files.NewAlbum("album34", nil, "artist2/album34"),
							issues:  []string{"worst album I own"},
							tracks: []*checkedTrack{
								{
									backing: files.NewTrack(files.NewAlbum("album34", files.NewArtist("artist2", "./artist2"), "./artist2/album34"), "40 track40.mp3", "track40", 40),
									issues:  []string{"singer died in mid act and that improved the track"},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := merge(tt.args.sets)
			if len(got) != len(tt.want) {
				t.Errorf("%s artist len = %d, want %d", fnName, len(got), len(tt.want))
			} else {
				for i := range got {
					gotArtist := got[i]
					wantArtist := tt.want[i]
					if gotArtist.backing.Name() != wantArtist.backing.Name() {
						t.Errorf("%s artist[%d] name %q, want %q", fnName, i, gotArtist.backing.Name(), wantArtist.backing.Name())
					}
					if !reflect.DeepEqual(gotArtist.issues, wantArtist.issues) {
						t.Errorf("%s artist[%d] issues %v, want %v", fnName, i, gotArtist.issues, wantArtist.issues)
					}
					if len(gotArtist.albums) != len(wantArtist.albums) {
						t.Errorf("%s artist[%d] albums len = %d, want %d", fnName, i, len(gotArtist.albums), len(wantArtist.albums))
					} else {
						for j := range gotArtist.albums {
							gotAlbum := gotArtist.albums[j]
							wantAlbum := wantArtist.albums[j]
							if gotAlbum.backing.Name() != wantAlbum.backing.Name() {
								t.Errorf("%s artist[%d] album[%d] name %q, want %q", fnName, i, j, gotAlbum.backing.Name(), wantAlbum.backing.Name())
							}
							if !reflect.DeepEqual(gotAlbum.issues, wantAlbum.issues) {
								t.Errorf("%s artist[%d] album[%d] issues %v, want %v", fnName, i, j, gotAlbum.issues, wantAlbum.issues)
							}
							if len(gotAlbum.tracks) != len(wantAlbum.tracks) {
								t.Errorf("%s artist[%d] album[%d] tracks len = %d, want %d", fnName, i, j, len(gotAlbum.tracks), len(wantAlbum.tracks))
							} else {
								for k := range gotAlbum.tracks {
									gotTrack := gotAlbum.tracks[k]
									wantTrack := wantAlbum.tracks[k]
									if gotTrack.backing.Number() != wantTrack.backing.Number() {
										t.Errorf("%s artist[%d] album[%d] track[%d] number %d, want %d", fnName, i, j, k, gotTrack.backing.Number(), wantTrack.backing.Number())
									}
									if gotTrack.backing.Name() != wantTrack.backing.Name() {
										t.Errorf("%s artist[%d] album[%d] track[%d] name %q, want %q", fnName, i, j, k, gotTrack.backing.Name(), wantTrack.backing.Name())
									}
									if !reflect.DeepEqual(gotTrack.issues, wantTrack.issues) {
										t.Errorf("%s artist[%d] album[%d] track[%d] issues %v, want %v", fnName, i, j, k, gotTrack.issues, wantTrack.issues)
									}
								}
							}
						}
					}
				}
			}
		})
	}
}

func Test_sortCheckedArtistsWithIssues(t *testing.T) {
	const fnName = "sortCheckedArtistsWithIssues()"
	tests := []struct {
		name  string
		input checkedArtistSlice
	}{
		{name: "degenerate case", input: nil},
		{name: "scrambled input", input: checkedArtistSlice([]*checkedArtist{
			{backing: files.NewArtist("10", "10")},
			{backing: files.NewArtist("2", "2")},
			{backing: files.NewArtist("35", "35")},
			{backing: files.NewArtist("1", "1")},
			{backing: files.NewArtist("2", "2")},
		})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sort.Sort(tt.input)
			for i := range tt.input {
				if i == 0 {
					continue
				}
				if tt.input[i-1].backing.Name() > tt.input[i].backing.Name() {
					t.Errorf("%s artist[%d] with name %q comes before artist[%d] with name %q", fnName, i-1, tt.input[i-1].backing.Name(), i, tt.input[i].backing.Name())
				}
			}
		})
	}
}

func Test_sortAlbumsWithIssues(t *testing.T) {
	const fnName = "sortAlbumsWithIssues()"
	tests := []struct {
		name  string
		input checkedAlbumSlice
	}{
		{name: "degenerate case", input: nil},
		{name: "scrambled input", input: checkedAlbumSlice([]*checkedAlbum{
			{backing: files.NewAlbum("10", nil, "artist/10")},
			{backing: files.NewAlbum("2", nil, "artist/2")},
			{backing: files.NewAlbum("35", nil, "artist/35")},
			{backing: files.NewAlbum("1", nil, "artist/1")},
			{backing: files.NewAlbum("3", nil, "artist/3")},
		})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sort.Sort(tt.input)
			for i := range tt.input {
				if i == 0 {
					continue
				}
				if tt.input[i-1].backing.Name() > tt.input[i].backing.Name() {
					t.Errorf("%s album[%d] with name %q comes before album[%d] with name %q", fnName, i-1, tt.input[i-1].backing.Name(), i, tt.input[i].backing.Name())
				}
			}
		})
	}
}

func Test_sortTracksWithIssues(t *testing.T) {
	const fnName = "sortTracksWithIssues()"
	tests := []struct {
		name  string
		input checkedTrackSlice
	}{
		{name: "degenerate case", input: nil},
		{name: "scrambled input", input: checkedTrackSlice([]*checkedTrack{
			{backing: files.NewTrack(files.NewAlbum("album1", files.NewArtist("artist1", "./artist1"), "./artist1/album1"), "10 track10.mp3", "track10", 10)},
			{backing: files.NewTrack(files.NewAlbum("album1", files.NewArtist("artist1", "./artist1"), "./artist1/album1"), "02 track2.mp3", "track2", 2)},
			{backing: files.NewTrack(files.NewAlbum("album1", files.NewArtist("artist1", "./artist1"), "./artist1/album1"), "35 track35.mp3", "track35", 35)},
			{backing: files.NewTrack(files.NewAlbum("album1", files.NewArtist("artist1", "./artist1"), "./artist1/album1"), "01 track1.mp3", "track1", 1)},
			{backing: files.NewTrack(files.NewAlbum("album1", files.NewArtist("artist1", "./artist1"), "./artist1/album1"), "03 track3.mp3", "track3", 3)},
		})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sort.Sort(tt.input)
			for i := range tt.input {
				if i == 0 {
					continue
				}
				if tt.input[i-1].backing.Number() > tt.input[i].backing.Number() {
					t.Errorf("%s track[%d] with number %d comes before track[%d] with number %d", fnName, i-1, tt.input[i-1].backing.Number(), i, tt.input[i].backing.Number())
				}
			}
		})
	}
}

func Test_reportResults(t *testing.T) {
	const fnName = "reportResults()"
	type args struct {
		artistsWithIssues [][]*checkedArtist
	}
	tests := []struct {
		name string
		args
		output.WantedRecording
	}{
		{name: "degenerate case", args: args{}},
		{
			name: "more interesting case",
			args: args{artistsWithIssues: [][]*checkedArtist{
				// set 1
				{
					{
						backing: files.NewArtist("artist1", "artist1"),
						issues:  []string{"bad artist"},
						albums: []*checkedAlbum{
							{
								backing: files.NewAlbum("album1", nil, "artist1/album1"),
								issues:  []string{"skips badly"},
								tracks: []*checkedTrack{
									{
										backing: files.NewTrack(files.NewAlbum("album1", files.NewArtist("artist1", "./artist1"), "./artist1/album1"), "01 track1.mp3", "track1", 1),
										issues:  []string{"inaudible"},
									},
								},
							},
						},
					},
				},
				// set 2
				{
					{
						backing: files.NewArtist("artist1", "artist1"),
						issues:  []string{"really awful artist"},
						albums: []*checkedAlbum{
							{
								backing: files.NewAlbum("album1", nil, "artist1/album1"),
								issues:  []string{"bad cover art"},
								tracks: []*checkedTrack{
									{
										backing: files.NewTrack(files.NewAlbum("album1", files.NewArtist("artist1", "./artist1"), "./artist1/album1"), "01 track1.mp3", "track1", 1),
										issues:  []string{"plays backwards!"},
									},
									{
										backing: files.NewTrack(files.NewAlbum("album1", files.NewArtist("artist1", "./artist1"), "./artist1/album1"), "02 track2.mp3", "track2", 2),
										issues:  []string{"truly insipid"},
									},
								},
							},
							{
								backing: files.NewAlbum("album2", nil, "artist1/album2"),
								issues:  []string{"horrible sequel"},
								tracks: []*checkedTrack{
									{
										backing: files.NewTrack(files.NewAlbum("album2", files.NewArtist("artist1", "./artist1"), "./artist1/album2"), "03 track3.mp3", "track3", 3),
										issues:  []string{"singer is dreadful, band is worse"},
									},
								},
							},
						},
					},
					{
						backing: files.NewArtist("artist2", "artist2"),
						issues:  []string{"tone deaf"},
						albums: []*checkedAlbum{
							{
								backing: files.NewAlbum("album34", nil, "artist2/album34"),
								issues:  []string{"worst album I own"},
								tracks: []*checkedTrack{
									{
										backing: files.NewTrack(files.NewAlbum("album34", files.NewArtist("artist2", "./artist2"), "./artist2/album34"), "40 track40.mp3", "track40", 40),
										issues:  []string{"singer died in mid act and that improved the track"},
									},
								},
							},
						},
					},
				},
			}},
			WantedRecording: output.WantedRecording{
				Console: strings.Join([]string{
					"artist1",
					"  bad artist",
					"  really awful artist",
					"    album1",
					"      bad cover art",
					"      skips badly",
					"         1 track1",
					"          inaudible",
					"          plays backwards!",
					"         2 track2",
					"          truly insipid",
					"    album2",
					"      horrible sequel",
					"         3 track3",
					"          singer is dreadful, band is worse",
					"artist2",
					"  tone deaf",
					"    album34",
					"      worst album I own",
					"        40 track40",
					"          singer died in mid act and that improved the track",
					"",
				}, "\n"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			reportResults(o, tt.args.artistsWithIssues...)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}
