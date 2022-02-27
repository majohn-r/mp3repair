package subcommands

import (
	"bytes"
	"flag"
	"fmt"
	"mp3/internal"
	"mp3/internal/files"
	"os"
	"sort"
	"strings"
	"testing"
)

func Test_ls_validateTrackSorting(t *testing.T) {
	tests := []struct {
		name          string
		sortingInput  string
		includeAlbums bool
		wantSorting   string
	}{
		{name: "alpha sorting with albums", sortingInput: "alpha", includeAlbums: true, wantSorting: "alpha"},
		{name: "alpha sorting without albums", sortingInput: "alpha", includeAlbums: false, wantSorting: "alpha"},
		{name: "numeric sorting with albums", sortingInput: "numeric", includeAlbums: true, wantSorting: "numeric"},
		{name: "numeric sorting without albums", sortingInput: "numeric", includeAlbums: false, wantSorting: "alpha"},
		{name: "invalid sorting with albums", sortingInput: "nonsense", includeAlbums: true, wantSorting: "numeric"},
		{name: "invalid sorting without albums", sortingInput: "nonsense", includeAlbums: false, wantSorting: "alpha"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("ls", flag.ContinueOnError)
			lsCommand := newLsSubCommand(fs)
			lsCommand.trackSorting = &tt.sortingInput
			lsCommand.includeAlbums = &tt.includeAlbums
			lsCommand.validateTrackSorting()
			if *lsCommand.trackSorting != tt.wantSorting {
				t.Errorf("validateTrackSorting(): got %q, want %q", *lsCommand.trackSorting, tt.wantSorting)
			}
		})
	}
}

