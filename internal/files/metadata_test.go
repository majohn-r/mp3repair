package files_test

import (
	"mp3repair/internal/files"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bogem/id3v2/v2"
	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
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

func Test_trackMetadata_setId3v1Values(t *testing.T) {
	const fnName = "trackMetadata.setId3v1Values()"
	type args struct {
		v1 *files.Id3v1Metadata
	}
	tests := map[string]struct {
		tM *files.TrackMetadata
		args
		wantTM *files.TrackMetadata
	}{
		"complete test": {
			tM:   files.NewTrackMetadata(),
			args: args{v1: NewID3v1MetadataWithData(id3v1DataSet1)},
			wantTM: files.NewTrackMetadata().WithAlbumNames(
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

func Test_trackMetadata_setId3v2Values(t *testing.T) {
	const fnName = "trackMetadata.setId3v1Values()"
	type args struct {
		d *files.Id3v2Metadata
	}
	tests := map[string]struct {
		tM *files.TrackMetadata
		args
		wantTM *files.TrackMetadata
	}{
		"complete test": {
			tM: files.NewTrackMetadata(),
			args: args{
				d: files.NewId3v2Metadata().WithAlbumName("Great album").WithArtistName(
					"Great artist").WithTrackName("Great track").WithGenre("Pop").WithYear(
					"2022").WithTrackNumber(1).WithMusicCDIdentifier([]byte{0, 2, 4}),
			},
			wantTM: files.NewTrackMetadata().WithAlbumNames(
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

func Test_readMetadata(t *testing.T) {
	const fnName = "readMetadata()"
	testDir := "readMetadata"
	if err := cmd_toolkit.Mkdir(testDir); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, testDir, err)
	}
	taglessFile := "01 tagless.mp3"
	if err := createFile(testDir, taglessFile); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, taglessFile, err)
	}
	id3v1OnlyFile := "02 id3v1.mp3"
	payloadID3v1Only := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	payloadID3v1Only = append(payloadID3v1Only, id3v1DataSet1...)
	if err := createFileWithContent(testDir, id3v1OnlyFile,
		payloadID3v1Only); err != nil {
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
	payloadID3v2Only := createID3v2TaggedData([]byte{}, frames)
	if err := createFileWithContent(testDir, id3v2OnlyFile,
		payloadID3v2Only); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, id3v2OnlyFile, err)
	}
	completeFile := "04 complete.mp3"
	payloadComplete := payloadID3v2Only
	payloadComplete = append(payloadComplete, payloadID3v1Only...)
	if err := createFileWithContent(testDir, completeFile,
		payloadComplete); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, completeFile, err)
	}
	defer func() {
		destroyDirectory(fnName, testDir)
	}()
	type args struct {
		path string
	}
	tests := map[string]struct {
		args
		want *files.TrackMetadata
	}{
		"missing file": {
			args: args{path: filepath.Join(testDir, "no such file.mp3")},
			want: files.NewTrackMetadata().WithErrorCauses([]string{
				"", cannotOpenFile, cannotOpenFile}),
		},
		"no metadata": {
			args: args{path: filepath.Join(testDir, taglessFile)},
			want: files.NewTrackMetadata().WithErrorCauses([]string{
				"", negativeSeek, files.ErrMissingTrackNumber.Error()}),
		},
		"only id3v1 metadata": {
			args: args{path: filepath.Join(testDir, id3v1OnlyFile)},
			want: files.NewTrackMetadata().WithAlbumNames(
				[]string{"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames(
				[]string{"", "The Beatles", ""}).WithTrackNames(
				[]string{"", "Ringo - Pop Profile [Interview", ""}).WithGenres(
				[]string{"", "Other", ""}).WithYears([]string{
				"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses(
				[]string{"", "", files.ErrMissingTrackNumber.Error()}),
		},
		"only id3v2 metadata": {
			args: args{path: filepath.Join(testDir, id3v2OnlyFile)},
			want: files.NewTrackMetadata().WithAlbumNames(
				[]string{"", "", "unknown album"}).WithArtistNames(
				[]string{"", "", "unknown artist"}).WithTrackNames(
				[]string{"", "", "unknown track"}).WithGenres(
				[]string{"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers(
				[]int{0, 0, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2).WithErrorCauses(
				[]string{"", noID3V1Metadata, ""}),
		},
		"all metadata": {
			args: args{path: filepath.Join(testDir, completeFile)},
			want: files.NewTrackMetadata().WithAlbumNames(
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
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_trackMetadata_isValid(t *testing.T) {
	const fnName = "trackMetadata.isValid()"
	tests := map[string]struct {
		tM   *files.TrackMetadata
		want bool
	}{
		"uninitialized data": {tM: files.NewTrackMetadata(), want: false},
		"after read failure": {
			tM: files.NewTrackMetadata().WithErrorCauses([]string{
				"", cannotOpenFile, cannotOpenFile,
			}),
			want: false,
		},
		"after reading no metadata": {
			tM: files.NewTrackMetadata().WithErrorCauses([]string{
				"", negativeSeek, zeroBytes}),
			want: false,
		},
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			want: true,
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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

func Test_trackMetadata_canonicalArtist(t *testing.T) {
	const fnName = "trackMetadata.CanonicalArtist()"
	tests := map[string]struct {
		tM   *files.TrackMetadata
		want string
	}{
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			want: "The Beatles",
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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

func Test_trackMetadata_canonicalAlbum(t *testing.T) {
	const fnName = "trackMetadata.CanonicalAlbum()"
	tests := map[string]struct {
		tM   *files.TrackMetadata
		want string
	}{
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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

func Test_trackMetadata_canonicalGenre(t *testing.T) {
	const fnName = "trackMetadata.CanonicalGenre()"
	tests := map[string]struct {
		tM   *files.TrackMetadata
		want string
	}{
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			want: "Other",
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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

func Test_trackMetadata_canonicalYear(t *testing.T) {
	const fnName = "trackMetadata.CanonicalYear()"
	tests := map[string]struct {
		tM   *files.TrackMetadata
		want string
	}{
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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

func Test_trackMetadata_canonicalMusicCDIdentifier(t *testing.T) {
	const fnName = "trackMetadata.CanonicalMusicCDIdentifier()"
	tests := map[string]struct {
		tM   *files.TrackMetadata
		want id3v2.UnknownFrame
	}{
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{
				"", "2013", ""}).WithTrackNumbers([]int{0, 29, 0}).WithPrimarySource(
				files.ID3V1).WithErrorCauses([]string{"", "", zeroBytes}),
			want: id3v2.UnknownFrame{},
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears([]string{
				"", "", "2022"}).WithTrackNumbers([]int{0, 0, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			want: id3v2.UnknownFrame{Body: []byte{0}},
		},
		"after reading all metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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

func Test_trackMetadata_errors(t *testing.T) {
	const fnName = "trackMetadata.errors()"
	tests := map[string]struct {
		tM   *files.TrackMetadata
		want []string
	}{
		"uninitialized data": {tM: files.NewTrackMetadata(), want: []string{}},
		"after read failure": {
			tM: files.NewTrackMetadata().WithErrorCauses([]string{
				"", cannotOpenFile, cannotOpenFile}),
			want: []string{cannotOpenFile, cannotOpenFile},
		},
		"after reading no metadata": {
			tM: files.NewTrackMetadata().WithErrorCauses([]string{
				"", negativeSeek, zeroBytes}),
			want: []string{negativeSeek, zeroBytes},
		},
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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

func Test_trackMetadata_TrackDiffers(t *testing.T) {
	const fnName = "trackMetadata.TrackDiffers()"
	type args struct {
		track int
	}
	tests := map[string]struct {
		tM *files.TrackMetadata
		args
		want   bool
		wantTM *files.TrackMetadata
	}{
		"after read failure": {
			tM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
			args: args{track: 20},
			want: false,
			wantTM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
		},
		"after reading no metadata": {
			tM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
			args: args{track: 20},
			want: false,
			wantTM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
		},
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears(
				[]string{"", "2013", ""}).WithTrackNumbers([]int{
				0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			args: args{track: 20},
			want: true,
			wantTM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers([]int{
				0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			args: args{track: 20},
			want: true,
			wantTM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers([]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			args: args{track: 20},
			want: true,
			wantTM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			if got := tt.tM.TrackDiffers(tt.args.track); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
			if !reflect.DeepEqual(tt.tM, tt.wantTM) {
				t.Errorf("%s got TM %v, want TM %v", fnName, tt.tM, tt.wantTM)
			}
		})
	}
}

func Test_trackMetadata_TrackTitleDiffers(t *testing.T) {
	const fnName = "trackMetadata.TrackTitleDiffers()"
	type args struct {
		title string
	}
	tests := map[string]struct {
		tM *files.TrackMetadata
		args
		wantDiffers bool
		wantTM      *files.TrackMetadata
	}{
		"after read failure": {
			tM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
			args:        args{title: "track name"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
		},
		"after reading no metadata": {
			tM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
			args:        args{title: "track name"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
		},
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			args:        args{title: "track name"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}).WithCorrectedTrackNames([]string{
				"", "track name", ""}).WithRequiresEdits([]bool{false, true, false}),
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers([]int{
				0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			args:        args{title: "track name"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2),
			args:        args{title: "track name"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Theme from M*A*S*H", "Theme from M*A*S*H"}).WithGenres([]string{
				"", "Other", "dance music"}).WithYears(
				[]string{"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2),
			args:        args{title: "Theme From M-A-S-H"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadata().WithAlbumNames([]string{
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

func Test_trackMetadata_AlbumTitleDiffers(t *testing.T) {
	const fnName = "trackMetadata.AlbumTitleDiffers()"
	type args struct {
		albumTitle string
	}
	tests := map[string]struct {
		tM *files.TrackMetadata
		args
		wantDiffers bool
		wantTM      *files.TrackMetadata
	}{
		"after read failure": {
			tM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
			args:        args{albumTitle: "album name"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
		},
		"after reading no metadata": {
			tM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
			args:        args{albumTitle: "album name"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
		},
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			args:        args{albumTitle: "album name"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}).WithCorrectedAlbumNames([]string{
				"", "album name", ""}).WithRequiresEdits([]bool{false, true, false}),
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers([]int{
				0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			args:        args{albumTitle: "album name"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			args:        args{albumTitle: "album name"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadata().WithAlbumNames([]string{
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

func Test_trackMetadata_ArtistNameDiffers(t *testing.T) {
	const fnName = "trackMetadata.ArtistNameDiffers()"
	type args struct {
		artistName string
	}
	tests := map[string]struct {
		tM *files.TrackMetadata
		args
		wantDiffers bool
		wantTM      *files.TrackMetadata
	}{
		"after read failure": {
			tM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
			args:        args{artistName: "artist name"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
		},
		"after reading no metadata": {
			tM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
			args:        args{artistName: "artist name"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
		},
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			args:        args{artistName: "artist name"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}).WithCorrectedArtistNames([]string{
				"", "artist name", ""}).WithRequiresEdits([]bool{false, true, false}),
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers([]int{
				0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			args:        args{artistName: "artist name"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			args:        args{artistName: "artist name"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadata().WithAlbumNames([]string{
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

func Test_trackMetadata_GenreDiffers(t *testing.T) {
	const fnName = "trackMetadata.GenreDiffers()"
	type args struct {
		genre string
	}
	tests := map[string]struct {
		tM *files.TrackMetadata
		args
		wantDiffers bool
		wantTM      *files.TrackMetadata
	}{
		"after read failure": {
			tM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
			args:        args{genre: "Indie Pop"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
		},
		"after reading no metadata": {
			tM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
			args:        args{genre: "Indie Pop"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
		},
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses(
				[]string{"", "", zeroBytes}),
			args:        args{genre: "Indie Pop"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses(
				[]string{"", "", zeroBytes}),
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers(
				[]int{0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			args:        args{genre: "Indie Pop"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			args:        args{genre: "Indie Pop"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadata().WithAlbumNames([]string{
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

func Test_trackMetadata_YearDiffers(t *testing.T) {
	const fnName = "trackMetadata.YearDiffers()"
	type args struct {
		year string
	}
	tests := map[string]struct {
		tM *files.TrackMetadata
		args
		wantDiffers bool
		wantTM      *files.TrackMetadata
	}{
		"after read failure": {
			tM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
			args:        args{year: "1999"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
		},
		"after reading no metadata": {
			tM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
			args:        args{year: "1999"},
			wantDiffers: false,
			wantTM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
		},
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			args:        args{year: "1999"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}).WithCorrectedYears([]string{
				"", "1999", ""}).WithRequiresEdits([]bool{false, true, false}),
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers(
				[]int{0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			args:        args{year: "1999"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2),
			args:        args{year: "1999"},
			wantDiffers: true,
			wantTM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears([]string{
				"", "2013", "2022"}).WithTrackNumbers(
				[]int{0, 29, 2}).WithMusicCDIdentifier(
				[]byte{0}).WithPrimarySource(files.ID3V2).WithCorrectedYears([]string{
				"", "1999", "1999"}).WithRequiresEdits([]bool{false, true, true}),
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

func Test_trackMetadata_MCDIDiffers(t *testing.T) {
	const fnName = "trackMetadata.MCDIDiffers()"
	type args struct {
		f id3v2.UnknownFrame
	}
	tests := map[string]struct {
		tM *files.TrackMetadata
		args
		wantDiffers bool
		wantTM      *files.TrackMetadata
	}{
		"after read failure": {
			tM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
			args:        args{f: id3v2.UnknownFrame{Body: []byte{1, 2, 3}}},
			wantDiffers: false,
			wantTM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", cannotOpenFile, cannotOpenFile}),
		},
		"after reading no metadata": {
			tM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
			args:        args{f: id3v2.UnknownFrame{Body: []byte{1, 2, 3}}},
			wantDiffers: false,
			wantTM: files.NewTrackMetadata().WithErrorCauses(
				[]string{"", negativeSeek, zeroBytes}),
		},
		"after reading only id3v1 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
			args:        args{f: id3v2.UnknownFrame{Body: []byte{1, 2, 3}}},
			wantDiffers: false,
			wantTM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", ""}).WithArtistNames([]string{
				"", "The Beatles", ""}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", ""}).WithGenres([]string{
				"", "Other", ""}).WithYears([]string{"", "2013", ""}).WithTrackNumbers(
				[]int{0, 29, 0}).WithPrimarySource(files.ID3V1).WithErrorCauses([]string{
				"", "", zeroBytes}),
		},
		"after reading only id3v2 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "", "unknown album"}).WithArtistNames([]string{
				"", "", "unknown artist"}).WithTrackNames([]string{
				"", "", "unknown track"}).WithGenres([]string{
				"", "", "dance music"}).WithYears(
				[]string{"", "", "2022"}).WithTrackNumbers(
				[]int{0, 0, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(
				files.ID3V2).WithErrorCauses([]string{"", noID3V1Metadata, ""}),
			args:        args{f: id3v2.UnknownFrame{Body: []byte{1, 2, 3}}},
			wantDiffers: true,
			wantTM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
				"", "On Air: Live At The BBC, Volum", "unknown album"}).WithArtistNames(
				[]string{"", "The Beatles", "unknown artist"}).WithTrackNames([]string{
				"", "Ringo - Pop Profile [Interview", "unknown track"}).WithGenres(
				[]string{"", "Other", "dance music"}).WithYears(
				[]string{"", "2013", "2022"}).WithTrackNumbers([]int{
				0, 29, 2}).WithMusicCDIdentifier([]byte{0}).WithPrimarySource(files.ID3V2),
			args:        args{f: id3v2.UnknownFrame{Body: []byte{1, 2, 3}}},
			wantDiffers: true,
			wantTM: files.NewTrackMetadata().WithAlbumNames([]string{
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

func Test_trackMetadata_CanonicalAlbumTitleMatches(t *testing.T) {
	const fnName = "trackMetadata.CanonicalAlbumTitleMatches()"
	type args struct {
		albumTitle string
	}
	tests := map[string]struct {
		tM *files.TrackMetadata
		args
		want bool
	}{
		"mismatch after reading only id3v1 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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

func Test_trackMetadata_CanonicalArtistNameMatches(t *testing.T) {
	const fnName = "trackMetadata.CanonicalArtistNameMatches()"
	type args struct {
		artistName string
	}
	tests := map[string]struct {
		tM *files.TrackMetadata
		args
		want bool
	}{
		"mismatch after reading only id3v1 metadata": {
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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
			tM: files.NewTrackMetadata().WithAlbumNames([]string{
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
