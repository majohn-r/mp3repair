package files

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/spf13/afero"
)

// values per https://id3.org/ID3v1 as of August 16, 2022
const (
	nameLength = 30
	// always 'TAG' if present
	tagOffset = 0
	tagLength = 3
	// first 30 characters of the track title
	titleOffset = tagOffset + tagLength
	titleLength = nameLength
	// first 30 characters of the artist name
	artistOffset = titleOffset + titleLength
	artistLength = nameLength
	// first 30 characters of the album name
	albumOffset = artistOffset + artistLength
	albumLength = nameLength
	// four digit year, e.g., '2', '0', '2', '2'
	yearOffset = albumOffset + albumLength
	yearLength = 4
	// comment, rarely used, and not interesting
	commentOffset = yearOffset + yearLength
	commentLength = 28
	// should always be zero; if not, then the track is not valid
	zeroByteOffset = commentOffset + commentLength
	zeroByteLength = 1
	// track number; if zeroByte is not zero, not valid
	trackOffset = zeroByteOffset + zeroByteLength
	trackLength = 1
	// genre list index
	genreOffset = trackOffset + trackLength
	genreLength = 1
	// total length of the ID3V1 block
	id3v1Length = genreOffset + genreLength
)

type id3v1Field struct {
	startOffset int
	length      int
	endOffset   int
}

type BiDirectionalMap[K comparable, V comparable] struct {
	k2v        map[K]V
	v2k        map[V]K
	kNormalize func(K) K
	vNormalize func(V) V
}

func NewBiDirectionalMap[K comparable, V comparable](keyNormalize func(K) K, valueNormalize func(V) V) *BiDirectionalMap[K, V] {
	bdMap := &BiDirectionalMap[K, V]{
		k2v:        map[K]V{},
		v2k:        map[V]K{},
		kNormalize: keyNormalize,
		vNormalize: valueNormalize,
	}
	if keyNormalize == nil {
		bdMap.kNormalize = func(k K) K {
			return k
		}
	}
	if valueNormalize == nil {
		bdMap.vNormalize = func(v V) V {
			return v
		}
	}
	return bdMap
}

func (bdMap *BiDirectionalMap[K, V]) LookupKey(key K) (V, bool) {
	v, found := bdMap.k2v[bdMap.kNormalize(key)]
	return v, found
}

func (bdMap *BiDirectionalMap[K, V]) LookupValue(value V) (K, bool) {
	k, found := bdMap.v2k[bdMap.vNormalize(value)]
	return k, found
}

var Genres *BiDirectionalMap[int, string]

func normalizeGenreNames(name string) string {
	n := strings.ToLower(name)
	if n == "rhythm and blues" {
		return "r&b"
	}
	return n
}

func Populate(m map[int]string) (*BiDirectionalMap[int, string], error) {
	g := NewBiDirectionalMap[int, string](nil, normalizeGenreNames)
	for k, v := range m {
		if err := g.AddPair(k, v); err != nil {
			return nil, err
		}
	}
	return g, nil
}

func (bdMap *BiDirectionalMap[K, V]) AddPair(k K, v V) error {
	normalizedV := bdMap.vNormalize(v)
	normalizedK := bdMap.kNormalize(k)
	if _, exists := bdMap.v2k[normalizedV]; exists {
		return fmt.Errorf("value %v exists", normalizedV)
	}
	if _, exists := bdMap.k2v[normalizedK]; exists {
		return fmt.Errorf("key %v exists", normalizedK)
	}
	bdMap.k2v[normalizedK] = normalizedV
	bdMap.v2k[normalizedV] = normalizedK
	return nil
}

func (bdMap *BiDirectionalMap[K, V]) KeyMap() map[K]V {
	return bdMap.k2v
}

func (bdMap *BiDirectionalMap[K, V]) ValueMap() map[V]K {
	return bdMap.v2k
}

func InitGenres() error {
	if Genres == nil {
		Genres, _ = Populate(documentedGenres)
	}
	return nil
}

func genreIndex(genreName string) (int, bool) {
	_ = InitGenres()
	return Genres.LookupValue(genreName)
}

