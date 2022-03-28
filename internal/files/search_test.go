package files

import (
	"flag"
	"io/fs"
	"mp3/internal"
	"os"
	"reflect"
	"testing"
)

func TestSearch_TopDirectory(t *testing.T) {
	tests := []struct {
		name string
		s    *Search
		want string
	}{
		{name: "expected", s: &Search{topDirectory: "check"}, want: "check"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.TopDirectory(); got != tt.want {
				t.Errorf("Search.TopDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearch_TargetExtension(t *testing.T) {
	tests := []struct {
		name string
		s    *Search
		want string
	}{
		{name: "expected", s: &Search{targetExtension: ".txt"}, want: ".txt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.TargetExtension(); got != tt.want {
				t.Errorf("Search.TargetExtension() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearch_LoadUnfilteredData(t *testing.T) {
	// generate test data
	topDir := "loadTest"
	if err := internal.Mkdir(t, "Search.LoadUnfilteredData", topDir); err != nil {
		return
	}
	defer func() {
		if err := os.RemoveAll(topDir); err != nil {
			t.Errorf("Search.LoadUnfilteredData error destroying test directory %q: %v", topDir, err)
		}
	}()
	internal.PopulateTopDir(t, topDir)
	emptyDir := "empty directory"
	if err := internal.Mkdir(t, "Search.LoadUnfilteredData", emptyDir); err != nil {
		return
	}
	defer func() {
		if err := os.RemoveAll(emptyDir); err != nil {
			t.Errorf("Search.LoadUnfilteredData error destroying test directory %q: %v", emptyDir, err)
		}
	}()
	realFlagSet := flag.NewFlagSet("real", flag.ContinueOnError)
	emptyFlagSet := flag.NewFlagSet("empty", flag.ContinueOnError)
	tests := []struct {
		name        string
		s           *Search
		wantArtists []*Artist
	}{
		{
			name:        "read all",
			s:           NewSearchFlags(realFlagSet).ProcessArgs(os.Stdout, []string{"-topDir", topDir}),
			wantArtists: CreateAllArtists(topDir, true),
		},
		{
			name: "empty dir",
			s:    NewSearchFlags(emptyFlagSet).ProcessArgs(os.Stdout, []string{"-topDir", emptyDir}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotArtists := tt.s.LoadUnfilteredData(); !reflect.DeepEqual(gotArtists, tt.wantArtists) {
				t.Errorf("Search.LoadUnfilteredData() = %v, want %v", gotArtists, tt.wantArtists)
			}
		})
	}
}

func TestSearch_FilterArtists(t *testing.T) {
	// generate test data
	topDir := "loadTest"
	if err := internal.Mkdir(t, "LoadData", topDir); err != nil {
		return
	}
	defer func() {
		if err := os.RemoveAll(topDir); err != nil {
			t.Errorf("TestFilterArtists() error destroying test directory %q: %v", topDir, err)
		}
	}()
	internal.PopulateTopDir(t, topDir)
	realFlagSet := flag.NewFlagSet("real", flag.ContinueOnError)
	realS := NewSearchFlags(realFlagSet).ProcessArgs(os.Stdout, []string{"-topDir", topDir})
	type args struct {
		unfilteredArtists []*Artist
	}
	tests := []struct {
		name        string
		s           *Search
		args        args
		wantArtists []*Artist
	}{
		{name: "default", s: realS, args: args{unfilteredArtists: realS.LoadUnfilteredData()}, wantArtists: realS.LoadData()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotArtists := tt.s.FilterArtists(tt.args.unfilteredArtists); !reflect.DeepEqual(gotArtists, tt.wantArtists) {
				t.Errorf("Search.FilterArtists() = %v, want %v", gotArtists, tt.wantArtists)
			}
		})
	}
}

func TestSearch_LoadData(t *testing.T) {
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
	internal.PopulateTopDir(t, topDir)
	fsCase1 := flag.NewFlagSet("case1", flag.ContinueOnError)
	fsCase2 := flag.NewFlagSet("case2", flag.ContinueOnError)
	fsCase3 := flag.NewFlagSet("case3", flag.ContinueOnError)
	fsCase4 := flag.NewFlagSet("case4", flag.ContinueOnError)
	tests := []struct {
		name        string
		s           *Search
		wantArtists []*Artist
	}{
		{
			name:        "read all",
			s:           NewSearchFlags(fsCase1).ProcessArgs(os.Stdout, []string{"-topDir", topDir, "-albums", "^.*$", "-artists", "^.*$"}),
			wantArtists: CreateAllArtists(topDir, false),
		},
		{
			name:        "read with filtering",
			s:           NewSearchFlags(fsCase2).ProcessArgs(os.Stdout, []string{"-topDir", topDir, "-albums", "^.*[02468]$", "-artists", "^.*[13579]$"}),
			wantArtists: CreateAllOddArtistsWithEvenAlbums(topDir),
		},
		{
			name: "read with all artists filtered out",
			s:    NewSearchFlags(fsCase3).ProcessArgs(os.Stdout, []string{"-topDir", topDir, "-albums", "^.*$", "-artists", "^.*X$"}),
		},
		{
			name: "read with all albums filtered out",
			s:    NewSearchFlags(fsCase4).ProcessArgs(os.Stdout, []string{"-topDir", topDir, "-albums", "^.*X$", "-artists", "^.*X$"}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotArtists := tt.s.LoadData(); !reflect.DeepEqual(gotArtists, tt.wantArtists) {
				t.Errorf("Search.LoadData() = %v, want %v", gotArtists, tt.wantArtists)
			}
		})
	}
}

func Test_readDirectory(t *testing.T) {
	// generate test data
	topDir := "loadTest"
	if err := internal.Mkdir(t, "readDirectory", topDir); err != nil {
		return
	}
	defer func() {
		if err := os.RemoveAll(topDir); err != nil {
			t.Errorf("readDirectory() error destroying test directory %q: %v", topDir, err)
		}
	}()
	type args struct {
		dir string
	}
	tests := []struct {
		name      string
		args      args
		wantFiles []fs.FileInfo
		wantErr   bool
	}{
		{name: "default", args: args{topDir}, wantFiles: []fs.FileInfo{}, wantErr: false},
		{name: "non-existent dir", args: args{"non-existent directory"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFiles, err := readDirectory(tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("readDirectory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if !reflect.DeepEqual(gotFiles, tt.wantFiles) {
					t.Errorf("readDirectory() = %v, want %v", gotFiles, tt.wantFiles)
				}
			}
		})
	}
}
