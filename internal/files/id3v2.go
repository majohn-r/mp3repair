/*
Copyright © 2026 Marc Johnson (marc.johnson27591@gmail.com)
*/
package files

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
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
	errMalformedTrackNumber  = fmt.Errorf("track number first character is not a digit")
	errMissingTrackNumber    = fmt.Errorf("track number is zero length")
	errNoID3V2MetadataFound  = fmt.Errorf("no ID3V2 metadata found")
	windowsLegacyMCDIPattern = regexp.MustCompile(`([0-9A-F]+\+)+([0-9A-F]+)`)
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
	value []string
}

// String returns the contents of an ID3V2TrackFrame formatted in the form
// name = \"value\".
func (itf *id3v2TrackFrame) String() string {
	return fmt.Sprintf("%s = %q", itf.name, itf.value)
}

type ID3V2Info struct {
	version   byte
	encoding  string
	frames    map[string][]string
	rawFrames []*id3v2TrackFrame
}

func (info *ID3V2Info) Frames() map[string][]string { return info.frames }

func (info *ID3V2Info) Version() byte {
	return info.version
}

func (info *ID3V2Info) Encoding() string {
	return info.encoding
}

func NewID3V2Info(version byte, encoding string, frames map[string][]string, rawFrames []*id3v2TrackFrame) *ID3V2Info {
	return &ID3V2Info{
		version:   version,
		encoding:  encoding,
		frames:    frames,
		rawFrames: rawFrames,
	}
}

func readID3V2Metadata(path string) (*ID3V2Info, error) {
	tag, readErr := readID3V2Tag(path)
	if readErr != nil {
		return &ID3V2Info{
			frames:    map[string][]string{},
			rawFrames: []*id3v2TrackFrame{},
		}, readErr
	}
	defer func() {
		_ = tag.Close()
	}()
	info := NewID3V2Info(tag.Version(), tag.DefaultEncoding().String(), map[string][]string{}, []*id3v2TrackFrame{})
	frameMap := tag.AllFrames()
	frameNames := slices.Sorted(maps.Keys(frameMap))
	for _, n := range frameNames {
		var value []string
		switch {
		case n == "TLEN":
			rawValue := removeLeadingBOMs(tag.GetTextFrame(n).Text)
			length := 0
			for _, c := range rawValue {
				switch c {
				case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
					length *= 10
					length += int(c - '0')
				}
			}
			rawSeconds := length / 1000
			minutes := rawSeconds / 60
			value = []string{fmt.Sprintf("%d:%02d.%03d", minutes, rawSeconds%60, length%1000)}
		case strings.HasPrefix(n, "T"): // tag
			value = []string{removeLeadingBOMs(tag.GetTextFrame(n).Text)}
		default:
			value = framerSliceAsString(frameMap[n])
		}
		frame := &id3v2TrackFrame{name: n, value: value}
		info.frames[n] = append(info.frames[n], value...)
		info.rawFrames = append(info.rawFrames, frame)
	}
	return info, nil
}

func framerSliceAsString(f []id3v2.Framer) []string {
	substrings := make([]string, 0, 100) // be generous!
	for _, framer := range f {
		if unknownFrame, isUnknownFrame := framer.(id3v2.UnknownFrame); isUnknownFrame {
			substrings = append(substrings, interpretUnknownFrame(unknownFrame)...)
			continue
		}
		if pictureFrame, isPictureFrame := framer.(id3v2.PictureFrame); isPictureFrame {
			substrings = append(substrings, interpretPictureFrame(&pictureFrame)...)
			continue
		}
		substrings = append(substrings, fmt.Sprintf("%#v", framer))
	}
	return substrings
}

var pictureTypes = map[byte]string{
	0x00: "Other",
	0x01: "32x32 pixels 'file icon' (PNG only)",
	0x02: "Other file icon",
	0x03: "Cover (front)",
	0x04: "Cover (back)",
	0x05: "Leaflet page",
	0x06: "Media (e.g. label side of CD)",
	0x07: "Lead artist/lead performer/soloist",
	0x08: "Artist/performer",
	0x09: "Conductor",
	0x0A: "Band/Orchestra",
	0x0B: "Composer",
	0x0C: "Lyricist/text writer",
	0x0D: "Recording Location",
	0x0E: "During recording",
	0x0F: "During performance",
	0x10: "Movie/video screen capture",
	0x11: "A bright coloured fish",
	0x12: "Illustration",
	0x13: "Band/artist logotype",
	0x14: "Publisher/Studio logotype",
}

