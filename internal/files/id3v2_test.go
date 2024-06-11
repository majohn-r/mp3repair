package files_test

import (
	"fmt"
	"io"
	"mp3repair/internal/files"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/bogem/id3v2/v2"
	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/spf13/afero"
)

func makeTextFrame(id, content string) []byte {
	frame := make([]byte, 0)
	frame = append(frame, []byte(id)...)
	contentSize := 1 + len(content)
	factor := 256 * 256 * 256
	for k := 0; k < 4; k++ {
		frame = append(frame, byte(contentSize/factor))
		contentSize %= factor
		factor /= 256
	}
	frame = append(frame, []byte{0, 0, 0}...)
	frame = append(frame, []byte(content)...)
	return frame
}

// createID3v2TaggedData creates ID3V2-tagged content. This code is
// based on reading https://id3.org/id3v2.3.0 and on looking at a hex dump of a
// real mp3 file.
func createID3v2TaggedData(audio []byte, frames map[string]string) []byte {
	content := make([]byte, 0)
	// block off tag header
	content = append(content, []byte("ID3")...)
	content = append(content, []byte{3, 0, 0, 0, 0, 0, 0}...)
	// add some text frames; order is fixed for testing
	var keys []string
	for key := range frames {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		content = append(content, makeTextFrame(key, frames[key])...)
	}
	contentLength := len(content) - 10
	factor := 128 * 128 * 128
	for k := 0; k < 4; k++ {
		content[6+k] = byte(contentLength / factor)
		contentLength %= factor
		factor /= 128
	}
	// add payload
	content = append(content, audio...)
	return content
}

func Test_rawReadID3V2Metadata(t *testing.T) {
	originalFileSystem := cmd_toolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmd_toolkit.AssignFileSystem(originalFileSystem)
	}()
	const fnName = "rawReadID3V2Metadata()"
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
	content := createID3v2TaggedData(payload, frames)
	createFileWithContent(".", "goodFile.mp3", content)
	frames["TRCK"] = "oops"
	createFileWithContent(".", "badFile.mp3", createID3v2TaggedData(payload, frames))
	type args struct {
		path string
	}
	tests := map[string]struct {
		args
		wantD *files.Id3v2Metadata
	}{
		"bad test": {
			args:  args{path: "./noSuchFile!.mp3"},
			wantD: &files.Id3v2Metadata{Err: fmt.Errorf("foo")},
		},
		"good test": {
			args: args{path: "./goodFile.mp3"},
			wantD: &files.Id3v2Metadata{
				AlbumTitle:  "unknown album",
				ArtistName:  "unknown artist",
				TrackName:   "unknown track",
				TrackNumber: 2,
			},
		},
		"bad data test": {
			args:  args{path: "./badFile.mp3"},
			wantD: &files.Id3v2Metadata{Err: files.ErrMalformedTrackNumber},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotD := files.RawReadID3V2Metadata(tt.args.path)
			if gotD.HasError() {
				if !tt.wantD.HasError() {
					t.Errorf("%s = %v, want %v", fnName, gotD, tt.wantD)
				}
			} else if tt.wantD.HasError() {
				t.Errorf("%s = %v, want %v", fnName, gotD, tt.wantD)
			}
		})
	}
}

func Test_removeLeadingBOMs(t *testing.T) {
	const fnName = "removeLeadingBOMs()"
	type args struct {
		s string
	}
	tests := map[string]struct {
		args
		want string
	}{
		"normal string":   {args: args{s: "normal"}, want: "normal"},
		"abnormal string": {args: args{s: "\ufeff\ufeffnormal"}, want: "normal"},
		"empty string":    {args: args{}},
		"nothing but BOM": {args: args{s: "\ufeff\ufeff"}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.RemoveLeadingBOMs(tt.args.s); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func Test_toTrackNumber(t *testing.T) {
	const fnName = "toTrackNumber()"
	type args struct {
		s string
	}
	tests := map[string]struct {
		args
		wantI   int
		wantErr bool
	}{
		"good value": {
			args:    args{s: "12"},
			wantI:   12,
			wantErr: false,
		},
		"empty value": {
			args:    args{s: ""},
			wantI:   0,
			wantErr: true,
		},
		"BOM-infested empty value": {
			args:    args{s: "\ufeff"},
			wantI:   0,
			wantErr: true,
		},
		"negative value": {
			args:    args{s: "-12"},
			wantI:   0,
			wantErr: true,
		},
		"invalid value": {
			args:    args{s: "foo"},
			wantI:   0,
			wantErr: true,
		},
		"complicated value": {
			args:    args{s: "12/39"},
			wantI:   12,
			wantErr: false,
		},
		"BOM-infested complicated value": {
			args:    args{s: "\ufeff12/39"},
			wantI:   12,
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotI, gotErr := files.ToTrackNumber(tt.args.s)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, gotErr, tt.wantErr)
				return
			}
			if gotErr == nil && gotI != tt.wantI {
				t.Errorf("%s = %d, want %d", fnName, gotI, tt.wantI)
			}
		})
	}
}

