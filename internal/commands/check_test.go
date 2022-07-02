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
		t.Errorf("%s error creating %s: %v", fnName, dirtyDirName, err)
	}
	if err := internal.PopulateTopDirForTesting(dirtyDirName); err != nil {
		t.Errorf("%s error populating %s: %v", fnName, dirtyDirName, err)
	}
	if err := internal.Mkdir(goodFolderDirName); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, goodFolderDirName, err)
	}
	if err := internal.Mkdir(filepath.Join(goodFolderDirName, "goodArtist")); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, "goodArtist", err)
	}
	if err := internal.Mkdir(filepath.Join(goodFolderDirName, "goodArtist", "goodAlbum")); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, "good album", err)
	}
	if err := internal.CreateFileForTestingWithContent(filepath.Join(goodFolderDirName, "goodArtist", "goodAlbum"), "01 goodTrack.mp3", "good content"); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, "01 goodTrack.mp3", err)
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
		name                string
		c                   *check
		args                args
		wantArtists         []*files.Artist
		wantConsoleOutput   string
		wantErrorOutput     string
		wantLogOutput       string
		wantFilteredArtists []*artistWithIssues
		wantOk              bool
	}{
		{name: "no work to do", c: &check{checkEmptyFolders: &fFlag}, args: args{}, wantOk: true},
		{
			name:            "empty topDir",
			c:               &check{checkEmptyFolders: &tFlag},
			args:            args{s: files.CreateSearchForTesting(emptyDirName)},
			wantErrorOutput: "No music files could be found using the specified parameters.\n",
			wantLogOutput:   "level='warn' -ext='.mp3' -topDir='empty' msg='cannot find any artist directories'\n",
		},
		{
			name:                "folders, no empty folders present",
			c:                   &check{checkEmptyFolders: &tFlag},
			args:                args{s: files.CreateSearchForTesting(goodFolderDirName)},
			wantArtists:         []*files.Artist{goodArtist},
			wantFilteredArtists: nil,
			wantConsoleOutput:   "Empty Folder Analysis: no empty folders found\n",
			wantOk:              true,
		},
		{
			name:        "empty folders present",
			c:           &check{checkEmptyFolders: &tFlag},
			args:        args{s: files.CreateSearchForTesting(dirtyDirName)},
			wantArtists: files.CreateAllArtistsForTesting(dirtyDirName, true),
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
			wantOk: true,
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
			if gotConsoleOutput := o.ConsoleOutput(); gotConsoleOutput != tt.wantConsoleOutput {
				t.Errorf("%s console output = %q, want %q", fnName, gotConsoleOutput, tt.wantConsoleOutput)
			}
			if gotErrorOutput := o.ErrorOutput(); gotErrorOutput != tt.wantErrorOutput {
				t.Errorf("%s error output = %q, want %q", fnName, gotErrorOutput, tt.wantErrorOutput)
			}
			if gotLogOutput := o.LogOutput(); gotLogOutput != tt.wantLogOutput {
				t.Errorf("%s log output = %q, want %q", fnName, gotLogOutput, tt.wantLogOutput)
			}
		})
	}
}

