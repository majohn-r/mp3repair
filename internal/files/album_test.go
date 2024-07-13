package files

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/bogem/id3v2/v2"
)

func TestAlbum_RecordingArtistName(t *testing.T) {
	tests := map[string]struct {
		a    *Album
		want string
	}{
		"with recording artist": {
			a: AlbumMaker{
				Title:  "album1",
				Artist: NewArtist("artist1", ""),
			}.NewAlbum(),
			want: "artist1",
		},
		"no recording artist": {
			a:    AlbumMaker{Title: "album1"}.NewAlbum(),
			want: "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.a.RecordingArtistName(); got != tt.want {
				t.Errorf("Album.RecordingArtistName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAlbum_Copy(t *testing.T) {
	complexAlbum := &Album{
		Title:             "my album",
		RecordingArtist:   NewArtist("my artist", "Music/my artist"),
		FilePath:          "Music/my artist/my album",
		CanonicalGenre:    "rap",
		CanonicalTitle:    "my special album",
		CanonicalYear:     "1993",
		MusicCDIdentifier: id3v2.UnknownFrame{Body: []byte{0, 1, 2}},
	}
	for k := 1; k <= 10; k++ {
		track := TrackMaker{
			Album:      complexAlbum,
			FileName:   fmt.Sprintf("%d track %d.mp3", k, k),
			SimpleName: fmt.Sprintf("track %d", k),
			Number:     k,
		}.NewTrack()
		complexAlbum.AddTrack(track)
	}
	complexAlbum2 := &Album{
		Title:             "my album",
		RecordingArtist:   NewArtist("my artist", "Music/my artist"),
		FilePath:          "Music/my artist/my album",
		CanonicalGenre:    "rap",
		CanonicalTitle:    "my special album",
		CanonicalYear:     "1993",
		MusicCDIdentifier: id3v2.UnknownFrame{Body: []byte{0, 1, 2}},
	}
	for k := 1; k <= 10; k++ {
		track := TrackMaker{
			Album:      complexAlbum2,
			FileName:   fmt.Sprintf("%d track %d.mp3", k, k),
			SimpleName: fmt.Sprintf("track %d", k),
			Number:     k,
		}.NewTrack()
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
			a: AlbumMaker{
				Title:  "album name",
				Artist: NewArtist("artist", "Music/artist"),
				Path:   "Music/artist/album name",
			}.NewAlbum(),
			args: args{
				ar:            NewArtist("artist", "Music/artist"),
				includeTracks: true,
			},
			want: AlbumMaker{
				Title:  "album name",
				Artist: NewArtist("artist", "Music/artist"),
				Path:   "Music/artist/album name",
			}.NewAlbum(),
		},
		"complex test": {
			a: complexAlbum,
			args: args{
				ar:            complexAlbum.RecordingArtist.Copy(),
				includeTracks: true,
			},
			want: complexAlbum2,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.a.Copy(tt.args.ar, tt.args.includeTracks); !reflect.DeepEqual(got,
				tt.want) {
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
			a:    AlbumMaker{Title: "album", Path: "artist/album"}.NewAlbum(),
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

/*
	Name() string       // base name of the file
	Size() int64        // length in bytes for regular files; system-dependent for others
	Mode() FileMode     // file mode bits
	ModTime() time.Time // modification time
	IsDir() bool        // abbreviation for Mode().IsDir()
	Sys() any           // underlying data source (can return nil)
*/

type testFile struct {
	name string
	mode fs.FileMode
}

func (tf *testFile) Name() string {
	return tf.name
}

func (tf *testFile) Size() int64 {
	return 0
}

func (tf *testFile) Mode() fs.FileMode {
	return tf.mode
}

func (tf *testFile) ModTime() time.Time {
	return time.Now()
}

func (tf *testFile) IsDir() bool {
	return tf.mode.IsDir()
}

func (tf *testFile) Sys() any {
	return nil
}

func TestNewAlbumFromFile(t *testing.T) {
	testArtist := NewArtist("artist name", filepath.Join("Music", "artist name"))
	type args struct {
		file fs.FileInfo
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
			want: AlbumMaker{
				Title:  "simple file",
				Artist: testArtist,
				Path:   filepath.Join(testArtist.FilePath, "simple file"),
			}.NewAlbum(),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := NewAlbumFromFile(tt.args.file,
				tt.args.ar); !reflect.DeepEqual(got, tt.want) {
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
			a: &Album{
				Tracks: []*Track{{}},
			},
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
