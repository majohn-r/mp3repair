package subcommands

import (
	"bytes"
	"mp3/internal"
	"mp3/internal/files"
	"reflect"
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
	if err := internal.Mkdir(emptyDirName); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, emptyDirName, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, emptyDirName)
		internal.DestroyDirectoryForTesting(fnName, dirtyDirName)
	}()
	if err := internal.Mkdir(dirtyDirName); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, dirtyDirName, err)
	}
	if err := internal.PopulateTopDirForTesting(dirtyDirName); err != nil {
		t.Errorf("%s error populating %s: %v", fnName, dirtyDirName, err)
	}
	type args struct {
		s *files.Search
	}
	tests := []struct {
		name        string
		c           *check
		args        args
		wantArtists []*files.Artist
		wantW       string
	}{
		{name: "no work to do", c: &check{checkEmptyFolders: &fFlag}, args: args{}},
		{
			name:  "empty topDir",
			c:     &check{checkEmptyFolders: &tFlag},
			args:  args{s: files.CreateSearchForTesting(emptyDirName)},
			wantW: "Empty Folder Analysis: no empty folders found\n",
		},
		{
			name: "empty folders present",
			c:    &check{checkEmptyFolders: &tFlag},
			args: args{s: files.CreateSearchForTesting(dirtyDirName)},
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
			if gotArtists := tt.c.performEmptyFolderAnalysis(w, tt.args.s); !reflect.DeepEqual(gotArtists, tt.wantArtists) {
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
	type args struct {
		s       *files.Search
		artists []*files.Artist
	}
	tests := []struct {
		name                string
		c                   *check
		args                args
		wantFilteredArtists []*files.Artist
	}{
		{
			name: "neither gap analysis nor integrity enabled",
			c:    &check{checkGapsInTrackNumbering: &fFlag, checkIntegrity: &fFlag},
			args: args{},
		},
		{
			name:                "only gap analysis enabled, no artists supplied",
			c:                   &check{checkGapsInTrackNumbering: &tFlag, checkIntegrity: &fFlag},
			args:                args{s: searchStruct},
			wantFilteredArtists: filteredArtists,
		},
		{
			name:                "only gap analysis enabled, artists supplied",
			c:                   &check{checkGapsInTrackNumbering: &tFlag, checkIntegrity: &fFlag},
			args:                args{s: searchStruct, artists: fullArtists},
			wantFilteredArtists: filteredArtists,
		},
		{
			name:                "only integrity check enabled, no artists supplied",
			c:                   &check{checkGapsInTrackNumbering: &fFlag, checkIntegrity: &tFlag},
			args:                args{s: searchStruct},
			wantFilteredArtists: filteredArtists,
		},
		{
			name:                "only integrity check enabled, artists supplied",
			c:                   &check{checkGapsInTrackNumbering: &fFlag, checkIntegrity: &tFlag},
			args:                args{s: searchStruct, artists: fullArtists},
			wantFilteredArtists: filteredArtists,
		},
		{
			name:                "gap analysis and integrity check enabled, no artists supplied",
			c:                   &check{checkGapsInTrackNumbering: &tFlag, checkIntegrity: &tFlag},
			args:                args{s: searchStruct},
			wantFilteredArtists: filteredArtists,
		},
		{
			name:                "gap analysis and integrity check enabled, artists supplied",
			c:                   &check{checkGapsInTrackNumbering: &tFlag, checkIntegrity: &tFlag},
			args:                args{s: searchStruct, artists: fullArtists},
			wantFilteredArtists: filteredArtists,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotFilteredArtists := tt.c.filterArtists(tt.args.s, tt.args.artists); !reflect.DeepEqual(gotFilteredArtists, tt.wantFilteredArtists) {
				t.Errorf("%s = %v, want %v", fnName, gotFilteredArtists, tt.wantFilteredArtists)
			}
		})
	}
}

func Test_check_performGapAnalysis(t *testing.T) {
	goodArtist := &files.Artist{Name: "My Good Artist"}
	goodAlbum := &files.Album{Name: "An Excellent Album", RecordingArtist: goodArtist}
	goodArtist.Albums = append(goodArtist.Albums, goodAlbum)
	goodAlbum.Tracks = append(goodAlbum.Tracks, &files.Track{Name: "First Track", TrackNumber: 1, ContainingAlbum: goodAlbum})
	goodAlbum.Tracks = append(goodAlbum.Tracks, &files.Track{Name: "Second Track", TrackNumber: 2, ContainingAlbum: goodAlbum})
	goodAlbum.Tracks = append(goodAlbum.Tracks, &files.Track{Name: "Third Track", TrackNumber: 3, ContainingAlbum: goodAlbum})
	badArtist := &files.Artist{Name: "BadArtist"}
	badAlbum := &files.Album{Name: "No Biscuits For You!", RecordingArtist: badArtist}
	badArtist.Albums = append(badArtist.Albums, badAlbum)
	badAlbum.Tracks = append(badAlbum.Tracks, &files.Track{Name: "Awful Track", TrackNumber: 0, ContainingAlbum: badAlbum})
	badAlbum.Tracks = append(badAlbum.Tracks, &files.Track{Name: "Nasty Track", TrackNumber: 1, ContainingAlbum: badAlbum})
	badAlbum.Tracks = append(badAlbum.Tracks, &files.Track{Name: "Worse Track", TrackNumber: 1, ContainingAlbum: badAlbum})
	badAlbum.Tracks = append(badAlbum.Tracks, &files.Track{Name: "Bonus Track", TrackNumber: 9, ContainingAlbum: badAlbum})
	type args struct {
		artists []*files.Artist
	}
	tests := []struct {
		name  string
		c     *check
		args  args
		wantW string
	}{
		{name: "no analysis", c: &check{checkGapsInTrackNumbering: &fFlag}, args: args{}, wantW: ""},
		{
			name:  "no content",
			c:     &check{checkGapsInTrackNumbering: &tFlag},
			args:  args{},
			wantW: "Check Gaps: no gaps found\n",
		},
		{
			name:  "good artist",
			c:     &check{checkGapsInTrackNumbering: &tFlag},
			args:  args{artists: []*files.Artist{goodArtist}},
			wantW: "Check Gaps: no gaps found\n",
		},
		{
			name: "bad artist",
			c:    &check{checkGapsInTrackNumbering: &tFlag},
			args: args{artists: []*files.Artist{badArtist}},
			wantW: strings.Join([]string{
				"Check Gaps",
				"Artist: \"BadArtist\" album \"No Biscuits For You!\": missing track 2",
				"Artist: \"BadArtist\" album \"No Biscuits For You!\": missing track 3",
				"Artist: \"BadArtist\" album \"No Biscuits For You!\": missing track 4",
				"Artist: \"BadArtist\" album \"No Biscuits For You!\": track 0 (\"Awful Track\") is not a valid track number; valid tracks are 1..7",
				"Artist: \"BadArtist\" album \"No Biscuits For You!\": track 1 used by \"Nasty Track\" and \"Worse Track\"",
				"Artist: \"BadArtist\" album \"No Biscuits For You!\": track 9 (\"Bonus Track\") is not a valid track number; valid tracks are 1..7",
				"",
			}, "\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			tt.c.performGapAnalysis(w, tt.args.artists)
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("check.performGapAnalysis() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