func Test_filterArtists(t *testing.T) {
	fnName := "filterArtists()"
	topDirName := "filterArtists"
	if err := internal.Mkdir(topDirName); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, topDirName, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDirName)
	}()
	if err := internal.PopulateTopDirForTesting(topDirName); err != nil {
		t.Errorf("%s error populating %s: %v", fnName, topDirName, err)
	}
	searchStruct := files.CreateSearchForTesting(topDirName)
	fullArtists, _ := searchStruct.LoadUnfilteredData(internal.NewOutputDeviceForTesting())
	filteredArtists, _ := searchStruct.LoadData(internal.NewOutputDeviceForTesting())
	type args struct {
		s       *files.Search
		artists []*files.Artist
	}
	tests := []struct {
		name                string
		c                   *check
		args                args
		wantFilteredArtists []*files.Artist
		wantOk              bool
		wantConsoleOutput   string
		wantErrorOutput     string
		wantLogOutput       string
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
		},
		{
			name:                "only gap analysis enabled, artists supplied",
			c:                   &check{checkGapsInTrackNumbering: &tFlag, checkIntegrity: &fFlag},
			args:                args{s: searchStruct, artists: fullArtists},
			wantFilteredArtists: filteredArtists,
			wantOk:              true,
		},
		{
			name:                "only integrity check enabled, no artists supplied",
			c:                   &check{checkGapsInTrackNumbering: &fFlag, checkIntegrity: &tFlag},
			args:                args{s: searchStruct},
			wantFilteredArtists: filteredArtists,
			wantOk:              true,
		},
		{
			name:                "only integrity check enabled, artists supplied",
			c:                   &check{checkGapsInTrackNumbering: &fFlag, checkIntegrity: &tFlag},
			args:                args{s: searchStruct, artists: fullArtists},
			wantFilteredArtists: filteredArtists,
			wantOk:              true,
		},
		{
			name:                "gap analysis and integrity check enabled, no artists supplied",
			c:                   &check{checkGapsInTrackNumbering: &tFlag, checkIntegrity: &tFlag},
			args:                args{s: searchStruct},
			wantFilteredArtists: filteredArtists,
			wantOk:              true,
		},
		{
			name:                "gap analysis and integrity check enabled, artists supplied",
			c:                   &check{checkGapsInTrackNumbering: &tFlag, checkIntegrity: &tFlag},
			args:                args{s: searchStruct, artists: fullArtists},
			wantFilteredArtists: filteredArtists,
			wantOk:              true,
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
			if gotConsoleOutput := o.ConsoleOutput(); gotConsoleOutput != tt.wantConsoleOutput {
				t.Errorf("%s console output = %v, want %v", fnName, gotConsoleOutput, tt.wantConsoleOutput)
			}
			if gotErrorOutput := o.ErrorOutput(); gotErrorOutput != tt.wantErrorOutput {
				t.Errorf("%s error output = %v, want %v", fnName, gotErrorOutput, tt.wantErrorOutput)
			}
			if gotLogOutput := o.LogOutput(); gotLogOutput != tt.wantLogOutput {
				t.Errorf("%s log output = %v, want %v", fnName, gotLogOutput, tt.wantLogOutput)
			}
		})
	}
}

func Test_check_performGapAnalysis(t *testing.T) {
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
		name                  string
		c                     *check
		args                  args
		wantConsoleOutput     string
		wantErrorOutput       string
		wantLogOutput         string
		wantConflictedArtists []*artistWithIssues
	}{
		{name: "no analysis", c: &check{checkGapsInTrackNumbering: &fFlag}, args: args{}, wantConsoleOutput: ""},
		{
			name:              "no content",
			c:                 &check{checkGapsInTrackNumbering: &tFlag},
			args:              args{},
			wantConsoleOutput: "Check Gaps: no gaps found\n",
		},
		{
			name:              "good artist",
			c:                 &check{checkGapsInTrackNumbering: &tFlag},
			args:              args{artists: []*files.Artist{goodArtist}},
			wantConsoleOutput: "Check Gaps: no gaps found\n",
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
			gotConflictedArtists := filterAndSortArtists(tt.c.performGapAnalysis(o, tt.args.artists))
			if gotConsoleOutput := o.ConsoleOutput(); gotConsoleOutput != tt.wantConsoleOutput {
				t.Errorf("check.performGapAnalysis() console output = %v, want %v", gotConsoleOutput, tt.wantConsoleOutput)
			}
			if gotErrorOutput := o.ErrorOutput(); gotErrorOutput != tt.wantErrorOutput {
				t.Errorf("check.performGapAnalysis() error output = %v, want %v", gotErrorOutput, tt.wantErrorOutput)
			}
			if gotLogOutput := o.LogOutput(); gotLogOutput != tt.wantLogOutput {
				t.Errorf("check.performGapAnalysis() log output = %v, want %v", gotLogOutput, tt.wantLogOutput)
			}
			if !reflect.DeepEqual(gotConflictedArtists, tt.wantConflictedArtists) {
				t.Errorf("check.performGapAnalysis() = %v, want %v", gotConflictedArtists, tt.wantConflictedArtists)
			}
		})
	}
}

