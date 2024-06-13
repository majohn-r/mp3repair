package files

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bogem/id3v2/v2"
	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
)

type Id3v2Metadata struct {
	AlbumTitle        string
	ArtistName        string
	Err               error
	Genre             string
	MusicCDIdentifier id3v2.UnknownFrame
	TrackName         string
	TrackNumber       int
	Year              string
}

func (im *Id3v2Metadata) HasError() bool {
	return im.Err != nil
}

func readID3V2Tag(path string) (*id3v2.Tag, error) {
	file, readError := cmd_toolkit.FileSystem().Open(path)
	if readError != nil {
		return nil, readError
	}
	tag, parseError := id3v2.ParseReader(file, id3v2.Options{Parse: true, ParseFrames: nil})
	if IsTagAbsent(tag) {
		return tag, ErrNoID3V2MetadataFound
	}
	return tag, parseError
}

func IsTagAbsent(tag *id3v2.Tag) bool {
	if tag == nil {
		return true
	}
	return tag.Count() == 0
}

func RawReadID3V2Metadata(path string) (d *Id3v2Metadata) {
	d = &Id3v2Metadata{}
	tag, readErr := readID3V2Tag(path)
	if readErr != nil {
		d.Err = readErr
		return
	}
	defer tag.Close()
	trackNumber, trackErr := ToTrackNumber(tag.GetTextFrame(trackFrame).Text)
	if trackErr != nil {
		d.Err = trackErr
		return
	}
	d.AlbumTitle = RemoveLeadingBOMs(tag.Album())
	d.ArtistName = RemoveLeadingBOMs(tag.Artist())
	d.Genre = NormalizeGenre(RemoveLeadingBOMs(tag.Genre()))
	d.TrackName = RemoveLeadingBOMs(tag.Title())
	d.TrackNumber = trackNumber
	d.Year = RemoveLeadingBOMs(tag.Year())
	mcdiFramers := tag.AllFrames()[mcdiFrame]
	d.MusicCDIdentifier = SelectUnknownFrame(mcdiFramers)
	return
}

// sometimes an id3v2 genre tries to show some solidarity with the old id3v1
// genre, by making the genre string "(key)value", where key is the integer
// index and value is the canonical string for that key, as defined for id3v1.
// This function also has to take into account that the mapping is imperfect in
// the case of "Rhythm and Blues", which is abbreviated to "R&B". This function
// detects these "(key)value" strings, verifies that the value is correct for
// the specified key, and, if so, returns the plain value without the
// parenthetical key. Everything else passes through 'as is'.
func NormalizeGenre(s string) string {
	var i int
	var value string
	if n, scanErr := fmt.Sscanf(s, "(%d)%s", &i, &value); n == 2 && scanErr == nil {
		// discard value
		if splits := strings.SplitAfter(s, ")"); len(splits) >= 2 {
			value = splits[1]
			mappedValue := GenreMap[i]
			if value == mappedValue || (value == "R&B" && mappedValue == "Rhythm and Blues") {
				return mappedValue
			}
		}
	}
	return s
}

var (
	ErrMalformedTrackNumber = fmt.Errorf("track number first character is not a digit")
	ErrMissingTrackNumber   = fmt.Errorf("track number is zero length")
	ErrNoID3V2MetadataFound = fmt.Errorf("no ID3V2 metadata found")
)

func ToTrackNumber(s string) (int, error) {
	s = RemoveLeadingBOMs(s)
	if s == "" {
		return 0, ErrMissingTrackNumber
	}
	// this is more complicated than I wanted, because some mp3 rippers produce
	// track numbers like "12/14", meaning 12th track of 14
	n := 0
	bs := []byte(s)
	for j, b := range bs {
		c := int(b)
		switch {
		case c >= '0' && c <= '9':
			n *= 10
			n += c - '0'
		default:
			// found something other than a digit
			switch j {
			case 0: // never saw a digit
				return 0, ErrMalformedTrackNumber
			default: // did read at least one digit
				return n, nil
			}
		}
	}
	// normal path, whole string was digits
	return n, nil
}

// depending on encoding, frame values may begin with a BOM (byte order mark)
func RemoveLeadingBOMs(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	if r[0] != '\ufeff' {
		return s
	}
	for r[0] == '\ufeff' {
		r = r[1:]
		if len(r) == 0 {
			break
		}
	}
	return string(r)
}

func SelectUnknownFrame(mcdiFramers []id3v2.Framer) id3v2.UnknownFrame {
	if len(mcdiFramers) == 1 {
		frame := mcdiFramers[0]
		if f, ok := frame.(id3v2.UnknownFrame); ok {
			return f
		}
	}
	return id3v2.UnknownFrame{Body: []byte{0}}
}