func genreName(genreIndex int) (string, bool) {
	_ = InitGenres()
	return Genres.LookupKey(genreIndex)
}

var (
	// per https://en.wikipedia.org/wiki/List_of_ID3v1_Genres as of August 16, 2022
	documentedGenres = map[int]string{
		0:   "Blues",
		1:   "Classic Rock",
		2:   "Country",
		3:   "Dance",
		4:   "Disco",
		5:   "Funk",
		6:   "Grunge",
		7:   "Hip-Hop",
		8:   "Jazz",
		9:   "Metal",
		10:  "New Age",
		11:  "Oldies",
		12:  "Other",
		13:  "Pop",
		14:  "Rhythm and Blues",
		15:  "Rap",
		16:  "Reggae",
		17:  "Rock",
		18:  "Techno",
		19:  "Industrial",
		20:  "Alternative",
		21:  "Ska",
		22:  "Death Metal",
		23:  "Pranks",
		24:  "Soundtrack",
		25:  "Euro-Techno",
		26:  "Ambient",
		27:  "Trip-Hop",
		28:  "Vocal",
		29:  "Jazz & Funk",
		30:  "Fusion",
		31:  "Trance",
		32:  "Classical",
		33:  "Instrumental",
		34:  "Acid",
		35:  "House",
		36:  "Game",
		37:  "Sound clip",
		38:  "Gospel",
		39:  "Noise",
		40:  "Alternative Rock",
		41:  "Bass",
		42:  "Soul",
		43:  "Punk",
		44:  "Space",
		45:  "Meditative",
		46:  "Instrumental Pop",
		47:  "Instrumental Rock",
		48:  "Ethnic",
		49:  "Gothic",
		50:  "Darkwave",
		51:  "Techno-Industrial",
		52:  "Electronic",
		53:  "Pop-Folk",
		54:  "Eurodance",
		55:  "Dream",
		56:  "Southern Rock",
		57:  "Comedy",
		58:  "Cult",
		59:  "Gangsta",
		60:  "Top 40",
		61:  "Christian Rap",
		62:  "Pop/Funk",
		63:  "Jungle music",
		64:  "Native US",
		65:  "Cabaret",
		66:  "New Wave",
		67:  "Psychedelic",
		68:  "Rave",
		69:  "Showtunes",
		70:  "Trailer",
		71:  "Lo-Fi",
		72:  "Tribal",
		73:  "Acid Punk",
		74:  "Acid Jazz",
		75:  "Polka",
		76:  "Retro",
		77:  "Musical",
		78:  "Rock ’n’ Roll",
		79:  "Hard Rock",
		80:  "Folk",
		81:  "Folk-Rock",
		82:  "National Folk",
		83:  "Swing",
		84:  "Fast Fusion",
		85:  "Bebop",
		86:  "Latin",
		87:  "Revival",
		88:  "Celtic",
		89:  "Bluegrass",
		90:  "Avantgarde",
		91:  "Gothic Rock",
		92:  "Progressive Rock",
		93:  "Psychedelic Rock",
		94:  "Symphonic Rock",
		95:  "Slow Rock",
		96:  "Big Band",
		97:  "Chorus",
		98:  "Easy Listening",
		99:  "Acoustic",
		100: "Humour",
		101: "Speech",
		102: "Chanson",
		103: "Opera",
		104: "Chamber Music",
		105: "Sonata",
		106: "Symphony",
		107: "Booty Bass",
		108: "Primus",
		109: "Porn Groove",
		110: "Satire",
		111: "Slow Jam",
		112: "Club",
		113: "Tango",
		114: "Samba",
		115: "Folklore",
		116: "Ballad",
		117: "Power Ballad",
		118: "Rhythmic Soul",
		119: "Freestyle",
		120: "Duet",
		121: "Punk Rock",
		122: "Drum Solo",
		123: "A cappella",
		124: "Euro-House",
		125: "Dance Hall",
		126: "Goa music",
		127: "Drum & Bass",
		128: "Club-House",
		129: "Hardcore Techno",
		130: "Terror",
		131: "Indie",
		132: "BritPop",
		133: "Negerpunk",
		134: "Polsk Punk",
		135: "Beat",
		136: "Christian Gangsta Rap",
		137: "Heavy Metal",
		138: "Black Metal",
		139: "Crossover",
		140: "Contemporary Christian",
		141: "Christian Rock",
		142: "Merengue",
		143: "Salsa",
		144: "Thrash Metal",
		145: "Anime",
		146: "Jpop",
		147: "Synthpop",
		148: "Abstract",
		149: "Art Rock",
		150: "Baroque",
		151: "Bhangra",
		152: "Big beat",
		153: "Breakbeat",
		154: "Chillout",
		155: "Downtempo",
		156: "Dub",
		157: "EBM",
		158: "Eclectic",
		159: "Electro",
		160: "Electroclash",
		161: "Emo",
		162: "Experimental",
		163: "Garage",
		164: "Global",
		165: "IDM",
		166: "Illbient",
		167: "Industro-Goth",
		168: "Jam Band",
		169: "Krautrock",
		170: "Leftfield",
		171: "Lounge",
		172: "Math Rock",
		173: "New Romantic",
		174: "Nu-Breakz",
		175: "Post-Punk",
		176: "Post-Rock",
		177: "Psytrance",
		178: "Shoegaze",
		179: "Space Rock",
		180: "Trop Rock",
		181: "World Music",
		182: "Neoclassical",
		183: "Audiobook",
		184: "Audio Theatre",
		185: "Neue Deutsche Welle",
		186: "Podcast",
		187: "Indie-Rock",
		188: "G-Funk",
		189: "Dubstep",
		190: "Garage Rock",
		191: "Psybient",
	}
	runeByteMapping = map[rune][]byte{
		'…': {0x85},
		'¡': {0xA1},
		'¢': {0xA2},
		'£': {0xA3},
		'¤': {0xA4},
		'¥': {0xA5},
		'¦': {0xA6},
		'§': {0xA7},
		'¨': {0xA8},
		'©': {0xA9},
		'ª': {0xAA},
		'«': {0xAB},
		'¬': {0xAC},
		'®': {0xAE},
		'¯': {0xAF},
		'°': {0xB0},
		'±': {0xB1},
		'²': {0xB2},
		'³': {0xB3},
		'´': {0xB4},
		'µ': {0xB5},
		'¶': {0xB6},
		'·': {0xB7},
		'¸': {0xB8},
		'¹': {0xB9},
		'º': {0xBA},
		'»': {0xBB},
		'¼': {0xBC},
		'½': {0xBD},
		'¾': {0xBE},
		'¿': {0xBF},
		'À': {0xC0},
		'Á': {0xC1},
		'Â': {0xC2},
		'Ã': {0xC3},
		'Ä': {0xC4},
		'Å': {0xC5},
		'Æ': {0xC6},
		'Ç': {0xC7},
		'È': {0xC8},
		'É': {0xC9},
		'Ê': {0xCA},
		'Ë': {0xCB},
		'Ì': {0xCC},
		'Í': {0xCD},
		'Î': {0xCE},
		'Ï': {0xCF},
		'Ð': {0xD0},
		'Ñ': {0xD1},
		'Ò': {0xD2},
		'Ó': {0xD3},
		'Ô': {0xD4},
		'Õ': {0xD5},
		'Ö': {0xD6},
		'×': {0xD7},
		'Ø': {0xD8},
		'Ù': {0xD9},
		'Ú': {0xDA},
		'Û': {0xDB},
		'Ü': {0xDC},
		'Ý': {0xDD},
		'Þ': {0xDE},
		'ß': {0xDF},
		'à': {0xE0},
		'á': {0xE1},
		'â': {0xE2},
		'ã': {0xE3},
		'ä': {0xE4},
		'å': {0xE5},
		'æ': {0xE6},
		'ç': {0xE7},
		'è': {0xE8},
		'é': {0xE9},
		'ê': {0xEA},
		'ë': {0xEB},
		'ì': {0xEC},
		'í': {0xED},
		'î': {0xEE},
		'ï': {0xEF},
		'ñ': {0xF1},
		'ò': {0xF2},
		'ó': {0xF3},
		'ô': {0xF4},
		'õ': {0xF5},
		'ö': {0xF6},
		'÷': {0xF7},
		'ø': {0xF8},
		'ù': {0xF9},
		'ú': {0xFA},
		'û': {0xFB},
		'ü': {0xFC},
		'ý': {0xFD},
		'þ': {0xFE},
		'ÿ': {0xFF},
		'Ā': {'A'},      // Latin Capital letter A with macron
		'ā': {'a'},      // Latin Small letter A with macron
		'Ă': {'A'},      // Latin Capital letter A with breve
		'ă': {'a'},      // Latin Small letter A with breve
		'Ą': {'A'},      // Latin Capital letter A with ogonek
		'ą': {'a'},      // Latin Small letter A with ogonek
		'Ć': {'C'},      // Latin Capital letter C with acute
		'ć': {'c'},      // Latin Small letter C with acute
		'Ĉ': {'C'},      // Latin Capital letter C with circumflex
		'ĉ': {'c'},      // Latin Small letter C with circumflex
		'Ċ': {'C'},      // Latin Capital letter C with dot above
		'ċ': {'c'},      // Latin Small letter C with dot above
		'Č': {'C'},      // Latin Capital letter C with caron
		'č': {'c'},      // Latin Small letter C with caron
		'Ď': {'D'},      // Latin Capital letter D with caron
		'ď': {'d'},      // Latin Small letter D with caron
		'Đ': {'D'},      // Latin Capital letter D with stroke
		'đ': {'d'},      // Latin Small letter D with stroke
		'Ē': {'E'},      // Latin Capital letter E with macron
		'ē': {'e'},      // Latin Small letter E with macron
		'Ĕ': {'E'},      // Latin Capital letter E with breve
		'ĕ': {'e'},      // Latin Small letter E with breve
		'Ė': {'E'},      // Latin Capital letter E with dot above
		'ė': {'e'},      // Latin Small letter E with dot above
		'Ę': {'E'},      // Latin Capital letter E with ogonek
		'ę': {'e'},      // Latin Small letter E with ogonek
		'Ě': {'E'},      // Latin Capital letter E with caron
		'ě': {'e'},      // Latin Small letter E with caron
		'Ĝ': {'G'},      // Latin Capital letter G with circumflex
		'ĝ': {'g'},      // Latin Small letter G with circumflex
		'Ğ': {'G'},      // Latin Capital letter G with breve
		'ğ': {'g'},      // Latin Small letter G with breve
		'Ġ': {'G'},      // Latin Capital letter G with dot above
		'ġ': {'g'},      // Latin Small letter G with dot above
		'Ģ': {'G'},      // Latin Capital letter G with cedilla
		'ģ': {'g'},      // Latin Small letter G with cedilla
		'Ĥ': {'H'},      // Latin Capital letter H with circumflex
		'ĥ': {'h'},      // Latin Small letter H with circumflex
		'Ħ': {'H'},      // Latin Capital letter H with stroke
		'ħ': {'h'},      // Latin Small letter H with stroke
		'Ĩ': {'I'},      // Latin Capital letter I with tilde
		'ĩ': {'i'},      // Latin Small letter I with tilde
		'Ī': {'I'},      // Latin Capital letter I with macron
		'ī': {'i'},      // Latin Small letter I with macron
		'Ĭ': {'I'},      // Latin Capital letter I with breve
		'ĭ': {'i'},      // Latin Small letter I with breve
		'Į': {'I'},      // Latin Capital letter I with ogonek
		'į': {'i'},      // Latin Small letter I with ogonek
		'İ': {'I'},      // Latin Capital letter I with dot above
		'ı': {'i'},      // Latin Small letter dotless I
		'Ĳ': {'I', 'J'}, // Latin Capital Ligature IJ
		'ĳ': {'i', 'j'}, // Latin Small Ligature IJ
		'Ĵ': {'J'},      // Latin Capital letter J with circumflex
		'ĵ': {'j'},      // Latin Small letter J with circumflex
		'Ķ': {'K'},      // Latin Capital letter K with cedilla
		'ķ': {'k'},      // Latin Small letter K with cedilla
		'ĸ': {'k'},      // Latin Small letter Kra
		'Ĺ': {'L'},      // Latin Capital letter L with acute
		'ĺ': {'l'},      // Latin Small letter L with acute
		'Ļ': {'L'},      // Latin Capital letter L with cedilla
		'ļ': {'l'},      // Latin Small letter L with cedilla
		'Ľ': {'L'},      // Latin Capital letter L with caron
		'ľ': {'l'},      // Latin Small letter L with caron
		'Ŀ': {'L'},      // Latin Capital letter L with middle dot
		'ŀ': {'l'},      // Latin Small letter L with middle dot
		'Ł': {'L'},      // Latin Capital letter L with stroke
		'ł': {'L'},      // Latin Small letter L with stroke
		'Ń': {'N'},      // Latin Capital letter N with acute
		'ń': {'n'},      // Latin Small letter N with acute
		'Ņ': {'N'},      // Latin Capital letter N with cedilla
		'ņ': {'n'},      // Latin Small letter N with cedilla
		'Ň': {'N'},      // Latin Capital letter N with caron
		'ň': {'n'},      // Latin Small letter N with caron
		'ŉ': {'n'},      // Latin Small letter N preceded by apostrophe
		'Ŋ': {'N', 'G'}, // Latin Capital letter Eng
		'ŋ': {'n', 'g'}, // Latin Small letter Eng
		'Ō': {'O'},      // Latin Capital letter O with macron
		'ō': {'o'},      // Latin Small letter O with macron
		'Ŏ': {'O'},      // Latin Capital letter O with breve
		'ŏ': {'o'},      // Latin Small letter O with breve
		'Ő': {'O'},      // Latin Capital Letter O with double acute
		'ő': {'o'},      // Latin Small Letter O with double acute
		'Œ': {'O', 'E'}, // Latin Capital Ligature OE
		'œ': {'o', 'e'}, // Latin Small Ligature OE
		'Ŕ': {'R'},      // Latin Capital letter R with acute
		'ŕ': {'r'},      // Latin Small letter R with acute
		'Ŗ': {'R'},      // Latin Capital letter R with cedilla
		'ŗ': {'t'},      // Latin Small letter R with cedilla
		'Ř': {'R'},      // Latin Capital letter R with caron
		'ř': {'r'},      // Latin Small letter R with caron
		'Ś': {'S'},      // Latin Capital letter S with acute
		'ś': {'s'},      // Latin Small letter S with acute
		'Ŝ': {'S'},      // Latin Capital letter S with circumflex
		'ŝ': {'s'},      // Latin Small letter S with circumflex
		'Ş': {'S'},      // Latin Capital letter S with cedilla
		'ş': {'s'},      // Latin Small letter S with cedilla
		'Š': {'S'},      // Latin Capital letter S with caron
		'š': {'s'},      // Latin Small letter S with caron
		'Ţ': {'T'},      // Latin Capital letter T with cedilla
		'ţ': {'t'},      // Latin Small letter T with cedilla
		'Ť': {'T'},      // Latin Capital letter T with caron
		'ť': {'t'},      // Latin Small letter T with caron
		'Ŧ': {'T'},      // Latin Capital letter T with stroke
		'ŧ': {'t'},      // Latin Small letter T with stroke
		'Ũ': {'U'},      // Latin Capital letter U with tilde
		'ũ': {'u'},      // Latin Small letter U with tilde
		'Ū': {'U'},      // Latin Capital letter U with macron
		'ū': {'u'},      // Latin Small letter U with macron
		'Ŭ': {'U'},      // Latin Capital letter U with breve
		'ŭ': {'u'},      // Latin Small letter U with breve
		'Ů': {'U'},      // Latin Capital letter U with ring above
		'ů': {'u'},      // Latin Small letter U with ring above
		'Ű': {'U'},      // Latin Capital Letter U with double acute
		'ű': {'u'},      // Latin Small Letter U with double acute
		'Ų': {'U'},      // Latin Capital letter U with ogonek
		'ų': {'u'},      // Latin Small letter U with ogonek
		'Ŵ': {'W'},      // Latin Capital letter W with circumflex
		'ŵ': {'w'},      // Latin Small letter W with circumflex
		'Ŷ': {'Y'},      // Latin Capital letter Y with circumflex
		'ŷ': {'y'},      // Latin Small letter Y with circumflex
		'Ÿ': {'Y'},      // Latin Capital letter Y with diaeresis
		'Ź': {'Z'},      // Latin Capital letter Z with acute
		'ź': {'z'},      // Latin Small letter Z with acute
		'Ż': {'Z'},      // Latin Capital letter Z with dot above
		'ż': {'z'},      // Latin Small letter Z with dot above
		'Ž': {'Z'},      // Latin Capital letter Z with caron
		'ž': {'z'},      // Latin Small letter Z with caron
		'ſ': {'S'},      // Latin Small letter long S
	}
	// TagField is the initial field for ID3V1 metadata, and should contain 'TAG'
	TagField = initID3v1Field(tagOffset, tagLength)
	// these are the remaining fields making up ID3V1 metadata
	titleField    = initID3v1Field(titleOffset, titleLength)
	artistField   = initID3v1Field(artistOffset, artistLength)
	albumField    = initID3v1Field(albumOffset, albumLength)
	yearField     = initID3v1Field(yearOffset, yearLength)
	commentField  = initID3v1Field(commentOffset, commentLength)
	zeroByteField = initID3v1Field(zeroByteOffset, zeroByteLength)
	trackField    = initID3v1Field(trackOffset, trackLength)
	genreField    = initID3v1Field(genreOffset, genreLength)

	ErrNoID3V1MetadataFound = fmt.Errorf("no ID3V1 metadata found")
)

