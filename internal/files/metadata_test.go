package files_test

import (
	"fmt"
	"mp3repair/internal/files"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bogem/id3v2/v2"
	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/spf13/afero"
)

const (
	cannotOpenFile = "open readMetadata\\no such file.mp3:" +
		" The system cannot find the file specified."
	negativeSeek = "seek readMetadata\\01 tagless.mp3:" +
		" An attempt was made to move the file pointer before the beginning of the file."
	noID3V1Metadata = "no id3v1 metadata found in file" +
		" \"readMetadata\\\\03 id3v2.mp3\""
	zeroBytes = "zero length"
)

func TestTrackMetadataV1_setId3v1Values(t *testing.T) {
	const fnName = "trackMetadata.setId3v1Values()"
	type args struct {
		v1 *files.Id3v1Metadata
	}
	tests := map[string]struct {
		tM *files.TrackMetadataV1
		args
		wantTM *files.TrackMetadataV1
	}{
		"complete test": {
			tM:   files.NewTrackMetadataV1(),
			args: args{v1: NewID3v1MetadataWithData(id3v1DataSet1)},
			wantTM: files.NewTrackMetadataV1().WithAlbumNames(
				[]string{"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames(
				[]string{"", "The Beatles", ""}).WithTrackNames(
				[]string{"", "Ringo - Pop Profile [Interview", ""}).WithGenres(
				[]string{"", "Other", ""}).WithYears(
				[]string{"", "2013", ""}).WithTrackNumbers([]int{0, 29, 0}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.tM.SetID3v1Values(tt.args.v1)
			if !reflect.DeepEqual(tt.tM, tt.wantTM) {
				t.Errorf("%s got %v want %v", fnName, tt.tM, tt.wantTM)
			}
		})
	}
}

func TestTrackMetadataV1_setId3v2Values(t *testing.T) {
	const fnName = "trackMetadata.setId3v1Values()"
	type args struct {
		d *files.Id3v2Metadata
	}
	tests := map[string]struct {
		tM *files.TrackMetadataV1
		args
		wantTM *files.TrackMetadataV1
	}{
		"complete test": {
			tM: files.NewTrackMetadataV1(),
			args: args{
				d: &files.Id3v2Metadata{
					AlbumTitle:        "Great album",
					ArtistName:        "Great artist",
					TrackName:         "Great track",
					Genre:             "Pop",
					Year:              "2022",
					TrackNumber:       1,
					MusicCDIdentifier: id3v2.UnknownFrame{Body: []byte{0, 2, 4}},
				},
			},
			wantTM: files.NewTrackMetadataV1().WithAlbumNames(
				[]string{"", "", "Great album"}).WithArtistNames(
				[]string{"", "", "Great artist"}).WithTrackNames(
				[]string{"", "", "Great track"}).WithGenres(
				[]string{"", "", "Pop"}).WithYears([]string{
				"", "", "2022"}).WithTrackNumbers([]int{
				0, 0, 1}).WithMusicCDIdentifier([]byte{0, 2, 4}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.tM.SetID3v2Values(tt.args.d)
			if !reflect.DeepEqual(tt.tM, tt.wantTM) {
				t.Errorf("%s got %v want %v", fnName, tt.tM, tt.wantTM)
			}
		})
	}
}

func TestReadRawMetadata(t *testing.T) {
	originalFileSystem := cmd_toolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmd_toolkit.AssignFileSystem(originalFileSystem)
	}()
	testDir := "ReadRawMetadata"
	cmd_toolkit.Mkdir(testDir)
	taglessFile := "01 tagless.mp3"
	createFile(testDir, taglessFile)
	id3v1OnlyFile := "02 id3v1.mp3"
	payloadID3v1Only := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	payloadID3v1Only = append(payloadID3v1Only, id3v1DataSet1...)
	createFileWithContent(testDir, id3v1OnlyFile, payloadID3v1Only)
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
	payloadID3v2Only := createID3v2TaggedData([]byte{}, frames)
	createFileWithContent(testDir, id3v2OnlyFile, payloadID3v2Only)
	completeFile := "04 complete.mp3"
	payloadComplete := payloadID3v2Only
	payloadComplete = append(payloadComplete, payloadID3v1Only...)
	createFileWithContent(testDir, completeFile, payloadComplete)
	type args struct {
		path string
	}
	tests := map[string]struct {
		args
		want *files.TrackMetadataV1
	}{
		"missing file": {
			args: args{path: filepath.Join(testDir, "no such file.mp3")},
			want: files.NewTrackMetadataV1().WithErrorCauses([]string{
				"",
				"open ReadRawMetadata\\no such file.mp3: file does not exist",
				"open ReadRawMetadata\\no such file.mp3: file does not exist",
			}),
		},
		"no metadata": {
			args: args{path: filepath.Join(testDir, taglessFile)},
			want: files.NewTrackMetadataV1().WithErrorCauses([]string{
				"",
				"no ID3V1 metadata found",
				"no ID3V2 metadata found",
			}),
		},
		"only id3v1 metadata": {
			args: args{path: filepath.Join(testDir, id3v1OnlyFile)},
			want: files.NewTrackMetadataV1().WithAlbumNames(
				[]string{"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames(
				[]string{"", "The Beatles", ""}).WithTrackNames(
				[]string{"", "Ringo - Pop Profile [Interview", ""}).WithGenres(
				[]string{"", "Other", ""}).WithYears([]string{
				"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses(
				[]string{"", "", "no ID3V2 metadata found"}),
		},
		"only id3v2 metadata": {
			args: args{path: filepath.Join(testDir, id3v2OnlyFile)},
			want: files.NewTrackMetadataV1().WithAlbumNames(
				[]string{"", "", "unknown album"}).WithArtistNames(
				[]string{"", "", "unknown artist"}).WithTrackNames(
				[]string{"", "", "unknown track"}).WithGenres(
				[]string{"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers(
				[]int{0, 0, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2).WithErrorCauses(
				[]string{"", "no ID3V1 metadata found", ""}),
		},
		"all metadata": {
			args: args{path: filepath.Join(testDir, completeFile)},
			want: files.NewTrackMetadataV1().WithAlbumNames(
				[]string{"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames(
				[]string{"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears(
				[]string{"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.ReadRawMetadata(tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", "ReadRawMetadata()", got, tt.want)
			}
		})
	}
}

func TestTrackMetadataV1_isValid(t *testing.T) {
	const fnName = "trackMetadata.isValid()"
	tests := map[string]struct {
		tM   *files.TrackMetadataV1
		want bool
	}{
		"uninitialized data": {tM: files.NewTrackMetadataV1(), want: false},
		"after read failure": {
			tM: files.NewTrackMetadataV1().WithErrorCauses([]string{
				"", cannotOpenFile, cannotOpenFile,
			}),
			want: false,
		},
		"after reading no metadata": {
			tM: files.NewTrackMetadataV1().WithErrorCauses([]string{
				"", negativeSeek, zeroBytes}),
			want: false,
		},
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			want: true,
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers(
				[]int{0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			want: true,
		},
		"after reading all metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tM.IsValid(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestTrackMetadataV1_canonicalArtist(t *testing.T) {
	const fnName = "trackMetadata.CanonicalArtist()"
	tests := map[string]struct {
		tM   *files.TrackMetadataV1
		want string
	}{
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			want: "The Beatles",
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears([]string{
				"", "", "2022"}).WithTrackNumbers([]int{0, 0, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2).WithErrorCauses([]string{
				"", noID3V1Metadata, ""}),
			want: "unknown artist",
		},
		"after reading all metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			want: "unknown artist",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tM.CanonicalArtist(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestTrackMetadataV1_canonicalAlbum(t *testing.T) {
	const fnName = "trackMetadata.CanonicalAlbum()"
	tests := map[string]struct {
		tM   *files.TrackMetadataV1
		want string
	}{
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears(
				[]string{"", "2013", ""}).WithTrackNumbers([]int{
				0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			want: "On Air: Live At The BBC, Volum",
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers(
				[]int{0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			want: "unknown album",
		},
		"after reading all metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			want: "unknown album",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tM.CanonicalAlbum(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestTrackMetadataV1_canonicalGenre(t *testing.T) {
	const fnName = "trackMetadata.CanonicalGenre()"
	tests := map[string]struct {
		tM   *files.TrackMetadataV1
		want string
	}{
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			want: "Other",
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers(
				[]int{0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			want: "dance music",
		},
		"after reading all metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			want: "dance music",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tM.CanonicalGenre(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestTrackMetadataV1_canonicalYear(t *testing.T) {
	const fnName = "trackMetadata.CanonicalYear()"
	tests := map[string]struct {
		tM   *files.TrackMetadataV1
		want string
	}{
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears(
				[]string{"", "2013", ""}).WithTrackNumbers([]int{
				0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			want: "2013",
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers(
				[]int{0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			want: "2022",
		},
		"after reading all metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers([]int{
				0, 29, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(files.ID3V2),
			want: "2022",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tM.CanonicalYear(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestTrackMetadataV1_canonicalMusicCDIdentifier(t *testing.T) {
	const fnName = "trackMetadata.CanonicalMusicCDIdentifier()"
	tests := map[string]struct {
		tM   *files.TrackMetadataV1
		want id3v2.UnknownFrame
	}{
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{
				"", "2013", ""}).WithTrackNumbers([]int{0, 29, 0}).WithPrimarySource(
				files.ID3V1).WithErrorCauses([]string{"", "", zeroBytes}),
			want: id3v2.UnknownFrame{},
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears([]string{
				"", "", "2022"}).WithTrackNumbers([]int{0, 0, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			want: id3v2.UnknownFrame{Body: []byte{0}},
		},
		"after reading all metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album",
			}).WithArtistNames([]string{
				"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track",
			}).WithGenres([]string{
				"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			want: id3v2.UnknownFrame{Body: []byte{0}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tM.CanonicalMusicCDIdentifier(); !reflect.DeepEqual(got,
				tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestTrackMetadataV1_errors(t *testing.T) {
	const fnName = "trackMetadata.errors()"
	tests := map[string]struct {
		tM   *files.TrackMetadataV1
		want []string
	}{
		"uninitialized data": {tM: files.NewTrackMetadataV1(), want: []string{}},
		"after read failure": {
			tM: files.NewTrackMetadataV1().WithErrorCauses([]string{
				"", cannotOpenFile, cannotOpenFile}),
			want: []string{cannotOpenFile, cannotOpenFile},
		},
		"after reading no metadata": {
			tM: files.NewTrackMetadataV1().WithErrorCauses([]string{
				"", negativeSeek, zeroBytes}),
			want: []string{negativeSeek, zeroBytes},
		},
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears(
				[]string{"", "2013", ""}).WithTrackNumbers([]int{
				0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			want: []string{zeroBytes},
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers([]int{
				0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			want: []string{noID3V1Metadata},
		},
		"after reading all metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album",
			}).WithArtistNames([]string{
				"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track",
			}).WithGenres([]string{
				"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			want: []string{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tM.ErrorCauses(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestTrackMetadataV1_TrackNumberDiffers(t *testing.T) {
	tests := map[string]struct {
		tM     *files.TrackMetadataV1
		track  int
		want   bool
		wantTM *files.TrackMetadataV1
	}{
		"after read failure": {
			tM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
			track: 20,
			want:  false,
			wantTM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
		},
		"after reading no metadata": {
			tM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
			track: 20,
			want:  false,
			wantTM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
		},
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears(
				[]string{"", "2013", ""}).WithTrackNumbers([]int{
				0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			track: 20,
			want:  true,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears(
				[]string{"", "2013", ""}).WithTrackNumbers([]int{
				0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}).WithCorrectedTrackNumbers(
				[]int{0, 20, 0}).WithRequiresEdits([]bool{
				false, true, false}),
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers([]int{
				0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			track: 20,
			want:  true,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers([]int{
				0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{
				"", noID3V1Metadata, ""}).WithCorrectedTrackNumbers(
				[]int{0, 0, 20}).WithRequiresEdits(
				[]bool{false, false, true}),
		},
		"after reading all metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers([]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			track: 20,
			want:  true,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears(
				[]string{"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithCorrectedTrackNumbers(
				[]int{0, 20, 20}).WithRequiresEdits([]bool{
				false, true, true}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tM.TrackNumberDiffers(tt.track); got != tt.want {
				t.Errorf("%s = %v, want %v", "trackMetadata.TrackDiffers()", got, tt.want)
			}
			if !reflect.DeepEqual(tt.tM, tt.wantTM) {
				t.Errorf("%s got TM %v, want TM %v", "trackMetadata.TrackDiffers()", tt.tM, tt.wantTM)
			}
		})
	}
}

func TestTrackMetadataV1_TrackTitleDiffers(t *testing.T) {
	const fnName = "trackMetadata.TrackTitleDiffers()"
	type args struct {
		title string
	}
	tests := map[string]struct {
		tM *files.TrackMetadataV1
		args
		wantDiffers bool
		wantTM      *files.TrackMetadataV1
	}{
		"after read failure": {
			tM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
			args:        args{title: "track name"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
		},
		"after reading no metadata": {
			tM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
			args:        args{title: "track name"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
		},
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			args:        args{title: "track name"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}).WithCorrectedTrackNames([]string{
				"", "track name", ""}).WithRequiresEdits([]bool{false, true, false}),
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers([]int{
				0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			args:        args{title: "track name"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers([]int{
				0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{
				"", noID3V1Metadata, ""}).WithCorrectedTrackNames([]string{
				"", "", "track name"}).WithRequiresEdits([]bool{false, false, true}),
		},
		"after reading all metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2),
			args:        args{title: "track name"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears(
				[]string{"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithCorrectedTrackNames([]string{
				"", "track name", "track name"}).WithRequiresEdits([]bool{false, true, true}),
		},
		"valid name": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Theme from M*A*S*H", "Theme from M*A*S*H"}).WithGenres([]string{
				"", "Other", "dance music"}).WithYears(
				[]string{"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2),
			args:        args{title: "Theme From M-A-S-H"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Theme from M*A*S*H", "Theme from M*A*S*H"}).WithGenres([]string{
				"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if gotDiffers := tt.tM.TrackTitleDiffers(tt.args.title); gotDiffers != tt.wantDiffers {
				t.Errorf("%s = %v, want %v", fnName, gotDiffers, tt.wantDiffers)
			}
			if !reflect.DeepEqual(tt.tM, tt.wantTM) {
				t.Errorf("%s got TM %v, want TM %v", fnName, tt.tM, tt.wantTM)
			}
		})
	}
}

func TestTrackMetadataV1_AlbumTitleDiffers(t *testing.T) {
	const fnName = "trackMetadata.AlbumTitleDiffers()"
	type args struct {
		albumTitle string
	}
	tests := map[string]struct {
		tM *files.TrackMetadataV1
		args
		wantDiffers bool
		wantTM      *files.TrackMetadataV1
	}{
		"after read failure": {
			tM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
			args:        args{albumTitle: "album name"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
		},
		"after reading no metadata": {
			tM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
			args:        args{albumTitle: "album name"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
		},
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			args:        args{albumTitle: "album name"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}).WithCorrectedAlbumNames([]string{
				"", "album name", ""}).WithRequiresEdits([]bool{false, true, false}),
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers([]int{
				0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			args:        args{albumTitle: "album name"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers([]int{
				0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{
				"", noID3V1Metadata, ""}).WithCorrectedAlbumNames([]string{
				"", "", "album name"}).WithRequiresEdits([]bool{false, false, true}),
		},
		"after reading all metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			args:        args{albumTitle: "album name"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2).WithCorrectedAlbumNames([]string{
				"", "album name", "album name"}).WithRequiresEdits([]bool{false, true, true}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if gotDiffers := tt.tM.AlbumTitleDiffers(
				tt.args.albumTitle); gotDiffers != tt.wantDiffers {
				t.Errorf("%s = %v, want %v", fnName, gotDiffers, tt.wantDiffers)
			}
			if !reflect.DeepEqual(tt.tM, tt.wantTM) {
				t.Errorf("%s got TM %v, want TM %v", fnName, tt.tM, tt.wantTM)
			}
		})
	}
}

func TestTrackMetadataV1_ArtistNameDiffers(t *testing.T) {
	const fnName = "trackMetadata.ArtistNameDiffers()"
	type args struct {
		artistName string
	}
	tests := map[string]struct {
		tM *files.TrackMetadataV1
		args
		wantDiffers bool
		wantTM      *files.TrackMetadataV1
	}{
		"after read failure": {
			tM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
			args:        args{artistName: "artist name"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
		},
		"after reading no metadata": {
			tM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
			args:        args{artistName: "artist name"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
		},
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			args:        args{artistName: "artist name"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}).WithCorrectedArtistNames([]string{
				"", "artist name", ""}).WithRequiresEdits([]bool{false, true, false}),
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers([]int{
				0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			args:        args{artistName: "artist name"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers([]int{
				0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{
				"", noID3V1Metadata, ""}).WithCorrectedArtistNames([]string{
				"", "", "artist name"}).WithRequiresEdits([]bool{false, false, true}),
		},
		"after reading all metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			args:        args{artistName: "artist name"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2).WithCorrectedArtistNames([]string{
				"", "artist name", "artist name"}).WithRequiresEdits([]bool{
				false, true, true}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if gotDiffers := tt.tM.ArtistNameDiffers(tt.args.artistName); gotDiffers != tt.wantDiffers {
				t.Errorf("%s = %v, want %v", fnName, gotDiffers, tt.wantDiffers)
			}
			if !reflect.DeepEqual(tt.tM, tt.wantTM) {
				t.Errorf("%s got TM %v, want TM %v", fnName, tt.tM, tt.wantTM)
			}
		})
	}
}

func TestTrackMetadataV1_GenreDiffers(t *testing.T) {
	const fnName = "trackMetadata.GenreDiffers()"
	type args struct {
		genre string
	}
	tests := map[string]struct {
		tM *files.TrackMetadataV1
		args
		wantDiffers bool
		wantTM      *files.TrackMetadataV1
	}{
		"after read failure": {
			tM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
			args:        args{genre: "Indie Pop"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
		},
		"after reading no metadata": {
			tM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
			args:        args{genre: "Indie Pop"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
		},
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses(
				[]string{"", "", zeroBytes}),
			args:        args{genre: "Indie Pop"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses(
				[]string{"", "", zeroBytes}),
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers(
				[]int{0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			args:        args{genre: "Indie Pop"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers(
				[]int{0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{
				"", noID3V1Metadata, ""}).WithCorrectedGenres([]string{
				"", "", "Indie Pop"}).WithRequiresEdits([]bool{false, false, true}),
		},
		"after reading all metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			args:        args{genre: "Indie Pop"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2).WithCorrectedGenres([]string{
				"", "", "Indie Pop"}).WithRequiresEdits([]bool{false, false, true}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if gotDiffers := tt.tM.GenreDiffers(
				tt.args.genre); gotDiffers != tt.wantDiffers {
				t.Errorf("%s = %v, want %v", fnName, gotDiffers, tt.wantDiffers)
			}
			if !reflect.DeepEqual(tt.tM, tt.wantTM) {
				t.Errorf("%s got TM %v, want TM %v", fnName, tt.tM, tt.wantTM)
			}
		})
	}
}

func TestTrackMetadataV1_YearDiffers(t *testing.T) {
	const fnName = "trackMetadata.YearDiffers()"
	type args struct {
		year string
	}
	tests := map[string]struct {
		tM *files.TrackMetadataV1
		args
		wantDiffers bool
		wantTM      *files.TrackMetadataV1
	}{
		"after read failure": {
			tM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
			args:        args{year: "1999"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
		},
		"after reading no metadata": {
			tM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
			args:        args{year: "1999"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
		},
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			args:        args{year: "1999"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}).WithCorrectedYears([]string{
				"", "1999", ""}).WithRequiresEdits([]bool{false, true, false}),
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers(
				[]int{0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			args:        args{year: "1999"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers(
				[]int{0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{
				"", noID3V1Metadata, ""}).WithCorrectedYears([]string{
				"", "", "1999"}).WithRequiresEdits([]bool{false, false, true}),
		},
		"after reading all metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			args:        args{year: "1999"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2).WithCorrectedYears([]string{
				"", "1999", "1999"}).WithRequiresEdits([]bool{false, true, true}),
		},
		"no mismatch on years": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "1968", "1968 (2018)"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			args:        args{year: "1968"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "1968", "1968 (2018)"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if gotDiffers := tt.tM.YearDiffers(
				tt.args.year); gotDiffers != tt.wantDiffers {
				t.Errorf("%s = %v, want %v", fnName, gotDiffers, tt.wantDiffers)
			}
			if !reflect.DeepEqual(tt.tM, tt.wantTM) {
				t.Errorf("%s got TM %v, want TM %v", fnName, tt.tM, tt.wantTM)
			}
		})
	}
}

func TestTrackMetadataV1_MCDIDiffers(t *testing.T) {
	const fnName = "trackMetadata.MCDIDiffers()"
	type args struct {
		f id3v2.UnknownFrame
	}
	tests := map[string]struct {
		tM *files.TrackMetadataV1
		args
		wantDiffers bool
		wantTM      *files.TrackMetadataV1
	}{
		"after read failure": {
			tM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
			args:        args{f: id3v2.UnknownFrame{Body: []byte{1, 2, 3}}},
			wantDiffers: false,
			wantTM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
		},
		"after reading no metadata": {
			tM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
			args:        args{f: id3v2.UnknownFrame{Body: []byte{1, 2, 3}}},
			wantDiffers: false,
			wantTM: files.NewTrackMetadataV1().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
		},
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			args:        args{f: id3v2.UnknownFrame{Body: []byte{1, 2, 3}}},
			wantDiffers: false,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers(
				[]int{0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			args:        args{f: id3v2.UnknownFrame{Body: []byte{1, 2, 3}}},
			wantDiffers: true,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers(
				[]int{0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{
				"", noID3V1Metadata, ""}).WithCorrectedMusicCDIdentifier([]byte{
				1, 2, 3}).WithRequiresEdits([]bool{false, false, true}),
		},
		"after reading all metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears(
				[]string{"", "2013", "2022"}).WithTrackNumbers([]int{
				0, 29, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(files.ID3V2),
			args:        args{f: id3v2.UnknownFrame{Body: []byte{1, 2, 3}}},
			wantDiffers: true,
			wantTM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears(
				[]string{"", "2013", "2022"}).WithTrackNumbers([]int{
				0, 29, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithCorrectedMusicCDIdentifier([]byte{
				1, 2, 3}).WithRequiresEdits([]bool{false, false, true}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if gotDiffers := tt.tM.MCDIDiffers(tt.args.f); gotDiffers != tt.wantDiffers {
				t.Errorf("%s = %v, want %v", fnName, gotDiffers, tt.wantDiffers)
			}
			if !reflect.DeepEqual(tt.tM, tt.wantTM) {
				t.Errorf("%s got TM %v, want TM %v", fnName, tt.tM, tt.wantTM)
			}
		})
	}
}

func TestTrackMetadataV1_CanonicalAlbumTitleMatches(t *testing.T) {
	const fnName = "trackMetadata.CanonicalAlbumTitleMatches()"
	type args struct {
		albumTitle string
	}
	tests := map[string]struct {
		tM *files.TrackMetadataV1
		args
		want bool
	}{
		"mismatch after reading only id3v1 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears(
				[]string{"", "2013", ""}).WithTrackNumbers([]int{
				0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			args: args{albumTitle: "album name"},
			want: false,
		},
		"mismatch after reading only id3v2 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers(
				[]int{0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			args: args{albumTitle: "album name"},
			want: false,
		},
		"mismatch after reading all metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			args: args{albumTitle: "album name"},
			want: false,
		},
		"match after reading only id3v1 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears(
				[]string{"", "2013", ""}).WithTrackNumbers([]int{
				0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			args: args{albumTitle: "On Air: Live At The BBC, Volume 1"},
			want: true,
		},
		"match after reading only id3v2 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers([]int{
				0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			args: args{albumTitle: "unknown album"},
			want: true,
		},
		"match after reading all metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			args: args{albumTitle: "unknown album"},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tM.CanonicalAlbumTitleMatches(
				tt.args.albumTitle); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestTrackMetadataV1_CanonicalArtistNameMatches(t *testing.T) {
	const fnName = "trackMetadata.CanonicalArtistNameMatches()"
	type args struct {
		artistName string
	}
	tests := map[string]struct {
		tM *files.TrackMetadataV1
		args
		want bool
	}{
		"mismatch after reading only id3v1 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears(
				[]string{"", "2013", ""}).WithTrackNumbers([]int{
				0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			args: args{artistName: "artist name"},
			want: false,
		},
		"mismatch after reading only id3v2 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers([]int{
				0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			args: args{artistName: "artist name"},
			want: false,
		},
		"mismatch after reading all metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album",
			}).WithArtistNames([]string{
				"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track",
			}).WithGenres([]string{
				"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			args: args{artistName: "artist name"},
			want: false,
		},
		"match after reading only id3v1 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears(
				[]string{"", "2013", ""}).WithTrackNumbers([]int{
				0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			args: args{artistName: "The Beatles"},
			want: true,
		},
		"match after reading only id3v2 metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers([]int{
				0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			args: args{artistName: "unknown artist"},
			want: true,
		},
		"match after reading all metadata": {
			tM: files.NewTrackMetadataV1().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			args: args{artistName: "unknown artist"},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tM.CanonicalArtistNameMatches(
				tt.args.artistName); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestSourceType_name(t *testing.T) {
	const fnName = "SourceType.name()"
	tests := map[string]struct {
		sT   files.SourceType
		want string
	}{
		"undefined": {sT: files.UndefinedSource, want: "undefined"},
		"ID3V1":     {sT: files.ID3V1, want: "ID3V1"},
		"ID3V2":     {sT: files.ID3V2, want: "ID3V2"},
		"total":     {sT: files.TotalSources, want: "total"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.sT.Name(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestYearsMatch(t *testing.T) {
	type args struct {
		metadataYear string
		albumYear    string
	}
	tests := map[string]struct {
		args
		want bool
	}{
		"two empty years": {
			args: args{metadataYear: "", albumYear: ""},
			want: true,
		},
		"empty metadata year": {
			args: args{metadataYear: "", albumYear: "1968"},
			want: false,
		},
		"empty album year": {
			args: args{metadataYear: "1968", albumYear: ""},
			want: false,
		},
		"match equal lengths": {
			args: args{metadataYear: "1968", albumYear: "1968"},
			want: true,
		},
		"mismatch equal lengths": {
			args: args{metadataYear: "1968", albumYear: "1969"},
			want: false,
		},
		"match album > metadata": {
			args: args{metadataYear: "1968", albumYear: "1968 (2018)"},
			want: true,
		},
		"mismatch album > metadata": {
			args: args{metadataYear: "1968", albumYear: "1969 (2019)"},
			want: false,
		},
		"match album < metadata": {
			args: args{metadataYear: "1968 (2018)", albumYear: "1968"},
			want: true,
		},
		"mismatch album < metadata": {
			args: args{metadataYear: "1968 (2018)", albumYear: "1969"},
			want: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.YearsMatch(tt.args.metadataYear, tt.args.albumYear); got != tt.want {
				t.Errorf("YearsMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrackMetadata_SetArtistName(t *testing.T) {
	type args struct {
		src  files.SourceType
		name string
	}
	tests := map[string]struct {
		tm *files.TrackMetadata
		args
		want string
	}{
		"id3v1": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V1, name: "my favorite old artist"},
			want: "my favorite old artist",
		},
		"id3v2": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V2, name: "my favorite new artist"},
			want: "my favorite new artist",
		},
		"unknown": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.UndefinedSource, name: "what artist?"},
			want: "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.tm.SetArtistName(tt.args.src, tt.args.name)
			if got := tt.tm.ArtistName(tt.args.src).Original(); got != tt.want {
				t.Errorf("TrackMetadata.ArtistName got %q want %q", got, tt.want)
			}
			tt.tm.SetCanonicalSource(tt.args.src)
			if got := tt.tm.CanonicalArtistName(); got != tt.want {
				t.Errorf("TrackMetadata.CanonicalArtistName got %q want %q", got, tt.want)
			}
		})
	}
}

func TestTrackMetadata_CorrectArtistName(t *testing.T) {
	type args struct {
		src  files.SourceType
		name string
	}
	tests := map[string]struct {
		tm *files.TrackMetadata
		args
		want string
	}{
		"id3v1": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V1, name: "my favorite old artist"},
			want: "my favorite old artist",
		},
		"id3v2": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V2, name: "my favorite new artist"},
			want: "my favorite new artist",
		},
		"unknown": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.UndefinedSource, name: "what artist?"},
			want: "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.tm.CorrectArtistName(tt.args.src, tt.args.name)
			if got := tt.tm.ArtistName(tt.args.src).Correction(); got != tt.want {
				t.Errorf("TrackMetadata.ArtistName got %q want %q", got, tt.want)
			}
		})
	}
}

func TestTrackMetadata_SetAlbumName(t *testing.T) {
	type args struct {
		src  files.SourceType
		name string
	}
	tests := map[string]struct {
		tm *files.TrackMetadata
		args
		want string
	}{
		"id3v1": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V1, name: "my favorite old album"},
			want: "my favorite old album",
		},
		"id3v2": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V2, name: "my favorite new album"},
			want: "my favorite new album",
		},
		"unknown": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.UndefinedSource, name: "what album?"},
			want: "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.tm.SetAlbumName(tt.args.src, tt.args.name)
			if got := tt.tm.AlbumName(tt.args.src).Original(); got != tt.want {
				t.Errorf("TrackMetadata.AlbumName got %q want %q", got, tt.want)
			}
			tt.tm.SetCanonicalSource(tt.args.src)
			if got := tt.tm.CanonicalAlbumName(); got != tt.want {
				t.Errorf("TrackMetadata.CanonicalAlbumName got %q want %q", got, tt.want)
			}
		})
	}
}

func TestTrackMetadata_CorrectAlbumName(t *testing.T) {
	type args struct {
		src  files.SourceType
		name string
	}
	tests := map[string]struct {
		tm *files.TrackMetadata
		args
		want string
	}{
		"id3v1": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V1, name: "my favorite old album"},
			want: "my favorite old album",
		},
		"id3v2": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V2, name: "my favorite new album"},
			want: "my favorite new album",
		},
		"unknown": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.UndefinedSource, name: "what album?"},
			want: "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.tm.CorrectAlbumName(tt.args.src, tt.args.name)
			if got := tt.tm.AlbumName(tt.args.src).Correction(); got != tt.want {
				t.Errorf("TrackMetadata.AlbumName got %q want %q", got, tt.want)
			}
		})
	}
}

func TestTrackMetadata_SetAlbumGenre(t *testing.T) {
	type args struct {
		src  files.SourceType
		name string
	}
	tests := map[string]struct {
		tm *files.TrackMetadata
		args
		want string
	}{
		"id3v1": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V1, name: "old genre"},
			want: "old genre",
		},
		"id3v2": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V2, name: "new genre"},
			want: "new genre",
		},
		"unknown": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.UndefinedSource, name: "what genre?"},
			want: "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.tm.SetAlbumGenre(tt.args.src, tt.args.name)
			if got := tt.tm.AlbumGenre(tt.args.src).Original(); got != tt.want {
				t.Errorf("TrackMetadata.AlbumGenre got %q want %q", got, tt.want)
			}
			tt.tm.SetCanonicalSource(tt.args.src)
			if got := tt.tm.CanonicalAlbumGenre(); got != tt.want {
				t.Errorf("TrackMetadata.CanonicalAlbumGenre got %q want %q", got, tt.want)
			}
		})
	}
}

func TestTrackMetadata_CorrectAlbumGenre(t *testing.T) {
	type args struct {
		src  files.SourceType
		name string
	}
	tests := map[string]struct {
		tm *files.TrackMetadata
		args
		want string
	}{
		"id3v1": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V1, name: "old genre"},
			want: "old genre",
		},
		"id3v2": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V2, name: "new genre"},
			want: "new genre",
		},
		"unknown": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.UndefinedSource, name: "what genre?"},
			want: "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.tm.CorrectAlbumGenre(tt.args.src, tt.args.name)
			if got := tt.tm.AlbumGenre(tt.args.src).Correction(); got != tt.want {
				t.Errorf("TrackMetadata.AlbumGenre got %q want %q", got, tt.want)
			}
		})
	}
}

func TestTrackMetadata_SetAlbumYear(t *testing.T) {
	type args struct {
		src  files.SourceType
		name string
	}
	tests := map[string]struct {
		tm *files.TrackMetadata
		args
		want string
	}{
		"id3v1": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V1, name: "1900"},
			want: "1900",
		},
		"id3v2": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V2, name: "2000"},
			want: "2000",
		},
		"unknown": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.UndefinedSource, name: "1984?"},
			want: "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.tm.SetAlbumYear(tt.args.src, tt.args.name)
			if got := tt.tm.AlbumYear(tt.args.src).Original(); got != tt.want {
				t.Errorf("TrackMetadata.AlbumYear got %q want %q", got, tt.want)
			}
			tt.tm.SetCanonicalSource(tt.args.src)
			if got := tt.tm.CanonicalAlbumYear(); got != tt.want {
				t.Errorf("TrackMetadata.CanonicalAlbumYear got %q want %q", got, tt.want)
			}
		})
	}
}

