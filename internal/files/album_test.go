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
			}.NewAlbum(false),
			want: "artist1",
		},
		"no recording artist": {
			a:    AlbumMaker{Title: "album1"}.NewAlbum(false),
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
		title:           "my album",
		recordingArtist: NewArtist("my artist", "Music/my artist"),
		directory:       "Music/my artist/my album",
		genre:           "rap",
		canonicalTitle:  "my special album",
		year:            "1993",
		cdIdentifier:    id3v2.UnknownFrame{Body: []byte{0, 1, 2}},
	}
	for k := 1; k <= 10; k++ {
		track := TrackMaker{
			Album:      complexAlbum,
			FileName:   fmt.Sprintf("%d track %d.mp3", k, k),
			SimpleName: fmt.Sprintf("track %d", k),
			Number:     k,
		}.NewTrack(false)
		complexAlbum.addTrack(track)
	}
	complexAlbum2 := &Album{
		title:           "my album",
		recordingArtist: NewArtist("my artist", "Music/my artist"),
		directory:       "Music/my artist/my album",
		genre:           "rap",
		canonicalTitle:  "my special album",
		year:            "1993",
		cdIdentifier:    id3v2.UnknownFrame{Body: []byte{0, 1, 2}},
	}
	for k := 1; k <= 10; k++ {
		track := TrackMaker{
			Album:      complexAlbum2,
			FileName:   fmt.Sprintf("%d track %d.mp3", k, k),
			SimpleName: fmt.Sprintf("track %d", k),
			Number:     k,
		}.NewTrack(false)
		complexAlbum2.addTrack(track)
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
				Title:     "album name",
				Artist:    NewArtist("artist", "Music/artist"),
				Directory: "Music/artist/album name",
			}.NewAlbum(false),
			args: args{
				ar:            NewArtist("artist", "Music/artist"),
				includeTracks: true,
			},
			want: AlbumMaker{
				Title:     "album name",
				Artist:    NewArtist("artist", "Music/artist"),
				Directory: "Music/artist/album name",
			}.NewAlbum(false),
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
			if got := tt.a.Copy(tt.args.ar, tt.args.includeTracks, false); !reflect.DeepEqual(got,
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
			a:    AlbumMaker{Title: "album", Directory: "artist/album"}.NewAlbum(false),
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
				Title:     "simple file",
				Artist:    testArtist,
				Directory: filepath.Join(testArtist.Directory(), "simple file"),
			}.NewAlbum(true),
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
			a: &Album{
				tracks: []*Track{{}},
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

func TestAlbum_Tracks(t *testing.T) {
	type fields struct {
		tracks          []*Track
		directory       string
		recordingArtist *Artist
		title           string
		genre           string
		canonicalTitle  string
		year            string
		cdIdentifier    id3v2.UnknownFrame
	}
	tracks := []*Track{
		{
			album:      nil,
			filePath:   `my artist\my album\01 track 1.mp3`,
			metadata:   nil,
			simpleName: "track 1",
			number:     1,
		},
		{
			album:      nil,
			filePath:   `my artist\my album\02 track 2.mp3`,
			metadata:   nil,
			simpleName: "track 2",
			number:     2,
		},
		{
			album:      nil,
			filePath:   `my artist\my album\03 track 3.mp3`,
			metadata:   nil,
			simpleName: "track 3",
			number:     3,
		},
	}
	tests := map[string]struct {
		fields
		want []*Track
	}{
		"trivial": {
			fields: fields{
				tracks:          tracks,
				directory:       `my artist\my album`,
				recordingArtist: nil,
				title:           "my album",
				genre:           "rock",
				canonicalTitle:  "my album",
				year:            "2024",
				cdIdentifier:    id3v2.UnknownFrame{},
			},
			want: tracks,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			a := &Album{
				tracks:          tt.fields.tracks,
				directory:       tt.fields.directory,
				recordingArtist: tt.fields.recordingArtist,
				title:           tt.fields.title,
				genre:           tt.fields.genre,
				canonicalTitle:  tt.fields.canonicalTitle,
				year:            tt.fields.year,
				cdIdentifier:    tt.fields.cdIdentifier,
			}
			if got := a.Tracks(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Tracks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlbum_Directory(t *testing.T) {
	type fields struct {
		tracks          []*Track
		directory       string
		recordingArtist *Artist
		title           string
		genre           string
		canonicalTitle  string
		year            string
		cdIdentifier    id3v2.UnknownFrame
	}
	tests := map[string]struct {
		fields
		want string
	}{
		"trivial": {
			fields: fields{
				tracks:          nil,
				directory:       `my artist\my album`,
				recordingArtist: nil,
				title:           "my album",
				genre:           "rock",
				canonicalTitle:  "my album",
				year:            "2024",
				cdIdentifier:    id3v2.UnknownFrame{},
			},
			want: `my artist\my album`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			a := &Album{
				tracks:          tt.fields.tracks,
				directory:       tt.fields.directory,
				recordingArtist: tt.fields.recordingArtist,
				title:           tt.fields.title,
				genre:           tt.fields.genre,
				canonicalTitle:  tt.fields.canonicalTitle,
				year:            tt.fields.year,
				cdIdentifier:    tt.fields.cdIdentifier,
			}
			if got := a.Directory(); got != tt.want {
				t.Errorf("Directory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlbum_Title(t *testing.T) {
	type fields struct {
		tracks          []*Track
		directory       string
		recordingArtist *Artist
		title           string
		genre           string
		canonicalTitle  string
		year            string
		cdIdentifier    id3v2.UnknownFrame
	}
	tests := map[string]struct {
		fields
		want string
	}{
		"trivial": {
			fields: fields{
				tracks:          nil,
				directory:       `my artist\my album`,
				recordingArtist: nil,
				title:           "my album",
				genre:           "rock",
				canonicalTitle:  "my album",
				year:            "2024",
				cdIdentifier:    id3v2.UnknownFrame{},
			},
			want: "my album",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			a := &Album{
				tracks:          tt.fields.tracks,
				directory:       tt.fields.directory,
				recordingArtist: tt.fields.recordingArtist,
				title:           tt.fields.title,
				genre:           tt.fields.genre,
				canonicalTitle:  tt.fields.canonicalTitle,
				year:            tt.fields.year,
				cdIdentifier:    tt.fields.cdIdentifier,
			}
			if got := a.Title(); got != tt.want {
				t.Errorf("Title() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortAlbums(t *testing.T) {
	tests := map[string]struct {
		albums []*Album
		want   []*Album
	}{
		"definitive": {
			albums: []*Album{
				AlbumMaker{
					Title:  "b",
					Artist: NewArtist("c", `music\c`),
				}.NewAlbum(false),
				AlbumMaker{
					Title:  "a",
					Artist: NewArtist("c", `music\c`),
				}.NewAlbum(false),
				AlbumMaker{
					Title:  "b",
					Artist: NewArtist("a", `music\a`),
				}.NewAlbum(false),
			},
			want: []*Album{
				AlbumMaker{
					Title:  "a",
					Artist: NewArtist("c", `music\c`),
				}.NewAlbum(false),
				AlbumMaker{
					Title:  "b",
					Artist: NewArtist("a", `music\a`),
				}.NewAlbum(false),
				AlbumMaker{
					Title:  "b",
					Artist: NewArtist("c", `music\c`),
				}.NewAlbum(false),
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			SortAlbums(tt.albums)
			if !reflect.DeepEqual(tt.albums, tt.want) {
				t.Errorf("SortAlbums() = %v, want %v", tt.albums, tt.want)
			}
		})
	}
}