func initID3v1Field(offset, length int) id3v1Field {
	return id3v1Field{startOffset: offset, length: length, endOffset: offset + length}
}

type Id3v1Metadata struct {
	data []byte
}

func (im *Id3v1Metadata) RawData() []byte {
	return im.data
}

func (im *Id3v1Metadata) WithData(b []byte) *Id3v1Metadata {
	im.data = make([]byte, id3v1Length)
	switch {
	case len(b) >= id3v1Length:
		for k := 0; k < id3v1Length; k++ {
			im.data[k] = b[k]
		}
	default:
		copy(im.data, b)
		for k := len(b); k < id3v1Length; k++ {
			im.data[k] = 0
		}
	}
	return im
}

func NewID3v1Metadata() *Id3v1Metadata {
	return &Id3v1Metadata{data: make([]byte, id3v1Length)}
}

func (im *Id3v1Metadata) readString(f id3v1Field) string {
	return Trim(string(im.data[f.startOffset:f.endOffset]))
}

func (im *Id3v1Metadata) IsValid() bool {
	return im.readString(TagField) == "TAG"
}

func (im *Id3v1Metadata) Title() string {
	return im.readString(titleField)
}

func (im *Id3v1Metadata) WriteString(s string, f id3v1Field) {
	copy(im.data[f.startOffset:f.endOffset], bytes.Repeat([]byte{0}, f.length))
	// truncate long strings ...
	if len(s) > f.length {
		s = s[0:f.length]
	}
	copy(im.data[f.startOffset:f.endOffset], s)
}

