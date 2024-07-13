package files

import (
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/bogem/id3v2/v2"
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
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
	originalFileSystem := cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmdtoolkit.AssignFileSystem(originalFileSystem)
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
	_ = createFileWithContent(".", "goodFile.mp3", content)
	frames["TRCK"] = "oops"
	_ = createFileWithContent(".", "badFile.mp3", createID3v2TaggedData(payload, frames))
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
			if gotD.hasError() {
				if !tt.wantD.hasError() {
					t.Errorf("rawReadID3V2Metadata() = %v, want %v", gotD, tt.wantD)
				}
			} else if tt.wantD.hasError() {
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
			f:    &id3v2TrackFrame{name: "T1", value: "V1"},
			want: "T1 = \"V1\"",
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
		"Fake": "huh",
	}
	content := createID3v2TaggedData(payload, frames)
	goodFileName := "goodFile.mp3"
	_ = createFileWithContent(".", goodFileName, content)
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
				`Fake = "<<[]byte{0x0, 0x68, 0x75, 0x68}>>"`,
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
			gotInfo, gotErr := readID3V2Metadata(tt.path)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("readID3V2Metadata() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if gotErr == nil {
				if gotInfo.Version != tt.wantVersion {
					t.Errorf("readID3V2Metadata() gotInfo.Version = %v, want %v",
						gotInfo.Version, tt.wantVersion)
				}
				if gotInfo.Encoding != tt.wantEncoding {
					t.Errorf("readID3V2Metadata gotInfo.Encoding = %v, want %v",
						gotInfo.Encoding, tt.wantEncoding)
				}
				if !reflect.DeepEqual(gotInfo.FrameStrings, tt.wantFrameStrings) {
					t.Errorf("readID3V2Metadata gotInfo.FrameStrings = %v, want %v",
						gotInfo.FrameStrings, tt.wantFrameStrings)
				}
			}
		})
	}
}

func Test_framerSliceAsString(t *testing.T) {
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
			want: "<<files.unspecifiedFrame{content:\"hello world\"}>>",
		},
		"multiple frames": {
			f: []id3v2.Framer{
				id3v2.UnknownFrame{Body: []byte{0, 1, 2}},
				unspecifiedFrame{content: "hello world"},
			},
			want: "<<[0 []byte{0x0, 0x1, 0x2}]," +
				" [1 files.unspecifiedFrame{content:\"hello world\"}]>>",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := framerSliceAsString(tt.f); got != tt.want {
				t.Errorf("framerSliceAsString() = %q, want %q", got, tt.want)
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
