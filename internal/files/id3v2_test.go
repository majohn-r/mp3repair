package files

import (
	"bytes"
	"mp3/internal"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bogem/id3v2/v2"
)

func TestNewID3V2TaggedTrackDataForTesting(t *testing.T) {
	fnName := "NewID3V2TaggedTrackDataForTesting()"
	type args struct {
		albumFrame           string
		artistFrame          string
		titleFrame           string
		evaluatedNumberFrame int
		musicCDIdentifier    []byte
	}
	tests := []struct {
		name string
		args
		want *ID3V2TaggedTrackData
	}{
		{
			name: "usual",
			args: args{
				albumFrame:           "the album",
				artistFrame:          "the artist",
				titleFrame:           "the title",
				evaluatedNumberFrame: 1,
				musicCDIdentifier:    []byte{0, 1, 2},
			},
			want: &ID3V2TaggedTrackData{
				album:             "the album",
				artist:            "the artist",
				title:             "the title",
				track:             1,
				musicCDIdentifier: id3v2.UnknownFrame{Body: []byte{0, 1, 2}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewID3V2TaggedTrackDataForTesting(tt.args.albumFrame, tt.args.artistFrame, tt.args.titleFrame, tt.args.evaluatedNumberFrame, tt.args.musicCDIdentifier)
			if got.album != tt.want.album ||
				got.artist != tt.want.artist ||
				got.title != tt.want.title ||
				got.track != tt.want.track ||
				string(got.musicCDIdentifier.Body) != string(tt.want.musicCDIdentifier.Body) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestRawReadID3V2Tag(t *testing.T) {
	fnName := "RawReadID3V2Tag()"
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
	tests := []struct {
		name string
		args
		wantD *ID3V2TaggedTrackData
	}{
		{name: "bad test", args: args{path: "./noSuchFile!.mp3"}, wantD: &ID3V2TaggedTrackData{err: "foo"}},
		{
			name: "good test",
			args: args{path: "./goodFile.mp3"},
			wantD: &ID3V2TaggedTrackData{
				album:  "unknown album",
				artist: "unknown artist",
				title:  "unknown track",
				track:  2,
			},
		},
		{
			name: "bad data test",
			args: args{path: "./badFile.mp3"},
			wantD: &ID3V2TaggedTrackData{
				err: internal.ERROR_DOES_NOT_BEGIN_WITH_DIGIT,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotD := RawReadID3V2Tag(tt.args.path)
			if len(gotD.err) != 0 {
				if len(tt.wantD.err) == 0 {
					t.Errorf("%s = %v, want %v", fnName, gotD, tt.wantD)
				}
			} else if len(tt.wantD.err) != 0 {
				t.Errorf("%s = %v, want %v", fnName, gotD, tt.wantD)
			}
		})
	}
}

func Test_removeLeadingBOMs(t *testing.T) {
	fnName := "removeLeadingBOMs()"
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{name: "normal string", args: args{s: "normal"}, want: "normal"},
		{name: "abnormal string", args: args{s: "\ufeff\ufeffnormal"}, want: "normal"},
		{name: "empty string", args: args{}},
		{name: "nothing but BOM", args: args{s: "\ufeff\ufeff"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeLeadingBOMs(tt.args.s); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func Test_toTrackNumber(t *testing.T) {
	fnName := "toTrackNumber()"
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args
		wantI   int
		wantErr bool
	}{
		{name: "good value", args: args{s: "12"}, wantI: 12, wantErr: false},
		{name: "empty value", args: args{s: ""}, wantI: 0, wantErr: true},
		{name: "negative value", args: args{s: "-12"}, wantI: 0, wantErr: true},
		{name: "invalid value", args: args{s: "foo"}, wantI: 0, wantErr: true},
		{name: "complicated value", args: args{s: "12/39"}, wantI: 12, wantErr: false},
		{name: "BOM-infested complicated value", args: args{s: "\ufeff12/39"}, wantI: 12, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	fnName := "selectUnknownFrame()"
	type args struct {
		mcdiFramers []id3v2.Framer
	}
	tests := []struct {
		name string
		args
		want id3v2.UnknownFrame
	}{
		{
			name: "degenerate case",
			args: args{mcdiFramers: nil},
			want: id3v2.UnknownFrame{Body: []byte{0}},
		},
		{
			name: "too many framers",
			args: args{
				mcdiFramers: []id3v2.Framer{
					id3v2.UnknownFrame{Body: []byte{1, 2, 3}},
					id3v2.UnknownFrame{Body: []byte{4, 5, 6}},
				},
			},
			want: id3v2.UnknownFrame{Body: []byte{0}},
		},
		{
			name: "wrong kind of framer",
			args: args{
				mcdiFramers: []id3v2.Framer{
					unspecifiedFrame{content: "no good"},
				},
			},
			want: id3v2.UnknownFrame{Body: []byte{0}},
		},
		{
			name: "desired use case",
			args: args{
				mcdiFramers: []id3v2.Framer{
					id3v2.UnknownFrame{Body: []byte{0, 1, 2}},
				},
			},
			want: id3v2.UnknownFrame{Body: []byte{0, 1, 2}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selectUnknownFrame(tt.args.mcdiFramers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_updateID3V2Tag(t *testing.T) {
	fnName := "updateID3V2Tag()"
	topDir := "updateID3V2Tag"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, topDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	testFileName := "test.mp3"
	fullPath := filepath.Join(topDir, testFileName)
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
	if err := internal.CreateFileForTestingWithContent(topDir, testFileName, content); err != nil {
		t.Errorf("%s cannot create file %q: %v", fnName, fullPath, err)
	}
	type args struct {
		t *Track
		a taggedTrackState
	}
	tests := []struct {
		name string
		args
		wantErr bool
	}{
		{
			name: "track got deleted!",
			args: args{
				t: &Track{
					number:          1,
					name:            "defective track",
					path:            filepath.Join(topDir, "non-existent-file.mp3"),
					containingAlbum: NewAlbum("poor album", NewArtist("sorry artist", ""), ""),
					ID3V2TaggedTrackData: ID3V2TaggedTrackData{
						track:  1,
						title:  "unknown track",
						album:  "unknown album",
						artist: "unknown artist",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "fixable track",
			args: args{
				t: &Track{
					number:          1,
					name:            "fixable track",
					path:            fullPath,
					containingAlbum: NewAlbum("poor album", NewArtist("sorry artist", ""), ""),
					ID3V2TaggedTrackData: ID3V2TaggedTrackData{
						track:             2,
						title:             "unknown track",
						album:             "unknown album",
						artist:            "unknown artist",
						genre:             "unknown genre",
						year:              "2022",
						musicCDIdentifier: id3v2.UnknownFrame{Body: []byte{1, 2, 3}},
					},
				},
				a: taggedTrackState{
					numberingConflict:  true,
					trackNameConflict:  true,
					albumNameConflict:  true,
					artistNameConflict: true,
					genreConflict:      true,
					yearConflict:       true,
					mcdiConflict:       true,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := updateID3V2Tag(tt.args.t, tt.args.a); (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
			}
		})
	}
	if edited, err := os.ReadFile(fullPath); err != nil {
		t.Errorf("%s cannot read file %q: %v", fnName, fullPath, err)
	} else {
		if tag, err := id3v2.ParseReader(bytes.NewReader(edited), id3v2.Options{Parse: true}); err != nil {
			t.Errorf("%s edited mp3 file %q cannot be read for tags: %v", fnName, fullPath, err)
		} else {
			m := map[string]string{
				// changed by editing
				"TALB": "poor album",
				"TIT2": "fixable track",
				"TPE1": "sorry artist",
				"TRCK": "1",
				"TCON": "",
				"TYER": "",
				// preserved from original file
				"TCOM": "a couple of idiots",
				"TLEN": "1000",
			}
			for key, value := range m {
				if got := tag.GetTextFrame(key).Text; got != value {
					t.Errorf("%s edited mp3 file key %q got %q want %q", fnName, key, got, value)
				}
			}
		}
		// verify "music" is present
		musicStarts := len(edited) - 256
		for k := 0; k < 256; k++ {
			if edited[musicStarts+k] != byte(k) {
				t.Errorf("%s edited mp3 file music at index %d mismatch (%d v. %d)", fnName, k, edited[musicStarts+k], k)
			}
		}
	}
}

func TestID3V2TrackFrame_String(t *testing.T) {
	fnName := "ID3V2TrackFrame.String()"
	tests := []struct {
		name string
		f    *id3v2TrackFrame
		want string
	}{{name: "usual", f: &id3v2TrackFrame{name: "T1", value: "V1"}, want: "T1 = \"V1\""}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.String(); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func Test_readID3V3Metadata(t *testing.T) {
	fnName := "readID3V3Metadata()"
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
	tests := []struct {
		name string
		args
		wantVersion byte
		wantEnc     string
		wantF       []string
		wantErr     bool
	}{
		{
			name:    "error case",
			args:    args{path: "./no such file"},
			wantErr: true,
		},
		{
			name:        "good case",
			args:        args{path: "./goodfile.mp3"},
			wantEnc:     "ISO-8859-1",
			wantVersion: 3,
			wantF: []string{
				"Fake = \"<<[]byte{0x0, 0x75, 0x6d, 0x6d, 0x6d}>>\"",
				"T??? = \"who knows?\"",
				"TALB = \"unknown album\"",
				"TCOM = \"a couple of idiots\"",
				"TCON = \"dance music\"",
				"TIT2 = \"unknown track\"",
				"TLEN = \"1000\"",
				"TPE1 = \"unknown artist\"",
				"TRCK = \"2\"",
				"TYER = \"2022\"",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVersion, gotEnc, gotF, err := readID3V3Metadata(tt.args.path)
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

func Test_stringifyFramerArray(t *testing.T) {
	fnName := "stringifyFramerArray()"
	type args struct {
		f []id3v2.Framer
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "single UnknownFrame",
			args: args{f: []id3v2.Framer{id3v2.UnknownFrame{Body: []byte{0, 1, 2}}}},
			want: "<<[]byte{0x0, 0x1, 0x2}>>",
		},
		{
			name: "single unexpected frame",
			args: args{f: []id3v2.Framer{unspecifiedFrame{content: "hello world"}}},
			want: "<<files.unspecifiedFrame{content:\"hello world\"}>>",
		},
		{
			name: "multiple frames",
			args: args{
				f: []id3v2.Framer{
					id3v2.UnknownFrame{Body: []byte{0, 1, 2}},
					unspecifiedFrame{content: "hello world"},
				},
			},
			want: "<<[0 []byte{0x0, 0x1, 0x2}], [1 files.unspecifiedFrame{content:\"hello world\"}]>>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringifyFramerArray(tt.args.f); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}
