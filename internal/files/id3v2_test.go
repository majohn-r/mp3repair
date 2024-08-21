package files

import (
	"fmt"
	"io"
	"maps"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"testing"

	"github.com/bogem/id3v2/v2"
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/spf13/afero"
)

func makeTextFrame(id, content string) []byte {
	frame := make([]byte, 0, len(id)+len(content)+7)
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

var cannedPayload = makePayload()

func makePayload() []byte {
	payload := make([]byte, 0, 256)
	for k := 0; k < 256; k++ {
		payload = append(payload, byte(k))
	}
	return payload
}

// createID3v2TaggedData creates ID3V2-tagged content. This code is based on reading
// https://id3.org/id3v2.3.0 and on examining hex dumps of real mp3 files.
func createID3v2TaggedData(audio []byte, frames map[string]string) []byte {
	// create text frames; order is fixed for testing
	keys := slices.Collect(maps.Keys(frames))
	sort.Strings(keys)
	frameContents := make([][]byte, 0, len(keys))
	frameLength := 0
	for _, key := range keys {
		frame := makeTextFrame(key, frames[key])
		frameLength += len(frame)
		frameContents = append(frameContents, frame)
	}
	content := make([]byte, 0, 10+frameLength+len(audio))
	// ID3V2 header
	// ID3v2 file identifier   "ID3"
	// ID3v2 version           $03 00
	//                         major version 3 minor version 0, so ID3V2.3.0
	// ID3v2 flags             %abc00000
	//                          a: Unsynchronisation
	//                             Bit 7 in the 'ID3v2 flags' indicates whether unsynchronisation
	//                             is used (see section 5 for details); a set bit indicates usage.
	//                          b: Extended header
	//                             The second bit (bit 6) indicates whether the header is followed
	//                             by an extended header. The extended header is described in
	//                             section 3.2.
	//                          c: Experimental indicator
	//                             The third bit (bit 5) should be used as an 'experimental
	//                             indicator'. This flag should always be set when the tag is in
	//                             an experimental stage.
	// ID3v2 size              4 * %0xxxxxxx
	content = append(content, []byte("ID3")...)
	content = append(content, []byte{3, 0, 0}...)
	factor := 128 * 128 * 128
	for k := 0; k < 4; k++ {
		content = append(content, byte(frameLength/factor))
		frameLength %= factor
		factor /= 128
	}
	for _, frameContent := range frameContents {
		content = append(content, frameContent...)
	}
	// add audio
	content = append(content, audio...)
	return content

}

func Test_rawReadID3V2Metadata(t *testing.T) {
	originalFileSystem := cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmdtoolkit.AssignFileSystem(originalFileSystem)
	}()
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
	content := createID3v2TaggedData(cannedPayload, frames)
	_ = createFileWithContent(".", "goodFile.mp3", content)
	frames["TRCK"] = "oops"
	_ = createFileWithContent(".", "badFile.mp3", createID3v2TaggedData(cannedPayload, frames))
	tests := map[string]struct {
		path  string
		wantD *id3v2Metadata
	}{
		"bad test": {
			path:  "./noSuchFile!.mp3",
			wantD: &id3v2Metadata{err: fmt.Errorf("foo")},
		},
		"good test": {
			path: "./goodFile.mp3",
			wantD: &id3v2Metadata{
				albumTitle:  "unknown album",
				artistName:  "unknown artist",
				trackName:   "unknown track",
				trackNumber: 2,
			},
		},
		"bad data test": {
			path:  "./badFile.mp3",
			wantD: &id3v2Metadata{err: errMalformedTrackNumber},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotD := rawReadID3V2Metadata(tt.path)
			if gotD.err != nil {
				if tt.wantD.err == nil {
					t.Errorf("rawReadID3V2Metadata() = %v, want %v", gotD, tt.wantD)
				}
			} else if tt.wantD.err != nil {
				t.Errorf("rawReadID3V2Metadata() = %v, want %v", gotD, tt.wantD)
			}
		})
	}
}

