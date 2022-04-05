package subcommands

import (
	"bytes"
	"mp3/internal"
	"mp3/internal/files"
	"reflect"
	"strings"
	"testing"
)

func Test_performEmptyFolderAnalysis(t *testing.T) {
	fnName := "performEmptyFolderAnalysis()"
	noCheck := false
	performCheck := true
	emptyDirName := "empty"
	if err := internal.Mkdir(emptyDirName); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, emptyDirName, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, emptyDirName)
	}()
	dirtyDirName := "dirty"
	if err := internal.Mkdir(dirtyDirName); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, dirtyDirName, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, dirtyDirName)
	}()
	if err := internal.PopulateTopDirForTesting(dirtyDirName); err != nil {
		t.Errorf("%s error populating %s: %v", fnName, dirtyDirName, err)
	}
	type args struct {
		c *check
		s *files.Search
	}
	tests := []struct {
		name        string
		args        args
		wantArtists []*files.Artist
		wantW       string
	}{
		{name: "no work to do", args: args{c: &check{checkEmptyFolders: &noCheck}}},
		{
			name:  "empty topDir",
			args:  args{c: &check{checkEmptyFolders: &performCheck}, s: files.CreateSearchForTesting(emptyDirName)},
			wantW: "Empty Folder Analysis: no empty folders found\n",
		},
		{
			name: "empty folders present",
			args: args{
				c: &check{checkEmptyFolders: &performCheck},
				s: files.CreateSearchForTesting(dirtyDirName),
			},
			wantW: strings.Join([]string{
				"Empty Folder Analysis",
				"Artist \"Test Artist 0\" album \"Test Album 999\": no tracks found",
				"Artist \"Test Artist 1\" album \"Test Album 999\": no tracks found",
				"Artist \"Test Artist 2\" album \"Test Album 999\": no tracks found",
				"Artist \"Test Artist 3\" album \"Test Album 999\": no tracks found",
				"Artist \"Test Artist 4\" album \"Test Album 999\": no tracks found",
				"Artist \"Test Artist 5\" album \"Test Album 999\": no tracks found",
				"Artist \"Test Artist 6\" album \"Test Album 999\": no tracks found",
				"Artist \"Test Artist 7\" album \"Test Album 999\": no tracks found",
				"Artist \"Test Artist 8\" album \"Test Album 999\": no tracks found",
				"Artist \"Test Artist 9\" album \"Test Album 999\": no tracks found",
				"Artist \"Test Artist 999\": no albums found",
				"",
			}, "\n"),
			wantArtists: files.CreateAllArtistsForTesting(dirtyDirName, true),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			if gotArtists := performEmptyFolderAnalysis(w, tt.args.c, tt.args.s); !reflect.DeepEqual(gotArtists, tt.wantArtists) {
				t.Errorf("%s = %v, want %v", fnName, gotArtists, tt.wantArtists)
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("%s = %v, want %v", fnName, gotW, tt.wantW)
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
	fullArtists := searchStruct.LoadUnfilteredData()
	filteredArtists := searchStruct.LoadData()
	fFlag := false
	tFlag := true
	type args struct {
		c       *check
		s       *files.Search
		artists []*files.Artist
	}
	tests := []struct {
		name                string
		args                args
		wantFilteredArtists []*files.Artist
	}{
		{
			name: "neither gap analysis nor integrity enabled",
			args: args{c: &check{checkGapsInTrackNumbering: &fFlag, checkIntegrity: &fFlag}},
		},
		{
			name: "only gap analysis enabled, no artists supplied",
			args: args{c: &check{
				checkGapsInTrackNumbering: &tFlag, checkIntegrity: &fFlag},
				s: searchStruct,
			},
			wantFilteredArtists: filteredArtists,
		},
		{
			name: "only gap analysis enabled, artists supplied",
			args: args{
				c:       &check{checkGapsInTrackNumbering: &tFlag, checkIntegrity: &fFlag},
				s:       searchStruct,
				artists: fullArtists,
			},
			wantFilteredArtists: filteredArtists,
		},
		{
			name: "only integrity check enabled, no artists supplied",
			args: args{
				c: &check{checkGapsInTrackNumbering: &fFlag, checkIntegrity: &tFlag},
				s: searchStruct,
			},
			wantFilteredArtists: filteredArtists,
		},
		{
			name: "only integrity check enabled, artists supplied",
			args: args{
				c:       &check{checkGapsInTrackNumbering: &fFlag, checkIntegrity: &tFlag},
				s:       searchStruct,
				artists: fullArtists,
			},
			wantFilteredArtists: filteredArtists,
		},
		{
			name: "gap analysis and integrity check enabled, no artists supplied",
			args: args{
				c: &check{checkGapsInTrackNumbering: &tFlag, checkIntegrity: &tFlag},
				s: searchStruct,
			},
			wantFilteredArtists: filteredArtists,
		},
		{
			name: "gap analysis and integrity check enabled, artists supplied",
			args: args{
				c:       &check{checkGapsInTrackNumbering: &tFlag, checkIntegrity: &tFlag},
				s:       searchStruct,
				artists: fullArtists,
			},
			wantFilteredArtists: filteredArtists,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotFilteredArtists := filterArtists(tt.args.c, tt.args.s, tt.args.artists); !reflect.DeepEqual(gotFilteredArtists, tt.wantFilteredArtists) {
				t.Errorf("%s = %v, want %v", fnName, gotFilteredArtists, tt.wantFilteredArtists)
			}
		})
	}
}
