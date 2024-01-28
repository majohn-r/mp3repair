package files

import (
	"io/fs"
	"path/filepath"
	"reflect"
	"testing"
)

func TestArtist_Copy(t *testing.T) {
	complexArtist := NewArtist("artist's name", "Music/artist's name")
	complexArtist.canonicalName = "Actually, Fred"
	complexArtist2 := NewArtist("artist's name", "Music/artist's name")
	complexArtist2.canonicalName = "Actually, Fred"
	tests := map[string]struct {
		a    *Artist
		want *Artist
	}{
		"simple test": {
			a:    NewArtist("artist name", "Music/artist name"),
			want: NewArtist("artist name", "Music/artist name"),
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
		want *Artist
	}{
		"simple": {
			args: args{
				f:   &testFile{name: "my artist"},
				dir: "Music",
			},
			want: NewArtist("my artist", filepath.Join("Music", "my artist")),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := NewArtistFromFile(tt.args.f, tt.args.dir); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewArtistFromFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArtist_HasAlbums(t *testing.T) {
	tests := map[string]struct {
		a    *Artist
		want bool
	}{
		"empty":       {a: &Artist{}, want: false},
		"with albums": {a: &Artist{albums: []*Album{{}}}, want: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.a.HasAlbums(); got != tt.want {
				t.Errorf("Artist.HasAlbums() = %v, want %v", got, tt.want)
			}
		})
	}
}