func repairName(s string) string {
	bs := make([]byte, 0, 2*len(s))
	for _, r := range s {
		b, mappingFound := runeByteMapping[r]
		switch {
		case mappingFound:
			bs = append(bs, b...)
		default:
			bs = append(bs, byte(r))
		}
	}
	return string(bs)
}

func (im *Id3v1Metadata) SetTitle(s string) {
	im.WriteString(repairName(s), titleField)
}

func (im *Id3v1Metadata) Artist() string {
	return im.readString(artistField)
}

func (im *Id3v1Metadata) SetArtist(s string) {
	im.WriteString(repairName(s), artistField)
}

func (im *Id3v1Metadata) Album() string {
	return im.readString(albumField)
}

func (im *Id3v1Metadata) SetAlbum(s string) {
	im.WriteString(repairName(s), albumField)
}

func (im *Id3v1Metadata) Year() string {
	return im.readString(yearField)
}

func (im *Id3v1Metadata) SetYear(s string) {
	im.WriteString(s, yearField)
}

func (im *Id3v1Metadata) Comment() string {
	return im.readString(commentField)
}

func (im *Id3v1Metadata) SetComment(s string) {
	im.WriteString(s, commentField)
}

func (im *Id3v1Metadata) readInt(f id3v1Field) int {
	return int(im.data[f.startOffset])
}

