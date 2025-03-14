package files

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
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
					title:           "some album",
					recordingArtist: NewArtist("some artist", `music\some artist`),
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
					title:           "some album",
					recordingArtist: NewArtist("some artist", `music\some artist`),
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
					title:           "some album",
					recordingArtist: NewArtist("some artist", `music\some artist`),
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
					title:           "some album",
					recordingArtist: NewArtist("some artist", `music\some artist`),
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
					title:           "some album",
					recordingArtist: NewArtist("some artist", `music\some artist`),
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

func TestTrackString(t *testing.T) {
	tests := map[string]struct {
		t    *Track
		want string
	}{
		"expected": {
			t:    &Track{filePath: "my path"},
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
			t:    &Track{album: &Album{title: "my album name"}},
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
				album: &Album{recordingArtist: NewArtist("my artist", `music\my artist`)},
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
			t:    &Track{filePath: "Music/my artist/my album/03 track.mp3"},
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
			t:    &Track{filePath: "Music/my artist/my album/03 track.mp3"},
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
			gotS, gotOk := canonicalChoice(tt.m)
			if gotS != tt.wantS {
				t.Errorf("canonicalChoice gotS = %v, want %v", gotS, tt.wantS)
			}
			if gotOk != tt.wantOk {
				t.Errorf("canonicalChoice gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestTrack_ID3V2Diagnostics(t *testing.T) {
	originalFileSystem := cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmdtoolkit.AssignFileSystem(originalFileSystem)
	}()
	content := createID3v2TaggedData(cannedPayload, map[string]string{
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
	})
	goodFileName := "goodFile.mp3"
	_ = createFileWithContent(".", goodFileName, content)
	tests := map[string]struct {
		t                *Track
		wantEncoding     string
		wantVersion      byte
		wantFrameStrings map[string][]string
		wantErr          bool
	}{
		"error case": {
			t:       &Track{filePath: "./no such file"},
			wantErr: true,
		},
		"good case": {
			t:            &Track{filePath: filepath.Join(".", goodFileName)},
			wantEncoding: "ISO-8859-1",
			wantVersion:  3,
			wantFrameStrings: map[string][]string{
				"Fake": {"00 68 75 68                                     •huh"},
				"T???": {"who knows?"},
				"TALB": {"unknown album"},
				"TCOM": {"a couple of idiots"},
				"TCON": {"dance music"},
				"TIT2": {"unknown track"},
				"TLEN": {"0:01.000"},
				"TPE1": {"unknown artist"},
				"TRCK": {"2"},
				"TYER": {"2022"},
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
				if gotInfo.Encoding() != tt.wantEncoding {
					t.Errorf("Track.ID3V2Diagnostics() gotInfo.encoding = %q, want %q",
						gotInfo.Encoding(), tt.wantEncoding)
				}
				if gotInfo.Version() != tt.wantVersion {
					t.Errorf("Track.ID3V2Diagnostics() gotInfo.version = %d, want %d",
						gotInfo.Version(), tt.wantVersion)
				}
				if !reflect.DeepEqual(gotInfo.Frames(), tt.wantFrameStrings) {
					t.Errorf("Track.ID3V2Diagnostics() gotInfo.frames = %v, want %v",
						gotInfo.Frames(), tt.wantFrameStrings)
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
			t:       &Track{filePath: filepath.Join(testDir, smallFile)},
			wantErr: true,
		},
		"invalid file": {
			t:       &Track{filePath: filepath.Join(testDir, invalidFile)},
			wantErr: true,
		},
		"good file": {
			t: &Track{filePath: filepath.Join(testDir, goodFile)},
			want: []string{
				"Artist: The Beatles",
				"Album: On Air: Live At The BBC, Volum",
				"Title: Ringo - Pop Profile [Interview",
				"Track: 29",
				"Year: 2013",
				"Genre: other",
				"Comment: silly",
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
	return v1.data
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
	postReadTm := newTrackMetadata()
	for _, src := range sourceTypes {
		postReadTm.setArtistName(src, artistName)
		postReadTm.setAlbumName(src, albumName)
		postReadTm.setAlbumGenre(src, genre)
		postReadTm.setAlbumYear(src, year)
		postReadTm.setTrackName(src, trackName)
		postReadTm.setTrackNumber(src, track)
	}
	postReadTm.setCDIdentifier([]byte{0})
	postReadTm.setCanonicalSource(ID3V2)
	tests := map[string]struct {
		t    *Track
		want *TrackMetadata
	}{
		"no read needed": {
			t:    &Track{metadata: newTrackMetadata()},
			want: newTrackMetadata()},
		"read file": {
			t:    &Track{filePath: filepath.Join(testDir, fileName)},
			want: postReadTm,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			bar := pb.New(1)
			bar.SetWriter(output.NewNilBus().ErrorWriter())
			bar.Start()
			tt.t.loadMetadata(bar)
			waitForFilesClosed()
			bar.Finish()
			if !reflect.DeepEqual(tt.t.metadata, tt.want) {
				t.Errorf("Track.loadMetadata() got %#v want %#v", tt.t.metadata, tt.want)
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
				Title:     albumName,
				Artist:    artist,
				Directory: albumName,
			}.NewAlbum(true)
			for n := 0; n < 50; n++ {
				trackName := fmt.Sprintf("track %d-%d-%d", k, m, n)
				trackFileName := fmt.Sprintf("%02d %s.mp3", n+1, trackName)
				track := &Track{
					filePath:   filepath.Join(albumPath, trackFileName),
					simpleName: trackName,
					album:      album,
					number:     n + 1,
					metadata:   nil,
				}
				metadata := map[string]any{
					"artist": artistName,
					"album":  albumName,
					"title":  trackName,
					"genre":  "Classic Rock",
					"year":   "2022",
					"track":  n + 1,
				}
				content := createConsistentlyTaggedData(
					[]byte{0, 1, 2, 3, 4, 5, 6, byte(k), byte(m), byte(n)},
					metadata,
				)
				_ = createFileWithContent(albumPath, trackFileName, content)
				album.addTrack(track)
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
				for _, album := range artist.Albums() {
					for _, track := range album.tracks {
						if track.needsMetadata() {
							t.Errorf("ReadMetadata() track %q has no metadata",
								track.filePath)
						} else if track.hasMetadataError() {
							t.Errorf("ReadMetadata() track %q is defective: %v",
								track.filePath, track.metadata.errorCauses())
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
		title:           "problematic:album",
		recordingArtist: problematicArtist,
		genre:           "hard rock",
		year:            "1999",
		canonicalTitle:  "problematic:album",
	}
	src := ID3V2
	metadata := newTrackMetadata()
	metadata.setCanonicalSource(src)
	metadata.setCDIdentifier([]byte{1, 3, 5})
	metadata.setArtistName(src, "unknown artist")
	metadata.setAlbumName(src, "unknown album")
	metadata.setAlbumGenre(src, "unknown")
	metadata.setAlbumYear(src, "2001")
	metadata.setTrackName(src, "unknown title")
	metadata.setTrackNumber(src, 2)
	problematicTrack := TrackMaker{
		Album:      problematicAlbum,
		FileName:   "03 bad track.mp3",
		SimpleName: "bad track",
		Number:     3,
	}.NewTrack(false)
	problematicTrack.metadata = metadata
	problematicAlbum.addTrack(problematicTrack)
	problematicArtist.addAlbum(problematicAlbum)
	goodArtist := NewArtist("good artist", "")
	goodAlbum := &Album{
		title:           "good album",
		recordingArtist: goodArtist,
		genre:           "Classic Rock",
		year:            "1999",
		canonicalTitle:  "good album",
	}
	src2 := ID3V1
	metadata2 := newTrackMetadata()
	metadata2.setCanonicalSource(src2)
	metadata2.setArtistName(src2, "good artist")
	metadata2.setAlbumName(src2, "good album")
	metadata2.setAlbumGenre(src2, "Classic Rock")
	metadata2.setAlbumYear(src2, "1999")
	metadata2.setTrackName(src2, "good track")
	metadata2.setTrackNumber(src2, 3)
	metadata2.setErrorCause(ID3V2, "no id3v2 metadata, how odd")
	goodTrack := TrackMaker{
		Album:      goodAlbum,
		FileName:   "03 good track.mp3",
		SimpleName: "good track",
		Number:     3,
	}.NewTrack(false)
	goodTrack.metadata = metadata2
	goodAlbum.addTrack(goodTrack)
	goodArtist.addAlbum(goodAlbum)
	errorMetadata := newTrackMetadata()
	errorMetadata.setErrorCause(ID3V1, "oops")
	errorMetadata.setErrorCause(ID3V2, "oops")
	noMetadata := newTrackMetadata()
	noMetadata.setErrorCause(ID3V1, errNoID3V1MetadataFound.Error())
	noMetadata.setErrorCause(ID3V2, errNoID3V2MetadataFound.Error())
	tests := map[string]struct {
		t    *Track
		want []string
	}{
		"unread metadata": {
			t:    &Track{metadata: nil},
			want: []string{"differences cannot be determined: metadata has not been read"},
		},
		"track with error": {
			t:    &Track{metadata: errorMetadata},
			want: []string{"differences cannot be determined: track metadata may be corrupted"},
		},
		"track with no metadata": {
			t:    &Track{metadata: noMetadata},
			want: []string{"differences cannot be determined: the track file contains no metadata"},
		},
		"track with metadata differences": {
			t: problematicTrack,
			want: []string{
				"ID3V1 metadata [0] does not agree with track number 3",
				"ID3V1 metadata [] does not agree with album genre \"hard rock\"",
				"ID3V1 metadata [] does not agree with album name \"problematic:album\"",
				"ID3V1 metadata [] does not agree with album year \"1999\"",
				"ID3V1 metadata [] does not agree with artist name \"problematic:artist\"",
				"ID3V1 metadata [] does not agree with track name \"bad track\"",
				"ID3V2 metadata [2001] does not agree with album year \"1999\"",
				"ID3V2 metadata [2] does not agree with track number 3",
				"ID3V2 metadata [[1 3 5]] does not agree with the MCDI frame \"\"",
				"ID3V2 metadata [unknown album] does not agree with album name \"problematic:album\"",
				"ID3V2 metadata [unknown artist] does not agree with artist name \"problematic:artist\"",
				"ID3V2 metadata [unknown title] does not agree with track name \"bad track\"",
				"ID3V2 metadata [unknown] does not agree with album genre \"hard rock\"",
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
	expectedMetadata := newTrackMetadata()
	for _, src := range sourceTypes {
		expectedMetadata.setArtistName(src, "unknown artist")
		expectedMetadata.setAlbumName(src, "unknown album")
		expectedMetadata.setAlbumGenre(src, "unknown")
		expectedMetadata.setAlbumYear(src, "1900")
		expectedMetadata.setTrackName(src, "unknown title")
		expectedMetadata.setTrackNumber(src, 1)
	}
	expectedMetadata.setCanonicalSource(ID3V2)
	track := &Track{
		filePath:   filepath.Join(testDir, trackName),
		simpleName: strings.TrimSuffix(trackName, ".mp3"),
		number:     2,
		album: &Album{
			title:          "fine album",
			genre:          "classic rock",
			year:           "2022",
			canonicalTitle: "fine album",
			cdIdentifier: id3v2.UnknownFrame{
				Body: []byte("fine album"),
			},
			recordingArtist: NewArtist("fine artist", `music\fine artist`),
		},
		metadata: expectedMetadata,
	}
	deletedTrack := &Track{
		filePath:   filepath.Join(testDir, "no such file"),
		simpleName: strings.TrimSuffix(trackName, ".mp3"),
		number:     2,
		album: &Album{
			title:           "fine album",
			genre:           "classic rock",
			year:            "2022",
			canonicalTitle:  "fine album",
			cdIdentifier:    id3v2.UnknownFrame{Body: []byte("fine album")},
			recordingArtist: NewArtist("fine artist", `music\fine artist`),
		},
		metadata: expectedMetadata,
	}
	editedTm := newTrackMetadata()
	for _, src := range sourceTypes {
		editedTm.setArtistName(src, "fine artist")
		editedTm.setAlbumName(src, "fine album")
		editedTm.setAlbumGenre(src, "classic rock")
		editedTm.setAlbumYear(src, "2022")
		editedTm.setTrackName(src, "edit this track")
		editedTm.setTrackNumber(src, 2)
	}
	editedTm.setCDIdentifier([]byte("fine album"))
	editedTm.setCanonicalSource(ID3V2)
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
			t:     &Track{metadata: nil},
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
			} else if len(gotE) == 0 && tt.t.metadata != nil {
				// verify file was correctly rewritten
				gotTm := initializeMetadata(tt.t.filePath)
				if !reflect.DeepEqual(gotTm, tt.wantTm) {
					t.Errorf("Track.UpdateMetadata() read %#v, want %#v", gotTm, tt.wantTm)
				}
			}
		})
	}
}

func TestProcessArtistMetadata(t *testing.T) {
	artist1 := NewArtist("artist_name", "")
	album1 := AlbumMaker{Title: "album1", Artist: artist1}.NewAlbum(true)
	for k := 1; k <= 10; k++ {
		src := ID3V2
		tm := newTrackMetadata()
		tm.setCanonicalSource(src)
		tm.setArtistName(src, "artist:name")
		track := TrackMaker{
			Album:      album1,
			FileName:   fmt.Sprintf("%02d track%d.mp3", k, k),
			SimpleName: fmt.Sprintf("track%d", k),
			Number:     k,
		}.NewTrack(false)
		track.metadata = tm
		album1.addTrack(track)
	}
	artist2 := NewArtist("artist_name", "")
	album2 := AlbumMaker{Title: "album2", Artist: artist2}.NewAlbum(true)
	for k := 1; k <= 10; k++ {
		src := ID3V2
		tm := newTrackMetadata()
		tm.setCanonicalSource(src)
		tm.setArtistName(src, "unknown artist")
		track := TrackMaker{
			Album:      album2,
			FileName:   fmt.Sprintf("%02d track%d.mp3", k, k),
			SimpleName: fmt.Sprintf("track%d", k),
			Number:     k,
		}.NewTrack(false)
		track.metadata = tm
		album2.addTrack(track)
	}
	artist3 := NewArtist("artist_name", "")
	album3 := AlbumMaker{Title: "album3", Artist: artist3}.NewAlbum(true)
	for k := 1; k <= 10; k++ {
		src := ID3V2
		tm := newTrackMetadata()
		tm.setCanonicalSource(src)
		if k%2 == 0 {
			tm.setArtistName(src, "artist:name")
		} else {
			tm.setArtistName(src, "artist_name")
		}
		track := TrackMaker{
			Album:      album3,
			FileName:   fmt.Sprintf("%02d track%d.mp3", k, k),
			SimpleName: fmt.Sprintf("track%d", k),
			Number:     k,
		}.NewTrack(false)
		track.metadata = tm
		album3.addTrack(track)
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
			processArtistMetadata(o, tt.artists)
			o.Report(t, "processArtistMetadata()", tt.WantedRecording)
		})
	}
}

func TestProcessAlbumMetadata(t *testing.T) {
	// ordinary test data
	src := ID3V2
	var artists1 []*Artist
	artist1 := NewArtist("good artist", "")
	artists1 = append(artists1, artist1)
	album1 := AlbumMaker{Title: "good-album", Artist: artist1}.NewAlbum(true)
	tm := newTrackMetadata()
	tm.setCanonicalSource(src)
	tm.setAlbumName(src, "good:album")
	tm.setAlbumGenre(src, "pop")
	tm.setAlbumYear(src, "2022")
	track1 := TrackMaker{
		Album:      album1,
		FileName:   "01 track1.mp3",
		SimpleName: "track1",
		Number:     1,
	}.NewTrack(false)
	track1.metadata = tm
	album1.addTrack(track1)
	// more interesting test data
	var artists2 []*Artist
	artist2 := NewArtist("another good artist", "")
	artists2 = append(artists2, artist2)
	album2 := AlbumMaker{
		Title:  "another good_album",
		Artist: artist2,
	}.NewAlbum(true)
	tm2a := newTrackMetadata()
	tm2a.setCanonicalSource(src)
	tm2a.setAlbumName(src, "unknown album")
	tm2a.setAlbumGenre(src, "unknown")
	tm2a.setAlbumYear(src, "")
	track2a := TrackMaker{
		Album:      album2,
		FileName:   "01 track1.mp3",
		SimpleName: "track1",
		Number:     1,
	}.NewTrack(false)
	track2a.metadata = tm2a
	album2.addTrack(track2a)
	tm2b := newTrackMetadata()
	tm2b.setCanonicalSource(src)
	tm2b.setAlbumName(src, "another good:album")
	tm2b.setAlbumGenre(src, "pop")
	tm2b.setAlbumYear(src, "2022")
	track2b := TrackMaker{
		Album:      album1,
		FileName:   "02 track2.mp3",
		SimpleName: "track2",
		Number:     2,
	}.NewTrack(false)
	track2b.metadata = tm2b
	album2.addTrack(track2b)
	tm2c := newTrackMetadata()
	tm2c.setCanonicalSource(src)
	tm2c.setAlbumName(src, "another good:album")
	tm2c.setAlbumGenre(src, "pop")
	tm2c.setAlbumYear(src, "2022")
	track2c := TrackMaker{
		Album:      album1,
		FileName:   "03 track3.mp3",
		SimpleName: "track3",
		Number:     3,
	}.NewTrack(false)
	track2c.metadata = tm2c
	album2.addTrack(track2c)
	// error case data
	var artists3 []*Artist
	artist3 := NewArtist("problematic artist", "")
	artists3 = append(artists3, artist3)
	album3 := AlbumMaker{
		Title:  "problematic_album",
		Artist: artist3,
	}.NewAlbum(true)
	tm3a := newTrackMetadata()
	tm3a.setCanonicalSource(src)
	tm3a.setCDIdentifier([]byte{1, 2, 3})
	tm3a.setAlbumName(src, "problematic:album")
	tm3a.setAlbumGenre(src, "rock")
	tm3a.setAlbumYear(src, "2023")
	track3a := TrackMaker{
		Album:      album2,
		FileName:   "01 track1.mp3",
		SimpleName: "track1",
		Number:     1,
	}.NewTrack(false)
	track3a.metadata = tm3a
	album3.addTrack(track3a)
	tm3b := newTrackMetadata()
	tm3b.setCanonicalSource(src)
	tm3b.setCDIdentifier([]byte{1, 2, 3, 4})
	tm3b.setAlbumName(src, "problematic:Album")
	tm3b.setAlbumGenre(src, "pop")
	tm3b.setAlbumYear(src, "2022")
	track3b := TrackMaker{
		Album:      album1,
		FileName:   "02 track2.mp3",
		SimpleName: "track2",
		Number:     2,
	}.NewTrack(false)
	track3b.metadata = tm3b
	album3.addTrack(track3b)
	tm3c := newTrackMetadata()
	tm3c.setCanonicalSource(src)
	tm3c.setCDIdentifier([]byte{1, 2, 3, 4, 5})
	tm3c.setAlbumName(src, "Problematic:album")
	tm3c.setAlbumGenre(src, "folk")
	tm3c.setAlbumYear(src, "2021")
	track3c := TrackMaker{
		Album:      album1,
		FileName:   "03 track3.mp3",
		SimpleName: "track3",
		Number:     3,
	}.NewTrack(false)
	track3c.metadata = tm3c
	album3.addTrack(track3c)
	// verify code can handle missing metadata
	track4 := TrackMaker{
		Album:      album1,
		FileName:   "04 track4.mp3",
		SimpleName: "track4",
		Number:     4,
	}.NewTrack(false)
	album3.addTrack(track4)
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
			processAlbumMetadata(o, tt.artists)
			o.Report(t, "processAlbumMetadata()", tt.WantedRecording)
		})
	}
}

func TestTrack_ReportMetadataErrors(t *testing.T) {
	tm := newTrackMetadata()
	tm.setErrorCause(ID3V1, "id3v1 error!")
	tm.setErrorCause(ID3V2, "id3v2 error!")
	tests := map[string]struct {
		t *Track
		output.WantedRecording
	}{
		"error handling": {
			t: &Track{
				simpleName: "silly track",
				filePath:   "Music\\silly artist\\silly album\\01 silly track.mp3",
				metadata:   tm,
				album: &Album{
					title:           "silly album",
					recordingArtist: NewArtist("silly artist", `music\silly artist`),
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
			tt.t.reportMetadataErrors(o)
			o.Report(t, "Track.reportMetadataErrors()", tt.WantedRecording)
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
func (sB *sampleBus) BeginConsoleList(_ bool)                     {}
func (sB *sampleBus) EndConsoleList()                             {}
func (sB *sampleBus) BeginErrorList(_ bool)                       {}
func (sB *sampleBus) EndErrorList()                               {}
func (sB *sampleBus) ConsoleListDecorator() *output.ListDecorator { return nil }
func (sB *sampleBus) ErrorListDecorator() *output.ListDecorator   { return nil }
func (sB *sampleBus) ConsolePrintf(_ string, _ ...any)            {}
func (sB *sampleBus) ErrorPrintf(_ string, _ ...any)              {}
func (sB *sampleBus) ConsolePrintln(_ string)                     {}
func (sB *sampleBus) ErrorPrintln(_ string)                       {}

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
			if got := progressWriter(tt.o); got != tt.want {
				t.Errorf("progressWriter() = %v, want %v", got, tt.want)
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

func TestSortTracks(t *testing.T) {
	tests := map[string]struct {
		tracks []*Track
		want   []*Track
	}{
		"thorough": {
			tracks: []*Track{
				{
					simpleName: "b",
					album: &Album{
						title:           "b",
						recordingArtist: NewArtist("c", `music\c`),
					},
				},
				{
					simpleName: "b",
					album: &Album{
						title:           "a",
						recordingArtist: NewArtist("c", `music\c`),
					},
				},
				{
					simpleName: "b",
					album: &Album{
						title:           "b",
						recordingArtist: NewArtist("a", `music\a`),
					},
				},
				{
					simpleName: "a",
					album: &Album{
						title:           "b",
						recordingArtist: NewArtist("c", `music\c`),
					},
				},
				{
					simpleName: "a",
					album: &Album{
						title:           "a",
						recordingArtist: NewArtist("c", `music\c`),
					},
				},
				{
					simpleName: "a",
					album: &Album{
						title:           "b",
						recordingArtist: NewArtist("a", `music\a`),
					},
				},
			},
			want: []*Track{
				{
					simpleName: "a",
					album: &Album{
						title:           "a",
						recordingArtist: NewArtist("c", `music\c`),
					},
				},
				{
					simpleName: "a",
					album: &Album{
						title:           "b",
						recordingArtist: NewArtist("a", `music\a`),
					},
				},
				{
					simpleName: "a",
					album: &Album{
						title:           "b",
						recordingArtist: NewArtist("c", `music\c`),
					},
				},
				{
					simpleName: "b",
					album: &Album{
						title:           "a",
						recordingArtist: NewArtist("c", `music\c`),
					},
				},
				{
					simpleName: "b",
					album: &Album{
						title:           "b",
						recordingArtist: NewArtist("a", `music\a`),
					},
				},
				{
					simpleName: "b",
					album: &Album{
						title:           "b",
						recordingArtist: NewArtist("c", `music\c`),
					},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			SortTracks(tt.tracks)
			if got := tt.tracks; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SortTracks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_Path(t1 *testing.T) {
	type fields struct {
		album      *Album
		FilePath   string
		metadata   *TrackMetadata
		SimpleName string
		Number     int
	}
	tests := map[string]struct {
		fields
		want string
	}{
		"simple reader, simple test": {
			fields: fields{
				album:      nil,
				FilePath:   "/c/music/my artist/my album/01 this track.mp3",
				metadata:   nil,
				SimpleName: "this track",
				Number:     1,
			},
			want: "/c/music/my artist/my album/01 this track.mp3",
		},
	}
	for name, tt := range tests {
		t1.Run(name, func(t1 *testing.T) {
			t := &Track{
				album:      tt.fields.album,
				filePath:   tt.fields.FilePath,
				metadata:   tt.fields.metadata,
				simpleName: tt.fields.SimpleName,
				number:     tt.fields.Number,
			}
			if got := t.Path(); got != tt.want {
				t1.Errorf("Path() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_Name(t1 *testing.T) {
	type fields struct {
		album      *Album
		filePath   string
		metadata   *TrackMetadata
		SimpleName string
		Number     int
	}
	tests := map[string]struct {
		fields
		want string
	}{
		"simple reader, simple test": {
			fields: fields{
				album:      nil,
				filePath:   "/c/music/my artist/my album/01 this track.mp3",
				metadata:   nil,
				SimpleName: "this track",
				Number:     1,
			},
			want: "this track",
		},
	}
	for name, tt := range tests {
		t1.Run(name, func(t1 *testing.T) {
			t := &Track{
				album:      tt.fields.album,
				filePath:   tt.fields.filePath,
				metadata:   tt.fields.metadata,
				simpleName: tt.fields.SimpleName,
				number:     tt.fields.Number,
			}
			if got := t.Name(); got != tt.want {
				t1.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_Number(t1 *testing.T) {
	type fields struct {
		album      *Album
		filePath   string
		metadata   *TrackMetadata
		simpleName string
		number     int
	}
	tests := map[string]struct {
		fields
		want int
	}{
		"simple reader, simple test": {
			fields: fields{
				album:      nil,
				filePath:   "/c/music/my artist/my album/01 this track.mp3",
				metadata:   nil,
				simpleName: "this track",
				number:     1,
			},
			want: 1,
		},
	}
	for name, tt := range tests {
		t1.Run(name, func(t1 *testing.T) {
			t := &Track{
				album:      tt.fields.album,
				filePath:   tt.fields.filePath,
				metadata:   tt.fields.metadata,
				simpleName: tt.fields.simpleName,
				number:     tt.fields.number,
			}
			if got := t.Number(); got != tt.want {
				t1.Errorf("Number() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrackMaker_NewTrack(t *testing.T) {
	testAlbum1 := AlbumMaker{
		Title:     "my album",
		Artist:    nil,
		Directory: "my artist/my album",
	}.NewAlbum(false)
	testAlbum2 := AlbumMaker{
		Title:     "my other album",
		Artist:    nil,
		Directory: "my artist/my other album",
	}.NewAlbum(false)
	type fields struct {
		Album      *Album
		FileName   string
		SimpleName string
		Number     int
		Metadata   *TrackMetadata
	}
	tests := map[string]struct {
		fields
		addToAlbum          bool
		want                *Track
		wantAlbumTrackCount int
	}{
		"typical": {
			fields: fields{
				Album:      testAlbum1,
				FileName:   "01 typical track.mp3",
				SimpleName: "typical track",
				Number:     1,
				Metadata:   nil,
			},
			addToAlbum: true,
			want: &Track{
				album:      testAlbum1,
				filePath:   `my artist\my album\01 typical track.mp3`,
				metadata:   nil,
				simpleName: "typical track",
				number:     1,
			},
			wantAlbumTrackCount: 1,
		},
		"atypical": {
			fields: fields{
				Album:      testAlbum2,
				FileName:   "01 typical track.mp3",
				SimpleName: "typical track",
				Number:     1,
				Metadata:   nil,
			},
			addToAlbum: false,
			want: &Track{
				album:      testAlbum2,
				filePath:   `my artist\my other album\01 typical track.mp3`,
				metadata:   nil,
				simpleName: "typical track",
				number:     1,
			},
			wantAlbumTrackCount: 0,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ti := TrackMaker{
				Album:      tt.fields.Album,
				FileName:   tt.fields.FileName,
				SimpleName: tt.fields.SimpleName,
				Number:     tt.fields.Number,
				Metadata:   tt.fields.Metadata,
			}
			if got := ti.NewTrack(tt.addToAlbum); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTrack() = %v, want %v", got, tt.want)
			}
			if got := len(ti.Album.tracks); got != tt.wantAlbumTrackCount {
				t.Errorf("NewTrack() = %v, want %v", got, tt.wantAlbumTrackCount)
			}
		})
	}
}

func TestTrack_Copy(t1 *testing.T) {
	testAlbum1 := AlbumMaker{
		Title:     "my album",
		Artist:    nil,
		Directory: "my artist/my album",
	}.NewAlbum(false)
	testAlbum2 := AlbumMaker{
		Title:     "my other album",
		Artist:    nil,
		Directory: "my artist/my other album",
	}.NewAlbum(false)
	type fields struct {
		album      *Album
		filePath   string
		metadata   *TrackMetadata
		simpleName string
		number     int
	}
	type args struct {
		a          *Album
		addToAlbum bool
	}
	tests := map[string]struct {
		fields
		args
		want                *Track
		wantAlbumTrackCount int
	}{
		"typical": {
			fields: fields{
				album:      nil,
				filePath:   `my artist\my album\01 typical track.mp3`,
				simpleName: "typical track",
				number:     1,
				metadata:   nil,
			},
			args: args{
				a:          testAlbum1,
				addToAlbum: true,
			},
			want: &Track{
				album:      testAlbum1,
				filePath:   `my artist\my album\01 typical track.mp3`,
				metadata:   nil,
				simpleName: "typical track",
				number:     1,
			},
			wantAlbumTrackCount: 1,
		},
		"atypical": {
			fields: fields{
				album:      testAlbum2,
				filePath:   `my artist\my other album\01 typical track.mp3`,
				simpleName: "typical track",
				number:     1,
				metadata:   nil,
			},
			args: args{a: testAlbum2, addToAlbum: false},
			want: &Track{
				album:      testAlbum2,
				filePath:   `my artist\my other album\01 typical track.mp3`,
				metadata:   nil,
				simpleName: "typical track",
				number:     1,
			},
			wantAlbumTrackCount: 0,
		},
	}
	for name, tt := range tests {
		t1.Run(name, func(t1 *testing.T) {
			t := &Track{
				album:      tt.fields.album,
				filePath:   tt.fields.filePath,
				metadata:   tt.fields.metadata,
				simpleName: tt.fields.simpleName,
				number:     tt.fields.number,
			}
			if got := t.Copy(tt.args.a, tt.args.addToAlbum); !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("Copy() = %v, want %v", got, tt.want)
			}
			if got := len(tt.args.a.tracks); got != tt.wantAlbumTrackCount {
				t1.Errorf("Copy() = %v, want %v", got, tt.wantAlbumTrackCount)
			}
		})
	}
}

func TestFrameDescription(t *testing.T) {
	tests := map[string]struct {
		name string
		want string
	}{
		"known case": {
			name: "MCDI",
			want: "Music CD identifier",
		},
		"unknown case": {
			name: "Music",
			want: "No description found",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := FrameDescription(tt.name); got != tt.want {
				t.Errorf("FrameDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}