func TestTrackMetadata_CorrectAlbumYear(t *testing.T) {
	type args struct {
		src  files.SourceType
		name string
	}
	tests := map[string]struct {
		tm *files.TrackMetadata
		args
		want string
	}{
		"id3v1": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V1, name: "1900"},
			want: "1900",
		},
		"id3v2": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V2, name: "2000"},
			want: "2000",
		},
		"unknown": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.UndefinedSource, name: "1984?"},
			want: "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.tm.CorrectAlbumYear(tt.args.src, tt.args.name)
			if got := tt.tm.AlbumYear(tt.args.src).Correction(); got != tt.want {
				t.Errorf("TrackMetadata.AlbumYear got %q want %q", got, tt.want)
			}
		})
	}
}

func TestTrackMetadata_SetTrackName(t *testing.T) {
	type args struct {
		src  files.SourceType
		name string
	}
	tests := map[string]struct {
		tm *files.TrackMetadata
		args
		want string
	}{
		"id3v1": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V1, name: "My old track"},
			want: "My old track",
		},
		"id3v2": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V2, name: "My new track"},
			want: "My new track",
		},
		"unknown": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.UndefinedSource, name: "I can has track?"},
			want: "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.tm.SetTrackName(tt.args.src, tt.args.name)
			if got := tt.tm.TrackName(tt.args.src).Original(); got != tt.want {
				t.Errorf("TrackMetadata.TrackName got %q want %q", got, tt.want)
			}
			tt.tm.SetCanonicalSource(tt.args.src)
			if got := tt.tm.CanonicalTrackName(); got != tt.want {
				t.Errorf("TrackMetadata.CanonicalTrackName got %q want %q", got, tt.want)
			}
		})
	}
}

