package files

import (
	"flag"
	"mp3/internal"
	"path/filepath"
	"reflect"
	"testing"

	tools "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
)

func TestSearch_FilterArtists(t *testing.T) {
	const fnName = "Search.FilterArtists()"
	// generate test data
	topDir := "loadTest"
	if err := tools.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, topDir, err)
	}
	if err := internal.PopulateTopDirForTesting(topDir); err != nil {
		t.Errorf("%s error populating %q: %v", fnName, topDir, err)
	}
	o := output.NewNilBus()
	c := tools.EmptyConfiguration()
	realSF, _ := NewSearchFlags(o, c, flag.NewFlagSet("real", flag.ContinueOnError))
	realS, _ := realSF.ProcessArgs(o, []string{"-topDir", topDir})
	realArtists, _ := realS.Load(o)
	overFilteredSF, _ := NewSearchFlags(o, c, flag.NewFlagSet("overFiltered", flag.ContinueOnError))
	overFilteredS, _ := overFilteredSF.ProcessArgs(o, []string{"-topDir", topDir, "-artistFilter", "^Filter all out$"})
	a, _ := realS.LoadUnfiltered(o)
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	type args struct {
		unfilteredArtists []*Artist
	}
	tests := map[string]struct {
		s *Search
		args
		wantArtists []*Artist
		wantOk      bool
		output.WantedRecording
	}{
		"default": {
			s:           realS,
			args:        args{unfilteredArtists: a},
			wantArtists: realArtists,
			wantOk:      true,
			WantedRecording: output.WantedRecording{
				Log: "level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='filtering music files'\n",
			},
		},
		"all filtered out": {
			s:    overFilteredS,
			args: args{unfilteredArtists: a},
			WantedRecording: output.WantedRecording{
				Error: "No music files could be found using the specified parameters.\n",
				Log: "level='info' -albumFilter='.*' -artistFilter='^Filter all out$' -ext='.mp3' -topDir='loadTest' msg='filtering music files'\n" +
					"level='error' -albumFilter='.*' -artistFilter='^Filter all out$' -ext='.mp3' -topDir='loadTest' msg='cannot find any artist directories'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			gotArtists, gotOk := tt.s.FilterArtists(o, tt.args.unfilteredArtists)
			if !reflect.DeepEqual(gotArtists, tt.wantArtists) {
				t.Errorf("%s = %v, want %v", fnName, gotArtists, tt.wantArtists)
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

func TestSearch_Load(t *testing.T) {
	const fnName = "Search.Load()"
	// generate test data
	topDir := "loadTest"
	if err := tools.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, topDir, err)
	}
	if err := internal.PopulateTopDirForTesting(topDir); err != nil {
		t.Errorf("%s error populating %q: %v", fnName, topDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	tests := map[string]struct {
		s           *Search
		wantArtists []*Artist
		wantOk      bool
		output.WantedRecording
	}{
		"read all": {
			s:           CreateFilteredSearchForTesting(topDir, "^.*$", "^.*$"),
			wantArtists: CreateAllArtistsForTesting(topDir, false),
			wantOk:      true,
			WantedRecording: output.WantedRecording{
				Log: "level='info' -albumFilter='^.*$' -artistFilter='^.*$' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		"read with filtering": {
			s:           CreateFilteredSearchForTesting(topDir, "^.*[13579]$", "^.*[02468]$"),
			wantArtists: createAllOddArtistsWithEvenAlbums(topDir),
			wantOk:      true,
			WantedRecording: output.WantedRecording{
				Log: "level='info' -albumFilter='^.*[02468]$' -artistFilter='^.*[13579]$' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		"read with all artists filtered out": {
			s: CreateFilteredSearchForTesting(topDir, "^.*X$", "^.*$"),
			WantedRecording: output.WantedRecording{
				Error: "No music files could be found using the specified parameters.\n",
				Log: "level='info' -albumFilter='^.*$' -artistFilter='^.*X$' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n" +
					"level='error' -albumFilter='^.*$' -artistFilter='^.*X$' -ext='.mp3' -topDir='loadTest' msg='cannot find any artist directories'\n",
			},
		},
		"read with all albums filtered out": {
			s: CreateFilteredSearchForTesting(topDir, "^.*$", "^.*X$"),
			WantedRecording: output.WantedRecording{
				Error: "No music files could be found using the specified parameters.\n",
				Log: "level='info' -albumFilter='^.*X$' -artistFilter='^.*$' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n" +
					"level='error' -albumFilter='^.*X$' -artistFilter='^.*$' -ext='.mp3' -topDir='loadTest' msg='cannot find any artist directories'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			gotArtists, gotOk := tt.s.Load(o)
			if !reflect.DeepEqual(gotArtists, tt.wantArtists) {
				t.Errorf("%s = %v, want %v", fnName, gotArtists, tt.wantArtists)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s ok = %v, want %v", fnName, gotOk, tt.wantOk)
			}
		})
	}
}

func TestSearch_LoadUnfiltered(t *testing.T) {
	const fnName = "Search.LoadUnfiltered()"
	// generate test data
	topDir := "loadTest"
	emptyDir := "empty directory"
	if err := tools.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, topDir, err)
	}
	if err := internal.PopulateTopDirForTesting(topDir); err != nil {
		t.Errorf("%s error populating %q: %v", fnName, topDir, err)
	}
	if err := tools.Mkdir(emptyDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, emptyDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
		internal.DestroyDirectoryForTesting(fnName, emptyDir)
	}()
	tests := map[string]struct {
		s           *Search
		wantArtists []*Artist
		wantOk      bool
		output.WantedRecording
	}{
		"read all": {
			s:               CreateSearchForTesting(topDir),
			wantArtists:     CreateAllArtistsForTesting(topDir, true),
			wantOk:          true,
			WantedRecording: output.WantedRecording{Log: "level='info' -ext='.mp3' -topDir='loadTest' msg='reading unfiltered music files'\n"},
		},
		"empty dir": {
			s: CreateSearchForTesting(emptyDir),
			WantedRecording: output.WantedRecording{
				Error: "No music files could be found using the specified parameters.\n",
				Log: "level='info' -ext='.mp3' -topDir='empty directory' msg='reading unfiltered music files'\n" +
					"level='error' -ext='.mp3' -topDir='empty directory' msg='cannot find any artist directories'\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var gotArtists []*Artist
			var gotOk bool
			o := output.NewRecorder()
			if gotArtists, gotOk = tt.s.LoadUnfiltered(o); !reflect.DeepEqual(gotArtists, tt.wantArtists) {
				t.Errorf("%s = %v, want %v", fnName, gotArtists, tt.wantArtists)
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

func createAllOddArtistsWithEvenAlbums(topDir string) []*Artist {
	var artists []*Artist
	for k := 1; k < 10; k += 2 {
		artistName := internal.CreateArtistNameForTesting(k)
		artistDir := filepath.Join(topDir, artistName)
		artist := NewArtist(artistName, artistDir)
		for n := 0; n < 10; n += 2 {
			albumName := internal.CreateAlbumNameForTesting(n)
			albumDir := filepath.Join(artistDir, albumName)
			album := NewAlbum(albumName, artist, albumDir)
			for p := 0; p < 10; p++ {
				trackName := internal.CreateTrackNameForTesting(p)
				name, _, _ := parseTrackName(nil, trackName, album, defaultFileExtension)
				album.AddTrack(NewTrack(album, trackName, name, p))
			}
			artist.AddAlbum(album)
		}
		artists = append(artists, artist)
	}
	return artists
}
