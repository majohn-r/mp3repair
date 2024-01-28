package files

import (
	"mp3/internal"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bogem/id3v2/v2"
	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
)

const (
	cannotOpenFile  = "open readMetadata\\no such file.mp3: The system cannot find the file specified."
	negativeSeek    = "seek readMetadata\\01 tagless.mp3: An attempt was made to move the file pointer before the beginning of the file."
	noID3V1Metadata = "no id3v1 metadata found in file \"readMetadata\\\\03 id3v2.mp3\""
	zeroBytes       = "zero length"
)

func Test_trackMetadata_setId3v1Values(t *testing.T) {
	const fnName = "trackMetadata.setId3v1Values()"
	type args struct {
		v1 *id3v1Metadata
	}
	tests := map[string]struct {
		tM *TrackMetadata
		args
		wantTM *TrackMetadata
	}{
		"complete test": {
			tM: newTrackMetadata(), args: args{v1: newID3v1MetadataWithData(internal.ID3V1DataSet1)},
			wantTM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.tM.setID3v1Values(tt.args.v1)
			if !reflect.DeepEqual(tt.tM, tt.wantTM) {
				t.Errorf("%s got %v want %v", fnName, tt.tM, tt.wantTM)
			}
		})
	}
}

func Test_trackMetadata_setId3v2Values(t *testing.T) {
	const fnName = "trackMetadata.setId3v1Values()"
	type args struct {
		d *id3v2Metadata
	}
	tests := map[string]struct {
		tM *TrackMetadata
		args
		wantTM *TrackMetadata
	}{
		"complete test": {
			tM: newTrackMetadata(),
			args: args{
				d: &id3v2Metadata{
					album:             "Great album",
					artist:            "Great artist",
					title:             "Great track",
					genre:             "Pop",
					year:              "2022",
					track:             1,
					musicCDIdentifier: id3v2.UnknownFrame{Body: []byte{0, 2, 4}},
				},
			},
			wantTM: &TrackMetadata{
				Album:                      []string{"", "", "Great album"},
				Artist:                     []string{"", "", "Great artist"},
				Title:                      []string{"", "", "Great track"},
				Genre:                      []string{"", "", "Pop"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 1},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0, 2, 4}},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.tM.setID3v2Values(tt.args.d)
			if !reflect.DeepEqual(tt.tM, tt.wantTM) {
				t.Errorf("%s got %v want %v", fnName, tt.tM, tt.wantTM)
			}
		})
	}
}

