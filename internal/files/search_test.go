package files

import (
	"flag"
	"mp3/internal"
	"reflect"
	"testing"

	"github.com/majohn-r/output"
)

func TestSearch_FilterArtists(t *testing.T) {
	fnName := "Search.FilterArtists()"
	// generate test data
	topDir := "loadTest"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, topDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	if err := internal.PopulateTopDirForTesting(topDir); err != nil {
		t.Errorf("%s error populating %q: %v", fnName, topDir, err)
	}
	realFlagSet := flag.NewFlagSet("real", flag.ContinueOnError)
	realSF, _ := NewSearchFlags(output.NewNilBus(), internal.EmptyConfiguration(), realFlagSet)
	realS, _ := realSF.ProcessArgs(output.NewNilBus(), []string{"-topDir", topDir})
	realArtists, _ := realS.LoadData(output.NewNilBus())
	overFilteredSF, _ := NewSearchFlags(
		output.NewNilBus(),
		internal.EmptyConfiguration(),
		flag.NewFlagSet("overFiltered", flag.ContinueOnError))
	overFilteredS, _ := overFilteredSF.ProcessArgs(
		output.NewNilBus(), []string{"-topDir", topDir, "-artistFilter", "^Filter all out$"})
	a, _ := realS.LoadUnfilteredData(output.NewNilBus())
	type args struct {
		unfilteredArtists []*Artist
	}
	tests := []struct {
		name string
		s    *Search
		args
		wantArtists []*Artist
		wantOk      bool
		output.WantedRecording
	}{
		{
			name:        "default",
			s:           realS,
			args:        args{unfilteredArtists: a},
			wantArtists: realArtists,
			wantOk:      true,
			WantedRecording: output.WantedRecording{
				Log: "level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='loadTest' msg='filtering music files'\n",
			},
		},
		{
			name: "all filtered out",
			s:    overFilteredS,
			args: args{unfilteredArtists: a},
			WantedRecording: output.WantedRecording{
				Error: "No music files could be found using the specified parameters.\n",
				Log: "level='info' -albumFilter='.*' -artistFilter='^Filter all out$' -ext='.mp3' -topDir='loadTest' msg='filtering music files'\n" +
					"level='error' -albumFilter='.*' -artistFilter='^Filter all out$' -ext='.mp3' -topDir='loadTest' msg='cannot find any artist directories'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestSearch_LoadData(t *testing.T) {
	fnName := "Search.LoadData()"
	// generate test data
	topDir := "loadTest"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, topDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	if err := internal.PopulateTopDirForTesting(topDir); err != nil {
		t.Errorf("%s error populating %q: %v", fnName, topDir, err)
	}
	tests := []struct {
		name        string
		s           *Search
		wantArtists []*Artist
		wantOk      bool
		output.WantedRecording
	}{
		{
			name:        "read all",
			s:           CreateFilteredSearchForTesting(topDir, "^.*$", "^.*$"),
			wantArtists: CreateAllArtistsForTesting(topDir, false),
			wantOk:      true,
			WantedRecording: output.WantedRecording{
				Log: "level='info' -albumFilter='^.*$' -artistFilter='^.*$' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		{
			name:        "read with filtering",
			s:           CreateFilteredSearchForTesting(topDir, "^.*[13579]$", "^.*[02468]$"),
			wantArtists: CreateAllOddArtistsWithEvenAlbumsForTesting(topDir),
			wantOk:      true,
			WantedRecording: output.WantedRecording{
				Log: "level='info' -albumFilter='^.*[02468]$' -artistFilter='^.*[13579]$' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n",
			},
		},
		{
			name: "read with all artists filtered out",
			s:    CreateFilteredSearchForTesting(topDir, "^.*X$", "^.*$"),
			WantedRecording: output.WantedRecording{
				Error: "No music files could be found using the specified parameters.\n",
				Log: "level='info' -albumFilter='^.*$' -artistFilter='^.*X$' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n" +
					"level='error' -albumFilter='^.*$' -artistFilter='^.*X$' -ext='.mp3' -topDir='loadTest' msg='cannot find any artist directories'\n",
			},
		},
		{
			name: "read with all albums filtered out",
			s:    CreateFilteredSearchForTesting(topDir, "^.*$", "^.*X$"),
			WantedRecording: output.WantedRecording{
				Error: "No music files could be found using the specified parameters.\n",
				Log: "level='info' -albumFilter='^.*X$' -artistFilter='^.*$' -ext='.mp3' -topDir='loadTest' msg='reading filtered music files'\n" +
					"level='error' -albumFilter='^.*X$' -artistFilter='^.*$' -ext='.mp3' -topDir='loadTest' msg='cannot find any artist directories'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			gotArtists, gotOk := tt.s.LoadData(o)
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

func TestSearch_LoadUnfilteredData(t *testing.T) {
	fnName := "Search.LoadUnfilteredData()"
	// generate test data
	topDir := "loadTest"
	emptyDir := "empty directory"
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
		internal.DestroyDirectoryForTesting(fnName, emptyDir)
	}()
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, topDir, err)
	}
	if err := internal.PopulateTopDirForTesting(topDir); err != nil {
		t.Errorf("%s error populating %q: %v", fnName, topDir, err)
	}
	if err := internal.Mkdir(emptyDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, emptyDir, err)
	}
	tests := []struct {
		name        string
		s           *Search
		wantArtists []*Artist
		wantOk      bool
		output.WantedRecording
	}{
		{
			name:        "read all",
			s:           CreateSearchForTesting(topDir),
			wantArtists: CreateAllArtistsForTesting(topDir, true),
			wantOk:      true,
			WantedRecording: output.WantedRecording{
				Log: "level='info' -ext='.mp3' -topDir='loadTest' msg='reading unfiltered music files'\n",
			},
		},
		{
			name: "empty dir",
			s:    CreateSearchForTesting(emptyDir),
			WantedRecording: output.WantedRecording{
				Error: "No music files could be found using the specified parameters.\n",
				Log: "level='info' -ext='.mp3' -topDir='empty directory' msg='reading unfiltered music files'\n" +
					"level='error' -ext='.mp3' -topDir='empty directory' msg='cannot find any artist directories'\n"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotArtists []*Artist
			var gotOk bool
			o := output.NewRecorder()
			if gotArtists, gotOk = tt.s.LoadUnfilteredData(o); !reflect.DeepEqual(gotArtists, tt.wantArtists) {
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