func updateID3V2TrackMetadata(tm *TrackMetadata, path string) error {
	const src = ID3V2
	if !tm.EditRequired(src) {
		return nil
	}
	tag, readErr := readID3V2Tag(path)
	if readErr != nil {
		return readErr
	}
	defer tag.Close()
	tag.SetDefaultEncoding(id3v2.EncodingUTF8)
	if artistName := tm.ArtistName(src).Correction(); artistName != "" {
		tag.SetArtist(artistName)
	}
	if albumName := tm.AlbumName(src).Correction(); albumName != "" {
		tag.SetAlbum(albumName)
	}
	if albumGenre := tm.AlbumGenre(src).Correction(); albumGenre != "" {
		tag.SetGenre(albumGenre)
	}
	if albumYear := tm.AlbumYear(src).Correction(); albumYear != "" {
		tag.SetYear(albumYear)
	}
	if trackName := tm.TrackName(src).Correction(); trackName != "" {
		tag.SetTitle(trackName)
	}
	if trackNumber := tm.TrackNumber(src).Correction(); trackNumber != 0 {
		tag.AddTextFrame("TRCK", tag.DefaultEncoding(), fmt.Sprintf("%d", trackNumber))
	}
	cdIdentifier := tm.CDIdentifier().Correction()
	if len(cdIdentifier.Body) != 0 {
		tag.DeleteFrames(mcdiFrame)
		tag.AddFrame(mcdiFrame, cdIdentifier)
	}
	return tag.Save()
}

type Id3v2TrackFrame struct {
	Name  string
	Value string
}

// String returns the contents of an ID3V2TrackFrame formatted in the form
// "name = \"value\"".
func (itf *Id3v2TrackFrame) String() string {
	return fmt.Sprintf("%s = %q", itf.Name, itf.Value)
}

type ID3V2Info struct {
	Version      byte
	Encoding     string
	FrameStrings []string
	RawFrames    []*Id3v2TrackFrame
}

func ReadID3V2Metadata(path string) (info *ID3V2Info, e error) {
	info = &ID3V2Info{
		FrameStrings: []string{},
		RawFrames:    []*Id3v2TrackFrame{},
	}
	tag, readErr := readID3V2Tag(path)
	if readErr != nil {
		e = readErr
		return
	}
	defer tag.Close()
	info.Version = tag.Version()
	info.Encoding = tag.DefaultEncoding().Name
	frameMap := tag.AllFrames()
	frameNames := make([]string, 0, len(frameMap))
	for k := range frameMap {
		frameNames = append(frameNames, k)
	}
	sort.Strings(frameNames)
	for _, n := range frameNames {
		var value string
		switch {
		case strings.HasPrefix(n, "T"): // tag
			value = RemoveLeadingBOMs(tag.GetTextFrame(n).Text)
		default:
			value = FramerSliceAsString(frameMap[n])
		}
		frame := &Id3v2TrackFrame{Name: n, Value: value}
		info.FrameStrings = append(info.FrameStrings, frame.String())
		info.RawFrames = append(info.RawFrames, frame)
	}
	return
}

func FramerSliceAsString(f []id3v2.Framer) string {
	substrings := make([]string, 0, len(f))
	switch {
	case len(f) == 1:
		data, ok := f[0].(id3v2.UnknownFrame)
		switch {
		case ok:
			substrings = append(substrings, fmt.Sprintf("%#v", data.Body))
		default:
			substrings = append(substrings, fmt.Sprintf("%#v", f[0]))
		}
	default:
		for k, framer := range f {
			data, ok := framer.(id3v2.UnknownFrame)
			switch {
			case ok:
				substrings = append(substrings, fmt.Sprintf("[%d %#v]", k, data.Body))
			default:
				substrings = append(substrings, fmt.Sprintf("[%d %#v]", k, framer))
			}
		}
	}
	return fmt.Sprintf("<<%s>>", strings.Join(substrings, ", "))
}

func Id3v2NameDiffers(cS *ComparableStrings) bool {
	externalName := strings.ToLower(cS.External)
	metadataName := strings.ToLower(cS.Metadata)
	// strip off trailing space from the metadata value
	for strings.HasSuffix(metadataName, " ") {
		metadataName = metadataName[:len(metadataName)-1]
	}
	if externalName == metadataName {
		return false
	}
	metadataRunes := []rune(metadataName)
	externalRunes := []rune(externalName)
	if len(metadataRunes) != len(externalRunes) {
		return true
	}
	for index, c := range metadataRunes {
		if externalRunes[index] == c {
			continue
		}
		// allow for the metadata rune to be one that is illegal for file names:
		// the external name is likely to be a file name
		if !IsIllegalRuneForFileNames(c) {
			return true
		}
	}
	return false // rune by rune comparison was successful
}

func Id3v2GenreDiffers(cS *ComparableStrings) bool {
	// differs unless exact match. Period.
	return cS.External != cS.Metadata
}
