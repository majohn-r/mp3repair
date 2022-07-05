package files

import (
	"bytes"
	"fmt"
	"mp3/internal"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/bogem/id3v2/v2"
)

func Test_parseTrackName(t *testing.T) {
	fnName := "parseTrackName()"
	type args struct {
		name  string
		album *Album
		ext   string
	}
	tests := []struct {
		name            string
		args            args
		wantSimpleName  string
		wantTrackNumber int
		wantValid       bool
		internal.WantedOutput
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
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The track \"59 track name.mp4\" on album \"some album\" by artist \"some artist\" cannot be parsed.\n",
				WantLogOutput:   "level='warn' albumName='some album' artistName='some artist' trackName='59 track name.mp4' msg='the track name cannot be parsed'\n",
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
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The track \"track name.mp3\" on album \"some album\" by artist \"some artist\" cannot be parsed.\n",
				WantLogOutput:   "level='warn' albumName='some album' artistName='some artist' trackName='track name.mp3' msg='the track name cannot be parsed'\n",
			},
			args: args{
				name:  "track name.mp3",
				album: &Album{name: "some album", recordingArtist: &Artist{name: "some artist"}},
				ext:   ".mp3",
			},
		},
		{
			name: "missing track number, simple name",
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The track \"trackName.mp3\" on album \"some album\" by artist \"some artist\" cannot be parsed.\n",
				WantLogOutput:   "level='warn' albumName='some album' artistName='some artist' trackName='trackName.mp3' msg='the track name cannot be parsed'\n",
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
			o := internal.NewOutputDeviceForTesting()
			gotSimpleName, gotTrackNumber, gotValid := parseTrackName(o, tt.args.name, tt.args.album, tt.args.ext)
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
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
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
		{name: "needs tagged data", tr: &Track{track: 0}, want: true},
		{name: "format error", tr: &Track{tagError: "format error"}, want: false},
		{name: "valid track number", tr: &Track{track: 1}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.needsTaggedData(); got != tt.want {
				t.Errorf("Track.needsTaggedData() = %v, want %v", got, tt.want)
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
		d *TaggedTrackData
	}
	tests := []struct {
		name       string
		tr         *Track
		args       args
		wantAlbum  string
		wantArtist string
		wantTitle  string
		wantNumber int
		wantError  string
	}{
		{
			name: "good input",
			tr:   &Track{},
			args: args{&TaggedTrackData{
				album:  "my excellent album",
				artist: "great artist",
				title:  "best track ever",
				track:  "1",
			}},
			wantAlbum:  "my excellent album",
			wantArtist: "great artist",
			wantTitle:  "best track ever",
			wantNumber: 1,
		},
		{
			name: "badly formatted input",
			tr:   &Track{},
			args: args{&TaggedTrackData{
				album:  "my excellent album",
				artist: "great artist",
				title:  "best track ever",
				track:  "foo",
			}},
		},
		{
			name: "negative track",
			tr:   &Track{},
			args: args{&TaggedTrackData{
				album:  "my excellent album",
				artist: "great artist",
				title:  "best track ever",
				track:  "-1",
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tr.SetTags(tt.args.d)
			if tt.tr.track != tt.wantNumber {
				t.Errorf("track.SetTags() tagged track = %d, want %d ", tt.tr.track, tt.wantNumber)
			}
			if tt.wantNumber != 0 {
				if tt.tr.album != tt.wantAlbum {
					t.Errorf("track.SetTags() tagged album = %q, want %q", tt.tr.album, tt.wantAlbum)
				}
				if tt.tr.artist != tt.wantArtist {
					t.Errorf("track.SetTags() tagged artist = %q, want %q", tt.tr.artist, tt.wantArtist)
				}
				if tt.tr.title != tt.wantTitle {
					t.Errorf("track.SetTags() tagged title = %q, want %q", tt.tr.title, tt.wantTitle)
				}
			}
		})
	}
}

func TestTrack_readTags(t *testing.T) {
	normalReader := func(path string) *TaggedTrackData {
		return &TaggedTrackData{
			album:  "beautiful album",
			artist: "great artist",
			title:  "terrific track",
			track:  "1",
		}
	}
	bentReader := func(path string) *TaggedTrackData {
		return &TaggedTrackData{
			album:  "beautiful album",
			artist: "great artist",
			title:  "terrific track",
			track:  "-2",
		}
	}
	brokenReader := func(path string) *TaggedTrackData {
		return &TaggedTrackData{err: "read error"}
	}
	type args struct {
		reader func(string) *TaggedTrackData
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
			tr:         &Track{track: trackUnknownTagsNotRead},
			args:       args{normalReader},
			wantAlbum:  "beautiful album",
			wantArtist: "great artist",
			wantTitle:  "terrific track",
			wantNumber: 1,
		},
		{
			name: "replay",
			tr: &Track{
				track:  2,
				album:  "nice album",
				artist: "good artist",
				title:  "pretty song",
			},
			args:       args{normalReader},
			wantAlbum:  "nice album",
			wantArtist: "good artist",
			wantTitle:  "pretty song",
			wantNumber: 2,
		},
		{
			name: "read error",
			tr:   &Track{path: "./unreadable track"},
			args: args{brokenReader},
		},
		{
			name: "format error",
			tr:   &Track{path: "./badly formatted track"},
			args: args{bentReader},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tr.readTags(tt.args.reader)
			waitForSemaphoresDrained()
			if tt.tr.track != tt.wantNumber {
				t.Errorf("track.readTags() tagged track = %d, want %d ", tt.tr.track, tt.wantNumber)
			}
			if tt.wantNumber >= 0 {
				if tt.tr.album != tt.wantAlbum {
					t.Errorf("track.readTags() tagged album = %q, want %q", tt.tr.album, tt.wantAlbum)
				}
				if tt.tr.artist != tt.wantArtist {
					t.Errorf("track.readTags() tagged artist = %q, want %q", tt.tr.artist, tt.wantArtist)
				}
				if tt.tr.title != tt.wantTitle {
					t.Errorf("track.readTags() tagged title = %q, want %q", tt.tr.title, tt.wantTitle)
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
		{name: "final period", args: args{nameTagPair{name: "simple name.", tag: "simple name."}}, want: true},
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
				number:          1,
				name:            "track name",
				containingAlbum: NewAlbum("album name", NewArtist("artist name", ""), ""),
				track:           1,
				title:           "track name",
				album:           "album name",
				artist:          "artist name",
			},
			want: nil,
		},
		{
			name: "another OK use case",
			tr: &Track{
				number:          1,
				name:            "track name",
				containingAlbum: NewAlbum("album name", NewArtist("artist name", ""), ""),
				track:           1,
				title:           "track:name",
				album:           "album:name",
				artist:          "artist:name",
			},
			want: nil,
		},
		{
			name: "oops",
			tr: &Track{
				number:          2,
				name:            "track:name",
				containingAlbum: NewAlbum("album:name", NewArtist("artist:name", ""), ""),
				track:           1,
				title:           "track name",
				album:           "album name",
				artist:          "artist name",
			},
			want: []string{
				"album \"album:name\" does not agree with album tag \"album name\"",
				"artist \"artist:name\" does not agree with artist tag \"artist name\"",
				"title \"track:name\" does not agree with title tag \"track name\"",
				"track number 2 does not agree with track tag 1",
			},
		},
		{name: "unread tags", tr: &Track{track: 0}, want: []string{trackDiffUnreadTags}},
		{name: "track with error", tr: &Track{track: 0, tagError: "oops"}, want: []string{trackDiffError}},
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
		artist := NewArtist(fmt.Sprintf("artist %d", k), "")
		artists = append(artists, artist)
		for m := 0; m < 20; m++ {
			album := NewAlbum(fmt.Sprintf("album %d-%d", k, m), artist, "")
			artist.AddAlbum(album)
			for n := 0; n < 50; n++ {
				track := &Track{
					name:            fmt.Sprintf("track %d-%d-%d", k, m, n),
					track:           trackUnknownTagsNotRead,
					containingAlbum: album,
				}
				album.AddTrack(track)
			}
		}
	}
	normalReader := func(path string) *TaggedTrackData {
		return &TaggedTrackData{
			album:  "beautiful album",
			artist: "great artist",
			title:  "terrific track",
			track:  "1",
		}
	}
	badReader := func(path string) *TaggedTrackData {
		return &TaggedTrackData{err: "read error"}
	}
	var artists2 []*Artist
	for k := 0; k < 500; k++ {
		artist := NewArtist(fmt.Sprintf("artist %d", k), "")
		artists2 = append(artists2, artist)
		for m := 0; m < 20; m++ {
			album := NewAlbum(fmt.Sprintf("album %d-%d", k, m), artist, "")
			artist.AddAlbum(album)
			for n := 0; n < 50; n++ {
				track := &Track{
					name:            fmt.Sprintf("track %d-%d-%d", k, m, n),
					track:           trackUnknownTagsNotRead,
					containingAlbum: album,
				}
				album.AddTrack(track)
			}
		}
	}
	var errors []string
	var logs []string
	for _, artist := range artists2 {
		for _, album := range artist.Albums() {
			for _, track := range album.Tracks() {
				errors = append(errors, fmt.Sprintf(internal.USER_TAG_ERROR, track.name, album.name, artist.name, "read error"))
				logs = append(logs, fmt.Sprintf("level='warn' albumName='%s' artistName='%s' error='read error' trackName='%s' msg='tag error'\n", album.name, artist.name, track.name))
			}
		}
	}
	type args struct {
		artists []*Artist
		reader  func(string) *TaggedTrackData
	}
	tests := []struct {
		name             string
		args             args
		checkTrackNumber bool
		internal.WantedOutput
	}{
		{name: "big test", args: args{artists: artists, reader: normalReader}, checkTrackNumber: true},
		{
			name: "massive failure",
			args: args{artists: artists2, reader: badReader},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: strings.Join(errors, ""),
				WantLogOutput:   strings.Join(logs, ""),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			UpdateTracks(o, tt.args.artists, tt.args.reader)
			if tt.checkTrackNumber {
				for _, artist := range tt.args.artists {
					for _, album := range artist.Albums() {
						for _, track := range album.Tracks() {
							if track.track != 1 {
								t.Errorf("UpdateTracks() %q track = %d", track.name, track.track)
							}
						}
					}
				}
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("UpdateTracks() %s", issue)
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
		name  string
		args  args
		wantD *TaggedTrackData
	}{
		{name: "bad test", args: args{path: "./noSuchFile!.mp3"}, wantD: &TaggedTrackData{err: "foo"}},
		{name: "good test", args: args{path: "./goodFile.mp3"}, wantD: &TaggedTrackData{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotD := RawReadTags(tt.args.path)
			if len(gotD.err) != 0 {
				if len(tt.wantD.err) == 0 {
					t.Errorf("RawReadTags() = %v, want %v", gotD, tt.wantD)
				}
			} else if len(tt.wantD.err) != 0 {
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
				t.Errorf("Sort(Tracks) track[%d] artist name %q comes after track[%d] artist name %q", i-1, artist1, i, artist2)
			} else {
				if artist1 == artist2 {
					if album1.Name() > album2.Name() {
						t.Errorf("Sort(Tracks) track[%d] album name %q comes after track[%d] album name %q", i-1, album1.Name(), i, album2.Name())
					} else {
						if album1.Name() == album2.Name() {
							if track1.number > track2.number {
								t.Errorf("Sort(Tracks) track[%d] track %d comes after track[%d] track %d", i-1, track1.number, i, track2.number)
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
	content := CreateTaggedDataForTesting(payload, frames)
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
				number:          1,
				name:            "defective track",
				track:           trackUnknownTagsNotRead,
				containingAlbum: NewAlbum("poor album", NewArtist("sorry artist", ""), ""),
			},
			wantErr: true,
		},
		{
			name: "track got deleted!",
			tr: &Track{
				number:          1,
				name:            "defective track",
				track:           1,
				title:           "unknown track",
				album:           "unknown album",
				artist:          "unknown artist",
				path:            filepath.Join(topDir, "non-existent-file.mp3"),
				containingAlbum: NewAlbum("poor album", NewArtist("sorry artist", ""), ""),
			},
			wantErr: true,
		},
		{
			name: "fixable track",
			tr: &Track{
				number:          1,
				name:            "fixable track",
				track:           2,
				title:           "unknown track",
				album:           "unknown album",
				artist:          "unknown artist",
				path:            fullPath,
				containingAlbum: NewAlbum("poor album", NewArtist("sorry artist", ""), ""),
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
			tr:   &Track{containingAlbum: NewAlbum("", nil, "albumPath")},
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

func TestTrack_AlbumPath(t *testing.T) {
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
				t.Errorf("Track.AlbumPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewTaggedTrackData(t *testing.T) {
	type args struct {
		albumFrame  string
		artistFrame string
		titleFrame  string
		numberFrame string
	}
	tests := []struct {
		name string
		args args
		want *TaggedTrackData
	}{
		{
			name: "usual",
			args: args{
				albumFrame:  "the album",
				artistFrame: "the artist",
				titleFrame:  "the title",
				numberFrame: "1",
			},
			want: &TaggedTrackData{
				album:  "the album",
				artist: "the artist",
				title:  "the title",
				track:  "1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTaggedTrackData(tt.args.albumFrame, tt.args.artistFrame, tt.args.titleFrame, tt.args.numberFrame)
			if got.album != tt.want.album || got.artist != tt.want.artist || got.title != tt.want.title || got.track != tt.want.track {
				t.Errorf("NewTaggedTrackData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_Copy(t *testing.T) {
	fnName := "Track.Copy()"
	topDir := "copies"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, topDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	srcName := "source.mp3"
	srcPath := filepath.Join(topDir, srcName)
	if err := internal.CreateFileForTesting(topDir, srcName); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, srcPath, err)
	}
	type args struct {
		destination string
	}
	tests := []struct {
		name    string
		tr      *Track
		args    args
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
				t.Errorf("Track.Copy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTrack_String(t *testing.T) {
	tests := []struct {
		name string
		tr   *Track
		want string
	}{{name: "expected", tr: &Track{path: "my path"}, want: "my path"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.String(); got != tt.want {
				t.Errorf("Track.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_Name(t *testing.T) {
	tests := []struct {
		name string
		tr   *Track
		want string
	}{{name: "expected", tr: &Track{name: "track name"}, want: "track name"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.Name(); got != tt.want {
				t.Errorf("Track.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_Number(t *testing.T) {
	tests := []struct {
		name string
		tr   *Track
		want int
	}{{name: "expected", tr: &Track{number: 42}, want: 42}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.Number(); got != tt.want {
				t.Errorf("Track.Number() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_AlbumName(t *testing.T) {
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
				t.Errorf("Track.AlbumName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_RecordingArtist(t *testing.T) {
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
				t.Errorf("Track.RecordingArtist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTrackNameForTesting(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name            string
		args            args
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
				t.Errorf("ParseTrackNameForTesting() gotSimpleName = %v, want %v", gotSimpleName, tt.wantSimpleName)
			}
			if gotTrackNumber != tt.wantTrackNumber {
				t.Errorf("ParseTrackNameForTesting() gotTrackNumber = %v, want %v", gotTrackNumber, tt.wantTrackNumber)
			}
		})
	}
}

func TestTrack_Path(t *testing.T) {
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
				t.Errorf("Track.Path() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_Directory(t *testing.T) {
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
				t.Errorf("Track.Directory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_FileName(t *testing.T) {
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
				t.Errorf("Track.FileName() = %v, want %v", got, tt.want)
			}
		})
	}
}
