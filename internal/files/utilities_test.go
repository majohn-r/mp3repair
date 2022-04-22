package files

import (
	"fmt"
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
		{name: "negative value", args: args{s: "-12"}, wantI: 0, wantErr: true},
		{name: "invalid value", args: args{s: "foo"}, wantI: 0, wantErr: true},
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
