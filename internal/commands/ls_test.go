package commands

import (
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
	fnName := "ls.validateTrackSorting()"
	tests := []struct {
		name              string
		sortingInput      string
		includeAlbums     bool
		wantSorting       string
		wantConsoleOutput string
		wantErrorOutput   string
		wantLogOutput     string
	}{
		{name: "alpha sorting with albums", sortingInput: "alpha", includeAlbums: true, wantSorting: "alpha"},
		{name: "alpha sorting without albums", sortingInput: "alpha", includeAlbums: false, wantSorting: "alpha"},
		{name: "numeric sorting with albums", sortingInput: "numeric", includeAlbums: true, wantSorting: "numeric"},
		{
			name:            "numeric sorting without albums",
			sortingInput:    "numeric",
			includeAlbums:   false,
			wantSorting:     "alpha",
			wantErrorOutput: "The value of the -sort flag, 'numeric', cannot be used unless '-includeAlbums' is true; track sorting will be alphabetic.\n",
			wantLogOutput:   "level='warn' -includeAlbums='false' -sort='numeric' msg='numeric track sorting is not applicable'\n",
		},
		{
			name:            "invalid sorting with albums",
			sortingInput:    "nonsense",
			includeAlbums:   true,
			wantSorting:     "numeric",
			wantErrorOutput: "The \"-sort\" value you specified, \"nonsense\", is not valid.\n",
			wantLogOutput:   "level='warn' -sort='nonsense' command='ls' msg='flag value is not valid'\n",
		},
		{
			name:            "invalid sorting without albums",
			sortingInput:    "nonsense",
			includeAlbums:   false,
			wantSorting:     "alpha",
			wantErrorOutput: "The \"-sort\" value you specified, \"nonsense\", is not valid.\n",
			wantLogOutput:   "level='warn' -sort='nonsense' command='ls' msg='flag value is not valid'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("ls", flag.ContinueOnError)
			o := internal.NewOutputDeviceForTesting()
			lsCommand := newLsCommand(internal.EmptyConfiguration(), fs)
			lsCommand.trackSorting = &tt.sortingInput
			lsCommand.includeAlbums = &tt.includeAlbums
			lsCommand.validateTrackSorting(o)
			if *lsCommand.trackSorting != tt.wantSorting {
				t.Errorf("%s: got %q, want %q", fnName, *lsCommand.trackSorting, tt.wantSorting)
			}
			if gotConsoleOutput := o.ConsoleOutput(); gotConsoleOutput != tt.wantConsoleOutput {
				t.Errorf("%s: console output = %q, want %q", fnName, gotConsoleOutput, tt.wantConsoleOutput)
			}
			if gotErrorOutput := o.ErrorOutput(); gotErrorOutput != tt.wantErrorOutput {
				t.Errorf("%s: error output = %q, want %q", fnName, gotErrorOutput, tt.wantErrorOutput)
			}
			if gotLogOutput := o.LogOutput(); gotLogOutput != tt.wantLogOutput {
				t.Errorf("%s: log output = %q, want %q", fnName, gotLogOutput, tt.wantLogOutput)
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

func Test_ls_Exec(t *testing.T) {
	fnName := "ls.Exec()"
	// generate test data
	topDir := "loadTest"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, topDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	if err := internal.PopulateTopDirForTesting(topDir); err != nil {
		t.Errorf("%s error populating %s: %v", fnName, topDir, err)
	}
	type args struct {
		args []string
	}
	tests := []struct {
		name              string
		l                 *ls
		args              args
		wantConsoleOutput string
		wantErrorOutput   string
		wantLogOutput     string
	}{
		{
			name: "no output",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=false",
					"-includeArtists=false",
					"-includeTracks=false",
				},
			},
			wantConsoleOutput: generateListing(false, false, false, false, false),
			wantErrorOutput:   "You disabled all functionality for the command \"ls\".\n",
			wantLogOutput:     "level='warn' -annotate='false' -includeAlbums='false' -includeArtists='false' -includeTracks='false' -sort='numeric' command='ls' msg='the user disabled all functionality'\n",
		},
		// tracks only
		{
			name: "unannotated tracks only",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=false",
					"-includeArtists=false",
					"-includeTracks=true",
					"-annotate=false",
				},
			},
			wantConsoleOutput: generateListing(false, false, true, false, false),
			wantErrorOutput:   "The value of the -sort flag, 'numeric', cannot be used unless '-includeAlbums' is true; track sorting will be alphabetic.\n",
			wantLogOutput: "level='info' -annotate='false' -includeAlbums='false' -includeArtists='false' -includeTracks='true' -sort='numeric' command='ls' msg='executing command'\n" +
				"level='warn' -includeAlbums='false' -sort='numeric' msg='numeric track sorting is not applicable'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
		},
		{
			name: "annotated tracks only",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=false",
					"-includeArtists=false",
					"-includeTracks=true",
					"-annotate=true",
				},
			},
			wantConsoleOutput: generateListing(false, false, true, true, false),
			wantErrorOutput:   "The value of the -sort flag, 'numeric', cannot be used unless '-includeAlbums' is true; track sorting will be alphabetic.\n",
			wantLogOutput: "level='info' -annotate='true' -includeAlbums='false' -includeArtists='false' -includeTracks='true' -sort='numeric' command='ls' msg='executing command'\n" +
				"level='warn' -includeAlbums='false' -sort='numeric' msg='numeric track sorting is not applicable'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
		},
		{
			name: "unannotated tracks only with numeric sorting",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
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
			wantConsoleOutput: generateListing(false, false, true, false, true),
			wantErrorOutput:   "The value of the -sort flag, 'numeric', cannot be used unless '-includeAlbums' is true; track sorting will be alphabetic.\n",
			wantLogOutput: "level='info' -annotate='false' -includeAlbums='false' -includeArtists='false' -includeTracks='true' -sort='numeric' command='ls' msg='executing command'\n" +
				"level='warn' -includeAlbums='false' -sort='numeric' msg='numeric track sorting is not applicable'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
		},
		{
			name: "annotated tracks only with numeric sorting",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
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
			wantConsoleOutput: generateListing(false, false, true, true, true),
			wantErrorOutput:   "The value of the -sort flag, 'numeric', cannot be used unless '-includeAlbums' is true; track sorting will be alphabetic.\n",
			wantLogOutput: "level='info' -annotate='true' -includeAlbums='false' -includeArtists='false' -includeTracks='true' -sort='numeric' command='ls' msg='executing command'\n" +
				"level='warn' -includeAlbums='false' -sort='numeric' msg='numeric track sorting is not applicable'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
		},
		// albums only
		{
			name: "unannotated albums only",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=true",
					"-includeArtists=false",
					"-includeTracks=false",
					"-annotate=false",
				},
			},
			wantConsoleOutput: generateListing(false, true, false, false, false),
			wantLogOutput: "level='info' -annotate='false' -includeAlbums='true' -includeArtists='false' -includeTracks='false' -sort='numeric' command='ls' msg='executing command'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
		},
		{
			name: "annotated albums only",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=true",
					"-includeArtists=false",
					"-includeTracks=false",
					"-annotate=true",
				},
			},
			wantConsoleOutput: generateListing(false, true, false, true, false),
			wantLogOutput: "level='info' -annotate='true' -includeAlbums='true' -includeArtists='false' -includeTracks='false' -sort='numeric' command='ls' msg='executing command'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
		},
		// artists only
		{
			name: "unannotated artists only",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=false",
					"-includeArtists=true",
					"-includeTracks=false",
					"-annotate=false",
				},
			},
			wantConsoleOutput: generateListing(true, false, false, false, false),
			wantLogOutput: "level='info' -annotate='false' -includeAlbums='false' -includeArtists='true' -includeTracks='false' -sort='numeric' command='ls' msg='executing command'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
		},
		{
			name: "annotated artists only",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=false",
					"-includeArtists=true",
					"-includeTracks=false",
					"-annotate=true",
				},
			},
			wantConsoleOutput: generateListing(true, false, false, true, false),
			wantLogOutput: "level='info' -annotate='true' -includeAlbums='false' -includeArtists='true' -includeTracks='false' -sort='numeric' command='ls' msg='executing command'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
		},
		// albums and artists
		{
			name: "unannotated artists and albums only",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=true",
					"-includeArtists=true",
					"-includeTracks=false",
					"-annotate=false",
				},
			},
			wantConsoleOutput: generateListing(true, true, false, false, false),
			wantLogOutput: "level='info' -annotate='false' -includeAlbums='true' -includeArtists='true' -includeTracks='false' -sort='numeric' command='ls' msg='executing command'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
		},
		{
			name: "annotated artists and albums only",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
			args: args{
				[]string{
					"-topDir", topDir,
					"-includeAlbums=true",
					"-includeArtists=true",
					"-includeTracks=false",
					"-annotate=true",
				},
			},
			wantConsoleOutput: generateListing(true, true, false, true, false),
			wantLogOutput: "level='info' -annotate='true' -includeAlbums='true' -includeArtists='true' -includeTracks='false' -sort='numeric' command='ls' msg='executing command'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
		},
		// albums and tracks
		{
			name: "unannotated albums and tracks with alpha sorting",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
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
			wantConsoleOutput: generateListing(false, true, true, false, false),
			wantLogOutput: "level='info' -annotate='false' -includeAlbums='true' -includeArtists='false' -includeTracks='true' -sort='alpha' command='ls' msg='executing command'\n" +
				"level='info' -annotate='false' -includeAlbums='true' -includeArtists='false' -includeTracks='true' -sort='alpha' command='ls' msg='one or more flags were overridden'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
		},
		{
			name: "annotated albums and tracks with alpha sorting",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
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
			wantConsoleOutput: generateListing(false, true, true, true, false),
			wantLogOutput: "level='info' -annotate='true' -includeAlbums='true' -includeArtists='false' -includeTracks='true' -sort='alpha' command='ls' msg='executing command'\n" +
				"level='info' -annotate='true' -includeAlbums='true' -includeArtists='false' -includeTracks='true' -sort='alpha' command='ls' msg='one or more flags were overridden'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
		},
		{
			name: "unannotated albums and tracks with numeric sorting",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
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
			wantConsoleOutput: generateListing(false, true, true, false, true),
			wantLogOutput: "level='info' -annotate='false' -includeAlbums='true' -includeArtists='false' -includeTracks='true' -sort='numeric' command='ls' msg='executing command'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
		},
		{
			name: "annotated albums and tracks with numeric sorting",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
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
			wantConsoleOutput: generateListing(false, true, true, true, true),
			wantLogOutput: "level='info' -annotate='true' -includeAlbums='true' -includeArtists='false' -includeTracks='true' -sort='numeric' command='ls' msg='executing command'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
		},
		// artists and tracks
		{
			name: "unannotated artists and tracks with alpha sorting",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
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
			wantConsoleOutput: generateListing(true, false, true, false, false),
			wantLogOutput: "level='info' -annotate='false' -includeAlbums='false' -includeArtists='true' -includeTracks='true' -sort='alpha' command='ls' msg='executing command'\n" +
				"level='info' -annotate='false' -includeAlbums='false' -includeArtists='true' -includeTracks='true' -sort='alpha' command='ls' msg='one or more flags were overridden'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
		},
		{
			name: "annotated artists and tracks with alpha sorting",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
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
			wantConsoleOutput: generateListing(true, false, true, true, false),
			wantLogOutput: "level='info' -annotate='true' -includeAlbums='false' -includeArtists='true' -includeTracks='true' -sort='alpha' command='ls' msg='executing command'\n" +
				"level='info' -annotate='true' -includeAlbums='false' -includeArtists='true' -includeTracks='true' -sort='alpha' command='ls' msg='one or more flags were overridden'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
		},
		{
			name: "unannotated artists and tracks with numeric sorting",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
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
			wantConsoleOutput: generateListing(true, false, true, false, true),
			wantErrorOutput:   "The value of the -sort flag, 'numeric', cannot be used unless '-includeAlbums' is true; track sorting will be alphabetic.\n",
			wantLogOutput: "level='info' -annotate='false' -includeAlbums='false' -includeArtists='true' -includeTracks='true' -sort='numeric' command='ls' msg='executing command'\n" +
				"level='warn' -includeAlbums='false' -sort='numeric' msg='numeric track sorting is not applicable'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
		},
		{
			name: "annotated artists and tracks with numeric sorting",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
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
			wantConsoleOutput: generateListing(true, false, true, true, true),
			wantErrorOutput:   "The value of the -sort flag, 'numeric', cannot be used unless '-includeAlbums' is true; track sorting will be alphabetic.\n",
			wantLogOutput: "level='info' -annotate='true' -includeAlbums='false' -includeArtists='true' -includeTracks='true' -sort='numeric' command='ls' msg='executing command'\n" +
				"level='warn' -includeAlbums='false' -sort='numeric' msg='numeric track sorting is not applicable'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
		},
		// albums, artists, and tracks
		{
			name: "unannotated artists, albums, and tracks with alpha sorting",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
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
			wantConsoleOutput: generateListing(true, true, true, false, false),
			wantLogOutput: "level='info' -annotate='false' -includeAlbums='true' -includeArtists='true' -includeTracks='true' -sort='alpha' command='ls' msg='executing command'\n" +
				"level='info' -annotate='false' -includeAlbums='true' -includeArtists='true' -includeTracks='true' -sort='alpha' command='ls' msg='one or more flags were overridden'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
		},
		{
			name: "annotated artists, albums, and tracks with alpha sorting",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
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
			wantConsoleOutput: generateListing(true, true, true, true, false),
			wantLogOutput: "level='info' -annotate='true' -includeAlbums='true' -includeArtists='true' -includeTracks='true' -sort='alpha' command='ls' msg='executing command'\n" +
				"level='info' -annotate='true' -includeAlbums='true' -includeArtists='true' -includeTracks='true' -sort='alpha' command='ls' msg='one or more flags were overridden'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
		},
		{
			name: "unannotated artists, albums, and tracks with numeric sorting",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
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
			wantConsoleOutput: generateListing(true, true, true, false, true),
			wantLogOutput: "level='info' -annotate='false' -includeAlbums='true' -includeArtists='true' -includeTracks='true' -sort='numeric' command='ls' msg='executing command'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
		},
		{
			name: "annotated artists, albums, and tracks with numeric sorting",
			l:    newLsCommand(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ContinueOnError)),
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
			wantConsoleOutput: generateListing(true, true, true, true, true),
			wantLogOutput: "level='info' -annotate='true' -includeAlbums='true' -includeArtists='true' -includeTracks='true' -sort='numeric' command='ls' msg='executing command'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			tt.l.Exec(o, tt.args.args)
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

