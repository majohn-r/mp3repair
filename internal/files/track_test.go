package files

import (
	"bytes"
	"fmt"
	"mp3/internal"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/bogem/id3v2/v2"
)

func Test_parseTrackName(t *testing.T) {
	fnName := "parseTrackName()"
	type args struct {
		name   string
		album  string
		artist string
		ext    string
	}
	tests := []struct {
		name            string
		args            args
		wantSimpleName  string
		wantTrackNumber int
		wantValid       bool
	}{
		{
			name: "expected use case", wantSimpleName: "track name", wantTrackNumber: 59, wantValid: true,
			args: args{name: "59 track name.mp3", album: "some album", artist: "some artist", ext: ".mp3"},
		},
		{
			name: "expected use case with hyphen separator", wantSimpleName: "other track name", wantTrackNumber: 60, wantValid: true,
			args: args{name: "60-other track name.mp3", album: "some album", artist: "some artist", ext: ".mp3"},
		},
		{
			name: "wrong extension", wantSimpleName: "track name.mp4", wantTrackNumber: 59, wantValid: false,
			args: args{name: "59 track name.mp4", album: "some album", artist: "some artist", ext: ".mp3"},
		},
		{
			name: "missing track number", wantSimpleName: "name", wantTrackNumber: 0, wantValid: false,
			args: args{name: "track name.mp3", album: "some album", artist: "some artist", ext: ".mp3"},
		},
		{
			name: "missing track number, simple name", wantSimpleName: "", wantTrackNumber: 0, wantValid: false,
			args: args{name: "trackName.mp3", album: "some album", artist: "some artist", ext: ".mp3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSimpleName, gotTrackNumber, gotValid := ParseTrackName(tt.args.name, tt.args.album, tt.args.artist, tt.args.ext)
			if tt.wantValid {
				if gotSimpleName != tt.wantSimpleName {
					t.Errorf("%s gotSimpleName = %v, want %v", fnName, gotSimpleName, tt.wantSimpleName)
				}
				if gotTrackNumber != tt.wantTrackNumber {
					t.Errorf("%s gotTrackNumber = %v, want %v", fnName, gotTrackNumber, tt.wantTrackNumber)
				}
			}
			if gotValid != tt.wantValid {
				t.Errorf("%s gotValid = %v, want %v", fnName, gotValid, tt.wantValid)
			}
		})
	}
}

