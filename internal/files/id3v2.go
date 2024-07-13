package files

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bogem/id3v2/v2"
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
)

type id3v2Metadata struct {
	albumTitle        string
	artistName        string
	err               error
	genre             string
	musicCDIdentifier id3v2.UnknownFrame
	trackName         string
	trackNumber       int
	year              string
}

func (im *id3v2Metadata) hasError() bool {
	return im.err != nil
}

func readID3V2Tag(path string) (*id3v2.Tag, error) {
	file, readError := cmdtoolkit.FileSystem().Open(path)
	if readError != nil {
		return nil, readError
	}
	tag, parseError := id3v2.ParseReader(file, id3v2.Options{Parse: true, ParseFrames: nil})
	if isTagAbsent(tag) {
		return tag, errNoID3V2MetadataFound
	}
	return tag, parseError
}

func isTagAbsent(tag *id3v2.Tag) bool {
	if tag == nil {
		return true
	}
	return tag.Count() == 0
}

func rawReadID3V2Metadata(path string) (d *id3v2Metadata) {
	d = &id3v2Metadata{}
	tag, readErr := readID3V2Tag(path)
	if readErr != nil {
		d.err = readErr
		return
	}
	defer func() {
		_ = tag.Close()
	}()
	trackNumber, trackErr := toTrackNumber(tag.GetTextFrame(trackFrame).Text)
	if trackErr != nil {
		d.err = trackErr
		return
	}
	d.albumTitle = removeLeadingBOMs(tag.Album())
	d.artistName = removeLeadingBOMs(tag.Artist())
	d.genre = normalizeGenre(removeLeadingBOMs(tag.Genre()))
	d.trackName = removeLeadingBOMs(tag.Title())
	d.trackNumber = trackNumber
	d.year = removeLeadingBOMs(tag.Year())
	mcdiFramers := tag.AllFrames()[mcdiFrame]
	d.musicCDIdentifier = selectUnknownFrame(mcdiFramers)
	return
}

// normalizeGenre handles issues relating to a common practice in mp3 files, where the ID3V2
// genre field 'recognizes' the older ID3V1 genre field, by its value being written as
// "(key)value", where 'key' is the integer index (as used by ID3V1) and 'value' is the
// canonical ID3V1 string for that key. This function detects these "(key)value" strings,
// verifies that 'value' is correct for the specified key, and, if so, returns the 'value'
// piece without the parenthetical key. Everything else passes through 'as is'.
func normalizeGenre(s string) string {
	var i int
	var value string
	if n, scanErr := fmt.Sscanf(s, "(%d)%s", &i, &value); n == 2 && scanErr == nil {
		// discard value
		if splits := strings.SplitAfter(s, ")"); len(splits) >= 2 {
			value = strings.ToLower(splits[1])
			mappedValue, _ := genreName(i)
			if value == mappedValue {
				return mappedValue
			}
		}
	}
	return s
}

var (
	errMalformedTrackNumber = fmt.Errorf("track number first character is not a digit")
	errMissingTrackNumber   = fmt.Errorf("track number is zero length")
	errNoID3V2MetadataFound = fmt.Errorf("no ID3V2 metadata found")
)

func toTrackNumber(s string) (int, error) {
	s = removeLeadingBOMs(s)
	if s == "" {
		return 0, errMissingTrackNumber
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
				return 0, errMalformedTrackNumber
			default: // did read at least one digit
				return n, nil
			}
		}
	}
	// normal path, whole string was digits
	return n, nil
}

// removeLeadingBOMs removes leading byte order marks (BOMs); frame values may begin with BOMs,
// depending on encoding
func removeLeadingBOMs(s string) string {
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

func selectUnknownFrame(mcdiFramers []id3v2.Framer) id3v2.UnknownFrame {
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
	if !tm.editRequired(src) {
		return nil
	}
	tag, readErr := readID3V2Tag(path)
	if readErr != nil {
		return readErr
	}
	defer func() {
		_ = tag.Close()
	}()
	tag.SetDefaultEncoding(id3v2.EncodingUTF8)
	if artistName := tm.artistName(src).correctedValue(); artistName != "" {
		tag.SetArtist(artistName)
	}
	if albumName := tm.albumName(src).correctedValue(); albumName != "" {
		tag.SetAlbum(albumName)
	}
	if albumGenre := tm.albumGenre(src).correctedValue(); albumGenre != "" {
		tag.SetGenre(albumGenre)
	}
	if albumYear := tm.albumYear(src).correctedValue(); albumYear != "" {
		tag.SetYear(albumYear)
	}
	if trackName := tm.trackName(src).correctedValue(); trackName != "" {
		tag.SetTitle(trackName)
	}
	if trackNumber := tm.trackNumber(src).correctedValue(); trackNumber != 0 {
		tag.AddTextFrame("TRCK", tag.DefaultEncoding(), fmt.Sprintf("%d", trackNumber))
	}
	cdIdentifier := tm.cdIdentifier().correctedValue()
	if len(cdIdentifier.Body) != 0 {
		tag.DeleteFrames(mcdiFrame)
		tag.AddFrame(mcdiFrame, cdIdentifier)
	}
	return tag.Save()
}

type id3v2TrackFrame struct {
	name  string
	value string
}

// String returns the contents of an ID3V2TrackFrame formatted in the form
// "name = \"value\"".
func (itf *id3v2TrackFrame) String() string {
	return fmt.Sprintf("%s = %q", itf.name, itf.value)
}

type ID3V2Info struct {
	Version      byte
	Encoding     string
	FrameStrings []string
	RawFrames    []*id3v2TrackFrame
}

func readID3V2Metadata(path string) (info *ID3V2Info, e error) {
	info = &ID3V2Info{
		FrameStrings: []string{},
		RawFrames:    []*id3v2TrackFrame{},
	}
	tag, readErr := readID3V2Tag(path)
	if readErr != nil {
		e = readErr
		return
	}
	defer func() {
		_ = tag.Close()
	}()
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
			value = removeLeadingBOMs(tag.GetTextFrame(n).Text)
		default:
			value = framerSliceAsString(frameMap[n])
		}
		frame := &id3v2TrackFrame{name: n, value: value}
		info.FrameStrings = append(info.FrameStrings, frame.String())
		info.RawFrames = append(info.RawFrames, frame)
	}
	return
}

func framerSliceAsString(f []id3v2.Framer) string {
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

func id3v2NameDiffers(cS *comparableStrings) bool {
	externalName := strings.ToLower(cS.external)
	metadataName := strings.ToLower(cS.metadata)
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
		if !isIllegalRuneForFileNames(c) {
			return true
		}
	}
	return false // rune by rune comparison was successful
}

func id3v2GenreDiffers(cS *comparableStrings) bool {
	// differs unless exact match. Period.
	return cS.external != cS.metadata
}