// interpretPictureType translates the byte into a string; the string values come from
// https://id3.org/id3v2.3.0#Attached_picture
func interpretPictureType(pictureType byte) string {
	if description, found := pictureTypes[pictureType]; found {
		return description
	}
	return fmt.Sprintf("Undocumented value %d", pictureType)
}

func interpretPictureFrame(frame *id3v2.PictureFrame) []string {
	dump := hexDump(frame.Picture)
	substrings := make([]string, 0, 5+len(dump))
	substrings = append(
		substrings,
		fmt.Sprintf("Encoding: %s", frame.Encoding.String()),
		fmt.Sprintf("Mime Type: %s", frame.MimeType),
		fmt.Sprintf("Picture Type: %s", interpretPictureType(frame.PictureType)),
		fmt.Sprintf("Description: %s", frame.Description),
		"Picture Data:",
	)
	substrings = append(substrings, dump...)
	return substrings
}

func bytesToInt(content []byte) int {
	result := 0
	for i := 0; i < len(content); i++ {
		result <<= 8
		result += int(content[i])
	}
	return result
}

func interpretUnknownFrame(data id3v2.UnknownFrame) []string {
	content := data.Body
	if result, ok := decodeFreeRipMCDI(content); ok {
		return result
	}
	if s, ok := displayString(content); ok {
		if result, ok := decodeWindowsLegacyMediaPlayerMCDI(s, content); ok {
			return result
		}
		// some other pattern, not seen before, so no attempt to decode the string further
		dump := hexDump(content)
		result := make([]string, 0, 1+len(dump))
		result = append(result, s)
		result = append(result, dump...)
		return result
	}
	if s, ok := decodeLAMEGeneratedMCDI(content); ok {
		return s
	}
	return hexDump(content)
}

func decodeLAMEGeneratedMCDI(content []byte) ([]string, bool) {
	// this code is based on inspection of content generated by LAME, and by reading this:
	// https://musicbrainz.org/doc/Disc_IDs_and_Tagging, which says:
	//
	// "Basic format in pseudo-C code (all are big-endian, MSB first):
	//
	//  struct cdtoc
	//  {
	//    unsigned short toc_data_length;
	//    unsigned char first_track_number;
	//    unsigned char last_track_number;
	//    /* the following fields are repeated once per track on
	//     * the CD, and then one extra time for the lead-out */
	//    unsigned char reserved1; /* This should be 0 */
	//    unsigned char adr_ctrl; /* first 4 bits for the ADR data last 4 bits for Control data */
	//    unsigned char track_number; /* This is 0xAA for the lead-out */
	//    unsigned char reserved2; /* This should be 0 */
	//    unsigned long lba_address; /* NOT MSF. */
	//    /* The lba_address may be misapplied in my hack of discid
	//     * (off by 150 frames, track 1 starts at LBA 0), but I don't
	//     * have a way to verify that this is the case. */
	//  };
	// "
	//
	// regarding the 150 frame correction applied here, this passage read from
	// https://musicbrainz.org/doc/Disc_ID_Calculation was most informative:
	//
	// "Also note that the LBA (Logical Block Address) offsets start at address 0, but the first
	//  track starts actually at 00:02:00 (the standard length of the lead-in track). So we need
	//  to add 150 logical blocks to each LBA offset."
	//
	// My experimentation with using LAME demonstrates that it generates LBA offsets that begin
	// with zero, and so need the 150 frame logical block correction
	contentLength := len(content)
	if contentLength >= 4 {
		result := bytesToInt(content[0:2])
		if result < contentLength {
			trackFirst := int(content[2])
			trackLast := int(content[3])
			trackCount := trackLast + 1 - trackFirst
			expectedLength := 2 + (8 * (trackCount + 1))
			if expectedLength == result {
				dump := hexDump(content)
				formatted := make([]string, 0, trackCount+3+len(dump))
				formatted = append(
					formatted,
					fmt.Sprintf("first track: %d", trackFirst),
					fmt.Sprintf("last track: %d", trackLast))
				const logicalBlockAddressCorrection = 150
				for k := 0; k < trackCount; k++ {
					offset := k * 8
					lbaAddress := bytesToInt(content[offset+8:offset+12]) + logicalBlockAddressCorrection
					formatted = append(
						formatted,
						fmt.Sprintf("track %d logical block address %d", content[6+offset], lbaAddress),
					)
				}
				offset := trackCount * 8
				lbaAddress := bytesToInt(content[offset+8:offset+12]) + logicalBlockAddressCorrection
				formatted = append(formatted, fmt.Sprintf("leadout track logical block address %d", lbaAddress))
				formatted = append(formatted, dump...)
				return formatted, true
			}
		}
	}
	return nil, false
}