func TestTrack_needsTaggedData(t *testing.T) {
	tests := []struct {
		name string
		tr   *Track
		want bool
	}{
		{name: "needs tagged data", tr: &Track{TaggedTrack: trackUnknownTagsNotRead}, want: true},
		{name: "format error", tr: &Track{TaggedTrack: trackUnknownFormatError}, want: false},
		{name: "tag read error", tr: &Track{TaggedTrack: TrackUnknownTagReadError}, want: false},
		{name: "valid track number", tr: &Track{TaggedTrack: 1}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.needsTaggedData(); got != tt.want {
				t.Errorf("Track.needsTaggedData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_setTagReadError(t *testing.T) {
	tests := []struct {
		name string
		tr   *Track
	}{
		{name: "simple test", tr: &Track{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tr.setTagReadError()
			if tt.tr.TaggedTrack != TrackUnknownTagReadError {
				t.Errorf("Track.setTagReadError() failed to set TaggedTrack: %d", tt.tr.TaggedTrack)
			}
		})
	}
}

func TestTrack_setTagFormatError(t *testing.T) {
	tests := []struct {
		name string
		tr   *Track
	}{
		{name: "simple test", tr: &Track{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tr.setTagFormatError()
			if tt.tr.TaggedTrack != trackUnknownFormatError {
				t.Errorf("Track.setTagFormatError() failed to set TaggedTrack: %d", tt.tr.TaggedTrack)
			}
		})
	}
}

func Test_toTrackNumber(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		wantI   int
		wantErr bool
	}{
		{name: "good value", args: args{s: "12"}, wantI: 12, wantErr: false},
		{name: "empty value", args: args{s: ""}, wantI: 0, wantErr: true},
		{name: "negative value", args: args{s: "-12"}, wantI: 0, wantErr: true},
		{name: "invalid value", args: args{s: "foo"}, wantI: 0, wantErr: true},
		{name: "complicated value", args: args{s: "12/39"}, wantI: 12, wantErr: false},
		{name: "BOM-infested complicated value", args: args{s: "\ufeff12/39"}, wantI: 12, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotI, err := toTrackNumber(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("toTrackNumber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && gotI != tt.wantI {
				t.Errorf("toTrackNumber() = %v, want %v", gotI, tt.wantI)
			}
		})
	}
}

func TestTrack_setTags(t *testing.T) {
	type args struct {
		d *taggedTrackData
	}
	tests := []struct {
		name       string
		tr         *Track
		args       args
		wantAlbum  string
		wantArtist string
		wantTitle  string
		wantNumber int
	}{
		{
			name: "good input",
			tr:   &Track{},
			args: args{&taggedTrackData{
				album:  "my excellent album",
				artist: "great artist",
				title:  "best track ever",
				number: "1",
			}},
			wantAlbum:  "my excellent album",
			wantArtist: "great artist",
			wantTitle:  "best track ever",
			wantNumber: 1,
		},
		{
			name: "badly formatted input",
			tr:   &Track{},
			args: args{&taggedTrackData{
				album:  "my excellent album",
				artist: "great artist",
				title:  "best track ever",
				number: "foo",
			}},
			wantNumber: trackUnknownFormatError,
		},
		{
			name: "negative track",
			tr:   &Track{},
			args: args{&taggedTrackData{
				album:  "my excellent album",
				artist: "great artist",
				title:  "best track ever",
				number: "-1",
			}},
			wantNumber: trackUnknownFormatError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tr.setTags(tt.args.d)
			if tt.tr.TaggedTrack != tt.wantNumber {
				t.Errorf("track.SetTags() tagged track = %d, want %d ", tt.tr.TaggedTrack, tt.wantNumber)
			}
			if tt.wantNumber != trackUnknownFormatError {
				if tt.tr.TaggedAlbum != tt.wantAlbum {
					t.Errorf("track.SetTags() tagged album = %q, want %q", tt.tr.TaggedAlbum, tt.wantAlbum)
				}
				if tt.tr.TaggedArtist != tt.wantArtist {
					t.Errorf("track.SetTags() tagged artist = %q, want %q", tt.tr.TaggedArtist, tt.wantArtist)
				}
				if tt.tr.TaggedTitle != tt.wantTitle {
					t.Errorf("track.SetTags() tagged title = %q, want %q", tt.tr.TaggedTitle, tt.wantTitle)
				}
			}
		})
	}
}

func TestTrack_readTags(t *testing.T) {
	normalReader := func(path string) (*taggedTrackData, error) {
		return &taggedTrackData{
			album:  "beautiful album",
			artist: "great artist",
			title:  "terrific track",
			number: "1",
		}, nil
	}
	bentReader := func(path string) (*taggedTrackData, error) {
		return &taggedTrackData{
			album:  "beautiful album",
			artist: "great artist",
			title:  "terrific track",
			number: "-2",
		}, nil
	}
	brokenReader := func(path string) (*taggedTrackData, error) {
		return nil, fmt.Errorf("read error")
	}
	type args struct {
		reader func(string) (*taggedTrackData, error)
	}
	tests := []struct {
		name       string
		tr         *Track
		args       args
		wantAlbum  string
		wantArtist string
		wantTitle  string
		wantNumber int
	}{
		{
			name:       "normal",
			tr:         &Track{TaggedTrack: trackUnknownTagsNotRead},
			args:       args{normalReader},
			wantAlbum:  "beautiful album",
			wantArtist: "great artist",
			wantTitle:  "terrific track",
			wantNumber: 1,
		},
		{
			name: "replay",
			tr: &Track{
				TaggedTrack:  2,
				TaggedAlbum:  "nice album",
				TaggedArtist: "good artist",
				TaggedTitle:  "pretty song",
			},
			args:       args{normalReader},
			wantAlbum:  "nice album",
			wantArtist: "good artist",
			wantTitle:  "pretty song",
			wantNumber: 2,
		},
		{
			name:       "replay after read error",
			tr:         &Track{TaggedTrack: TrackUnknownTagReadError},
			args:       args{normalReader},
			wantNumber: TrackUnknownTagReadError,
		},
		{
			name:       "replay after format error",
			tr:         &Track{TaggedTrack: trackUnknownFormatError},
			args:       args{normalReader},
			wantNumber: trackUnknownFormatError,
		},
		{
			name:       "read error",
			tr:         &Track{TaggedTrack: trackUnknownTagsNotRead},
			args:       args{brokenReader},
			wantNumber: TrackUnknownTagReadError,
		},
		{
			name:       "format error",
			tr:         &Track{TaggedTrack: trackUnknownTagsNotRead},
			args:       args{bentReader},
			wantNumber: trackUnknownFormatError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tr.readTags(tt.args.reader)
			waitForSemaphoresDrained()
			if tt.tr.TaggedTrack != tt.wantNumber {
				t.Errorf("track.readTags() tagged track = %d, want %d ", tt.tr.TaggedTrack, tt.wantNumber)
			}
			if tt.wantNumber >= 0 {
				if tt.tr.TaggedAlbum != tt.wantAlbum {
					t.Errorf("track.readTags() tagged album = %q, want %q", tt.tr.TaggedAlbum, tt.wantAlbum)
				}
				if tt.tr.TaggedArtist != tt.wantArtist {
					t.Errorf("track.readTags() tagged artist = %q, want %q", tt.tr.TaggedArtist, tt.wantArtist)
				}
				if tt.tr.TaggedTitle != tt.wantTitle {
					t.Errorf("track.readTags() tagged title = %q, want %q", tt.tr.TaggedTitle, tt.wantTitle)
				}
			}
		})
	}
}

func Test_isComparable(t *testing.T) {
	type args struct {
		p nameTagPair
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "simple case", args: args{nameTagPair{name: "simple name", tag: "simple name"}}, want: true},
		{name: "case insensitive", args: args{nameTagPair{name: "SIMPLE name", tag: "simple NAME"}}, want: true},
		{name: "expected fail", args: args{nameTagPair{name: "simple name", tag: "artist: simple name"}}, want: false},
		{name: "illegal char case", args: args{nameTagPair{name: "simple_name", tag: "simple:name"}}, want: true},
		{name: "illegal final space", args: args{nameTagPair{name: "simple name", tag: "simple name "}}, want: true},
		{name: "illegal final period", args: args{nameTagPair{name: "simple name", tag: "simple name."}}, want: true},
		{name: "complex fail", args: args{nameTagPair{name: "simple_name", tag: "simple: nam"}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isComparable(tt.args.p); got != tt.want {
				t.Errorf("isComparable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_FindDifferences(t *testing.T) {
	tests := []struct {
		name string
		tr   *Track
		want []string
	}{
		{
			name: "typical use case",
			tr: &Track{
				TrackNumber:     1,
				Name:            "track name",
				ContainingAlbum: &Album{Name: "album name", RecordingArtist: &Artist{Name: "artist name"}},
				TaggedTrack:     1,
				TaggedTitle:     "track name",
				TaggedAlbum:     "album name",
				TaggedArtist:    "artist name",
			},
			want: nil,
		},
		{
			name: "another OK use case",
			tr: &Track{
				TrackNumber:     1,
				Name:            "track name",
				ContainingAlbum: &Album{Name: "album name", RecordingArtist: &Artist{Name: "artist name"}},
				TaggedTrack:     1,
				TaggedTitle:     "track:name",
				TaggedAlbum:     "album:name",
				TaggedArtist:    "artist:name",
			},
			want: nil,
		},
		{
			name: "oops",
			tr: &Track{
				TrackNumber:     2,
				Name:            "track:name",
				ContainingAlbum: &Album{Name: "album:name", RecordingArtist: &Artist{Name: "artist:name"}},
				TaggedTrack:     1,
				TaggedTitle:     "track name",
				TaggedAlbum:     "album name",
				TaggedArtist:    "artist name",
			},
			want: []string{
				"album \"album:name\" does not agree with album tag \"album name\"",
				"artist \"artist:name\" does not agree with artist tag \"artist name\"",
				"title \"track:name\" does not agree with title tag \"track name\"",
				"track number 2 does not agree with track tag 1",
			},
		},
		{name: "unread tags", tr: &Track{TaggedTrack: trackUnknownTagsNotRead}, want: []string{trackDiffUnreadTags}},
		{name: "unreadable tags", tr: &Track{TaggedTrack: TrackUnknownTagReadError}, want: []string{trackDiffUnreadableTags}},
		{name: "garbage tags", tr: &Track{TaggedTrack: trackUnknownFormatError}, want: []string{trackDiffBadTags}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tr.FindDifferences()
			sort.Strings(got)
			sort.Strings(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Track.FindDifferences() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateTracks(t *testing.T) {
	// 500 artists, 20 albums each, 50 tracks apiece ... total: 500,000 tracks
	var artists []*Artist
	for k := 0; k < 500; k++ {
		artist := &Artist{
			Name:   fmt.Sprintf("artist %d", k),
			Albums: make([]*Album, 0),
		}
		artists = append(artists, artist)
		for m := 0; m < 20; m++ {
			album := &Album{
				Name:            fmt.Sprintf("album %d-%d", k, m),
				Tracks:          make([]*Track, 0),
				RecordingArtist: artist,
			}
			artist.Albums = append(artist.Albums, album)
			for n := 0; n < 50; n++ {
				track := &Track{
					Name:            fmt.Sprintf("track %d-%d-%d", k, m, n),
					TaggedTrack:     trackUnknownTagsNotRead,
					ContainingAlbum: album,
				}
				album.Tracks = append(album.Tracks, track)
			}
		}
	}
	normalReader := func(path string) (*taggedTrackData, error) {
		return &taggedTrackData{
			album:  "beautiful album",
			artist: "great artist",
			title:  "terrific track",
			number: "1",
		}, nil
	}
	type args struct {
		artists []*Artist
		reader  func(string) (*taggedTrackData, error)
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "big test", args: args{artists: artists, reader: normalReader}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdateTracks(tt.args.artists, tt.args.reader)
			for _, artist := range tt.args.artists {
				for _, album := range artist.Albums {
					for _, track := range album.Tracks {
						if track.TaggedTrack != 1 {
							t.Errorf("UpdateTracks() %q track = %d", track.Name, track.TaggedTrack)
						}
					}
				}
			}
		})
	}
}

func TestRawReadTags(t *testing.T) {
	if err := internal.CreateFileForTesting(".", "goodFile.mp3"); err != nil {
		t.Errorf("failed to create ./goodFile.mp3")
	}
	defer func() {
		if err := os.Remove("./goodFile.mp3"); err != nil {
			t.Errorf("failed to delete ./goodFile.mp3")
		}
	}()
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantD   *taggedTrackData
		wantErr bool
	}{
		{name: "bad test", args: args{path: "./noSuchFile!.mp3"}, wantD: nil, wantErr: true},
		{name: "good test", args: args{path: "./goodFile.mp3"}, wantD: &taggedTrackData{}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotD, err := RawReadTags(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("RawReadTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(gotD, tt.wantD) {
				t.Errorf("RawReadTags() = %v, want %v", gotD, tt.wantD)
			}
		})
	}
}

func Test_removeLeadingBOMs(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "normal string", args: args{s: "normal"}, want: "normal"},
		{name: "abnormal string", args: args{s: "\ufeff\ufeffnormal"}, want: "normal"},
		{name: "empty string", args: args{}},
		{name: "nothing but BOM", args: args{s: "\ufeff\ufeff"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeLeadingBOMs(tt.args.s); got != tt.want {
				t.Errorf("removeLeadingBOMs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sortTracks(t *testing.T) {
	tests := []struct {
		name   string
		tracks []*Track
	}{
		{name: "degenerate case"},
		{
			name: "mixed tracks",
			tracks: []*Track{
				{
					TrackNumber:     10,
					ContainingAlbum: &Album{Name: "album2", RecordingArtist: &Artist{Name: "artist3"}},
				},
				{
					TrackNumber:     1,
					ContainingAlbum: &Album{Name: "album2", RecordingArtist: &Artist{Name: "artist3"}},
				},
				{
					TrackNumber:     2,
					ContainingAlbum: &Album{Name: "album1", RecordingArtist: &Artist{Name: "artist3"}},
				},
				{
					TrackNumber:     3,
					ContainingAlbum: &Album{Name: "album3", RecordingArtist: &Artist{Name: "artist2"}},
				},
				{
					TrackNumber:     3,
					ContainingAlbum: &Album{Name: "album3", RecordingArtist: &Artist{Name: "artist4"}},
				},
				{
					TrackNumber:     3,
					ContainingAlbum: &Album{Name: "album5", RecordingArtist: &Artist{Name: "artist2"}},
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
			album1 := track1.ContainingAlbum
			album2 := track2.ContainingAlbum
			artist1 := album1.RecordingArtist.Name
			artist2 := album2.RecordingArtist.Name
			if artist1 > artist2 {
				t.Errorf("Sort(Tracks) track[%d] artist name %q comes after track[%d] artist name %q", i-1, artist1, i, artist2)
			} else {
				if artist1 == artist2 {
					if album1.Name > album2.Name {
						t.Errorf("Sort(Tracks) track[%d] album name %q comes after track[%d] album name %q", i-1, album1.Name, i, album2.Name)
					} else {
						if album1.Name == album2.Name {
							if track1.TrackNumber > track2.TrackNumber {
								t.Errorf("Sort(Tracks) track[%d] track %d comes after track[%d] track %d", i-1, track1.TrackNumber, i, track2.TrackNumber)
							}
						}
					}
				}
			}
		}
	}
}

func TestTrack_EditTags(t *testing.T) {
	fnName := "Track.EditTags()"
	topDir := "editTags"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, topDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	testFileName := "test.mp3"
	fullPath := filepath.Join(topDir, testFileName)
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
	}
	content := CreateTaggedData(payload, frames)
	// // block off tag header
	// content = append(content, []byte("ID3")...)
	// content = append(content, []byte{3, 0, 0, 0, 0, 0, 0}...)
	// // add some text frames
	// contentLength := len(content) - 10
	// factor := 128 * 128 * 128
	// for k := 0; k < 4; k++ {
	// 	content[6+k] = byte(contentLength / factor)
	// 	contentLength = contentLength % factor
	// 	factor = factor / 128
	// }
	// // add "music"
	// for k := 0; k < 256; k++ {
	// 	content = append(content, byte(k))
	// }
	if err := internal.CreateFileForTestingWithContent(topDir, testFileName, string(content)); err != nil {
		t.Errorf("%s cannot create file %q: %v", fnName, fullPath, err)
	}
	tests := []struct {
		name    string
		tr      *Track
		wantErr bool
	}{
		{
			name: "defective track",
			tr: &Track{
				TrackNumber:     1,
				Name:            "defective track",
				TaggedTrack:     trackUnknownTagsNotRead,
				ContainingAlbum: &Album{Name: "poor album", RecordingArtist: &Artist{Name: "sorry artist"}},
			},
			wantErr: true,
		},
		{
			name: "track got deleted!",
			tr: &Track{
				TrackNumber:     1,
				Name:            "defective track",
				TaggedTrack:     1,
				TaggedTitle:     "unknown track",
				TaggedAlbum:     "unknown album",
				TaggedArtist:    "unknown artist",
				Path:            filepath.Join(topDir, "non-existent-file.mp3"),
				ContainingAlbum: &Album{Name: "poor album", RecordingArtist: &Artist{Name: "sorry artist"}},
			},
			wantErr: true,
		},
		{
			name: "fixable track",
			tr: &Track{
				TrackNumber:     1,
				Name:            "fixable track",
				TaggedTrack:     2,
				TaggedTitle:     "unknown track",
				TaggedAlbum:     "unknown album",
				TaggedArtist:    "unknown artist",
				Path:            fullPath,
				ContainingAlbum: &Album{Name: "poor album", RecordingArtist: &Artist{Name: "sorry artist"}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.tr.EditTags(); (err != nil) != tt.wantErr {
				t.Errorf("Track.EditTags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	if edited, err := os.ReadFile(fullPath); err != nil {
		t.Errorf("%s cannot read file %q: %v", fnName, fullPath, err)
	} else {
		if tag, err := id3v2.ParseReader(bytes.NewReader(edited), id3v2.Options{Parse: true}); err != nil {
			t.Errorf("%s edited mp3 file %q cannot be read for tags: %v", fnName, fullPath, err)
		} else {
			m := map[string]string{
				// changed by editing
				"TALB": "poor album",
				"TIT2": "fixable track",
				"TPE1": "sorry artist",
				"TRCK": "1",
				// preserved from original file
				"TCOM": "a couple of idiots",
				"TCON": "dance music",
				"TLEN": "1000",
				"TYER": "2022",
			}
			for key, value := range m {
				if got := tag.GetTextFrame(key).Text; got != value {
					t.Errorf("%s edited mp3 file key %q got %q want %q", fnName, key, got, value)
				}
			}
		}
		// verify "music" is present
		musicStarts := len(edited) - 256
		for k := 0; k < 256; k++ {
			if edited[musicStarts+k] != byte(k) {
				t.Errorf("%s edited mp3 file music at index %d mismatch (%d v. %d)", fnName, k, edited[musicStarts+k], k)
			}
		}
	}
}

func TestTrack_BackupDirectory(t *testing.T) {
	tests := []struct {
		name string
		tr   *Track
		want string
	}{
		{
			name: "simple case",
			tr:   &Track{ContainingAlbum: &Album{Path: "albumPath"}},
			want: "albumPath\\pre-repair-backup",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.BackupDirectory(); got != tt.want {
				t.Errorf("Track.BackupDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}
