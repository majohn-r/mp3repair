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
)

var (
	fFlag bool = false
	tFlag bool = true
)

func Test_performEmptyFolderAnalysis(t *testing.T) {
	fnName := "performEmptyFolderAnalysis()"
	emptyDirName := "empty"
	dirtyDirName := "dirty"
	goodFolderDirName := "good"
	if err := internal.Mkdir(emptyDirName); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, emptyDirName, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, emptyDirName)
		internal.DestroyDirectoryForTesting(fnName, dirtyDirName)
		internal.DestroyDirectoryForTesting(fnName, goodFolderDirName)
	}()
	if err := internal.Mkdir(dirtyDirName); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, dirtyDirName, err)
	}
	if err := internal.PopulateTopDirForTesting(dirtyDirName); err != nil {
		t.Errorf("%s error populating %q: %v", fnName, dirtyDirName, err)
	}
	if err := internal.Mkdir(goodFolderDirName); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, goodFolderDirName, err)
	}
	if err := internal.Mkdir(filepath.Join(goodFolderDirName, "goodArtist")); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, "goodArtist", err)
	}
	if err := internal.Mkdir(filepath.Join(goodFolderDirName, "goodArtist", "goodAlbum")); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, "good album", err)
	}
	if err := internal.CreateFileForTestingWithContent(filepath.Join(goodFolderDirName, "goodArtist", "goodAlbum"), "01 goodTrack.mp3", []byte("good content")); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, "01 goodTrack.mp3", err)
	}
	goodArtist := files.NewArtist("goodArtist", filepath.Join(goodFolderDirName, "goodArtist"))
	goodAlbum := files.NewAlbum("goodAlbum", goodArtist, filepath.Join(goodFolderDirName, "goodArtist", "goodAlbum"))
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
		wantFilteredArtists []*artistWithIssues
		wantOk              bool
		internal.WantedOutput
	}{
		{name: "no work to do", c: &check{checkEmptyFolders: &fFlag}, args: args{}, wantOk: true},
		{
			name: "empty topDir",
			c:    &check{checkEmptyFolders: &tFlag},
			args: args{s: files.CreateSearchForTesting(emptyDirName)},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "No music files could be found using the specified parameters.\n",
				WantLogOutput: "level='info' -ext='.mp3' -topDir='empty' msg='reading unfiltered music files'\n" +
					"level='error' -ext='.mp3' -topDir='empty' msg='cannot find any artist directories'\n",
			},
		},
		{
			name:        "folders, no empty folders present",
			c:           &check{checkEmptyFolders: &tFlag},
			args:        args{s: files.CreateSearchForTesting(goodFolderDirName)},
			wantArtists: []*files.Artist{goodArtist},
			wantOk:      true,
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: "Empty Folder Analysis: no empty folders found.\n",
				WantLogOutput:     "level='info' -ext='.mp3' -topDir='good' msg='reading unfiltered music files'\n",
			},
		},
		{
			name:        "empty folders present",
			c:           &check{checkEmptyFolders: &tFlag},
			args:        args{s: files.CreateSearchForTesting(dirtyDirName)},
			wantArtists: files.CreateAllArtistsForTesting(dirtyDirName, true),
			wantOk:      true,
			wantFilteredArtists: []*artistWithIssues{
				{
					name:   "Test Artist 0",
					albums: []*albumWithIssues{{name: "Test Album 999", issues: []string{"no tracks found"}}},
				},
				{
					name:   "Test Artist 1",
					albums: []*albumWithIssues{{name: "Test Album 999", issues: []string{"no tracks found"}}},
				},
				{
					name:   "Test Artist 2",
					albums: []*albumWithIssues{{name: "Test Album 999", issues: []string{"no tracks found"}}},
				},
				{
					name:   "Test Artist 3",
					albums: []*albumWithIssues{{name: "Test Album 999", issues: []string{"no tracks found"}}},
				},
				{
					name:   "Test Artist 4",
					albums: []*albumWithIssues{{name: "Test Album 999", issues: []string{"no tracks found"}}},
				},
				{
					name:   "Test Artist 5",
					albums: []*albumWithIssues{{name: "Test Album 999", issues: []string{"no tracks found"}}},
				},
				{
					name:   "Test Artist 6",
					albums: []*albumWithIssues{{name: "Test Album 999", issues: []string{"no tracks found"}}},
				},
				{
					name:   "Test Artist 7",
					albums: []*albumWithIssues{{name: "Test Album 999", issues: []string{"no tracks found"}}},
				},
				{
					name:   "Test Artist 8",
					albums: []*albumWithIssues{{name: "Test Album 999", issues: []string{"no tracks found"}}},
				},
				{
					name:   "Test Artist 9",
					albums: []*albumWithIssues{{name: "Test Album 999", issues: []string{"no tracks found"}}},
				},
				{
					name:   "Test Artist 999",
					issues: []string{"no albums found"},
				},
			},
			WantedOutput: internal.WantedOutput{
				WantLogOutput: "level='info' -ext='.mp3' -topDir='dirty' msg='reading unfiltered music files'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			gotArtists, gotArtistsWithIssues, gotOk := tt.c.performEmptyFolderAnalysis(o, tt.args.s)
			if !reflect.DeepEqual(gotArtists, tt.wantArtists) {
				t.Errorf("%s = %v, want %v", fnName, gotArtists, tt.wantArtists)
			} else {
				filteredArtists := filterAndSortArtists(gotArtistsWithIssues)
				if !reflect.DeepEqual(filteredArtists, tt.wantFilteredArtists) {
					t.Errorf("%s = %v, want %v", fnName, filteredArtists, tt.wantFilteredArtists)
				}
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s ok = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_filterArtists(t *testing.T) {
	fnName := "filterArtists()"
	topDirName := "filterArtists"
	if err := internal.Mkdir(topDirName); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, topDirName, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDirName)
	}()
	if err := internal.PopulateTopDirForTesting(topDirName); err != nil {
		t.Errorf("%s error populating %q: %v", fnName, topDirName, err)
	}
	searchStruct := files.CreateSearchForTesting(topDirName)
	fullArtists, _ := searchStruct.LoadUnfilteredData(internal.NewOutputDeviceForTesting())
	filteredArtists, _ := searchStruct.LoadData(internal.NewOutputDeviceForTesting())
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
		internal.WantedOutput
	}{
		{
			name:   "neither gap analysis nor integrity enabled",
			c:      &check{checkGapsInTrackNumbering: &fFlag, checkIntegrity: &fFlag},
			args:   args{s: nil, artists: nil},
			wantOk: true,
		},
		{
			name:                "only gap analysis enabled, no artists supplied",
			c:                   &check{checkGapsInTrackNumbering: &tFlag, checkIntegrity: &fFlag},
			args:                args{s: searchStruct},
			wantFilteredArtists: filteredArtists,
			wantOk:              true,
			WantedOutput: internal.WantedOutput{
				WantLogOutput: "level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='filterArtists' msg='reading filtered music files'\n",
			},
		},
		{
			name:                "only gap analysis enabled, artists supplied",
			c:                   &check{checkGapsInTrackNumbering: &tFlag, checkIntegrity: &fFlag},
			args:                args{s: searchStruct, artists: fullArtists},
			wantFilteredArtists: filteredArtists,
			wantOk:              true,
			WantedOutput: internal.WantedOutput{
				WantLogOutput: "level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='filterArtists' msg='filtering music files'\n",
			},
		},
		{
			name:                "only integrity check enabled, no artists supplied",
			c:                   &check{checkGapsInTrackNumbering: &fFlag, checkIntegrity: &tFlag},
			args:                args{s: searchStruct},
			wantFilteredArtists: filteredArtists,
			wantOk:              true,
			WantedOutput: internal.WantedOutput{
				WantLogOutput: "level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='filterArtists' msg='reading filtered music files'\n",
			},
		},
		{
			name:                "only integrity check enabled, artists supplied",
			c:                   &check{checkGapsInTrackNumbering: &fFlag, checkIntegrity: &tFlag},
			args:                args{s: searchStruct, artists: fullArtists},
			wantFilteredArtists: filteredArtists,
			wantOk:              true,
			WantedOutput: internal.WantedOutput{
				WantLogOutput: "level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='filterArtists' msg='filtering music files'\n",
			},
		},
		{
			name:                "gap analysis and integrity check enabled, no artists supplied",
			c:                   &check{checkGapsInTrackNumbering: &tFlag, checkIntegrity: &tFlag},
			args:                args{s: searchStruct},
			wantFilteredArtists: filteredArtists,
			wantOk:              true,
			WantedOutput: internal.WantedOutput{
				WantLogOutput: "level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='filterArtists' msg='reading filtered music files'\n",
			},
		},
		{
			name:                "gap analysis and integrity check enabled, artists supplied",
			c:                   &check{checkGapsInTrackNumbering: &tFlag, checkIntegrity: &tFlag},
			args:                args{s: searchStruct, artists: fullArtists},
			wantFilteredArtists: filteredArtists,
			wantOk:              true,
			WantedOutput: internal.WantedOutput{
				WantLogOutput: "level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='filterArtists' msg='filtering music files'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			gotFilteredArtists, gotOk := tt.c.filterArtists(o, tt.args.s, tt.args.artists)
			if !reflect.DeepEqual(gotFilteredArtists, tt.wantFilteredArtists) {
				t.Errorf("%s = %v, want %v", fnName, gotFilteredArtists, tt.wantFilteredArtists)
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s ok = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_check_performGapAnalysis(t *testing.T) {
	fnName := "check.performGapAnalysis()"
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
		wantConflictedArtists []*artistWithIssues
		internal.WantedOutput
	}{
		{name: "no analysis", c: &check{checkGapsInTrackNumbering: &fFlag}, args: args{}},
		{
			name: "no content",
			c:    &check{checkGapsInTrackNumbering: &tFlag},
			args: args{},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: "Check Gaps: no gaps found.\n",
			},
		},
		{
			name: "good artist",
			c:    &check{checkGapsInTrackNumbering: &tFlag},
			args: args{artists: []*files.Artist{goodArtist}},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: "Check Gaps: no gaps found.\n",
			},
		},
		{
			name: "bad artist",
			c:    &check{checkGapsInTrackNumbering: &tFlag},
			args: args{artists: []*files.Artist{badArtist}},
			wantConflictedArtists: []*artistWithIssues{
				{
					name: "BadArtist",
					albums: []*albumWithIssues{
						{
							name: "No Biscuits For You!",
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
			o := internal.NewOutputDeviceForTesting()
			if gotConflictedArtists := filterAndSortArtists(tt.c.performGapAnalysis(o, tt.args.artists)); !reflect.DeepEqual(gotConflictedArtists, tt.wantConflictedArtists) {
				t.Errorf("%s = %v, want %v", fnName, gotConflictedArtists, tt.wantConflictedArtists)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_check_performIntegrityCheck(t *testing.T) {
	fnName := "check.performIntegrityCheck()"
	// create some data to work with
	topDirName := "integrity"
	if err := internal.Mkdir(topDirName); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, topDirName, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDirName)
	}()
	// keep it simple: one artist, one album, one track
	artistPath := filepath.Join(topDirName, "artist")
	if err := internal.Mkdir(artistPath); err != nil {
		t.Errorf("%s error creating artist folder", fnName)
	}
	albumPath := filepath.Join(artistPath, "album")
	if err := internal.Mkdir(albumPath); err != nil {
		t.Errorf("%s error creating album folder", fnName)
	}
	if err := internal.CreateFileForTestingWithContent(albumPath, "01 track.mp3", nil); err != nil {
		t.Errorf("%s error creating track", fnName)
	}
	s := files.CreateSearchForTesting(topDirName)
	a, _ := s.LoadUnfilteredData(internal.NewOutputDeviceForTesting())
	type args struct {
		artists []*files.Artist
	}
	tests := []struct {
		name string
		c    *check
		args
		wantConflictedArtists []*artistWithIssues
		internal.WantedOutput
	}{
		{name: "degenerate case", c: &check{checkIntegrity: &fFlag}, args: args{}},
		{
			name: "no artists",
			c:    &check{checkIntegrity: &tFlag},
			args: args{},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: "Integrity Analysis: no issues found.\n"},
		},
		{
			name: "meaningful case",
			c:    &check{checkIntegrity: &tFlag},
			args: args{artists: a},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "An error occurred when trying to read ID3V2 tag information for track \"track\" on album \"album\" by artist \"artist\": \"zero length\".\n",
				WantLogOutput:   "level='error' albumName='album' artistName='artist' error='zero length' trackName='track' msg='id3v2 tag error'\n",
			},
			wantConflictedArtists: []*artistWithIssues{
				{
					name: "artist",
					albums: []*albumWithIssues{
						{
							name: "album",
							tracks: []*trackWithIssues{
								{
									name:   "track",
									number: 1,
									issues: []string{"differences cannot be determined: there was an error reading ID3V2 tags"},
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
			o := internal.NewOutputDeviceForTesting()
			if gotConflictedArtists := filterAndSortArtists(tt.c.performIntegrityCheck(o, tt.args.artists)); !reflect.DeepEqual(gotConflictedArtists, tt.wantConflictedArtists) {
				t.Errorf("%s = %v, want %v", fnName, gotConflictedArtists, tt.wantConflictedArtists)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func makeCheckCommand() *check {
	c, _ := newCheckCommand(internal.NewOutputDeviceForTesting(), internal.EmptyConfiguration(), flag.NewFlagSet("check", flag.ContinueOnError))
	return c
}

func Test_check_Exec(t *testing.T) {
	fnName := "check.Exec()"
	topDirName := "checkExec"
	if err := internal.Mkdir(topDirName); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, topDirName, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDirName)
	}()
	if err := internal.PopulateTopDirForTesting(topDirName); err != nil {
		t.Errorf("%s error populating directory %q: %v", fnName, topDirName, err)
	}
	type args struct {
		args []string
	}
	tests := []struct {
		name string
		c    *check
		args
		wantOk bool
		internal.WantedOutput
	}{
		{
			name: "do nothing",
			c:    makeCheckCommand(),
			args: args{[]string{"-topDir", topDirName, "-empty=false", "-gaps=false", "-integrity=false"}},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "You disabled all functionality for the command \"check\".\n",
				WantLogOutput:   "level='error' -empty='false' -gaps='false' -integrity='false' command='check' msg='the user disabled all functionality'\n",
			},
		},
		{
			name:   "do something",
			c:      makeCheckCommand(),
			args:   args{[]string{"-topDir", topDirName, "-empty=true", "-gaps=false", "-integrity=false"}},
			wantOk: true,
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: strings.Join([]string{
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
				WantLogOutput: "level='info' -empty='true' -gaps='false' -integrity='false' command='check' msg='executing command'\n" +
					"level='info' -ext='.mp3' -topDir='checkExec' msg='reading unfiltered music files'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			if gotOk := tt.c.Exec(o, tt.args.args); gotOk != tt.wantOk {
				t.Errorf("%s ok = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_newCheckCommand(t *testing.T) {
	fnName := "newCheckCommand()"
	savedState := internal.SaveEnvVarForTesting("APPDATA")
	os.Setenv("APPDATA", internal.SecureAbsolutePathForTesting("."))
	defer func() {
		savedState.RestoreForTesting()
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
	defaultConfig, _ := internal.ReadConfigurationFile(internal.NewOutputDeviceForTesting())
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
		internal.WantedOutput
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
			args:                     args{c: defaultConfig},
			wantEmptyFolders:         true,
			wantGapsInTrackNumbering: true,
			wantIntegrity:            false,
			wantOk:                   true,
		},
		{
			name: "bad default empty folder",
			args: args{
				c: internal.CreateConfiguration(internal.NewOutputDeviceForTesting(), map[string]interface{}{
					"check": map[string]interface{}{
						emptyFoldersFlag: "Empty!!",
					},
				}),
			},
			wantOk: false,
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The configuration file \"defaults.yaml\" contains an invalid value for \"check\": invalid boolean value \"Empty!!\" for -empty: parse error.\n",
				WantLogOutput:   "level='error' error='invalid boolean value \"Empty!!\" for -empty: parse error' section='check' msg='invalid content in configuration file'\n",
			},
		},
		{
			name: "bad default gaps",
			args: args{
				c: internal.CreateConfiguration(internal.NewOutputDeviceForTesting(), map[string]interface{}{
					"check": map[string]interface{}{
						gapsInTrackNumberingFlag: "No",
					},
				}),
			},
			wantOk: false,
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The configuration file \"defaults.yaml\" contains an invalid value for \"check\": invalid boolean value \"No\" for -gaps: parse error.\n",
				WantLogOutput:   "level='error' error='invalid boolean value \"No\" for -gaps: parse error' section='check' msg='invalid content in configuration file'\n",
			},
		},
		{
			name: "bad default integrity",
			args: args{
				c: internal.CreateConfiguration(internal.NewOutputDeviceForTesting(), map[string]interface{}{
					"check": map[string]interface{}{
						integrityFlag: "Off",
					},
				}),
			},
			wantOk: false,
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The configuration file \"defaults.yaml\" contains an invalid value for \"check\": invalid boolean value \"Off\" for -integrity: parse error.\n",
				WantLogOutput:   "level='error' error='invalid boolean value \"Off\" for -integrity: parse error' section='check' msg='invalid content in configuration file'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			check, gotOk := newCheckCommand(o, tt.args.c, flag.NewFlagSet("check", flag.ContinueOnError))
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk %t wantOk %t", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
			if check != nil {
				if _, ok := check.sf.ProcessArgs(internal.NewOutputDeviceForTesting(), []string{
					"-topDir", topDir,
					"-ext", ".mp3",
				}); ok {
					if *check.checkEmptyFolders != tt.wantEmptyFolders {
						t.Errorf("%s %q: got checkEmptyFolders %t want %t", fnName, tt.name, *check.checkEmptyFolders, tt.wantEmptyFolders)
					}
					if *check.checkGapsInTrackNumbering != tt.wantGapsInTrackNumbering {
						t.Errorf("%s %q: got checkGapsInTrackNumbering %t want %t", fnName, tt.name, *check.checkGapsInTrackNumbering, tt.wantGapsInTrackNumbering)
					}
					if *check.checkIntegrity != tt.wantIntegrity {
						t.Errorf("%s %q: got checkIntegrity %t want %t", fnName, tt.name, *check.checkIntegrity, tt.wantIntegrity)
					}
				} else {
					t.Errorf("%s %q: error processing arguments", fnName, tt.name)
				}
			}
		})
	}
}

func Test_merge(t *testing.T) {
	fnName := "merge()"
	type args struct {
		sets [][]*artistWithIssues
	}
	tests := []struct {
		name string
		args
		want []*artistWithIssues
	}{
		{name: "degenerate case", args: args{}},
		{
			name: "more interesting case",
			args: args{sets: [][]*artistWithIssues{
				// set 1
				{
					{
						name:   "artist1",
						issues: []string{"bad artist"},
						albums: []*albumWithIssues{
							{
								name:   "album1",
								issues: []string{"skips badly"},
								tracks: []*trackWithIssues{
									{
										number: 1,
										name:   "track1",
										issues: []string{"inaudible"},
									},
								},
							},
						},
					},
				},
				// set 2
				{
					{
						name:   "artist1",
						issues: []string{"really awful artist"},
						albums: []*albumWithIssues{
							{
								name:   "album1",
								issues: []string{"bad cover art"},
								tracks: []*trackWithIssues{
									{
										number: 1,
										name:   "track1",
										issues: []string{"plays backwards!"},
									},
									{
										number: 2,
										name:   "track2",
										issues: []string{"truly insipid"},
									},
								},
							},
							{
								name:   "album2",
								issues: []string{"horrible sequel"},
								tracks: []*trackWithIssues{
									{
										number: 3,
										name:   "track3",
										issues: []string{"singer is dreadful, band is worse"},
									},
								},
							},
						},
					},
					{
						name:   "artist2",
						issues: []string{"tone deaf"},
						albums: []*albumWithIssues{
							{
								name:   "album34",
								issues: []string{"worst album I own"},
								tracks: []*trackWithIssues{
									{
										number: 40,
										name:   "track40",
										issues: []string{"singer died in mid act and that improved the track"},
									},
								},
							},
						},
					},
				},
			}},
			want: []*artistWithIssues{
				{
					name:   "artist1",
					issues: []string{"bad artist", "really awful artist"},
					albums: []*albumWithIssues{
						{
							name:   "album1",
							issues: []string{"bad cover art", "skips badly"},
							tracks: []*trackWithIssues{
								{
									number: 1,
									name:   "track1",
									issues: []string{"inaudible", "plays backwards!"},
								},
								{
									number: 2,
									name:   "track2",
									issues: []string{"truly insipid"},
								},
							},
						},
						{
							name:   "album2",
							issues: []string{"horrible sequel"},
							tracks: []*trackWithIssues{
								{
									number: 3,
									name:   "track3",
									issues: []string{"singer is dreadful, band is worse"},
								},
							},
						},
					},
				},
				{
					name:   "artist2",
					issues: []string{"tone deaf"},
					albums: []*albumWithIssues{
						{
							name:   "album34",
							issues: []string{"worst album I own"},
							tracks: []*trackWithIssues{
								{
									number: 40,
									name:   "track40",
									issues: []string{"singer died in mid act and that improved the track"},
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
					if gotArtist.name != wantArtist.name {
						t.Errorf("%s artist[%d] name %q, want %q", fnName, i, gotArtist.name, wantArtist.name)
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
							if gotAlbum.name != wantAlbum.name {
								t.Errorf("%s artist[%d] album[%d] name %q, want %q", fnName, i, j, gotAlbum.name, wantAlbum.name)
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
									if gotTrack.number != wantTrack.number {
										t.Errorf("%s artist[%d] album[%d] track[%d] number %d, want %d", fnName, i, j, k, gotTrack.number, wantTrack.number)
									}
									if gotTrack.name != wantTrack.name {
										t.Errorf("%s artist[%d] album[%d] track[%d] name %q, want %q", fnName, i, j, k, gotTrack.name, wantTrack.name)
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

func Test_sortArtistsWithIssues(t *testing.T) {
	fnName := "sortArtistsWithIssues()"
	tests := []struct {
		name  string
		input artistSlice
	}{
		{name: "degenerate case", input: nil},
		{name: "scrambled input", input: artistSlice([]*artistWithIssues{
			{name: "10"},
			{name: "2"},
			{name: "35"},
			{name: "1"},
			{name: "3"},
		})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sort.Sort(tt.input)
			for i := range tt.input {
				if i == 0 {
					continue
				}
				if tt.input[i-1].name > tt.input[i].name {
					t.Errorf("%s artist[%d] with name %q comes before artist[%d] with name %q", fnName, i-1, tt.input[i-1].name, i, tt.input[i].name)
				}
			}
		})
	}
}

func Test_sortAlbumsWithIssues(t *testing.T) {
	fnName := "sortAlbumsWithIssues()"
	tests := []struct {
		name  string
		input albumSlice
	}{
		{name: "degenerate case", input: nil},
		{name: "scrambled input", input: albumSlice([]*albumWithIssues{
			{name: "10"},
			{name: "2"},
			{name: "35"},
			{name: "1"},
			{name: "3"},
		})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sort.Sort(tt.input)
			for i := range tt.input {
				if i == 0 {
					continue
				}
				if tt.input[i-1].name > tt.input[i].name {
					t.Errorf("%s album[%d] with name %q comes before album[%d] with name %q", fnName, i-1, tt.input[i-1].name, i, tt.input[i].name)
				}
			}
		})
	}
}

func Test_sortTracksWithIssues(t *testing.T) {
	fnName := "sortTracksWithIssues()"
	tests := []struct {
		name  string
		input trackSlice
	}{
		{name: "degenerate case", input: nil},
		{name: "scrambled input", input: trackSlice([]*trackWithIssues{
			{number: 10},
			{number: 2},
			{number: 35},
			{number: 1},
			{number: 3},
		})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sort.Sort(tt.input)
			for i := range tt.input {
				if i == 0 {
					continue
				}
				if tt.input[i-1].number > tt.input[i].number {
					t.Errorf("%s track[%d] with number %d comes before track[%d] with number %d", fnName, i-1, tt.input[i-1].number, i, tt.input[i].number)
				}
			}
		})
	}
}

func Test_reportResults(t *testing.T) {
	fnName := "reportResults()"
	type args struct {
		artistsWithIssues [][]*artistWithIssues
	}
	tests := []struct {
		name string
		args
		internal.WantedOutput
	}{
		{name: "degenerate case", args: args{}},
		{
			name: "more interesting case",
			args: args{artistsWithIssues: [][]*artistWithIssues{
				// set 1
				{
					{
						name:   "artist1",
						issues: []string{"bad artist"},
						albums: []*albumWithIssues{
							{
								name:   "album1",
								issues: []string{"skips badly"},
								tracks: []*trackWithIssues{
									{
										number: 1,
										name:   "track1",
										issues: []string{"inaudible"},
									},
								},
							},
						},
					},
				},
				// set 2
				{
					{
						name:   "artist1",
						issues: []string{"really awful artist"},
						albums: []*albumWithIssues{
							{
								name:   "album1",
								issues: []string{"bad cover art"},
								tracks: []*trackWithIssues{
									{
										number: 1,
										name:   "track1",
										issues: []string{"plays backwards!"},
									},
									{
										number: 2,
										name:   "track2",
										issues: []string{"truly insipid"},
									},
								},
							},
							{
								name:   "album2",
								issues: []string{"horrible sequel"},
								tracks: []*trackWithIssues{
									{
										number: 3,
										name:   "track3",
										issues: []string{"singer is dreadful, band is worse"},
									},
								},
							},
						},
					},
					{
						name:   "artist2",
						issues: []string{"tone deaf"},
						albums: []*albumWithIssues{
							{
								name:   "album34",
								issues: []string{"worst album I own"},
								tracks: []*trackWithIssues{
									{
										number: 40,
										name:   "track40",
										issues: []string{"singer died in mid act and that improved the track"},
									},
								},
							},
						},
					},
				},
			}},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: strings.Join([]string{
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
			o := internal.NewOutputDeviceForTesting()
			reportResults(o, tt.args.artistsWithIssues...)
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}