func Test_ls_runSubcommand(t *testing.T) {
	// generate test data
	topDir := "loadTest"
	if err := internal.Mkdir(t, "LoadData", topDir); err != nil {
		return
	}
	defer func() {
		if err := os.RemoveAll(topDir); err != nil {
			t.Errorf("LoadData() error destroying test directory %q: %v", topDir, err)
		}
	}()
	for k := 0; k < 10; k++ {
		internal.CreateArtist(t, topDir, k)
	}
	tests := []struct {
		name             string
		includeAlbums    bool
		includeArtists   bool
		includeTracks    bool
		trackSorting     string
		annotateListings bool
		wantW            string
	}{
		{
			name:           "no output",
			includeAlbums:  false,
			includeArtists: false,
			includeTracks:  false,
			wantW:          generateListing(false, false, false, false, false),
		},
		// tracks only
		{
			name:             "unannotated tracks only",
			includeAlbums:    false,
			includeArtists:   false,
			includeTracks:    true,
			annotateListings: false,
			wantW:            generateListing(false, false, true, false, false),
		},
		{
			name:             "annotated tracks only",
			includeAlbums:    false,
			includeArtists:   false,
			includeTracks:    true,
			annotateListings: true,
			wantW:            generateListing(false, false, true, true, false),
		},
		{
			name:             "unannotated tracks only with numeric sorting",
			includeAlbums:    false,
			includeArtists:   false,
			includeTracks:    true,
			annotateListings: false,
			trackSorting:     "numeric",
			wantW:            generateListing(false, false, true, false, true),
		},
		{
			name:             "annotated tracks only with numeric sorting",
			includeAlbums:    false,
			includeArtists:   false,
			includeTracks:    true,
			annotateListings: true,
			trackSorting:     "numeric",
			wantW:            generateListing(false, false, true, true, true),
		},
		// albums only
		{
			name:             "unannotated albums only",
			includeAlbums:    true,
			includeArtists:   false,
			includeTracks:    false,
			annotateListings: false,
			wantW:            generateListing(false, true, false, false, false),
		},
		{
			name:             "annotated albums only",
			includeAlbums:    true,
			includeArtists:   false,
			includeTracks:    false,
			annotateListings: true,
			wantW:            generateListing(false, true, false, true, false),
		},
		// artists only
		{
			name:             "unannotated artists only",
			includeAlbums:    false,
			includeArtists:   true,
			includeTracks:    false,
			annotateListings: false,
			wantW:            generateListing(true, false, false, false, false),
		},
		{
			name:             "annotated artists only",
			includeAlbums:    false,
			includeArtists:   true,
			includeTracks:    false,
			annotateListings: true,
			wantW:            generateListing(true, false, false, true, false),
		},
		// albums and artists
		{
			name:             "unannotated artists and albums only",
			includeAlbums:    true,
			includeArtists:   true,
			includeTracks:    false,
			annotateListings: false,
			wantW:            generateListing(true, true, false, false, false),
		},
		{
			name:             "annotated artists and albums only",
			includeAlbums:    true,
			includeArtists:   true,
			includeTracks:    false,
			annotateListings: true,
			wantW:            generateListing(true, true, false, true, false),
		},
		// albums and tracks
		{
			name:             "unannotated albums and tracks with alpha sorting",
			includeAlbums:    true,
			includeArtists:   false,
			includeTracks:    true,
			annotateListings: false,
			trackSorting:     "alpha",
			wantW:            generateListing(false, true, true, false, false),
		},
		{
			name:             "annotated albums and tracks with alpha sorting",
			includeAlbums:    true,
			includeArtists:   false,
			includeTracks:    true,
			annotateListings: true,
			trackSorting:     "alpha",
			wantW:            generateListing(false, true, true, true, false),
		},
		{
			name:             "unannotated albums and tracks with numeric sorting",
			includeAlbums:    true,
			includeArtists:   false,
			includeTracks:    true,
			annotateListings: false,
			trackSorting:     "numeric",
			wantW:            generateListing(false, true, true, false, true),
		},
		{
			name:             "annotated albums and tracks with numeric sorting",
			includeAlbums:    true,
			includeArtists:   false,
			includeTracks:    true,
			annotateListings: true,
			trackSorting:     "numeric",
			wantW:            generateListing(false, true, true, true, true),
		},
		// artists and tracks
		{
			name:             "unannotated artists and tracks with alpha sorting",
			includeAlbums:    false,
			includeArtists:   true,
			includeTracks:    true,
			annotateListings: false,
			trackSorting:     "alpha",
			wantW:            generateListing(true, false, true, false, false),
		},
		{
			name:             "annotated artists and tracks with alpha sorting",
			includeAlbums:    false,
			includeArtists:   true,
			includeTracks:    true,
			annotateListings: true,
			trackSorting:     "alpha",
			wantW:            generateListing(true, false, true, true, false),
		},
		{
			name:             "unannotated artists and tracks with numeric sorting",
			includeAlbums:    false,
			includeArtists:   true,
			includeTracks:    true,
			annotateListings: false,
			trackSorting:     "numeric",
			wantW:            generateListing(true, false, true, false, true),
		},
		{
			name:             "annotated artists and tracks with numeric sorting",
			includeAlbums:    false,
			includeArtists:   true,
			includeTracks:    true,
			annotateListings: true,
			trackSorting:     "numeric",
			wantW:            generateListing(true, false, true, true, true),
		},
		// albums, artists, and tracks
		{
			name:             "unannotated artists, albums, and tracks with alpha sorting",
			includeAlbums:    true,
			includeArtists:   true,
			includeTracks:    true,
			annotateListings: false,
			trackSorting:     "alpha",
			wantW:            generateListing(true, true, true, false, false),
		},
		{
			name:             "annotated artists, albums, and tracks with alpha sorting",
			includeAlbums:    true,
			includeArtists:   true,
			includeTracks:    true,
			annotateListings: true,
			trackSorting:     "alpha",
			wantW:            generateListing(true, true, true, true, false),
		},
		{
			name:             "unannotated artists, albums, and tracks with numeric sorting",
			includeAlbums:    true,
			includeArtists:   true,
			includeTracks:    true,
			annotateListings: false,
			trackSorting:     "numeric",
			wantW:            generateListing(true, true, true, false, true),
		},
		{
			name:             "annotated artists, albums, and tracks with numeric sorting",
			includeAlbums:    true,
			includeArtists:   true,
			includeTracks:    true,
			annotateListings: true,
			trackSorting:     "numeric",
			wantW:            generateListing(true, true, true, true, true),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			fs := flag.NewFlagSet("ls", flag.ContinueOnError)
			lsCommand := newLsSubCommand(fs)
			lsCommand.includeAlbums = &tt.includeAlbums
			lsCommand.includeArtists = &tt.includeArtists
			lsCommand.includeTracks = &tt.includeTracks
			lsCommand.trackSorting = &tt.trackSorting
			lsCommand.annotateListings = &tt.annotateListings
			lsCommand.commons.topDirectory = &topDir
			params := lsCommand.commons.processArgs(os.Stderr, nil)
			lsCommand.runSubcommand(w, params)
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("ls.runSubcommand() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

type testTrack struct {
	artistName string
	albumName  string
	trackName  string
}

func generateListing(artists, albums, tracks, annotated, sortNumerically bool) string {
	var trackCollection []*testTrack
	for j := 0; j < 10; j++ {
		artist := internal.CreateArtistName(j)
		for k := 0; k < 10; k++ {
			album := internal.CreateAlbumName(k)
			for m := 0; m < 10; m++ {
				track := internal.CreateTrackName(m)
				trackCollection = append(trackCollection, &testTrack{artistName: artist, albumName: album, trackName: track})
			}
		}
	}
	var output []string
	switch artists {
	case true:
		tracksByArtist := make(map[string][]*testTrack)
		for _, tt := range trackCollection {
			artistName := tt.artistName
			tracksByArtist[artistName] = append(tracksByArtist[artistName], tt)
		}
		var artistNames []string
		for key := range tracksByArtist {
			artistNames = append(artistNames, key)
		}
		sort.Strings(artistNames)
		for _, artistName := range artistNames {
			output = append(output, fmt.Sprintf("Artist: %s", artistName))
			output = append(output, generateAlbumListings(tracksByArtist[artistName], "  ", artists, albums, tracks, annotated, sortNumerically)...)
		}
	case false:
		output = append(output, generateAlbumListings(trackCollection, "", artists, albums, tracks, annotated, sortNumerically)...)
	}
	if len(output) != 0 {
		output = append(output, "") // force trailing newline
	}
	return strings.Join(output, "\n")
}

type albumType struct {
	artistName string
	albumName  string
}

type albumTypes []albumType

func (a albumTypes) Len() int {
	return len(a)
}

func (a albumTypes) Less(i, j int) bool {
	if a[i].albumName == a[j].albumName {
		return a[i].artistName < a[j].artistName
	}
	return a[i].albumName < a[j].albumName
}

func (a albumTypes) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func generateAlbumListings(testTracks []*testTrack, spacer string, artists, albums, tracks, annotated, sortNumerically bool) []string {
	var output []string
	switch albums {
	case true:
		albumsToList := make(map[albumType][]*testTrack)
		for _, tt := range testTracks {
			albumName := tt.albumName
			var albumTitle string
			if annotated && !artists {
				albumTitle = fmt.Sprintf("%s by %s", albumName, tt.artistName)
			} else {
				albumTitle = albumName
			}
			album := albumType{artistName: tt.artistName, albumName: albumTitle}
			albumsToList[album] = append(albumsToList[album], tt)
		}
		var albumNames albumTypes
		for key := range albumsToList {
			albumNames = append(albumNames, key)
		}

		sort.Sort(albumNames)
		for _, albumTitle := range albumNames {
			output = append(output, fmt.Sprintf("%sAlbum: %s", spacer, albumTitle.albumName))
			output = append(output, generateTrackListings(albumsToList[albumTitle], spacer+"  ", artists, albums, tracks, annotated, sortNumerically)...)
		}
	case false:
		output = append(output, generateTrackListings(testTracks, spacer, artists, albums, tracks, annotated, sortNumerically)...)
	}
	return output
}

func generateTrackListings(testTracks []*testTrack, spacer string, artists, albums, tracks, annotated, sortNumerically bool) []string {
	var output []string
	if tracks {
		var tracksToList []string
		for _, tt := range testTracks {
			trackName, trackNumber, _ := files.ParseTrackName(tt.trackName, tt.albumName, tt.artistName, files.DefaultFileExtension)
			key := trackName
			if annotated {
				if !albums {
					if !artists {
						key = fmt.Sprintf("%s on %s by %s", trackName, tt.albumName, tt.artistName)
					} else {
						key = fmt.Sprintf("%s on %s", trackName, tt.albumName)
					}
				}
			}
			if sortNumerically && albums {
				key = fmt.Sprintf("%2d. %s", trackNumber, trackName)
			}
			tracksToList = append(tracksToList, key)
		}
		sort.Strings(tracksToList)
		for _, trackName := range tracksToList {
			output = append(output, fmt.Sprintf("%s%s", spacer, trackName))
		}
	}
	return output
}
