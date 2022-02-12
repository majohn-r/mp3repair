package files

import (
	"reflect"
	"regexp"
	"testing"
)

func Test_validateExtension(t *testing.T) {
	type args struct {
		ext string
	}
	tests := []struct {
		name      string
		args      args
		wantValid bool
	}{
		{
			name: "valid extension",
			args: args{
				ext: ".mp3",
			},
			wantValid: true,
		},
		{
			name: "extension does not start with '.'",
			args: args{
				ext: "mp3",
			},
			wantValid: false,
		},
		{
			name: "extension contains multiple '.'",
			args: args{
				ext: ".m.p3",
			},
			wantValid: false,
		},
		{
			name: "extension contains invalid characters",
			args: args{
				ext: ".m[p3",
			},
			wantValid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotValid := validateExtension(tt.args.ext); gotValid != tt.wantValid {
				t.Errorf("validateExtension() = %v, want %v", gotValid, tt.wantValid)
			}
		})
	}
}

func Test_validateRegexp(t *testing.T) {
	type args struct {
		pattern string
		name    string
	}
	tests := []struct {
		name         string
		args         args
		wantFilter   *regexp.Regexp
		wantBadRegex bool
	}{
		{
			name: "valid filter with regex",
			args: args{
				pattern: "^.*$",
				name:    "artist",
			},
			wantFilter:   regexp.MustCompile("^.*$"),
			wantBadRegex: false,
		},
		{
			name: "valid simple filter",
			args: args{
				pattern: "Beatles",
				name:    "artist",
			},
			wantFilter:   regexp.MustCompile("Beatles"),
			wantBadRegex: false,
		},
		{
			name: "invalid filter",
			args: args{
				pattern: "disc[",
				name:    "album",
			},
			wantFilter:   nil,
			wantBadRegex: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFilter, gotBadRegex := validateRegexp(tt.args.pattern, tt.args.name)
			if !tt.wantBadRegex && !reflect.DeepEqual(gotFilter, tt.wantFilter) {
				t.Errorf("validateRegexp() gotFilter = %v, want %v", gotFilter, tt.wantFilter)
			}
			if gotBadRegex != tt.wantBadRegex {
				t.Errorf("validateRegexp() gotBadRegex = %v, want %v", gotBadRegex, tt.wantBadRegex)
			}
		})
	}
}

func Test_validateSearchParameters(t *testing.T) {
	type args struct {
		ext     string
		albums  string
		artists string
	}
	tests := []struct {
		name              string
		args              args
		wantAlbumsFilter  *regexp.Regexp
		wantArtistsFilter *regexp.Regexp
		wantProblemsExist bool
	}{
		{
			name: "valid input",
			args: args{
				ext:     ".mp3",
				albums:  ".*",
				artists: ".*",
			},
			wantAlbumsFilter:  regexp.MustCompile(".*"),
			wantArtistsFilter: regexp.MustCompile(".*"),
			wantProblemsExist: false,
		},
		{
			name: "bad extension 1",
			args: args{
				ext:     "mp3",
				albums:  ".*",
				artists: ".*",
			},
			wantProblemsExist: true,
		},
		{
			name: "bad extension 2",
			args: args{
				ext:     ".m.p3",
				albums:  ".*",
				artists: ".*",
			},
			wantProblemsExist: true,
		},
		{
			name: "bad extension 3",
			args: args{
				ext:     ".mp[3",
				albums:  ".*",
				artists: ".*",
			},
			wantProblemsExist: true,
		},
		{
			name: "bad album filter",
			args: args{
				ext:     ".mp3",
				albums:  ".[*",
				artists: ".*",
			},
			wantArtistsFilter: regexp.MustCompile(".*"),
			wantProblemsExist: true,
		},
		{
			name: "bad album filter",
			args: args{
				ext:     ".mp3",
				albums:  ".*",
				artists: ".[*",
			},
			wantAlbumsFilter:  regexp.MustCompile(".*"),
			wantProblemsExist: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAlbumsFilter, gotArtistsFilter, gotProblemsExist := validateSearchParameters(tt.args.ext, tt.args.albums, tt.args.artists)
			if !tt.wantProblemsExist {
				if !reflect.DeepEqual(gotAlbumsFilter, tt.wantAlbumsFilter) {
					t.Errorf("validateSearchParameters() gotAlbumsFilter = %v, want %v", gotAlbumsFilter, tt.wantAlbumsFilter)
				}
				if !reflect.DeepEqual(gotArtistsFilter, tt.wantArtistsFilter) {
					t.Errorf("validateSearchParameters() gotArtistsFilter = %v, want %v", gotArtistsFilter, tt.wantArtistsFilter)
				}
			}
			if gotProblemsExist != tt.wantProblemsExist {
				t.Errorf("validateSearchParameters() gotProblemsExist = %v, want %v", gotProblemsExist, tt.wantProblemsExist)
			}
		})
	}
}

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
			name: "expected use case",
			args: args{
				name:   "59 track name.mp3",
				album:  "some album",
				artist: "some artist",
				ext:    ".mp3",
			},
			wantSimpleName:  "track name",
			wantTrackNumber: 59,
			wantValid:       true,
		},
		{
			name: "wrong extension",
			args: args{
				name:   "59 track name.mp4",
				album:  "some album",
				artist: "some artist",
				ext:    ".mp3",
			},
			wantSimpleName:  "track name.mp4",
			wantTrackNumber: 59,
			wantValid:       false,
		},
		{
			name: "missing track number",
			args: args{
				name:   "track name.mp3",
				album:  "some album",
				artist: "some artist",
				ext:    ".mp3",
			},
			wantSimpleName:  "name",
			wantTrackNumber: 0,
			wantValid:       false,
		},
		{
			name: "missing track number, simple name",
			args: args{
				name:   "trackName.mp3",
				album:  "some album",
				artist: "some artist",
				ext:    ".mp3",
			},
			wantSimpleName:  "",
			wantTrackNumber: 0,
			wantValid:       false,
		},
	}
	validateExtension(".mp3")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSimpleName, gotTrackNumber, gotValid := parseTrackName(tt.args.name, tt.args.album, tt.args.artist, tt.args.ext)
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
