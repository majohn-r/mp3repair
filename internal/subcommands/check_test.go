package subcommands

import (
	"bytes"
	"mp3/internal"
	"mp3/internal/files"
	"os"
	"reflect"
	"strings"
	"testing"
)

func Test_performEmptyFolderAnalysis(t *testing.T) {
	fnName := "performEmptyFolderAnalysis()"
	noCheck := false
	performCheck := true
	emptyDirName := "empty"
	if internal.Mkdir(t, fnName, emptyDirName) != nil {
		return
	}
	defer func() {
		if err := os.RemoveAll(emptyDirName); err != nil {
			t.Errorf("%s error destroying test directory %q: %v", fnName, emptyDirName, err)
		}
	}()
	dirtyDirName := "dirty"
	if internal.Mkdir(t, fnName, dirtyDirName) != nil {
		return
	}
	defer func() {
		if err := os.RemoveAll(dirtyDirName); err != nil {
			t.Errorf("%s error destroying test directory %q: %v", fnName, dirtyDirName, err)
		}
	}()
	internal.PopulateTopDir(t, fnName, dirtyDirName)
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
			name: "empty topDir",
			args: args{
				c: &check{checkEmptyFolders: &performCheck},
				s: files.CreateSearchForTesting(emptyDirName),
			},
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
			wantArtists: files.CreateAllArtists(dirtyDirName, true),
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
