package files

import (
	"fmt"
	"mp3/internal"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bogem/id3v2/v2"
)

func Test_trackMetadata_setId3v1Values(t *testing.T) {
	fnName := "trackMetadata.setId3v1Values()"
	type args struct {
		v1 *id3v1Metadata
	}
	tests := []struct {
		name string
		tM   *trackMetadata
		args
		wantTM *trackMetadata
	}{
		{
			name: "complete test",
			tM:   newTrackMetadata(),
			args: args{v1: newID3v1MetadataWithData(internal.ID3V1DataSet1)},
			wantTM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              undefinedSource,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tM.setID3v1Values(tt.args.v1)
			if !reflect.DeepEqual(tt.tM, tt.wantTM) {
				t.Errorf("%s got %v want %v", fnName, tt.tM, tt.wantTM)
			}
		})
	}
}

func Test_trackMetadata_setId3v2Values(t *testing.T) {
	fnName := "trackMetadata.setId3v1Values()"
	type args struct {
		d *ID3V2TaggedTrackData
	}
	tests := []struct {
		name string
		tM   *trackMetadata
		args
		wantTM *trackMetadata
	}{
		{
			name: "complete test",
			tM:   newTrackMetadata(),
			args: args{
				d: &ID3V2TaggedTrackData{
					album:             "Great album",
					artist:            "Great artist",
					title:             "Great track",
					genre:             "Pop",
					year:              "2022",
					track:             1,
					musicCDIdentifier: id3v2.UnknownFrame{Body: []byte{0, 2, 4}},
				},
			},
			wantTM: &trackMetadata{
				album:                      []string{"", "", "Great album"},
				artist:                     []string{"", "", "Great artist"},
				title:                      []string{"", "", "Great track"},
				genre:                      []string{"", "", "Pop"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 1},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0, 2, 4}},
				canonicalType:              undefinedSource,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tM.setID3v2Values(tt.args.d)
			if !reflect.DeepEqual(tt.tM, tt.wantTM) {
				t.Errorf("%s got %v want %v", fnName, tt.tM, tt.wantTM)
			}
		})
	}
}

