package files_test

import (
	"errors"
	"fmt"
	"io"
	"mp3repair/internal/files"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/bogem/id3v2/v2"
	"github.com/cheggaaa/pb/v3"
	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/afero"
)

func TestParseTrackName(t *testing.T) {
	tests := map[string]struct {
		parser         files.TrackNameParser
		wantParsedName *files.ParsedTrackName
		wantValid      bool
		output.WantedRecording
	}{
		"expected use case": {
			parser: files.TrackNameParser{
				FileName: "59 track name.mp3",
				Album: &files.Album{
					Title:           "some album",
					RecordingArtist: &files.Artist{Name: "some artist"},
				},
				Extension: ".mp3",
			},
			wantParsedName: &files.ParsedTrackName{SimpleName: "track name", Number: 59},
			wantValid:      true,
		},
		"expected use case with hyphen separator": {
			parser: files.TrackNameParser{
				FileName: "60-other track name.mp3",
				Album: &files.Album{
					Title:           "some album",
					RecordingArtist: &files.Artist{Name: "some artist"},
				},
				Extension: ".mp3",
			},
			wantParsedName: &files.ParsedTrackName{
				SimpleName: "other track name",
				Number:     60,
			},
			wantValid: true,
		},
		"wrong extension": {
			parser: files.TrackNameParser{
				FileName: "59 track name.mp4",
				Album: &files.Album{
					Title:           "some album",
					RecordingArtist: &files.Artist{Name: "some artist"},
				},
				Extension: ".mp3",
			},
			wantParsedName: &files.ParsedTrackName{
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
			parser: files.TrackNameParser{
				FileName: "track name.mp3",
				Album: &files.Album{
					Title:           "some album",
					RecordingArtist: &files.Artist{Name: "some artist"},
				},
				Extension: ".mp3",
			},
			wantParsedName: &files.ParsedTrackName{SimpleName: "name"},
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
			parser: files.TrackNameParser{
				FileName: "trackName.mp3",
				Album: &files.Album{
					Title:           "some album",
					RecordingArtist: &files.Artist{Name: "some artist"},
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
					t.Errorf("ParseTrackName() gotParsedName = %v, want %v", gotParsedName, tt.wantParsedName)
				}
			}
			if gotValid != tt.wantValid {
				t.Errorf("ParseTrackName() gotValid = %v, want %v", gotValid, tt.wantValid)
			}
			o.Report(t, "ParseTrackName()", tt.WantedRecording)
		})
	}
}

func Test_sortTracks(t *testing.T) {
	const fnName = "sortTracks()"
	tests := map[string]struct {
		tracks []*files.Track
	}{
		"degenerate case": {},
		"mixed tracks": {
			tracks: []*files.Track{
				{
					Number: 10,
					Album: files.AlbumMaker{
						Title:  "album2",
						Artist: files.NewArtist("artist3", ""),
					}.NewAlbum(),
				},
				{
					Number: 1,
					Album: files.AlbumMaker{
						Title:  "album2",
						Artist: files.NewArtist("artist3", ""),
					}.NewAlbum(),
				},
				{
					Number: 3,
					Album: files.AlbumMaker{
						Title:  "album3",
						Artist: files.NewArtist("artist2", ""),
					}.NewAlbum(),
				},
				{
					Number: 3,
					Album: files.AlbumMaker{
						Title:  "album3",
						Artist: files.NewArtist("artist4", ""),
					}.NewAlbum(),
				},
				{
					Number: 3,
					Album: files.AlbumMaker{
						Title:  "album5",
						Artist: files.NewArtist("artist2", ""),
					}.NewAlbum(),
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			sort.Sort(files.Tracks(tt.tracks))
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
					t.Errorf("%s track[%d] artist name %q comes after"+
						" track[%d] artist name %q", fnName, i-1, artist1, i, artist2)
				} else if artist1 == artist2 {
					if album1.Title > album2.Title {
						t.Errorf("%s track[%d] album name %q comes after"+
							" track[%d] album name %q", fnName, i-1, album1.Title, i,
							album2.Title)
					} else if album1.Title == album2.Title {
						if track1.Number > track2.Number {
							t.Errorf("%s track[%d] track %d comes after"+
								" track[%d] track %d", fnName, i-1, track1.Number, i,
								track2.Number)
						}
					}
				}
			}
		})
	}
}

