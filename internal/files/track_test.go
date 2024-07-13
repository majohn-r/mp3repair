package files

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/bogem/id3v2/v2"
	"github.com/cheggaaa/pb/v3"
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/afero"
)

func TestTrackNameParserParse(t *testing.T) {
	tests := map[string]struct {
		parser         TrackNameParser
		wantParsedName *ParsedTrackName
		wantValid      bool
		output.WantedRecording
	}{
		"expected use case": {
			parser: TrackNameParser{
				FileName: "59 track name.mp3",
				Album: &Album{
					Title:           "some album",
					RecordingArtist: &Artist{Name: "some artist"},
				},
				Extension: ".mp3",
			},
			wantParsedName: &ParsedTrackName{SimpleName: "track name", Number: 59},
			wantValid:      true,
		},
		"expected use case with hyphen separator": {
			parser: TrackNameParser{
				FileName: "60-other track name.mp3",
				Album: &Album{
					Title:           "some album",
					RecordingArtist: &Artist{Name: "some artist"},
				},
				Extension: ".mp3",
			},
			wantParsedName: &ParsedTrackName{
				SimpleName: "other track name",
				Number:     60,
			},
			wantValid: true,
		},
		"wrong extension": {
			parser: TrackNameParser{
				FileName: "59 track name.mp4",
				Album: &Album{
					Title:           "some album",
					RecordingArtist: &Artist{Name: "some artist"},
				},
				Extension: ".mp3",
			},
			wantParsedName: &ParsedTrackName{
				SimpleName: "track name.mp4",
				Number:     59,
			},
			WantedRecording: output.WantedRecording{
				Error: "The track \"59 track name.mp4\" on album \"some album\" by" +
					" artist \"some artist\" cannot be parsed.\n",
				Log: "level='error'" +
					" albumName='some album'" +
					" artistName='some artist'" +
					" trackName='59 track name.mp4'" +
					" msg='the track name cannot be parsed'\n",
			},
		},
		"missing track number": {
			parser: TrackNameParser{
				FileName: "track name.mp3",
				Album: &Album{
					Title:           "some album",
					RecordingArtist: &Artist{Name: "some artist"},
				},
				Extension: ".mp3",
			},
			wantParsedName: &ParsedTrackName{SimpleName: "name"},
			WantedRecording: output.WantedRecording{
				Error: "The track \"track name.mp3\" on album \"some album\" by artist" +
					" \"some artist\" cannot be parsed.\n",
				Log: "level='error'" +
					" albumName='some album'" +
					" artistName='some artist'" +
					" trackName='track name.mp3'" +
					" msg='the track name cannot be parsed'\n",
			},
		},
		"missing track number, simple name": {
			parser: TrackNameParser{
				FileName: "trackName.mp3",
				Album: &Album{
					Title:           "some album",
					RecordingArtist: &Artist{Name: "some artist"},
				},
				Extension: ".mp3",
			},
			WantedRecording: output.WantedRecording{
				Error: "The track \"trackName.mp3\" on album \"some album\" by artist" +
					" \"some artist\" cannot be parsed.\n",
				Log: "level='error'" +
					" albumName='some album'" +
					" artistName='some artist'" +
					" trackName='trackName.mp3'" +
					" msg='the track name cannot be parsed'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			gotParsedName, gotValid := tt.parser.Parse(o)
			if tt.wantValid {
				if !reflect.DeepEqual(gotParsedName, tt.wantParsedName) {
					t.Errorf("TrackNameParser.Parse() gotParsedName = %v, want %v", gotParsedName, tt.wantParsedName)
				}
			}
			if gotValid != tt.wantValid {
				t.Errorf("TrackNameParser.Parse() gotValid = %v, want %v", gotValid, tt.wantValid)
			}
			o.Report(t, "TrackNameParser.Parse()", tt.WantedRecording)
		})
	}
}

