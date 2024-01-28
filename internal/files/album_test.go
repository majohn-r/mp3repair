package files

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"reflect"
	"testing"
)

func TestAlbum_RecordingArtistName(t *testing.T) {
	const fnName = "Album.RecordingArtistName()"
	tests := map[string]struct {
		a    *Album
		want string
	}{
		"with recording artist": {a: NewAlbum("album1", NewArtist("artist1", ""), ""), want: "artist1"},
		"no recording artist":   {a: NewAlbum("album1", nil, ""), want: ""},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.a.RecordingArtistName(); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestAlbum_Copy(t *testing.T) {
	complexAlbum := NewAlbum("my album", NewArtist("my artist", "Music/my artist"), "Music/my artist/my album")
	complexAlbum.canonicalGenre = "rap"
	complexAlbum.canonicalTitle = "my special album"
	complexAlbum.canonicalYear = "1993"
	complexAlbum.musicCDIdentifier.Body = []byte{0, 1, 2}
	for k := 1; k <= 10; k++ {
		track := NewTrack(complexAlbum, fmt.Sprintf("%d track %d.mp3", k, k), fmt.Sprintf("track %d.mp3", k), k)
		complexAlbum.AddTrack(track)
	}
	complexAlbum2 := NewAlbum("my album", NewArtist("my artist", "Music/my artist"), "Music/my artist/my album")
	complexAlbum2.canonicalGenre = "rap"
	complexAlbum2.canonicalTitle = "my special album"
	complexAlbum2.canonicalYear = "1993"
	complexAlbum2.musicCDIdentifier.Body = []byte{0, 1, 2}
	for k := 1; k <= 10; k++ {
		track := NewTrack(complexAlbum2, fmt.Sprintf("%d track %d.mp3", k, k), fmt.Sprintf("track %d.mp3", k), k)
		complexAlbum2.AddTrack(track)
	}
	type args struct {
		ar            *Artist
		includeTracks bool
	}
	tests := map[string]struct {
		a *Album
		args
		want *Album
	}{
		"simple test": {
			a: NewAlbum("album name", NewArtist("artist", "Music/artist"), "Music/artist/album name"),
			args: args{
				ar:            NewArtist("artist", "Music/artist"),
				includeTracks: true,
			},
			want: NewAlbum("album name", NewArtist("artist", "Music/artist"), "Music/artist/album name"),
		},
		"complex test": {
			a: complexAlbum,
			args: args{
				ar:            complexAlbum.recordingArtist.Copy(),
				includeTracks: true,
			},
			want: complexAlbum2,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.a.Copy(tt.args.ar, tt.args.includeTracks); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Album.Copy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlbum_BackupDirectory(t *testing.T) {
	tests := map[string]struct {
		a    *Album
		want string
	}{
		"simple": {
			a:    NewAlbum("album", nil, "artist/album"),
			want: "artist\\album\\pre-repair-backup",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.a.BackupDirectory(); got != tt.want {
				t.Errorf("Album.BackupDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}

type testFile struct {
	name  string
	files []*testFile
}

func (tf *testFile) Name() string {
	return tf.name
}

func (tf *testFile) IsDir() bool {
	return len(tf.files) > 0
}

func (tf *testFile) Type() fs.FileMode {
	if tf.IsDir() {
		return fs.ModeDir
	}
	return 0
}

func (tf *testFile) Info() (fs.FileInfo, error) {
	return nil, nil
}

func TestNewAlbumFromFile(t *testing.T) {
	testArtist := NewArtist("artist name", filepath.Join("Music", "artist name"))
	type args struct {
		file fs.DirEntry
		ar   *Artist
	}
	tests := map[string]struct {
		args
		want *Album
	}{
		"simple": {
			args: args{
				file: &testFile{name: "simple file"},
				ar:   testArtist,
			},
			want: NewAlbum("simple file", testArtist, filepath.Join(testArtist.path, "simple file")),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := NewAlbumFromFile(tt.args.file, tt.args.ar); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAlbumFromFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlbum_HasTracks(t *testing.T) {
	tests := map[string]struct {
		a    *Album
		want bool
	}{
		"empty": {
			a:    &Album{},
			want: false,
		},
		"with tracks": {
			a:    &Album{tracks: []*Track{{}}},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.a.HasTracks(); got != tt.want {
				t.Errorf("Album.HasTracks() = %v, want %v", got, tt.want)
			}
		})
	}
}