// this struct implements id3v2.Framer as a means to provide an unexpected kind
// of Framer
type unspecifiedFrame struct {
	content string
}

func (u unspecifiedFrame) Size() int {
	return len(u.content)
}

func (u unspecifiedFrame) UniqueIdentifier() string {
	return ""
}

func (u unspecifiedFrame) WriteTo(w io.Writer) (int64, error) {
	count, fileErr := w.Write([]byte(u.content))
	return int64(count), fileErr
}

func Test_selectUnknownFrame(t *testing.T) {
	const fnName = "selectUnknownFrame()"
	type args struct {
		mcdiFramers []id3v2.Framer
	}
	tests := map[string]struct {
		args
		want id3v2.UnknownFrame
	}{
		"degenerate case": {
			args: args{mcdiFramers: nil},
			want: id3v2.UnknownFrame{Body: []byte{0}},
		},
		"too many framers": {
			args: args{mcdiFramers: []id3v2.Framer{id3v2.UnknownFrame{Body: []byte{1, 2, 3}}, id3v2.UnknownFrame{Body: []byte{4, 5, 6}}}},
			want: id3v2.UnknownFrame{Body: []byte{0}},
		},
		"wrong kind of framer": {
			args: args{mcdiFramers: []id3v2.Framer{unspecifiedFrame{content: "no good"}}},
			want: id3v2.UnknownFrame{Body: []byte{0}},
		},
		"desired use case": {
			args: args{mcdiFramers: []id3v2.Framer{id3v2.UnknownFrame{Body: []byte{0, 1, 2}}}},
			want: id3v2.UnknownFrame{Body: []byte{0, 1, 2}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.SelectUnknownFrame(tt.args.mcdiFramers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestID3V2TrackFrameStringType(t *testing.T) {
	const fnName = "ID3V2TrackFrame.String()"
	tests := map[string]struct {
		f    *files.Id3v2TrackFrame
		want string
	}{
		"usual": {
			f:    &files.Id3v2TrackFrame{Name: "T1", Value: "V1"},
			want: "T1 = \"V1\"",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.f.String(); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func Test_readID3V2Metadata(t *testing.T) {
	originalFileSystem := cmd_toolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmd_toolkit.AssignFileSystem(originalFileSystem)
	}()
	const fnName = "readID3V2Metadata()"
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
		"T???": "who knows?",
		"Fake": "ummm",
	}
	content := createID3v2TaggedData(payload, frames)
	goodFileName := "goodFile.mp3"
	createFileWithContent(".", goodFileName, content)
	type args struct {
		path string
	}
	tests := map[string]struct {
		args
		wantVersion      byte
		wantEncoding     string
		wantFrameStrings []string
		wantErr          bool
	}{
		"error case": {args: args{path: "./no such file"}, wantErr: true},
		"good case": {
			args:         args{path: filepath.Join(".", goodFileName)},
			wantEncoding: "ISO-8859-1",
			wantVersion:  3,
			wantFrameStrings: []string{
				`Fake = "<<[]byte{0x0, 0x75, 0x6d, 0x6d, 0x6d}>>"`,
				`T??? = "who knows?"`,
				`TALB = "unknown album"`,
				`TCOM = "a couple of idiots"`,
				`TCON = "dance music"`,
				`TIT2 = "unknown track"`,
				`TLEN = "1000"`,
				`TPE1 = "unknown artist"`,
				`TRCK = "2"`,
				`TYER = "2022"`,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// ignoring raw frames ...
			gotInfo, gotErr := files.ReadID3V2Metadata(tt.args.path)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, gotErr, tt.wantErr)
				return
			}
			if gotErr == nil {
				if gotInfo.Version != tt.wantVersion {
					t.Errorf("%s gotInfo.Version = %v, want %v", fnName, gotInfo.Version, tt.wantVersion)
				}
				if gotInfo.Encoding != tt.wantEncoding {
					t.Errorf("%s gotInfo.Encoding = %v, want %v", fnName, gotInfo.Encoding, tt.wantEncoding)
				}
				if !reflect.DeepEqual(gotInfo.FrameStrings, tt.wantFrameStrings) {
					t.Errorf("%s gotInfo.FrameStrings = %v, want %v", fnName, gotInfo.FrameStrings, tt.wantFrameStrings)
				}
			}
		})
	}
}

func Test_framerSliceAsString(t *testing.T) {
	const fnName = "framerSliceAsString()"
	type args struct {
		f []id3v2.Framer
	}
	tests := map[string]struct {
		args
		want string
	}{
		"single UnknownFrame": {
			args: args{f: []id3v2.Framer{id3v2.UnknownFrame{Body: []byte{0, 1, 2}}}},
			want: "<<[]byte{0x0, 0x1, 0x2}>>",
		},
		"unexpected frame": {
			args: args{f: []id3v2.Framer{unspecifiedFrame{content: "hello world"}}},
			want: "<<files_test.unspecifiedFrame{content:\"hello world\"}>>",
		},
		"multiple frames": {
			args: args{
				f: []id3v2.Framer{
					id3v2.UnknownFrame{Body: []byte{0, 1, 2}},
					unspecifiedFrame{content: "hello world"},
				},
			},
			want: "<<[0 []byte{0x0, 0x1, 0x2}]," +
				" [1 files_test.unspecifiedFrame{content:\"hello world\"}]>>",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.FramerSliceAsString(tt.args.f); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func Test_id3v2NameDiffers(t *testing.T) {
	const fnName = "id3v2NameDiffers()"
	type args struct {
		cS *files.ComparableStrings
	}
	tests := map[string]struct {
		args
		want bool
	}{
		"identical strings": {
			args: args{
				files.NewComparableStrings().WithExternal(
					"simple name").WithMetadata("simple name"),
			},
			want: false,
		},
		"identical strings with case differences": {
			args: args{
				files.NewComparableStrings().WithExternal(
					"SIMPLE name").WithMetadata("simple NAME"),
			},
			want: false,
		},
		"strings of different length": {
			args: args{
				files.NewComparableStrings().WithExternal(
					"simple name").WithMetadata("artist: simple name"),
			},
			want: true,
		},
		"use of runes that are illegal for file names": {
			args: args{
				files.NewComparableStrings().WithExternal(
					"simple_name").WithMetadata("simple:name"),
			},
			want: false,
		},
		"metadata with trailing space": {
			args: args{
				files.NewComparableStrings().WithExternal(
					"simple name").WithMetadata("simple name "),
			},
			want: false,
		},
		"period on the end": {
			args: args{
				files.NewComparableStrings().WithExternal(
					"simple name.").WithMetadata("simple name."),
			},
			want: false,
		},
		"complex mismatch": {
			args: args{
				files.NewComparableStrings().WithExternal(
					"simple_name").WithMetadata("simple: nam"),
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.Id3v2NameDiffers(tt.args.cS); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_normalizeGenre(t *testing.T) {
	const fnName = "normalizeGenre()"
	type args struct {
		g string
	}
	type test struct {
		args
		want string
	}
	tests := map[string]test{}
	for k, v := range files.GenreMap {
		tests[v] = test{args: args{g: fmt.Sprintf("(%d)%s", k, v)}, want: v}
		if v == "Rhythm and Blues" {
			tests["R&B"] = test{args: args{g: fmt.Sprintf("(%d)R&B", k)}, want: v}
		}
	}
	tests["prog rock"] = test{args: args{g: "prog rock"}, want: "prog rock"}
	tests["unexpected k/v"] = test{
		args: args{g: "(256)martian folk rock"},
		want: "(256)martian folk rock",
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.NormalizeGenre(tt.args.g); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_id3v2GenreDiffers(t *testing.T) {
	const fnName = "id3v2GenreDiffers()"
	type args struct {
		cS *files.ComparableStrings
	}
	tests := map[string]struct {
		args
		want bool
	}{
		"match": {
			args: args{
				cS: files.NewComparableStrings().WithExternal(
					"Classic Rock").WithMetadata("Classic Rock"),
			},
			want: false,
		},
		"no match": {
			args: args{
				cS: files.NewComparableStrings().WithExternal(
					"Classic Rock").WithMetadata("classic rock"),
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.Id3v2GenreDiffers(tt.args.cS); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_updateID3V2Metadata(t *testing.T) {
	type args struct {
		tM   *files.TrackMetadataV1
		path string
	}
	tests := map[string]struct {
		args
		wantErr bool
	}{
		"no need for edit": {
			args:    args{tM: files.NewTrackMetadataV1()}, // empty struct implicitly does not need an edit
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if gotErr := files.UpdateID3V2Metadata(tt.args.tM, tt.args.path); (gotErr != nil) != tt.wantErr {
				t.Errorf("UpdateID3V2Metadata() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestIsTagAbsent(t *testing.T) {
	tagWithContent := id3v2.NewEmptyTag()
	tagWithContent.AddTextFrame("TFOO", id3v2.EncodingISO, "foo")
	tests := map[string]struct {
		tag  *id3v2.Tag
		want bool
	}{
		"nil":          {tag: nil, want: true},
		"empty":        {tag: id3v2.NewEmptyTag(), want: true},
		"with a frame": {tag: tagWithContent, want: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.IsTagAbsent(tt.tag); got != tt.want {
				t.Errorf("IsTagAbsent() = %v, want %v", got, tt.want)
			}
		})
	}
}
