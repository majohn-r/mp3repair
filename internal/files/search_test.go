package files

import (
	"bytes"
	"flag"
	"io/fs"
	"mp3/internal"
	"reflect"
	"testing"
)

func Test_readDirectory(t *testing.T) {
	fnName := "readDirectory()"
	// generate test data
	topDir := "loadTest"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, topDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	type args struct {
		dir string
	}
	tests := []struct {
		name      string
		args      args
		wantFiles []fs.FileInfo
		wantOk    bool
		wantWErr  string
	}{
		{name: "default", args: args{topDir}, wantFiles: []fs.FileInfo{}, wantOk: true},
		{
			name:     "non-existent dir",
			args:     args{"non-existent directory"},
			wantWErr: "The directory \"non-existent directory\" cannot be read: open non-existent directory: The system cannot find the file specified.\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wErr := &bytes.Buffer{}
			gotFiles, gotOk := readDirectory(wErr, tt.args.dir)
			if !reflect.DeepEqual(gotFiles, tt.wantFiles) {
				t.Errorf("readDirectory() gotFiles = %v, want %v", gotFiles, tt.wantFiles)
			}
			if gotOk != tt.wantOk {
				t.Errorf("readDirectory() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if gotWErr := wErr.String(); gotWErr != tt.wantWErr {
				t.Errorf("readDirectory() gotWErr = %v, want %v", gotWErr, tt.wantWErr)
			}
		})
	}
}