func (im *Id3v1Metadata) Track() (trackNumber int, trackNumberValid bool) {
	if im.readInt(zeroByteField) == 0 {
		trackNumber = im.readInt(trackField)
		trackNumberValid = true
	}
	return
}

func (im *Id3v1Metadata) writeInt(v int, f id3v1Field) {
	im.data[f.startOffset] = byte(v)
}

func (im *Id3v1Metadata) SetTrack(t int) (b bool) {
	if t >= 1 && t <= 255 {
		im.writeInt(0, zeroByteField)
		im.writeInt(t, trackField)
		b = true
	}
	return
}

func (im *Id3v1Metadata) Genre() (string, bool) {
	genre, genreFound := genreName(im.readInt(genreField))
	return genre, genreFound
}

func (im *Id3v1Metadata) SetGenre(s string) {
	index, found := genreIndex(s)
	if !found {
		v, _ := genreIndex("other")
		im.writeInt(v, genreField)
		return
	}
	im.writeInt(index, genreField)
}

func Trim(s string) string {
	for s != "" && (s[len(s)-1:] == " " || s[len(s)-1:] == "\u0000") {
		s = s[:len(s)-1]
	}
	return s
}

func ReadID3v1Metadata(path string) ([]string, error) {
	v1, readErr := InternalReadID3V1Metadata(path, FileReader)
	if readErr != nil {
		return nil, readErr
	}
	output := make([]string, 0, 5)
	output = append(output, fmt.Sprintf("Artist: %q", v1.Artist()),
		fmt.Sprintf("Album: %q", v1.Album()), fmt.Sprintf("Title: %q", v1.Title()))
	if track, trackNumberValid := v1.Track(); trackNumberValid {
		output = append(output, fmt.Sprintf("Track: %d", track))
	}
	output = append(output, fmt.Sprintf("Year: %q", v1.Year()))
	if genre, genreFound := v1.Genre(); genreFound {
		output = append(output, fmt.Sprintf("Genre: %q", genre))
	}
	if comment := v1.Comment(); comment != "" {
		output = append(output, fmt.Sprintf("Comment: %q", comment))
	}
	return output, nil
}