func decodeWindowsLegacyMediaPlayerMCDI(s string, raw []byte) ([]string, bool) {
	if windowsLegacyMCDIPattern.MatchString(s) {
		// windows legacy media player generates a string of hexadecimal numbers separated by plus
		// signs. The first number is the number of tracks; the remaining numbers are the logical
		// block addresses of the tracks. There will be one address for each track, plus one for the
		// leadout.
		hexStrings := strings.Split(s, "+")
		// nilaway is not convinced that s, having matched the windows legacy pattern, couldn't
		// somehow return a nil slice of string, so we make the check that should never fail
		if len(hexStrings) != 0 {
			track := 0
			// ignoring count and error returns because the regex the string matched
			// only matches a sequence of hexadecimal numbers separated by '+' characters
			_, _ = fmt.Sscanf(hexStrings[0], "%x", &track)
			if len(hexStrings) == track+2 {
				addresses := make([]int, 0, track+1)
				for _, ss := range hexStrings[1:] {
					address := 0
					_, _ = fmt.Sscanf(ss, "%x", &address)
					addresses = append(addresses, address)
				}
				dump := hexDump(raw)
				substrings := make([]string, 0, len(addresses)+1+len(dump))
				substrings = append(substrings, fmt.Sprintf("tracks %d", track))
				for n, address := range addresses {
					if n == track {
						substrings = append(substrings, fmt.Sprintf("leadout track logical block address %d", address))
					} else {
						substrings = append(substrings, fmt.Sprintf("track %d logical block address %d", n+1, address))
					}
				}
				substrings = append(substrings, dump...)
				return substrings, true
			}
		}
	}
	return nil, false
}

func decodeFreeRipMCDI(content []byte) ([]string, bool) {
	if len(content) >= 3 && content[0] == 1 && content[1] == 0xff && content[2] == 0xfe {
		if s, ok := displayString(content[3:]); ok {
			dump := hexDump(content)
			result := make([]string, 0, 1+len(dump))
			result = append(result, s)
			result = append(result, dump...)
			return result, true
		}
	}
	return nil, false
}

func hexDump(content []byte) []string {
	values := make([]string, 0, (len(content)+15)/16)
	fullLines := len(content) / 16
	for j := 0; j < fullLines; j++ {
		hex := make([]string, 0, 16)
		printable := make([]string, 0, 16)
		for k := 0; k < 16; k++ {
			currentChar := content[(j*16)+k]
			hex = append(hex, fmt.Sprintf("%02X", currentChar))
			if currentChar >= ' ' && currentChar <= '~' {
				printable = append(printable, fmt.Sprintf("%c", currentChar))
			} else {
				printable = append(printable, "•")
			}
		}
		hex = append(hex, strings.Join(printable, ""))
		values = append(values, strings.Join(hex, " "))
	}
	if len(content)%16 != 0 {
		remainder := len(content) % 16
		hex := make([]string, 0, 16)
		printable := make([]string, 0, remainder)
		offset := 16 * (len(content) / 16)
		for k := 0; k < 16; k++ {
			if k < remainder {
				currentChar := content[offset+k]
				hex = append(hex, fmt.Sprintf("%02X", currentChar))
				if currentChar >= ' ' && currentChar <= '~' {
					printable = append(printable, fmt.Sprintf("%c", currentChar))
				} else {
					printable = append(printable, "•")
				}
			} else {
				hex = append(hex, "  ")
			}
		}
		hex = append(hex, strings.Join(printable, ""))
		values = append(values, strings.Join(hex, " "))
	}
	return values
}

func displayString(content []byte) (string, bool) {
	if len(content)%2 == 0 {
		allOddBytesAreNull := true
		ascii := make([]byte, 0, len(content)/2)
		for index, b := range content {
			if index%2 == 0 {
				ascii = append(ascii, b)
			} else if b != 0 {
				allOddBytesAreNull = false
				break
			}
		}
		if allOddBytesAreNull {
			for ascii[len(ascii)-1] == 0x00 {
				ascii = ascii[:len(ascii)-1]
			}
			return string(ascii), true
		}
	}
	return "", false
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
