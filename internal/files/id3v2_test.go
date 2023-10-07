package files

import (
	"fmt"
	"mp3/internal"
	"os"

	"reflect"
	"testing"

	"github.com/bogem/id3v2/v2"
)

func Test_rawReadID3V2Metadata(t *testing.T) {
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
	content := CreateID3V2TaggedDataForTesting(payload, frames)
	if err := internal.CreateFileForTestingWithContent(".", "goodFile.mp3", content); err != nil {
		t.Errorf("%s failed to create ./goodFile.mp3: %v", fnName, err)
	}
	frames["TRCK"] = "oops"
	if err := internal.CreateFileForTestingWithContent(".", "badFile.mp3", CreateID3V2TaggedDataForTesting(payload, frames)); err != nil {
		t.Errorf("%s failed to create ./badFile.mp3: %v", fnName, err)
	}
	defer func() {
		if err := os.Remove("./goodFile.mp3"); err != nil {
			t.Errorf("%s failed to delete ./goodFile.mp3: %v", fnName, err)
		}
		if err := os.Remove("./badFile.mp3"); err != nil {
			t.Errorf("%s failed to delete ./badFile.mp3: %v", fnName, err)
		}
	}()
	type args struct {
		path string
	}
	tests := map[string]struct {
		args
		wantD *id3v2Metadata
	}{
		"bad test": {args: args{path: "./noSuchFile!.mp3"}, wantD: &id3v2Metadata{err: fmt.Errorf("foo")}},
		"good test": {
			args: args{path: "./goodFile.mp3"},
			wantD: &id3v2Metadata{
				album:  "unknown album",
				artist: "unknown artist",
				title:  "unknown track",
				track:  2,
			},
		},
		"bad data test": {
			args:  args{path: "./badFile.mp3"},
			wantD: &id3v2Metadata{err: errMalformedTrackNumber},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotD := rawReadID3V2Metadata(tt.args.path)
			if gotD.err != nil {
				if tt.wantD.err == nil {
					t.Errorf("%s = %v, want %v", fnName, gotD, tt.wantD)
				}
			} else if tt.wantD.err != nil {
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
			if got := removeLeadingBOMs(tt.args.s); got != tt.want {
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
		"good value":                     {args: args{s: "12"}, wantI: 12, wantErr: false},
		"empty value":                    {args: args{s: ""}, wantI: 0, wantErr: true},
		"BOM-infested empty value":       {args: args{s: "\ufeff"}, wantI: 0, wantErr: true},
		"negative value":                 {args: args{s: "-12"}, wantI: 0, wantErr: true},
		"invalid value":                  {args: args{s: "foo"}, wantI: 0, wantErr: true},
		"complicated value":              {args: args{s: "12/39"}, wantI: 12, wantErr: false},
		"BOM-infested complicated value": {args: args{s: "\ufeff12/39"}, wantI: 12, wantErr: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotI, err := toTrackNumber(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
				return
			}
			if err == nil && gotI != tt.wantI {
				t.Errorf("%s = %d, want %d", fnName, gotI, tt.wantI)
			}
		})
	}
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
		"degenerate case": {args: args{mcdiFramers: nil}, want: id3v2.UnknownFrame{Body: []byte{0}}},
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
			if got := selectUnknownFrame(tt.args.mcdiFramers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestID3V2TrackFrame_String(t *testing.T) {
	const fnName = "ID3V2TrackFrame.String()"
	tests := map[string]struct {
		f    *id3v2TrackFrame
		want string
	}{"usual": {f: &id3v2TrackFrame{name: "T1", value: "V1"}, want: "T1 = \"V1\""}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.f.String(); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func Test_readID3V2Metadata(t *testing.T) {
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
	content := CreateID3V2TaggedDataForTesting(payload, frames)
	if err := internal.CreateFileForTestingWithContent(".", "goodFile.mp3", content); err != nil {
		t.Errorf("%s failed to create ./goodFile.mp3: %v", fnName, err)
	}
	defer func() {
		if err := os.Remove("./goodFile.mp3"); err != nil {
			t.Errorf("%s failed to delete ./goodFile.mp3: %v", fnName, err)
		}
	}()
	type args struct {
		path string
	}
	tests := map[string]struct {
		args
		wantVersion byte
		wantEnc     string
		wantF       []string
		wantErr     bool
	}{
		"error case": {args: args{path: "./no such file"}, wantErr: true},
		"good case": {
			args:        args{path: "./goodfile.mp3"},
			wantEnc:     "ISO-8859-1",
			wantVersion: 3,
			wantF: []string{
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
			gotVersion, gotEnc, gotF, _, err := readID3V2Metadata(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
				return
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("%s gotVersion = %v, want %v", fnName, gotVersion, tt.wantVersion)
			}
			if gotEnc != tt.wantEnc {
				t.Errorf("%s gotEnc = %v, want %v", fnName, gotEnc, tt.wantEnc)
			}
			if !reflect.DeepEqual(gotF, tt.wantF) {
				t.Errorf("%s gotF = %v, want %v", fnName, gotF, tt.wantF)
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
		"single UnknownFrame": {args: args{f: []id3v2.Framer{id3v2.UnknownFrame{Body: []byte{0, 1, 2}}}}, want: "<<[]byte{0x0, 0x1, 0x2}>>"},
		"unexpected frame":    {args: args{f: []id3v2.Framer{unspecifiedFrame{content: "hello world"}}}, want: "<<files.unspecifiedFrame{content:\"hello world\"}>>"},
		"multiple frames": {
			args: args{f: []id3v2.Framer{id3v2.UnknownFrame{Body: []byte{0, 1, 2}}, unspecifiedFrame{content: "hello world"}}},
			want: "<<[0 []byte{0x0, 0x1, 0x2}], [1 files.unspecifiedFrame{content:\"hello world\"}]>>",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := framerSliceAsString(tt.args.f); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func Test_id3v2NameDiffers(t *testing.T) {
	const fnName = "id3v2NameDiffers()"
	type args struct {
		cS comparableStrings
	}
	tests := map[string]struct {
		args
		want bool
	}{
		"identical strings":                            {args: args{comparableStrings{externalName: "simple name", metadataName: "simple name"}}, want: false},
		"identical strings with case differences":      {args: args{comparableStrings{externalName: "SIMPLE name", metadataName: "simple NAME"}}, want: false},
		"strings of different length":                  {args: args{comparableStrings{externalName: "simple name", metadataName: "artist: simple name"}}, want: true},
		"use of runes that are illegal for file names": {args: args{comparableStrings{externalName: "simple_name", metadataName: "simple:name"}}, want: false},
		"metadata with trailing space":                 {args: args{comparableStrings{externalName: "simple name", metadataName: "simple name "}}, want: false},
		"period on the end":                            {args: args{comparableStrings{externalName: "simple name.", metadataName: "simple name."}}, want: false},
		"complex mismatch":                             {args: args{comparableStrings{externalName: "simple_name", metadataName: "simple: nam"}}, want: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := id3v2NameDiffers(tt.args.cS); got != tt.want {
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
	for k, v := range genreMap {
		tests[v] = test{args: args{g: fmt.Sprintf("(%d)%s", k, v)}, want: v}
		if v == "Rhythm and Blues" {
			tests["R&B"] = test{args: args{g: fmt.Sprintf("(%d)R&B", k)}, want: v}
		}
	}
	tests["prog rock"] = test{args: args{g: "prog rock"}, want: "prog rock"}
	tests["unexpected k/v"] = test{args: args{g: "(256)martian folk rock"}, want: "(256)martian folk rock"}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := normalizeGenre(tt.args.g); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_id3v2GenreDiffers(t *testing.T) {
	const fnName = "id3v2GenreDiffers()"
	type args struct {
		cS comparableStrings
	}
	tests := map[string]struct {
		args
		want bool
	}{
		"match":    {args: args{cS: comparableStrings{externalName: "Classic Rock", metadataName: "Classic Rock"}}, want: false},
		"no match": {args: args{cS: comparableStrings{externalName: "Classic Rock", metadataName: "classic rock"}}, want: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := id3v2GenreDiffers(tt.args.cS); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}
