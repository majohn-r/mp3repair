package files

import (
	"fmt"
	"mp3/internal"
	"sort"
	"strings"

	"github.com/bogem/id3v2/v2"
)

// ID3V2TaggedTrackData contains raw ID3V2 tag frame data and is public so that
// tests can populate it.
type ID3V2TaggedTrackData struct {
	album             string
	artist            string
	title             string
	genre             string
	year              string
	track             int
	musicCDIdentifier id3v2.UnknownFrame
	err               string
}

// NewID3V2TaggedTrackDataForTesting creates a new instance of
// ID3V2TaggedTrackData. The method is public so it can be called from unit
// tests.
func NewID3V2TaggedTrackDataForTesting(albumFrame string, artistFrame string, titleFrame string, evaluatedNumberFrame int, mcdi []byte) *ID3V2TaggedTrackData {
	return &ID3V2TaggedTrackData{
		album:             albumFrame,
		artist:            artistFrame,
		title:             titleFrame,
		track:             evaluatedNumberFrame,
		musicCDIdentifier: id3v2.UnknownFrame{Body: mcdi},
		err:               "",
	}
}

func readID3V2Tag(path string) (*id3v2.Tag, error) {
	return id3v2.Open(path, id3v2.Options{Parse: true, ParseFrames: nil})
}

// RawReadID3V2Tag reads the ID3V2 tag from an MP3 file and collects interesting
// frame values.
func RawReadID3V2Tag(path string) (d *ID3V2TaggedTrackData) {
	d = &ID3V2TaggedTrackData{}
	var tag *id3v2.Tag
	var err error
	if tag, err = readID3V2Tag(path); err != nil {
		d.err = fmt.Sprintf("%v", err)
		return
	}
	defer tag.Close()
	if trackNumber, err := toTrackNumber(tag.GetTextFrame(trackFrame).Text); err != nil {
		d.err = fmt.Sprintf("%v", err)
	} else {
		d.album = removeLeadingBOMs(tag.Album())
		d.artist = removeLeadingBOMs(tag.Artist())
		d.genre = removeLeadingBOMs(tag.Genre())
		d.title = removeLeadingBOMs(tag.Title())
		d.track = trackNumber
		d.year = removeLeadingBOMs(tag.Year())
		mcdiFramers := tag.AllFrames()[mcdiFrame]
		d.musicCDIdentifier = selectUnknownFrame(mcdiFramers)
	}
	return
}

func toTrackNumber(s string) (i int, err error) {
	// this is more complicated than I wanted, because some mp3 rippers produce
	// track numbers like "12/14", meaning 12th track of 14
	if len(s) == 0 {
		err = fmt.Errorf(internal.ERROR_ZERO_LENGTH)
		return
	}
	s = removeLeadingBOMs(s)
	n := 0
	bs := []byte(s)
	for j, b := range bs {
		c := int(b)
		if c >= '0' && c <= '9' {
			n *= 10
			n += c - '0'
		} else {
			switch j {
			case 0: // never saw a digit
				err = fmt.Errorf(internal.ERROR_DOES_NOT_BEGIN_WITH_DIGIT)
				return
			default: // found something other than a digit, but read at least one
				i = n
				return
			}
		}
	}
	// normal path, whole string was digits
	i = n
	return
}

// depending on encoding, frame values may begin with a BOM (byte order mark)
func removeLeadingBOMs(s string) string {
	if len(s) == 0 {
		return s
	}
	r := []rune(s)
	if r[0] == '\ufeff' {
		return removeLeadingBOMs(string(r[1:]))
	}
	return s
}

func selectUnknownFrame(mcdiFramers []id3v2.Framer) id3v2.UnknownFrame {
	uf := id3v2.UnknownFrame{Body: []byte{0}}
	if len(mcdiFramers) == 1 {
		frame := mcdiFramers[0]
		if f, ok := frame.(id3v2.UnknownFrame); ok {
			uf = f
		}
	}
	return uf
}

func updateID3V2Tag(t *Track, a taggedTrackState) error {
	tag, err := readID3V2Tag(t.path)
	if err != nil {
		return err
	}
	defer tag.Close()
	tag.SetDefaultEncoding(id3v2.EncodingUTF8)
	if a.HasAlbumNameConflict() {
		tag.SetAlbum(t.containingAlbum.canonicalTitle)
	}
	if a.HasArtistNameConflict() {
		tag.SetArtist(t.containingAlbum.recordingArtist.canonicalName)
	}
	if a.HasTrackNameConflict() {
		tag.SetTitle(t.name)
	}
	if a.HasNumberingConflict() {
		tag.AddTextFrame("TRCK", tag.DefaultEncoding(), fmt.Sprintf("%d", t.number))
	}
	if a.HasGenreConflict() {
		tag.SetGenre(t.containingAlbum.genre)
	}
	if a.HasYearConflict() {
		tag.SetYear(t.containingAlbum.year)
	}
	if a.HasMCDIConflict() {
		tag.DeleteFrames(mcdiFrame)
		tag.AddFrame(mcdiFrame, t.containingAlbum.musicCDIdentifier)
	}
	return tag.Save()
}

type id3v2TrackFrame struct {
	name  string
	value string
}

// String returns the contents of an ID3V2TrackFrame formatted in the form
// "name = \"value\"".
func (f *id3v2TrackFrame) String() string {
	return fmt.Sprintf("%s = %q", f.name, f.value)
}

func readID3V3Metadata(path string) (version byte, enc string, f []string, e error) {
	var tag *id3v2.Tag
	var err error
	if tag, err = readID3V2Tag(path); err != nil {
		e = err
		return
	}
	defer tag.Close()
	frames := tag.AllFrames()
	var frameNames []string
	for k := range frames {
		frameNames = append(frameNames, k)
	}
	sort.Strings(frameNames)
	for _, n := range frameNames {
		var frame *id3v2TrackFrame
		if strings.HasPrefix(n, "T") {
			frame = &id3v2TrackFrame{name: n, value: removeLeadingBOMs(tag.GetTextFrame(n).Text)}
		} else {
			frame = &id3v2TrackFrame{name: n, value: stringifyFramerArray(frames[n])}
		}
		f = append(f, frame.String())
	}
	enc = tag.DefaultEncoding().Name
	version = tag.Version()
	return
}

func stringifyFramerArray(f []id3v2.Framer) string {
	var substrings []string
	if len(f) == 1 {
		if data, ok := f[0].(id3v2.UnknownFrame); ok {
			substrings = append(substrings, fmt.Sprintf("%#v", data.Body))
		} else {
			substrings = append(substrings, fmt.Sprintf("%#v", f[0]))
		}
	} else {
		for k, framer := range f {
			if data, ok := framer.(id3v2.UnknownFrame); ok {
				substrings = append(substrings, fmt.Sprintf("[%d %#v]", k, data.Body))
			} else {
				substrings = append(substrings, fmt.Sprintf("[%d %#v]", k, framer))
			}
		}
	}
	return fmt.Sprintf("<<%s>>", strings.Join(substrings, ", "))
}