func TestTrackMetadata_CorrectTrackName(t *testing.T) {
	type args struct {
		src  files.SourceType
		name string
	}
	tests := map[string]struct {
		tm *files.TrackMetadata
		args
		want string
	}{
		"id3v1": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V1, name: "My old track"},
			want: "My old track",
		},
		"id3v2": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V2, name: "My new track"},
			want: "My new track",
		},
		"unknown": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.UndefinedSource, name: "I can has track?"},
			want: "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.tm.CorrectTrackName(tt.args.src, tt.args.name)
			if got := tt.tm.TrackName(tt.args.src).Correction(); got != tt.want {
				t.Errorf("TrackMetadata.TrackName got %q want %q", got, tt.want)
			}
		})
	}
}

func TestTrackMetadata_SetTrackNumber(t *testing.T) {
	type args struct {
		src    files.SourceType
		number int
	}
	tests := map[string]struct {
		tm *files.TrackMetadata
		args
		want int
	}{
		"id3v1": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V1, number: 19},
			want: 19,
		},
		"id3v2": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V2, number: 20},
			want: 20,
		},
		"unknown": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.UndefinedSource, number: 45},
			want: 0,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.tm.SetTrackNumber(tt.args.src, tt.args.number)
			if got := tt.tm.TrackNumber(tt.args.src).Original(); got != tt.want {
				t.Errorf("TrackMetadata.TrackNumber got %q want %q", got, tt.want)
			}
			tt.tm.SetCanonicalSource(tt.args.src)
			if got := tt.tm.CanonicalTrackNumber(); got != tt.want {
				t.Errorf("TrackMetadata.CanonicalTrackNumber got %q want %q", got, tt.want)
			}
		})
	}
}

func TestTrackMetadata_CorrectTrackNumber(t *testing.T) {
	type args struct {
		src    files.SourceType
		number int
	}
	tests := map[string]struct {
		tm *files.TrackMetadata
		args
		want int
	}{
		"id3v1": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V1, number: 19},
			want: 19,
		},
		"id3v2": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V2, number: 20},
			want: 20,
		},
		"unknown": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.UndefinedSource, number: 45},
			want: 0,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.tm.CorrectTrackNumber(tt.args.src, tt.args.number)
			if got := tt.tm.TrackNumber(tt.args.src).Correction(); got != tt.want {
				t.Errorf("TrackMetadata.TrackNumber got %q want %q", got, tt.want)
			}
		})
	}
}

func TestTrackMetadata_SetErrorCause(t *testing.T) {
	type args struct {
		src   files.SourceType
		cause string
	}
	tests := map[string]struct {
		tm *files.TrackMetadata
		args
		want string
	}{
		"id3v1": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V1, cause: "failure to read ID3V1 data"},
			want: "failure to read ID3V1 data",
		},
		"id3v2": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.ID3V2, cause: "failure to read ID3V2 data"},
			want: "failure to read ID3V2 data",
		},
		"unknown": {
			tm:   files.NewTrackMetadata(),
			args: args{src: files.UndefinedSource, cause: "what happened?"},
			want: "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.tm.SetErrorCause(tt.args.src, tt.args.cause)
			if got := tt.tm.ErrorCause(tt.args.src); got != tt.want {
				t.Errorf("TrackMetadata.ErrorCause got %q want %q", got, tt.want)
			}
		})
	}
}

func TestTrackMetadata_SetEditRequired(t *testing.T) {
	tests := map[string]struct {
		tm   *files.TrackMetadata
		src  files.SourceType
		want bool
	}{
		"id3v1": {
			tm:   files.NewTrackMetadata(),
			src:  files.ID3V1,
			want: true,
		},
		"id3v2": {
			tm:   files.NewTrackMetadata(),
			src:  files.ID3V2,
			want: true,
		},
		"unknown": {
			tm:   files.NewTrackMetadata(),
			src:  files.UndefinedSource,
			want: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.tm.SetEditRequired(tt.src)
			if got := tt.tm.EditRequired(tt.src); got != tt.want {
				t.Errorf("TrackMetadata.EditRequired got %t want %t", got, tt.want)
			}
		})
	}
}