func Test_readMetadata(t *testing.T) {
	fnName := "readMetadata()"
	testDir := "readMetadata"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, testDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	taglessFile := "01 tagless.mp3"
	if err := internal.CreateFileForTesting(testDir, taglessFile); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, taglessFile, err)
	}
	id3v1OnlyFile := "02 id3v1.mp3"
	payloadID3v1Only := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	payloadID3v1Only = append(payloadID3v1Only, internal.ID3V1DataSet1...)
	if err := internal.CreateFileForTestingWithContent(testDir, id3v1OnlyFile, payloadID3v1Only); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, id3v1OnlyFile, err)
	}
	id3v2OnlyFile := "03 id3v2.mp3"
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
	payloadID3v2Only := CreateID3V2TaggedDataForTesting([]byte{}, frames)
	if err := internal.CreateFileForTestingWithContent(testDir, id3v2OnlyFile, payloadID3v2Only); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, id3v2OnlyFile, err)
	}
	completeFile := "04 complete.mp3"
	payloadComplete := payloadID3v2Only
	payloadComplete = append(payloadComplete, payloadID3v1Only...)
	if err := internal.CreateFileForTestingWithContent(testDir, completeFile, payloadComplete); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, completeFile, err)
	}
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want *trackMetadata
	}{
		{
			name: "missing file",
			args: args{path: filepath.Join(testDir, "no such file.mp3")},
			want: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
		{
			name: "no tags",
			args: args{path: filepath.Join(testDir, taglessFile)},
			want: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("seek readMetadata\\01 tagless.mp3: An attempt was made to move the file pointer before the beginning of the file."),
					fmt.Errorf("zero length"),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
		{
			name: "only id3v1 tag",
			args: args{path: filepath.Join(testDir, id3v1OnlyFile)},
			want: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
		{
			name: "only id3v2 tag",
			args: args{path: filepath.Join(testDir, id3v2OnlyFile)},
			want: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
		{
			name: "both tags",
			args: args{path: filepath.Join(testDir, completeFile)},
			want: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := readMetadata(tt.args.path); !metadataEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func metadataEqual(got, want *trackMetadata) (ok bool) {
	if !reflect.DeepEqual(got.album, want.album) {
		return
	}
	if !reflect.DeepEqual(got.artist, want.artist) {
		return
	}
	if !reflect.DeepEqual(got.title, want.title) {
		return
	}
	if !reflect.DeepEqual(got.genre, want.genre) {
		return
	}
	if !reflect.DeepEqual(got.year, want.year) {
		return
	}
	if !reflect.DeepEqual(got.track, want.track) {
		return
	}
	if !reflect.DeepEqual(got.musicCDIdentifier, want.musicCDIdentifier) {
		return
	}
	if !reflect.DeepEqual(got.canonicalType, want.canonicalType) {
		return
	}
	if len(got.err) != len(want.err) {
		return
	}
	if !errorSliceEqual(got.err, want.err) {
		return
	}
	if !reflect.DeepEqual(got.correctedAlbum, want.correctedAlbum) {
		return
	}
	if !reflect.DeepEqual(got.correctedArtist, want.correctedArtist) {
		return
	}
	if !reflect.DeepEqual(got.correctedTitle, want.correctedTitle) {
		return
	}
	if !reflect.DeepEqual(got.correctedGenre, want.correctedGenre) {
		return
	}
	if !reflect.DeepEqual(got.correctedYear, want.correctedYear) {
		return
	}
	if !reflect.DeepEqual(got.correctedTrack, want.correctedTrack) {
		return
	}
	if !reflect.DeepEqual(got.correctedMusicCDIdentifier, want.correctedMusicCDIdentifier) {
		return
	}
	if !reflect.DeepEqual(got.requiresEdit, want.requiresEdit) {
		return
	}
	ok = true
	return
}

func errorSliceEqual(got, want []error) (ok bool) {
	if len(got) != len(want) {
		return
	}
	for i, e := range got {
		if e == nil && want[i] != nil {
			return
		}
		if e != nil && want[i] == nil {
			return
		}
		if e != nil && want[i] != nil {
			if e.Error() != want[i].Error() {
				return
			}
		}
	}
	ok = true
	return
}

func Test_trackMetadata_isValid(t *testing.T) {
	fnName := "trackMetadata.isValid()"
	tests := []struct {
		name string
		tM   *trackMetadata
		want bool
	}{
		{
			name: "uninitialized data",
			tM:   newTrackMetadata(),
			want: false,
		},
		{
			name: "after read failure",
			tM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: false,
		},
		{
			name: "after reading no tags",
			tM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("seek readMetadata\\01 tagless.mp3: An attempt was made to move the file pointer before the beginning of the file."),
					fmt.Errorf("zero length"),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: false,
		},
		{
			name: "after reading only id3v1 tag",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: true,
		},
		{
			name: "after reading only id3v2 tag",
			tM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: true,
		},
		{
			name: "after reading both tags",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tM.isValid(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_trackMetadata_canonicalArtist(t *testing.T) {
	fnName := "trackMetadata.canonicalArtist()"
	tests := []struct {
		name string
		tM   *trackMetadata
		want string
	}{
		{
			name: "after reading only id3v1 tag",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: "The Beatles",
		},
		{
			name: "after reading only id3v2 tag",
			tM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: "unknown artist",
		},
		{
			name: "after reading both tags",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: "unknown artist",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tM.canonicalArtist(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_trackMetadata_canonicalAlbum(t *testing.T) {
	fnName := "trackMetadata.canonicalAlbum()"
	tests := []struct {
		name string
		tM   *trackMetadata
		want string
	}{
		{
			name: "after reading only id3v1 tag",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: "On Air: Live At The BBC, Volum",
		},
		{
			name: "after reading only id3v2 tag",
			tM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: "unknown album",
		},
		{
			name: "after reading both tags",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: "unknown album",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tM.canonicalAlbum(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_trackMetadata_canonicalGenre(t *testing.T) {
	fnName := "trackMetadata.canonicalGenre()"
	tests := []struct {
		name string
		tM   *trackMetadata
		want string
	}{
		{
			name: "after reading only id3v1 tag",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: "Other",
		},
		{
			name: "after reading only id3v2 tag",
			tM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: "dance music",
		},
		{
			name: "after reading both tags",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: "dance music",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tM.canonicalGenre(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_trackMetadata_canonicalYear(t *testing.T) {
	fnName := "trackMetadata.canonicalYear()"
	tests := []struct {
		name string
		tM   *trackMetadata
		want string
	}{
		{
			name: "after reading only id3v1 tag",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: "2013",
		},
		{
			name: "after reading only id3v2 tag",
			tM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: "2022",
		},
		{
			name: "after reading both tags",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: "2022",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tM.canonicalYear(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_trackMetadata_canonicalMusicCDIdentifier(t *testing.T) {
	fnName := "trackMetadata.canonicalMusicCDIdentifier()"
	tests := []struct {
		name string
		tM   *trackMetadata
		want id3v2.UnknownFrame
	}{
		{
			name: "after reading only id3v1 tag",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: id3v2.UnknownFrame{},
		},
		{
			name: "after reading only id3v2 tag",
			tM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: id3v2.UnknownFrame{Body: []byte{0}},
		},
		{
			name: "after reading both tags",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: id3v2.UnknownFrame{Body: []byte{0}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tM.canonicalMusicCDIdentifier(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_trackMetadata_errors(t *testing.T) {
	fnName := "trackMetadata.errors()"
	tests := []struct {
		name string
		tM   *trackMetadata
		want []error
	}{
		{
			name: "uninitialized data",
			tM:   newTrackMetadata(),
			want: []error{},
		},
		{
			name: "after read failure",
			tM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: []error{
				fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
				fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
			},
		},
		{
			name: "after reading no tags",
			tM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("seek readMetadata\\01 tagless.mp3: An attempt was made to move the file pointer before the beginning of the file."),
					fmt.Errorf("zero length"),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: []error{
				fmt.Errorf("seek readMetadata\\01 tagless.mp3: An attempt was made to move the file pointer before the beginning of the file."),
				fmt.Errorf("zero length"),
			},
		},
		{
			name: "after reading only id3v1 tag",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: []error{fmt.Errorf("zero length")},
		},
		{
			name: "after reading only id3v2 tag",
			tM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: []error{fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\"")},
		},
		{
			name: "after reading both tags",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			want: []error{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tM.errors(); !errorSliceEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_trackMetadata_trackDiffers(t *testing.T) {
	fnName := "trackMetadata.trackDiffers()"
	type args struct {
		track int
	}
	tests := []struct {
		name string
		tM   *trackMetadata
		args
		want   bool
		wantTM *trackMetadata
	}{
		{
			name: "after read failure",
			tM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args: args{track: 20},
			want: false,
			wantTM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
		{
			name: "after reading no tags",
			tM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("seek readMetadata\\01 tagless.mp3: An attempt was made to move the file pointer before the beginning of the file."),
					fmt.Errorf("zero length"),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args: args{track: 20},
			want: false,
			wantTM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("seek readMetadata\\01 tagless.mp3: An attempt was made to move the file pointer before the beginning of the file."),
					fmt.Errorf("zero length"),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
		{
			name: "after reading only id3v1 tag",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args: args{track: 20},
			want: true,
			wantTM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 20, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, true, false},
			},
		},
		{
			name: "after reading only id3v2 tag",
			tM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args: args{track: 20},
			want: true,
			wantTM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 20},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, true},
			},
		},
		{
			name: "after reading both tags",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args: args{track: 20},
			want: true,
			wantTM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 20, 20},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, true, true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tM.trackDiffers(tt.args.track); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
			if !reflect.DeepEqual(tt.tM, tt.wantTM) {
				t.Errorf("%s got TM %v, want TM %v", fnName, tt.tM, tt.wantTM)
			}
		})
	}
}

func Test_trackMetadata_trackTitleDiffers(t *testing.T) {
	fnName := "trackMetadata.trackTitleDiffers()"
	type args struct {
		title string
	}
	tests := []struct {
		name string
		tM   *trackMetadata
		args
		wantDiffers bool
		wantTM      *trackMetadata
	}{
		{
			name: "after read failure",
			tM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{title: "track name"},
			wantDiffers: false,
			wantTM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
		{
			name: "after reading no tags",
			tM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("seek readMetadata\\01 tagless.mp3: An attempt was made to move the file pointer before the beginning of the file."),
					fmt.Errorf("zero length"),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{title: "track name"},
			wantDiffers: false,
			wantTM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("seek readMetadata\\01 tagless.mp3: An attempt was made to move the file pointer before the beginning of the file."),
					fmt.Errorf("zero length"),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
		{
			name: "after reading only id3v1 tag",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{title: "track name"},
			wantDiffers: true,
			wantTM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "track name", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, true, false},
			},
		},
		{
			name: "after reading only id3v2 tag",
			tM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{title: "track name"},
			wantDiffers: true,
			wantTM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", "track name"},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, true},
			},
		},
		{
			name: "after reading both tags",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{title: "track name"},
			wantDiffers: true,
			wantTM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "track name", "track name"},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, true, true},
			},
		},
		{
			name: "valid name",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Theme from M*A*S*H", "Theme from M*A*S*H"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{title: "Theme From M-A-S-H"},
			wantDiffers: false,
			wantTM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Theme from M*A*S*H", "Theme from M*A*S*H"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotDiffers := tt.tM.trackTitleDiffers(tt.args.title); gotDiffers != tt.wantDiffers {
				t.Errorf("%s = %v, want %v", fnName, gotDiffers, tt.wantDiffers)
			}
			if !reflect.DeepEqual(tt.tM, tt.wantTM) {
				t.Errorf("%s got TM %v, want TM %v", fnName, tt.tM, tt.wantTM)
			}
		})
	}
}

func Test_trackMetadata_albumTitleDiffers(t *testing.T) {
	fnName := "trackMetadata.albumTitleDiffers()"
	type args struct {
		albumTitle string
	}
	tests := []struct {
		name string
		tM   *trackMetadata
		args
		wantDiffers bool
		wantTM      *trackMetadata
	}{
		{
			name: "after read failure",
			tM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{albumTitle: "album name"},
			wantDiffers: false,
			wantTM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
		{
			name: "after reading no tags",
			tM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("seek readMetadata\\01 tagless.mp3: An attempt was made to move the file pointer before the beginning of the file."),
					fmt.Errorf("zero length"),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{albumTitle: "album name"},
			wantDiffers: false,
			wantTM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("seek readMetadata\\01 tagless.mp3: An attempt was made to move the file pointer before the beginning of the file."),
					fmt.Errorf("zero length"),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
		{
			name: "after reading only id3v1 tag",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{albumTitle: "album name"},
			wantDiffers: true,
			wantTM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "album name", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, true, false},
			},
		},
		{
			name: "after reading only id3v2 tag",
			tM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{albumTitle: "album name"},
			wantDiffers: true,
			wantTM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", "album name"},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, true},
			},
		},
		{
			name: "after reading both tags",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{albumTitle: "album name"},
			wantDiffers: true,
			wantTM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "album name", "album name"},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, true, true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotDiffers := tt.tM.albumTitleDiffers(tt.args.albumTitle); gotDiffers != tt.wantDiffers {
				t.Errorf("%s = %v, want %v", fnName, gotDiffers, tt.wantDiffers)
			}
			if !reflect.DeepEqual(tt.tM, tt.wantTM) {
				t.Errorf("%s got TM %v, want TM %v", fnName, tt.tM, tt.wantTM)
			}
		})
	}
}

func Test_trackMetadata_artistNameDiffers(t *testing.T) {
	fnName := "trackMetadata.artistNameDiffers()"
	type args struct {
		artistName string
	}
	tests := []struct {
		name string
		tM   *trackMetadata
		args
		wantDiffers bool
		wantTM      *trackMetadata
	}{
		{
			name: "after read failure",
			tM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{artistName: "artist name"},
			wantDiffers: false,
			wantTM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
		{
			name: "after reading no tags",
			tM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("seek readMetadata\\01 tagless.mp3: An attempt was made to move the file pointer before the beginning of the file."),
					fmt.Errorf("zero length"),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{artistName: "artist name"},
			wantDiffers: false,
			wantTM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("seek readMetadata\\01 tagless.mp3: An attempt was made to move the file pointer before the beginning of the file."),
					fmt.Errorf("zero length"),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
		{
			name: "after reading only id3v1 tag",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{artistName: "artist name"},
			wantDiffers: true,
			wantTM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "artist name", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, true, false},
			},
		},
		{
			name: "after reading only id3v2 tag",
			tM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{artistName: "artist name"},
			wantDiffers: true,
			wantTM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", "artist name"},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, true},
			},
		},
		{
			name: "after reading both tags",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{artistName: "artist name"},
			wantDiffers: true,
			wantTM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "artist name", "artist name"},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, true, true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotDiffers := tt.tM.artistNameDiffers(tt.args.artistName); gotDiffers != tt.wantDiffers {
				t.Errorf("%s = %v, want %v", fnName, gotDiffers, tt.wantDiffers)
			}
			if !reflect.DeepEqual(tt.tM, tt.wantTM) {
				t.Errorf("%s got TM %v, want TM %v", fnName, tt.tM, tt.wantTM)
			}
		})
	}
}

func Test_trackMetadata_genreDiffers(t *testing.T) {
	fnName := "trackMetadata.genreDiffers()"
	type args struct {
		genre string
	}
	tests := []struct {
		name string
		tM   *trackMetadata
		args
		wantDiffers bool
		wantTM      *trackMetadata
	}{
		{
			name: "after read failure",
			tM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{genre: "Indie Pop"},
			wantDiffers: false,
			wantTM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
		{
			name: "after reading no tags",
			tM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("seek readMetadata\\01 tagless.mp3: An attempt was made to move the file pointer before the beginning of the file."),
					fmt.Errorf("zero length"),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{genre: "Indie Pop"},
			wantDiffers: false,
			wantTM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("seek readMetadata\\01 tagless.mp3: An attempt was made to move the file pointer before the beginning of the file."),
					fmt.Errorf("zero length"),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
		{
			name: "after reading only id3v1 tag",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{genre: "Indie Pop"},
			wantDiffers: false,
			wantTM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
		{
			name: "after reading only id3v2 tag",
			tM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{genre: "Indie Pop"},
			wantDiffers: true,
			wantTM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", "Indie Pop"},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, true},
			},
		},
		{
			name: "after reading both tags",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{genre: "Indie Pop"},
			wantDiffers: true,
			wantTM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", "Indie Pop"},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotDiffers := tt.tM.genreDiffers(tt.args.genre); gotDiffers != tt.wantDiffers {
				t.Errorf("%s = %v, want %v", fnName, gotDiffers, tt.wantDiffers)
			}
			if !reflect.DeepEqual(tt.tM, tt.wantTM) {
				t.Errorf("%s got TM %v, want TM %v", fnName, tt.tM, tt.wantTM)
			}
		})
	}
}

func Test_trackMetadata_yearDiffers(t *testing.T) {
	fnName := "trackMetadata.yearDiffers()"
	type args struct {
		year string
	}
	tests := []struct {
		name string
		tM   *trackMetadata
		args
		wantDiffers bool
		wantTM      *trackMetadata
	}{
		{
			name: "after read failure",
			tM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{year: "1999"},
			wantDiffers: false,
			wantTM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
		{
			name: "after reading no tags",
			tM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("seek readMetadata\\01 tagless.mp3: An attempt was made to move the file pointer before the beginning of the file."),
					fmt.Errorf("zero length"),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{year: "1999"},
			wantDiffers: false,
			wantTM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("seek readMetadata\\01 tagless.mp3: An attempt was made to move the file pointer before the beginning of the file."),
					fmt.Errorf("zero length"),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
		{
			name: "after reading only id3v1 tag",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{year: "1999"},
			wantDiffers: true,
			wantTM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "1999", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, true, false},
			},
		},
		{
			name: "after reading only id3v2 tag",
			tM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{year: "1999"},
			wantDiffers: true,
			wantTM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", "1999"},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, true},
			},
		},
		{
			name: "after reading both tags",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{year: "1999"},
			wantDiffers: true,
			wantTM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "1999", "1999"},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, true, true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotDiffers := tt.tM.yearDiffers(tt.args.year); gotDiffers != tt.wantDiffers {
				t.Errorf("%s = %v, want %v", fnName, gotDiffers, tt.wantDiffers)
			}
			if !reflect.DeepEqual(tt.tM, tt.wantTM) {
				t.Errorf("%s got TM %v, want TM %v", fnName, tt.tM, tt.wantTM)
			}
		})
	}
}

func Test_trackMetadata_mcdiDiffers(t *testing.T) {
	fnName := "trackMetadata.mcdiDiffers()"
	type args struct {
		f id3v2.UnknownFrame
	}
	tests := []struct {
		name string
		tM   *trackMetadata
		args
		wantDiffers bool
		wantTM      *trackMetadata
	}{
		{
			name: "after read failure",
			tM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{f: id3v2.UnknownFrame{Body: []byte{1, 2, 3}}},
			wantDiffers: false,
			wantTM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
					fmt.Errorf("open readMetadata\\no such file.mp3: The system cannot find the file specified."),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
		{
			name: "after reading no tags",
			tM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("seek readMetadata\\01 tagless.mp3: An attempt was made to move the file pointer before the beginning of the file."),
					fmt.Errorf("zero length"),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{f: id3v2.UnknownFrame{Body: []byte{1, 2, 3}}},
			wantDiffers: false,
			wantTM: &trackMetadata{
				album:             []string{"", "", ""},
				artist:            []string{"", "", ""},
				title:             []string{"", "", ""},
				genre:             []string{"", "", ""},
				year:              []string{"", "", ""},
				track:             []int{0, 0, 0},
				musicCDIdentifier: id3v2.UnknownFrame{},
				canonicalType:     undefinedSource,
				err: []error{
					nil,
					fmt.Errorf("seek readMetadata\\01 tagless.mp3: An attempt was made to move the file pointer before the beginning of the file."),
					fmt.Errorf("zero length"),
				},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
		{
			name: "after reading only id3v1 tag",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{f: id3v2.UnknownFrame{Body: []byte{1, 2, 3}}},
			wantDiffers: false,
			wantTM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
		},
		{
			name: "after reading only id3v2 tag",
			tM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{f: id3v2.UnknownFrame{Body: []byte{1, 2, 3}}},
			wantDiffers: true,
			wantTM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{Body: []byte{1, 2, 3}},
				requiresEdit:               []bool{false, false, true},
			},
		},
		{
			name: "after reading both tags",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args:        args{f: id3v2.UnknownFrame{Body: []byte{1, 2, 3}}},
			wantDiffers: true,
			wantTM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{Body: []byte{1, 2, 3}},
				requiresEdit:               []bool{false, false, true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotDiffers := tt.tM.mcdiDiffers(tt.args.f); gotDiffers != tt.wantDiffers {
				t.Errorf("%s = %v, want %v", fnName, gotDiffers, tt.wantDiffers)
			}
			if !reflect.DeepEqual(tt.tM, tt.wantTM) {
				t.Errorf("%s got TM %v, want TM %v", fnName, tt.tM, tt.wantTM)
			}
		})
	}
}

func Test_trackMetadata_canonicalAlbumTitleMatches(t *testing.T) {
	fnName := "trackMetadata.canonicalAlbumTitleMatches()"
	type args struct {
		albumTitle string
	}
	tests := []struct {
		name string
		tM   *trackMetadata
		args
		want bool
	}{
		{
			name: "mismatch after reading only id3v1 tag",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args: args{albumTitle: "album name"},
			want: false,
		},
		{
			name: "mismatch after reading only id3v2 tag",
			tM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args: args{albumTitle: "album name"},
			want: false,
		},
		{
			name: "mismatch after reading both tags",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args: args{albumTitle: "album name"},
			want: false,
		},
		{
			name: "match after reading only id3v1 tag",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args: args{albumTitle: "On Air: Live At The BBC, Volume 1"},
			want: true,
		},
		{
			name: "match after reading only id3v2 tag",
			tM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args: args{albumTitle: "unknown album"},
			want: true,
		},
		{
			name: "match after reading both tags",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args: args{albumTitle: "unknown album"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tM.canonicalAlbumTitleMatches(tt.args.albumTitle); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_trackMetadata_canonicalArtistNameMatches(t *testing.T) {
	fnName := "trackMetadata.canonicalArtistNameMatches()"
	type args struct {
		artistName string
	}
	tests := []struct {
		name string
		tM   *trackMetadata
		args
		want bool
	}{
		{
			name: "mismatch after reading only id3v1 tag",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args: args{artistName: "artist name"},
			want: false,
		},
		{
			name: "mismatch after reading only id3v2 tag",
			tM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args: args{artistName: "artist name"},
			want: false,
		},
		{
			name: "mismatch after reading both tags",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args: args{artistName: "artist name"},
			want: false,
		},
		{
			name: "match after reading only id3v1 tag",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				artist:                     []string{"", "The Beatles", ""},
				title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				genre:                      []string{"", "Other", ""},
				year:                       []string{"", "2013", ""},
				track:                      []int{0, 29, 0},
				musicCDIdentifier:          id3v2.UnknownFrame{},
				canonicalType:              id3v1Source,
				err:                        []error{nil, nil, fmt.Errorf("zero length")},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args: args{artistName: "The Beatles"},
			want: true,
		},
		{
			name: "match after reading only id3v2 tag",
			tM: &trackMetadata{
				album:                      []string{"", "", "unknown album"},
				artist:                     []string{"", "", "unknown artist"},
				title:                      []string{"", "", "unknown track"},
				genre:                      []string{"", "", "dance music"},
				year:                       []string{"", "", "2022"},
				track:                      []int{0, 0, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, fmt.Errorf("no id3v1 tag found in file \"readMetadata\\\\03 id3v2.mp3\""), nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args: args{artistName: "unknown artist"},
			want: true,
		},
		{
			name: "match after reading both tags",
			tM: &trackMetadata{
				album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				artist:                     []string{"", "The Beatles", "unknown artist"},
				title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				genre:                      []string{"", "Other", "dance music"},
				year:                       []string{"", "2013", "2022"},
				track:                      []int{0, 29, 2},
				musicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				canonicalType:              id3v2Source,
				err:                        []error{nil, nil, nil},
				correctedAlbum:             []string{"", "", ""},
				correctedArtist:            []string{"", "", ""},
				correctedTitle:             []string{"", "", ""},
				correctedGenre:             []string{"", "", ""},
				correctedYear:              []string{"", "", ""},
				correctedTrack:             []int{0, 0, 0},
				correctedMusicCDIdentifier: id3v2.UnknownFrame{},
				requiresEdit:               []bool{false, false, false},
			},
			args: args{artistName: "unknown artist"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tM.canonicalArtistNameMatches(tt.args.artistName); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}