func Test_removeLeadingBOMs(t *testing.T) {
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
			if got := removeLeadingBOMs(tt.s); got != tt.want {
				t.Errorf("removeLeadingBOMs() = %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_toTrackNumber(t *testing.T) {
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
			gotI, gotErr := toTrackNumber(tt.s)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("toTrackNumber() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if gotErr == nil && gotI != tt.wantI {
				t.Errorf("toTrackNumber() = %d, want %d", gotI, tt.wantI)
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
			if got := selectUnknownFrame(tt.mcdiFramers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("selectUnknownFrame() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_iD3V2TrackFrame_String(t *testing.T) {
	tests := map[string]struct {
		f    *id3v2TrackFrame
		want string
	}{
		"usual": {
			f:    &id3v2TrackFrame{name: "T1", value: []string{"V1"}},
			want: `T1 = ["V1"]`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.f.String(); got != tt.want {
				t.Errorf("id3v2TrackFrame.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_readID3V2Metadata(t *testing.T) {
	originalFileSystem := cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmdtoolkit.AssignFileSystem(originalFileSystem)
	}()
	content := createID3v2TaggedData(cannedPayload, map[string]string{
		"TYER": "2022",
		"TALB": "unknown album",
		"TRCK": "2",
		"TCON": "dance music",
		"TCOM": "a couple of idiots",
		"TIT2": "unknown track",
		"TPE1": "unknown artist",
		"TLEN": "1000",
		"T???": "who knows?",
		"Fake": "huh",
	})
	goodFileName := "goodFile.mp3"
	_ = createFileWithContent(".", goodFileName, content)
	tests := map[string]struct {
		path             string
		wantVersion      byte
		wantEncoding     string
		wantFrameStrings map[string][]string
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
			wantFrameStrings: map[string][]string{
				"Fake": {"00 68 75 68                                     •huh"},
				"T???": {"who knows?"},
				"TALB": {"unknown album"},
				"TCOM": {"a couple of idiots"},
				"TCON": {"dance music"},
				"TIT2": {"unknown track"},
				"TLEN": {"0:01.000"},
				"TPE1": {"unknown artist"},
				"TRCK": {"2"},
				"TYER": {"2022"},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// ignoring raw frames ...
			gotInfo, gotErr := readID3V2Metadata(tt.path)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("readID3V2Metadata() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if gotErr == nil {
				if gotInfo.Version() != tt.wantVersion {
					t.Errorf("readID3V2Metadata() gotInfo.version = %v, want %v",
						gotInfo.Version(), tt.wantVersion)
				}
				if gotInfo.Encoding() != tt.wantEncoding {
					t.Errorf("readID3V2Metadata gotInfo.encoding = %v, want %v",
						gotInfo.Encoding(), tt.wantEncoding)
				}
				if !reflect.DeepEqual(gotInfo.Frames(), tt.wantFrameStrings) {
					t.Errorf("readID3V2Metadata gotInfo.frames = %v, want %v",
						gotInfo.Frames(), tt.wantFrameStrings)
				}
			}
		})
	}
}

var (
	freeRipMCDI = []byte{
		0x01, 0xFF, 0xFE, '2', 0, '0', 0, '0', 0, 'f', 0, 'c', 0, '8', 0, '1', 0, '4', 0, 0, 0,
	}
	freeRipMCDIOutput = []string{
		"200fc814",
		"01 FF FE 32 00 30 00 30 00 66 00 63 00 38 00 31 •••2•0•0•f•c•8•1",
		"00 34 00 00 00                                  •4•••",
	}
	lameMCDI = []byte{
		0x00, 0xAA, 0x01, 0x14, 0x00, 0x10, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x02, 0x00,
		0x00, 0x00, 0x32, 0xF0, 0x00, 0x10, 0x03, 0x00, 0x00, 0x00, 0x5E, 0x93, 0x00, 0x10, 0x04, 0x00,
		0x00, 0x00, 0x91, 0xCE, 0x00, 0x10, 0x05, 0x00, 0x00, 0x00, 0xC6, 0x21, 0x00, 0x10, 0x06, 0x00,
		0x00, 0x00, 0xEB, 0x5D, 0x00, 0x10, 0x07, 0x00, 0x00, 0x01, 0x1C, 0x2C, 0x00, 0x10, 0x08, 0x00,
		0x00, 0x01, 0x47, 0x24, 0x00, 0x10, 0x09, 0x00, 0x00, 0x01, 0x71, 0xFF, 0x00, 0x10, 0x0A, 0x00,
		0x00, 0x01, 0xA2, 0x7E, 0x00, 0x10, 0x0B, 0x00, 0x00, 0x01, 0xCE, 0xF7, 0x00, 0x10, 0x0C, 0x00,
		0x00, 0x01, 0xF9, 0x3F, 0x00, 0x10, 0x0D, 0x00, 0x00, 0x02, 0x26, 0x70, 0x00, 0x10, 0x0E, 0x00,
		0x00, 0x02, 0x6B, 0xF6, 0x00, 0x10, 0x0F, 0x00, 0x00, 0x02, 0xB3, 0xE6, 0x00, 0x10, 0x10, 0x00,
		0x00, 0x03, 0x0B, 0xBD, 0x00, 0x10, 0x11, 0x00, 0x00, 0x03, 0x4D, 0x07, 0x00, 0x10, 0x12, 0x00,
		0x00, 0x03, 0x7C, 0xAC, 0x00, 0x10, 0x13, 0x00, 0x00, 0x03, 0xC6, 0xA1, 0x00, 0x10, 0x14, 0x00,
		0x00, 0x04, 0x0F, 0x19, 0x00, 0x10, 0xAA, 0x00, 0x00, 0x04, 0x9F, 0xB2, 0x0D, 0x00,
	}
	lameMCDIOutput = []string{
		"first track: 1",
		"last track: 20",
		"track 1 logical block address 150",
		"track 2 logical block address 13190",
		"track 3 logical block address 24361",
		"track 4 logical block address 37476",
		"track 5 logical block address 50871",
		"track 6 logical block address 60403",
		"track 7 logical block address 72898",
		"track 8 logical block address 83898",
		"track 9 logical block address 94869",
		"track 10 logical block address 107284",
		"track 11 logical block address 118669",
		"track 12 logical block address 129493",
		"track 13 logical block address 141062",
		"track 14 logical block address 158860",
		"track 15 logical block address 177276",
		"track 16 logical block address 199763",
		"track 17 logical block address 216477",
		"track 18 logical block address 228674",
		"track 19 logical block address 247607",
		"track 20 logical block address 266159",
		"leadout track logical block address 303176",
		"00 AA 01 14 00 10 01 00 00 00 00 00 00 10 02 00 ••••••••••••••••",
		"00 00 32 F0 00 10 03 00 00 00 5E 93 00 10 04 00 ••2•••••••^•••••",
		"00 00 91 CE 00 10 05 00 00 00 C6 21 00 10 06 00 •••••••••••!••••",
		"00 00 EB 5D 00 10 07 00 00 01 1C 2C 00 10 08 00 •••]•••••••,••••",
		"00 01 47 24 00 10 09 00 00 01 71 FF 00 10 0A 00 ••G$••••••q•••••",
		"00 01 A2 7E 00 10 0B 00 00 01 CE F7 00 10 0C 00 •••~••••••••••••",
		"00 01 F9 3F 00 10 0D 00 00 02 26 70 00 10 0E 00 •••?••••••&p••••",
		"00 02 6B F6 00 10 0F 00 00 02 B3 E6 00 10 10 00 ••k•••••••••••••",
		"00 03 0B BD 00 10 11 00 00 03 4D 07 00 10 12 00 ••••••••••M•••••",
		"00 03 7C AC 00 10 13 00 00 03 C6 A1 00 10 14 00 ••|•••••••••••••",
		"00 04 0F 19 00 10 AA 00 00 04 9F B2 0D 00       ••••••••••••••",
	}
	windowsLegacyReaderMCDI = []byte{
		0x31, 0x00, 0x34, 0x00, 0x2B, 0x00, 0x39, 0x00, 0x36, 0x00, 0x2B, 0x00, 0x33, 0x00, 0x33, 0x00,
		0x38, 0x00, 0x36, 0x00, 0x2B, 0x00, 0x35, 0x00, 0x46, 0x00, 0x32, 0x00, 0x39, 0x00, 0x2B, 0x00,
		0x39, 0x00, 0x32, 0x00, 0x36, 0x00, 0x34, 0x00, 0x2B, 0x00, 0x43, 0x00, 0x36, 0x00, 0x42, 0x00,
		0x37, 0x00, 0x2B, 0x00, 0x45, 0x00, 0x42, 0x00, 0x46, 0x00, 0x33, 0x00, 0x2B, 0x00, 0x31, 0x00,
		0x31, 0x00, 0x43, 0x00, 0x43, 0x00, 0x32, 0x00, 0x2B, 0x00, 0x31, 0x00, 0x34, 0x00, 0x37, 0x00,
		0x42, 0x00, 0x41, 0x00, 0x2B, 0x00, 0x31, 0x00, 0x37, 0x00, 0x32, 0x00, 0x39, 0x00, 0x35, 0x00,
		0x2B, 0x00, 0x31, 0x00, 0x41, 0x00, 0x33, 0x00, 0x31, 0x00, 0x34, 0x00, 0x2B, 0x00, 0x31, 0x00,
		0x43, 0x00, 0x46, 0x00, 0x38, 0x00, 0x44, 0x00, 0x2B, 0x00, 0x31, 0x00, 0x46, 0x00, 0x39, 0x00,
		0x44, 0x00, 0x35, 0x00, 0x2B, 0x00, 0x32, 0x00, 0x32, 0x00, 0x37, 0x00, 0x30, 0x00, 0x36, 0x00,
		0x2B, 0x00, 0x32, 0x00, 0x36, 0x00, 0x43, 0x00, 0x38, 0x00, 0x43, 0x00, 0x2B, 0x00, 0x32, 0x00,
		0x42, 0x00, 0x34, 0x00, 0x37, 0x00, 0x43, 0x00, 0x2B, 0x00, 0x33, 0x00, 0x30, 0x00, 0x43, 0x00,
		0x35, 0x00, 0x33, 0x00, 0x2B, 0x00, 0x33, 0x00, 0x34, 0x00, 0x44, 0x00, 0x39, 0x00, 0x44, 0x00,
		0x2B, 0x00, 0x33, 0x00, 0x37, 0x00, 0x44, 0x00, 0x34, 0x00, 0x32, 0x00, 0x2B, 0x00, 0x33, 0x00,
		0x43, 0x00, 0x37, 0x00, 0x33, 0x00, 0x37, 0x00, 0x2B, 0x00, 0x34, 0x00, 0x30, 0x00, 0x46, 0x00,
		0x41, 0x00, 0x46, 0x00, 0x2B, 0x00, 0x34, 0x00, 0x41, 0x00, 0x30, 0x00, 0x34, 0x00, 0x38, 0x00,
		0x00, 0x00,
	}
	windowsLegacyReaderMCDIString = "14+96+3386+5F29+9264+C6B7+EBF3+11CC2+147BA+17295+1A314+1CF8D+1F9D5+22706+26C8C+" +
		"2B47C+30C53+34D9D+37D42+3C737+40FAF+4A048"
	windowsLegacyReaderMCDIOutput = []string{
		"tracks 20",
		"track 1 logical block address 150",
		"track 2 logical block address 13190",
		"track 3 logical block address 24361",
		"track 4 logical block address 37476",
		"track 5 logical block address 50871",
		"track 6 logical block address 60403",
		"track 7 logical block address 72898",
		"track 8 logical block address 83898",
		"track 9 logical block address 94869",
		"track 10 logical block address 107284",
		"track 11 logical block address 118669",
		"track 12 logical block address 129493",
		"track 13 logical block address 141062",
		"track 14 logical block address 158860",
		"track 15 logical block address 177276",
		"track 16 logical block address 199763",
		"track 17 logical block address 216477",
		"track 18 logical block address 228674",
		"track 19 logical block address 247607",
		"track 20 logical block address 266159",
		"leadout track logical block address 303176",
		"31 00 34 00 2B 00 39 00 36 00 2B 00 33 00 33 00 1•4•+•9•6•+•3•3•",
		"38 00 36 00 2B 00 35 00 46 00 32 00 39 00 2B 00 8•6•+•5•F•2•9•+•",
		"39 00 32 00 36 00 34 00 2B 00 43 00 36 00 42 00 9•2•6•4•+•C•6•B•",
		"37 00 2B 00 45 00 42 00 46 00 33 00 2B 00 31 00 7•+•E•B•F•3•+•1•",
		"31 00 43 00 43 00 32 00 2B 00 31 00 34 00 37 00 1•C•C•2•+•1•4•7•",
		"42 00 41 00 2B 00 31 00 37 00 32 00 39 00 35 00 B•A•+•1•7•2•9•5•",
		"2B 00 31 00 41 00 33 00 31 00 34 00 2B 00 31 00 +•1•A•3•1•4•+•1•",
		"43 00 46 00 38 00 44 00 2B 00 31 00 46 00 39 00 C•F•8•D•+•1•F•9•",
		"44 00 35 00 2B 00 32 00 32 00 37 00 30 00 36 00 D•5•+•2•2•7•0•6•",
		"2B 00 32 00 36 00 43 00 38 00 43 00 2B 00 32 00 +•2•6•C•8•C•+•2•",
		"42 00 34 00 37 00 43 00 2B 00 33 00 30 00 43 00 B•4•7•C•+•3•0•C•",
		"35 00 33 00 2B 00 33 00 34 00 44 00 39 00 44 00 5•3•+•3•4•D•9•D•",
		"2B 00 33 00 37 00 44 00 34 00 32 00 2B 00 33 00 +•3•7•D•4•2•+•3•",
		"43 00 37 00 33 00 37 00 2B 00 34 00 30 00 46 00 C•7•3•7•+•4•0•F•",
		"41 00 46 00 2B 00 34 00 41 00 30 00 34 00 38 00 A•F•+•4•A•0•4•8•",
		"00 00                                           ••",
	}
	samplePictureFrame = id3v2.PictureFrame{
		Encoding:    id3v2.EncodingISO,
		MimeType:    "image/jpeg",
		PictureType: 3,
		Description: "CD Front Cover",
		Picture: []byte{
			0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01, 0x01, 0x01, 0x00, 0x11,
		},
	}
	samplePictureFrameOutput = []string{
		"Encoding: ISO-8859-1",
		"Mime Type: image/jpeg",
		"Picture Type: Cover (front)",
		"Description: CD Front Cover",
		"Picture Data:",
		"FF D8 FF E0 00 10 4A 46 49 46 00 01 01 01 00 11 ••••••JFIF••••••",
	}
)

func Test_framerSliceAsString(t *testing.T) {
	tests := map[string]struct {
		f    []id3v2.Framer
		want []string
	}{
		"freeRip MCDI": {
			f:    []id3v2.Framer{id3v2.UnknownFrame{Body: freeRipMCDI}},
			want: freeRipMCDIOutput,
		},
		"lame-generated MCDI": {
			f:    []id3v2.Framer{id3v2.UnknownFrame{Body: lameMCDI}},
			want: lameMCDIOutput,
		},
		"window legacy reader MCDI": {
			f:    []id3v2.Framer{id3v2.UnknownFrame{Body: windowsLegacyReaderMCDI}},
			want: windowsLegacyReaderMCDIOutput,
		},
		"APIC": {
			f:    []id3v2.Framer{samplePictureFrame},
			want: samplePictureFrameOutput,
		},
		"unrecognized string MCDI": {
			f: []id3v2.Framer{id3v2.UnknownFrame{Body: []byte{
				0x31, 0x00, 0x34, 0x00, 0x2B, 0x00, 0x39, 0x00, 0x36, 0x00, 0x2B, 0x00, 0x33, 0x00, 0x33, 0x00,
			}}},
			want: []string{
				"14+96+33",
				"31 00 34 00 2B 00 39 00 36 00 2B 00 33 00 33 00 1•4•+•9•6•+•3•3•",
			},
		},
		"single UnknownFrame": {
			f:    []id3v2.Framer{id3v2.UnknownFrame{Body: []byte{0, 1, 2}}},
			want: []string{"00 01 02                                        •••"},
		},
		"unexpected frame": {
			f:    []id3v2.Framer{unspecifiedFrame{content: "hello world"}},
			want: []string{"files.unspecifiedFrame{content:\"hello world\"}"},
		},
		"multiple frames": {
			f: []id3v2.Framer{
				id3v2.UnknownFrame{Body: []byte{0, 1, 2}},
				unspecifiedFrame{content: "hello world"},
			},
			want: []string{
				"00 01 02                                        •••",
				"files.unspecifiedFrame{content:\"hello world\"}",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := framerSliceAsString(tt.f); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("framerSliceAsString() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_id3v2NameDiffers(t *testing.T) {
	tests := map[string]struct {
		cS   *comparableStrings
		want bool
	}{
		"identical strings": {
			cS: &comparableStrings{
				external: "simple name",
				metadata: "simple name",
			},
			want: false,
		},
		"identical strings with case differences": {
			cS: &comparableStrings{
				external: "SIMPLE name",
				metadata: "simple NAME",
			},
			want: false,
		},
		"strings of different length": {
			cS: &comparableStrings{
				external: "simple name",
				metadata: "artist: simple name",
			},
			want: true,
		},
		"use of runes that are illegal for file names": {
			cS: &comparableStrings{
				external: "simple_name",
				metadata: "simple:name",
			},
			want: false,
		},
		"metadata with trailing space": {
			cS: &comparableStrings{
				external: "simple name",
				metadata: "simple name ",
			},
			want: false,
		},
		"period on the end": {
			cS: &comparableStrings{
				external: "simple name.",
				metadata: "simple name.",
			},
			want: false,
		},
		"complex mismatch": {
			cS: &comparableStrings{
				external: "simple_name",
				metadata: "simple: nam",
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := id3v2NameDiffers(tt.cS); got != tt.want {
				t.Errorf("id3v2NameDiffers() = %v, want %v", got, tt.want)
			}
		})
	}
}

var lcGenres = map[int]string{
	0:   "blues",
	1:   "classic rock",
	2:   "country",
	3:   "dance",
	4:   "disco",
	5:   "funk",
	6:   "grunge",
	7:   "hip-hop",
	8:   "jazz",
	9:   "metal",
	10:  "new age",
	11:  "oldies",
	12:  "other",
	13:  "pop",
	14:  "r&b",
	15:  "rap",
	16:  "reggae",
	17:  "rock",
	18:  "techno",
	19:  "industrial",
	20:  "alternative",
	21:  "ska",
	22:  "death metal",
	23:  "pranks",
	24:  "soundtrack",
	25:  "euro-techno",
	26:  "ambient",
	27:  "trip-hop",
	28:  "vocal",
	29:  "jazz & funk",
	30:  "fusion",
	31:  "trance",
	32:  "classical",
	33:  "instrumental",
	34:  "acid",
	35:  "house",
	36:  "game",
	37:  "sound clip",
	38:  "gospel",
	39:  "noise",
	40:  "alternative rock",
	41:  "bass",
	42:  "soul",
	43:  "punk",
	44:  "space",
	45:  "meditative",
	46:  "instrumental pop",
	47:  "instrumental rock",
	48:  "ethnic",
	49:  "gothic",
	50:  "darkwave",
	51:  "techno-industrial",
	52:  "electronic",
	53:  "pop-folk",
	54:  "eurodance",
	55:  "dream",
	56:  "southern rock",
	57:  "comedy",
	58:  "cult",
	59:  "gangsta",
	60:  "top 40",
	61:  "christian rap",
	62:  "pop/funk",
	63:  "jungle music",
	64:  "native us",
	65:  "cabaret",
	66:  "new wave",
	67:  "psychedelic",
	68:  "rave",
	69:  "showtunes",
	70:  "trailer",
	71:  "lo-fi",
	72:  "tribal",
	73:  "acid punk",
	74:  "acid jazz",
	75:  "polka",
	76:  "retro",
	77:  "musical",
	78:  "rock ’n’ roll",
	79:  "hard rock",
	80:  "folk",
	81:  "folk-rock",
	82:  "national folk",
	83:  "swing",
	84:  "fast fusion",
	85:  "bebop",
	86:  "latin",
	87:  "revival",
	88:  "celtic",
	89:  "bluegrass",
	90:  "avantgarde",
	91:  "gothic rock",
	92:  "progressive rock",
	93:  "psychedelic rock",
	94:  "symphonic rock",
	95:  "slow rock",
	96:  "big band",
	97:  "chorus",
	98:  "easy listening",
	99:  "acoustic",
	100: "humour",
	101: "speech",
	102: "chanson",
	103: "opera",
	104: "chamber music",
	105: "sonata",
	106: "symphony",
	107: "booty bass",
	108: "primus",
	109: "porn groove",
	110: "satire",
	111: "slow jam",
	112: "club",
	113: "tango",
	114: "samba",
	115: "folklore",
	116: "ballad",
	117: "power ballad",
	118: "rhythmic soul",
	119: "freestyle",
	120: "duet",
	121: "punk rock",
	122: "drum solo",
	123: "a cappella",
	124: "euro-house",
	125: "dance hall",
	126: "goa music",
	127: "drum & bass",
	128: "club-house",
	129: "hardcore techno",
	130: "terror",
	131: "indie",
	132: "britpop",
	133: "negerpunk",
	134: "polsk punk",
	135: "beat",
	136: "christian gangsta rap",
	137: "heavy metal",
	138: "black metal",
	139: "crossover",
	140: "contemporary christian",
	141: "christian rock",
	142: "merengue",
	143: "salsa",
	144: "thrash metal",
	145: "anime",
	146: "jpop",
	147: "synthpop",
	148: "abstract",
	149: "art rock",
	150: "baroque",
	151: "bhangra",
	152: "big beat",
	153: "breakbeat",
	154: "chillout",
	155: "downtempo",
	156: "dub",
	157: "ebm",
	158: "eclectic",
	159: "electro",
	160: "electroclash",
	161: "emo",
	162: "experimental",
	163: "garage",
	164: "global",
	165: "idm",
	166: "illbient",
	167: "industro-goth",
	168: "jam band",
	169: "krautrock",
	170: "leftfield",
	171: "lounge",
	172: "math rock",
	173: "new romantic",
	174: "nu-breakz",
	175: "post-punk",
	176: "post-rock",
	177: "psytrance",
	178: "shoegaze",
	179: "space rock",
	180: "trop rock",
	181: "world music",
	182: "neoclassical",
	183: "audiobook",
	184: "audio theatre",
	185: "neue deutsche welle",
	186: "podcast",
	187: "indie-rock",
	188: "g-funk",
	189: "dubstep",
	190: "garage rock",
	191: "psybient",
}

func Test_normalizeGenre(t *testing.T) {
	type test struct {
		g    string
		want string
	}
	tests := map[string]test{}
	for k, v := range lcGenres {
		tests[v] = test{
			g:    fmt.Sprintf("(%d)%s", k, v),
			want: v,
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
			if got := normalizeGenre(tt.g); got != tt.want {
				t.Errorf("normalizeGenre() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_id3v2GenreDiffers(t *testing.T) {
	tests := map[string]struct {
		cS   *comparableStrings
		want bool
	}{
		"match": {
			cS: &comparableStrings{
				external: "Classic Rock",
				metadata: "Classic Rock",
			},
			want: false,
		},
		"no match": {
			cS: &comparableStrings{
				external: "Classic Rock",
				metadata: "classic rock",
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := id3v2GenreDiffers(tt.cS); got != tt.want {
				t.Errorf("id3v2GenreDiffers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isTagAbsent(t *testing.T) {
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
			if got := isTagAbsent(tt.tag); got != tt.want {
				t.Errorf("isTagAbsent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hexDump(t *testing.T) {
	tests := map[string]struct {
		content []byte
		want    []string
	}{
		"empty": {
			content: nil,
			want:    []string{},
		},
		"short": {
			content: []byte{0x00},
			want:    []string{"00                                              •"},
		},
		"evenly divisible by 16": {
			content: []byte{
				16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
				48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63,
				80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95,
				112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127,
			},
			want: []string{
				"10 11 12 13 14 15 16 17 18 19 1A 1B 1C 1D 1E 1F ••••••••••••••••",
				"30 31 32 33 34 35 36 37 38 39 3A 3B 3C 3D 3E 3F 0123456789:;<=>?",
				"50 51 52 53 54 55 56 57 58 59 5A 5B 5C 5D 5E 5F PQRSTUVWXYZ[\\]^_",
				"70 71 72 73 74 75 76 77 78 79 7A 7B 7C 7D 7E 7F pqrstuvwxyz{|}~•",
			},
		},
		"long": {
			content: []byte{
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
				32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47,
				48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63,
				64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79,
				80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95,
				96, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111,
				112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127,
				128, 129, 130,
			},
			want: []string{
				"00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F ••••••••••••••••",
				"10 11 12 13 14 15 16 17 18 19 1A 1B 1C 1D 1E 1F ••••••••••••••••",
				"20 21 22 23 24 25 26 27 28 29 2A 2B 2C 2D 2E 2F  !\"#$%&'()*+,-./",
				"30 31 32 33 34 35 36 37 38 39 3A 3B 3C 3D 3E 3F 0123456789:;<=>?",
				"40 41 42 43 44 45 46 47 48 49 4A 4B 4C 4D 4E 4F @ABCDEFGHIJKLMNO",
				"50 51 52 53 54 55 56 57 58 59 5A 5B 5C 5D 5E 5F PQRSTUVWXYZ[\\]^_",
				"60 61 62 63 64 65 66 67 68 69 6A 6B 6C 6D 6E 6F `abcdefghijklmno",
				"70 71 72 73 74 75 76 77 78 79 7A 7B 7C 7D 7E 7F pqrstuvwxyz{|}~•",
				"80 81 82                                        •••",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := hexDump(tt.content); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("hexDump() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_displayString(t *testing.T) {
	tests := map[string]struct {
		content []byte
		want    string
		want1   bool
	}{
		"odd length": {
			content: []byte{'A'},
			want:    "",
			want1:   false,
		},
		"odd bytes are not all null": {
			content: []byte{'A', 'A'},
			want:    "",
			want1:   false,
		},
		"typical, including trailing null": {
			content: []byte{'A', 0, 'F', 0, '+', 0, '4', 0, 'A', 0, '0', 0, '4', 0, '8', 0, 0, 0},
			want:    "AF+4A048",
			want1:   true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, got1 := displayString(tt.content)
			if got != tt.want {
				t.Errorf("displayString() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("displayString() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_decodeFreeRipMCDI(t *testing.T) {
	tests := map[string]struct {
		content []byte
		want    []string
		want1   bool
	}{
		"too short": {
			content: []byte{1, 2},
			want:    nil,
			want1:   false,
		},
		"invalid key": {
			content: []byte{1, 2, 3},
			want:    nil,
			want1:   false,
		},
		"invalid content": {
			content: []byte{1, 0xff, 0xfe, '2'},
			want:    nil,
			want1:   false,
		},
		"valid content": {
			content: freeRipMCDI,
			want:    freeRipMCDIOutput,
			want1:   true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, got1 := decodeFreeRipMCDI(tt.content)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeFreeRipMCDI() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("decodeFreeRipMCDI() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_decodeLAMEGeneratedMCDI(t *testing.T) {
	tests := map[string]struct {
		content []byte
		want    []string
		want1   bool
	}{
		"too short": {
			content: []byte{1, 2, 3},
			want:    nil,
			want1:   false,
		},
		"length field too big": {
			content: []byte{0, 4, 1, 2},
			want:    nil,
			want1:   false,
		},
		"tracks inconsistent with content": {
			content: []byte{0, 4, 1, 2, 3},
			want:    nil,
			want1:   false,
		},
		"proper LAME output": {
			content: lameMCDI,
			want:    lameMCDIOutput,
			want1:   true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, got1 := decodeLAMEGeneratedMCDI(tt.content)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeLAMEGeneratedMCDI() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("decodeLAMEGeneratedMCDI() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_decodeWindowsLegacyMediaPlayerMCDI(t *testing.T) {
	type args struct {
		s   string
		raw []byte
	}
	tests := map[string]struct {
		args
		want  []string
		want1 bool
	}{
		"no match": {
			args: args{
				s:   "not a match",
				raw: []byte{},
			},
			want:  nil,
			want1: false,
		},
		"insufficient addresses": {
			args: args{
				s:   "20+0+1",
				raw: []byte{},
			},
			want:  nil,
			want1: false,
		},
		"good data": {
			args: args{
				s:   windowsLegacyReaderMCDIString,
				raw: windowsLegacyReaderMCDI,
			},
			want:  windowsLegacyReaderMCDIOutput,
			want1: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, got1 := decodeWindowsLegacyMediaPlayerMCDI(tt.args.s, tt.args.raw)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeWindowsLegacyMediaPlayerMCDI() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("decodeWindowsLegacyMediaPlayerMCDI() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_interpretPictureFrame(t *testing.T) {
	tests := map[string]struct {
		frame *id3v2.PictureFrame
		want  []string
	}{
		"typical": {
			frame: &samplePictureFrame,
			want:  samplePictureFrameOutput,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := interpretPictureFrame(tt.frame); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("interpretPictureFrame() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_interpretPictureType(t *testing.T) {
	tests := map[string]struct {
		pictureType byte
		want        string
	}{
		"typical": {
			pictureType: 3,
			want:        "Cover (front)",
		},
		"atypical": {
			pictureType: 0xff,
			want:        "Undocumented value 255",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := interpretPictureType(tt.pictureType); got != tt.want {
				t.Errorf("interpretPictureType() = %v, want %v", got, tt.want)
			}
		})
	}
}