func TestTrackMetadata_SetCDIdentifier(t *testing.T) {
	tests := map[string]struct {
		tm   *files.TrackMetadata
		body []byte
		want []byte
	}{
		"id3v2": {
			tm:   files.NewTrackMetadata(),
			body: []byte("new"),
			want: []byte("new"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.tm.SetCDIdentifier(tt.body)
			if got := tt.tm.CDIdentifier().Original(); !reflect.DeepEqual(got.Body, tt.want) {
				t.Errorf("TrackMetadata.CDIdentifier got %v want %v", got, tt.want)
			}
			if got := tt.tm.CanonicalCDIdentifier(); !reflect.DeepEqual(got.Body, tt.want) {
				t.Errorf("TrackMetadata.CanonicalCDIdentifier got %v want %v", got, tt.want)
			}
		})
	}
}

func TestTrackMetadata_CorrectCDIdentifier(t *testing.T) {
	tests := map[string]struct {
		tm   *files.TrackMetadata
		body []byte
		want []byte
	}{
		"id3v2": {
			tm:   files.NewTrackMetadata(),
			body: []byte("old"),
			want: []byte("old"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.tm.CorrectCDIdentifier(tt.body)
			if got := tt.tm.CDIdentifier().Correction(); !reflect.DeepEqual(got.Body, tt.want) {
				t.Errorf("TrackMetadata.CDIdentifier got %v want %v", got, tt.want)
			}
		})
	}
}

func TestNewTrackMetadata(t *testing.T) {
	tests := map[string]struct {
		want *files.TrackMetadata
	}{"test": {want: files.NewTrackMetadata()}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.NewTrackMetadata(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTrackMetadata() = %v, want %v", got, tt.want)
			}
			for _, src := range []files.SourceType{files.ID3V1, files.ID3V2} {
				artistName := tt.want.ArtistName(src)
				if got := artistName.Original(); got != "" {
					t.Errorf("NewTrackMetadata().ArtistName(%s).Original() = %q, want %q", src.Name(), got, "")
				}
				if got := artistName.Correction(); got != "" {
					t.Errorf("NewTrackMetadata().ArtistName(%s).Correction() = %q, want %q", src.Name(), got, "")
				}
				albumName := tt.want.AlbumName(src)
				if got := albumName.Original(); got != "" {
					t.Errorf("NewTrackMetadata().AlbumName(%s).Original() = %q, want %q", src.Name(), got, "")
				}
				if got := albumName.Correction(); got != "" {
					t.Errorf("NewTrackMetadata().AlbumName(%s).Correction() = %q, want %q", src.Name(), got, "")
				}
				albumGenre := tt.want.AlbumGenre(src)
				if got := albumGenre.Original(); got != "" {
					t.Errorf("NewTrackMetadata().AlbumGenre(%s).Original() = %q, want %q", src.Name(), got, "")
				}
				if got := albumGenre.Correction(); got != "" {
					t.Errorf("NewTrackMetadata().AlbumGenre(%s).Correction() = %q, want %q", src.Name(), got, "")
				}
				albumYear := tt.want.AlbumYear(src)
				if got := albumYear.Original(); got != "" {
					t.Errorf("NewTrackMetadata().AlbumYear(%s).Original() = %q, want %q", src.Name(), got, "")
				}
				if got := albumYear.Correction(); got != "" {
					t.Errorf("NewTrackMetadata().AlbumYear(%s).Correction() = %q, want %q", src.Name(), got, "")
				}
				trackName := tt.want.TrackName(src)
				if got := trackName.Original(); got != "" {
					t.Errorf("NewTrackMetadata().TrackName(%s).Original() = %q, want %q", src.Name(), got, "")
				}
				if got := trackName.Correction(); got != "" {
					t.Errorf("NewTrackMetadata().TrackName(%s).Correction() = %q, want %q", src.Name(), got, "")
				}
				trackNumber := tt.want.TrackNumber(src)
				if got := trackNumber.Original(); got != 0 {
					t.Errorf("NewTrackMetadata().TrackNumber(%s).Original() = %d, want %d", src.Name(), got, 0)
				}
				if got := trackNumber.Correction(); got != 0 {
					t.Errorf("NewTrackMetadata().TrackNumber(%s).Correction() = %d, want %d", src.Name(), got, 0)
				}
				if got := tt.want.ErrorCause(src); got != "" {
					t.Errorf("NewTrackMetadata().ErrorCause(%s) = %q, want %q", src.Name(), got, "")
				}
				if got := tt.want.EditRequired(src); got != false {
					t.Errorf("NewTrackMetadata().EditRequired(%s) = %t, want %t", src.Name(), got, false)
				}
			}
			cdi := tt.want.CDIdentifier()
			if got := cdi.Original(); len(got.Body) != 0 {
				t.Errorf("NewTrackMetadata().CDIdentifier().Original() = %v, want %v", got.Body, []byte{})
			}
			if got := cdi.Correction(); len(got.Body) != 0 {
				t.Errorf("NewTrackMetadata().CDIdentifier().Correction() = %v, want %v", got.Body, []byte{})
			}
			if got := tt.want.CanonicalSource(); got != files.UndefinedSource {
				t.Errorf("NewTrackMetadata().CanonicalSource() = %s, want %s", got.Name(), files.UndefinedSource.Name())
			}
		})
	}
}