func Test_check_performIntegrityCheck(t *testing.T) {
	fnName := "check.performIntegrityCheck()"
	// create some data to work with
	topDirName := "integrity"
	if err := internal.Mkdir(topDirName); err != nil {
		t.Errorf("cannot create %q: %v", topDirName, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDirName)
	}()
	// keep it simple: one artist, one album, one track
	artistPath := filepath.Join(topDirName, "artist")
	if err := internal.Mkdir(artistPath); err != nil {
		t.Errorf("error creating artist folder")
	}
	albumPath := filepath.Join(artistPath, "album")
	if err := internal.Mkdir(albumPath); err != nil {
		t.Errorf("error creating album folder")
	}
	if err := internal.CreateFileForTestingWithContent(albumPath, "01 track.mp3", ""); err != nil {
		t.Errorf("error creating track")
	}
	s := files.CreateSearchForTesting(topDirName)
	a, _ := s.LoadUnfilteredData(internal.NewOutputDeviceForTesting())
	type args struct {
		artists []*files.Artist
	}
	tests := []struct {
		name                  string
		c                     *check
		args                  args
		wantConsoleOutput     string
		wantErrorOutput       string
		wantLogOutput         string
		wantConflictedArtists []*artistWithIssues
	}{
		{name: "degenerate case", c: &check{checkIntegrity: &fFlag}, args: args{}},
		{name: "no artists", c: &check{checkIntegrity: &tFlag}, args: args{}, wantConsoleOutput: "Integrity Analysis: no issues found\n"},
		{
			name: "meaningful case",
			c:    &check{checkIntegrity: &tFlag},
			args: args{artists: a},
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
									issues: []string{"cannot determine differences, tags were not recognized"},
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
			gotConflictedArtists := filterAndSortArtists(tt.c.performIntegrityCheck(o, tt.args.artists))
			if gotConsoleOutput := o.ConsoleOutput(); gotConsoleOutput != tt.wantConsoleOutput {
				t.Errorf("%s console output = %v, want %v", fnName, gotConsoleOutput, tt.wantConsoleOutput)
			}
			if gotErrorOutput := o.ErrorOutput(); gotErrorOutput != tt.wantErrorOutput {
				t.Errorf("%s error output = %v, want %v", fnName, gotErrorOutput, tt.wantErrorOutput)
			}
			if gotLogOutput := o.LogOutput(); gotLogOutput != tt.wantLogOutput {
				t.Errorf("%s log output = %v, want %v", fnName, gotLogOutput, tt.wantLogOutput)
			}
			if !reflect.DeepEqual(gotConflictedArtists, tt.wantConflictedArtists) {
				t.Errorf("check.performGapAnalysis() = %v, want %v", gotConflictedArtists, tt.wantConflictedArtists)
			}
		})
	}
}

