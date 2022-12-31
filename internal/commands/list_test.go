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

	"github.com/majohn-r/output"
)

func Test_list_validateTrackSorting(t *testing.T) {
	const fnName = "list.validateTrackSorting()"
	tests := map[string]struct {
		sortingInput  string
		includeAlbums bool
		want          string
		output.WantedRecording
	}{
		"alpha sorting with albums":    {sortingInput: "alpha", includeAlbums: true, want: "alpha"},
		"alpha sorting without albums": {sortingInput: "alpha", includeAlbums: false, want: "alpha"},
		"numeric sorting with albums":  {sortingInput: "numeric", includeAlbums: true, want: "numeric"},
		"numeric sorting without albums": {
			sortingInput:  "numeric",
			includeAlbums: false,
			want:          "alpha",
			WantedRecording: output.WantedRecording{
				Error: "The \"-sort\" value you specified, \"numeric\", is not valid unless \"-includeAlbums\" is true; track sorting will be alphabetic.\n",
				Log:   "level='error' -includeAlbums='false' -sort='numeric' msg='numeric track sorting is not applicable'\n",
			},
		},
		"invalid sorting with albums": {
			sortingInput:  "nonsense",
			includeAlbums: true,
			want:          "numeric",
			WantedRecording: output.WantedRecording{
				Error: "The \"-sort\" value you specified, \"nonsense\", is not valid.\n",
				Log:   "level='error' -sort='nonsense' command='list' msg='flag value is not valid'\n",
			},
		},
		"invalid sorting without albums": {
			sortingInput:  "nonsense",
			includeAlbums: false,
			want:          "alpha",
			WantedRecording: output.WantedRecording{
				Error: "The \"-sort\" value you specified, \"nonsense\", is not valid.\n",
				Log:   "level='error' -sort='nonsense' command='list' msg='flag value is not valid'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fs := flag.NewFlagSet("list", flag.ContinueOnError)
			o := output.NewRecorder()
			l, _ := newListCommand(o, internal.EmptyConfiguration(), fs)
			l.trackSorting = &tt.sortingInput
			l.includeAlbums = &tt.includeAlbums
			l.validateTrackSorting(o)
			if *l.trackSorting != tt.want {
				t.Errorf("%s: got %q, want %q", fnName, *l.trackSorting, tt.want)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
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
	var ts []*testTrack
	for j := 0; j < 10; j++ {
		artist := internal.CreateArtistNameForTesting(j)
		for k := 0; k < 10; k++ {
			album := internal.CreateAlbumNameForTesting(k)
			for m := 0; m < 10; m++ {
				track := internal.CreateTrackNameForTesting(m)
				ts = append(ts, &testTrack{artistName: artist, albumName: album, trackName: track})
			}
		}
	}
	var listing []string
	switch artists {
	case true:
		m := make(map[string][]*testTrack)
		for _, tt := range ts {
			m[tt.artistName] = append(m[tt.artistName], tt)
		}
		var names []string
		for key := range m {
			names = append(names, key)
		}
		sort.Strings(names)
		for _, s := range names {
			listing = append(listing, fmt.Sprintf("Artist: %s", s))
			listing = append(listing, generateAlbumListings(m[s], "  ", artists, albums, tracks, annotated, sortNumerically)...)
		}
	case false:
		listing = append(listing, generateAlbumListings(ts, "", artists, albums, tracks, annotated, sortNumerically)...)
	}
	if len(listing) != 0 {
		listing = append(listing, "") // force trailing newline
	}
	return strings.Join(listing, "\n")
}

type albumType struct {
	artistName string
	albumName  string
}

type albumTypes []albumType

func (aT albumTypes) Len() int {
	return len(aT)
}

func (aT albumTypes) Less(i, j int) bool {
	if aT[i].albumName == aT[j].albumName {
		return aT[i].artistName < aT[j].artistName
	}
	return aT[i].albumName < aT[j].albumName
}

func (aT albumTypes) Swap(i, j int) {
	aT[i], aT[j] = aT[j], aT[i]
}

func generateAlbumListings(testTracks []*testTrack, spacer string, artists, albums, tracks, annotated, sortNumerically bool) []string {
	var listing []string
	switch albums {
	case true:
		m := make(map[albumType][]*testTrack)
		for _, tt := range testTracks {
			name := tt.albumName
			var title string
			if annotated && !artists {
				title = fmt.Sprintf("%q by %q", name, tt.artistName)
			} else {
				title = name
			}
			aT := albumType{artistName: tt.artistName, albumName: title}
			m[aT] = append(m[aT], tt)
		}
		var names albumTypes
		for key := range m {
			names = append(names, key)
		}

		sort.Sort(names)
		for _, aT := range names {
			listing = append(listing, fmt.Sprintf("%sAlbum: %s", spacer, aT.albumName))
			listing = append(listing, generateTrackListings(m[aT], spacer+"  ", artists, albums, tracks, annotated, sortNumerically)...)
		}
	case false:
		listing = append(listing, generateTrackListings(testTracks, spacer, artists, albums, tracks, annotated, sortNumerically)...)
	}
	return listing
}

func generateTrackListings(testTracks []*testTrack, spacer string, artists, albums, tracks, annotated, sortNumerically bool) []string {
	var listing []string
	if tracks {
		var tNames []string
		for _, tt := range testTracks {
			name, number := files.ParseTrackNameForTesting(tt.trackName)
			key := name
			if annotated {
				if !albums {
					key = fmt.Sprintf("%q on %q by %q", name, tt.albumName, tt.artistName)
					if artists {
						key = fmt.Sprintf("%q on %q", name, tt.albumName)
					}
				}
			}
			if sortNumerically && albums {
				key = fmt.Sprintf("%2d. %s", number, name)
			}
			tNames = append(tNames, key)
		}
		sort.Strings(tNames)
		for _, trackName := range tNames {
			listing = append(listing, fmt.Sprintf("%s%s", spacer, trackName))
		}
	}
	return listing
}

func newListForTesting() *list {
	l, _ := newListCommand(output.NewNilBus(), internal.EmptyConfiguration(), flag.NewFlagSet("list", flag.ContinueOnError))
	return l
}

func Test_list_Exec(t *testing.T) {
	const fnName = "list.Exec()"
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
	if err := internal.PopulateTopDirForTesting(topDir); err != nil {
		t.Errorf("%s error populating %q: %v", fnName, topDir, err)
	}
	defer func() {
		savedHome.RestoreForTesting()
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	type args struct {
		args []string
	}
	tests := map[string]struct {
		l *list
		args
		output.WantedRecording
	}{
		"help": {
			l:    newListForTesting(),
			args: args{[]string{"--help"}},
			WantedRecording: output.WantedRecording{
				Error: "Usage of list:\n" +
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
				Log: "level='error' arguments='[--help]' msg='flag: help requested'\n",
			},
		},
		"no output": {
			l: newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=false",
					"-includeArtists=false",
					"-includeTracks=false",
				},
			},
			WantedRecording: output.WantedRecording{
				Console: generateListing(false, false, false, false, false),
				Error:   "You disabled all functionality for the command \"list\".\n",
				Log:     "level='error' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='false' -includeTracks='false' -sort='numeric' command='list' msg='the user disabled all functionality'\n",
			},
		},
		// tracks only
		"unannotated tracks only": {
			l: newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=false",
					"-includeArtists=false",
					"-includeTracks=true",
					"-annotate=false",
				},
			},
			WantedRecording: output.WantedRecording{
				Console: generateListing(false, false, true, false, false),
				Error:   "The \"-sort\" value you specified, \"numeric\", is not valid unless \"-includeAlbums\" is true; track sorting will be alphabetic.\n",
				Log: "level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='false' -includeTracks='true' -sort='numeric' command='list' msg='executing command'\n" +
					"level='error' -includeAlbums='false' -sort='numeric' msg='numeric track sorting is not applicable'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		"annotated tracks only": {
			l: newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=false",
					"-includeArtists=false",
					"-includeTracks=true",
					"-annotate=true",
				},
			},
			WantedRecording: output.WantedRecording{
				Console: generateListing(false, false, true, true, false),
				Error:   "The \"-sort\" value you specified, \"numeric\", is not valid unless \"-includeAlbums\" is true; track sorting will be alphabetic.\n",
				Log: "level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='false' -includeTracks='true' -sort='numeric' command='list' msg='executing command'\n" +
					"level='error' -includeAlbums='false' -sort='numeric' msg='numeric track sorting is not applicable'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		"unannotated tracks only with numeric sorting": {
			l: newListForTesting(),
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
			WantedRecording: output.WantedRecording{
				Console: generateListing(false, false, true, false, true),
				Error:   "The \"-sort\" value you specified, \"numeric\", is not valid unless \"-includeAlbums\" is true; track sorting will be alphabetic.\n",
				Log: "level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='false' -includeTracks='true' -sort='numeric' command='list' msg='executing command'\n" +
					"level='error' -includeAlbums='false' -sort='numeric' msg='numeric track sorting is not applicable'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		"annotated tracks only with numeric sorting": {
			l: newListForTesting(),
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
			WantedRecording: output.WantedRecording{
				Console: generateListing(false, false, true, true, true),
				Error:   "The \"-sort\" value you specified, \"numeric\", is not valid unless \"-includeAlbums\" is true; track sorting will be alphabetic.\n",
				Log: "level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='false' -includeTracks='true' -sort='numeric' command='list' msg='executing command'\n" +
					"level='error' -includeAlbums='false' -sort='numeric' msg='numeric track sorting is not applicable'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		// albums only
		"unannotated albums only": {
			l: newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=true",
					"-includeArtists=false",
					"-includeTracks=false",
					"-annotate=false",
				},
			},
			WantedRecording: output.WantedRecording{
				Console: generateListing(false, true, false, false, false),
				Log: "level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='false' -includeTracks='false' -sort='numeric' command='list' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		"annotated albums only": {
			l: newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=true",
					"-includeArtists=false",
					"-includeTracks=false",
					"-annotate=true",
				},
			},
			WantedRecording: output.WantedRecording{
				Console: generateListing(false, true, false, true, false),
				Log: "level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='false' -includeTracks='false' -sort='numeric' command='list' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		// artists only
		"unannotated artists only": {
			l: newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=false",
					"-includeArtists=true",
					"-includeTracks=false",
					"-annotate=false",
				},
			},
			WantedRecording: output.WantedRecording{
				Console: generateListing(true, false, false, false, false),
				Log: "level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='true' -includeTracks='false' -sort='numeric' command='list' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		"annotated artists only": {
			l: newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=false",
					"-includeArtists=true",
					"-includeTracks=false",
					"-annotate=true",
				},
			},
			WantedRecording: output.WantedRecording{
				Console: generateListing(true, false, false, true, false),
				Log: "level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='true' -includeTracks='false' -sort='numeric' command='list' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		// albums and artists
		"unannotated artists and albums only": {
			l: newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=true",
					"-includeArtists=true",
					"-includeTracks=false",
					"-annotate=false",
				},
			},
			WantedRecording: output.WantedRecording{
				Console: generateListing(true, true, false, false, false),
				Log: "level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='true' -includeTracks='false' -sort='numeric' command='list' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		"annotated artists and albums only": {
			l: newListForTesting(),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=true",
					"-includeArtists=true",
					"-includeTracks=false",
					"-annotate=true",
				},
			},
			WantedRecording: output.WantedRecording{
				Console: generateListing(true, true, false, true, false),
				Log: "level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='true' -includeTracks='false' -sort='numeric' command='list' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		// albums and tracks
		"unannotated albums and tracks with alpha sorting": {
			l: newListForTesting(),
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
			WantedRecording: output.WantedRecording{
				Console: generateListing(false, true, true, false, false),
				Log: "level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='false' -includeTracks='true' -sort='alpha' command='list' msg='executing command'\n" +
					"level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='false' -includeTracks='true' -sort='alpha' command='list' msg='one or more flags were overridden'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		"annotated albums and tracks with alpha sorting": {
			l: newListForTesting(),
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
			WantedRecording: output.WantedRecording{
				Console: generateListing(false, true, true, true, false),
				Log: "level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='false' -includeTracks='true' -sort='alpha' command='list' msg='executing command'\n" +
					"level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='false' -includeTracks='true' -sort='alpha' command='list' msg='one or more flags were overridden'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		"unannotated albums and tracks with numeric sorting": {
			l: newListForTesting(),
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
			WantedRecording: output.WantedRecording{
				Console: generateListing(false, true, true, false, true),
				Log: "level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='false' -includeTracks='true' -sort='numeric' command='list' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		"annotated albums and tracks with numeric sorting": {
			l: newListForTesting(),
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
			WantedRecording: output.WantedRecording{
				Console: generateListing(false, true, true, true, true),
				Log: "level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='false' -includeTracks='true' -sort='numeric' command='list' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		// artists and tracks
		"unannotated artists and tracks with alpha sorting": {
			l: newListForTesting(),
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
			WantedRecording: output.WantedRecording{
				Console: generateListing(true, false, true, false, false),
				Log: "level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='true' -includeTracks='true' -sort='alpha' command='list' msg='executing command'\n" +
					"level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='true' -includeTracks='true' -sort='alpha' command='list' msg='one or more flags were overridden'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		"annotated artists and tracks with alpha sorting": {
			l: newListForTesting(),
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
			WantedRecording: output.WantedRecording{
				Console: generateListing(true, false, true, true, false),
				Log: "level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='true' -includeTracks='true' -sort='alpha' command='list' msg='executing command'\n" +
					"level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='true' -includeTracks='true' -sort='alpha' command='list' msg='one or more flags were overridden'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		"unannotated artists and tracks with numeric sorting": {
			l: newListForTesting(),
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
			WantedRecording: output.WantedRecording{
				Console: generateListing(true, false, true, false, true),
				Error:   "The \"-sort\" value you specified, \"numeric\", is not valid unless \"-includeAlbums\" is true; track sorting will be alphabetic.\n",
				Log: "level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='true' -includeTracks='true' -sort='numeric' command='list' msg='executing command'\n" +
					"level='error' -includeAlbums='false' -sort='numeric' msg='numeric track sorting is not applicable'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		"annotated artists and tracks with numeric sorting": {
			l: newListForTesting(),
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
			WantedRecording: output.WantedRecording{
				Console: generateListing(true, false, true, true, true),
				Error:   "The \"-sort\" value you specified, \"numeric\", is not valid unless \"-includeAlbums\" is true; track sorting will be alphabetic.\n",
				Log: "level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='false' -includeArtists='true' -includeTracks='true' -sort='numeric' command='list' msg='executing command'\n" +
					"level='error' -includeAlbums='false' -sort='numeric' msg='numeric track sorting is not applicable'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		// albums, artists, and tracks
		"unannotated artists, albums, and tracks with alpha sorting": {
			l: newListForTesting(),
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
			WantedRecording: output.WantedRecording{
				Console: generateListing(true, true, true, false, false),
				Log: "level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='true' -includeTracks='true' -sort='alpha' command='list' msg='executing command'\n" +
					"level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='true' -includeTracks='true' -sort='alpha' command='list' msg='one or more flags were overridden'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		"annotated artists, albums, and tracks with alpha sorting": {
			l: newListForTesting(),
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
			WantedRecording: output.WantedRecording{
				Console: generateListing(true, true, true, true, false),
				Log: "level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='true' -includeTracks='true' -sort='alpha' command='list' msg='executing command'\n" +
					"level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='true' -includeTracks='true' -sort='alpha' command='list' msg='one or more flags were overridden'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		"unannotated artists, albums, and tracks with numeric sorting": {
			l: newListForTesting(),
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
			WantedRecording: output.WantedRecording{
				Console: generateListing(true, true, true, false, true),
				Log: "level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='true' -includeTracks='true' -sort='numeric' command='list' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		"annotated artists, albums, and tracks with numeric sorting": {
			l: newListForTesting(),
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
			WantedRecording: output.WantedRecording{
				Console: generateListing(true, true, true, true, true),
				Log: "level='info' -annotate='true' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='true' -includeTracks='true' -sort='numeric' command='list' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.l.Exec(o, tt.args.args)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_newListCommand(t *testing.T) {
	const fnName = "newListCommand()"
	savedAppData := internal.SaveEnvVarForTesting("APPDATA")
	os.Setenv("APPDATA", internal.SecureAbsolutePathForTesting("."))
	oldAppPath := internal.ApplicationPath()
	internal.InitApplicationPath(output.NewNilBus())
	savedFoo := internal.SaveEnvVarForTesting("FOO")
	os.Unsetenv("FOO")
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
	defaultConfig, _ := internal.ReadConfigurationFile(output.NewNilBus())
	defer func() {
		savedAppData.RestoreForTesting()
		internal.SetApplicationPathForTesting(oldAppPath)
		savedFoo.RestoreForTesting()
		internal.DestroyDirectoryForTesting(fnName, topDir)
		internal.DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	type args struct {
		c *internal.Configuration
	}
	tests := map[string]struct {
		args
		wantIncludeAlbums    bool
		wantIncludeArtists   bool
		wantIncludeTracks    bool
		wantTrackSorting     string
		wantAnnotateListings bool
		wantOk               bool
		output.WantedRecording
	}{
		"ordinary defaults": {
			args:                 args{c: internal.EmptyConfiguration()},
			wantIncludeAlbums:    true,
			wantIncludeArtists:   true,
			wantIncludeTracks:    false,
			wantTrackSorting:     "numeric",
			wantAnnotateListings: false,
			wantOk:               true,
		},
		"overridden defaults": {
			args:                 args{c: defaultConfig},
			wantIncludeAlbums:    false,
			wantIncludeArtists:   false,
			wantIncludeTracks:    true,
			wantTrackSorting:     "alpha",
			wantAnnotateListings: true,
			wantOk:               true,
		},
		"bad default for includeAlbums": {
			args: args{
				c: internal.CreateConfiguration(output.NewNilBus(), map[string]any{
					"list": map[string]any{
						"includeAlbums": "nope",
					},
				}),
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"list\": invalid boolean value \"nope\" for -includeAlbums: parse error.\n",
				Log:   "level='error' error='invalid boolean value \"nope\" for -includeAlbums: parse error' section='list' msg='invalid content in configuration file'\n",
			},
		},
		"bad default for includeArtists": {
			args: args{
				c: internal.CreateConfiguration(output.NewNilBus(), map[string]any{
					"list": map[string]any{
						"includeArtists": "yes",
					},
				}),
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"list\": invalid boolean value \"yes\" for -includeArtists: parse error.\n",
				Log:   "level='error' error='invalid boolean value \"yes\" for -includeArtists: parse error' section='list' msg='invalid content in configuration file'\n",
			},
		},
		"bad default for includeTracks": {
			args: args{
				c: internal.CreateConfiguration(output.NewNilBus(), map[string]any{
					"list": map[string]any{
						"includeTracks": "sure",
					},
				}),
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"list\": invalid boolean value \"sure\" for -includeTracks: parse error.\n",
				Log:   "level='error' error='invalid boolean value \"sure\" for -includeTracks: parse error' section='list' msg='invalid content in configuration file'\n",
			},
		},
		"bad default for annotate": {
			args: args{
				c: internal.CreateConfiguration(output.NewNilBus(), map[string]any{
					"list": map[string]any{
						"annotate": "+2",
					},
				}),
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"list\": invalid boolean value \"+2\" for -annotate: parse error.\n",
				Log:   "level='error' error='invalid boolean value \"+2\" for -annotate: parse error' section='list' msg='invalid content in configuration file'\n",
			},
		},
		"bad default for details": {
			args: args{
				c: internal.CreateConfiguration(output.NewNilBus(), map[string]any{
					"list": map[string]any{
						"details": "no!",
					},
				}),
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"list\": invalid boolean value \"no!\" for -details: parse error.\n",
				Log:   "level='error' error='invalid boolean value \"no!\" for -details: parse error' section='list' msg='invalid content in configuration file'\n",
			},
		},
		"bad default for diagnostics": {
			args: args{
				c: internal.CreateConfiguration(output.NewNilBus(), map[string]any{
					"list": map[string]any{
						"diagnostic": "no!",
					},
				}),
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"list\": invalid boolean value \"no!\" for -diagnostic: parse error.\n",
				Log:   "level='error' error='invalid boolean value \"no!\" for -diagnostic: parse error' section='list' msg='invalid content in configuration file'\n",
			},
		},
		"bad default for sorting": {
			args: args{
				c: internal.CreateConfiguration(output.NewNilBus(), map[string]any{
					"list": map[string]any{
						"sort": "$FOO",
					},
				}),
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"list\": invalid value \"$FOO\" for flag -sort: missing environment variables: [FOO].\n",
				Log:   "level='error' error='invalid value \"$FOO\" for flag -sort: missing environment variables: [FOO]' section='list' msg='invalid content in configuration file'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			list, gotOk := newListCommand(o, tt.args.c, flag.NewFlagSet("list", flag.ContinueOnError))
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk %t wantOk %t", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
			if list != nil {
				if _, ok := list.sf.ProcessArgs(output.NewNilBus(), []string{
					"-topDir", topDir,
					"-ext", ".mp3",
				}); ok {
					if *list.includeAlbums != tt.wantIncludeAlbums {
						t.Errorf("%s %q: got includeAlbums %t want %t", fnName, name, *list.includeAlbums, tt.wantIncludeAlbums)
					}
					if *list.includeArtists != tt.wantIncludeArtists {
						t.Errorf("%s %q: got includeArtists %t want %t", fnName, name, *list.includeArtists, tt.wantIncludeArtists)
					}
					if *list.includeTracks != tt.wantIncludeTracks {
						t.Errorf("%s %q: got includeTracks %t want %t", fnName, name, *list.includeTracks, tt.wantIncludeTracks)
					}
					if *list.annotateListings != tt.wantAnnotateListings {
						t.Errorf("%s %q: got annotateListings %t want %t", fnName, name, *list.annotateListings, tt.wantAnnotateListings)
					}
					if *list.trackSorting != tt.wantTrackSorting {
						t.Errorf("%s %q: got trackSorting %q want %q", fnName, name, *list.trackSorting, tt.wantTrackSorting)
					}
				} else {
					t.Errorf("%s %q: error processing arguments", fnName, name)
				}
			}
		})
	}
}

func Test_list_outputTrackDiagnostics(t *testing.T) {
	const fnName = "list.outputTrackDiagnostics"
	badArtist := files.NewArtist("bad artist", "./BadArtist")
	badAlbum := files.NewAlbum("bad album", badArtist, "BadAlbum")
	badTrack := files.NewTrack(badAlbum, "01 bad track.mp3", "bad track", 1)
	makeList := func() *list {
		l, _ := newListCommand(output.NewNilBus(), internal.EmptyConfiguration(), flag.NewFlagSet("list", flag.ContinueOnError))
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
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	type args struct {
		t      *files.Track
		prefix string
	}
	tests := map[string]struct {
		name string
		l    *list
		args
		output.WantedRecording
	}{
		"error case": {
			l:    makeList(),
			args: args{t: badTrack},
			WantedRecording: output.WantedRecording{
				Error: "An error occurred when trying to read ID3V2 tag information for track \"bad track\" on album \"bad album\" by artist \"bad artist\": \"open BadAlbum\\\\01 bad track.mp3: The system cannot find the path specified.\".\n" +
					"An error occurred when trying to read ID3V1 tag information for track \"bad track\" on album \"bad album\" by artist \"bad artist\": \"open BadAlbum\\\\01 bad track.mp3: The system cannot find the path specified.\".\n",
				Log: "level='error' error='open BadAlbum\\01 bad track.mp3: The system cannot find the path specified.' track='BadAlbum\\01 bad track.mp3' msg='id3v2 tag error'\n" +
					"level='error' error='open BadAlbum\\01 bad track.mp3: The system cannot find the path specified.' track='BadAlbum\\01 bad track.mp3' msg='id3v1 tag error'\n",
			},
		},
		"success case": {
			l:    makeList(),
			args: args{t: goodTrack, prefix: "      "},
			WantedRecording: output.WantedRecording{
				Console: "      ID3V2 Version: 3\n" +
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
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.l.outputTrackDiagnostics(o, tt.args.t, tt.args.prefix)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_list_outputTrackDetails(t *testing.T) {
	const fnName = "list.outputTrackDetails()"
	badArtist := files.NewArtist("bad artist", "./BadArtist")
	badAlbum := files.NewAlbum("bad album", badArtist, "BadAlbum")
	badTrack := files.NewTrack(badAlbum, "01 bad track.mp3", "bad track", 1)
	makeList := func() *list {
		l, _ := newListCommand(output.NewNilBus(), internal.EmptyConfiguration(), flag.NewFlagSet("list", flag.ContinueOnError))
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
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	type args struct {
		t      *files.Track
		prefix string
	}
	tests := map[string]struct {
		l *list
		args
		output.WantedRecording
	}{
		"error case": {
			l:    makeList(),
			args: args{t: badTrack},
			WantedRecording: output.WantedRecording{
				Error: "The details are not available for track \"bad track\" on album \"bad album\" by artist \"bad artist\": \"open BadAlbum\\\\01 bad track.mp3: The system cannot find the path specified.\".\n",
				Log:   "level='error' error='open BadAlbum\\01 bad track.mp3: The system cannot find the path specified.' track='BadAlbum\\01 bad track.mp3' msg='cannot get details'\n",
			},
		},
		"success case": {
			l:    makeList(),
			args: args{t: goodTrack, prefix: "-->"},
			WantedRecording: output.WantedRecording{
				Console: "-->Details:\n" +
					"-->  Composer = \"a couple of idiots\"\n" +
					"-->  Conductor = \"Someone with a stick\"\n" +
					"-->  Key = \"D Major\"\n" +
					"-->  Lyricist = \"An infinite number of monkeys with a typewriter\"\n" +
					"-->  Orchestra/Band = \"The usual gang of idiots\"\n" +
					"-->  Subtitle = \"Part II\"\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.l.outputTrackDetails(o, tt.args.t, tt.args.prefix)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}