func Test_tracks_Sort(t *testing.T) {
	tests := map[string]struct {
		tracks []*Track
	}{
		"degenerate case": {},
		"mixed tracks": {
			tracks: []*Track{
				{
					Number: 10,
					Album: AlbumMaker{
						Title:  "album2",
						Artist: NewArtist("artist3", ""),
					}.NewAlbum(),
				},
				{
					Number: 1,
					Album: AlbumMaker{
						Title:  "album2",
						Artist: NewArtist("artist3", ""),
					}.NewAlbum(),
				},
				{
					Number: 3,
					Album: AlbumMaker{
						Title:  "album3",
						Artist: NewArtist("artist2", ""),
					}.NewAlbum(),
				},
				{
					Number: 3,
					Album: AlbumMaker{
						Title:  "album3",
						Artist: NewArtist("artist4", ""),
					}.NewAlbum(),
				},
				{
					Number: 3,
					Album: AlbumMaker{
						Title:  "album5",
						Artist: NewArtist("artist2", ""),
					}.NewAlbum(),
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			sort.Sort(tracks(tt.tracks))
			for i := range tt.tracks {
				if i == 0 {
					continue
				}
				track1 := tt.tracks[i-1]
				track2 := tt.tracks[i]
				album1 := track1.Album
				album2 := track2.Album
				artist1 := album1.RecordingArtistName()
				artist2 := album2.RecordingArtistName()
				if artist1 > artist2 {
					t.Errorf("tracks.Sort() track[%d] artist name %q comes after track[%d] artist name %q",
						i-1, artist1, i, artist2)
				} else if artist1 == artist2 {
					if album1.Title > album2.Title {
						t.Errorf("tracks.Sort() track[%d] album name %q comes after track[%d] album name %q",
							i-1, album1.Title, i, album2.Title)
					} else if album1.Title == album2.Title {
						if track1.Number > track2.Number {
							t.Errorf("tracks.Sort() track[%d] track %d comes after track[%d] track %d",
								i-1, track1.Number, i, track2.Number)
						}
					}
				}
			}
		})
	}
}

func TestTrack_AlbumPath(t *testing.T) {
	tests := map[string]struct {
		t    *Track
		want string
	}{
		"no containing album": {t: &Track{}, want: ""},
		"has containing album": {
			t:    &Track{Album: AlbumMaker{Path: "album-path"}.NewAlbum()},
			want: "album-path"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.t.AlbumPath(); got != tt.want {
				t.Errorf("Track.AlbumPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTrack_CopyFile(t *testing.T) {
	originalFileSystem := cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmdtoolkit.AssignFileSystem(originalFileSystem)
	}()
	topDir := "copies"
	_ = cmdtoolkit.Mkdir(topDir)
	srcName := "source.mp3"
	srcPath := filepath.Join(topDir, srcName)
	_ = createFile(topDir, srcName)
	tests := map[string]struct {
		t           *Track
		destination string
		wantErr     bool
	}{
		"error case": {
			t:           &Track{FilePath: "no such file"},
			destination: filepath.Join(topDir, "destination.mp3"),
			wantErr:     true,
		},
		"good case": {
			t:           &Track{FilePath: srcPath},
			destination: filepath.Join(topDir, "destination.mp3"),
			wantErr:     false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if gotErr := tt.t.CopyFile(tt.destination); (gotErr != nil) != tt.wantErr {
				t.Errorf("Track.CopyFile() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestTrackString(t *testing.T) {
	tests := map[string]struct {
		t    *Track
		want string
	}{
		"expected": {
			t:    &Track{FilePath: "my path"},
			want: "my path",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.t.String(); got != tt.want {
				t.Errorf("Track.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTrack_AlbumName(t *testing.T) {
	tests := map[string]struct {
		t    *Track
		want string
	}{
		"orphan track": {t: &Track{}, want: ""},
		"good track": {
			t:    &Track{Album: &Album{Title: "my album name"}},
			want: "my album name"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.t.AlbumName(); got != tt.want {
				t.Errorf("Track.AlbumName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTrack_RecordingArtist(t *testing.T) {
	tests := map[string]struct {
		t    *Track
		want string
	}{
		"orphan track": {t: &Track{}, want: ""},
		"good track": {
			t: &Track{
				Album: &Album{RecordingArtist: &Artist{Name: "my artist"}},
			},
			want: "my artist",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.t.RecordingArtist(); got != tt.want {
				t.Errorf("Track.RecordingArtist() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTrack_Directory(t *testing.T) {
	tests := map[string]struct {
		t    *Track
		want string
	}{
		"typical": {
			t:    &Track{FilePath: "Music/my artist/my album/03 track.mp3"},
			want: "Music\\my artist\\my album",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.t.Directory(); got != tt.want {
				t.Errorf("Track.Directory() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTrack_FileName(t *testing.T) {
	tests := map[string]struct {
		t    *Track
		want string
	}{
		"typical": {
			t:    &Track{FilePath: "Music/my artist/my album/03 track.mp3"},
			want: "03 track.mp3",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.t.FileName(); got != tt.want {
				t.Errorf("Track.FileName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCanonicalChoice(t *testing.T) {
	tests := map[string]struct {
		m      map[string]int
		wantS  string
		wantOk bool
	}{
		"unanimous choice": {
			m:      map[string]int{"pop": 2},
			wantS:  "pop",
			wantOk: true,
		},
		"majority for even size": {
			m: map[string]int{
				"pop": 3,
				"":    1,
			},
			wantS:  "pop",
			wantOk: true,
		},
		"majority for odd size": {
			m: map[string]int{
				"pop": 2,
				"":    1,
			},
			wantS:  "pop",
			wantOk: true,
		},
		"no majority even size": {
			m: map[string]int{
				"pop":      1,
				"alt-rock": 1,
			},
		},
		"no majority odd size": {
			m: map[string]int{
				"pop":      2,
				"alt-rock": 2,
				"folk":     1,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotS, gotOk := CanonicalChoice(tt.m)
			if gotS != tt.wantS {
				t.Errorf("CanonicalChoice gotS = %v, want %v", gotS, tt.wantS)
			}
			if gotOk != tt.wantOk {
				t.Errorf("CanonicalChoice gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestTrack_ID3V2Diagnostics(t *testing.T) {
	originalFileSystem := cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmdtoolkit.AssignFileSystem(originalFileSystem)
	}()
	audio := make([]byte, 0)
	for k := 0; k < 256; k++ {
		audio = append(audio, byte(k))
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
		"Fake": "huh",
	}
	content := createID3v2TaggedData(audio, frames)
	goodFileName := "goodFile.mp3"
	_ = createFileWithContent(".", goodFileName, content)
	tests := map[string]struct {
		t                *Track
		wantEncoding     string
		wantVersion      byte
		wantFrameStrings []string
		wantErr          bool
	}{
		"error case": {
			t:       &Track{FilePath: "./no such file"},
			wantErr: true,
		},
		"good case": {
			t:            &Track{FilePath: filepath.Join(".", goodFileName)},
			wantEncoding: "ISO-8859-1",
			wantVersion:  3,
			wantFrameStrings: []string{
				"Fake = \"<<[]byte{0x0, 0x68, 0x75, 0x68}>>\"",
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
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotInfo, gotErr := tt.t.ID3V2Diagnostics()
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("Track.ID3V2Diagnostics() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if gotErr == nil {
				if gotInfo.Encoding != tt.wantEncoding {
					t.Errorf("Track.ID3V2Diagnostics() gotInfo.Encoding = %q, want %q",
						gotInfo.Encoding, tt.wantEncoding)
				}
				if gotInfo.Version != tt.wantVersion {
					t.Errorf("Track.ID3V2Diagnostics() gotInfo.Version = %d, want %d",
						gotInfo.Version, tt.wantVersion)
				}
				if !reflect.DeepEqual(gotInfo.FrameStrings, tt.wantFrameStrings) {
					t.Errorf("Track.ID3V2Diagnostics() gotInfo.FrameStrings = %v, want %v",
						gotInfo.FrameStrings, tt.wantFrameStrings)
				}
			}
		})
	}
}

func TestTrack_ID3V1Diagnostics(t *testing.T) {
	originalFileSystem := cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmdtoolkit.AssignFileSystem(originalFileSystem)
	}()
	testDir := "id3v1Diagnostics"
	_ = cmdtoolkit.Mkdir(testDir)
	// three files: one good, one too small, one with an invalid tag
	smallFile := "01 small.mp3"
	_ = createFileWithContent(testDir, smallFile, []byte{0, 1, 2})
	invalidFile := "02 invalid.mp3"
	_ = createFileWithContent(testDir, invalidFile, []byte{
		'd', 'A', 'G', // 'd' for defective!
		'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r', 'o', 'f',
		'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i', 'e', 'w',
		'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A', 't', ' ',
		'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l', 'u', 'm',
		'2', '0', '1', '3',
		' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		0,
		29,
		12,
	})
	goodFile := "02 good.mp3"
	_ = createFileWithContent(testDir, goodFile, []byte{
		'T', 'A', 'G',
		'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r', 'o', 'f',
		'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i', 'e', 'w',
		'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A', 't', ' ',
		'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l', 'u', 'm',
		'2', '0', '1', '3',
		's', 'i', 'l', 'l', 'y', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		0,
		29,
		12,
	})
	tests := map[string]struct {
		t       *Track
		want    []string
		wantErr bool
	}{
		"small file": {
			t:       &Track{FilePath: filepath.Join(testDir, smallFile)},
			wantErr: true,
		},
		"invalid file": {
			t:       &Track{FilePath: filepath.Join(testDir, invalidFile)},
			wantErr: true,
		},
		"good file": {
			t: &Track{FilePath: filepath.Join(testDir, goodFile)},
			want: []string{
				"Artist: \"The Beatles\"",
				"Album: \"On Air: Live At The BBC, Volum\"",
				"Title: \"Ringo - Pop Profile [Interview\"",
				"Track: 29",
				"Year: \"2013\"",
				"Genre: \"other\"",
				"Comment: \"silly\"",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotErr := tt.t.ID3V1Diagnostics()
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("Track.ID3V1Diagnostics() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Track.ID3V1Diagnostics() = %v, want %v", got, tt.want)
			}
		})
	}
}

var nameToID3V2TagName = map[string]string{
	"artist": "TPE1",
	"album":  "TALB",
	"title":  "TIT2",
	"genre":  "TCON",
	"year":   "TYER",
	"track":  "TRCK",
}

var recognizedTagNames = []string{"artist", "album", "title", "genre", "year", "track"}

func createID3V1TaggedData(m map[string]any) []byte {
	v1 := newID3v1Metadata()
	v1.writeString("TAG", tagField)
	for _, tagName := range recognizedTagNames {
		if value, tagNameFound := m[tagName]; tagNameFound {
			switch tagName {
			case "artist":
				v1.setArtist(value.(string))
			case "album":
				v1.setAlbum(value.(string))
			case "title":
				v1.setTitle(value.(string))
			case "genre":
				v1.setGenre(value.(string))
			case "year":
				v1.setYear(value.(string))
			case "track":
				v1.setTrack(value.(int))
			}
		}
	}
	return v1.rawData()
}

func createConsistentlyTaggedData(audio []byte, m map[string]any) []byte {
	var frames = map[string]string{}
	for _, tagName := range recognizedTagNames {
		if value, tagNameFound := m[tagName]; tagNameFound {
			switch tagName {
			case "track":
				frames[nameToID3V2TagName[tagName]] = fmt.Sprintf("%d", value.(int))
			default:
				frames[nameToID3V2TagName[tagName]] = value.(string)
			}
		}
	}
	data := createID3v2TaggedData(audio, frames)
	data = append(data, createID3V1TaggedData(m)...)
	return data
}

func TestTrackLoadMetadata(t *testing.T) {
	originalFileSystem := cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmdtoolkit.AssignFileSystem(originalFileSystem)
	}()
	testDir := "loadMetadata"
	_ = cmdtoolkit.Mkdir(testDir)
	artistName := "A great artist"
	albumName := "A really good album"
	trackName := "A brilliant track"
	genre := "classic rock"
	year := "2022"
	track := 5
	payload := createConsistentlyTaggedData([]byte{0, 1, 2}, map[string]any{
		"artist": artistName,
		"album":  albumName,
		"title":  trackName,
		"genre":  genre,
		"year":   year,
		"track":  track,
	})
	fileName := "05 A brilliant track.mp3"
	_ = createFileWithContent(testDir, fileName, payload)
	postReadTm := NewTrackMetadata()
	for _, src := range []SourceType{ID3V1, ID3V2} {
		postReadTm.SetArtistName(src, artistName)
		postReadTm.SetAlbumName(src, albumName)
		postReadTm.SetAlbumGenre(src, genre)
		postReadTm.SetAlbumYear(src, year)
		postReadTm.SetTrackName(src, trackName)
		postReadTm.SetTrackNumber(src, track)
	}
	postReadTm.SetCDIdentifier([]byte{0})
	postReadTm.SetCanonicalSource(ID3V2)
	tests := map[string]struct {
		t    *Track
		want *TrackMetadata
	}{
		"no read needed": {
			t:    &Track{Metadata: NewTrackMetadata()},
			want: NewTrackMetadata()},
		"read file": {
			t:    &Track{FilePath: filepath.Join(testDir, fileName)},
			want: postReadTm,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			bar := pb.New(1)
			bar.SetWriter(output.NewNilBus().ErrorWriter())
			bar.Start()
			tt.t.LoadMetadata(bar)
			WaitForFilesClosed()
			bar.Finish()
			if !reflect.DeepEqual(tt.t.Metadata, tt.want) {
				t.Errorf("Track.LoadMetadata() got %#v want %#v", tt.t.Metadata, tt.want)
			}
		})
	}
}

func TestReadMetadata(t *testing.T) {
	originalFileSystem := cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmdtoolkit.AssignFileSystem(originalFileSystem)
	}()
	// 5 artists, 20 albums each, 50 tracks apiece ... total: 5,000 tracks
	testDir := "ReadMetadata"
	_ = cmdtoolkit.Mkdir(testDir)
	var artists []*Artist
	for k := 0; k < 5; k++ {
		artistName := fmt.Sprintf("artist %d", k)
		artistPath := filepath.Join(testDir, artistName)
		_ = cmdtoolkit.Mkdir(artistPath)
		artist := NewArtist(artistName, artistPath)
		artists = append(artists, artist)
		for m := 0; m < 20; m++ {
			albumName := fmt.Sprintf("album %d-%d", k, m)
			albumPath := filepath.Join(artistPath, albumName)
			_ = cmdtoolkit.Mkdir(albumPath)
			album := AlbumMaker{
				Title:  albumName,
				Artist: artist,
				Path:   albumName,
			}.NewAlbum()
			artist.AddAlbum(album)
			for n := 0; n < 50; n++ {
				trackName := fmt.Sprintf("track %d-%d-%d", k, m, n)
				trackFileName := fmt.Sprintf("%02d %s.mp3", n+1, trackName)
				track := &Track{
					FilePath:   filepath.Join(albumPath, trackFileName),
					SimpleName: trackName,
					Album:      album,
					Number:     n + 1,
					Metadata:   nil,
				}
				metadata := map[string]any{
					"artist": artistName,
					"album":  albumName,
					"title":  trackName,
					"genre":  "Classic Rock",
					"year":   "2022",
					"track":  n + 1,
				}
				content := createConsistentlyTaggedData([]byte{0, 1, 2, 3, 4, 5, 6, byte(k), byte(m), byte(n)}, metadata)
				_ = createFileWithContent(albumPath, trackFileName, content)
				album.AddTrack(track)
			}
		}
	}
	tests := map[string]struct {
		artists []*Artist
		output.WantedRecording
	}{
		"thorough test": {
			artists:         artists,
			WantedRecording: output.WantedRecording{Error: "Reading track metadata.\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			ReadMetadata(o, tt.artists)
			o.Report(t, "ReadMetadata()", tt.WantedRecording)
			for _, artist := range tt.artists {
				for _, album := range artist.Albums {
					for _, track := range album.Tracks {
						if track.needsMetadata() {
							t.Errorf("ReadMetadata() track %q has no metadata",
								track.FilePath)
						} else if track.hasMetadataError() {
							t.Errorf("ReadMetadata() track %q is defective: %v",
								track.FilePath, track.Metadata.ErrorCauses())
						}
					}
				}
			}
		})
	}
}

func TestTrack_ReportMetadataProblems(t *testing.T) {
	problematicArtist := NewArtist("problematic:artist", "")
	problematicAlbum := &Album{
		Title:           "problematic:album",
		RecordingArtist: problematicArtist,
		CanonicalGenre:  "hard rock",
		CanonicalYear:   "1999",
		CanonicalTitle:  "problematic:album",
	}
	src := ID3V2
	metadata := NewTrackMetadata()
	metadata.SetCanonicalSource(src)
	metadata.SetCDIdentifier([]byte{1, 3, 5})
	metadata.SetArtistName(src, "unknown artist")
	metadata.SetAlbumName(src, "unknown album")
	metadata.SetAlbumGenre(src, "unknown")
	metadata.SetAlbumYear(src, "2001")
	metadata.SetTrackName(src, "unknown title")
	metadata.SetTrackNumber(src, 2)
	problematicTrack := TrackMaker{
		Album:      problematicAlbum,
		FileName:   "03 bad track.mp3",
		SimpleName: "bad track",
		Number:     3,
	}.NewTrack()
	problematicTrack.Metadata = metadata
	problematicAlbum.AddTrack(problematicTrack)
	problematicArtist.AddAlbum(problematicAlbum)
	goodArtist := NewArtist("good artist", "")
	goodAlbum := &Album{
		Title:           "good album",
		RecordingArtist: goodArtist,
		CanonicalGenre:  "Classic Rock",
		CanonicalYear:   "1999",
		CanonicalTitle:  "good album",
	}
	src2 := ID3V1
	metadata2 := NewTrackMetadata()
	metadata2.SetCanonicalSource(src2)
	metadata2.SetArtistName(src2, "good artist")
	metadata2.SetAlbumName(src2, "good album")
	metadata2.SetAlbumGenre(src2, "Classic Rock")
	metadata2.SetAlbumYear(src2, "1999")
	metadata2.SetTrackName(src2, "good track")
	metadata2.SetTrackNumber(src2, 3)
	metadata2.SetErrorCause(ID3V2, "no id3v2 metadata, how odd")
	goodTrack := TrackMaker{
		Album:      goodAlbum,
		FileName:   "03 good track.mp3",
		SimpleName: "good track",
		Number:     3,
	}.NewTrack()
	goodTrack.Metadata = metadata2
	goodAlbum.AddTrack(goodTrack)
	goodArtist.AddAlbum(goodAlbum)
	errorMetadata := NewTrackMetadata()
	errorMetadata.SetErrorCause(ID3V1, "oops")
	errorMetadata.SetErrorCause(ID3V2, "oops")
	noMetadata := NewTrackMetadata()
	noMetadata.SetErrorCause(ID3V1, errNoID3V1MetadataFound.Error())
	noMetadata.SetErrorCause(ID3V2, errNoID3V2MetadataFound.Error())
	tests := map[string]struct {
		t    *Track
		want []string
	}{
		"unread metadata": {
			t:    &Track{Metadata: nil},
			want: []string{"differences cannot be determined: metadata has not been read"},
		},
		"track with error": {
			t:    &Track{Metadata: errorMetadata},
			want: []string{"differences cannot be determined: track metadata may be corrupted"},
		},
		"track with no metadata": {
			t:    &Track{Metadata: noMetadata},
			want: []string{"differences cannot be determined: the track file contains no metadata"},
		},
		"track with metadata differences": {
			t: problematicTrack,
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
		"track with no metadata differences": {t: goodTrack, want: nil},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.t.ReportMetadataProblems(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Track.ReportMetadataProblems() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_UpdateMetadata(t *testing.T) {
	// unfortunately, we cannot use a memory-mapped filesystem here, as the
	// library used for updating ID3V2 tags is hardcoded to use the os file
	// system.
	testDir := "updateMetadata"
	_ = cmdtoolkit.Mkdir(testDir)
	defer func() {
		_ = os.RemoveAll(testDir)
	}()
	trackName := "edit this track.mp3"
	trackContents := createConsistentlyTaggedData([]byte(trackName), map[string]any{
		"artist": "unknown artist",
		"album":  "unknown album",
		"title":  "unknown title",
		"genre":  "unknown",
		"year":   "1900",
		"track":  1,
	})
	_ = createFileWithContent(testDir, trackName, trackContents)
	expectedMetadata := NewTrackMetadata()
	for _, src := range []SourceType{ID3V1, ID3V2} {
		expectedMetadata.SetArtistName(src, "unknown artist")
		expectedMetadata.SetAlbumName(src, "unknown album")
		expectedMetadata.SetAlbumGenre(src, "unknown")
		expectedMetadata.SetAlbumYear(src, "1900")
		expectedMetadata.SetTrackName(src, "unknown title")
		expectedMetadata.SetTrackNumber(src, 1)
	}
	expectedMetadata.SetCanonicalSource(ID3V2)
	track := &Track{
		FilePath:   filepath.Join(testDir, trackName),
		SimpleName: strings.TrimSuffix(trackName, ".mp3"),
		Number:     2,
		Album: &Album{
			Title:          "fine album",
			CanonicalGenre: "classic rock",
			CanonicalYear:  "2022",
			CanonicalTitle: "fine album",
			MusicCDIdentifier: id3v2.UnknownFrame{
				Body: []byte("fine album"),
			},
			RecordingArtist: &Artist{
				Name:          "fine artist",
				CanonicalName: "fine artist",
			},
		},
		Metadata: expectedMetadata,
	}
	deletedTrack := &Track{
		FilePath:   filepath.Join(testDir, "no such file"),
		SimpleName: strings.TrimSuffix(trackName, ".mp3"),
		Number:     2,
		Album: &Album{
			Title:             "fine album",
			CanonicalGenre:    "classic rock",
			CanonicalYear:     "2022",
			CanonicalTitle:    "fine album",
			MusicCDIdentifier: id3v2.UnknownFrame{Body: []byte("fine album")},
			RecordingArtist: &Artist{
				Name:          "fine artist",
				CanonicalName: "fine artist",
			},
		},
		Metadata: expectedMetadata,
	}
	editedTm := NewTrackMetadata()
	for _, src := range []SourceType{ID3V1, ID3V2} {
		editedTm.SetArtistName(src, "fine artist")
		editedTm.SetAlbumName(src, "fine album")
		editedTm.SetAlbumGenre(src, "classic rock")
		editedTm.SetAlbumYear(src, "2022")
		editedTm.SetTrackName(src, "edit this track")
		editedTm.SetTrackNumber(src, 2)
	}
	editedTm.SetCDIdentifier([]byte("fine album"))
	editedTm.SetCanonicalSource(ID3V2)
	tests := map[string]struct {
		t      *Track
		wantE  []string
		wantTm *TrackMetadata
	}{
		"error checking": {
			t: deletedTrack,
			wantE: []string{
				"open updateMetadata\\no such file: The system cannot find the file specified.",
				"open updateMetadata\\no such file: The system cannot find the file specified.",
			},
		},
		"no edit required": {
			t:     &Track{Metadata: nil},
			wantE: []string{errNoEditNeeded.Error()},
		},
		"edit required": {t: track, wantTm: editedTm},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotE := tt.t.UpdateMetadata()
			var eStrings []string
			for _, e := range gotE {
				eStrings = append(eStrings, e.Error())
			}
			if !reflect.DeepEqual(eStrings, tt.wantE) {
				t.Errorf("Track.UpdateMetadata() = %v, want %v", eStrings, tt.wantE)
			} else if len(gotE) == 0 && tt.t.Metadata != nil {
				// verify file was correctly rewritten
				gotTm := InitializeMetadata(tt.t.FilePath)
				if !reflect.DeepEqual(gotTm, tt.wantTm) {
					t.Errorf("Track.UpdateMetadata() read %#v, want %#v", gotTm, tt.wantTm)
				}
			}
		})
	}
}

func TestProcessArtistMetadata(t *testing.T) {
	artist1 := NewArtist("artist_name", "")
	album1 := AlbumMaker{Title: "album1", Artist: artist1}.NewAlbum()
	artist1.AddAlbum(album1)
	for k := 1; k <= 10; k++ {
		src := ID3V2
		tm := NewTrackMetadata()
		tm.SetCanonicalSource(src)
		tm.SetArtistName(src, "artist:name")
		track := TrackMaker{
			Album:      album1,
			FileName:   fmt.Sprintf("%02d track%d.mp3", k, k),
			SimpleName: fmt.Sprintf("track%d", k),
			Number:     k,
		}.NewTrack()
		track.Metadata = tm
		album1.AddTrack(track)
	}
	artist2 := NewArtist("artist_name", "")
	album2 := AlbumMaker{Title: "album2", Artist: artist2}.NewAlbum()
	artist2.AddAlbum(album2)
	for k := 1; k <= 10; k++ {
		src := ID3V2
		tm := NewTrackMetadata()
		tm.SetCanonicalSource(src)
		tm.SetArtistName(src, "unknown artist")
		track := TrackMaker{
			Album:      album2,
			FileName:   fmt.Sprintf("%02d track%d.mp3", k, k),
			SimpleName: fmt.Sprintf("track%d", k),
			Number:     k,
		}.NewTrack()
		track.Metadata = tm
		album2.AddTrack(track)
	}
	artist3 := NewArtist("artist_name", "")
	album3 := AlbumMaker{Title: "album3", Artist: artist3}.NewAlbum()
	artist3.AddAlbum(album3)
	for k := 1; k <= 10; k++ {
		src := ID3V2
		tm := NewTrackMetadata()
		tm.SetCanonicalSource(src)
		if k%2 == 0 {
			tm.SetArtistName(src, "artist:name")
		} else {
			tm.SetArtistName(src, "artist_name")
		}
		track := TrackMaker{
			Album:      album3,
			FileName:   fmt.Sprintf("%02d track%d.mp3", k, k),
			SimpleName: fmt.Sprintf("track%d", k),
			Number:     k,
		}.NewTrack()
		track.Metadata = tm
		album3.AddTrack(track)
	}
	tests := map[string]struct {
		artists []*Artist
		output.WantedRecording
	}{
		"unanimous choice": {artists: []*Artist{artist1}},
		"unknown choice":   {artists: []*Artist{artist2}},
		"ambiguous choice": {
			artists: []*Artist{artist3},
			WantedRecording: output.WantedRecording{
				Error: "There are multiple artist name fields for \"artist_name\"," +
					" and there is no unambiguously preferred choice; candidates are" +
					" {\"artist:name\": 5 instances, \"artist_name\": 5 instances}.\n",
				Log: "level='error'" +
					" artistName='artist_name'" +
					" field='artist name'" +
					" settings='map[artist:name:5 artist_name:5]'" +
					" msg='no value has a majority of instances'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			ProcessArtistMetadata(o, tt.artists)
			o.Report(t, "ProcessArtistMetadata()", tt.WantedRecording)
		})
	}
}

func TestProcessAlbumMetadata(t *testing.T) {
	// ordinary test data
	src := ID3V2
	var artists1 []*Artist
	artist1 := NewArtist("good artist", "")
	artists1 = append(artists1, artist1)
	album1 := AlbumMaker{Title: "good-album", Artist: artist1}.NewAlbum()
	artist1.AddAlbum(album1)
	tm := NewTrackMetadata()
	tm.SetCanonicalSource(src)
	tm.SetAlbumName(src, "good:album")
	tm.SetAlbumGenre(src, "pop")
	tm.SetAlbumYear(src, "2022")
	track1 := TrackMaker{
		Album:      album1,
		FileName:   "01 track1.mp3",
		SimpleName: "track1",
		Number:     1,
	}.NewTrack()
	track1.Metadata = tm
	album1.AddTrack(track1)
	// more interesting test data
	var artists2 []*Artist
	artist2 := NewArtist("another good artist", "")
	artists2 = append(artists2, artist2)
	album2 := AlbumMaker{
		Title:  "another good_album",
		Artist: artist2,
	}.NewAlbum()
	artist2.AddAlbum(album2)
	tm2a := NewTrackMetadata()
	tm2a.SetCanonicalSource(src)
	tm2a.SetAlbumName(src, "unknown album")
	tm2a.SetAlbumGenre(src, "unknown")
	tm2a.SetAlbumYear(src, "")
	track2a := TrackMaker{
		Album:      album2,
		FileName:   "01 track1.mp3",
		SimpleName: "track1",
		Number:     1,
	}.NewTrack()
	track2a.Metadata = tm2a
	album2.AddTrack(track2a)
	tm2b := NewTrackMetadata()
	tm2b.SetCanonicalSource(src)
	tm2b.SetAlbumName(src, "another good:album")
	tm2b.SetAlbumGenre(src, "pop")
	tm2b.SetAlbumYear(src, "2022")
	track2b := TrackMaker{
		Album:      album1,
		FileName:   "02 track2.mp3",
		SimpleName: "track2",
		Number:     2,
	}.NewTrack()
	track2b.Metadata = tm2b
	album2.AddTrack(track2b)
	tm2c := NewTrackMetadata()
	tm2c.SetCanonicalSource(src)
	tm2c.SetAlbumName(src, "another good:album")
	tm2c.SetAlbumGenre(src, "pop")
	tm2c.SetAlbumYear(src, "2022")
	track2c := TrackMaker{
		Album:      album1,
		FileName:   "03 track3.mp3",
		SimpleName: "track3",
		Number:     3,
	}.NewTrack()
	track2c.Metadata = tm2c
	album2.AddTrack(track2c)
	// error case data
	var artists3 []*Artist
	artist3 := NewArtist("problematic artist", "")
	artists3 = append(artists3, artist3)
	album3 := AlbumMaker{
		Title:  "problematic_album",
		Artist: artist3,
	}.NewAlbum()
	artist3.AddAlbum(album3)
	tm3a := NewTrackMetadata()
	tm3a.SetCanonicalSource(src)
	tm3a.SetCDIdentifier([]byte{1, 2, 3})
	tm3a.SetAlbumName(src, "problematic:album")
	tm3a.SetAlbumGenre(src, "rock")
	tm3a.SetAlbumYear(src, "2023")
	track3a := TrackMaker{
		Album:      album2,
		FileName:   "01 track1.mp3",
		SimpleName: "track1",
		Number:     1,
	}.NewTrack()
	track3a.Metadata = tm3a
	album3.AddTrack(track3a)
	tm3b := NewTrackMetadata()
	tm3b.SetCanonicalSource(src)
	tm3b.SetCDIdentifier([]byte{1, 2, 3, 4})
	tm3b.SetAlbumName(src, "problematic:Album")
	tm3b.SetAlbumGenre(src, "pop")
	tm3b.SetAlbumYear(src, "2022")
	track3b := TrackMaker{
		Album:      album1,
		FileName:   "02 track2.mp3",
		SimpleName: "track2",
		Number:     2,
	}.NewTrack()
	track3b.Metadata = tm3b
	album3.AddTrack(track3b)
	tm3c := NewTrackMetadata()
	tm3c.SetCanonicalSource(src)
	tm3c.SetCDIdentifier([]byte{1, 2, 3, 4, 5})
	tm3c.SetAlbumName(src, "Problematic:album")
	tm3c.SetAlbumGenre(src, "folk")
	tm3c.SetAlbumYear(src, "2021")
	track3c := TrackMaker{
		Album:      album1,
		FileName:   "03 track3.mp3",
		SimpleName: "track3",
		Number:     3,
	}.NewTrack()
	track3c.Metadata = tm3c
	album3.AddTrack(track3c)
	// verify code can handle missing metadata
	track4 := TrackMaker{
		Album:      album1,
		FileName:   "04 track4.mp3",
		SimpleName: "track4",
		Number:     4,
	}.NewTrack()
	album3.AddTrack(track4)
	tests := map[string]struct {
		artists []*Artist
		output.WantedRecording
	}{
		"ordinary test":    {artists: artists1},
		"typical use case": {artists: artists2},
		"errors": {
			artists: artists3,
			WantedRecording: output.WantedRecording{
				Error: "There are multiple genre fields for \"problematic_album by" +
					" problematic artist\", and there is no unambiguously preferred choice;" +
					" candidates are {\"folk\": 1 instance, \"pop\": 1 instance, \"rock\":" +
					" 1 instance}.\n" +
					"There are multiple year fields for \"problematic_album by" +
					" problematic artist\", and there is no unambiguously preferred" +
					" choice; candidates are {\"2021\": 1 instance, \"2022\": 1 instance," +
					" \"2023\": 1 instance}.\n" +
					"There are multiple album title fields for \"problematic_album by" +
					" problematic artist\", and there is no unambiguously preferred" +
					" choice; candidates are {\"Problematic:album\": 1 instance," +
					" \"problematic:Album\": 1 instance, \"problematic:album\": 1 instance}.\n" +
					"There are multiple MCDI frame fields for \"problematic_album by" +
					" problematic artist\", and there is no unambiguously preferred" +
					" choice; candidates are {\"\\x01\\x02\\x03\": 1 instance," +
					" \"\\x01\\x02\\x03\\x04\": 1 instance," +
					" \"\\x01\\x02\\x03\\x04\\x05\": 1 instance}.\n",
				Log: "level='error'" +
					" albumName='problematic_album'" +
					" artistName='problematic artist'" +
					" field='genre'" +
					" settings='map[folk:1 pop:1 rock:1]'" +
					" msg='no value has a majority of instances'\n" +
					"level='error'" +
					" albumName='problematic_album'" +
					" artistName='problematic artist'" +
					" field='year'" +
					" settings='map[2021:1 2022:1 2023:1]'" +
					" msg='no value has a majority of instances'\n" +
					"level='error'" +
					" albumName='problematic_album'" +
					" artistName='problematic artist'" +
					" field='album title'" +
					" settings='map[Problematic:album:1 problematic:Album:1" +
					" problematic:album:1]'" +
					" msg='no value has a majority of instances'\n" +
					"level='error'" +
					" albumName='problematic_album'" +
					" artistName='problematic artist'" +
					" field='mcdi frame'" +
					" settings='map[\x01\x02\x03:1 \x01\x02\x03\x04:1" +
					" \x01\x02\x03\x04\x05:1]'" +
					" msg='no value has a majority of instances'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			ProcessAlbumMetadata(o, tt.artists)
			o.Report(t, "ProcessAlbumMetadata()", tt.WantedRecording)
		})
	}
}

func TestTrack_ReportMetadataErrors(t *testing.T) {
	tm := NewTrackMetadata()
	tm.SetErrorCause(ID3V1, "id3v1 error!")
	tm.SetErrorCause(ID3V2, "id3v2 error!")
	tests := map[string]struct {
		t *Track
		output.WantedRecording
	}{
		"error handling": {
			t: &Track{
				SimpleName: "silly track",
				FilePath:   "Music\\silly artist\\silly album\\01 silly track.mp3",
				Metadata:   tm,
				Album: &Album{
					Title:           "silly album",
					RecordingArtist: &Artist{Name: "silly artist"},
				},
			},
			WantedRecording: output.WantedRecording{
				Log: "level='error'" +
					" error='id3v1 error!'" +
					" metadata='ID3V1'" +
					" track='Music\\silly artist\\silly album\\01 silly track.mp3'" +
					" msg='metadata read error'\n" +
					"level='error'" +
					" error='id3v2 error!'" +
					" metadata='ID3V2'" +
					" track='Music\\silly artist\\silly album\\01 silly track.mp3'" +
					" msg='metadata read error'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.t.ReportMetadataErrors(o)
			o.Report(t, "Track.ReportMetadataErrors()", tt.WantedRecording)
		})
	}
}

func TestTrack_Details(t *testing.T) {
	originalFileSystem := cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmdtoolkit.AssignFileSystem(originalFileSystem)
	}()
	audio := make([]byte, 0)
	for k := 0; k < 256; k++ {
		audio = append(audio, byte(k))
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
	content := createID3v2TaggedData(audio, frames)
	goodFileName := "goodFile.mp3"
	_ = createFileWithContent(".", goodFileName, content)
	tests := map[string]struct {
		t       *Track
		want    map[string]string
		wantErr bool
	}{
		"error case": {
			t:       &Track{FilePath: "./no such file"},
			wantErr: true,
		},
		"good case": {
			t: &Track{FilePath: filepath.Join(".", goodFileName)},
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
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotErr := tt.t.Details()
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("Track.Details() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Track.Details() = %v, want %v", got, tt.want)
			}
		})
	}
}

type sampleWriter struct {
	name string
}

func (sw *sampleWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

type sampleBus struct {
	consoleTTY    bool
	errorTTY      bool
	consoleWriter io.Writer
	errorWriter   io.Writer
}

func (sB *sampleBus) Log(_ output.Level, _ string, _ map[string]any) {}
func (sB *sampleBus) WriteCanonicalConsole(_ string, _ ...any)       {}
func (sB *sampleBus) WriteConsole(_ string, _ ...any)                {}
func (sB *sampleBus) WriteCanonicalError(_ string, _ ...any)         {}
func (sB *sampleBus) WriteError(_ string, _ ...any)                  {}
func (sB *sampleBus) ConsoleWriter() io.Writer {
	return sB.consoleWriter
}
func (sB *sampleBus) ErrorWriter() io.Writer {
	return sB.errorWriter
}
func (sB *sampleBus) IsConsoleTTY() bool {
	return sB.consoleTTY
}
func (sB *sampleBus) IsErrorTTY() bool {
	return sB.errorTTY
}
func (sB *sampleBus) IncrementTab(_ uint8) {}
func (sB *sampleBus) DecrementTab(_ uint8) {}
func (sB *sampleBus) Tab() uint8 {
	return 0
}

func TestProgressWriter(t *testing.T) {
	errorWriter := &sampleWriter{name: "error"}
	consoleWriter := &sampleWriter{name: "console"}
	tests := map[string]struct {
		o    output.Bus
		want io.Writer
	}{
		"error is TTY": {
			o: &sampleBus{
				errorWriter: errorWriter,
				errorTTY:    true,
			},
			want: errorWriter,
		},
		"console is TTY": {
			o: &sampleBus{
				consoleWriter: consoleWriter,
				consoleTTY:    true,
			},
			want: consoleWriter,
		},
		"no TTY": {
			o:    &sampleBus{},
			want: output.NilWriter{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := ProgressWriter(tt.o); got != tt.want {
				t.Errorf("ProgressWriter() = %v, want %v", got, tt.want)
			}
		})
	}
}

// createFileWithContent creates a file in a specified directory.
func createFileWithContent(dir, name string, content []byte) error {
	fileName := filepath.Join(dir, name)
	return createNamedFile(fileName, content)
}

// createNamedFile creates a specified name with the specified content.
func createNamedFile(fileName string, content []byte) (fileErr error) {
	fs := cmdtoolkit.FileSystem()
	_, fileErr = fs.Stat(fileName)
	if fileErr == nil {
		fileErr = fmt.Errorf("file %q already exists", fileName)
	} else if errors.Is(fileErr, afero.ErrFileNotFound) {
		fileErr = afero.WriteFile(fs, fileName, content, cmdtoolkit.StdFilePermissions)
	}
	return
}

// createFile creates a file in a specified directory with standardized content
func createFile(dir, name string) (err error) {
	return createFileWithContent(dir, name, []byte("file contents for "+name))
}
