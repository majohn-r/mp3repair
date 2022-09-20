package commands

import (
	"flag"
	"fmt"
	"mp3/internal"
	"mp3/internal/files"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func Test_list_validateTrackSorting(t *testing.T) {
	fnName := "list.validateTrackSorting()"
	tests := []struct {
		name          string
		sortingInput  string
		includeAlbums bool
		wantSorting   string
		internal.WantedOutput
	}{
		{name: "alpha sorting with albums", sortingInput: "alpha", includeAlbums: true, wantSorting: "alpha"},
		{name: "alpha sorting without albums", sortingInput: "alpha", includeAlbums: false, wantSorting: "alpha"},
		{name: "numeric sorting with albums", sortingInput: "numeric", includeAlbums: true, wantSorting: "numeric"},
		{
			name:          "numeric sorting without albums",
			sortingInput:  "numeric",
			includeAlbums: false,
			wantSorting:   "alpha",
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The value of the -sort flag, 'numeric', cannot be used unless '-includeAlbums' is true; track sorting will be alphabetic.\n",
				WantLogOutput:   "level='error' -includeAlbums='false' -sort='numeric' msg='numeric track sorting is not applicable'\n",
			},
		},
		{
			name:          "invalid sorting with albums",
			sortingInput:  "nonsense",
			includeAlbums: true,
			wantSorting:   "numeric",
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The \"-sort\" value you specified, \"nonsense\", is not valid.\n",
				WantLogOutput:   "level='error' -sort='nonsense' command='list' msg='flag value is not valid'\n",
			},
		},
		{
			name:          "invalid sorting without albums",
			sortingInput:  "nonsense",
			includeAlbums: false,
			wantSorting:   "alpha",
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The \"-sort\" value you specified, \"nonsense\", is not valid.\n",
				WantLogOutput:   "level='error' -sort='nonsense' command='list' msg='flag value is not valid'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("list", flag.ContinueOnError)
			o := internal.NewOutputDeviceForTesting()
			listCommand, _ := newListCommand(o, internal.EmptyConfiguration(), fs)
			listCommand.trackSorting = &tt.sortingInput
			listCommand.includeAlbums = &tt.includeAlbums
			listCommand.validateTrackSorting(o)
			if *listCommand.trackSorting != tt.wantSorting {
				t.Errorf("%s: got %q, want %q", fnName, *listCommand.trackSorting, tt.wantSorting)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
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
		artist := internal.CreateArtistNameForTesting(j)
		for k := 0; k < 10; k++ {
			album := internal.CreateAlbumNameForTesting(k)
			for m := 0; m < 10; m++ {
				track := internal.CreateTrackNameForTesting(m)
				trackCollection = append(trackCollection, &testTrack{
					artistName: artist,
					albumName:  album,
					trackName:  track,
				})
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
				albumTitle = fmt.Sprintf("%q by %q", albumName, tt.artistName)
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
			trackName, trackNumber := files.ParseTrackNameForTesting(tt.trackName)
			key := trackName
			if annotated {
				if !albums {
					key = fmt.Sprintf("%q on %q by %q", trackName, tt.albumName, tt.artistName)
					if !artists {
					} else {
						key = fmt.Sprintf("%q on %q", trackName, tt.albumName)
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

func newListForTesting() *list {
	l, _ := newListCommand(internal.NewOutputDeviceForTesting(), internal.EmptyConfiguration(), flag.NewFlagSet("list", flag.ContinueOnError))
	return l
}

func Test_list_Exec(t *testing.T) {
	fnName := "list.Exec()"
	// generate test data
	topDir := "loadTest"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, topDir, err)
	}
	savedHome := internal.SaveEnvVarForTesting("HOMEPATH")
	home := internal.SavedEnvVar{
		Name:  "HOMEPATH",
		Value: "C:\\Users\\The User",
		Set:   true,
	}
	home.RestoreForTesting()
	defer func() {
		savedHome.RestoreForTesting()
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	if err := internal.PopulateTopDirForTesting(topDir); err != nil {
		t.Errorf("%s error populating %q: %v", fnName, topDir, err)
	}
	type args struct {
		args []string
	}
	tests := []struct {
		name string
		l    *list
		args
		internal.WantedOutput
	}{
		{
			name: "help",
			l:    newListForTesting(),
			args: args{[]string{"--help"}},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "Usage of list:\n" +
					"  -albumFilter regular expression\n" +
					"    \tregular expression specifying which albums to select (default \".*\")\n" +
					"  -annotate\n" +
					"    \tannotate listings with album and artist data (default false)\n" +
					"  -artistFilter regular expression\n" +
					"    \tregular expression specifying which artists to select (default \".*\")\n" +
					"  -details\n" +
					"    \tinclude details with tracks (default false)\n" +
					"  -diagnostic\n" +
					"    \tinclude diagnostic information with tracks (default false)\n" +
					"  -ext extension\n" +
					"    \textension identifying music files (default \".mp3\")\n" +
					"  -includeAlbums\n" +
					"    \tinclude album names in listing (default true)\n" +
					"  -includeArtists\n" +
					"    \tinclude artist names in listing (default true)\n" +
					"  -includeTracks\n" +
					"    \tinclude track names in listing (default false)\n" +
					"  -sort sorting\n" +
					"    \ttrack sorting, 'numeric' in track number order, or 'alpha' in track name order (default \"numeric\")\n" +
					"  -topDir directory\n" +
					"    \ttop directory specifying where to find music files (default \"C:\\\\Users\\\\The User\\\\Music\")\n",
				WantLogOutput: "level='error' arguments='[--help]' msg='flag: help requested'\n",
			},
		},
		{
			name: "no output",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=false",
					"-includeArtists=false",
					"-includeTracks=false",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(false, false, false, false, false),
				WantErrorOutput:   "You disabled all functionality for the command \"list\".\n",
				WantLogOutput:     "level='error' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='false' -includeTracks='false' -sort='numeric' command='list' msg='the user disabled all functionality'\n",
			},
		},
		// tracks only
		{
			name: "unannotated tracks only",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=false",
					"-includeArtists=false",
					"-includeTracks=true",
					"-annotate=false",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(false, false, true, false, false),
				WantErrorOutput:   "The value of the -sort flag, 'numeric', cannot be used unless '-includeAlbums' is true; track sorting will be alphabetic.\n",
				WantLogOutput: "level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='false' -includeTracks='true' -sort='numeric' command='list' msg='executing command'\n" +
					"level='error' -includeAlbums='false' -sort='numeric' msg='numeric track sorting is not applicable'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		{
			name: "annotated tracks only",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=false",
					"-includeArtists=false",
					"-includeTracks=true",
					"-annotate=true",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(false, false, true, true, false),
				WantErrorOutput:   "The value of the -sort flag, 'numeric', cannot be used unless '-includeAlbums' is true; track sorting will be alphabetic.\n",
				WantLogOutput: "level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='false' -includeTracks='true' -sort='numeric' command='list' msg='executing command'\n" +
					"level='error' -includeAlbums='false' -sort='numeric' msg='numeric track sorting is not applicable'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		{
			name: "unannotated tracks only with numeric sorting",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=false",
					"-includeArtists=false",
					"-includeTracks=true",
					"-annotate=false",
					"-sort=numeric",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(false, false, true, false, true),
				WantErrorOutput:   "The value of the -sort flag, 'numeric', cannot be used unless '-includeAlbums' is true; track sorting will be alphabetic.\n",
				WantLogOutput: "level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='false' -includeTracks='true' -sort='numeric' command='list' msg='executing command'\n" +
					"level='error' -includeAlbums='false' -sort='numeric' msg='numeric track sorting is not applicable'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		{
			name: "annotated tracks only with numeric sorting",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=false",
					"-includeArtists=false",
					"-includeTracks=true",
					"-annotate=true",
					"-sort", "numeric",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(false, false, true, true, true),
				WantErrorOutput:   "The value of the -sort flag, 'numeric', cannot be used unless '-includeAlbums' is true; track sorting will be alphabetic.\n",
				WantLogOutput: "level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='false' -includeTracks='true' -sort='numeric' command='list' msg='executing command'\n" +
					"level='error' -includeAlbums='false' -sort='numeric' msg='numeric track sorting is not applicable'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		// albums only
		{
			name: "unannotated albums only",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=true",
					"-includeArtists=false",
					"-includeTracks=false",
					"-annotate=false",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(false, true, false, false, false),
				WantLogOutput: "level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='false' -includeTracks='false' -sort='numeric' command='list' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		{
			name: "annotated albums only",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=true",
					"-includeArtists=false",
					"-includeTracks=false",
					"-annotate=true",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(false, true, false, true, false),
				WantLogOutput: "level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='false' -includeTracks='false' -sort='numeric' command='list' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		// artists only
		{
			name: "unannotated artists only",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=false",
					"-includeArtists=true",
					"-includeTracks=false",
					"-annotate=false",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(true, false, false, false, false),
				WantLogOutput: "level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='true' -includeTracks='false' -sort='numeric' command='list' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		{
			name: "annotated artists only",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=false",
					"-includeArtists=true",
					"-includeTracks=false",
					"-annotate=true",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(true, false, false, true, false),
				WantLogOutput: "level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='true' -includeTracks='false' -sort='numeric' command='list' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		// albums and artists
		{
			name: "unannotated artists and albums only",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=true",
					"-includeArtists=true",
					"-includeTracks=false",
					"-annotate=false",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(true, true, false, false, false),
				WantLogOutput: "level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='true' -includeTracks='false' -sort='numeric' command='list' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		{
			name: "annotated artists and albums only",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=true",
					"-includeArtists=true",
					"-includeTracks=false",
					"-annotate=true",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(true, true, false, true, false),
				WantLogOutput: "level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='true' -includeTracks='false' -sort='numeric' command='list' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		// albums and tracks
		{
			name: "unannotated albums and tracks with alpha sorting",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=true",
					"-includeArtists=false",
					"-includeTracks=true",
					"-annotate=false",
					"-sort", "alpha",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(false, true, true, false, false),
				WantLogOutput: "level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='false' -includeTracks='true' -sort='alpha' command='list' msg='executing command'\n" +
					"level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='false' -includeTracks='true' -sort='alpha' command='list' msg='one or more flags were overridden'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		{
			name: "annotated albums and tracks with alpha sorting",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=true",
					"-includeArtists=false",
					"-includeTracks=true",
					"-annotate=true",
					"-sort", "alpha",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(false, true, true, true, false),
				WantLogOutput: "level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='false' -includeTracks='true' -sort='alpha' command='list' msg='executing command'\n" +
					"level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='false' -includeTracks='true' -sort='alpha' command='list' msg='one or more flags were overridden'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		{
			name: "unannotated albums and tracks with numeric sorting",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=true",
					"-includeArtists=false",
					"-includeTracks=true",
					"-annotate=false",
					"-sort", "numeric",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(false, true, true, false, true),
				WantLogOutput: "level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='false' -includeTracks='true' -sort='numeric' command='list' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		{
			name: "annotated albums and tracks with numeric sorting",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=true",
					"-includeArtists=false",
					"-includeTracks=true",
					"-annotate=true",
					"-sort", "numeric",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(false, true, true, true, true),
				WantLogOutput: "level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='false' -includeTracks='true' -sort='numeric' command='list' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		// artists and tracks
		{
			name: "unannotated artists and tracks with alpha sorting",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=false",
					"-includeArtists=true",
					"-includeTracks=true",
					"-annotate=false",
					"-sort", "alpha",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(true, false, true, false, false),
				WantLogOutput: "level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='true' -includeTracks='true' -sort='alpha' command='list' msg='executing command'\n" +
					"level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='true' -includeTracks='true' -sort='alpha' command='list' msg='one or more flags were overridden'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		{
			name: "annotated artists and tracks with alpha sorting",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=false",
					"-includeArtists=true",
					"-includeTracks=true",
					"-annotate=true",
					"-sort", "alpha",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(true, false, true, true, false),
				WantLogOutput: "level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='true' -includeTracks='true' -sort='alpha' command='list' msg='executing command'\n" +
					"level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='true' -includeTracks='true' -sort='alpha' command='list' msg='one or more flags were overridden'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		{
			name: "unannotated artists and tracks with numeric sorting",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=false",
					"-includeArtists=true",
					"-includeTracks=true",
					"-annotate=false",
					"-sort", "numeric",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(true, false, true, false, true),
				WantErrorOutput:   "The value of the -sort flag, 'numeric', cannot be used unless '-includeAlbums' is true; track sorting will be alphabetic.\n",
				WantLogOutput: "level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='true' -includeTracks='true' -sort='numeric' command='list' msg='executing command'\n" +
					"level='error' -includeAlbums='false' -sort='numeric' msg='numeric track sorting is not applicable'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		{
			name: "annotated artists and tracks with numeric sorting",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=false",
					"-includeArtists=true",
					"-includeTracks=true",
					"-annotate=true",
					"-sort", "numeric",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(true, false, true, true, true),
				WantErrorOutput:   "The value of the -sort flag, 'numeric', cannot be used unless '-includeAlbums' is true; track sorting will be alphabetic.\n",
				WantLogOutput: "level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='true' -includeTracks='true' -sort='numeric' command='list' msg='executing command'\n" +
					"level='error' -includeAlbums='false' -sort='numeric' msg='numeric track sorting is not applicable'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		// albums, artists, and tracks
		{
			name: "unannotated artists, albums, and tracks with alpha sorting",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=true",
					"-includeArtists=true",
					"-includeTracks=true",
					"-annotate=false",
					"-sort", "alpha",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(true, true, true, false, false),
				WantLogOutput: "level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='true' -includeTracks='true' -sort='alpha' command='list' msg='executing command'\n" +
					"level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='true' -includeTracks='true' -sort='alpha' command='list' msg='one or more flags were overridden'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		{
			name: "annotated artists, albums, and tracks with alpha sorting",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=true",
					"-includeArtists=true",
					"-includeTracks=true",
					"-annotate=true",
					"-sort", "alpha",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(true, true, true, true, false),
				WantLogOutput: "level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='true' -includeTracks='true' -sort='alpha' command='list' msg='executing command'\n" +
					"level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='true' -includeTracks='true' -sort='alpha' command='list' msg='one or more flags were overridden'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		{
			name: "unannotated artists, albums, and tracks with numeric sorting",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=true",
					"-includeArtists=true",
					"-includeTracks=true",
					"-annotate=false",
					"-sort", "numeric",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(true, true, true, false, true),
				WantLogOutput: "level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='true' -includeTracks='true' -sort='numeric' command='list' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		{
			name: "annotated artists, albums, and tracks with numeric sorting",
			l:    newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=true",
					"-includeArtists=true",
					"-includeTracks=true",
					"-annotate=true",
					"-sort", "numeric",
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: generateListing(true, true, true, true, true),
				WantLogOutput: "level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='true' -includeTracks='true' -sort='numeric' command='list' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			tt.l.Exec(o, tt.args.args)
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_newListCommand(t *testing.T) {
	fnName := "newListCommand()"
	savedAppData := internal.SaveEnvVarForTesting("APPDATA")
	os.Setenv("APPDATA", internal.SecureAbsolutePathForTesting("."))
	savedFoo := internal.SaveEnvVarForTesting("FOO")
	os.Unsetenv("FOO")
	defer func() {
		savedAppData.RestoreForTesting()
		savedFoo.RestoreForTesting()
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
		wantIncludeAlbums    bool
		wantIncludeArtists   bool
		wantIncludeTracks    bool
		wantTrackSorting     string
		wantAnnotateListings bool
		wantOk               bool
		internal.WantedOutput
	}{
		{
			name:                 "ordinary defaults",
			args:                 args{c: internal.EmptyConfiguration()},
			wantIncludeAlbums:    true,
			wantIncludeArtists:   true,
			wantIncludeTracks:    false,
			wantTrackSorting:     "numeric",
			wantAnnotateListings: false,
			wantOk:               true,
		},
		{
			name:                 "overridden defaults",
			args:                 args{c: defaultConfig},
			wantIncludeAlbums:    false,
			wantIncludeArtists:   false,
			wantIncludeTracks:    true,
			wantTrackSorting:     "alpha",
			wantAnnotateListings: true,
			wantOk:               true,
		},
		{
			name: "bad default for includeAlbums",
			args: args{
				c: internal.CreateConfiguration(internal.NewOutputDeviceForTesting(), map[string]interface{}{
					"list": map[string]interface{}{
						"includeAlbums": "nope",
					},
				}),
			},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The configuration file \"defaults.yaml\" contains an invalid value for \"list\": invalid boolean value \"nope\" for -includeAlbums: parse error.\n",
				WantLogOutput:   "level='error' error='invalid boolean value \"nope\" for -includeAlbums: parse error' section='list' msg='invalid content in configuration file'\n",
			},
		},
		{
			name: "bad default for includeArtists",
			args: args{
				c: internal.CreateConfiguration(internal.NewOutputDeviceForTesting(), map[string]interface{}{
					"list": map[string]interface{}{
						"includeArtists": "yes",
					},
				}),
			},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The configuration file \"defaults.yaml\" contains an invalid value for \"list\": invalid boolean value \"yes\" for -includeArtists: parse error.\n",
				WantLogOutput:   "level='error' error='invalid boolean value \"yes\" for -includeArtists: parse error' section='list' msg='invalid content in configuration file'\n",
			},
		},
		{
			name: "bad default for includeTracks",
			args: args{
				c: internal.CreateConfiguration(internal.NewOutputDeviceForTesting(), map[string]interface{}{
					"list": map[string]interface{}{
						"includeTracks": "sure",
					},
				}),
			},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The configuration file \"defaults.yaml\" contains an invalid value for \"list\": invalid boolean value \"sure\" for -includeTracks: parse error.\n",
				WantLogOutput:   "level='error' error='invalid boolean value \"sure\" for -includeTracks: parse error' section='list' msg='invalid content in configuration file'\n",
			},
		},
		{
			name: "bad default for annotate",
			args: args{
				c: internal.CreateConfiguration(internal.NewOutputDeviceForTesting(), map[string]interface{}{
					"list": map[string]interface{}{
						"annotate": "+2",
					},
				}),
			},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The configuration file \"defaults.yaml\" contains an invalid value for \"list\": invalid boolean value \"+2\" for -annotate: parse error.\n",
				WantLogOutput:   "level='error' error='invalid boolean value \"+2\" for -annotate: parse error' section='list' msg='invalid content in configuration file'\n",
			},
		},
		{
			name: "bad default for details",
			args: args{
				c: internal.CreateConfiguration(internal.NewOutputDeviceForTesting(), map[string]interface{}{
					"list": map[string]interface{}{
						"details": "no!",
					},
				}),
			},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The configuration file \"defaults.yaml\" contains an invalid value for \"list\": invalid boolean value \"no!\" for -details: parse error.\n",
				WantLogOutput:   "level='error' error='invalid boolean value \"no!\" for -details: parse error' section='list' msg='invalid content in configuration file'\n",
			},
		},
		{
			name: "bad default for diagnostics",
			args: args{
				c: internal.CreateConfiguration(internal.NewOutputDeviceForTesting(), map[string]interface{}{
					"list": map[string]interface{}{
						"diagnostic": "no!",
					},
				}),
			},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The configuration file \"defaults.yaml\" contains an invalid value for \"list\": invalid boolean value \"no!\" for -diagnostic: parse error.\n",
				WantLogOutput:   "level='error' error='invalid boolean value \"no!\" for -diagnostic: parse error' section='list' msg='invalid content in configuration file'\n",
			},
		},
		{
			name: "bad default for sorting",
			args: args{
				c: internal.CreateConfiguration(internal.NewOutputDeviceForTesting(), map[string]interface{}{
					"list": map[string]interface{}{
						"sort": "$FOO",
					},
				}),
			},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The configuration file \"defaults.yaml\" contains an invalid value for \"list\": invalid value \"$FOO\" for flag -sort: missing environment variables: [FOO].\n",
				WantLogOutput:   "level='error' error='invalid value \"$FOO\" for flag -sort: missing environment variables: [FOO]' section='list' msg='invalid content in configuration file'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			list, gotOk := newListCommand(o, tt.args.c, flag.NewFlagSet("list", flag.ContinueOnError))
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk %t wantOk %t", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
			if list != nil {
				if _, ok := list.sf.ProcessArgs(internal.NewOutputDeviceForTesting(), []string{
					"-topDir", topDir,
					"-ext", ".mp3",
				}); ok {
					if *list.includeAlbums != tt.wantIncludeAlbums {
						t.Errorf("%s %q: got includeAlbums %t want %t", fnName, tt.name, *list.includeAlbums, tt.wantIncludeAlbums)
					}
					if *list.includeArtists != tt.wantIncludeArtists {
						t.Errorf("%s %q: got includeArtists %t want %t", fnName, tt.name, *list.includeArtists, tt.wantIncludeArtists)
					}
					if *list.includeTracks != tt.wantIncludeTracks {
						t.Errorf("%s %q: got includeTracks %t want %t", fnName, tt.name, *list.includeTracks, tt.wantIncludeTracks)
					}
					if *list.annotateListings != tt.wantAnnotateListings {
						t.Errorf("%s %q: got annotateListings %t want %t", fnName, tt.name, *list.annotateListings, tt.wantAnnotateListings)
					}
					if *list.trackSorting != tt.wantTrackSorting {
						t.Errorf("%s %q: got trackSorting %q want %q", fnName, tt.name, *list.trackSorting, tt.wantTrackSorting)
					}
				} else {
					t.Errorf("%s %q: error processing arguments", fnName, tt.name)
				}
			}
		})
	}
}

func Test_list_outputTrackDiagnostics(t *testing.T) {
	fnName := "list.outputTrackDiagnostics"
	badArtist := files.NewArtist("bad artist", "./BadArtist")
	badAlbum := files.NewAlbum("bad album", badArtist, "BadAlbum")
	badTrack := files.NewTrack(badAlbum, "01 bad track.mp3", "bad track", 1)
	makeList := func() *list {
		l, _ := newListCommand(internal.NewOutputDeviceForTesting(), internal.EmptyConfiguration(), flag.NewFlagSet("list", flag.ContinueOnError))
		t := true
		l.diagnostics = &t
		return l
	}
	frames := map[string]string{
		"TYER": "2022",
		"TALB": "unknown album",
		"TRCK": "2",
		"TCON": "dance music",
		"TCOM": "a couple of idiots",
		"TIT2": "unknown track",
		"TPE1": "unknown artist",
		"TLEN": "1000",
	}
	topDir := "runDiagnostics"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, topDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	goodArtistDir := filepath.Join(topDir, "good artist")
	if err := internal.Mkdir(goodArtistDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, goodArtistDir, err)
	}
	goodAlbumDir := filepath.Join(goodArtistDir, "good album")
	if err := internal.Mkdir(goodAlbumDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, goodAlbumDir, err)
	}
	content := createTaggedContent(frames)
	content = append(content, internal.ID3V1DataSet1...)
	trackName := "01 new track.mp3"
	if err := internal.CreateFileForTestingWithContent(goodAlbumDir, trackName, content); err != nil {
		t.Errorf("%s error creating file %q: %v", fnName, trackName, err)
	}
	artist := files.NewArtist("good artist", goodArtistDir)
	album := files.NewAlbum("good album", artist, goodAlbumDir)
	artist.AddAlbum(album)
	goodTrack := files.NewTrack(album, trackName, "new track", 1)
	album.AddTrack(goodTrack)
	type args struct {
		t      *files.Track
		prefix string
	}
	tests := []struct {
		name string
		l    *list
		args
		internal.WantedOutput
	}{
		{
			name: "error case",
			l:    makeList(),
			args: args{t: badTrack},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "An error occurred when trying to read ID3V2 tag information for track \"bad track\" on album \"bad album\" by artist \"bad artist\": \"open BadAlbum\\\\01 bad track.mp3: The system cannot find the path specified.\".\n" +
					"An error occurred when trying to read ID3V1 tag information for track \"bad track\" on album \"bad album\" by artist \"bad artist\": \"open BadAlbum\\\\01 bad track.mp3: The system cannot find the path specified.\".\n",
				WantLogOutput: "level='error' error='open BadAlbum\\01 bad track.mp3: The system cannot find the path specified.' track='BadAlbum\\01 bad track.mp3' msg='id3v2 tag error'\n" +
					"level='error' error='open BadAlbum\\01 bad track.mp3: The system cannot find the path specified.' track='BadAlbum\\01 bad track.mp3' msg='id3v1 tag error'\n",
			},
		},
		{
			name: "success case",
			l:    makeList(),
			args: args{t: goodTrack, prefix: "      "},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: "      ID3V2 Version: 3\n" +
					"      ID3V2 Encoding: \"ISO-8859-1\"\n" +
					"      ID3V2 TALB = \"unknown album\"\n" +
					"      ID3V2 TCOM = \"a couple of idiots\"\n" +
					"      ID3V2 TCON = \"dance music\"\n" +
					"      ID3V2 TIT2 = \"unknown track\"\n" +
					"      ID3V2 TLEN = \"1000\"\n" +
					"      ID3V2 TPE1 = \"unknown artist\"\n" +
					"      ID3V2 TRCK = \"2\"\n" +
					"      ID3V2 TYER = \"2022\"\n" +
					"      ID3V1 Artist: \"The Beatles\"\n" +
					"      ID3V1 Album: \"On Air: Live At The BBC, Volum\"\n" +
					"      ID3V1 Title: \"Ringo - Pop Profile [Interview\"\n" +
					"      ID3V1 Track: 29\n" +
					"      ID3V1 Year: \"2013\"\n" +
					"      ID3V1 Genre: \"Other\"\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			tt.l.outputTrackDiagnostics(o, tt.args.t, tt.args.prefix)
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_list_outputTrackDetails(t *testing.T) {
	fnName := "list.outputTrackDetails()"
	badArtist := files.NewArtist("bad artist", "./BadArtist")
	badAlbum := files.NewAlbum("bad album", badArtist, "BadAlbum")
	badTrack := files.NewTrack(badAlbum, "01 bad track.mp3", "bad track", 1)
	makeList := func() *list {
		l, _ := newListCommand(internal.NewOutputDeviceForTesting(), internal.EmptyConfiguration(), flag.NewFlagSet("list", flag.ContinueOnError))
		t := true
		l.details = &t
		return l
	}
	frames := map[string]string{
		"TYER": "2022",
		"TALB": "unknown album",
		"TRCK": "2",
		"TCON": "dance music",
		"TCOM": "a couple of idiots",
		"TIT2": "unknown track",
		"TPE1": "unknown artist",
		"TLEN": "1000",
		"T???": "who knows?",
		"TEXT": "An infinite number of monkeys with a typewriter",
		"TIT3": "Part II",
		"TKEY": "D Major",
		"TPE2": "The usual gang of idiots",
		"TPE3": "Someone with a stick",
	}
	topDir := "details"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, topDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	goodArtistDir := filepath.Join(topDir, "good artist")
	if err := internal.Mkdir(goodArtistDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, goodArtistDir, err)
	}
	goodAlbumDir := filepath.Join(goodArtistDir, "good album")
	if err := internal.Mkdir(goodAlbumDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, goodAlbumDir, err)
	}
	content := createTaggedContent(frames)
	content = append(content, internal.ID3V1DataSet1...)
	trackName := "01 new track.mp3"
	if err := internal.CreateFileForTestingWithContent(goodAlbumDir, trackName, content); err != nil {
		t.Errorf("%s error creating file %q: %v", fnName, trackName, err)
	}
	artist := files.NewArtist("good artist", goodArtistDir)
	album := files.NewAlbum("good album", artist, goodAlbumDir)
	artist.AddAlbum(album)
	goodTrack := files.NewTrack(album, trackName, "new track", 1)
	album.AddTrack(goodTrack)
	type args struct {
		t      *files.Track
		prefix string
	}
	tests := []struct {
		name string
		l    *list
		args
		internal.WantedOutput
	}{
		{
			name: "error case",
			l:    makeList(),
			args: args{t: badTrack},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The details are not available for track \"bad track\" on album \"bad album\" by artist \"bad artist\": \"open BadAlbum\\\\01 bad track.mp3: The system cannot find the path specified.\".\n",
				WantLogOutput:   "level='error' error='open BadAlbum\\01 bad track.mp3: The system cannot find the path specified.' track='BadAlbum\\01 bad track.mp3' msg='cannot get details'\n",
			},
		},
		{
			name: "success case",
			l:    makeList(),
			args: args{t: goodTrack, prefix: "-->"},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: "-->Details:\n" +
					"-->  Composer = \"a couple of idiots\"\n" +
					"-->  Conductor = \"Someone with a stick\"\n" +
					"-->  Key = \"D Major\"\n" +
					"-->  Lyricist = \"An infinite number of monkeys with a typewriter\"\n" +
					"-->  Orchestra/Band = \"The usual gang of idiots\"\n" +
					"-->  Subtitle = \"Part II\"\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			tt.l.outputTrackDetails(o, tt.args.t, tt.args.prefix)
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}