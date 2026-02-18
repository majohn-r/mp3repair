/*
Copyright © 2026 Marc Johnson (marc.johnson27591@gmail.com)
*/
package files

import (
	"io/fs"
	"path/filepath"
	"reflect"
	"testing"
)

func TestArtist_Copy(t *testing.T) {
	complexArtist := &Artist{
		name:       "artist's name",
		directory:  "Music/artist's name",
		sharedName: "Actually, Fred",
	}
	complexArtist2 := &Artist{
		name:       "artist's name",
		directory:  "Music/artist's name",
		sharedName: "Actually, Fred",
	}
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
		f   fs.FileInfo
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
			if got := NewArtistFromFile(tt.args.f,
				tt.args.dir); !reflect.DeepEqual(got, tt.want) {
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
		"empty": {a: &Artist{}, want: false},
		"with albums": {
			a:    &Artist{albums: []*Album{{}}},
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
