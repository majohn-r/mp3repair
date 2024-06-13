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

func TestRawReadID3V2Metadata(t *testing.T) {
	originalFileSystem := cmd_toolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmd_toolkit.AssignFileSystem(originalFileSystem)
	}()
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
	tests := map[string]struct {
		path  string
		wantD *files.Id3v2Metadata
	}{
		"bad test": {
			path:  "./noSuchFile!.mp3",
			wantD: &files.Id3v2Metadata{Err: fmt.Errorf("foo")},
		},
		"good test": {
			path: "./goodFile.mp3",
			wantD: &files.Id3v2Metadata{
				AlbumTitle:  "unknown album",
				ArtistName:  "unknown artist",
				TrackName:   "unknown track",
				TrackNumber: 2,
			},
		},
		"bad data test": {
			path:  "./badFile.mp3",
			wantD: &files.Id3v2Metadata{Err: files.ErrMalformedTrackNumber},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotD := files.RawReadID3V2Metadata(tt.path)
			if gotD.HasError() {
				if !tt.wantD.HasError() {
					t.Errorf("RawReadID3V2Metadata() = %v, want %v", gotD, tt.wantD)
				}
			} else if tt.wantD.HasError() {
				t.Errorf("RawReadID3V2Metadata() = %v, want %v", gotD, tt.wantD)
			}
		})
	}
}

