/*
Copyright © 2026 Marc Johnson (marc.johnson27591@gmail.com)
*/
package files

import (
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

type biDirectionalMap[K comparable, V comparable] struct {
	k2v        map[K]V
	v2k        map[V]K
	kNormalize func(K) K
	vNormalize func(V) V
}

func newBiDirectionalMap[K comparable, V comparable](
	keyNormalize func(K) K,
	valueNormalize func(V) V,
) *biDirectionalMap[K, V] {
	bdMap := &biDirectionalMap[K, V]{
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

func (bdMap *biDirectionalMap[K, V]) lookupKey(key K) (V, bool) {
	v, found := bdMap.k2v[bdMap.kNormalize(key)]
	return v, found
}

func (bdMap *biDirectionalMap[K, V]) lookupValue(value V) (K, bool) {
	k, found := bdMap.v2k[bdMap.vNormalize(value)]
	return k, found
}

var genres *biDirectionalMap[int, string]

func normalizeGenreNames(name string) string {
	n := strings.ToLower(name)
	if n == "rhythm and blues" {
		return "r&b"
	}
	return n
}

func populate(m map[int]string) (*biDirectionalMap[int, string], error) {
	g := newBiDirectionalMap[int, string](nil, normalizeGenreNames)
	for k, v := range m {
		if err := g.addPair(k, v); err != nil {
			return nil, err
		}
	}
	return g, nil
}

func (bdMap *biDirectionalMap[K, V]) addPair(k K, v V) error {
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

func initGenres() error {
	if genres == nil {
		genres, _ = populate(documentedGenres)
	}
	return nil
}

func genreIndex(genreName string) (int, bool) {
	_ = initGenres()
	return genres.lookupValue(genreName)
}

func genreName(genreIndex int) (string, bool) {
	_ = initGenres()
	return genres.lookupKey(genreIndex)
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
		'…': []byte("..."),
		'¡': {'!'},
		'¢': []byte("cent"),
		'£': []byte("pound"),
		'¤': []byte("currency"),
		'¥': []byte("yen"),
		'¦': {'|'},
		'§': []byte("ss"),
		'¨': []byte("umlaut"),
		'©': []byte("(C)"),
		'ª': {'a'},
		'«': []byte("<<"),
		'¬': []byte("not"),
		'®': []byte("(R)"),
		'¯': {'-'},
		'°': []byte("degree"),
		'±': []byte("+/-"),
		'²': {'2'},
		'³': {'3'},
		'´': {'\''},
		'µ': []byte("nano"),
		'¶': []byte("pp"),
		'·': {'.'},
		'¸': {'?'},
		'¹': {'?'},
		'º': {'o'},
		'»': []byte(">>"),
		'¼': []byte("1/4"),
		'½': []byte("1/2"),
		'¾': []byte("3/4"),
		'¿': {'?'},
		'À': {'A'},
		'Á': {'A'},
		'Â': {'A'},
		'Ã': {'A'},
		'Ä': {'A'},
		'Å': {'A'},
		'Æ': []byte("AE"),
		'Ç': {'C'},
		'È': {'E'},
		'É': {'E'},
		'Ê': {'E'},
		'Ë': {'E'},
		'Ì': {'I'},
		'Í': {'I'},
		'Î': {'I'},
		'Ï': {'I'},
		'Ð': {'D'},
		'Ñ': {'N'},
		'Ò': {'O'},
		'Ó': {'O'},
		'Ô': {'O'},
		'Õ': {'O'},
		'Ö': {'O'},
		'×': {'x'},
		'Ø': {'O'},
		'Ù': {'U'},
		'Ú': {'U'},
		'Û': {'U'},
		'Ü': {'U'},
		'Ý': {'Y'},
		'Þ': []byte("TH"),
		'ß': {'S'},
		'à': {'a'},
		'á': {'a'},
		'â': {'a'},
		'ã': {'a'},
		'ä': {'a'},
		'å': {'a'},
		'æ': []byte("ae"),
		'ç': {'c'},
		'è': {'e'},
		'é': {'e'},
		'ê': {'e'},
		'ë': {'e'},
		'ì': {'i'},
		'í': {'i'},
		'î': {'i'},
		'ï': {'i'},
		'ñ': {'n'},
		'ò': {'o'},
		'ó': {'o'},
		'ô': {'o'},
		'õ': {'o'},
		'ö': {'o'},
		'÷': {'/'},
		'ø': {'o'},
		'ù': {'u'},
		'ú': {'u'},
		'û': {'u'},
		'ü': {'u'},
		'ý': {'y'},
		'þ': []byte("th"),
		'ÿ': {'y'},
		'Ā': {'A'},        // Latin Capital letter A with macron
		'ā': {'a'},        // Latin Small letter A with macron
		'Ă': {'A'},        // Latin Capital letter A with breve
		'ă': {'a'},        // Latin Small letter A with breve
		'Ą': {'A'},        // Latin Capital letter A with ogonek
		'ą': {'a'},        // Latin Small letter A with ogonek
		'Ć': {'C'},        // Latin Capital letter C with acute
		'ć': {'c'},        // Latin Small letter C with acute
		'Ĉ': {'C'},        // Latin Capital letter C with circumflex
		'ĉ': {'c'},        // Latin Small letter C with circumflex
		'Ċ': {'C'},        // Latin Capital letter C with dot above
		'ċ': {'c'},        // Latin Small letter C with dot above
		'Č': {'C'},        // Latin Capital letter C with caron
		'č': {'c'},        // Latin Small letter C with caron
		'Ď': {'D'},        // Latin Capital letter D with caron
		'ď': {'d'},        // Latin Small letter D with caron
		'Đ': {'D'},        // Latin Capital letter D with stroke
		'đ': {'d'},        // Latin Small letter D with stroke
		'Ē': {'E'},        // Latin Capital letter E with macron
		'ē': {'e'},        // Latin Small letter E with macron
		'Ĕ': {'E'},        // Latin Capital letter E with breve
		'ĕ': {'e'},        // Latin Small letter E with breve
		'Ė': {'E'},        // Latin Capital letter E with dot above
		'ė': {'e'},        // Latin Small letter E with dot above
		'Ę': {'E'},        // Latin Capital letter E with ogonek
		'ę': {'e'},        // Latin Small letter E with ogonek
		'Ě': {'E'},        // Latin Capital letter E with caron
		'ě': {'e'},        // Latin Small letter E with caron
		'Ĝ': {'G'},        // Latin Capital letter G with circumflex
		'ĝ': {'g'},        // Latin Small letter G with circumflex
		'Ğ': {'G'},        // Latin Capital letter G with breve
		'ğ': {'g'},        // Latin Small letter G with breve
		'Ġ': {'G'},        // Latin Capital letter G with dot above
		'ġ': {'g'},        // Latin Small letter G with dot above
		'Ģ': {'G'},        // Latin Capital letter G with cedilla
		'ģ': {'g'},        // Latin Small letter G with cedilla
		'Ĥ': {'H'},        // Latin Capital letter H with circumflex
		'ĥ': {'h'},        // Latin Small letter H with circumflex
		'Ħ': {'H'},        // Latin Capital letter H with stroke
		'ħ': {'h'},        // Latin Small letter H with stroke
		'Ĩ': {'I'},        // Latin Capital letter I with tilde
		'ĩ': {'i'},        // Latin Small letter I with tilde
		'Ī': {'I'},        // Latin Capital letter I with macron
		'ī': {'i'},        // Latin Small letter I with macron
		'Ĭ': {'I'},        // Latin Capital letter I with breve
		'ĭ': {'i'},        // Latin Small letter I with breve
		'Į': {'I'},        // Latin Capital letter I with ogonek
		'į': {'i'},        // Latin Small letter I with ogonek
		'İ': {'I'},        // Latin Capital letter I with dot above
		'ı': {'i'},        // Latin Small letter dotless I
		'Ĳ': []byte("IJ"), // Latin Capital Ligature IJ
		'ĳ': []byte("ij"), // Latin Small Ligature IJ
		'Ĵ': {'J'},        // Latin Capital letter J with circumflex
		'ĵ': {'j'},        // Latin Small letter J with circumflex
		'Ķ': {'K'},        // Latin Capital letter K with cedilla
		'ķ': {'k'},        // Latin Small letter K with cedilla
		'ĸ': {'k'},        // Latin Small letter Kra
		'Ĺ': {'L'},        // Latin Capital letter L with acute
		'ĺ': {'l'},        // Latin Small letter L with acute
		'Ļ': {'L'},        // Latin Capital letter L with cedilla
		'ļ': {'l'},        // Latin Small letter L with cedilla
		'Ľ': {'L'},        // Latin Capital letter L with caron
		'ľ': {'l'},        // Latin Small letter L with caron
		'Ŀ': {'L'},        // Latin Capital letter L with middle dot
		'ŀ': {'l'},        // Latin Small letter L with middle dot
		'Ł': {'L'},        // Latin Capital letter L with stroke
		'ł': {'L'},        // Latin Small letter L with stroke
		'Ń': {'N'},        // Latin Capital letter N with acute
		'ń': {'n'},        // Latin Small letter N with acute
		'Ņ': {'N'},        // Latin Capital letter N with cedilla
		'ņ': {'n'},        // Latin Small letter N with cedilla
		'Ň': {'N'},        // Latin Capital letter N with caron
		'ň': {'n'},        // Latin Small letter N with caron
		'ŉ': {'n'},        // Latin Small letter N preceded by apostrophe
		'Ŋ': []byte("NG"), // Latin Capital letter Eng
		'ŋ': []byte("ng"), // Latin Small letter Eng
		'Ō': {'O'},        // Latin Capital letter O with macron
		'ō': {'o'},        // Latin Small letter O with macron
		'Ŏ': {'O'},        // Latin Capital letter O with breve
		'ŏ': {'o'},        // Latin Small letter O with breve
		'Ő': {'O'},        // Latin Capital Letter O with double acute
		'ő': {'o'},        // Latin Small Letter O with double acute
		'Œ': []byte("OE"), // Latin Capital Ligature OE
		'œ': []byte("oe"), // Latin Small Ligature OE
		'Ŕ': {'R'},        // Latin Capital letter R with acute
		'ŕ': {'r'},        // Latin Small letter R with acute
		'Ŗ': {'R'},        // Latin Capital letter R with cedilla
		'ŗ': {'t'},        // Latin Small letter R with cedilla
		'Ř': {'R'},        // Latin Capital letter R with caron
		'ř': {'r'},        // Latin Small letter R with caron
		'Ś': {'S'},        // Latin Capital letter S with acute
		'ś': {'s'},        // Latin Small letter S with acute
		'Ŝ': {'S'},        // Latin Capital letter S with circumflex
		'ŝ': {'s'},        // Latin Small letter S with circumflex
		'Ş': {'S'},        // Latin Capital letter S with cedilla
		'ş': {'s'},        // Latin Small letter S with cedilla
		'Š': {'S'},        // Latin Capital letter S with caron
		'š': {'s'},        // Latin Small letter S with caron
		'Ţ': {'T'},        // Latin Capital letter T with cedilla
		'ţ': {'t'},        // Latin Small letter T with cedilla
		'Ť': {'T'},        // Latin Capital letter T with caron
		'ť': {'t'},        // Latin Small letter T with caron
		'Ŧ': {'T'},        // Latin Capital letter T with stroke
		'ŧ': {'t'},        // Latin Small letter T with stroke
		'Ũ': {'U'},        // Latin Capital letter U with tilde
		'ũ': {'u'},        // Latin Small letter U with tilde
		'Ū': {'U'},        // Latin Capital letter U with macron
		'ū': {'u'},        // Latin Small letter U with macron
		'Ŭ': {'U'},        // Latin Capital letter U with breve
		'ŭ': {'u'},        // Latin Small letter U with breve
		'Ů': {'U'},        // Latin Capital letter U with ring above
		'ů': {'u'},        // Latin Small letter U with ring above
		'Ű': {'U'},        // Latin Capital Letter U with double acute
		'ű': {'u'},        // Latin Small Letter U with double acute
		'Ų': {'U'},        // Latin Capital letter U with ogonek
		'ų': {'u'},        // Latin Small letter U with ogonek
		'Ŵ': {'W'},        // Latin Capital letter W with circumflex
		'ŵ': {'w'},        // Latin Small letter W with circumflex
		'Ŷ': {'Y'},        // Latin Capital letter Y with circumflex
		'ŷ': {'y'},        // Latin Small letter Y with circumflex
		'Ÿ': {'Y'},        // Latin Capital letter Y with diaeresis
		'Ź': {'Z'},        // Latin Capital letter Z with acute
		'ź': {'z'},        // Latin Small letter Z with acute
		'Ż': {'Z'},        // Latin Capital letter Z with dot above
		'ż': {'z'},        // Latin Small letter Z with dot above
		'Ž': {'Z'},        // Latin Capital letter Z with caron
		'ž': {'z'},        // Latin Small letter Z with caron
		'ſ': {'S'},        // Latin Small letter long S
	}
	// tagField is the initial field for ID3V1 metadata, and should contain 'TAG'
	tagField = initID3v1Field(tagOffset, tagLength)
	// these are the remaining fields making up ID3V1 metadata
	titleField    = initID3v1Field(titleOffset, titleLength)
	artistField   = initID3v1Field(artistOffset, artistLength)
	albumField    = initID3v1Field(albumOffset, albumLength)
	yearField     = initID3v1Field(yearOffset, yearLength)
	commentField  = initID3v1Field(commentOffset, commentLength)
	zeroByteField = initID3v1Field(zeroByteOffset, zeroByteLength)
	trackField    = initID3v1Field(trackOffset, trackLength)
	genreField    = initID3v1Field(genreOffset, genreLength)

	errNoID3V1MetadataFound = fmt.Errorf("no ID3V1 metadata found")
)

func initID3v1Field(offset, length int) id3v1Field {
	return id3v1Field{startOffset: offset, length: length, endOffset: offset + length}
}

type id3v1Metadata struct {
	data []byte
}

func newID3v1Metadata() *id3v1Metadata {
	return &id3v1Metadata{data: make([]byte, id3v1Length)}
}

func (im *id3v1Metadata) readString(f id3v1Field) string {
	return trim(string(im.data[f.startOffset:f.endOffset]))
}

func (im *id3v1Metadata) IsValid() bool {
	return im.readString(tagField) == "TAG"
}

func (im *id3v1Metadata) title() string {
	return im.readString(titleField)
}

func (im *id3v1Metadata) writeString(s string, f id3v1Field) {
	copy(im.data[f.startOffset:f.endOffset], make([]byte, f.length))
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

func (im *id3v1Metadata) setTitle(s string) {
	im.writeString(repairName(s), titleField)
}

func (im *id3v1Metadata) artist() string {
	return im.readString(artistField)
}

func (im *id3v1Metadata) setArtist(s string) {
	im.writeString(repairName(s), artistField)
}

func (im *id3v1Metadata) album() string {
	return im.readString(albumField)
}

func (im *id3v1Metadata) setAlbum(s string) {
	im.writeString(repairName(s), albumField)
}

func (im *id3v1Metadata) year() string {
	return im.readString(yearField)
}

func (im *id3v1Metadata) setYear(s string) {
	im.writeString(s, yearField)
}

func (im *id3v1Metadata) comment() string {
	return im.readString(commentField)
}

func (im *id3v1Metadata) readInt(f id3v1Field) int {
	return int(im.data[f.startOffset])
}

func (im *id3v1Metadata) track() (trackNumber int, trackNumberValid bool) {
	if im.readInt(zeroByteField) == 0 {
		trackNumber = im.readInt(trackField)
		trackNumberValid = true
	}
	return
}

func (im *id3v1Metadata) writeInt(v int, f id3v1Field) {
	im.data[f.startOffset] = byte(v)
}

func (im *id3v1Metadata) setTrack(t int) (b bool) {
	if t >= 1 && t <= 255 {
		im.writeInt(0, zeroByteField)
		im.writeInt(t, trackField)
		b = true
	}
	return
}

func (im *id3v1Metadata) genre() (string, bool) {
	genre, genreFound := genreName(im.readInt(genreField))
	return genre, genreFound
}

func (im *id3v1Metadata) setGenre(s string) {
	index, found := genreIndex(s)
	if !found {
		v, _ := genreIndex("other")
		im.writeInt(v, genreField)
		return
	}
	im.writeInt(index, genreField)
}

func trim(s string) string {
	for s != "" && (s[len(s)-1:] == " " || s[len(s)-1:] == "\u0000") {
		s = s[:len(s)-1]
	}
	return s
}

func readID3v1Metadata(path string) ([]string, error) {
	v1, readErr := internalReadID3V1Metadata(path, fileReader)
	if readErr != nil {
		return nil, readErr
	}
	output := make([]string, 0, 5)
	output = append(
		output,
		fmt.Sprintf("Artist: %s", v1.artist()),
		fmt.Sprintf("Album: %s", v1.album()),
		fmt.Sprintf("Title: %s", v1.title()),
	)
	if track, trackNumberValid := v1.track(); trackNumberValid {
		output = append(output, fmt.Sprintf("Track: %d", track))
	}
	output = append(output, fmt.Sprintf("Year: %s", v1.year()))
	if genre, genreFound := v1.genre(); genreFound {
		output = append(output, fmt.Sprintf("Genre: %s", genre))
	}
	if comment := v1.comment(); comment != "" {
		output = append(output, fmt.Sprintf("Comment: %s", comment))
	}
	return output, nil
}

func fileReader(f afero.File, b []byte) (int, error) {
	return f.Read(b)
}

func internalReadID3V1Metadata(
	path string,
	readFunc func(f afero.File, b []byte) (int, error),
) (*id3v1Metadata, error) {
	file, fileErr := cmdtoolkit.FileSystem().Open(path)
	if fileErr != nil {
		return nil, fileErr
	}
	defer func() {
		_ = file.Close()
	}()
	fileStat, _ := file.Stat()
	if fileStat != nil && fileStat.Size() < id3v1Length {
		return nil, errNoID3V1MetadataFound
	}
	_, _ = file.Seek(-id3v1Length, io.SeekEnd)
	v1 := newID3v1Metadata()
	r, readErr := readFunc(file, v1.data)
	if readErr != nil {
		return nil, readErr
	}
	if r < id3v1Length {
		return nil,
			fmt.Errorf("cannot read id3v1 metadata from file %q; only %d bytes read", path, r)
	}
	if !v1.IsValid() {
		return nil, errNoID3V1MetadataFound
	}
	return v1, nil
}

func (im *id3v1Metadata) write(path string) error {
	return im.internalWrite(path, writeToFile)
}

func writeToFile(f afero.File, b []byte) (int, error) {
	return f.Write(b)
}

func updateID3V1TrackMetadata(tm *TrackMetadata, path string) error {
	const src = ID3V1
	if !tm.editRequired(src) {
		return nil
	}
	var v1 *id3v1Metadata
	var fileErr error
	v1, fileErr = internalReadID3V1Metadata(path, fileReader)
	if fileErr != nil {
		return fileErr
	}
	if artistName := tm.artistName(src).correctedValue(); artistName != "" {
		v1.setArtist(artistName)
	}
	if albumName := tm.albumName(src).correctedValue(); albumName != "" {
		v1.setAlbum(albumName)
	}
	if albumGenre := tm.albumGenre(src).correctedValue(); albumGenre != "" {
		v1.setGenre(albumGenre)
	}
	if albumYear := tm.albumYear(src).correctedValue(); albumYear != "" {
		v1.setYear(albumYear)
	}
	if trackName := tm.trackName(src).correctedValue(); trackName != "" {
		v1.setTitle(trackName)
	}
	if trackNumber := tm.trackNumber(src).correctedValue(); trackNumber != 0 {
		_ = v1.setTrack(trackNumber)
	}
	return v1.write(path)
}

func (im *id3v1Metadata) internalWrite(path string,
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

func id3v1NameDiffers(cS *comparableStrings) bool {
	bs := make([]byte, 0, 2*len(cS.external))
	for _, r := range strings.ToLower(cS.external) {
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
	metadata := strings.ToLower(cS.metadata)
	metadataRunes := []rune(metadata)
	external := strings.ToLower(string(bs))
	externalRunes := []rune(external)
	if len(metadataRunes) != len(externalRunes) {
		return true
	}
	for i, c := range metadataRunes {
		r := externalRunes[i]
		if r == c {
			continue
		}
		// allow for the metadata rune to be one that is illegal for file names:
		// the external name is likely to be a file name
		if !isIllegalRuneForFileNames(c) {
			return true
		}
	}
	return false
}

func id3v1GenreDiffers(cS *comparableStrings) bool {
	if _, genreFound := genreIndex(cS.external); !genreFound {
		// the external genre does not map to a known id3v1 genre but "other"
		// always matches the external name
		if cS.metadata == "other" {
			return false
		}
	}
	// external name is a known id3v1 genre, or metadata name is not "other"
	return !strings.EqualFold(cS.external, cS.metadata)
}