func Test_check_Exec(t *testing.T) {
	topDirName := "checkExec"
	if err := internal.Mkdir(topDirName); err != nil {
		t.Errorf("error creating directory %q", topDirName)
	}
	fnName := "check.Exec()"
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDirName)
	}()
	if err := internal.PopulateTopDirForTesting(topDirName); err != nil {
		t.Errorf("error populating directory %q", topDirName)
	}
	type args struct {
		args []string
	}
	tests := []struct {
		name              string
		c                 *check
		args              args
		wantOk            bool
		wantConsoleOutput string
		wantErrorOutput   string
		wantLogOutput     string
	}{
		{
			name:            "do nothing",
			c:               newCheckCommand(internal.EmptyConfiguration(), flag.NewFlagSet("check", flag.ContinueOnError)),
			args:            args{[]string{"-topDir", topDirName, "-empty=false", "-gaps=false", "-integrity=false"}},
			wantErrorOutput: "You disabled all functionality for the command \"check\".\n",
			wantLogOutput:   "level='warn' -empty='false' -gaps='false' -integrity='false' command='check' msg='the user disabled all functionality'\n",
		},
		{
			name: "do something",
			c:    newCheckCommand(internal.EmptyConfiguration(), flag.NewFlagSet("check", flag.ContinueOnError)),
			args: args{[]string{"-topDir", topDirName, "-empty=true", "-gaps=false", "-integrity=false"}},
			wantConsoleOutput: strings.Join([]string{
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
			wantLogOutput: "level='info' -empty='true' -gaps='false' -integrity='false' command='check' msg='executing command'\n",
			wantOk:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			if gotOk := tt.c.Exec(o, tt.args.args); gotOk != tt.wantOk {
				t.Errorf("%s ok = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if gotConsoleOutput := o.ConsoleOutput(); gotConsoleOutput != tt.wantConsoleOutput {
				t.Errorf("%s console output = %v, want %v", fnName, gotConsoleOutput, tt.wantConsoleOutput)
			}
			if gotErrorOutput := o.ErrorOutput(); gotErrorOutput != tt.wantErrorOutput {
				t.Errorf("%s error output = %v, want %v", fnName, gotErrorOutput, tt.wantErrorOutput)
			}
			if gotLogOutput := o.LogOutput(); gotLogOutput != tt.wantLogOutput {
				t.Errorf("%s log output = %v, want %v", fnName, gotLogOutput, tt.wantLogOutput)
			}
		})
	}
}

func Test_newCheckCommand(t *testing.T) {
	savedState := internal.SaveEnvVarForTesting("APPDATA")
	os.Setenv("APPDATA", internal.SecureAbsolutePathForTesting("."))
	defer func() {
		savedState.RestoreForTesting()
	}()
	topDir := "loadTest"
	fnName := "newCheckCommand()"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, topDir, err)
	}
	if err := internal.PopulateTopDirForTesting(topDir); err != nil {
		t.Errorf("%s error populating %s: %v", fnName, topDir, err)
	}
	if err := internal.CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("error creating defaults.yaml: %v", err)
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
		name                     string
		args                     args
		wantEmptyFolders         bool
		wantGapsInTrackNumbering bool
		wantIntegrity            bool
	}{
		{
			name:                     "ordinary defaults",
			args:                     args{c: internal.EmptyConfiguration()},
			wantEmptyFolders:         false,
			wantGapsInTrackNumbering: false,
			wantIntegrity:            true,
		},
		{
			name:                     "overridden defaults",
			args:                     args{c: defaultConfig},
			wantEmptyFolders:         true,
			wantGapsInTrackNumbering: true,
			wantIntegrity:            false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := newCheckCommand(tt.args.c, flag.NewFlagSet("check", flag.ContinueOnError))
			if _, ok := check.sf.ProcessArgs(internal.NewOutputDeviceForTesting(), []string{
				"-topDir", topDir,
				"-ext", ".mp3",
			}); ok {
				if *check.checkEmptyFolders != tt.wantEmptyFolders {
					t.Errorf("%s %s: got checkEmptyFolders %t want %t", fnName, tt.name, *check.checkEmptyFolders, tt.wantEmptyFolders)
				}
				if *check.checkGapsInTrackNumbering != tt.wantGapsInTrackNumbering {
					t.Errorf("%s %s: got checkGapsInTrackNumbering %t want %t", fnName, tt.name, *check.checkGapsInTrackNumbering, tt.wantGapsInTrackNumbering)
				}
				if *check.checkIntegrity != tt.wantIntegrity {
					t.Errorf("%s %s: got checkIntegrity %t want %t", fnName, tt.name, *check.checkIntegrity, tt.wantIntegrity)
				}
			} else {
				t.Errorf("%s %s: error processing arguments", fnName, tt.name)
			}
		})
	}
}

