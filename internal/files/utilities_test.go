package files

import (
	"fmt"
	"mp3/internal"
	"os"
	"reflect"
	"sort"
	"testing"
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
		{name: "tag read error", tr: &Track{TaggedTrack: trackUnknownTagReadError}, want: false},
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
			if tt.tr.TaggedTrack != trackUnknownTagReadError {
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
			tr:         &Track{TaggedTrack: trackUnknownTagReadError},
			args:       args{normalReader},
			wantNumber: trackUnknownTagReadError,
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
			wantNumber: trackUnknownTagReadError,
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

func Test_isIllegalRuneForFileNames(t *testing.T) {
	type args struct {
		r rune
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "0", args: args{r: 0}, want: true},
		{name: "1", args: args{r: 1}, want: true},
		{name: "2", args: args{r: 2}, want: true},
		{name: "3", args: args{r: 3}, want: true},
		{name: "4", args: args{r: 4}, want: true},
		{name: "5", args: args{r: 5}, want: true},
		{name: "6", args: args{r: 6}, want: true},
		{name: "7", args: args{r: 7}, want: true},
		{name: "8", args: args{r: 8}, want: true},
		{name: "9", args: args{r: 9}, want: true},
		{name: "10", args: args{r: 10}, want: true},
		{name: "11", args: args{r: 11}, want: true},
		{name: "12", args: args{r: 12}, want: true},
		{name: "13", args: args{r: 13}, want: true},
		{name: "14", args: args{r: 14}, want: true},
		{name: "15", args: args{r: 15}, want: true},
		{name: "16", args: args{r: 16}, want: true},
		{name: "17", args: args{r: 17}, want: true},
		{name: "18", args: args{r: 18}, want: true},
		{name: "19", args: args{r: 19}, want: true},
		{name: "20", args: args{r: 20}, want: true},
		{name: "21", args: args{r: 21}, want: true},
		{name: "22", args: args{r: 22}, want: true},
		{name: "23", args: args{r: 23}, want: true},
		{name: "24", args: args{r: 24}, want: true},
		{name: "25", args: args{r: 25}, want: true},
		{name: "26", args: args{r: 26}, want: true},
		{name: "27", args: args{r: 27}, want: true},
		{name: "28", args: args{r: 28}, want: true},
		{name: "29", args: args{r: 29}, want: true},
		{name: "30", args: args{r: 30}, want: true},
		{name: "31", args: args{r: 31}, want: true},
		{name: "<", args: args{r: '<'}, want: true},
		{name: ">", args: args{r: '>'}, want: true},
		{name: ":", args: args{r: ':'}, want: true},
		{name: "\"", args: args{r: '"'}, want: true},
		{name: "/", args: args{r: '/'}, want: true},
		{name: "\\", args: args{r: '\\'}, want: true},
		{name: "|", args: args{r: '|'}, want: true},
		{name: "?", args: args{r: '?'}, want: true},
		{name: "*", args: args{r: '*'}, want: true},
		{name: "!", args: args{r: '!'}, want: false},
		{name: "#", args: args{r: '#'}, want: false},
		{name: "$", args: args{r: '$'}, want: false},
		{name: "&", args: args{r: '&'}, want: false},
		{name: "'", args: args{r: '\''}, want: false},
		{name: "(", args: args{r: '('}, want: false},
		{name: ")", args: args{r: ')'}, want: false},
		{name: "+", args: args{r: '+'}, want: false},
		{name: ",", args: args{r: ','}, want: false},
		{name: "-", args: args{r: '-'}, want: false},
		{name: ".", args: args{r: '.'}, want: false},
		{name: "0", args: args{r: '0'}, want: false},
		{name: "1", args: args{r: '1'}, want: false},
		{name: "2", args: args{r: '2'}, want: false},
		{name: "3", args: args{r: '3'}, want: false},
		{name: "4", args: args{r: '4'}, want: false},
		{name: "5", args: args{r: '5'}, want: false},
		{name: "6", args: args{r: '6'}, want: false},
		{name: "7", args: args{r: '7'}, want: false},
		{name: "8", args: args{r: '8'}, want: false},
		{name: "9", args: args{r: '9'}, want: false},
		{name: ";", args: args{r: ';'}, want: false},
		{name: "A", args: args{r: 'A'}, want: false},
		{name: "B", args: args{r: 'B'}, want: false},
		{name: "C", args: args{r: 'C'}, want: false},
		{name: "D", args: args{r: 'D'}, want: false},
		{name: "E", args: args{r: 'E'}, want: false},
		{name: "F", args: args{r: 'F'}, want: false},
		{name: "G", args: args{r: 'G'}, want: false},
		{name: "H", args: args{r: 'H'}, want: false},
		{name: "I", args: args{r: 'I'}, want: false},
		{name: "J", args: args{r: 'J'}, want: false},
		{name: "K", args: args{r: 'K'}, want: false},
		{name: "L", args: args{r: 'L'}, want: false},
		{name: "M", args: args{r: 'M'}, want: false},
		{name: "N", args: args{r: 'N'}, want: false},
		{name: "O", args: args{r: 'O'}, want: false},
		{name: "P", args: args{r: 'P'}, want: false},
		{name: "Q", args: args{r: 'Q'}, want: false},
		{name: "R", args: args{r: 'R'}, want: false},
		{name: "S", args: args{r: 'S'}, want: false},
		{name: "T", args: args{r: 'T'}, want: false},
		{name: "U", args: args{r: 'U'}, want: false},
		{name: "V", args: args{r: 'V'}, want: false},
		{name: "W", args: args{r: 'W'}, want: false},
		{name: "X", args: args{r: 'X'}, want: false},
		{name: "Y", args: args{r: 'Y'}, want: false},
		{name: "Z", args: args{r: 'Z'}, want: false},
		{name: "[", args: args{r: '['}, want: false},
		{name: "]", args: args{r: ']'}, want: false},
		{name: "_", args: args{r: '_'}, want: false},
		{name: "a", args: args{r: 'a'}, want: false},
		{name: "b", args: args{r: 'b'}, want: false},
		{name: "c", args: args{r: 'c'}, want: false},
		{name: "d", args: args{r: 'd'}, want: false},
		{name: "e", args: args{r: 'e'}, want: false},
		{name: "f", args: args{r: 'f'}, want: false},
		{name: "g", args: args{r: 'g'}, want: false},
		{name: "h", args: args{r: 'h'}, want: false},
		{name: "i", args: args{r: 'i'}, want: false},
		{name: "j", args: args{r: 'j'}, want: false},
		{name: "k", args: args{r: 'k'}, want: false},
		{name: "l", args: args{r: 'l'}, want: false},
		{name: "m", args: args{r: 'm'}, want: false},
		{name: "n", args: args{r: 'n'}, want: false},
		{name: "o", args: args{r: 'o'}, want: false},
		{name: "p", args: args{r: 'p'}, want: false},
		{name: "q", args: args{r: 'q'}, want: false},
		{name: "r", args: args{r: 'r'}, want: false},
		{name: "s", args: args{r: 's'}, want: false},
		{name: "space", args: args{r: ' '}, want: false},
		{name: "t", args: args{r: 't'}, want: false},
		{name: "u", args: args{r: 'u'}, want: false},
		{name: "v", args: args{r: 'v'}, want: false},
		{name: "w", args: args{r: 'w'}, want: false},
		{name: "x", args: args{r: 'x'}, want: false},
		{name: "y", args: args{r: 'y'}, want: false},
		{name: "z", args: args{r: 'z'}, want: false},
		{name: "Á", args: args{r: 'Á'}, want: false},
		{name: "È", args: args{r: 'È'}, want: false},
		{name: "É", args: args{r: 'É'}, want: false},
		{name: "Ô", args: args{r: 'Ô'}, want: false},
		{name: "à", args: args{r: 'à'}, want: false},
		{name: "á", args: args{r: 'á'}, want: false},
		{name: "ã", args: args{r: 'ã'}, want: false},
		{name: "ä", args: args{r: 'ä'}, want: false},
		{name: "å", args: args{r: 'å'}, want: false},
		{name: "ç", args: args{r: 'ç'}, want: false},
		{name: "è", args: args{r: 'è'}, want: false},
		{name: "é", args: args{r: 'é'}, want: false},
		{name: "ê", args: args{r: 'ê'}, want: false},
		{name: "ë", args: args{r: 'ë'}, want: false},
		{name: "í", args: args{r: 'í'}, want: false},
		{name: "î", args: args{r: 'î'}, want: false},
		{name: "ï", args: args{r: 'ï'}, want: false},
		{name: "ñ", args: args{r: 'ñ'}, want: false},
		{name: "ò", args: args{r: 'ò'}, want: false},
		{name: "ó", args: args{r: 'ó'}, want: false},
		{name: "ô", args: args{r: 'ô'}, want: false},
		{name: "ö", args: args{r: 'ö'}, want: false},
		{name: "ø", args: args{r: 'ø'}, want: false},
		{name: "ù", args: args{r: 'ù'}, want: false},
		{name: "ú", args: args{r: 'ú'}, want: false},
		{name: "ü", args: args{r: 'ü'}, want: false},
		{name: "ř", args: args{r: 'ř'}, want: false},
		{name: "…", args: args{r: '…'}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isIllegalRuneForFileNames(tt.args.r); got != tt.want {
				t.Errorf("isIllegalRuneForFileNames() = %v, want %v", got, tt.want)
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
		{name: "unreadable tags", tr: &Track{TaggedTrack: trackUnknownTagReadError}, want: []string{trackDiffUnreadableTags}},
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
					TrackNumber: 10,
					ContainingAlbum: &Album{
						Name:            "album2",
						RecordingArtist: &Artist{Name: "artist3"},
					},
				},
				{
					TrackNumber: 1,
					ContainingAlbum: &Album{
						Name:            "album2",
						RecordingArtist: &Artist{Name: "artist3"},
					},
				},
				{
					TrackNumber: 2,
					ContainingAlbum: &Album{
						Name:            "album1",
						RecordingArtist: &Artist{Name: "artist3"},
					},
				},
				{
					TrackNumber: 3,
					ContainingAlbum: &Album{
						Name:            "album3",
						RecordingArtist: &Artist{Name: "artist2"},
					},
				},
				{
					TrackNumber: 3,
					ContainingAlbum: &Album{
						Name:            "album3",
						RecordingArtist: &Artist{Name: "artist4"},
					},
				},
				{
					TrackNumber: 3,
					ContainingAlbum: &Album{
						Name:            "album5",
						RecordingArtist: &Artist{Name: "artist2"},
					},
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
