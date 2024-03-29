package files_test

import (
	"io/fs"
	"mp3repair/internal/files"
	"path/filepath"
	"reflect"
	"testing"
)

func TestArtist_Copy(t *testing.T) {
	complexArtist := files.NewArtist(
		"artist's name", "Music/artist's name").WithCanonicalName("Actually, Fred")
	complexArtist2 := files.NewArtist(
		"artist's name", "Music/artist's name").WithCanonicalName("Actually, Fred")
	tests := map[string]struct {
		a    *files.Artist
		want *files.Artist
	}{
		"simple test": {
			a:    files.NewArtist("artist name", "Music/artist name"),
			want: files.NewArtist("artist name", "Music/artist name"),
		},
		"complex test": {
			a:    complexArtist,
			want: complexArtist2,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.a.Copy(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Artist.Copy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewArtistFromFile(t *testing.T) {
	type args struct {
		f   fs.DirEntry
		dir string
	}
	tests := map[string]struct {
		args
		want *files.Artist
	}{
		"simple": {
			args: args{
				f:   &testFile{name: "my artist"},
				dir: "Music",
			},
			want: files.NewArtist("my artist", filepath.Join("Music", "my artist")),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.NewArtistFromFile(tt.args.f,
				tt.args.dir); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewArtistFromFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArtist_HasAlbums(t *testing.T) {
	tests := map[string]struct {
		a    *files.Artist
		want bool
	}{
		"empty": {a: files.NewEmptyArtist(), want: false},
		"with albums": {
			a:    files.NewEmptyArtist().WithAlbums([]*files.Album{{}}),
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.a.HasAlbums(); got != tt.want {
				t.Errorf("Artist.HasAlbums() = %v, want %v", got, tt.want)
			}
		})
	}
}