func FileReader(f afero.File, b []byte) (int, error) {
	return f.Read(b)
}

func InternalReadID3V1Metadata(path string, readFunc func(f afero.File, b []byte) (int, error)) (*Id3v1Metadata, error) {
	file, fileErr := cmdtoolkit.FileSystem().Open(path)
	if fileErr != nil {
		return nil, fileErr
	}
	defer func() {
		_ = file.Close()
	}()
	fileStat, _ := file.Stat()
	if fileStat != nil && fileStat.Size() < id3v1Length {
		return nil, ErrNoID3V1MetadataFound
	}
	_, _ = file.Seek(-id3v1Length, io.SeekEnd)
	v1 := NewID3v1Metadata()
	r, readErr := readFunc(file, v1.data)
	if readErr != nil {
		return nil, readErr
	}
	if r < id3v1Length {
		return nil,
			fmt.Errorf("cannot read id3v1 metadata from file %q; only %d bytes read", path, r)
	}
	if !v1.IsValid() {
		return nil, ErrNoID3V1MetadataFound
	}
	return v1, nil
}

func (im *Id3v1Metadata) Write(path string) error {
	return im.InternalWrite(path, WriteToFile)
}

func WriteToFile(f afero.File, b []byte) (int, error) {
	return f.Write(b)
}