func TestRemoveLeadingBOMs(t *testing.T) {
	tests := map[string]struct {
		s    string
		want string
	}{
		"normal string":   {s: "normal", want: "normal"},
		"abnormal string": {s: "\ufeff\ufeffnormal", want: "normal"},
		"empty string":    {s: "", want: ""},
		"nothing but BOM": {s: "\ufeff\ufeff", want: ""},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.RemoveLeadingBOMs(tt.s); got != tt.want {
				t.Errorf("RemoveLeadingBOMs() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestToTrackNumber(t *testing.T) {
	tests := map[string]struct {
		s       string
		wantI   int
		wantErr bool
	}{
		"good value": {
			s:       "12",
			wantI:   12,
			wantErr: false,
		},
		"empty value": {
			s:       "",
			wantI:   0,
			wantErr: true,
		},
		"BOM-infested empty value": {
			s:       "\ufeff",
			wantI:   0,
			wantErr: true,
		},
		"negative value": {
			s:       "-12",
			wantI:   0,
			wantErr: true,
		},
		"invalid value": {
			s:       "foo",
			wantI:   0,
			wantErr: true,
		},
		"complicated value": {
			s:       "12/39",
			wantI:   12,
			wantErr: false,
		},
		"BOM-infested complicated value": {
			s:       "\ufeff12/39",
			wantI:   12,
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotI, gotErr := files.ToTrackNumber(tt.s)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("ToTrackNumber() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if gotErr == nil && gotI != tt.wantI {
				t.Errorf("ToTrackNumber() = %d, want %d", gotI, tt.wantI)
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

func TestSelectUnknownFrame(t *testing.T) {
	tests := map[string]struct {
		mcdiFramers []id3v2.Framer
		want        id3v2.UnknownFrame
	}{
		"degenerate case": {
			mcdiFramers: nil,
			want:        id3v2.UnknownFrame{Body: []byte{0}},
		},
		"too many framers": {
			mcdiFramers: []id3v2.Framer{
				id3v2.UnknownFrame{Body: []byte{1, 2, 3}},
				id3v2.UnknownFrame{Body: []byte{4, 5, 6}},
			},
			want: id3v2.UnknownFrame{Body: []byte{0}},
		},
		"wrong kind of framer": {
			mcdiFramers: []id3v2.Framer{unspecifiedFrame{content: "no good"}},
			want:        id3v2.UnknownFrame{Body: []byte{0}},
		},
		"desired use case": {
			mcdiFramers: []id3v2.Framer{id3v2.UnknownFrame{Body: []byte{0, 1, 2}}},
			want:        id3v2.UnknownFrame{Body: []byte{0, 1, 2}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.SelectUnknownFrame(tt.mcdiFramers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SelectUnknownFrame() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestID3V2TrackFrameStringType(t *testing.T) {
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
				t.Errorf("Id3v2TrackFrame.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReadID3V2Metadata(t *testing.T) {
	originalFileSystem := cmd_toolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmd_toolkit.AssignFileSystem(originalFileSystem)
	}()
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
	tests := map[string]struct {
		path             string
		wantVersion      byte
		wantEncoding     string
		wantFrameStrings []string
		wantErr          bool
	}{
		"error case": {
			path:    "./no such file",
			wantErr: true,
		},
		"good case": {
			path:         filepath.Join(".", goodFileName),
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
			gotInfo, gotErr := files.ReadID3V2Metadata(tt.path)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("ReadID3V2Metadata() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if gotErr == nil {
				if gotInfo.Version != tt.wantVersion {
					t.Errorf("ReadID3V2Metadata() gotInfo.Version = %v, want %v",
						gotInfo.Version, tt.wantVersion)
				}
				if gotInfo.Encoding != tt.wantEncoding {
					t.Errorf("ReadID3V2Metadata gotInfo.Encoding = %v, want %v",
						gotInfo.Encoding, tt.wantEncoding)
				}
				if !reflect.DeepEqual(gotInfo.FrameStrings, tt.wantFrameStrings) {
					t.Errorf("ReadID3V2Metadata gotInfo.FrameStrings = %v, want %v",
						gotInfo.FrameStrings, tt.wantFrameStrings)
				}
			}
		})
	}
}

func TestFramerSliceAsString(t *testing.T) {
	tests := map[string]struct {
		f    []id3v2.Framer
		want string
	}{
		"single UnknownFrame": {
			f:    []id3v2.Framer{id3v2.UnknownFrame{Body: []byte{0, 1, 2}}},
			want: "<<[]byte{0x0, 0x1, 0x2}>>",
		},
		"unexpected frame": {
			f:    []id3v2.Framer{unspecifiedFrame{content: "hello world"}},
			want: "<<files_test.unspecifiedFrame{content:\"hello world\"}>>",
		},
		"multiple frames": {
			f: []id3v2.Framer{
				id3v2.UnknownFrame{Body: []byte{0, 1, 2}},
				unspecifiedFrame{content: "hello world"},
			},
			want: "<<[0 []byte{0x0, 0x1, 0x2}]," +
				" [1 files_test.unspecifiedFrame{content:\"hello world\"}]>>",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.FramerSliceAsString(tt.f); got != tt.want {
				t.Errorf("FramerSliceAsString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestId3v2NameDiffers(t *testing.T) {
	tests := map[string]struct {
		cS   *files.ComparableStrings
		want bool
	}{
		"identical strings": {
			cS: &files.ComparableStrings{
				External: "simple name",
				Metadata: "simple name",
			},
			want: false,
		},
		"identical strings with case differences": {
			cS: &files.ComparableStrings{
				External: "SIMPLE name",
				Metadata: "simple NAME",
			},
			want: false,
		},
		"strings of different length": {
			cS: &files.ComparableStrings{
				External: "simple name",
				Metadata: "artist: simple name",
			},
			want: true,
		},
		"use of runes that are illegal for file names": {
			cS: &files.ComparableStrings{
				External: "simple_name",
				Metadata: "simple:name",
			},
			want: false,
		},
		"metadata with trailing space": {
			cS: &files.ComparableStrings{
				External: "simple name",
				Metadata: "simple name ",
			},
			want: false,
		},
		"period on the end": {
			cS: &files.ComparableStrings{
				External: "simple name.",
				Metadata: "simple name.",
			},
			want: false,
		},
		"complex mismatch": {
			cS: &files.ComparableStrings{
				External: "simple_name",
				Metadata: "simple: nam",
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.Id3v2NameDiffers(tt.cS); got != tt.want {
				t.Errorf("Id3v2NameDiffers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeGenre(t *testing.T) {
	type test struct {
		g    string
		want string
	}
	tests := map[string]test{}
	for k, v := range files.GenreMap {
		tests[v] = test{
			g:    fmt.Sprintf("(%d)%s", k, v),
			want: v,
		}
		if v == "Rhythm and Blues" {
			tests["R&B"] = test{
				g:    fmt.Sprintf("(%d)R&B", k),
				want: v,
			}
		}
	}
	tests["prog rock"] = test{
		g:    "prog rock",
		want: "prog rock",
	}
	tests["unexpected k/v"] = test{
		g:    "(256)martian folk rock",
		want: "(256)martian folk rock",
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.NormalizeGenre(tt.g); got != tt.want {
				t.Errorf("NormalizeGenre() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestId3v2GenreDiffers(t *testing.T) {
	tests := map[string]struct {
		cS   *files.ComparableStrings
		want bool
	}{
		"match": {
			cS: &files.ComparableStrings{
				External: "Classic Rock",
				Metadata: "Classic Rock",
			},
			want: false,
		},
		"no match": {
			cS: &files.ComparableStrings{
				External: "Classic Rock",
				Metadata: "classic rock",
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.Id3v2GenreDiffers(tt.cS); got != tt.want {
				t.Errorf("Id3v2GenreDiffers() = %v, want %v", got, tt.want)
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