func TestTrack_AlbumPath(t *testing.T) {
	const fnName = "Track.AlbumPath()"
	tests := map[string]struct {
		t    *files.Track
		want string
	}{
		"no containing album": {t: &files.Track{}, want: ""},
		"has containing album": {
			t:    &files.Track{Album: files.AlbumMaker{Path: "album-path"}.NewAlbum()},
			want: "album-path"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.t.AlbumPath(); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestTrack_CopyFile(t *testing.T) {
	originalFileSystem := cmd_toolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmd_toolkit.AssignFileSystem(originalFileSystem)
	}()
	const fnName = "Track.CopyFile()"
	topDir := "copies"
	cmd_toolkit.Mkdir(topDir)
	srcName := "source.mp3"
	srcPath := filepath.Join(topDir, srcName)
	createFile(topDir, srcName)
	type args struct {
		destination string
	}
	tests := map[string]struct {
		t *files.Track
		args
		wantErr bool
	}{
		"error case": {
			t:       &files.Track{FilePath: "no such file"},
			args:    args{destination: filepath.Join(topDir, "destination.mp3")},
			wantErr: true,
		},
		"good case": {
			t:       &files.Track{FilePath: srcPath},
			args:    args{destination: filepath.Join(topDir, "destination.mp3")},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if gotErr := tt.t.CopyFile(tt.args.destination); (gotErr != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, gotErr, tt.wantErr)
			}
		})
	}
}