func Test_newLsCommand(t *testing.T) {
	savedState := internal.SaveEnvVarForTesting("APPDATA")
	os.Setenv("APPDATA", internal.SecureAbsolutePathForTesting("."))
	defer func() {
		savedState.RestoreForTesting()
	}()
	topDir := "loadTest"
	fnName := "newLsCommand()"
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
		name                 string
		args                 args
		wantIncludeAlbums    bool
		wantIncludeArtists   bool
		wantIncludeTracks    bool
		wantTrackSorting     string
		wantAnnotateListings bool
	}{
		{
			name:                 "ordinary defaults",
			args:                 args{c: internal.EmptyConfiguration()},
			wantIncludeAlbums:    true,
			wantIncludeArtists:   true,
			wantIncludeTracks:    false,
			wantTrackSorting:     "numeric",
			wantAnnotateListings: false,
		},
		{
			name:                 "overridden defaults",
			args:                 args{c: defaultConfig},
			wantIncludeAlbums:    false,
			wantIncludeArtists:   false,
			wantIncludeTracks:    true,
			wantTrackSorting:     "alpha",
			wantAnnotateListings: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := newLsCommand(tt.args.c, flag.NewFlagSet("ls", flag.ContinueOnError))
			if _, ok := ls.sf.ProcessArgs(internal.NewOutputDeviceForTesting(), []string{
				"-topDir", topDir,
				"-ext", ".mp3",
			}); ok {
				if *ls.includeAlbums != tt.wantIncludeAlbums {
					t.Errorf("%s %s: got includeAlbums %t want %t", fnName, tt.name, *ls.includeAlbums, tt.wantIncludeAlbums)
				}
				if *ls.includeArtists != tt.wantIncludeArtists {
					t.Errorf("%s %s: got includeArtists %t want %t", fnName, tt.name, *ls.includeArtists, tt.wantIncludeArtists)
				}
				if *ls.includeTracks != tt.wantIncludeTracks {
					t.Errorf("%s %s: got includeTracks %t want %t", fnName, tt.name, *ls.includeTracks, tt.wantIncludeTracks)
				}
				if *ls.annotateListings != tt.wantAnnotateListings {
					t.Errorf("%s %s: got annotateListings %t want %t", fnName, tt.name, *ls.annotateListings, tt.wantAnnotateListings)
				}
				if *ls.trackSorting != tt.wantTrackSorting {
					t.Errorf("%s %s: got trackSorting %q want %q", fnName, tt.name, *ls.trackSorting, tt.wantTrackSorting)
				}
			} else {
				t.Errorf("%s %s: error processing arguments", fnName, tt.name)
			}
		})
	}
}