func Test_readMetadata(t *testing.T) {
	const fnName = "readMetadata()"
	testDir := "readMetadata"
	if err := cmd_toolkit.Mkdir(testDir); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, testDir, err)
	}
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
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	type args struct {
		path string
	}
	tests := map[string]struct {
		args
		want *TrackMetadata
	}{
		"missing file": {
			args: args{path: filepath.Join(testDir, "no such file.mp3")},
			want: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", cannotOpenFile, cannotOpenFile},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
		"no metadata": {
			args: args{path: filepath.Join(testDir, taglessFile)},
			want: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", negativeSeek, errMissingTrackNumber.Error()},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
		"only id3v1 metadata": {
			args: args{path: filepath.Join(testDir, id3v1OnlyFile)},
			want: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", errMissingTrackNumber.Error()},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
		"only id3v2 metadata": {
			args: args{path: filepath.Join(testDir, id3v2OnlyFile)},
			want: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
		"all metadata": {
			args: args{path: filepath.Join(testDir, completeFile)},
			want: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := readMetadata(tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_trackMetadata_isValid(t *testing.T) {
	const fnName = "trackMetadata.isValid()"
	tests := map[string]struct {
		tM   *TrackMetadata
		want bool
	}{
		"uninitialized data": {tM: newTrackMetadata(), want: false},
		"after read failure": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", cannotOpenFile, cannotOpenFile},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: false,
		},
		"after reading no metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", negativeSeek, zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: false,
		},
		"after reading only id3v1 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: true,
		},
		"after reading only id3v2 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: true,
		},
		"after reading all metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tM.isValid(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_trackMetadata_canonicalArtist(t *testing.T) {
	const fnName = "trackMetadata.canonicalArtist()"
	tests := map[string]struct {
		tM   *TrackMetadata
		want string
	}{
		"after reading only id3v1 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: "The Beatles",
		},
		"after reading only id3v2 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: "unknown artist",
		},
		"after reading all metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: "unknown artist",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tM.canonicalArtist(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_trackMetadata_canonicalAlbum(t *testing.T) {
	const fnName = "trackMetadata.canonicalAlbum()"
	tests := map[string]struct {
		tM   *TrackMetadata
		want string
	}{
		"after reading only id3v1 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: "On Air: Live At The BBC, Volum",
		},
		"after reading only id3v2 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: "unknown album",
		},
		"after reading all metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: "unknown album",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tM.canonicalAlbum(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_trackMetadata_canonicalGenre(t *testing.T) {
	const fnName = "trackMetadata.canonicalGenre()"
	tests := map[string]struct {
		tM   *TrackMetadata
		want string
	}{
		"after reading only id3v1 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: "Other",
		},
		"after reading only id3v2 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: "dance music",
		},
		"after reading all metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: "dance music",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tM.canonicalGenre(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_trackMetadata_canonicalYear(t *testing.T) {
	const fnName = "trackMetadata.canonicalYear()"
	tests := map[string]struct {
		tM   *TrackMetadata
		want string
	}{
		"after reading only id3v1 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: "2013",
		},
		"after reading only id3v2 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: "2022",
		},
		"after reading all metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: "2022",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tM.canonicalYear(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_trackMetadata_canonicalMusicCDIdentifier(t *testing.T) {
	const fnName = "trackMetadata.canonicalMusicCDIdentifier()"
	tests := map[string]struct {
		tM   *TrackMetadata
		want id3v2.UnknownFrame
	}{
		"after reading only id3v1 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: id3v2.UnknownFrame{},
		},
		"after reading only id3v2 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: id3v2.UnknownFrame{Body: []byte{0}},
		},
		"after reading all metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: id3v2.UnknownFrame{Body: []byte{0}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tM.canonicalMusicCDIdentifier(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_trackMetadata_errors(t *testing.T) {
	const fnName = "trackMetadata.errors()"
	tests := map[string]struct {
		tM   *TrackMetadata
		want []string
	}{
		"uninitialized data": {tM: newTrackMetadata(), want: []string{}},
		"after read failure": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", cannotOpenFile, cannotOpenFile},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: []string{cannotOpenFile, cannotOpenFile},
		},
		"after reading no metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", negativeSeek, zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: []string{negativeSeek, zeroBytes},
		},
		"after reading only id3v1 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: []string{zeroBytes},
		},
		"after reading only id3v2 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: []string{noID3V1Metadata},
		},
		"after reading all metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			want: []string{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tM.errorCauses(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_trackMetadata_trackDiffers(t *testing.T) {
	const fnName = "trackMetadata.trackDiffers()"
	type args struct {
		track int
	}
	tests := map[string]struct {
		tM *TrackMetadata
		args
		want   bool
		wantTM *TrackMetadata
	}{
		"after read failure": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", cannotOpenFile, cannotOpenFile},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args: args{track: 20},
			want: false,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", cannotOpenFile, cannotOpenFile},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
		"after reading no metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", negativeSeek, zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args: args{track: 20},
			want: false,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", negativeSeek, zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
		"after reading only id3v1 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args: args{track: 20},
			want: true,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 20, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, true, false},
			},
		},
		"after reading only id3v2 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args: args{track: 20},
			want: true,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 20},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, true},
			},
		},
		"after reading all metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args: args{track: 20},
			want: true,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 20, 20},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, true, true},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
	const fnName = "trackMetadata.trackTitleDiffers()"
	type args struct {
		title string
	}
	tests := map[string]struct {
		tM *TrackMetadata
		args
		wantDiffers bool
		wantTM      *TrackMetadata
	}{
		"after read failure": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", cannotOpenFile, cannotOpenFile},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{title: "track name"},
			wantDiffers: false,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", cannotOpenFile, cannotOpenFile},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
		"after reading no metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", negativeSeek, zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{title: "track name"},
			wantDiffers: false,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", negativeSeek, zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
		"after reading only id3v1 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{title: "track name"},
			wantDiffers: true,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "track name", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, true, false},
			},
		},
		"after reading only id3v2 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{title: "track name"},
			wantDiffers: true,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", "track name"},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, true},
			},
		},
		"after reading all metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{title: "track name"},
			wantDiffers: true,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "track name", "track name"},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, true, true},
			},
		},
		"valid name": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Theme from M*A*S*H", "Theme from M*A*S*H"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{title: "Theme From M-A-S-H"},
			wantDiffers: false,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Theme from M*A*S*H", "Theme from M*A*S*H"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
	const fnName = "trackMetadata.albumTitleDiffers()"
	type args struct {
		albumTitle string
	}
	tests := map[string]struct {
		tM *TrackMetadata
		args
		wantDiffers bool
		wantTM      *TrackMetadata
	}{
		"after read failure": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", cannotOpenFile, cannotOpenFile},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{albumTitle: "album name"},
			wantDiffers: false,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", cannotOpenFile, cannotOpenFile},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
		"after reading no metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", negativeSeek, zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{albumTitle: "album name"},
			wantDiffers: false,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", negativeSeek, zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
		"after reading only id3v1 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{albumTitle: "album name"},
			wantDiffers: true,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "album name", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, true, false},
			},
		},
		"after reading only id3v2 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{albumTitle: "album name"},
			wantDiffers: true,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", "album name"},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, true},
			},
		},
		"after reading all metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{albumTitle: "album name"},
			wantDiffers: true,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "album name", "album name"},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, true, true},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
	const fnName = "trackMetadata.artistNameDiffers()"
	type args struct {
		artistName string
	}
	tests := map[string]struct {
		tM *TrackMetadata
		args
		wantDiffers bool
		wantTM      *TrackMetadata
	}{
		"after read failure": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", cannotOpenFile, cannotOpenFile},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{artistName: "artist name"},
			wantDiffers: false,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", cannotOpenFile, cannotOpenFile},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
		"after reading no metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", negativeSeek, zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{artistName: "artist name"},
			wantDiffers: false,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", negativeSeek, zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
		"after reading only id3v1 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{artistName: "artist name"},
			wantDiffers: true,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "artist name", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, true, false},
			},
		},
		"after reading only id3v2 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{artistName: "artist name"},
			wantDiffers: true,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", "artist name"},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, true},
			},
		},
		"after reading all metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{artistName: "artist name"},
			wantDiffers: true,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "artist name", "artist name"},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, true, true},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
	const fnName = "trackMetadata.genreDiffers()"
	type args struct {
		genre string
	}
	tests := map[string]struct {
		tM *TrackMetadata
		args
		wantDiffers bool
		wantTM      *TrackMetadata
	}{
		"after read failure": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", cannotOpenFile, cannotOpenFile},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{genre: "Indie Pop"},
			wantDiffers: false,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", cannotOpenFile, cannotOpenFile},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
		"after reading no metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", negativeSeek, zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{genre: "Indie Pop"},
			wantDiffers: false,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", negativeSeek, zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
		"after reading only id3v1 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{genre: "Indie Pop"},
			wantDiffers: false,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
		"after reading only id3v2 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{genre: "Indie Pop"},
			wantDiffers: true,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", "Indie Pop"},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, true},
			},
		},
		"after reading all metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{genre: "Indie Pop"},
			wantDiffers: true,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", "Indie Pop"},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, true},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
	const fnName = "trackMetadata.yearDiffers()"
	type args struct {
		year string
	}
	tests := map[string]struct {
		tM *TrackMetadata
		args
		wantDiffers bool
		wantTM      *TrackMetadata
	}{
		"after read failure": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", cannotOpenFile, cannotOpenFile},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{year: "1999"},
			wantDiffers: false,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", cannotOpenFile, cannotOpenFile},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
		"after reading no metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", negativeSeek, zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{year: "1999"},
			wantDiffers: false,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", negativeSeek, zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
		"after reading only id3v1 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{year: "1999"},
			wantDiffers: true,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "1999", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, true, false},
			},
		},
		"after reading only id3v2 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{year: "1999"},
			wantDiffers: true,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", "1999"},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, true},
			},
		},
		"after reading all metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{year: "1999"},
			wantDiffers: true,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "1999", "1999"},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, true, true},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
	const fnName = "trackMetadata.mcdiDiffers()"
	type args struct {
		f id3v2.UnknownFrame
	}
	tests := map[string]struct {
		tM *TrackMetadata
		args
		wantDiffers bool
		wantTM      *TrackMetadata
	}{
		"after read failure": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", cannotOpenFile, cannotOpenFile},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{f: id3v2.UnknownFrame{Body: []byte{1, 2, 3}}},
			wantDiffers: false,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", cannotOpenFile, cannotOpenFile},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
		"after reading no metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", negativeSeek, zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{f: id3v2.UnknownFrame{Body: []byte{1, 2, 3}}},
			wantDiffers: false,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "", ""},
				Artist:                     []string{"", "", ""},
				Title:                      []string{"", "", ""},
				Genre:                      []string{"", "", ""},
				Year:                       []string{"", "", ""},
				Track:                      []int{0, 0, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              UndefinedSource,
				ErrCause:                   []string{"", negativeSeek, zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
		"after reading only id3v1 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{f: id3v2.UnknownFrame{Body: []byte{1, 2, 3}}},
			wantDiffers: false,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
		},
		"after reading only id3v2 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{f: id3v2.UnknownFrame{Body: []byte{1, 2, 3}}},
			wantDiffers: true,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{Body: []byte{1, 2, 3}},
				RequiresEdit:               []bool{false, false, true},
			},
		},
		"after reading all metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args:        args{f: id3v2.UnknownFrame{Body: []byte{1, 2, 3}}},
			wantDiffers: true,
			wantTM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{Body: []byte{1, 2, 3}},
				RequiresEdit:               []bool{false, false, true},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
	const fnName = "trackMetadata.canonicalAlbumTitleMatches()"
	type args struct {
		albumTitle string
	}
	tests := map[string]struct {
		tM *TrackMetadata
		args
		want bool
	}{
		"mismatch after reading only id3v1 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args: args{albumTitle: "album name"},
			want: false,
		},
		"mismatch after reading only id3v2 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args: args{albumTitle: "album name"},
			want: false,
		},
		"mismatch after reading all metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args: args{albumTitle: "album name"},
			want: false,
		},
		"match after reading only id3v1 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args: args{albumTitle: "On Air: Live At The BBC, Volume 1"},
			want: true,
		},
		"match after reading only id3v2 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args: args{albumTitle: "unknown album"},
			want: true,
		},
		"match after reading all metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args: args{albumTitle: "unknown album"},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tM.canonicalAlbumTitleMatches(tt.args.albumTitle); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_trackMetadata_canonicalArtistNameMatches(t *testing.T) {
	const fnName = "trackMetadata.canonicalArtistNameMatches()"
	type args struct {
		artistName string
	}
	tests := map[string]struct {
		tM *TrackMetadata
		args
		want bool
	}{
		"mismatch after reading only id3v1 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args: args{artistName: "artist name"},
			want: false,
		},
		"mismatch after reading only id3v2 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args: args{artistName: "artist name"},
			want: false,
		},
		"mismatch after reading all metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args: args{artistName: "artist name"},
			want: false,
		},
		"match after reading only id3v1 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", ""},
				Artist:                     []string{"", "The Beatles", ""},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", ""},
				Genre:                      []string{"", "Other", ""},
				Year:                       []string{"", "2013", ""},
				Track:                      []int{0, 29, 0},
				MusicCDIdentifier:          id3v2.UnknownFrame{},
				CanonicalType:              ID3V1,
				ErrCause:                   []string{"", "", zeroBytes},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args: args{artistName: "The Beatles"},
			want: true,
		},
		"match after reading only id3v2 metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "", "unknown album"},
				Artist:                     []string{"", "", "unknown artist"},
				Title:                      []string{"", "", "unknown track"},
				Genre:                      []string{"", "", "dance music"},
				Year:                       []string{"", "", "2022"},
				Track:                      []int{0, 0, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", noID3V1Metadata, ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args: args{artistName: "unknown artist"},
			want: true,
		},
		"match after reading all metadata": {
			tM: &TrackMetadata{
				Album:                      []string{"", "On Air: Live At The BBC, Volum", "unknown album"},
				Artist:                     []string{"", "The Beatles", "unknown artist"},
				Title:                      []string{"", "Ringo - Pop Profile [Interview", "unknown track"},
				Genre:                      []string{"", "Other", "dance music"},
				Year:                       []string{"", "2013", "2022"},
				Track:                      []int{0, 29, 2},
				MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{0}},
				CanonicalType:              ID3V2,
				ErrCause:                   []string{"", "", ""},
				CorrectedAlbum:             []string{"", "", ""},
				CorrectedArtist:            []string{"", "", ""},
				CorrectedTitle:             []string{"", "", ""},
				CorrectedGenre:             []string{"", "", ""},
				CorrectedYear:              []string{"", "", ""},
				CorrectedTrack:             []int{0, 0, 0},
				CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
				RequiresEdit:               []bool{false, false, false},
			},
			args: args{artistName: "unknown artist"},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tM.canonicalArtistNameMatches(tt.args.artistName); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestSourceType_name(t *testing.T) {
	const fnName = "SourceType.name()"
	tests := map[string]struct {
		sT   SourceType
		want string
	}{
		"undefined": {sT: UndefinedSource, want: "undefined"},
		"ID3V1":     {sT: ID3V1, want: "ID3V1"},
		"ID3V2":     {sT: ID3V2, want: "ID3V2"},
		"total":     {sT: TotalSources, want: "total"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.sT.name(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}
