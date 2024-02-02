package files

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bogem/id3v2/v2"
)

type Id3v2Metadata struct {
	Album             string
	Artist            string
	Title             string
	Genre             string
	Year              string
	Track             int
	MusicCDIdentifier id3v2.UnknownFrame
	Err               error
}

func readID3V2Tag(path string) (*id3v2.Tag, error) {
	return id3v2.Open(path, id3v2.Options{Parse: true, ParseFrames: nil})
}

func RawReadID3V2Metadata(path string) (d *Id3v2Metadata) {
	d = &Id3v2Metadata{}
	if tag, err := readID3V2Tag(path); err != nil {
		d.Err = err
	} else {
		defer tag.Close()
		if trackNumber, err := ToTrackNumber(tag.GetTextFrame(trackFrame).Text); err != nil {
			d.Err = err
		} else {
			d.Album = RemoveLeadingBOMs(tag.Album())
			d.Artist = RemoveLeadingBOMs(tag.Artist())
			d.Genre = NormalizeGenre(RemoveLeadingBOMs(tag.Genre()))
			d.Title = RemoveLeadingBOMs(tag.Title())
			d.Track = trackNumber
			d.Year = RemoveLeadingBOMs(tag.Year())
			mcdiFramers := tag.AllFrames()[mcdiFrame]
			d.MusicCDIdentifier = SelectUnknownFrame(mcdiFramers)
		}
	}
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
	if n, err := fmt.Sscanf(s, "(%d)%s", &i, &value); n == 2 && err == nil {
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
)

func ToTrackNumber(s string) (i int, err error) {
	s = RemoveLeadingBOMs(s)
	if s == "" {
		err = ErrMissingTrackNumber
		return
	}
	// this is more complicated than I wanted, because some mp3 rippers produce
	// track numbers like "12/14", meaning 12th track of 14
	n := 0
	bs := []byte(s)
	for j, b := range bs {
		c := int(b)
		if c >= '0' && c <= '9' {
			n *= 10
			n += c - '0'
		} else {
			// found something other than a digit
			switch j {
			case 0: // never saw a digit
				err = ErrMalformedTrackNumber
				return
			default: // did read at least one digit
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

func updateID3V2Metadata(tM *TrackMetadata, path string, sT SourceType) (e error) {
	if tM.RequiresEdit[sT] {
		if tag, err := readID3V2Tag(path); err != nil {
			e = err
		} else {
			defer tag.Close()
			tag.SetDefaultEncoding(id3v2.EncodingUTF8)
			album := tM.CorrectedAlbum[sT]
			if album != "" {
				tag.SetAlbum(album)
			}
			artist := tM.CorrectedArtist[sT]
			if artist != "" {
				tag.SetArtist(artist)
			}
			title := tM.CorrectedTitle[sT]
			if title != "" {
				tag.SetTitle(title)
			}
			track := tM.CorrectedTrack[sT]
			if track != 0 {
				tag.AddTextFrame("TRCK", tag.DefaultEncoding(), fmt.Sprintf("%d", track))
			}
			genre := tM.CorrectedGenre[sT]
			if genre != "" {
				tag.SetGenre(genre)
			}
			year := tM.CorrectedYear[sT]
			if year != "" {
				tag.SetYear(year)
			}
			mcdi := tM.CorrectedMusicCDIdentifier
			if len(mcdi.Body) != 0 {
				tag.DeleteFrames(mcdiFrame)
				tag.AddFrame(mcdiFrame, mcdi)
			}
			e = tag.Save()
		}
	}
	return
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

func ReadID3V2Metadata(path string) (version byte, encoding string, frameStrings []string, rawFrames []*Id3v2TrackFrame, e error) {
	if tag, err := readID3V2Tag(path); err != nil {
		e = err
	} else {
		defer tag.Close()
		version = tag.Version()
		encoding = tag.DefaultEncoding().Name
		frameMap := tag.AllFrames()
		var frameNames []string
		for k := range frameMap {
			frameNames = append(frameNames, k)
		}
		sort.Strings(frameNames)
		for _, n := range frameNames {
			var value string
			if strings.HasPrefix(n, "T") {
				value = RemoveLeadingBOMs(tag.GetTextFrame(n).Text)
			} else {
				value = FramerSliceAsString(frameMap[n])
			}
			frame := &Id3v2TrackFrame{Name: n, Value: value}
			frameStrings = append(frameStrings, frame.String())
			rawFrames = append(rawFrames, frame)
		}
	}
	return
}

func FramerSliceAsString(f []id3v2.Framer) string {
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

func Id3v2NameDiffers(cS ComparableStrings) bool {
	externalName := strings.ToLower(cS.ExternalName)
	metadataName := strings.ToLower(cS.MetadataName)
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

func Id3v2GenreDiffers(cS ComparableStrings) bool {
	// differs unless exact match. Period.
	return cS.ExternalName != cS.MetadataName
}