func Test_merge(t *testing.T) {
	type args struct {
		sets [][]*artistWithIssues
	}
	tests := []struct {
		name string
		args args
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
				t.Errorf("merge() artist len = %d, want %d", len(got), len(tt.want))
			} else {
				for i := range got {
					gotArtist := got[i]
					wantArtist := tt.want[i]
					if gotArtist.name != wantArtist.name {
						t.Errorf("merge() artist[%d] name %q, want %q", i, gotArtist.name, wantArtist.name)
					}
					if !reflect.DeepEqual(gotArtist.issues, wantArtist.issues) {
						t.Errorf("merge() artist[%d] issues %v, want %v", i, gotArtist.issues, wantArtist.issues)
					}
					if len(gotArtist.albums) != len(wantArtist.albums) {
						t.Errorf("merge() artist[%d] albums len = %d, want %d", i, len(gotArtist.albums), len(wantArtist.albums))
					} else {
						for j := range gotArtist.albums {
							gotAlbum := gotArtist.albums[j]
							wantAlbum := wantArtist.albums[j]
							if gotAlbum.name != wantAlbum.name {
								t.Errorf("merge() artist[%d] album[%d] name %q, want %q", i, j, gotAlbum.name, wantAlbum.name)
							}
							if !reflect.DeepEqual(gotAlbum.issues, wantAlbum.issues) {
								t.Errorf("merge() artist[%d] album[%d] issues %v, want %v", i, j, gotAlbum.issues, wantAlbum.issues)
							}
							if len(gotAlbum.tracks) != len(wantAlbum.tracks) {
								t.Errorf("merge() artist[%d] album[%d] tracks len = %d, want %d", i, j, len(gotAlbum.tracks), len(wantAlbum.tracks))
							} else {
								for k := range gotAlbum.tracks {
									gotTrack := gotAlbum.tracks[k]
									wantTrack := wantAlbum.tracks[k]
									if gotTrack.number != wantTrack.number {
										t.Errorf("merge() artist[%d] album[%d] track[%d] number %d, want %d", i, j, k, gotTrack.number, wantTrack.number)
									}
									if gotTrack.name != wantTrack.name {
										t.Errorf("merge() artist[%d] album[%d] track[%d] name %q, want %q", i, j, k, gotTrack.name, wantTrack.name)
									}
									if !reflect.DeepEqual(gotTrack.issues, wantTrack.issues) {
										t.Errorf("merge() artist[%d] album[%d] track[%d] issues %v, want %v", i, j, k, gotTrack.issues, wantTrack.issues)
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
					t.Errorf("sortArtistsWithIssues artist[%d] with name %q comes before artist[%d] with name %q", i-1, tt.input[i-1].name, i, tt.input[i].name)
				}
			}
		})
	}
}

func Test_sortAlbumsWithIssues(t *testing.T) {
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
					t.Errorf("sortAlbumsWithIssues album[%d] with name %q comes before album[%d] with name %q", i-1, tt.input[i-1].name, i, tt.input[i].name)
				}
			}
		})
	}
}

func Test_sortTracksWithIssues(t *testing.T) {
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
					t.Errorf("sortTracksWithIssues track[%d] with number %d comes before track[%d] with number %d", i-1, tt.input[i-1].number, i, tt.input[i].number)
				}
			}
		})
	}
}

func Test_reportResults(t *testing.T) {
	type args struct {
		artistsWithIssues [][]*artistWithIssues
	}
	tests := []struct {
		name              string
		args              args
		wantConsoleOutput string
		wantErrorOutput   string
		wantLogOutput     string
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
			wantConsoleOutput: strings.Join([]string{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			reportResults(o, tt.args.artistsWithIssues...)
			if gotConsoleOutput := o.ConsoleOutput(); gotConsoleOutput != tt.wantConsoleOutput {
				t.Errorf("reportResults() console output = %v, want %v", gotConsoleOutput, tt.wantConsoleOutput)
			}
			if gotErrorOutput := o.ErrorOutput(); gotErrorOutput != tt.wantErrorOutput {
				t.Errorf("reportResults() error output = %v, want %v", gotErrorOutput, tt.wantErrorOutput)
			}
			if gotLogOutput := o.LogOutput(); gotLogOutput != tt.wantLogOutput {
				t.Errorf("reportResults() log output = %v, want %v", gotLogOutput, tt.wantLogOutput)
			}
		})
	}
}