func updateID3V1TrackMetadata(tm *TrackMetadata, path string) error {
	const src = ID3V1
	if !tm.EditRequired(src) {
		return nil
	}
	var v1 *Id3v1Metadata
	var fileErr error
	v1, fileErr = InternalReadID3V1Metadata(path, FileReader)
	if fileErr != nil {
		return fileErr
	}
	if artistName := tm.ArtistName(src).Correction(); artistName != "" {
		v1.SetArtist(artistName)
	}
	if albumName := tm.AlbumName(src).Correction(); albumName != "" {
		v1.SetAlbum(albumName)
	}
	if albumGenre := tm.AlbumGenre(src).Correction(); albumGenre != "" {
		v1.SetGenre(albumGenre)
	}
	if albumYear := tm.AlbumYear(src).Correction(); albumYear != "" {
		v1.SetYear(albumYear)
	}
	if trackName := tm.TrackName(src).Correction(); trackName != "" {
		v1.SetTitle(trackName)
	}
	if trackNumber := tm.TrackNumber(src).Correction(); trackNumber != 0 {
		_ = v1.SetTrack(trackNumber)
	}
	return v1.Write(path)
}

func (im *Id3v1Metadata) InternalWrite(path string,
	writeFunc func(f afero.File, b []byte) (int, error)) (fileErr error) {
	fS := cmdtoolkit.FileSystem()
	var src afero.File
	if src, fileErr = fS.Open(path); fileErr == nil {
		defer func() {
			_ = src.Close()
		}()
		var stat fs.FileInfo
		if stat, fileErr = src.Stat(); fileErr == nil {
			tmpPath := path + "-id3v1"
			var tmpFile afero.File
			if tmpFile, fileErr = fS.OpenFile(tmpPath, os.O_RDWR|os.O_CREATE, stat.Mode()); fileErr == nil {
				defer func() {
					_ = tmpFile.Close()
				}()
				// borrowed this piece of logic from id3v2 tag.Save() method
				tempFileShouldBeRemoved := true
				defer func() {
					if tempFileShouldBeRemoved {
						_ = fS.Remove(tmpPath)
					}
				}()
				if _, fileErr = io.Copy(tmpFile, src); fileErr == nil {
					_ = src.Close()
					fileInfo, _ := fS.Stat(tmpPath)
					if fileInfo != nil && fileInfo.Size() < id3v1Length {
						fileErr = fmt.Errorf("file %q is too short", tmpPath)
						return
					}
					if _, fileErr = tmpFile.Seek(-id3v1Length, io.SeekEnd); fileErr == nil {
						var n int
						if n, fileErr = writeFunc(tmpFile, im.data); fileErr == nil {
							_ = tmpFile.Close()
							if n != id3v1Length {
								fileErr = fmt.Errorf(
									"wrote %d bytes to %q, expected to write %d bytes", n,
									tmpPath, id3v1Length)
								return
							}
							if fileErr = fS.Rename(tmpPath, path); fileErr == nil {
								tempFileShouldBeRemoved = false
							}
						}
					}
				}
			}
		}
	}
	return
}