func TestTrackStringType(t *testing.T) {
	const fnName = "Track.String()"
	tests := map[string]struct {
		t    *files.Track
		want string
	}{
		"expected": {
			t:    &files.Track{FilePath: "my path"},
			want: "my path",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.t.String(); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestTrack_AlbumName(t *testing.T) {
	tests := map[string]struct {
		t    *files.Track
		want string
	}{
		"orphan track": {t: &files.Track{}, want: ""},
		"good track": {
			t:    &files.Track{Album: &files.Album{Title: "my album name"}},
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
	const fnName = "Track.RecordingArtist()"
	tests := map[string]struct {
		t    *files.Track
		want string
	}{
		"orphan track": {t: &files.Track{}, want: ""},
		"good track": {
			t: &files.Track{
				Album: &files.Album{RecordingArtist: &files.Artist{Name: "my artist"}},
			},
			want: "my artist",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.t.RecordingArtist(); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestTrack_Directory(t *testing.T) {
	const fnName = "Track.Directory()"
	tests := map[string]struct {
		t    *files.Track
		want string
	}{
		"typical": {
			t:    &files.Track{FilePath: "Music/my artist/my album/03 track.mp3"},
			want: "Music\\my artist\\my album",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.t.Directory(); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestTrack_FileName(t *testing.T) {
	const fnName = "Track.FileName()"
	tests := map[string]struct {
		t    *files.Track
		want string
	}{
		"typical": {
			t:    &files.Track{FilePath: "Music/my artist/my album/03 track.mp3"},
			want: "03 track.mp3",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.t.FileName(); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func Test_pickKey(t *testing.T) {
	const fnName = "pickKey()"
	type args struct {
		m map[string]int
	}
	tests := map[string]struct {
		args
		wantS  string
		wantOk bool
	}{
		"unanimous choice": {
			args:   args{m: map[string]int{"pop": 2}},
			wantS:  "pop",
			wantOk: true,
		},
		"majority for even size": {
			args:   args{m: map[string]int{"pop": 3, "": 1}},
			wantS:  "pop",
			wantOk: true,
		},
		"majority for odd size": {
			args:   args{m: map[string]int{"pop": 2, "": 1}},
			wantS:  "pop",
			wantOk: true,
		},
		"no majority even size": {
			args: args{m: map[string]int{"pop": 1, "alt-rock": 1}},
		},
		"no majority odd size": {
			args: args{m: map[string]int{"pop": 2, "alt-rock": 2, "folk": 1}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotS, gotOk := files.CanonicalChoice(tt.args.m)
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
	originalFileSystem := cmd_toolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmd_toolkit.AssignFileSystem(originalFileSystem)
	}()
	const fnName = "Track.ID3V2Diagnostics()"
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
		"Fake": "ummm",
	}
	content := createID3v2TaggedData(audio, frames)
	goodFileName := "goodFile.mp3"
	createFileWithContent(".", goodFileName, content)
	tests := map[string]struct {
		t                *files.Track
		wantEncoding     string
		wantVersion      byte
		wantFrameStrings []string
		wantErr          bool
	}{
		"error case": {
			t:       &files.Track{FilePath: "./no such file"},
			wantErr: true,
		},
		"good case": {
			t:            &files.Track{FilePath: filepath.Join(".", goodFileName)},
			wantEncoding: "ISO-8859-1",
			wantVersion:  3,
			wantFrameStrings: []string{
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
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotInfo, gotErr := tt.t.ID3V2Diagnostics()
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, gotErr, tt.wantErr)
				return
			}
			if gotErr == nil {
				if gotInfo.Encoding != tt.wantEncoding {
					t.Errorf("%s gotInfo.Encoding = %q, want %q", fnName, gotInfo.Encoding, tt.wantEncoding)
				}
				if gotInfo.Version != tt.wantVersion {
					t.Errorf("%s gotInfo.Version = %d, want %d", fnName, gotInfo.Version, tt.wantVersion)
				}
				if !reflect.DeepEqual(gotInfo.FrameStrings, tt.wantFrameStrings) {
					t.Errorf("%s gotInfo.FrameStrings = %v, want %v", fnName, gotInfo.FrameStrings, tt.wantFrameStrings)
				}
			}
		})
	}
}

func TestTrack_ID3V1Diagnostics(t *testing.T) {
	originalFileSystem := cmd_toolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmd_toolkit.AssignFileSystem(originalFileSystem)
	}()
	const fnName = "Track.ID3V1Diagnostics()"
	testDir := "id3v1Diagnostics"
	cmd_toolkit.Mkdir(testDir)
	// three files: one good, one too small, one with an invalid tag
	smallFile := "01 small.mp3"
	createFileWithContent(testDir, smallFile, []byte{0, 1, 2})
	invalidFile := "02 invalid.mp3"
	createFileWithContent(testDir, invalidFile, []byte{
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
	createFileWithContent(testDir, goodFile, []byte{
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
		t       *files.Track
		want    []string
		wantErr bool
	}{
		"small file": {
			t:       &files.Track{FilePath: filepath.Join(testDir, smallFile)},
			wantErr: true,
		},
		"invalid file": {
			t:       &files.Track{FilePath: filepath.Join(testDir, invalidFile)},
			wantErr: true,
		},
		"good file": {
			t: &files.Track{FilePath: filepath.Join(testDir, goodFile)},
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
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotErr := tt.t.ID3V1Diagnostics()
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, gotErr, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
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
	v1 := files.NewID3v1Metadata()
	v1.WriteString("TAG", files.TagField)
	for _, tagName := range recognizedTagNames {
		if value, tagNameFound := m[tagName]; tagNameFound {
			switch tagName {
			case "artist":
				v1.SetArtist(value.(string))
			case "album":
				v1.SetAlbum(value.(string))
			case "title":
				v1.SetTitle(value.(string))
			case "genre":
				v1.SetGenre(value.(string))
			case "year":
				v1.SetYear(value.(string))
			case "track":
				v1.SetTrack(value.(int))
			}
		}
	}
	return v1.RawData()
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
	originalFileSystem := cmd_toolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmd_toolkit.AssignFileSystem(originalFileSystem)
	}()
	testDir := "loadMetadata"
	cmd_toolkit.Mkdir(testDir)
	artistName := "A great artist"
	albumName := "A really good album"
	trackName := "A brilliant track"
	genre := "Classic Rock"
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
	createFileWithContent(testDir, fileName, payload)
	tests := map[string]struct {
		t    *files.Track
		want *files.TrackMetadataV1
	}{
		"no read needed": {
			t:    &files.Track{Metadata: &files.TrackMetadataV1{}},
			want: &files.TrackMetadataV1{}},
		"read file": {
			t: &files.Track{FilePath: filepath.Join(testDir, fileName)},
			want: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", albumName, albumName}).WithArtistNames([]string{
				"", artistName, artistName}).WithTrackNames([]string{
				"", trackName, trackName}).WithGenres([]string{
				"", genre, genre}).WithYears([]string{"", year, year}).WithTrackNumbers(
				[]int{0, track, track}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			bar := pb.New(1)
			bar.SetWriter(output.NewNilBus().ErrorWriter())
			bar.Start()
			tt.t.LoadMetadata(bar)
			files.WaitForFilesClosed()
			bar.Finish()
			if !reflect.DeepEqual(tt.t.Metadata, tt.want) {
				t.Errorf("track.LoadMetadata() got %#v want %#v", tt.t.Metadata, tt.want)
			}
		})
	}
}

func TestReadMetadata(t *testing.T) {
	originalFileSystem := cmd_toolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmd_toolkit.AssignFileSystem(originalFileSystem)
	}()
	const fnName = "ReadMetadata()"
	// 5 artists, 20 albums each, 50 tracks apiece ... total: 5,000 tracks
	testDir := "ReadMetadata"
	cmd_toolkit.Mkdir(testDir)
	var artists []*files.Artist
	for k := 0; k < 5; k++ {
		artistName := fmt.Sprintf("artist %d", k)
		artistPath := filepath.Join(testDir, artistName)
		cmd_toolkit.Mkdir(artistPath)
		artist := files.NewArtist(artistName, artistPath)
		artists = append(artists, artist)
		for m := 0; m < 20; m++ {
			albumName := fmt.Sprintf("album %d-%d", k, m)
			albumPath := filepath.Join(artistPath, albumName)
			cmd_toolkit.Mkdir(albumPath)
			album := files.AlbumMaker{
				Title:  albumName,
				Artist: artist,
				Path:   albumName,
			}.NewAlbum()
			artist.AddAlbum(album)
			for n := 0; n < 50; n++ {
				trackName := fmt.Sprintf("track %d-%d-%d", k, m, n)
				trackFileName := fmt.Sprintf("%02d %s.mp3", n+1, trackName)
				track := &files.Track{
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
				createFileWithContent(albumPath, trackFileName, content)
				album.AddTrack(track)
			}
		}
	}
	type args struct {
		artists []*files.Artist
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"thorough test": {
			args:            args{artists: artists},
			WantedRecording: output.WantedRecording{Error: "Reading track metadata.\n"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			files.ReadMetadata(o, tt.args.artists)
			o.Report(t, fnName, tt.WantedRecording)
			for _, artist := range tt.args.artists {
				for _, album := range artist.Albums {
					for _, track := range album.Tracks {
						if track.NeedsMetadata() {
							t.Errorf("%s track %q has no metadata", fnName, track.FilePath)
						} else if track.HasMetadataError() {
							t.Errorf("%s track %q is defective: %v", fnName, track.FilePath, track.Metadata.ErrorCauses())
						}
					}
				}
			}
		})
	}
}

func TestTrack_ReportMetadataProblems(t *testing.T) {
	const fnName = "Track.ReportMetadataProblems()"
	problematicArtist := files.NewArtist("problematic:artist", "")
	problematicAlbum := &files.Album{
		Title:           "problematic:album",
		RecordingArtist: problematicArtist,
		CanonicalGenre:  "hard rock",
		CanonicalYear:   "1999",
		CanonicalTitle:  "problematic:album",
	}
	src := files.ID3V2
	metadata := files.NewTrackMetadataV1().WithPrimarySource(src).WithMusicCDIdentifier(
		[]byte{1, 3, 5})
	metadata.SetAlbumName(src, "unknown album")
	metadata.SetArtistName(src, "unknown artist")
	metadata.SetGenre(src, "unknown")
	metadata.SetTrackName(src, "unknown title")
	metadata.SetTrackNumber(src, 2)
	metadata.SetYear(src, "2001")
	problematicTrack := files.TrackMaker{
		Album:      problematicAlbum,
		FileName:   "03 bad track.mp3",
		SimpleName: "bad track",
		Number:     3,
	}.NewTrack()
	problematicTrack.Metadata = metadata
	problematicAlbum.AddTrack(problematicTrack)
	problematicArtist.AddAlbum(problematicAlbum)
	goodArtist := files.NewArtist("good artist", "")
	goodAlbum := &files.Album{
		Title:           "good album",
		RecordingArtist: goodArtist,
		CanonicalGenre:  "Classic Rock",
		CanonicalYear:   "1999",
		CanonicalTitle:  "good album",
	}
	src2 := files.ID3V1
	metadata2 := files.NewTrackMetadataV1().WithPrimarySource(src2)
	metadata2.SetAlbumName(src2, "good album")
	metadata2.SetArtistName(src2, "good artist")
	metadata2.SetErrorCause(files.ID3V2, "no id3v2 metadata, how odd")
	metadata2.SetGenre(src2, "Classic Rock")
	metadata2.SetTrackName(src2, "good track")
	metadata2.SetTrackNumber(src2, 3)
	metadata2.SetYear(src2, "1999")
	goodTrack := files.TrackMaker{
		Album:      goodAlbum,
		FileName:   "03 good track.mp3",
		SimpleName: "good track",
		Number:     3,
	}.NewTrack()
	goodTrack.Metadata = metadata2
	goodAlbum.AddTrack(goodTrack)
	goodArtist.AddAlbum(goodAlbum)
	tests := map[string]struct {
		t    *files.Track
		want []string
	}{
		"unread metadata": {
			t:    &files.Track{Metadata: nil},
			want: []string{"differences cannot be determined: metadata has not been read"},
		},
		"track with error": {
			t: &files.Track{Metadata: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", "oops", "oops"}),
			},
			want: []string{"differences cannot be determined: track metadata may be corrupted"},
		},
		"track with no metadata": {
			t: &files.Track{
				Metadata: files.NewTrackMetadataV1().WithErrorCauses([]string{
					"",
					files.ErrNoID3V1MetadataFound.Error(),
					files.ErrNoID3V2MetadataFound.Error(),
				}),
			},
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
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestTrack_UpdateMetadata(t *testing.T) {
	// unfortunately, we cannot use a memory-mapped filesystem here, as the
	// library used for updating ID3V2 tags is hardcoded to use the os file
	// system.
	const fnName = "Track.UpdateMetadata()"
	testDir := "updateMetadata"
	cmd_toolkit.Mkdir(testDir)
	defer func() {
		os.RemoveAll(testDir)
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
	createFileWithContent(testDir, trackName, trackContents)
	track := &files.Track{
		FilePath:   filepath.Join(testDir, trackName),
		SimpleName: strings.TrimSuffix(trackName, ".mp3"),
		Number:     2,
		Album: &files.Album{
			Title:          "fine album",
			CanonicalGenre: "Classic Rock",
			CanonicalYear:  "2022",
			CanonicalTitle: "fine album",
			MusicCDIdentifier: id3v2.UnknownFrame{
				Body: []byte("fine album"),
			},
			RecordingArtist: &files.Artist{
				Name:          "fine artist",
				CanonicalName: "fine artist",
			},
		},
		Metadata: files.NewTrackMetadataV1().WithAlbumNames([]string{
			"", "unknown album", "unknown album"}).WithArtistNames([]string{
			"", "unknown artist", "unknown artist"}).WithTrackNames([]string{
			"", "unknown title", "unknown title"}).WithGenres([]string{
			"", "unknown", "unknown"}).WithYears(
			[]string{"", "1900", "1900"}).WithTrackNumbers(
			[]int{0, 1, 1}).WithPrimarySource(files.ID3V2),
	}
	deletedTrack := &files.Track{
		FilePath:   filepath.Join(testDir, "no such file"),
		SimpleName: strings.TrimSuffix(trackName, ".mp3"),
		Number:     2,
		Album: &files.Album{
			Title:             "fine album",
			CanonicalGenre:    "Classic Rock",
			CanonicalYear:     "2022",
			CanonicalTitle:    "fine album",
			MusicCDIdentifier: id3v2.UnknownFrame{Body: []byte("fine album")},
			RecordingArtist: &files.Artist{
				Name:          "fine artist",
				CanonicalName: "fine artist",
			},
		},
		Metadata: files.NewTrackMetadataV1().WithAlbumNames([]string{
			"", "unknown album", "unknown album"}).WithArtistNames([]string{
			"", "unknown artist", "unknown artist"}).WithTrackNames([]string{
			"", "unknown title", "unknown title"}).WithGenres([]string{
			"", "unknown", "unknown"}).WithYears(
			[]string{"", "1900", "1900"}).WithTrackNumbers(
			[]int{0, 1, 1}).WithPrimarySource(files.ID3V2),
	}
	editedTm := files.NewTrackMetadataV1().WithAlbumNames([]string{
		"", "fine album", "fine album"}).WithArtistNames([]string{
		"", "fine artist", "fine artist"}).WithTrackNames([]string{
		"", "edit this track", "edit this track"}).WithGenres([]string{
		"", "Classic Rock", "Classic Rock"}).WithYears([]string{
		"", "2022", "2022"}).WithTrackNumbers([]int{0, 2, 2}).WithMusicCDIdentifier(
		[]byte("fine album")).WithPrimarySource(files.ID3V2)
	tests := map[string]struct {
		t      *files.Track
		wantE  []string
		wantTm *files.TrackMetadataV1
	}{
		"error checking": {
			t: deletedTrack,
			wantE: []string{
				"open updateMetadata\\no such file: The system cannot find the file specified.",
				"open updateMetadata\\no such file: The system cannot find the file specified.",
			},
		},
		"no edit required": {
			t:     &files.Track{Metadata: nil},
			wantE: []string{files.ErrNoEditNeeded.Error()},
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
				t.Errorf("%s = %v, want %v", fnName, eStrings, tt.wantE)
			} else if len(gotE) == 0 && tt.t.Metadata != nil {
				// verify file was correctly rewritten
				gotTm := files.ReadRawMetadata(tt.t.FilePath)
				if !reflect.DeepEqual(gotTm, tt.wantTm) {
					t.Errorf("%s read %#v, want %#v", fnName, gotTm, tt.wantTm)
				}
			}
		})
	}
}

func Test_processArtistMetadata(t *testing.T) {
	const fnName = "processArtistMetadata()"
	artist1 := files.NewArtist("artist_name", "")
	album1 := files.AlbumMaker{Title: "album1", Artist: artist1}.NewAlbum()
	artist1.AddAlbum(album1)
	for k := 1; k <= 10; k++ {
		src := files.ID3V2
		tM := files.NewTrackMetadataV1().WithPrimarySource(src)
		tM.SetArtistName(src, "artist:name")
		track := files.TrackMaker{
			Album:      album1,
			FileName:   fmt.Sprintf("%02d track%d.mp3", k, k),
			SimpleName: fmt.Sprintf("track%d", k),
			Number:     k,
		}.NewTrack()
		track.Metadata = tM
		album1.AddTrack(track)
	}
	artist2 := files.NewArtist("artist_name", "")
	album2 := files.AlbumMaker{Title: "album2", Artist: artist2}.NewAlbum()
	artist2.AddAlbum(album2)
	for k := 1; k <= 10; k++ {
		src := files.ID3V2
		tM := files.NewTrackMetadataV1().WithPrimarySource(src)
		tM.SetArtistName(src, "unknown artist")
		track := files.TrackMaker{
			Album:      album2,
			FileName:   fmt.Sprintf("%02d track%d.mp3", k, k),
			SimpleName: fmt.Sprintf("track%d", k),
			Number:     k,
		}.NewTrack()
		track.Metadata = tM
		album2.AddTrack(track)
	}
	artist3 := files.NewArtist("artist_name", "")
	album3 := files.AlbumMaker{Title: "album3", Artist: artist3}.NewAlbum()
	artist3.AddAlbum(album3)
	for k := 1; k <= 10; k++ {
		src := files.ID3V2
		tM := files.NewTrackMetadataV1().WithPrimarySource(src)
		if k%2 == 0 {
			tM.SetArtistName(src, "artist:name")
		} else {
			tM.SetArtistName(src, "artist_name")
		}
		track := files.TrackMaker{
			Album:      album3,
			FileName:   fmt.Sprintf("%02d track%d.mp3", k, k),
			SimpleName: fmt.Sprintf("track%d", k),
			Number:     k,
		}.NewTrack()
		track.Metadata = tM
		album3.AddTrack(track)
	}
	type args struct {
		artists []*files.Artist
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"unanimous choice": {args: args{artists: []*files.Artist{artist1}}},
		"unknown choice":   {args: args{artists: []*files.Artist{artist2}}},
		"ambiguous choice": {
			args: args{artists: []*files.Artist{artist3}},
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
			files.ProcessArtistMetadata(o, tt.args.artists)
			o.Report(t, fnName, tt.WantedRecording)
		})
	}
}

func Test_processAlbumMetadata(t *testing.T) {
	const fnName = "processAlbumMetadata()"
	// ordinary test data
	src := files.ID3V2
	var artists1 []*files.Artist
	artist1 := files.NewArtist("good artist", "")
	artists1 = append(artists1, artist1)
	album1 := files.AlbumMaker{Title: "good-album", Artist: artist1}.NewAlbum()
	artist1.AddAlbum(album1)
	tm := files.NewTrackMetadataV1().WithPrimarySource(src)
	tm.SetAlbumName(src, "good:album")
	tm.SetGenre(src, "pop")
	tm.SetYear(src, "2022")
	track1 := files.TrackMaker{
		Album:      album1,
		FileName:   "01 track1.mp3",
		SimpleName: "track1",
		Number:     1,
	}.NewTrack()
	track1.Metadata = tm
	album1.AddTrack(track1)
	// more interesting test data
	var artists2 []*files.Artist
	artist2 := files.NewArtist("another good artist", "")
	artists2 = append(artists2, artist2)
	album2 := files.AlbumMaker{
		Title:  "another good_album",
		Artist: artist2,
	}.NewAlbum()
	artist2.AddAlbum(album2)
	tm2a := files.NewTrackMetadataV1().WithPrimarySource(src)
	tm2a.SetAlbumName(src, "unknown album")
	tm2a.SetGenre(src, "unknown")
	tm2a.SetYear(src, "")
	track2a := files.TrackMaker{
		Album:      album2,
		FileName:   "01 track1.mp3",
		SimpleName: "track1",
		Number:     1,
	}.NewTrack()
	track2a.Metadata = tm2a
	album2.AddTrack(track2a)
	tm2b := files.NewTrackMetadataV1().WithPrimarySource(src)
	tm2b.SetAlbumName(src, "another good:album")
	tm2b.SetGenre(src, "pop")
	tm2b.SetYear(src, "2022")
	track2b := files.TrackMaker{
		Album:      album1,
		FileName:   "02 track2.mp3",
		SimpleName: "track2",
		Number:     2,
	}.NewTrack()
	track2b.Metadata = tm2b
	album2.AddTrack(track2b)
	tm2c := files.NewTrackMetadataV1().WithPrimarySource(src)
	tm2c.SetAlbumName(src, "another good:album")
	tm2c.SetGenre(src, "pop")
	tm2c.SetYear(src, "2022")
	track2c := files.TrackMaker{
		Album:      album1,
		FileName:   "03 track3.mp3",
		SimpleName: "track3",
		Number:     3,
	}.NewTrack()
	track2c.Metadata = tm2c
	album2.AddTrack(track2c)
	// error case data
	var artists3 []*files.Artist
	artist3 := files.NewArtist("problematic artist", "")
	artists3 = append(artists3, artist3)
	album3 := files.AlbumMaker{
		Title:  "problematic_album",
		Artist: artist3,
	}.NewAlbum()
	artist3.AddAlbum(album3)
	tm3a := files.NewTrackMetadataV1().WithPrimarySource(src).WithMusicCDIdentifier(
		[]byte{1, 2, 3})
	tm3a.SetAlbumName(src, "problematic:album")
	tm3a.SetGenre(src, "rock")
	tm3a.SetYear(src, "2023")
	track3a := files.TrackMaker{
		Album:      album2,
		FileName:   "01 track1.mp3",
		SimpleName: "track1",
		Number:     1,
	}.NewTrack()
	track3a.Metadata = tm3a
	album3.AddTrack(track3a)
	tm3b := files.NewTrackMetadataV1().WithPrimarySource(src).WithMusicCDIdentifier(
		[]byte{1, 2, 3, 4})
	tm3b.SetAlbumName(src, "problematic:Album")
	tm3b.SetGenre(src, "pop")
	tm3b.SetYear(src, "2022")
	track3b := files.TrackMaker{
		Album:      album1,
		FileName:   "02 track2.mp3",
		SimpleName: "track2",
		Number:     2,
	}.NewTrack()
	track3b.Metadata = tm3b
	album3.AddTrack(track3b)
	tm3c := files.NewTrackMetadataV1().WithPrimarySource(src).WithMusicCDIdentifier(
		[]byte{1, 2, 3, 4, 5})
	tm3c.SetAlbumName(src, "Problematic:album")
	tm3c.SetGenre(src, "folk")
	tm3c.SetYear(src, "2021")
	track3c := files.TrackMaker{
		Album:      album1,
		FileName:   "03 track3.mp3",
		SimpleName: "track3",
		Number:     3,
	}.NewTrack()
	track3c.Metadata = tm3c
	album3.AddTrack(track3c)
	// verify code can handle missing metadata
	track4 := files.TrackMaker{
		Album:      album1,
		FileName:   "04 track4.mp3",
		SimpleName: "track4",
		Number:     4,
	}.NewTrack()
	album3.AddTrack(track4)
	type args struct {
		artists []*files.Artist
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"ordinary test":    {args: args{artists: artists1}},
		"typical use case": {args: args{artists: artists2}},
		"errors": {
			args: args{artists: artists3},
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
			files.ProcessAlbumMetadata(o, tt.args.artists)
			o.Report(t, fnName, tt.WantedRecording)
		})
	}
}

func TestTrack_reportMetadataErrors(t *testing.T) {
	const fnName = "Track.reportMetadataErrors()"
	type args struct {
		t *files.Track
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"error handling": {
			args: args{
				t: &files.Track{
					SimpleName: "silly track",
					FilePath:   "Music\\silly artist\\silly album\\01 silly track.mp3",
					Metadata: files.NewTrackMetadataV1().WithErrorCauses(
						[]string{"", "id3v1 error!", "id3v2 error!"}),
					Album: &files.Album{
						Title:           "silly album",
						RecordingArtist: &files.Artist{Name: "silly artist"},
					},
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
			tt.args.t.ReportMetadataErrors(o)
			o.Report(t, fnName, tt.WantedRecording)
		})
	}
}

func TestTrack_Details(t *testing.T) {
	originalFileSystem := cmd_toolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmd_toolkit.AssignFileSystem(originalFileSystem)
	}()
	const fnName = "Track.Details()"
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
	createFileWithContent(".", goodFileName, content)
	tests := map[string]struct {
		t       *files.Track
		want    map[string]string
		wantErr bool
	}{
		"error case": {
			t:       &files.Track{FilePath: "./no such file"},
			wantErr: true,
		},
		"good case": {
			t: &files.Track{FilePath: filepath.Join(".", goodFileName)},
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
				t.Errorf("%s error = %v, wantErr %v", fnName, gotErr, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
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
func (sb *sampleBus) WriteCanonicalConsole(_ string, _ ...any)       {}
func (sb *sampleBus) WriteConsole(_ string, _ ...any)                {}
func (sb *sampleBus) WriteCanonicalError(_ string, _ ...any)         {}
func (sb *sampleBus) WriteError(_ string, _ ...any)                  {}
func (sb *sampleBus) ConsoleWriter() io.Writer {
	return sb.consoleWriter
}
func (sb *sampleBus) ErrorWriter() io.Writer {
	return sb.errorWriter
}
func (sb *sampleBus) IsConsoleTTY() bool {
	return sb.consoleTTY
}
func (sb *sampleBus) IsErrorTTY() bool {
	return sb.errorTTY
}
func (sb *sampleBus) IncrementTab(_ uint8) {}
func (sb *sampleBus) DecrementTab(_ uint8) {}
func (sb *sampleBus) Tab() uint8 {
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
			if got := files.ProgressWriter(tt.o); got != tt.want {
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
	fs := cmd_toolkit.FileSystem()
	_, fileErr = fs.Stat(fileName)
	if fileErr == nil {
		fileErr = fmt.Errorf("file %q already exists", fileName)
	} else if errors.Is(fileErr, afero.ErrFileNotFound) {
		fileErr = afero.WriteFile(fs, fileName, content, cmd_toolkit.StdFilePermissions)
	}
	return
}

// createFile creates a file in a specified directory with standardized content
func createFile(dir, name string) (err error) {
	return createFileWithContent(dir, name, []byte("file contents for "+name))
}
