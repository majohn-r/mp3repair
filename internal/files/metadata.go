package files

import (
	"github.com/bogem/id3v2/v2"
)

type sourceType int

const (
	undefinedSource sourceType = iota
	id3v1Source
	id3v2Source
	totalSources
)

// outside of unit tests
type trackMetadata struct {
	album             []string
	artist            []string
	title             []string
	genre             []string
	year              []string
	track             []int
	musicCDIdentifier id3v2.UnknownFrame
	canonicalType     sourceType
	err               []string
	// these fields are set by the various xDiffers methods
	correctedAlbum             []string
	correctedArtist            []string
	correctedTitle             []string
	correctedGenre             []string
	correctedYear              []string
	correctedTrack             []int
	correctedMusicCDIdentifier id3v2.UnknownFrame
	requiresEdit               []bool
}

func newTrackMetadata() *trackMetadata {
	return &trackMetadata{
		album:           make([]string, totalSources),
		artist:          make([]string, totalSources),
		title:           make([]string, totalSources),
		genre:           make([]string, totalSources),
		year:            make([]string, totalSources),
		track:           make([]int, totalSources),
		err:             make([]string, totalSources),
		correctedAlbum:  make([]string, totalSources),
		correctedArtist: make([]string, totalSources),
		correctedTitle:  make([]string, totalSources),
		correctedGenre:  make([]string, totalSources),
		correctedYear:   make([]string, totalSources),
		correctedTrack:  make([]int, totalSources),
		requiresEdit:    make([]bool, totalSources),
	}
}

func readMetadata(path string) *trackMetadata {
	v1, id3v1Err := internalReadId3V1Metadata(path, fileReader)
	d := RawReadID3V2Tag(path)
	tM := newTrackMetadata()
	switch {
	case id3v1Err != nil && len(d.err) != 0:
		tM.err[id3v1Source] = id3v1Err.Error()
		tM.err[id3v2Source] = d.err
	case id3v1Err != nil:
		tM.err[id3v1Source] = id3v1Err.Error()
		tM.setId3v2Values(d)
		tM.canonicalType = id3v2Source
	case len(d.err) != 0:
		tM.err[id3v2Source] = d.err
		tM.setId3v1Values(v1)
		tM.canonicalType = id3v1Source
	default:
		tM.setId3v2Values(d)
		tM.setId3v1Values(v1)
		tM.canonicalType = id3v2Source
	}
	return tM
}

func (tM *trackMetadata) setId3v2Values(d *ID3V2TaggedTrackData) {
	index := id3v2Source
	tM.album[index] = d.album
	tM.artist[index] = d.artist
	tM.title[index] = d.title
	tM.genre[index] = d.genre
	tM.year[index] = d.year
	tM.track[index] = d.track
	tM.musicCDIdentifier = d.musicCDIdentifier
}

func (tM *trackMetadata) setId3v1Values(v1 *id3v1Metadata) {
	index := id3v1Source
	tM.album[index] = v1.getAlbum()
	tM.artist[index] = v1.getArtist()
	tM.title[index] = v1.getTitle()
	if genre, ok := v1.getGenre(); ok {
		tM.genre[index] = genre
	}
	tM.year[index] = v1.getYear()
	if track, ok := v1.getTrack(); ok {
		tM.track[index] = track
	}
}

func (tM *trackMetadata) isValid() bool {
	return tM.canonicalType == id3v1Source || tM.canonicalType == id3v2Source
}

func (tM *trackMetadata) canonicalArtist() string {
	return tM.artist[tM.canonicalType]
}

func (tM *trackMetadata) canonicalAlbum() string {
	return tM.album[tM.canonicalType]
}

func (tM *trackMetadata) canonicalGenre() string {
	return tM.genre[tM.canonicalType]
}

func (tM *trackMetadata) canonicalYear() string {
	return tM.year[tM.canonicalType]
}

func (tM *trackMetadata) canonicalMusicCDIdentifier() id3v2.UnknownFrame {
	return tM.musicCDIdentifier
}

func (tM *trackMetadata) errors() []string {
	s := []string{}
	for _, e := range tM.err {
		if len(e) > 0 {
			s = append(s, e)
		}
	}
	return s
}

