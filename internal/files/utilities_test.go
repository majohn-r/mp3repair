package files

import (
	"testing"
)

func Test_parseTrackName(t *testing.T) {
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
					t.Errorf("parseTrackName() gotSimpleName = %v, want %v", gotSimpleName, tt.wantSimpleName)
				}
				if gotTrackNumber != tt.wantTrackNumber {
					t.Errorf("parseTrackName() gotTrackNumber = %v, want %v", gotTrackNumber, tt.wantTrackNumber)
				}
			}
			if gotValid != tt.wantValid {
				t.Errorf("parseTrackName() gotValid = %v, want %v", gotValid, tt.wantValid)
			}
		})
	}
}
