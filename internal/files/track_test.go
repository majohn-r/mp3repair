package files

import (
	"fmt"
	"io"
	"mp3/internal"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/bogem/id3v2/v2"
	"github.com/majohn-r/output"
)

func Test_parseTrackName(t *testing.T) {
	fnName := "parseTrackName()"
	type args struct {
		name  string
		album *Album
		ext   string
	}
	tests := []struct {
		name string
		args
		wantSimpleName  string
		wantTrackNumber int
		wantValid       bool
		output.WantedRecording
	}{
		{
			name:            "expected use case",
			wantSimpleName:  "track name",
			wantTrackNumber: 59,
			wantValid:       true,
			args: args{
				name:  "59 track name.mp3",
				album: &Album{name: "some album", recordingArtist: &Artist{name: "some artist"}},
				ext:   ".mp3",
			},
		},
		{
			name:            "expected use case with hyphen separator",
			wantSimpleName:  "other track name",
			wantTrackNumber: 60,
			wantValid:       true,
			args: args{
				name:  "60-other track name.mp3",
				album: &Album{name: "some album", recordingArtist: &Artist{name: "some artist"}},
				ext:   ".mp3",
			},
		},
		{
			name:            "wrong extension",
			wantSimpleName:  "track name.mp4",
			wantTrackNumber: 59,
			WantedRecording: output.WantedRecording{
				Error: "The track \"59 track name.mp4\" on album \"some album\" by artist \"some artist\" cannot be parsed.\n",
				Log:   "level='error' albumName='some album' artistName='some artist' trackName='59 track name.mp4' msg='the track name cannot be parsed'\n",
			},
			args: args{
				name:  "59 track name.mp4",
				album: &Album{name: "some album", recordingArtist: &Artist{name: "some artist"}},
				ext:   ".mp3",
			},
		},
		{
			name:           "missing track number",
			wantSimpleName: "name",
			WantedRecording: output.WantedRecording{
				Error: "The track \"track name.mp3\" on album \"some album\" by artist \"some artist\" cannot be parsed.\n",
				Log:   "level='error' albumName='some album' artistName='some artist' trackName='track name.mp3' msg='the track name cannot be parsed'\n",
			},
			args: args{
				name:  "track name.mp3",
				album: &Album{name: "some album", recordingArtist: &Artist{name: "some artist"}},
				ext:   ".mp3",
			},
		},
		{
			name: "missing track number, simple name",
			WantedRecording: output.WantedRecording{
				Error: "The track \"trackName.mp3\" on album \"some album\" by artist \"some artist\" cannot be parsed.\n",
				Log:   "level='error' albumName='some album' artistName='some artist' trackName='trackName.mp3' msg='the track name cannot be parsed'\n",
			},
			args: args{
				name:  "trackName.mp3",
				album: &Album{name: "some album", recordingArtist: &Artist{name: "some artist"}},
				ext:   ".mp3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			gotSimpleName, gotTrackNumber, gotValid := parseTrackName(o, tt.args.name, tt.args.album, tt.args.ext)
			if tt.wantValid {
				if gotSimpleName != tt.wantSimpleName {
					t.Errorf("%s gotSimpleName = %q, want %q", fnName, gotSimpleName, tt.wantSimpleName)
				}
				if gotTrackNumber != tt.wantTrackNumber {
					t.Errorf("%s gotTrackNumber = %d, want %d", fnName, gotTrackNumber, tt.wantTrackNumber)
				}
			}
			if gotValid != tt.wantValid {
				t.Errorf("%s gotValid = %v, want %v", fnName, gotValid, tt.wantValid)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_sortTracks(t *testing.T) {
	fnName := "sortTracks()"
	tests := []struct {
		name   string
		tracks []*Track
	}{
		{name: "degenerate case"},
		{
			name: "mixed tracks",
			tracks: []*Track{
				{
					number:          10,
					containingAlbum: NewAlbum("album2", NewArtist("artist3", ""), ""),
				},
				{
					number:          1,
					containingAlbum: NewAlbum("album2", NewArtist("artist3", ""), ""),
				},
				{
					number:          2,
					containingAlbum: NewAlbum("album1", NewArtist("artist3", ""), ""),
				},
				{
					number:          3,
					containingAlbum: NewAlbum("album3", NewArtist("artist2", ""), ""),
				},
				{
					number:          3,
					containingAlbum: NewAlbum("album3", NewArtist("artist4", ""), ""),
				},
				{
					number:          3,
					containingAlbum: NewAlbum("album5", NewArtist("artist2", ""), ""),
				},
			},
		},
	}
	for _, tt := range tests {
		sort.Sort(Tracks(tt.tracks))
		for i := range tt.tracks {
			if i == 0 {
				continue
			}
			track1 := tt.tracks[i-1]
			track2 := tt.tracks[i]
			album1 := track1.containingAlbum
			album2 := track2.containingAlbum
			artist1 := album1.RecordingArtistName()
			artist2 := album2.RecordingArtistName()
			if artist1 > artist2 {
				t.Errorf("%s track[%d] artist name %q comes after track[%d] artist name %q", fnName, i-1, artist1, i, artist2)
			} else {
				if artist1 == artist2 {
					if album1.Name() > album2.Name() {
						t.Errorf("%s track[%d] album name %q comes after track[%d] album name %q", fnName, i-1, album1.Name(), i, album2.Name())
					} else {
						if album1.Name() == album2.Name() {
							if track1.number > track2.number {
								t.Errorf("%s track[%d] track %d comes after track[%d] track %d", fnName, i-1, track1.number, i, track2.number)
							}
						}
					}
				}
			}
		}
	}
}

func TestTrack_BackupDirectory(t *testing.T) {
	fnName := "Track.BackupDirectory()"
	tests := []struct {
		name string
		tr   *Track
		want string
	}{
		{
			name: "simple case",
			tr:   &Track{containingAlbum: NewAlbum("", nil, "albumPath")},
			want: "albumPath\\pre-repair-backup",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.BackupDirectory(); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestTrack_AlbumPath(t *testing.T) {
	fnName := "Track.AlbumPath()"
	tests := []struct {
		name string
		tr   *Track
		want string
	}{
		{name: "no containing album", tr: &Track{}, want: ""},
		{name: "has containing album", tr: &Track{containingAlbum: NewAlbum("", nil, "album-path")}, want: "album-path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.AlbumPath(); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestTrack_Copy(t *testing.T) {
	fnName := "Track.Copy()"
	topDir := "copies"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, topDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	srcName := "source.mp3"
	srcPath := filepath.Join(topDir, srcName)
	if err := internal.CreateFileForTesting(topDir, srcName); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, srcPath, err)
	}
	type args struct {
		destination string
	}
	tests := []struct {
		name string
		tr   *Track
		args
		wantErr bool
	}{
		{
			name:    "error case",
			tr:      &Track{path: "no such file"},
			args:    args{destination: filepath.Join(topDir, "destination.mp3")},
			wantErr: true,
		},
		{
			name:    "good case",
			tr:      &Track{path: srcPath},
			args:    args{destination: filepath.Join(topDir, "destination.mp3")},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.tr.Copy(tt.args.destination); (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
			}
		})
	}
}

func TestTrack_String(t *testing.T) {
	fnName := "Track.String()"
	tests := []struct {
		name string
		tr   *Track
		want string
	}{{name: "expected", tr: &Track{path: "my path"}, want: "my path"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.String(); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestTrack_Name(t *testing.T) {
	fnName := "Track.Name()"
	tests := []struct {
		name string
		tr   *Track
		want string
	}{{name: "expected", tr: &Track{name: "track name"}, want: "track name"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.Name(); got != tt.want {
				t.Errorf("%s = %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestTrack_Number(t *testing.T) {
	fnName := "Track.Number()"
	tests := []struct {
		name string
		tr   *Track
		want int
	}{{name: "expected", tr: &Track{number: 42}, want: 42}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.Number(); got != tt.want {
				t.Errorf("%s = %d, want %d", fnName, got, tt.want)
			}
		})
	}
}

func TestTrack_AlbumName(t *testing.T) {
	fnName := "Track.AlbumName()"
	tests := []struct {
		name string
		tr   *Track
		want string
	}{
		{name: "orphan track", tr: &Track{}, want: ""},
		{name: "good track", tr: &Track{containingAlbum: &Album{name: "my album name"}}, want: "my album name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.AlbumName(); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestTrack_RecordingArtist(t *testing.T) {
	fnName := "Track.RecordingArtist()"
	tests := []struct {
		name string
		tr   *Track
		want string
	}{
		{name: "orphan track", tr: &Track{}, want: ""},
		{
			name: "good track",
			tr:   &Track{containingAlbum: &Album{recordingArtist: &Artist{name: "my artist"}}},
			want: "my artist",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.RecordingArtist(); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestParseTrackNameForTesting(t *testing.T) {
	fnName := "ParseTrackNameForTesting()"
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args
		wantSimpleName  string
		wantTrackNumber int
	}{
		{name: "hyphenated test", args: args{name: "03-track3.mp3"}, wantSimpleName: "track3", wantTrackNumber: 3},
		{name: "spaced test", args: args{name: "99 track99.mp3"}, wantSimpleName: "track99", wantTrackNumber: 99},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSimpleName, gotTrackNumber := ParseTrackNameForTesting(tt.args.name)
			if gotSimpleName != tt.wantSimpleName {
				t.Errorf("%s gotSimpleName = %q, want %q", fnName, gotSimpleName, tt.wantSimpleName)
			}
			if gotTrackNumber != tt.wantTrackNumber {
				t.Errorf("%s gotTrackNumber = %d, want %d", fnName, gotTrackNumber, tt.wantTrackNumber)
			}
		})
	}
}

func TestTrack_Path(t *testing.T) {
	fnName := "Track.Path()"
	tests := []struct {
		name string
		tr   *Track
		want string
	}{
		{
			name: "typical",
			tr:   &Track{path: "Music/my artist/my album/03 track.mp3"},
			want: "Music/my artist/my album/03 track.mp3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.Path(); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestTrack_Directory(t *testing.T) {
	fnName := "Track.Directory()"
	tests := []struct {
		name string
		tr   *Track
		want string
	}{
		{
			name: "typical",
			tr:   &Track{path: "Music/my artist/my album/03 track.mp3"},
			want: "Music\\my artist\\my album",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.Directory(); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestTrack_FileName(t *testing.T) {
	fnName := "Track.FileName()"
	tests := []struct {
		name string
		tr   *Track
		want string
	}{
		{
			name: "typical",
			tr:   &Track{path: "Music/my artist/my album/03 track.mp3"},
			want: "03 track.mp3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.FileName(); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func Test_pickKey(t *testing.T) {
	fnName := "pickKey()"
	type args struct {
		m map[string]int
	}
	tests := []struct {
		name string
		args
		wantS  string
		wantOk bool
	}{
		{
			name:   "unanimous choice",
			args:   args{m: map[string]int{"pop": 2}},
			wantS:  "pop",
			wantOk: true,
		},
		{
			name:   "majority for even size",
			args:   args{m: map[string]int{"pop": 3, "": 1}},
			wantS:  "pop",
			wantOk: true,
		},
		{
			name:   "majority for odd size",
			args:   args{m: map[string]int{"pop": 2, "": 1}},
			wantS:  "pop",
			wantOk: true,
		},
		{
			name: "no majority even size",
			args: args{m: map[string]int{"pop": 1, "alt-rock": 1}},
		},
		{
			name: "no majority odd size",
			args: args{m: map[string]int{"pop": 2, "alt-rock": 2, "folk": 1}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotS, gotOk := pickKey(tt.args.m)
			if gotS != tt.wantS {
				t.Errorf("%s gotS = %v, want %v", fnName, gotS, tt.wantS)
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk = %v, want %v", fnName, gotOk, tt.wantOk)
			}
		})
	}
}

func TestTrack_ID3V2Diagnostics(t *testing.T) {
	fnName := "Track.ID3V2Diagnostics()"
	payload := make([]byte, 0)
	for k := 0; k < 256; k++ {
		payload = append(payload, byte(k))
	}
	frames := map[string]string{
		"TYER": "2022",
		"TALB": "unknown album",
		"TRCK": "2",
		"TCON": "dance music",
		"TCOM": "a couple of idiots",
		"TIT2": "unknown track",
		"TPE1": "unknown artist",
		"TLEN": "1000",
		"T???": "who knows?",
		"Fake": "ummm",
	}
	content := CreateID3V2TaggedDataForTesting(payload, frames)
	if err := internal.CreateFileForTestingWithContent(".", "goodFile.mp3", content); err != nil {
		t.Errorf("%s failed to create ./goodFile.mp3: %v", fnName, err)
	}
	defer func() {
		if err := os.Remove("./goodFile.mp3"); err != nil {
			t.Errorf("%s failed to delete ./goodFile.mp3: %v", fnName, err)
		}
	}()
	tests := []struct {
		name        string
		tr          *Track
		wantEnc     string
		wantVersion byte
		wantF       []string
		wantErr     bool
	}{
		{
			name:    "error case",
			tr:      &Track{path: "./no such file"},
			wantErr: true,
		},
		{
			name:        "good case",
			tr:          &Track{path: "./goodfile.mp3"},
			wantEnc:     "ISO-8859-1",
			wantVersion: 3,
			wantF: []string{
				"Fake = \"<<[]byte{0x0, 0x75, 0x6d, 0x6d, 0x6d}>>\"",
				"T??? = \"who knows?\"",
				"TALB = \"unknown album\"",
				"TCOM = \"a couple of idiots\"",
				"TCON = \"dance music\"",
				"TIT2 = \"unknown track\"",
				"TLEN = \"1000\"",
				"TPE1 = \"unknown artist\"",
				"TRCK = \"2\"",
				"TYER = \"2022\"",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVersion, gotEnc, gotF, err := tt.tr.ID3V2Diagnostics()
			if (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
				return
			}
			if gotEnc != tt.wantEnc {
				t.Errorf("%s gotEnc = %q, want %q", fnName, gotEnc, tt.wantEnc)
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("%s gotVersion = %d, want %d", fnName, gotVersion, tt.wantVersion)
			}
			if !reflect.DeepEqual(gotF, tt.wantF) {
				t.Errorf("%s gotF = %v, want %v", fnName, gotF, tt.wantF)
			}
		})
	}
}

// this struct implements idev2.Framer as a means to provide an unexpected kind
// of Framer
type unspecifiedFrame struct {
	content string
}

func (u unspecifiedFrame) Size() int {
	return len(u.content)
}

func (u unspecifiedFrame) UniqueIdentifier() string {
	return ""
}

func (u unspecifiedFrame) WriteTo(w io.Writer) (n int64, err error) {
	var count int
	count, err = w.Write([]byte(u.content))
	n = int64(count)
	return
}

func TestTrack_ID3V1Diagnostics(t *testing.T) {
	fnName := "Track.ID3V1Diagnostics()"
	testDir := "id3v1Diagnostics"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, testDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	// three files: one good, one too small, one with an invalid tag
	smallFile := "01 small.mp3"
	if err := internal.CreateFileForTestingWithContent(testDir, smallFile, []byte{0, 1, 2}); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, smallFile, err)
	}
	invalidFile := "02 invalid.mp3"
	if err := internal.CreateFileForTestingWithContent(testDir, invalidFile, []byte{
		'd', 'A', 'G', // 'd' for defective!
		'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r', 'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i', 'e', 'w',
		'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A', 't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l', 'u', 'm',
		'2', '0', '1', '3',
		' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		0,
		29,
		12,
	}); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, invalidFile, err)
	}
	goodFile := "02 good.mp3"
	if err := internal.CreateFileForTestingWithContent(testDir, goodFile, []byte{
		'T', 'A', 'G',
		'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r', 'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i', 'e', 'w',
		'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A', 't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l', 'u', 'm',
		'2', '0', '1', '3',
		's', 'i', 'l', 'l', 'y', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		0,
		29,
		12,
	}); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, goodFile, err)
	}
	tests := []struct {
		name    string
		tr      *Track
		want    []string
		wantErr bool
	}{
		{
			name:    "small file",
			tr:      &Track{path: filepath.Join(testDir, smallFile)},
			wantErr: true,
		},
		{
			name:    "invalid file",
			tr:      &Track{path: filepath.Join(testDir, invalidFile)},
			wantErr: true,
		},
		{
			name: "good file",
			tr:   &Track{path: filepath.Join(testDir, goodFile)},
			want: []string{
				"Artist: \"The Beatles\"",
				"Album: \"On Air: Live At The BBC, Volum\"",
				"Title: \"Ringo - Pop Profile [Interview\"",
				"Track: 29",
				"Year: \"2013\"",
				"Genre: \"Other\"",
				"Comment: \"silly\"",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.tr.ID3V1Diagnostics()
			if (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestTrack_readTags(t *testing.T) {
	fnName := "track.readTags()"
	testDir := "readTags"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	artistName := "A great artist"
	albumName := "A really good album"
	trackName := "A brilliant track"
	genre := "Classic Rock"
	year := "2022"
	track := 5
	payload := CreateConsistentlyTaggedDataForTesting([]byte{0, 1, 2}, map[string]any{
		"artist": artistName,
		"album":  albumName,
		"title":  trackName,
		"genre":  genre,
		"year":   year,
		"track":  track,
	})
	fileName := "05 A brilliant track.mp3"
	if err := internal.CreateFileForTestingWithContent(testDir, fileName, payload); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, fileName, err)
	}
	tests := []struct {
		name string
		tr   *Track
		want *trackMetadata
	}{
		{
			name: "no read needed",
			tr: &Track{
				tM: &trackMetadata{},
			},
			want: &trackMetadata{},
		},
		{
			name: "read file",
			tr: &Track{
				path: filepath.Join(testDir, fileName),
			},
			want: &trackMetadata{
				album:             []string{"", albumName, albumName},
				artist:            []string{"", artistName, artistName},
				title:             []string{"", trackName, trackName},
				genre:             []string{"", genre, genre},
				year:              []string{"", year, year},
				track:             []int{0, track, track},
				musicCDIdentifier: id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:     id3v2Source,
				err:               []string{"", "", ""},
				correctedAlbum:    []string{"", "", ""},
				correctedArtist:   []string{"", "", ""},
				correctedTitle:    []string{"", "", ""},
				correctedGenre:    []string{"", "", ""},
				correctedYear:     []string{"", "", ""},
				correctedTrack:    []int{0, 0, 0},
				requiresEdit:      []bool{false, false, false},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tr.readTags()
			waitForSemaphoresDrained()
			if !reflect.DeepEqual(tt.tr.tM, tt.want) {
				t.Errorf("%s got %#v want %#v", fnName, tt.tr.tM, tt.want)
			}
		})
	}
}

func TestReadMetadata(t *testing.T) {
	fnName := "ReadMetadata()"
	// 5 artists, 20 albums each, 50 tracks apiece ... total: 5,000 tracks
	testDir := "ReadMetadata"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	var artists []*Artist
	for k := 0; k < 5; k++ {
		artistName := fmt.Sprintf("artist %d", k)
		artistPath := filepath.Join(testDir, artistName)
		if err := internal.Mkdir(artistPath); err != nil {
			t.Errorf("%s error creating %q: %v", fnName, artistPath, err)
		}
		artist := NewArtist(artistName, artistPath)
		artists = append(artists, artist)
		for m := 0; m < 20; m++ {
			albumName := fmt.Sprintf("album %d-%d", k, m)
			albumPath := filepath.Join(artistPath, albumName)
			if err := internal.Mkdir(albumPath); err != nil {
				t.Errorf("%s error creating %q: %v", fnName, albumPath, err)
			}
			album := NewAlbum(albumName, artist, albumName)
			artist.AddAlbum(album)
			for n := 0; n < 50; n++ {
				trackName := fmt.Sprintf("track %d-%d-%d", k, m, n)
				trackFileName := fmt.Sprintf("%02d %s.mp3", n+1, trackName)
				track := &Track{
					path:            filepath.Join(albumPath, trackFileName),
					name:            trackName,
					containingAlbum: album,
					number:          n + 1,
					tM:              nil,
				}
				metadata := map[string]any{
					"artist": artistName,
					"album":  albumName,
					"title":  trackName,
					"genre":  "Classic Rock",
					"year":   "2022",
					"track":  n + 1,
				}
				content := CreateConsistentlyTaggedDataForTesting([]byte{0, 1, 2, 3, 4, 5, 6, byte(k), byte(m), byte(n)}, metadata)
				if err := internal.CreateFileForTestingWithContent(albumPath, trackFileName, content); err != nil {
					t.Errorf("%s error creating %q: %v", fnName, trackName, err)
				}
				album.AddTrack(track)
			}
		}
	}
	type args struct {
		artists []*Artist
	}
	tests := []struct {
		name string
		args
		output.WantedRecording
	}{
		{
			name: "thorough test",
			args: args{artists: artists},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			ReadMetadata(o, tt.args.artists)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
			for _, artist := range tt.args.artists {
				for _, album := range artist.albums {
					for _, track := range album.tracks {
						if track.needsMetadata() {
							t.Errorf("%s track %q has no metadata", fnName, track.path)
						} else {
							if track.hasTagError() {
								t.Errorf("%s track %q is defective: %v", fnName, track.path, track.tM.errors())
							}
						}
					}
				}
			}
		})
	}
}

func TestTrack_ReportMetadataProblems(t *testing.T) {
	fnName := "Track.ReportMetadataProblems()"

	problematicArtist := NewArtist("problematic:artist", "")
	problematicAlbum := NewAlbum("problematic:album", problematicArtist, "")
	problematicAlbum.canonicalGenre = "hard rock"
	problematicAlbum.canonicalYear = "1999"
	problematicTrack := NewTrack(problematicAlbum, "03 bad track.mp3", "bad track", 3)
	metadata := newTrackMetadata()
	problematicTrack.tM = metadata
	src := id3v2Source
	metadata.canonicalType = src
	metadata.genre[src] = "unknown"
	metadata.year[src] = "2001"
	metadata.track[src] = 2
	metadata.album[src] = "unknown album"
	metadata.artist[src] = "unknown artist"
	metadata.title[src] = "unknown title"
	metadata.musicCDIdentifier = id3v2.UnknownFrame{Body: []byte{1, 3, 5}}
	problematicAlbum.AddTrack(problematicTrack)
	problematicArtist.AddAlbum(problematicAlbum)

	goodArtist := NewArtist("good artist", "")
	goodAlbum := NewAlbum("good album", goodArtist, "")
	goodAlbum.canonicalGenre = "Classic Rock"
	goodAlbum.canonicalYear = "1999"
	goodTrack := NewTrack(goodAlbum, "03 good track.mp3", "good track", 3)
	metadata2 := newTrackMetadata()
	goodTrack.tM = metadata2
	src2 := id3v1Source
	metadata2.canonicalType = src2
	metadata2.genre[src2] = "Classic Rock"
	metadata2.year[src2] = "1999"
	metadata2.track[src2] = 3
	metadata2.album[src2] = "good album"
	metadata2.artist[src2] = "good artist"
	metadata2.title[src2] = "good track"
	metadata2.err[id3v2Source] = "no id3v2 metadata, how odd"
	goodAlbum.AddTrack(goodTrack)
	goodArtist.AddAlbum(goodAlbum)

	tests := []struct {
		name string
		tr   *Track
		want []string
	}{
		// {
		// 	name: "typical use case",
		// 	tr: &Track{
		// 		number:          1,
		// 		name:            "track name",
		// 		containingAlbum: NewAlbum("album name", NewArtist("artist name", ""), ""),
		// 		tM: &trackMetadata{
		// 			track:  1,
		// 			title:  "track name",
		// 			album:  "album name",
		// 			artist: "artist name",
		// 		},
		// 	},
		// 	want: nil,
		// },
		// 		{
		// 			name: "another OK use case",
		// 			tr: &Track{
		// 				number:          1,
		// 				name:            "track name",
		// 				containingAlbum: NewAlbum("album name", NewArtist("artist name", ""), ""),
		// 				ID3V2TaggedTrackData: ID3V2TaggedTrackData{
		// 					track:  1,
		// 					title:  "track:name",
		// 					album:  "album name",
		// 					artist: "artist name",
		// 				},
		// 			},
		// 			want: nil,
		// 		},
		// 		{
		// 			name: "oops",
		// 			tr: &Track{
		// 				number:          2,
		// 				name:            "track:name",
		// 				containingAlbum: NewAlbum("album:name", NewArtist("artist:name", ""), ""),
		// 				ID3V2TaggedTrackData: ID3V2TaggedTrackData{
		// 					track:  1,
		// 					title:  "track name",
		// 					album:  "album name",
		// 					artist: "artist name",
		// 				},
		// 			},
		// 			want: []string{
		// 				"album \"album:name\" does not agree with album tag \"album name\"",
		// 				"artist \"artist:name\" does not agree with artist tag \"artist name\"",
		// 				"title \"track:name\" does not agree with title tag \"track name\"",
		// 				"track number 2 does not agree with track tag 1",
		// 			},
		// 		},
		{
			name: "unread tags",
			tr:   &Track{tM: nil},
			want: []string{noMetadata},
		},
		{
			name: "track with error",
			tr:   &Track{tM: &trackMetadata{err: []string{"", "oops", "oops"}}},
			want: []string{metadataReadError},
		},
		{
			name: "track with metadata differences",
			tr:   problematicTrack,
			want: []string{
				"metadata does not agree with album genre \"hard rock\"",
				"metadata does not agree with album name \"problematic:album\"",
				"metadata does not agree with album year \"1999\"",
				"metadata does not agree with artist name \"problematic:artist\"",
				"metadata does not agree with the MCDI frame \"\"",
				"metadata does not agree with track name \"bad track\"",
				"metadata does not agree with track number 3",
			},
		},
		{
			name: "track with no metadata differences",
			tr:   goodTrack,
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.ReportMetadataProblems(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestTrack_EditTags(t *testing.T) {
	fnName := "Track.EditTags()"
	testDir := "editTags"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	trackName := "edit this track.mp3"
	trackContents := CreateConsistentlyTaggedDataForTesting([]byte(trackName), map[string]any{
		"artist": "unknown artist",
		"album":  "unknown album",
		"title":  "unknown title",
		"genre":  "unknown",
		"year":   "1900",
		"track":  1,
	})
	if err := internal.CreateFileForTestingWithContent(testDir, trackName, trackContents); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, trackName, err)
	}
	track := &Track{
		path:   filepath.Join(testDir, trackName),
		name:   strings.TrimSuffix(trackName, ".mp3"),
		number: 2,
		containingAlbum: &Album{
			name:              "fine album",
			canonicalGenre:    "Classic Rock",
			canonicalYear:     "2022",
			canonicalTitle:    "fine album",
			musicCDIdentifier: id3v2.UnknownFrame{Body: []byte("fine album")},
			recordingArtist: &Artist{
				name:          "fine artist",
				canonicalName: "fine artist",
			},
		},
		tM: &trackMetadata{
			album:           []string{"", "unknown album", "unknown album"},
			artist:          []string{"", "unknown artist", "unknown artist"},
			title:           []string{"", "unknown title", "unknown title"},
			genre:           []string{"", "unknown", "unknown"},
			year:            []string{"", "1900", "1900"},
			track:           []int{0, 1, 1},
			canonicalType:   id3v2Source,
			err:             []string{"", "", ""},
			correctedAlbum:  make([]string, 3),
			correctedArtist: make([]string, 3),
			correctedTitle:  make([]string, 3),
			correctedGenre:  make([]string, 3),
			correctedYear:   make([]string, 3),
			correctedTrack:  make([]int, 3),
			requiresEdit:    make([]bool, 3),
		},
	}
	deletedTrack := &Track{
		path:   filepath.Join(testDir, "no such file"),
		name:   strings.TrimSuffix(trackName, ".mp3"),
		number: 2,
		containingAlbum: &Album{
			name:              "fine album",
			canonicalGenre:    "Classic Rock",
			canonicalYear:     "2022",
			canonicalTitle:    "fine album",
			musicCDIdentifier: id3v2.UnknownFrame{Body: []byte("fine album")},
			recordingArtist: &Artist{
				name:          "fine artist",
				canonicalName: "fine artist",
			},
		},
		tM: &trackMetadata{
			album:           []string{"", "unknown album", "unknown album"},
			artist:          []string{"", "unknown artist", "unknown artist"},
			title:           []string{"", "unknown title", "unknown title"},
			genre:           []string{"", "unknown", "unknown"},
			year:            []string{"", "1900", "1900"},
			track:           []int{0, 1, 1},
			canonicalType:   id3v2Source,
			err:             []string{"", "", ""},
			correctedAlbum:  make([]string, 3),
			correctedArtist: make([]string, 3),
			correctedTitle:  make([]string, 3),
			correctedGenre:  make([]string, 3),
			correctedYear:   make([]string, 3),
			correctedTrack:  make([]int, 3),
			requiresEdit:    make([]bool, 3),
		},
	}
	editedTm := &trackMetadata{
		album:             []string{"", "fine album", "fine album"},
		artist:            []string{"", "fine artist", "fine artist"},
		title:             []string{"", "edit this track", "edit this track"},
		genre:             []string{"", "Classic Rock", "Classic Rock"},
		year:              []string{"", "2022", "2022"},
		track:             []int{0, 2, 2},
		musicCDIdentifier: id3v2.UnknownFrame{Body: []byte("fine album")},
		canonicalType:     id3v2Source,
		err:               []string{"", "", ""},
		correctedAlbum:    make([]string, 3),
		correctedArtist:   make([]string, 3),
		correctedTitle:    make([]string, 3),
		correctedGenre:    make([]string, 3),
		correctedYear:     make([]string, 3),
		correctedTrack:    make([]int, 3),
		requiresEdit:      make([]bool, 3),
	}
	tests := []struct {
		name   string
		tr     *Track
		wantE  []string
		wantTm *trackMetadata
	}{
		{
			name: "error checking",
			tr:   deletedTrack,
			wantE: []string{
				"open editTags\\no such file: The system cannot find the file specified.",
				"open editTags\\no such file: The system cannot find the file specified.",
			},
		},
		{
			name:  "no edit required",
			tr:    &Track{tM: nil},
			wantE: []string{internal.ErrorEditUnnecessary},
		},
		{
			name:   "edit required",
			tr:     track,
			wantTm: editedTm,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotE := tt.tr.EditTags()
			var eStrings []string
			for _, e := range gotE {
				eStrings = append(eStrings, fmt.Sprintf("%v", e))
			}
			if !reflect.DeepEqual(eStrings, tt.wantE) {
				t.Errorf("%s = %v, want %v", fnName, eStrings, tt.wantE)
			} else {
				if len(gotE) == 0 && tt.tr.tM != nil {
					// verify file was correctly rewritten
					gotTm := readMetadata(tt.tr.path)
					if !reflect.DeepEqual(gotTm, tt.wantTm) {
						t.Errorf("%s read %#v, want %#v", fnName, gotTm, tt.wantTm)
					}
				}
			}
		})
	}
}

func Test_processArtistMetadata(t *testing.T) {
	fnName := "processArtistMetadata()"
	artist1 := NewArtist("artist_name", "")
	album1 := NewAlbum("album1", artist1, "")
	artist1.AddAlbum(album1)
	for k := 1; k <= 10; k++ {
		track := NewTrack(album1, fmt.Sprintf("%02d track%d.mp3", k, k), fmt.Sprintf("track%d", k), k)
		tM := newTrackMetadata()
		src := id3v2Source
		tM.canonicalType = src
		tM.artist[src] = "artist:name"
		track.tM = tM
		album1.AddTrack(track)
	}
	artist2 := NewArtist("artist_name", "")
	album2 := NewAlbum("album2", artist2, "")
	artist2.AddAlbum(album2)
	for k := 1; k <= 10; k++ {
		track := NewTrack(album2, fmt.Sprintf("%02d track%d.mp3", k, k), fmt.Sprintf("track%d", k), k)
		tM := newTrackMetadata()
		src := id3v2Source
		tM.canonicalType = src
		tM.artist[src] = "unknown artist"
		track.tM = tM
		album2.AddTrack(track)
	}
	artist3 := NewArtist("artist_name", "")
	album3 := NewAlbum("album3", artist3, "")
	artist3.AddAlbum(album3)
	for k := 1; k <= 10; k++ {
		track := NewTrack(album3, fmt.Sprintf("%02d track%d.mp3", k, k), fmt.Sprintf("track%d", k), k)
		tM := newTrackMetadata()
		src := id3v2Source
		tM.canonicalType = src
		track.tM = tM
		if k%2 == 0 {
			tM.artist[src] = "artist:name"
		} else {
			tM.artist[src] = "artist_name"
		}
		album3.AddTrack(track)
	}
	type args struct {
		artists []*Artist
	}
	tests := []struct {
		name string
		args
		output.WantedRecording
	}{
		{
			name: "unanimous choice",
			args: args{artists: []*Artist{artist1}},
		},
		{
			name: "unknown choice",
			args: args{artists: []*Artist{artist2}},
		},
		{
			name: "ambiguous choice",
			args: args{artists: []*Artist{artist3}},
			WantedRecording: output.WantedRecording{
				Error: "There are multiple artist name fields for \"artist_name\", and there is no unambiguously preferred choice; candidates are {\"artist:name\": 5 instances, \"artist_name\": 5 instances}.\n",
				Log:   "level='error' artistName='artist_name' field='artist name' settings='map[artist:name:5 artist_name:5]' msg='no value has a majority of instances'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			processArtistMetadata(o, tt.args.artists)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_processAlbumMetadata(t *testing.T) {
	fnName := "processAlbumMetadata()"
	// ordinary test data
	src := id3v2Source
	var artists1 []*Artist
	artist1 := NewArtist("good artist", "")
	artists1 = append(artists1, artist1)
	album1 := NewAlbum("good-album", artist1, "")
	artist1.AddAlbum(album1)
	track1 := NewTrack(album1, "01 track1.mp3", "track1", 1)
	track1.tM = newTrackMetadata()
	track1.tM.canonicalType = src
	track1.tM.genre[src] = "pop"
	track1.tM.year[src] = "2022"
	track1.tM.album[src] = "good:album"
	album1.AddTrack(track1)
	// more interesting test data
	var artists2 []*Artist
	artist2 := NewArtist("another good artist", "")
	artists2 = append(artists2, artist2)
	album2 := NewAlbum("another good_album", artist2, "")
	artist2.AddAlbum(album2)
	track2a := NewTrack(album2, "01 track1.mp3", "track1", 1)
	track2a.tM = newTrackMetadata()
	track2a.tM.canonicalType = src
	track2a.tM.genre[src] = "unknown"
	track2a.tM.year[src] = ""
	track2a.tM.album[src] = "unknown album"
	album2.AddTrack(track2a)
	track2b := NewTrack(album1, "02 track2.mp3", "track2", 2)
	track2b.tM = newTrackMetadata()
	track2b.tM.canonicalType = src
	track2b.tM.genre[src] = "pop"
	track2b.tM.year[src] = "2022"
	track2b.tM.album[src] = "another good:album"
	album2.AddTrack(track2b)
	track2c := NewTrack(album1, "03 track3.mp3", "track3", 3)
	track2c.tM = newTrackMetadata()
	track2c.tM.canonicalType = src
	track2c.tM.genre[src] = "pop"
	track2c.tM.year[src] = "2022"
	track2c.tM.album[src] = "another good:album"
	album2.AddTrack(track2c)
	// error case data
	var artists3 []*Artist
	artist3 := NewArtist("problematic artist", "")
	artists3 = append(artists3, artist3)
	album3 := NewAlbum("problematic_album", artist3, "")
	artist3.AddAlbum(album3)
	track3a := NewTrack(album2, "01 track1.mp3", "track1", 1)
	track3a.tM = newTrackMetadata()
	track3a.tM.canonicalType = src
	track3a.tM.genre[src] = "rock"
	track3a.tM.year[src] = "2023"
	track3a.tM.album[src] = "problematic:album"
	track3a.tM.musicCDIdentifier = id3v2.UnknownFrame{Body: []byte{1, 2, 3}}
	album3.AddTrack(track3a)
	track3b := NewTrack(album1, "02 track2.mp3", "track2", 2)
	track3b.tM = newTrackMetadata()
	track3b.tM.canonicalType = src
	track3b.tM.genre[src] = "pop"
	track3b.tM.year[src] = "2022"
	track3b.tM.album[src] = "problematic:Album"
	track3b.tM.musicCDIdentifier = id3v2.UnknownFrame{Body: []byte{1, 2, 3, 4}}
	album3.AddTrack(track3b)
	track3c := NewTrack(album1, "03 track3.mp3", "track3", 3)
	track3c.tM = newTrackMetadata()
	track3c.tM.canonicalType = src
	track3c.tM.genre[src] = "folk"
	track3c.tM.year[src] = "2021"
	track3c.tM.album[src] = "Problematic:album"
	track3c.tM.musicCDIdentifier = id3v2.UnknownFrame{Body: []byte{1, 2, 3, 4, 5}}
	album3.AddTrack(track3c)
	type args struct {
		artists []*Artist
	}
	tests := []struct {
		name string
		args
		output.WantedRecording
	}{
		{
			name: "ordinary test",
			args: args{artists: artists1},
		},
		{
			name: "typical use case",
			args: args{artists: artists2},
		},
		{
			name: "errors",
			args: args{artists: artists3},
			WantedRecording: output.WantedRecording{
				Error: "There are multiple genre fields for \"problematic_album by problematic artist\", and there is no unambiguously preferred choice; candidates are {\"folk\": 1 instance, \"pop\": 1 instance, \"rock\": 1 instance}.\n" +
					"There are multiple year fields for \"problematic_album by problematic artist\", and there is no unambiguously preferred choice; candidates are {\"2021\": 1 instance, \"2022\": 1 instance, \"2023\": 1 instance}.\n" +
					"There are multiple album title fields for \"problematic_album by problematic artist\", and there is no unambiguously preferred choice; candidates are {\"Problematic:album\": 1 instance, \"problematic:Album\": 1 instance, \"problematic:album\": 1 instance}.\n" +
					"There are multiple MCDI frame fields for \"problematic_album by problematic artist\", and there is no unambiguously preferred choice; candidates are {\"\\x01\\x02\\x03\": 1 instance, \"\\x01\\x02\\x03\\x04\": 1 instance, \"\\x01\\x02\\x03\\x04\\x05\": 1 instance}.\n",
				Log: "level='error' albumName='problematic_album' artistName='problematic artist' field='genre' settings='map[folk:1 pop:1 rock:1]' msg='no value has a majority of instances'\n" +
					"level='error' albumName='problematic_album' artistName='problematic artist' field='year' settings='map[2021:1 2022:1 2023:1]' msg='no value has a majority of instances'\n" +
					"level='error' albumName='problematic_album' artistName='problematic artist' field='album title' settings='map[Problematic:album:1 problematic:Album:1 problematic:album:1]' msg='no value has a majority of instances'\n" +
					"level='error' albumName='problematic_album' artistName='problematic artist' field='mcdi frame' settings='map[\x01\x02\x03:1 \x01\x02\x03\x04:1 \x01\x02\x03\x04\x05:1]' msg='no value has a majority of instances'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			processAlbumMetadata(o, tt.args.artists)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_reportTrackErrors(t *testing.T) {
	fnName := "reportTrackErrors()"
	type args struct {
		track  *Track
		album  *Album
		artist *Artist
	}
	tests := []struct {
		name string
		args
		output.WantedRecording
	}{
		{
			name: "error handling",
			args: args{
				track: &Track{
					name: "silly track",
					tM: &trackMetadata{
						err: []string{"", "id3v1 error!", "id3v2 error!"},
					},
				},
				album:  &Album{name: "silly album"},
				artist: &Artist{name: "silly artist"},
			},
			WantedRecording: output.WantedRecording{
				Error: "An error occurred when trying to read ID3V1 tag information for track \"silly track\" on album \"silly album\" by artist \"silly artist\": \"id3v1 error!\".\n" +
					"An error occurred when trying to read ID3V2 tag information for track \"silly track\" on album \"silly album\" by artist \"silly artist\": \"id3v2 error!\".\n",
				Log: "level='error' albumName='silly album' artistName='silly artist' error='id3v1 error!' trackName='silly track' msg='id3v1 tag error'\n" +
					"level='error' albumName='silly album' artistName='silly artist' error='id3v2 error!' trackName='silly track' msg='id3v2 tag error'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			reportTrackErrors(o, tt.args.track, tt.args.album, tt.args.artist)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func TestTrack_Details(t *testing.T) {
	fnName := "Track.Details()"
	payload := make([]byte, 0)
	for k := 0; k < 256; k++ {
		payload = append(payload, byte(k))
	}
	frames := map[string]string{
		"TYER": "2022",
		"TALB": "unknown album",
		"TRCK": "2",
		"TCON": "dance music",
		"TCOM": "a couple of idiots",
		"TIT2": "unknown track",
		"TPE1": "unknown artist",
		"TLEN": "1000",
		"T???": "who knows?",
		"TEXT": "An infinite number of monkeys with a typewriter",
		"TIT3": "Part II",
		"TKEY": "D Major",
		"TPE2": "The usual gang of idiots",
		"TPE3": "Someone with a stick",
	}
	content := CreateID3V2TaggedDataForTesting(payload, frames)
	if err := internal.CreateFileForTestingWithContent(".", "goodFile.mp3", content); err != nil {
		t.Errorf("%s failed to create ./goodFile.mp3: %v", fnName, err)
	}
	defer func() {
		if err := os.Remove("./goodFile.mp3"); err != nil {
			t.Errorf("%s failed to delete ./goodFile.mp3: %v", fnName, err)
		}
	}()
	tests := []struct {
		name    string
		tr      *Track
		want    map[string]string
		wantErr bool
	}{
		{
			name:    "error case",
			tr:      &Track{path: "./no such file"},
			wantErr: true,
		},
		{
			name: "good case",
			tr:   &Track{path: "./goodfile.mp3"},
			want: map[string]string{
				"Composer":       "a couple of idiots",
				"Lyricist":       "An infinite number of monkeys with a typewriter",
				"Subtitle":       "Part II",
				"Key":            "D Major",
				"Orchestra/Band": "The usual gang of idiots",
				"Conductor":      "Someone with a stick",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.tr.Details()
			if (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}
