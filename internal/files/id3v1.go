package files

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"
	"strings"
)

// values per https://id3.org/ID3v1 as of August 16 2022
const (
	tagOffset      = 0 // always 'TAG' if present
	tagLength      = 3
	titleOffset    = tagOffset + tagLength // first 30 characters of the track title
	titleLength    = 30
	artistOffset   = titleOffset + titleLength // first 30 characters of the artist name
	artistLength   = 30
	albumOffset    = artistOffset + artistLength // first 30 characters of the album name
	albumLength    = 30
	yearOffset     = albumOffset + albumLength // four digit year, e.g., '2', '0', '2', '2'
	yearLength     = 4
	commentOffset  = yearOffset + yearLength // comment, rarely used, and not interesting
	commentLength  = 28
	zeroByteOffset = commentOffset + commentLength // should always be zero; if not, then the track is not valid
	zeroByteLength = 1
	trackOffset    = zeroByteOffset + zeroByteLength // track number; if zeroByte is not zero, not valid
	trackLength    = 1
	genreOffset    = trackOffset + trackLength // genre list index
	genreLength    = 1
	id3v1Length    = genreOffset + genreLength // total length of the ID3V1 block
)

// per https://en.wikipedia.org/wiki/List_of_ID3v1_Genres as of August 16 2022
var genreMap = map[int]string{
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

var genreIndicesMap = map[string]int{} // lazily initialized when needed; keys are all lowercase

type id3v1Field struct {
	startOffset int
	length      int
	endOffset   int
}

var (
	id3v1Tag      = initId3v1Field(tagOffset, tagLength)
	id3v1Title    = initId3v1Field(titleOffset, titleLength)
	id3v1Artist   = initId3v1Field(artistOffset, artistLength)
	id3v1Album    = initId3v1Field(albumOffset, albumLength)
	id3v1Year     = initId3v1Field(yearOffset, yearLength)
	id3v1Comment  = initId3v1Field(commentOffset, commentLength)
	id3v1ZeroByte = initId3v1Field(zeroByteOffset, zeroByteLength)
	id3v1Track    = initId3v1Field(trackOffset, trackLength)
	id3v1Genre    = initId3v1Field(genreOffset, genreLength)
)

func initId3v1Field(offset int, length int) id3v1Field {
	return id3v1Field{
		startOffset: offset,
		length:      length,
		endOffset:   offset + length,
	}
}

type id3v1Metadata struct {
	data []byte
}

func newId3v1MetadataWithData(b []byte) *id3v1Metadata {
	v1 := newId3v1Metadata()
	if len(b) >= id3v1Length {
		for k := 0; k < id3v1Length; k++ {
			v1.data[k] = b[k]
		}
	} else {
		copy(v1.data, b)
		for k := len(b); k < id3v1Length; k++ {
			v1.data[k] = 0
		}
	}
	return v1
}

func newId3v1Metadata() *id3v1Metadata {
	return &id3v1Metadata{data: make([]byte, id3v1Length)}
}

func (v1 *id3v1Metadata) readStringField(f id3v1Field) string {
	s := string(v1.data[f.startOffset:f.endOffset])
	return trim(s)
}

func (v1 *id3v1Metadata) isValid() bool {
	tag := v1.readStringField(id3v1Tag)
	return tag == "TAG"
}

func (v1 *id3v1Metadata) getTitle() string {
	return v1.readStringField(id3v1Title)
}

func createField(i int) []byte {
	b := make([]byte, i)
	for k := 0; k < i; k++ {
		b[k] = 0
	}
	return b
}

func (v1 *id3v1Metadata) writeStringField(s string, f id3v1Field) {
	b := createField(f.length)
	if len(s) > f.length {
		s = s[0:f.length]
	}
	copy(b, s)
	copy(v1.data[f.startOffset:f.endOffset], b)
}

func (v1 *id3v1Metadata) setTitle(s string) {
	v1.writeStringField(s, id3v1Title)
}

func (v1 *id3v1Metadata) getArtist() string {
	return v1.readStringField(id3v1Artist)
}

func (v1 *id3v1Metadata) setArtist(s string) {
	v1.writeStringField(s, id3v1Artist)
}

func (v1 *id3v1Metadata) getAlbum() string {
	return v1.readStringField(id3v1Album)
}

func (v1 *id3v1Metadata) setAlbum(s string) {
	v1.writeStringField(s, id3v1Album)
}

func (v1 *id3v1Metadata) getYear() (y int, ok bool) {
	s := v1.readStringField(id3v1Year)
	if year, err := strconv.Atoi(s); err == nil {
		y = year
		ok = true
	}
	return
}

func (v1 *id3v1Metadata) setYear(y int) bool {
	if y < 1000 || y > 9999 {
		return false
	}
	v1.writeStringField(fmt.Sprintf("%d", y), id3v1Year)
	return true
}

func (v1 *id3v1Metadata) getComment() string {
	return v1.readStringField(id3v1Comment)
}

func (v1 *id3v1Metadata) setComment(s string) {
	v1.writeStringField(s, id3v1Comment)
}

func (v1 *id3v1Metadata) readByteField(f id3v1Field) int {
	return int(v1.data[f.startOffset])
}

func (v1 *id3v1Metadata) getTrack() (i int, ok bool) {
	if v1.readByteField(id3v1ZeroByte) == 0 {
		i = v1.readByteField(id3v1Track)
		ok = true
	}
	return
}

func (v1 *id3v1Metadata) setByteField(v int, f id3v1Field) {
	v1.data[f.startOffset] = byte(v)
}

func (v1 *id3v1Metadata) setTrack(t int) bool {
	if t < 1 || t > 255 {
		return false
	}
	v1.setByteField(0, id3v1ZeroByte)
	v1.setByteField(t, id3v1Track)
	return true
}

func (v1 *id3v1Metadata) getGenre() (string, bool) {
	s, ok := genreMap[v1.readByteField(id3v1Genre)]
	return s, ok
}

func initGenreIndices() {
	if len(genreIndicesMap) == 0 {
		for k, v := range genreMap {
			genreIndicesMap[strings.ToLower(v)] = k
		}
	}
}

func (v1 *id3v1Metadata) setGenre(s string) bool {
	initGenreIndices()
	if index, ok := genreIndicesMap[strings.ToLower(s)]; !ok {
		return false
	} else {
		v1.setByteField(index, id3v1Genre)
		return true
	}
}

func trim(s string) string {
	if strings.HasSuffix(s, " ") {
		return stripTrailing(s, " ")
	}
	if strings.HasSuffix(s, "\u0000") {
		return stripTrailing(s, "\u0000")
	}
	return s
}

func stripTrailing(s string, suffix string) string {
	for strings.HasSuffix(s, suffix) {
		s = strings.TrimSuffix(s, suffix)
	}
	return s
}

func readId3v1Metadata(path string) ([]string, error) {
	if v1, err := internalReadId3V1Metadata(path, readFromFile); err != nil {
		return nil, err
	} else {
		var output []string
		output = append(output, fmt.Sprintf("Artist: %q", v1.getArtist()))
		output = append(output, fmt.Sprintf("Album: %q", v1.getAlbum()))
		output = append(output, fmt.Sprintf("Title: %q", v1.getTitle()))
		if track, ok := v1.getTrack(); ok {
			output = append(output, fmt.Sprintf("Track: %d", track))
		}
		if year, ok := v1.getYear(); ok {
			output = append(output, fmt.Sprintf("Year: %d", year))
		}
		if genre, ok := v1.getGenre(); ok {
			output = append(output, fmt.Sprintf("Genre: %q", genre))
		}
		if comment := v1.getComment(); len(comment) > 0 {
			output = append(output, fmt.Sprintf("Comment: %q", comment))
		}
		return output, nil

	}
}

func readFromFile(f *os.File, b []byte) (int, error) {
	return f.Read(b)
}

func internalReadId3V1Metadata(path string, readFunc func(f *os.File, b []byte) (int, error)) (*id3v1Metadata, error) {
	if file, err := os.Open(path); err != nil {
		return nil, err
	} else {
		defer file.Close()
		if _, err = file.Seek(-id3v1Length, io.SeekEnd); err != nil {
			return nil, err
		}
		v1 := newId3v1Metadata()
		if r, err := readFunc(file, v1.data); err != nil {
			return nil, err
		} else if r < id3v1Length {
			return nil, fmt.Errorf("cannot read id3v1 tag from file %q; only %d bytes read", path, r)
		}
		if v1.isValid() {
			return v1, nil
		}
		return nil, fmt.Errorf("no id3v1 tag found in file %q", path)
	}
}

func (v1 *id3v1Metadata) write(path string) error {
	return v1.internalWrite(path, writeToFile)
}

func writeToFile(f *os.File, b []byte) (int, error) {
	return f.Write(b)
}

func (v1 *id3v1Metadata) internalWrite(oldPath string, writeFunc func(f *os.File, b []byte) (int, error) ) (err error) {
	var oldFile *os.File
	if oldFile, err = os.Open(oldPath); err == nil {
		defer oldFile.Close()
		var stat fs.FileInfo
		if stat, err = oldFile.Stat(); err == nil {
			newPath := oldPath + "-id3v1"
			var newFile *os.File
			if newFile, err = os.OpenFile(newPath, os.O_RDWR|os.O_CREATE, stat.Mode()); err == nil {
				defer newFile.Close()
				// borrowed this piece of logic from id3v2 tag.Save() method
				tempfileShouldBeRemoved := true
				defer func() {
					if tempfileShouldBeRemoved {
						os.Remove(newPath)
					}
				}()
				if _, err = io.Copy(newFile, oldFile); err == nil {
					oldFile.Close()
					if _, err = newFile.Seek(-id3v1Length, io.SeekEnd); err == nil {
						var n int
						if n, err = writeFunc(newFile, v1.data); err == nil {
							newFile.Close()
							if n != id3v1Length {
								err = fmt.Errorf("wrote %d bytes to %q, expected to write %d bytes", n, newPath, id3v1Length)
								return
							}
							if err = os.Rename(newPath, oldPath); err == nil {
								tempfileShouldBeRemoved = false
							}
						}
					}
				}
			}
		}
	}
	return
}