func TestInitializeMetadata(t *testing.T) {
	originalFileSystem := cmd_toolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmd_toolkit.AssignFileSystem(originalFileSystem)
	}()
	testDir := "InitializeMetadata"
	cmd_toolkit.Mkdir(testDir)
	taglessFile := "01 tagless.mp3"
	createFile(testDir, taglessFile)
	id3v1OnlyFile := "02 id3v1.mp3"
	payloadID3v1Only := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	payloadID3v1Only = append(payloadID3v1Only, id3v1DataSet1...)
	createFileWithContent(testDir, id3v1OnlyFile, payloadID3v1Only)
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
	payloadID3v2Only := createID3v2TaggedData([]byte{}, frames)
	createFileWithContent(testDir, id3v2OnlyFile, payloadID3v2Only)
	completeFile := "04 complete.mp3"
	payloadComplete := payloadID3v2Only
	payloadComplete = append(payloadComplete, payloadID3v1Only...)
	createFileWithContent(testDir, completeFile, payloadComplete)
	noSuchFile := "no such file.mp3"
	missingFileData := files.NewTrackMetadata()
	missingFileData.SetErrorCause(files.ID3V1, "open "+testDir+"\\"+noSuchFile+": file does not exist")
	missingFileData.SetErrorCause(files.ID3V2, "open "+testDir+"\\"+noSuchFile+": file does not exist")
	noMetadata := files.NewTrackMetadata()
	noMetadata.SetErrorCause(files.ID3V1, "no ID3V1 metadata found")
	noMetadata.SetErrorCause(files.ID3V2, "no ID3V2 metadata found")
	onlyID3V1Metadata := files.NewTrackMetadata()
	onlyID3V1Metadata.SetArtistName(files.ID3V1, "The Beatles")
	onlyID3V1Metadata.SetAlbumName(files.ID3V1, "On Air: Live At The BBC, Volum")
	onlyID3V1Metadata.SetAlbumGenre(files.ID3V1, "Other")
	onlyID3V1Metadata.SetAlbumYear(files.ID3V1, "2013")
	onlyID3V1Metadata.SetTrackName(files.ID3V1, "Ringo - Pop Profile [Interview")
	onlyID3V1Metadata.SetTrackNumber(files.ID3V1, 29)
	onlyID3V1Metadata.SetErrorCause(files.ID3V2, "no ID3V2 metadata found")
	onlyID3V1Metadata.SetCanonicalSource(files.ID3V1)
	onlyID3V2Metadata := files.NewTrackMetadata()
	onlyID3V2Metadata.SetArtistName(files.ID3V2, "unknown artist")
	onlyID3V2Metadata.SetAlbumName(files.ID3V2, "unknown album")
	onlyID3V2Metadata.SetAlbumGenre(files.ID3V2, "dance music")
	onlyID3V2Metadata.SetAlbumYear(files.ID3V2, "2022")
	onlyID3V2Metadata.SetTrackName(files.ID3V2, "unknown track")
	onlyID3V2Metadata.SetTrackNumber(files.ID3V2, 2)
	onlyID3V2Metadata.SetCDIdentifier([]byte{0})
	onlyID3V2Metadata.SetCanonicalSource(files.ID3V2)
	onlyID3V2Metadata.SetErrorCause(files.ID3V1, "no ID3V1 metadata found")
	allMetadata := files.NewTrackMetadata()
	allMetadata.SetArtistName(files.ID3V1, "The Beatles")
	allMetadata.SetAlbumName(files.ID3V1, "On Air: Live At The BBC, Volum")
	allMetadata.SetAlbumGenre(files.ID3V1, "Other")
	allMetadata.SetAlbumYear(files.ID3V1, "2013")
	allMetadata.SetTrackName(files.ID3V1, "Ringo - Pop Profile [Interview")
	allMetadata.SetTrackNumber(files.ID3V1, 29)
	allMetadata.SetArtistName(files.ID3V2, "unknown artist")
	allMetadata.SetAlbumName(files.ID3V2, "unknown album")
	allMetadata.SetAlbumGenre(files.ID3V2, "dance music")
	allMetadata.SetAlbumYear(files.ID3V2, "2022")
	allMetadata.SetTrackName(files.ID3V2, "unknown track")
	allMetadata.SetTrackNumber(files.ID3V2, 2)
	allMetadata.SetCDIdentifier([]byte{0})
	allMetadata.SetCanonicalSource(files.ID3V2)
	tests := map[string]struct {
		path string
		want *files.TrackMetadata
	}{
		"missing file": {
			path: filepath.Join(testDir, noSuchFile),
			want: missingFileData,
		},
		"no metadata": {
			path: filepath.Join(testDir, taglessFile),
			want: noMetadata,
		},
		"only id3v1 metadata": {
			path: filepath.Join(testDir, id3v1OnlyFile),
			want: onlyID3V1Metadata,
		},
		"only id3v2 metadata": {
			path: filepath.Join(testDir, id3v2OnlyFile),
			want: onlyID3V2Metadata,
		},
		"all metadata": {
			path: filepath.Join(testDir, completeFile),
			want: allMetadata,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.InitializeMetadata(tt.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitializeMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrackMetadata_isValid(t *testing.T) {
	ID3V1Metadata := files.NewTrackMetadata()
	ID3V1Metadata.SetCanonicalSource(files.ID3V1)
	ID3V2Metadata := files.NewTrackMetadata()
	ID3V2Metadata.SetCanonicalSource(files.ID3V2)
	tests := map[string]struct {
		tm   *files.TrackMetadata
		want bool
	}{
		"uninitialized data": {tm: files.NewTrackMetadata(), want: false},
		"ID3V1 set":          {tm: ID3V1Metadata, want: true},
		"ID3V2 set":          {tm: ID3V2Metadata, want: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tm.IsValid(); got != tt.want {
				t.Errorf("TrackMetadata.IsValid() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestTrackMetadata_ErrorCauses(t *testing.T) {
	ID3V1Metadata := files.NewTrackMetadata()
	ID3V1Metadata.SetErrorCause(files.ID3V1, "id3v1 error")
	ID3V2Metadata := files.NewTrackMetadata()
	ID3V2Metadata.SetErrorCause(files.ID3V2, "id3v2 error")
	bothMetadata := files.NewTrackMetadata()
	bothMetadata.SetErrorCause(files.ID3V1, "id3v1 error")
	bothMetadata.SetErrorCause(files.ID3V2, "id3v2 error")
	tests := map[string]struct {
		tm   *files.TrackMetadata
		want []string
	}{
		"neither":    {tm: files.NewTrackMetadata(), want: []string{}},
		"id3v1 only": {tm: ID3V1Metadata, want: []string{"id3v1 error"}},
		"id3v2 only": {tm: ID3V2Metadata, want: []string{"id3v2 error"}},
		"both":       {tm: bothMetadata, want: []string{"id3v1 error", "id3v2 error"}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tm.ErrorCauses(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TrackMetadata.ErrorCauses() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrackMetadata_TrackNumberDiffers(t *testing.T) {
	expectedTrack := 14
	// 1. neither ID3V1 nor ID3v2 have errors, and neither ID3V1 nor ID3V2 track
	//    numbers differ
	tm1 := files.NewTrackMetadata()
	tm1.SetTrackNumber(files.ID3V1, expectedTrack)
	tm1.SetTrackNumber(files.ID3V2, expectedTrack)
	// 2. neither ID3V1 nor ID3v2 have errors, and only ID3V1 track number
	//    differs
	tm2 := files.NewTrackMetadata()
	tm2.SetTrackNumber(files.ID3V1, expectedTrack+1)
	tm2.SetTrackNumber(files.ID3V2, expectedTrack)
	// 3. neither ID3V1 nor ID3v2 have errors, and only ID3V2 track number
	//    differs
	tm3 := files.NewTrackMetadata()
	tm3.SetTrackNumber(files.ID3V1, expectedTrack)
	tm3.SetTrackNumber(files.ID3V2, expectedTrack-1)
	// 4. neither ID3V1 nor ID3v2 have errors, and both track numbers differ
	tm4 := files.NewTrackMetadata()
	tm4.SetTrackNumber(files.ID3V1, expectedTrack+1)
	tm4.SetTrackNumber(files.ID3V2, expectedTrack-1)
	// 5. ID3V1 has an error, ID3V2 track number does not differ
	tm5 := files.NewTrackMetadata()
	tm5.SetErrorCause(files.ID3V1, "bad format")
	tm5.SetTrackNumber(files.ID3V1, 0)
	tm5.SetTrackNumber(files.ID3V2, expectedTrack)
	// 6. ID3V1 has an error, ID3V2 track number differs
	tm6 := files.NewTrackMetadata()
	tm6.SetErrorCause(files.ID3V1, "bad format")
	tm6.SetTrackNumber(files.ID3V1, 0)
	tm6.SetTrackNumber(files.ID3V2, expectedTrack+1)
	// 7. ID3V2 has an error, ID3V1 track number does not differ
	tm7 := files.NewTrackMetadata()
	tm7.SetErrorCause(files.ID3V2, "bad format")
	tm7.SetTrackNumber(files.ID3V1, expectedTrack)
	tm7.SetTrackNumber(files.ID3V2, 0)
	// 8. ID3V2 has an error, ID3V1 track number differs
	tm8 := files.NewTrackMetadata()
	tm8.SetErrorCause(files.ID3V2, "bad format")
	tm8.SetTrackNumber(files.ID3V1, expectedTrack+1)
	tm8.SetTrackNumber(files.ID3V2, 0)
	// 9. both ID3V1 and ID3V2 have errors
	tm9 := files.NewTrackMetadata()
	tm9.SetErrorCause(files.ID3V1, "bad format")
	tm9.SetErrorCause(files.ID3V2, "bad format")
	tm9.SetTrackNumber(files.ID3V1, 0)
	tm9.SetTrackNumber(files.ID3V2, 0)
	tests := map[string]struct {
		tm                            *files.TrackMetadata
		trackNumberFromFileName       int
		wantDiffers                   bool
		wantCorrectedID3V1TrackNumber int
		wantCorrectedID3V2TrackNumber int
		wantID3V1EditRequired         bool
		wantID3V2EditRequired         bool
	}{
		"no errors, no differences": {
			tm:                            tm1,
			trackNumberFromFileName:       expectedTrack,
			wantDiffers:                   false,
			wantCorrectedID3V1TrackNumber: 0,
			wantCorrectedID3V2TrackNumber: 0,
			wantID3V1EditRequired:         false,
			wantID3V2EditRequired:         false,
		},
		"no errors, ID3V1 differs": {
			tm:                            tm2,
			trackNumberFromFileName:       expectedTrack,
			wantDiffers:                   true,
			wantCorrectedID3V1TrackNumber: expectedTrack,
			wantCorrectedID3V2TrackNumber: 0,
			wantID3V1EditRequired:         true,
			wantID3V2EditRequired:         false,
		},
		"no errors, ID3V2 differs": {
			tm:                            tm3,
			trackNumberFromFileName:       expectedTrack,
			wantDiffers:                   true,
			wantCorrectedID3V1TrackNumber: 0,
			wantCorrectedID3V2TrackNumber: expectedTrack,
			wantID3V1EditRequired:         false,
			wantID3V2EditRequired:         true,
		},
		"no errors, both differs": {
			tm:                            tm4,
			trackNumberFromFileName:       expectedTrack,
			wantDiffers:                   true,
			wantCorrectedID3V1TrackNumber: expectedTrack,
			wantCorrectedID3V2TrackNumber: expectedTrack,
			wantID3V1EditRequired:         true,
			wantID3V2EditRequired:         true,
		},
		"ID3V1 error, no differences": {
			tm:                            tm5,
			trackNumberFromFileName:       expectedTrack,
			wantDiffers:                   false,
			wantCorrectedID3V1TrackNumber: 0,
			wantCorrectedID3V2TrackNumber: 0,
			wantID3V1EditRequired:         false,
			wantID3V2EditRequired:         false,
		},
		"ID3V1 error, ID3V2 differs": {
			tm:                            tm6,
			trackNumberFromFileName:       expectedTrack,
			wantDiffers:                   true,
			wantCorrectedID3V1TrackNumber: 0,
			wantCorrectedID3V2TrackNumber: expectedTrack,
			wantID3V1EditRequired:         false,
			wantID3V2EditRequired:         true,
		},
		"ID3V2 error, no differences": {
			tm:                            tm7,
			trackNumberFromFileName:       expectedTrack,
			wantDiffers:                   false,
			wantCorrectedID3V1TrackNumber: 0,
			wantCorrectedID3V2TrackNumber: 0,
			wantID3V1EditRequired:         false,
			wantID3V2EditRequired:         false,
		},
		"ID3V2 error, ID3V1 differs": {
			tm:                            tm8,
			trackNumberFromFileName:       expectedTrack,
			wantDiffers:                   true,
			wantCorrectedID3V1TrackNumber: expectedTrack,
			wantCorrectedID3V2TrackNumber: 0,
			wantID3V1EditRequired:         true,
			wantID3V2EditRequired:         false,
		},
		"both errors": {
			tm:                            tm9,
			trackNumberFromFileName:       expectedTrack,
			wantDiffers:                   false,
			wantCorrectedID3V1TrackNumber: 0,
			wantCorrectedID3V2TrackNumber: 0,
			wantID3V1EditRequired:         false,
			wantID3V2EditRequired:         false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tm.TrackNumberDiffers(tt.trackNumberFromFileName); got != tt.wantDiffers {
				t.Errorf("TrackMetadata.TrackNumberDiffers() = %t, want %t", got, tt.wantDiffers)
			}
			if got := tt.tm.TrackNumber(files.ID3V1).Correction(); got != tt.wantCorrectedID3V1TrackNumber {
				t.Errorf("corrected ID3V1 track number = %d, want %d", got, tt.wantCorrectedID3V1TrackNumber)
			}
			if got := tt.tm.TrackNumber(files.ID3V2).Correction(); got != tt.wantCorrectedID3V2TrackNumber {
				t.Errorf("corrected ID3V2 track number = %d, want %d", got, tt.wantCorrectedID3V2TrackNumber)
			}
			if got := tt.tm.EditRequired(files.ID3V1); got != tt.wantID3V1EditRequired {
				t.Errorf("ID3V1 edit required = %t, want %t", got, tt.wantID3V1EditRequired)
			}
			if got := tt.tm.EditRequired(files.ID3V2); got != tt.wantID3V2EditRequired {
				t.Errorf("ID3V2 edit required = %t, want %t", got, tt.wantID3V2EditRequired)
			}
		})
	}
}

func TestTrackMetadata_TrackNameDiffers(t *testing.T) {
	expectedName := "my fine track"
	// 1. neither ID3V1 nor ID3v2 have errors, and neither ID3V1 nor ID3V2 track
	//    names differ
	tm1 := files.NewTrackMetadata()
	tm1.SetTrackName(files.ID3V1, expectedName)
	tm1.SetTrackName(files.ID3V2, expectedName)
	// 2. neither ID3V1 nor ID3v2 have errors, and only ID3V1 track name differs
	tm2 := files.NewTrackMetadata()
	tm2.SetTrackName(files.ID3V1, expectedName+"1")
	tm2.SetTrackName(files.ID3V2, expectedName)
	// 3. neither ID3V1 nor ID3v2 have errors, and only ID3V2 track name differs
	tm3 := files.NewTrackMetadata()
	tm3.SetTrackName(files.ID3V1, expectedName)
	tm3.SetTrackName(files.ID3V2, expectedName+"2")
	// 4. neither ID3V1 nor ID3v2 have errors, and both track names differ
	tm4 := files.NewTrackMetadata()
	tm4.SetTrackName(files.ID3V1, expectedName+"1")
	tm4.SetTrackName(files.ID3V2, expectedName+"2")
	// 5. ID3V1 has an error, ID3V2 track name does not differ
	tm5 := files.NewTrackMetadata()
	tm5.SetErrorCause(files.ID3V1, "bad format")
	tm5.SetTrackName(files.ID3V1, "")
	tm5.SetTrackName(files.ID3V2, expectedName)
	// 6. ID3V1 has an error, ID3V2 track name differs
	tm6 := files.NewTrackMetadata()
	tm6.SetErrorCause(files.ID3V1, "bad format")
	tm6.SetTrackName(files.ID3V1, "")
	tm6.SetTrackName(files.ID3V2, expectedName+"1")
	// 7. ID3V2 has an error, ID3V1 track name does not differ
	tm7 := files.NewTrackMetadata()
	tm7.SetErrorCause(files.ID3V2, "bad format")
	tm7.SetTrackName(files.ID3V1, expectedName)
	tm7.SetTrackName(files.ID3V2, "")
	// 8. ID3V2 has an error, ID3V1 track number differs
	tm8 := files.NewTrackMetadata()
	tm8.SetErrorCause(files.ID3V2, "bad format")
	tm8.SetTrackName(files.ID3V1, expectedName+"1")
	tm8.SetTrackName(files.ID3V2, "")
	// 9. both ID3V1 and ID3V2 have errors
	tm9 := files.NewTrackMetadata()
	tm9.SetErrorCause(files.ID3V1, "bad format")
	tm9.SetErrorCause(files.ID3V2, "bad format")
	tm9.SetTrackName(files.ID3V1, "")
	tm9.SetTrackName(files.ID3V2, "")
	tests := map[string]struct {
		tm                          *files.TrackMetadata
		nameFromFile                string
		wantDiffers                 bool
		wantCorrectedID3V1TrackName string
		wantCorrectedID3V2TrackName string
		wantID3V1EditRequired       bool
		wantID3V2EditRequired       bool
	}{
		"no errors, no differences": {
			tm:                          tm1,
			nameFromFile:                expectedName,
			wantDiffers:                 false,
			wantCorrectedID3V1TrackName: "",
			wantCorrectedID3V2TrackName: "",
			wantID3V1EditRequired:       false,
			wantID3V2EditRequired:       false,
		},
		"no errors, ID3V1 differs": {
			tm:                          tm2,
			nameFromFile:                expectedName,
			wantDiffers:                 true,
			wantCorrectedID3V1TrackName: expectedName,
			wantCorrectedID3V2TrackName: "",
			wantID3V1EditRequired:       true,
			wantID3V2EditRequired:       false,
		},
		"no errors, ID3V2 differs": {
			tm:                          tm3,
			nameFromFile:                expectedName,
			wantDiffers:                 true,
			wantCorrectedID3V1TrackName: "",
			wantCorrectedID3V2TrackName: expectedName,
			wantID3V1EditRequired:       false,
			wantID3V2EditRequired:       true,
		},
		"no errors, both differs": {
			tm:                          tm4,
			nameFromFile:                expectedName,
			wantDiffers:                 true,
			wantCorrectedID3V1TrackName: expectedName,
			wantCorrectedID3V2TrackName: expectedName,
			wantID3V1EditRequired:       true,
			wantID3V2EditRequired:       true,
		},
		"ID3V1 error, no differences": {
			tm:                          tm5,
			nameFromFile:                expectedName,
			wantDiffers:                 false,
			wantCorrectedID3V1TrackName: "",
			wantCorrectedID3V2TrackName: "",
			wantID3V1EditRequired:       false,
			wantID3V2EditRequired:       false,
		},
		"ID3V1 error, ID3V2 differs": {
			tm:                          tm6,
			nameFromFile:                expectedName,
			wantDiffers:                 true,
			wantCorrectedID3V1TrackName: "",
			wantCorrectedID3V2TrackName: expectedName,
			wantID3V1EditRequired:       false,
			wantID3V2EditRequired:       true,
		},
		"ID3V2 error, no differences": {
			tm:                          tm7,
			nameFromFile:                expectedName,
			wantDiffers:                 false,
			wantCorrectedID3V1TrackName: "",
			wantCorrectedID3V2TrackName: "",
			wantID3V1EditRequired:       false,
			wantID3V2EditRequired:       false,
		},
		"ID3V2 error, ID3V1 differs": {
			tm:                          tm8,
			nameFromFile:                expectedName,
			wantDiffers:                 true,
			wantCorrectedID3V1TrackName: expectedName,
			wantCorrectedID3V2TrackName: "",
			wantID3V1EditRequired:       true,
			wantID3V2EditRequired:       false,
		},
		"both errors": {
			tm:                          tm9,
			nameFromFile:                expectedName,
			wantDiffers:                 false,
			wantCorrectedID3V1TrackName: "",
			wantCorrectedID3V2TrackName: "",
			wantID3V1EditRequired:       false,
			wantID3V2EditRequired:       false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tm.TrackNameDiffers(tt.nameFromFile); got != tt.wantDiffers {
				t.Errorf("TrackMetadata.TrackNameDiffers() = %v, want %v", got, tt.wantDiffers)
			}
			if got := tt.tm.TrackName(files.ID3V1).Correction(); got != tt.wantCorrectedID3V1TrackName {
				t.Errorf("corrected ID3V1 track name = %q, want %q", got, tt.wantCorrectedID3V1TrackName)
			}
			if got := tt.tm.TrackName(files.ID3V2).Correction(); got != tt.wantCorrectedID3V2TrackName {
				t.Errorf("corrected ID3V2 track name = %q, want %q", got, tt.wantCorrectedID3V2TrackName)
			}
			if got := tt.tm.EditRequired(files.ID3V1); got != tt.wantID3V1EditRequired {
				t.Errorf("ID3V1 edit required = %t, want %t", got, tt.wantID3V1EditRequired)
			}
			if got := tt.tm.EditRequired(files.ID3V2); got != tt.wantID3V2EditRequired {
				t.Errorf("ID3V2 edit required = %t, want %t", got, tt.wantID3V2EditRequired)
			}
		})
	}
}

func TestTrackMetadata_AlbumNameDiffers(t *testing.T) {
	expectedName := "my fine album"
	// 1. neither ID3V1 nor ID3v2 have errors, and neither ID3V1 nor ID3V2 album
	//    names differ
	tm1 := files.NewTrackMetadata()
	tm1.SetAlbumName(files.ID3V1, expectedName)
	tm1.SetAlbumName(files.ID3V2, expectedName)
	// 2. neither ID3V1 nor ID3v2 have errors, and only ID3V1 album name differs
	tm2 := files.NewTrackMetadata()
	tm2.SetAlbumName(files.ID3V1, expectedName+"1")
	tm2.SetAlbumName(files.ID3V2, expectedName)
	// 3. neither ID3V1 nor ID3v2 have errors, and only ID3V2 album name differs
	tm3 := files.NewTrackMetadata()
	tm3.SetAlbumName(files.ID3V1, expectedName)
	tm3.SetAlbumName(files.ID3V2, expectedName+"2")
	// 4. neither ID3V1 nor ID3v2 have errors, and both album names differ
	tm4 := files.NewTrackMetadata()
	tm4.SetAlbumName(files.ID3V1, expectedName+"1")
	tm4.SetAlbumName(files.ID3V2, expectedName+"2")
	// 5. ID3V1 has an error, ID3V2 album name does not differ
	tm5 := files.NewTrackMetadata()
	tm5.SetErrorCause(files.ID3V1, "bad format")
	tm5.SetAlbumName(files.ID3V1, "")
	tm5.SetAlbumName(files.ID3V2, expectedName)
	// 6. ID3V1 has an error, ID3V2 album name differs
	tm6 := files.NewTrackMetadata()
	tm6.SetErrorCause(files.ID3V1, "bad format")
	tm6.SetAlbumName(files.ID3V1, "")
	tm6.SetAlbumName(files.ID3V2, expectedName+"1")
	// 7. ID3V2 has an error, ID3V1 album name does not differ
	tm7 := files.NewTrackMetadata()
	tm7.SetErrorCause(files.ID3V2, "bad format")
	tm7.SetAlbumName(files.ID3V1, expectedName)
	tm7.SetAlbumName(files.ID3V2, "")
	// 8. ID3V2 has an error, ID3V1 album number differs
	tm8 := files.NewTrackMetadata()
	tm8.SetErrorCause(files.ID3V2, "bad format")
	tm8.SetAlbumName(files.ID3V1, expectedName+"1")
	tm8.SetAlbumName(files.ID3V2, "")
	// 9. both ID3V1 and ID3V2 have errors
	tm9 := files.NewTrackMetadata()
	tm9.SetErrorCause(files.ID3V1, "bad format")
	tm9.SetErrorCause(files.ID3V2, "bad format")
	tm9.SetAlbumName(files.ID3V1, "")
	tm9.SetAlbumName(files.ID3V2, "")
	tests := map[string]struct {
		tm                          *files.TrackMetadata
		nameFromFile                string
		wantDiffers                 bool
		wantCorrectedID3V1AlbumName string
		wantCorrectedID3V2AlbumName string
		wantID3V1EditRequired       bool
		wantID3V2EditRequired       bool
	}{
		"no errors, no differences": {
			tm:                          tm1,
			nameFromFile:                expectedName,
			wantDiffers:                 false,
			wantCorrectedID3V1AlbumName: "",
			wantCorrectedID3V2AlbumName: "",
			wantID3V1EditRequired:       false,
			wantID3V2EditRequired:       false,
		},
		"no errors, ID3V1 differs": {
			tm:                          tm2,
			nameFromFile:                expectedName,
			wantDiffers:                 true,
			wantCorrectedID3V1AlbumName: expectedName,
			wantCorrectedID3V2AlbumName: "",
			wantID3V1EditRequired:       true,
			wantID3V2EditRequired:       false,
		},
		"no errors, ID3V2 differs": {
			tm:                          tm3,
			nameFromFile:                expectedName,
			wantDiffers:                 true,
			wantCorrectedID3V1AlbumName: "",
			wantCorrectedID3V2AlbumName: expectedName,
			wantID3V1EditRequired:       false,
			wantID3V2EditRequired:       true,
		},
		"no errors, both differs": {
			tm:                          tm4,
			nameFromFile:                expectedName,
			wantDiffers:                 true,
			wantCorrectedID3V1AlbumName: expectedName,
			wantCorrectedID3V2AlbumName: expectedName,
			wantID3V1EditRequired:       true,
			wantID3V2EditRequired:       true,
		},
		"ID3V1 error, no differences": {
			tm:                          tm5,
			nameFromFile:                expectedName,
			wantDiffers:                 false,
			wantCorrectedID3V1AlbumName: "",
			wantCorrectedID3V2AlbumName: "",
			wantID3V1EditRequired:       false,
			wantID3V2EditRequired:       false,
		},
		"ID3V1 error, ID3V2 differs": {
			tm:                          tm6,
			nameFromFile:                expectedName,
			wantDiffers:                 true,
			wantCorrectedID3V1AlbumName: "",
			wantCorrectedID3V2AlbumName: expectedName,
			wantID3V1EditRequired:       false,
			wantID3V2EditRequired:       true,
		},
		"ID3V2 error, no differences": {
			tm:                          tm7,
			nameFromFile:                expectedName,
			wantDiffers:                 false,
			wantCorrectedID3V1AlbumName: "",
			wantCorrectedID3V2AlbumName: "",
			wantID3V1EditRequired:       false,
			wantID3V2EditRequired:       false,
		},
		"ID3V2 error, ID3V1 differs": {
			tm:                          tm8,
			nameFromFile:                expectedName,
			wantDiffers:                 true,
			wantCorrectedID3V1AlbumName: expectedName,
			wantCorrectedID3V2AlbumName: "",
			wantID3V1EditRequired:       true,
			wantID3V2EditRequired:       false,
		},
		"both errors": {
			tm:                          tm9,
			nameFromFile:                expectedName,
			wantDiffers:                 false,
			wantCorrectedID3V1AlbumName: "",
			wantCorrectedID3V2AlbumName: "",
			wantID3V1EditRequired:       false,
			wantID3V2EditRequired:       false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tm.AlbumNameDiffers(tt.nameFromFile); got != tt.wantDiffers {
				t.Errorf("TrackMetadata.AlbumNameDiffers() = %v, want %v", got, tt.wantDiffers)
			}
			if got := tt.tm.AlbumName(files.ID3V1).Correction(); got != tt.wantCorrectedID3V1AlbumName {
				t.Errorf("corrected ID3V1 album name = %q, want %q", got, tt.wantCorrectedID3V1AlbumName)
			}
			if got := tt.tm.AlbumName(files.ID3V2).Correction(); got != tt.wantCorrectedID3V2AlbumName {
				t.Errorf("corrected ID3V2 album name = %q, want %q", got, tt.wantCorrectedID3V2AlbumName)
			}
			if got := tt.tm.EditRequired(files.ID3V1); got != tt.wantID3V1EditRequired {
				t.Errorf("ID3V1 edit required = %t, want %t", got, tt.wantID3V1EditRequired)
			}
			if got := tt.tm.EditRequired(files.ID3V2); got != tt.wantID3V2EditRequired {
				t.Errorf("ID3V2 edit required = %t, want %t", got, tt.wantID3V2EditRequired)
			}
		})
	}
}

func TestTrackMetadata_ArtistNameDiffers(t *testing.T) {
	expectedName := "my fine artist"
	// 1. neither ID3V1 nor ID3v2 have errors, and neither ID3V1 nor ID3V2
	//    artist names differ
	tm1 := files.NewTrackMetadata()
	tm1.SetArtistName(files.ID3V1, expectedName)
	tm1.SetArtistName(files.ID3V2, expectedName)
	// 2. neither ID3V1 nor ID3v2 have errors, and only ID3V1 artist name
	//    differs
	tm2 := files.NewTrackMetadata()
	tm2.SetArtistName(files.ID3V1, expectedName+"1")
	tm2.SetArtistName(files.ID3V2, expectedName)
	// 3. neither ID3V1 nor ID3v2 have errors, and only ID3V2 artist name
	//    differs
	tm3 := files.NewTrackMetadata()
	tm3.SetArtistName(files.ID3V1, expectedName)
	tm3.SetArtistName(files.ID3V2, expectedName+"2")
	// 4. neither ID3V1 nor ID3v2 have errors, and both artist names differ
	tm4 := files.NewTrackMetadata()
	tm4.SetArtistName(files.ID3V1, expectedName+"1")
	tm4.SetArtistName(files.ID3V2, expectedName+"2")
	// 5. ID3V1 has an error, ID3V2 artist name does not differ
	tm5 := files.NewTrackMetadata()
	tm5.SetErrorCause(files.ID3V1, "bad format")
	tm5.SetArtistName(files.ID3V1, "")
	tm5.SetArtistName(files.ID3V2, expectedName)
	// 6. ID3V1 has an error, ID3V2 artist name differs
	tm6 := files.NewTrackMetadata()
	tm6.SetErrorCause(files.ID3V1, "bad format")
	tm6.SetArtistName(files.ID3V1, "")
	tm6.SetArtistName(files.ID3V2, expectedName+"1")
	// 7. ID3V2 has an error, ID3V1 artist name does not differ
	tm7 := files.NewTrackMetadata()
	tm7.SetErrorCause(files.ID3V2, "bad format")
	tm7.SetArtistName(files.ID3V1, expectedName)
	tm7.SetArtistName(files.ID3V2, "")
	// 8. ID3V2 has an error, ID3V1 artist number differs
	tm8 := files.NewTrackMetadata()
	tm8.SetErrorCause(files.ID3V2, "bad format")
	tm8.SetArtistName(files.ID3V1, expectedName+"1")
	tm8.SetArtistName(files.ID3V2, "")
	// 9. both ID3V1 and ID3V2 have errors
	tm9 := files.NewTrackMetadata()
	tm9.SetErrorCause(files.ID3V1, "bad format")
	tm9.SetErrorCause(files.ID3V2, "bad format")
	tm9.SetArtistName(files.ID3V1, "")
	tm9.SetArtistName(files.ID3V2, "")
	tests := map[string]struct {
		tm                           *files.TrackMetadata
		nameFromFile                 string
		wantDiffers                  bool
		wantCorrectedID3V1ArtistName string
		wantCorrectedID3V2ArtistName string
		wantID3V1EditRequired        bool
		wantID3V2EditRequired        bool
	}{
		"no errors, no differences": {
			tm:                           tm1,
			nameFromFile:                 expectedName,
			wantDiffers:                  false,
			wantCorrectedID3V1ArtistName: "",
			wantCorrectedID3V2ArtistName: "",
			wantID3V1EditRequired:        false,
			wantID3V2EditRequired:        false,
		},
		"no errors, ID3V1 differs": {
			tm:                           tm2,
			nameFromFile:                 expectedName,
			wantDiffers:                  true,
			wantCorrectedID3V1ArtistName: expectedName,
			wantCorrectedID3V2ArtistName: "",
			wantID3V1EditRequired:        true,
			wantID3V2EditRequired:        false,
		},
		"no errors, ID3V2 differs": {
			tm:                           tm3,
			nameFromFile:                 expectedName,
			wantDiffers:                  true,
			wantCorrectedID3V1ArtistName: "",
			wantCorrectedID3V2ArtistName: expectedName,
			wantID3V1EditRequired:        false,
			wantID3V2EditRequired:        true,
		},
		"no errors, both differs": {
			tm:                           tm4,
			nameFromFile:                 expectedName,
			wantDiffers:                  true,
			wantCorrectedID3V1ArtistName: expectedName,
			wantCorrectedID3V2ArtistName: expectedName,
			wantID3V1EditRequired:        true,
			wantID3V2EditRequired:        true,
		},
		"ID3V1 error, no differences": {
			tm:                           tm5,
			nameFromFile:                 expectedName,
			wantDiffers:                  false,
			wantCorrectedID3V1ArtistName: "",
			wantCorrectedID3V2ArtistName: "",
			wantID3V1EditRequired:        false,
			wantID3V2EditRequired:        false,
		},
		"ID3V1 error, ID3V2 differs": {
			tm:                           tm6,
			nameFromFile:                 expectedName,
			wantDiffers:                  true,
			wantCorrectedID3V1ArtistName: "",
			wantCorrectedID3V2ArtistName: expectedName,
			wantID3V1EditRequired:        false,
			wantID3V2EditRequired:        true,
		},
		"ID3V2 error, no differences": {
			tm:                           tm7,
			nameFromFile:                 expectedName,
			wantDiffers:                  false,
			wantCorrectedID3V1ArtistName: "",
			wantCorrectedID3V2ArtistName: "",
			wantID3V1EditRequired:        false,
			wantID3V2EditRequired:        false,
		},
		"ID3V2 error, ID3V1 differs": {
			tm:                           tm8,
			nameFromFile:                 expectedName,
			wantDiffers:                  true,
			wantCorrectedID3V1ArtistName: expectedName,
			wantCorrectedID3V2ArtistName: "",
			wantID3V1EditRequired:        true,
			wantID3V2EditRequired:        false,
		},
		"both errors": {
			tm:                           tm9,
			nameFromFile:                 expectedName,
			wantDiffers:                  false,
			wantCorrectedID3V1ArtistName: "",
			wantCorrectedID3V2ArtistName: "",
			wantID3V1EditRequired:        false,
			wantID3V2EditRequired:        false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tm.ArtistNameDiffers(tt.nameFromFile); got != tt.wantDiffers {
				t.Errorf("TrackMetadata.ArtistNameDiffers() = %v, want %v", got, tt.wantDiffers)
			}
			if got := tt.tm.ArtistName(files.ID3V1).Correction(); got != tt.wantCorrectedID3V1ArtistName {
				t.Errorf("corrected ID3V1 artist name = %q, want %q", got, tt.wantCorrectedID3V1ArtistName)
			}
			if got := tt.tm.ArtistName(files.ID3V2).Correction(); got != tt.wantCorrectedID3V2ArtistName {
				t.Errorf("corrected ID3V2 artist name = %q, want %q", got, tt.wantCorrectedID3V2ArtistName)
			}
			if got := tt.tm.EditRequired(files.ID3V1); got != tt.wantID3V1EditRequired {
				t.Errorf("ID3V1 edit required = %t, want %t", got, tt.wantID3V1EditRequired)
			}
			if got := tt.tm.EditRequired(files.ID3V2); got != tt.wantID3V2EditRequired {
				t.Errorf("ID3V2 edit required = %t, want %t", got, tt.wantID3V2EditRequired)
			}
		})
	}
}

func TestTrackMetadata_AlbumGenreDiffers(t *testing.T) {
	expectedGenre := "rock"
	// 1. neither ID3V1 nor ID3v2 have errors, and neither ID3V1 nor ID3V2 album
	//    genres differ
	tm1 := files.NewTrackMetadata()
	tm1.SetAlbumGenre(files.ID3V1, expectedGenre)
	tm1.SetAlbumGenre(files.ID3V2, expectedGenre)
	// 2. neither ID3V1 nor ID3v2 have errors, and only ID3V1 album genre
	//    differs
	tm2 := files.NewTrackMetadata()
	tm2.SetAlbumGenre(files.ID3V1, "country")
	tm2.SetAlbumGenre(files.ID3V2, expectedGenre)
	// 3. neither ID3V1 nor ID3v2 have errors, and only ID3V2 album genre
	//    differs
	tm3 := files.NewTrackMetadata()
	tm3.SetAlbumGenre(files.ID3V1, expectedGenre)
	tm3.SetAlbumGenre(files.ID3V2, "rap")
	// 4. neither ID3V1 nor ID3v2 have errors, and both album genres differ
	tm4 := files.NewTrackMetadata()
	tm4.SetAlbumGenre(files.ID3V1, "country")
	tm4.SetAlbumGenre(files.ID3V2, "rap")
	// 5. ID3V1 has an error, ID3V2 album genre does not differ
	tm5 := files.NewTrackMetadata()
	tm5.SetErrorCause(files.ID3V1, "bad format")
	tm5.SetAlbumGenre(files.ID3V1, "")
	tm5.SetAlbumGenre(files.ID3V2, expectedGenre)
	// 6. ID3V1 has an error, ID3V2 album genre differs
	tm6 := files.NewTrackMetadata()
	tm6.SetErrorCause(files.ID3V1, "bad format")
	tm6.SetAlbumGenre(files.ID3V1, "")
	tm6.SetAlbumGenre(files.ID3V2, "country")
	// 7. ID3V2 has an error, ID3V1 album genre does not differ
	tm7 := files.NewTrackMetadata()
	tm7.SetErrorCause(files.ID3V2, "bad format")
	tm7.SetAlbumGenre(files.ID3V1, expectedGenre)
	tm7.SetAlbumGenre(files.ID3V2, "")
	// 8. ID3V2 has an error, ID3V1 album number differs
	tm8 := files.NewTrackMetadata()
	tm8.SetErrorCause(files.ID3V2, "bad format")
	tm8.SetAlbumGenre(files.ID3V1, "country")
	tm8.SetAlbumGenre(files.ID3V2, "")
	// 9. both ID3V1 and ID3V2 have errors
	tm9 := files.NewTrackMetadata()
	tm9.SetErrorCause(files.ID3V1, "bad format")
	tm9.SetErrorCause(files.ID3V2, "bad format")
	tm9.SetAlbumGenre(files.ID3V1, "")
	tm9.SetAlbumGenre(files.ID3V2, "")
	tests := map[string]struct {
		tm                           *files.TrackMetadata
		canonicalAlbumGenre          string
		wantDiffers                  bool
		wantCorrectedID3V1AlbumGenre string
		wantCorrectedID3V2AlbumGenre string
		wantID3V1EditRequired        bool
		wantID3V2EditRequired        bool
	}{
		"no errors, no differences": {
			tm:                           tm1,
			canonicalAlbumGenre:          expectedGenre,
			wantDiffers:                  false,
			wantCorrectedID3V1AlbumGenre: "",
			wantCorrectedID3V2AlbumGenre: "",
			wantID3V1EditRequired:        false,
			wantID3V2EditRequired:        false,
		},
		"no errors, ID3V1 differs": {
			tm:                           tm2,
			canonicalAlbumGenre:          expectedGenre,
			wantDiffers:                  true,
			wantCorrectedID3V1AlbumGenre: expectedGenre,
			wantCorrectedID3V2AlbumGenre: "",
			wantID3V1EditRequired:        true,
			wantID3V2EditRequired:        false,
		},
		"no errors, ID3V2 differs": {
			tm:                           tm3,
			canonicalAlbumGenre:          expectedGenre,
			wantDiffers:                  true,
			wantCorrectedID3V1AlbumGenre: "",
			wantCorrectedID3V2AlbumGenre: expectedGenre,
			wantID3V1EditRequired:        false,
			wantID3V2EditRequired:        true,
		},
		"no errors, both differs": {
			tm:                           tm4,
			canonicalAlbumGenre:          expectedGenre,
			wantDiffers:                  true,
			wantCorrectedID3V1AlbumGenre: expectedGenre,
			wantCorrectedID3V2AlbumGenre: expectedGenre,
			wantID3V1EditRequired:        true,
			wantID3V2EditRequired:        true,
		},
		"ID3V1 error, no differences": {
			tm:                           tm5,
			canonicalAlbumGenre:          expectedGenre,
			wantDiffers:                  false,
			wantCorrectedID3V1AlbumGenre: "",
			wantCorrectedID3V2AlbumGenre: "",
			wantID3V1EditRequired:        false,
			wantID3V2EditRequired:        false,
		},
		"ID3V1 error, ID3V2 differs": {
			tm:                           tm6,
			canonicalAlbumGenre:          expectedGenre,
			wantDiffers:                  true,
			wantCorrectedID3V1AlbumGenre: "",
			wantCorrectedID3V2AlbumGenre: expectedGenre,
			wantID3V1EditRequired:        false,
			wantID3V2EditRequired:        true,
		},
		"ID3V2 error, no differences": {
			tm:                           tm7,
			canonicalAlbumGenre:          expectedGenre,
			wantDiffers:                  false,
			wantCorrectedID3V1AlbumGenre: "",
			wantCorrectedID3V2AlbumGenre: "",
			wantID3V1EditRequired:        false,
			wantID3V2EditRequired:        false,
		},
		"ID3V2 error, ID3V1 differs": {
			tm:                           tm8,
			canonicalAlbumGenre:          expectedGenre,
			wantDiffers:                  true,
			wantCorrectedID3V1AlbumGenre: expectedGenre,
			wantCorrectedID3V2AlbumGenre: "",
			wantID3V1EditRequired:        true,
			wantID3V2EditRequired:        false,
		},
		"both errors": {
			tm:                           tm9,
			canonicalAlbumGenre:          expectedGenre,
			wantDiffers:                  false,
			wantCorrectedID3V1AlbumGenre: "",
			wantCorrectedID3V2AlbumGenre: "",
			wantID3V1EditRequired:        false,
			wantID3V2EditRequired:        false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tm.AlbumGenreDiffers(tt.canonicalAlbumGenre); got != tt.wantDiffers {
				t.Errorf("TrackMetadata.AlbumGenreDiffers() = %v, want %v", got, tt.wantDiffers)
			}
			if got := tt.tm.AlbumGenre(files.ID3V1).Correction(); got != tt.wantCorrectedID3V1AlbumGenre {
				t.Errorf("corrected ID3V1 album genre = %q, want %q", got, tt.wantCorrectedID3V1AlbumGenre)
			}
			if got := tt.tm.AlbumGenre(files.ID3V2).Correction(); got != tt.wantCorrectedID3V2AlbumGenre {
				t.Errorf("corrected ID3V2 album genre = %q, want %q", got, tt.wantCorrectedID3V2AlbumGenre)
			}
			if got := tt.tm.EditRequired(files.ID3V1); got != tt.wantID3V1EditRequired {
				t.Errorf("ID3V1 edit required = %t, want %t", got, tt.wantID3V1EditRequired)
			}
			if got := tt.tm.EditRequired(files.ID3V2); got != tt.wantID3V2EditRequired {
				t.Errorf("ID3V2 edit required = %t, want %t", got, tt.wantID3V2EditRequired)
			}
		})
	}
}

func TestTrackMetadata_AlbumYearDiffers(t *testing.T) {
	expectedYear := "1999"
	// 1. neither ID3V1 nor ID3v2 have errors, and neither ID3V1 nor ID3V2 album
	//    years differ
	tm1 := files.NewTrackMetadata()
	tm1.SetAlbumYear(files.ID3V1, expectedYear)
	tm1.SetAlbumYear(files.ID3V2, expectedYear)
	// 2. neither ID3V1 nor ID3v2 have errors, and only ID3V1 album year differs
	tm2 := files.NewTrackMetadata()
	tm2.SetAlbumYear(files.ID3V1, "1984")
	tm2.SetAlbumYear(files.ID3V2, expectedYear)
	// 3. neither ID3V1 nor ID3v2 have errors, and only ID3V2 album year differs
	tm3 := files.NewTrackMetadata()
	tm3.SetAlbumYear(files.ID3V1, expectedYear)
	tm3.SetAlbumYear(files.ID3V2, "2001")
	// 4. neither ID3V1 nor ID3v2 have errors, and both album years differ
	tm4 := files.NewTrackMetadata()
	tm4.SetAlbumYear(files.ID3V1, "1984")
	tm4.SetAlbumYear(files.ID3V2, "2001")
	// 5. ID3V1 has an error, ID3V2 album year does not differ
	tm5 := files.NewTrackMetadata()
	tm5.SetErrorCause(files.ID3V1, "bad format")
	tm5.SetAlbumYear(files.ID3V1, "")
	tm5.SetAlbumYear(files.ID3V2, expectedYear)
	// 6. ID3V1 has an error, ID3V2 album year differs
	tm6 := files.NewTrackMetadata()
	tm6.SetErrorCause(files.ID3V1, "bad format")
	tm6.SetAlbumYear(files.ID3V1, "")
	tm6.SetAlbumYear(files.ID3V2, "1984")
	// 7. ID3V2 has an error, ID3V1 album year does not differ
	tm7 := files.NewTrackMetadata()
	tm7.SetErrorCause(files.ID3V2, "bad format")
	tm7.SetAlbumYear(files.ID3V1, expectedYear)
	tm7.SetAlbumYear(files.ID3V2, "")
	// 8. ID3V2 has an error, ID3V1 album number differs
	tm8 := files.NewTrackMetadata()
	tm8.SetErrorCause(files.ID3V2, "bad format")
	tm8.SetAlbumYear(files.ID3V1, "1984")
	tm8.SetAlbumYear(files.ID3V2, "")
	// 9. both ID3V1 and ID3V2 have errors
	tm9 := files.NewTrackMetadata()
	tm9.SetErrorCause(files.ID3V1, "bad format")
	tm9.SetErrorCause(files.ID3V2, "bad format")
	tm9.SetAlbumYear(files.ID3V1, "")
	tm9.SetAlbumYear(files.ID3V2, "")
	tests := map[string]struct {
		tm                          *files.TrackMetadata
		canonicalAlbumYear          string
		wantDiffers                 bool
		wantCorrectedID3V1AlbumYear string
		wantCorrectedID3V2AlbumYear string
		wantID3V1EditRequired       bool
		wantID3V2EditRequired       bool
	}{
		"no errors, no differences": {
			tm:                          tm1,
			canonicalAlbumYear:          expectedYear,
			wantDiffers:                 false,
			wantCorrectedID3V1AlbumYear: "",
			wantCorrectedID3V2AlbumYear: "",
			wantID3V1EditRequired:       false,
			wantID3V2EditRequired:       false,
		},
		"no errors, ID3V1 differs": {
			tm:                          tm2,
			canonicalAlbumYear:          expectedYear,
			wantDiffers:                 true,
			wantCorrectedID3V1AlbumYear: expectedYear,
			wantCorrectedID3V2AlbumYear: "",
			wantID3V1EditRequired:       true,
			wantID3V2EditRequired:       false,
		},
		"no errors, ID3V2 differs": {
			tm:                          tm3,
			canonicalAlbumYear:          expectedYear,
			wantDiffers:                 true,
			wantCorrectedID3V1AlbumYear: "",
			wantCorrectedID3V2AlbumYear: expectedYear,
			wantID3V1EditRequired:       false,
			wantID3V2EditRequired:       true,
		},
		"no errors, both differs": {
			tm:                          tm4,
			canonicalAlbumYear:          expectedYear,
			wantDiffers:                 true,
			wantCorrectedID3V1AlbumYear: expectedYear,
			wantCorrectedID3V2AlbumYear: expectedYear,
			wantID3V1EditRequired:       true,
			wantID3V2EditRequired:       true,
		},
		"ID3V1 error, no differences": {
			tm:                          tm5,
			canonicalAlbumYear:          expectedYear,
			wantDiffers:                 false,
			wantCorrectedID3V1AlbumYear: "",
			wantCorrectedID3V2AlbumYear: "",
			wantID3V1EditRequired:       false,
			wantID3V2EditRequired:       false,
		},
		"ID3V1 error, ID3V2 differs": {
			tm:                          tm6,
			canonicalAlbumYear:          expectedYear,
			wantDiffers:                 true,
			wantCorrectedID3V1AlbumYear: "",
			wantCorrectedID3V2AlbumYear: expectedYear,
			wantID3V1EditRequired:       false,
			wantID3V2EditRequired:       true,
		},
		"ID3V2 error, no differences": {
			tm:                          tm7,
			canonicalAlbumYear:          expectedYear,
			wantDiffers:                 false,
			wantCorrectedID3V1AlbumYear: "",
			wantCorrectedID3V2AlbumYear: "",
			wantID3V1EditRequired:       false,
			wantID3V2EditRequired:       false,
		},
		"ID3V2 error, ID3V1 differs": {
			tm:                          tm8,
			canonicalAlbumYear:          expectedYear,
			wantDiffers:                 true,
			wantCorrectedID3V1AlbumYear: expectedYear,
			wantCorrectedID3V2AlbumYear: "",
			wantID3V1EditRequired:       true,
			wantID3V2EditRequired:       false,
		},
		"both errors": {
			tm:                          tm9,
			canonicalAlbumYear:          expectedYear,
			wantDiffers:                 false,
			wantCorrectedID3V1AlbumYear: "",
			wantCorrectedID3V2AlbumYear: "",
			wantID3V1EditRequired:       false,
			wantID3V2EditRequired:       false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tm.AlbumYearDiffers(tt.canonicalAlbumYear); got != tt.wantDiffers {
				t.Errorf("TrackMetadata.AlbumYearDiffers() = %v, want %v", got, tt.wantDiffers)
			}
			if got := tt.tm.AlbumYear(files.ID3V1).Correction(); got != tt.wantCorrectedID3V1AlbumYear {
				t.Errorf("corrected ID3V1 album year = %q, want %q", got, tt.wantCorrectedID3V1AlbumYear)
			}
			if got := tt.tm.AlbumYear(files.ID3V2).Correction(); got != tt.wantCorrectedID3V2AlbumYear {
				t.Errorf("corrected ID3V2 album year = %q, want %q", got, tt.wantCorrectedID3V2AlbumYear)
			}
			if got := tt.tm.EditRequired(files.ID3V1); got != tt.wantID3V1EditRequired {
				t.Errorf("ID3V1 edit required = %t, want %t", got, tt.wantID3V1EditRequired)
			}
			if got := tt.tm.EditRequired(files.ID3V2); got != tt.wantID3V2EditRequired {
				t.Errorf("ID3V2 edit required = %t, want %t", got, tt.wantID3V2EditRequired)
			}
		})
	}
}

func TestTrackMetadata_CDIdentifierDiffers(t *testing.T) {
	expectedBody := []byte("my lovely CD")
	canonicalFrame := id3v2.UnknownFrame{Body: expectedBody}
	// 1. ID3v2 does not have an error, the CD Identifier does not differ
	tm1 := files.NewTrackMetadata()
	tm1.SetCDIdentifier(expectedBody)
	// 2. ID3v2 does not have an error, the CD Identifier differs
	tm2 := files.NewTrackMetadata()
	tm2.SetCDIdentifier([]byte("some other CD"))
	// 3. ID3V2 has an error
	tm3 := files.NewTrackMetadata()
	tm3.SetErrorCause(files.ID3V2, "bad format")
	tm3.SetCDIdentifier([]byte{})
	tests := map[string]struct {
		tm                            *files.TrackMetadata
		canonicalCDIdentifier         id3v2.UnknownFrame
		wantDiffers                   bool
		wantCorrectedCDIdentifierBody []byte
		wantID3V2EditRequired         bool
	}{
		"no error, no difference": {
			tm:                            tm1,
			canonicalCDIdentifier:         canonicalFrame,
			wantDiffers:                   false,
			wantCorrectedCDIdentifierBody: []byte{},
			wantID3V2EditRequired:         false,
		},
		"no error, identifier differs": {
			tm:                            tm2,
			canonicalCDIdentifier:         canonicalFrame,
			wantDiffers:                   true,
			wantCorrectedCDIdentifierBody: expectedBody,
			wantID3V2EditRequired:         true,
		},
		"ID3V2 error": {
			tm:                            tm3,
			canonicalCDIdentifier:         canonicalFrame,
			wantDiffers:                   false,
			wantCorrectedCDIdentifierBody: []byte{},
			wantID3V2EditRequired:         false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tm.CDIdentifierDiffers(tt.canonicalCDIdentifier); got != tt.wantDiffers {
				t.Errorf("TrackMetadata.CDIdentifierDiffers() = %v, want %v", got, tt.wantDiffers)
			}
			if got := tt.tm.EditRequired(files.ID3V2); got != tt.wantID3V2EditRequired {
				t.Errorf("ID3V2 edit required = %t, want %t", got, tt.wantID3V2EditRequired)
			}
			got := tt.tm.CDIdentifier().Correction().Body
			if len(got) == 0 {
				if len(tt.wantCorrectedCDIdentifierBody) != 0 {
					t.Errorf("corrected CD Identifier = %v, want %v", got, tt.wantCorrectedCDIdentifierBody)
				}
			} else {
				if !reflect.DeepEqual(got, tt.wantCorrectedCDIdentifierBody) {
					t.Errorf("corrected CD Identifier = %v, want %v", got, tt.wantCorrectedCDIdentifierBody)
				}
			}
		})
	}
}

func TestTrackMetadata_CanonicalAlbumNameMatches(t *testing.T) {
	albumName := "my favorite album"
	tm1 := files.NewTrackMetadata()
	tm1.SetAlbumName(files.ID3V1, albumName)
	tm1.SetCanonicalSource(files.ID3V1)
	tm2 := files.NewTrackMetadata()
	tm2.SetAlbumName(files.ID3V1, "my other favorite album")
	tm2.SetCanonicalSource(files.ID3V1)
	tm3 := files.NewTrackMetadata()
	tm3.SetAlbumName(files.ID3V2, albumName)
	tm3.SetCanonicalSource(files.ID3V2)
	tm4 := files.NewTrackMetadata()
	tm4.SetAlbumName(files.ID3V2, "my other favorite album")
	tm4.SetCanonicalSource(files.ID3V2)
	tests := map[string]struct {
		tm           *files.TrackMetadata
		nameFromFile string
		want         bool
	}{
		"no data": {
			tm:           files.NewTrackMetadata(),
			nameFromFile: albumName,
			want:         false,
		},
		"ID3V1 match": {
			tm:           tm1,
			nameFromFile: albumName,
			want:         true,
		},
		"ID3V1 no match": {
			tm:           tm2,
			nameFromFile: albumName,
			want:         false,
		},
		"ID3V2 match": {
			tm:           tm3,
			nameFromFile: albumName,
			want:         true,
		},
		"ID3V2 no match": {
			tm:           tm4,
			nameFromFile: albumName,
			want:         false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tm.CanonicalAlbumNameMatches(tt.nameFromFile); got != tt.want {
				t.Errorf("TrackMetadata.CanonicalAlbumNameMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrackMetadata_CanonicalArtistNameMatches(t *testing.T) {
	artistName := "my favorite artist"
	tm1 := files.NewTrackMetadata()
	tm1.SetArtistName(files.ID3V1, artistName)
	tm1.SetCanonicalSource(files.ID3V1)
	tm2 := files.NewTrackMetadata()
	tm2.SetArtistName(files.ID3V1, "my other favorite artist")
	tm2.SetCanonicalSource(files.ID3V1)
	tm3 := files.NewTrackMetadata()
	tm3.SetArtistName(files.ID3V2, artistName)
	tm3.SetCanonicalSource(files.ID3V2)
	tm4 := files.NewTrackMetadata()
	tm4.SetArtistName(files.ID3V2, "my other favorite artist")
	tm4.SetCanonicalSource(files.ID3V2)
	tests := map[string]struct {
		tm           *files.TrackMetadata
		nameFromFile string
		want         bool
	}{
		"no data": {
			tm:           files.NewTrackMetadata(),
			nameFromFile: artistName,
			want:         false,
		},
		"ID3V1 match": {
			tm:           tm1,
			nameFromFile: artistName,
			want:         true,
		},
		"ID3V1 no match": {
			tm:           tm2,
			nameFromFile: artistName,
			want:         false,
		},
		"ID3V2 match": {
			tm:           tm3,
			nameFromFile: artistName,
			want:         true,
		},
		"ID3V2 no match": {
			tm:           tm4,
			nameFromFile: artistName,
			want:         false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.tm.CanonicalArtistNameMatches(tt.nameFromFile); got != tt.want {
				t.Errorf("TrackMetadata.CanonicalArtistNameMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrackMetadata_Update(t *testing.T) {
	// create some TrackMetadata to apply
	loadedTm := files.NewTrackMetadata()
	for _, src := range []files.SourceType{files.ID3V1, files.ID3V2} {
		loadedTm.CorrectArtistName(src, "corrected artist")
		loadedTm.CorrectAlbumName(src, "corrected album")
		loadedTm.CorrectAlbumGenre(src, "rock")
		loadedTm.CorrectAlbumYear(src, "2024")
		loadedTm.CorrectTrackName(src, "corrected name")
		loadedTm.CorrectTrackNumber(src, 42)
		loadedTm.SetEditRequired(src)
	}
	loadedTm.CorrectCDIdentifier([]byte("corrected CD identifier"))
	// create a valid file
	testDir := "Update"
	defer func() {
		cmd_toolkit.FileSystem().RemoveAll(testDir)
	}()
	cmd_toolkit.Mkdir(testDir)
	completeFile := "01 complete.mp3"
	payloadID3v1Only := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	payloadID3v1Only = append(payloadID3v1Only, id3v1DataSet1...)
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
	payloadID3v2Only := createID3v2TaggedData([]byte{}, frames)
	payloadComplete := payloadID3v2Only
	payloadComplete = append(payloadComplete, payloadID3v1Only...)
	createFileWithContent(testDir, completeFile, payloadComplete)
	tests := map[string]struct {
		tm    *files.TrackMetadata
		path  string
		wantE []error
	}{
		"no data": {
			tm:    files.NewTrackMetadata(),
			path:  "",
			wantE: []error{},
		},
		"bad file": {
			tm:   loadedTm,
			path: "no such path",
			wantE: []error{
				fmt.Errorf("open no such path: The system cannot find the file specified."),
				fmt.Errorf("open no such path: The system cannot find the file specified."),
			},
		},
		"good file": {
			tm:    loadedTm,
			path:  filepath.Join(testDir, completeFile),
			wantE: []error{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if gotE := tt.tm.Update(tt.path); len(gotE) != len(tt.wantE) {
				t.Errorf("TrackMetadata.Update() = %v, want %v", gotE, tt.wantE)
			}
		})
	}
}
