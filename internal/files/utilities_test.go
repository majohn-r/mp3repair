package files

import (
	"fmt"
	"os"
	"path/filepath"
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
		dir     string
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
				dir:     ".",
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
				dir:     ".",
				ext:     "mp3",
				albums:  ".*",
				artists: ".*",
			},
			wantProblemsExist: true,
		},
		{
			name: "bad extension 2",
			args: args{
				dir:     ".",
				ext:     ".m.p3",
				albums:  ".*",
				artists: ".*",
			},
			wantProblemsExist: true,
		},
		{
			name: "bad extension 3",
			args: args{
				dir:     ".",
				ext:     ".mp[3",
				albums:  ".*",
				artists: ".*",
			},
			wantProblemsExist: true,
		},
		{
			name: "bad album filter",
			args: args{
				dir:     ".",
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
				dir:     ".",
				ext:     ".mp3",
				albums:  ".*",
				artists: ".[*",
			},
			wantAlbumsFilter:  regexp.MustCompile(".*"),
			wantProblemsExist: true,
		},
		{
			name: "non-existent directory",
			args: args{
				dir:     "no such directory",
				ext:     ".mp3",
				albums:  ".*",
				artists: ".*",
			},
			wantAlbumsFilter:  regexp.MustCompile(".*"),
			wantArtistsFilter: regexp.MustCompile(".*"),
			wantProblemsExist: true,
		},
		{
			name: "directory is not a directory",
			args: args{
				dir:     "utilities_test.go",
				ext:     ".mp3",
				albums:  ".*",
				artists: ".*",
			},
			wantAlbumsFilter:  regexp.MustCompile(".*"),
			wantArtistsFilter: regexp.MustCompile(".*"),
			wantProblemsExist: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAlbumsFilter, gotArtistsFilter, gotProblemsExist := validateSearchParameters(tt.args.dir, tt.args.ext, tt.args.albums, tt.args.artists)
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

func Test_validateTopLevelDirectory(t *testing.T) {
	type args struct {
		dir string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "is directory",
			args: args{dir: "."},
			want: true,
		},
		{
			name: "non-existent directory",
			args: args{dir: "no such file"},
			want: false,
		},
		{
			name: "file that is not a directory",
			args: args{dir: "utilities_test.go"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateTopLevelDirectory(tt.args.dir); got != tt.want {
				t.Errorf("validateTopLevelDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadData(t *testing.T) {
	// generate test data
	topDir := "loadTest"
	if err := mkdir(t, "LoadData", topDir); err != nil {
		return
	}
	defer func() {
		if err := os.RemoveAll(topDir); err != nil {
			t.Errorf("LoadData() error destroying test directory %q: %v", topDir, err)
		}
	}()
	for k := 0; k < 10; k++ {
		createArtist(t, topDir, k)
	}
	type args struct {
		params *DirectorySearchParams
	}
	faultyParams := NewDirectorySearchParams(topDir, DefaultFileExtension, "^.*$", "^.*$")
	faultyParams.topDirectory = "no such directory"
	tests := []struct {
		name        string
		args        args
		wantArtists []*Artist
	}{
		{
			name:        "read all",
			args:        args{params: NewDirectorySearchParams(topDir, DefaultFileExtension, "^.*$", "^.*$")},
			wantArtists: createAllArtists(topDir),
		},
		{
			name: "no such top dir",
			args: args{faultyParams},
		},
		{
			name:        "read with filtering",
			args:        args{params: NewDirectorySearchParams(topDir, DefaultFileExtension, "^.*[02468]$", "^.*[13579]$")},
			wantArtists: createAllOddArtistsWithEvenAlbums(topDir),
		},
		{
			name: "read with all artists filtered out",
			args: args{params: NewDirectorySearchParams(topDir, DefaultFileExtension, "^.*$", "^.*X$")},
		},
		{
			name: "read with all albums filtered out",
			args: args{params: NewDirectorySearchParams(topDir, DefaultFileExtension, "^.*X$", "^.*$")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotArtists := LoadData(tt.args.params); !reflect.DeepEqual(gotArtists, tt.wantArtists) {
				t.Errorf("LoadData() = %v, want %v", gotArtists, tt.wantArtists)
			}
		})
	}
}

func createAllOddArtistsWithEvenAlbums(topDir string) []*Artist {
	var artists []*Artist
	for k := 1; k < 10; k += 2 {
		artistName := createArtistName(k)
		artistDir := filepath.Join(topDir, artistName)
		artist := &Artist{Name: artistName}
		for n := 0; n < 10; n += 2 {
			albumName := createAlbumName(n)
			albumDir := filepath.Join(artistDir, albumName)
			album := &Album{
				Name:            albumName,
				RecordingArtist: artist,
			}
			for p := 0; p < 10; p++ {
				trackName := createTrackName(p)
				name, _, _ := parseTrackName(trackName, albumName, artistName, DefaultFileExtension)
				track := &Track{
					fullPath:        filepath.Join(albumDir, trackName),
					fileName:        trackName,
					Name:            name,
					TrackNumber:     p,
					ContainingAlbum: album,
				}
				album.Tracks = append(album.Tracks, track)
			}
			artist.Albums = append(artist.Albums, album)
		}
		artists = append(artists, artist)
	}
	return artists
}

func createAllArtists(topDir string) []*Artist {
	var artists []*Artist
	for k := 0; k < 10; k++ {
		artistName := createArtistName(k)
		artistDir := filepath.Join(topDir, artistName)
		artist := &Artist{Name: artistName}
		for n := 0; n < 10; n++ {
			albumName := createAlbumName(n)
			albumDir := filepath.Join(artistDir, albumName)
			album := &Album{
				Name:            albumName,
				RecordingArtist: artist,
			}
			for p := 0; p < 10; p++ {
				trackName := createTrackName(p)
				name, _, _ := parseTrackName(trackName, albumName, artistName, DefaultFileExtension)
				track := &Track{
					fullPath:        filepath.Join(albumDir, trackName),
					fileName:        trackName,
					Name:            name,
					TrackNumber:     p,
					ContainingAlbum: album,
				}
				album.Tracks = append(album.Tracks, track)
			}
			artist.Albums = append(artist.Albums, album)
		}
		artists = append(artists, artist)
	}
	return artists
}

func createArtistName(k int) string {
	return fmt.Sprintf("Test Artist %d", k)
}

func createAlbumName(k int) string {
	return fmt.Sprintf("Test Album %d", k)
}

func createTrackName(k int) string {
	return fmt.Sprintf("%02d Test Track.mp3", k)
}

func mkdir(t *testing.T, fnName string, dirName string) error {
	if err := os.Mkdir(dirName, 0755); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, dirName, err)
		return err
	}
	return nil
}

func createArtist(t *testing.T, topDir string, k int) {
	artistDir := filepath.Join(topDir, createArtistName(k))
	if err := mkdir(t, "createArtist", artistDir); err == nil {
		for n := 0; n < 10; n++ {
			createAlbum(t, artistDir, n)
		}
		createFile(t, artistDir, "dummy file to be ignored.txt")
	}
}

func createFile(t *testing.T, dir, s string) {
	fileName := filepath.Join(dir, s)
	if err := os.WriteFile(fileName, []byte("file contents for "+s), 0644); err != nil {
		t.Errorf("createFile() error creating %q: %v", fileName, err)
	}
}

func createAlbum(t *testing.T, artistDir string, n int) {
	albumDir := filepath.Join(artistDir, createAlbumName(n))
	if err := mkdir(t, "createAlbum", albumDir); err == nil {
		for k := 0; k < 10; k++ {
			createTrack(t, albumDir, k)
		}
		createFile(t, albumDir, "album cover.jpeg")
		dummyDir := filepath.Join(albumDir, "ignore this folder")
		_ = mkdir(t, "createAlbum", dummyDir)
	}
}

func createTrack(t *testing.T, artistDir string, k int) {
	createFile(t, artistDir, createTrackName(k))
}