type comparableStrings struct {
	externalName string
	metadataName string
}

var (
	nameComparators = map[sourceType]func(comparableStrings) bool{
		id3v1Source: id3v1NameDiffers,
		id3v2Source: id3v2NameDiffers,
	}
	genreComparators = map[sourceType]func(comparableStrings) bool{
		id3v1Source: id3v1GenreDiffers,
		id3v2Source: id3v2GenreDiffers,
	}
)

func (tM *trackMetadata) trackDiffers(track int) (differs bool) {
	for _, source := range []sourceType{id3v1Source, id3v2Source} {
		if len(tM.err[source]) == 0 && tM.track[source] != track {
			differs = true
			tM.requiresEdit[source] = true
			tM.correctedTrack[source] = track
		}
	}
	return
}

func (tM *trackMetadata) trackTitleDiffers(title string) (differs bool) {
	for _, source := range []sourceType{id3v1Source, id3v2Source} {
		comparison := comparableStrings{externalName: title, metadataName: tM.title[source]}
		if len(tM.err[source]) == 0 && nameComparators[source](comparison) {
			differs = true
			tM.requiresEdit[source] = true
			tM.correctedTitle[source] = title
		}
	}
	return
}

func (tM *trackMetadata) albumTitleDiffers(albumTitle string) (differs bool) {
	for _, source := range []sourceType{id3v1Source, id3v2Source} {
		comparison := comparableStrings{externalName: albumTitle, metadataName: tM.album[source]}
		if len(tM.err[source]) == 0 && nameComparators[source](comparison) {
			differs = true
			tM.requiresEdit[source] = true
			tM.correctedAlbum[source] = albumTitle
		}
	}
	return
}

func (tM *trackMetadata) artistNameDiffers(artistName string) (differs bool) {
	for _, source := range []sourceType{id3v1Source, id3v2Source} {
		comparison := comparableStrings{externalName: artistName, metadataName: tM.artist[source]}
		if len(tM.err[source]) == 0 && nameComparators[source](comparison) {
			differs = true
			tM.requiresEdit[source] = true
			tM.correctedArtist[source] = artistName
		}
	}
	return
}

func (tM *trackMetadata) genreDiffers(genre string) (differs bool) {
	for _, source := range []sourceType{id3v1Source, id3v2Source} {
		comparison := comparableStrings{externalName: genre, metadataName: tM.genre[source]}
		if len(tM.err[source]) == 0 && genreComparators[source](comparison) {
			differs = true
			tM.requiresEdit[source] = true
			tM.correctedGenre[source] = genre
		}
	}
	return
}

func (tM *trackMetadata) yearDiffers(year string) (differs bool) {
	for _, source := range []sourceType{id3v1Source, id3v2Source} {
		if len(tM.err[source]) == 0 && tM.year[source] != year {
			differs = true
			tM.requiresEdit[source] = true
			tM.correctedYear[source] = year
		}
	}
	return
}

func (tM *trackMetadata) mcdiDiffers(f id3v2.UnknownFrame) (differs bool) {
	if len(tM.err[id3v2Source]) == 0 && string(tM.musicCDIdentifier.Body) != string(f.Body) {
		differs = true
		tM.requiresEdit[id3v2Source] = true
		tM.correctedMusicCDIdentifier = f
	}
	return
}

func (tM *trackMetadata) canonicalAlbumTitleMatches(albumTitle string) bool {
	comparison := comparableStrings{externalName: albumTitle, metadataName: tM.canonicalAlbum()}
	return !nameComparators[tM.canonicalType](comparison)
}

func (tM *trackMetadata) canonicalArtistNameMatches(artistName string) bool {
	comparison := comparableStrings{externalName: artistName, metadataName: tM.canonicalArtist()}
	return !nameComparators[tM.canonicalType](comparison)
}

var (
	tagEditors = map[sourceType]func(t *Track, src sourceType) error{
		id3v1Source: updateID3V1Tag,
		id3v2Source: updateID3V2Tag,
	}
)

func editTags(t *Track) (e []error) {
	for _, source := range []sourceType{id3v1Source, id3v2Source} {
		if err := tagEditors[source](t, source); err != nil {
			e = append(e, err)
		}
	}
	return
}