func Id3v1NameDiffers(cS *ComparableStrings) bool {
	bs := make([]byte, 0, 2*len(cS.External))
	for _, r := range strings.ToLower(cS.External) {
		b, mappingFound := runeByteMapping[r]
		switch {
		case mappingFound:
			bs = append(bs, b...)
		default:
			bs = append(bs, byte(r))
		}
	}
	if len(bs) > nameLength {
		bs = bs[:nameLength]
	}
	for bs[len(bs)-1] == ' ' {
		bs = bs[:len(bs)-1]
	}
	metadataRunes := []rune(strings.ToLower(cS.Metadata))
	externalRunes := []rune(string(bs))
	if len(metadataRunes) != len(externalRunes) {
		return true
	}
	for i, c := range metadataRunes {
		if externalRunes[i] == c {
			continue
		}
		// allow for the metadata rune to be one that is illegal for file names:
		// the external name is likely to be a file name
		if !IsIllegalRuneForFileNames(c) {
			return true
		}
	}
	return false
}

func Id3v1GenreDiffers(cS *ComparableStrings) bool {
	if _, genreFound := genreIndex(cS.External); !genreFound {
		// the external genre does not map to a known id3v1 genre but "other"
		// always matches the external name
		if cS.Metadata == "other" {
			return false
		}
	}
	// external name is a known id3v1 genre, or metadata name is not "other"
	return !strings.EqualFold(cS.External, cS.Metadata)
}