func TestSearch_FilterArtists(t *testing.T) {
	fnName := "Search.FilterArtists()"
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
	realFlagSet := flag.NewFlagSet("real", flag.ContinueOnError)
	realS, _ := NewSearchFlags(internal.EmptyConfiguration(), realFlagSet).ProcessArgs(
		internal.NewOutputDeviceForTesting(), []string{"-topDir", topDir})
	realArtists, _ := realS.LoadData(internal.NewOutputDeviceForTesting(), internal.NewOutputDevice().LogWriter(), internal.NewOutputDevice().ErrorWriter())
	overFilteredS, _ := NewSearchFlags(internal.EmptyConfiguration(),
		flag.NewFlagSet("overFiltered", flag.ContinueOnError)).ProcessArgs(
		internal.NewOutputDeviceForTesting(), []string{"-topDir", topDir, "-artistFilter", "^Filter all out$"})
	a, _ := realS.LoadUnfilteredData(internal.NewOutputDeviceForTesting())
	type args struct {
		unfilteredArtists []*Artist
	}
	tests := []struct {
		name              string
		s                 *Search
		args              args
		wantArtists       []*Artist
		wantOk            bool
		wantConsoleOutput string
		wantErrorOutput   string
		wantLogOutput     string
	}{
		{
			name:        "default",
			s:           realS,
			args:        args{unfilteredArtists: a},
			wantArtists: realArtists,
			wantOk:      true,
		},
		{
			name:            "all filtered out",
			s:               overFilteredS,
			args:            args{unfilteredArtists: a},
			wantArtists:     nil,
			wantOk:          false,
			wantErrorOutput: "No music files could be found using the specified parameters.\n",
			wantLogOutput:   "level='warn' -albumFilter='.*' -artistFilter='^Filter all out$' -ext='.mp3' -topDir='loadTest' msg='cannot find any artist directories'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			gotArtists, gotOk := tt.s.FilterArtists(o, tt.args.unfilteredArtists)
			if !reflect.DeepEqual(gotArtists, tt.wantArtists) {
				t.Errorf("%s = %v, want %v", fnName, gotArtists, tt.wantArtists)
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

func TestSearch_LoadData(t *testing.T) {
	fnName := "Search.LoadData()"
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
	tests := []struct {
		name              string
		s                 *Search
		wantArtists       []*Artist
		wantConsoleOutput string
		wantErrorOutput   string
		wantLogOutput     string
		wantOk            bool
	}{
		{
			name:        "read all",
			s:           CreateFilteredSearchForTesting(topDir, "^.*$", "^.*$"),
			wantArtists: CreateAllArtistsForTesting(topDir, false),
			wantOk:      true,
		},
		{
			name:        "read with filtering",
			s:           CreateFilteredSearchForTesting(topDir, "^.*[13579]$", "^.*[02468]$"),
			wantArtists: CreateAllOddArtistsWithEvenAlbumsForTesting(topDir),
			wantOk:      true,
		},
		{
			name:            "read with all artists filtered out",
			s:               CreateFilteredSearchForTesting(topDir, "^.*X$", "^.*$"),
			wantErrorOutput: "No music files could be found using the specified parameters.\n",
			wantLogOutput:   "level='warn' -albumFilter='^.*$' -artistFilter='^.*X$' -ext='.mp3' -topDir='loadTest' msg='cannot find any artist directories'\n",
		},
		{
			name:            "read with all albums filtered out",
			s:               CreateFilteredSearchForTesting(topDir, "^.*$", "^.*X$"),
			wantErrorOutput: "No music files could be found using the specified parameters.\n",
			wantLogOutput:   "level='warn' -albumFilter='^.*X$' -artistFilter='^.*$' -ext='.mp3' -topDir='loadTest' msg='cannot find any artist directories'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			gotArtists, gotOk := tt.s.LoadData(o, o.LogWriter(), o.ErrorWriter())
			if !reflect.DeepEqual(gotArtists, tt.wantArtists) {
				t.Errorf("Search.LoadData() = %v, want %v", gotArtists, tt.wantArtists)
			}
			if gotConsoleOutput := o.ConsoleOutput(); gotConsoleOutput != tt.wantConsoleOutput {
				t.Errorf("Search.LoadData() console output = %q, want %q", gotConsoleOutput, tt.wantConsoleOutput)
			}
			if gotErrorOutput := o.ErrorOutput(); gotErrorOutput != tt.wantErrorOutput {
				t.Errorf("Search.LoadData() error output = %q, want %q", gotErrorOutput, tt.wantErrorOutput)
			}
			if gotLogOutput := o.LogOutput(); gotLogOutput != tt.wantLogOutput {
				t.Errorf("Search.LoadData() log output = %q, want %q", gotLogOutput, tt.wantLogOutput)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Search.LoadData() ok = %v, want %v", gotOk, tt.wantOk)
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
		t.Errorf("%s error creating %s: %v", fnName, topDir, err)
	}
	if err := internal.PopulateTopDirForTesting(topDir); err != nil {
		t.Errorf("%s error populating %s: %v", fnName, topDir, err)
	}
	if err := internal.Mkdir(emptyDir); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, emptyDir, err)
	}
	tests := []struct {
		name              string
		s                 *Search
		wantArtists       []*Artist
		wantConsoleOutput string
		wantLogOutput     string
		wantErrorOutput   string
		wantOk            bool
	}{
		{
			name:        "read all",
			s:           CreateSearchForTesting(topDir),
			wantArtists: CreateAllArtistsForTesting(topDir, true),
			wantOk:      true,
		},
		{
			name:            "empty dir",
			s:               CreateSearchForTesting(emptyDir),
			wantErrorOutput: "No music files could be found using the specified parameters.\n",
			wantLogOutput:   "level='warn' -ext='.mp3' -topDir='empty directory' msg='cannot find any artist directories'\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotArtists []*Artist
			var gotOk bool
			o := internal.NewOutputDeviceForTesting()
			if gotArtists, gotOk = tt.s.LoadUnfilteredData(o); !reflect.DeepEqual(gotArtists, tt.wantArtists) {
				t.Errorf("Search.LoadUnfilteredData() = %v, want %v", gotArtists, tt.wantArtists)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Search.LoadUnfilteredData() ok = %v, want %v", gotOk, tt.wantOk)
			}
			if gotConsoleOutput := o.ConsoleOutput(); gotConsoleOutput != tt.wantConsoleOutput {
				t.Errorf("Search.LoadUnfilteredData() console output = %q, want %q", gotConsoleOutput, tt.wantConsoleOutput)
			}
			if gotErrorOutput := o.ErrorOutput(); gotErrorOutput != tt.wantErrorOutput {
				t.Errorf("Search.LoadUnfilteredData() error output = %q, want %q", gotErrorOutput, tt.wantErrorOutput)
			}
			if gotLogOutput := o.LogOutput(); gotLogOutput != tt.wantLogOutput {
				t.Errorf("Search.LoadUnfilteredData() log output = %q, want %q", gotLogOutput, tt.wantLogOutput)
			}
		})
	}
}
