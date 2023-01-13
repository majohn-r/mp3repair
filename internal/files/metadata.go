package files

import (
	"bytes"

	"github.com/bogem/id3v2/v2"
)

type sourceType int

const (
	undefinedSource sourceType = iota
	id3v1Source
	id3v2Source
	totalSources
)

var (
	nameComparators = map[sourceType]func(comparableStrings) bool{
		id3v1Source: id3v1NameDiffers,
		id3v2Source: id3v2NameDiffers,
	}
	genreComparators = map[sourceType]func(comparableStrings) bool{
		id3v1Source: id3v1GenreDiffers,
		id3v2Source: id3v2GenreDiffers,
	}
	tagEditors = map[sourceType]func(tM *trackMetadata, path string, src sourceType) error{
		id3v1Source: updateID3V1Tag,
		id3v2Source: updateID3V2Tag,
	}
	sourceTypes = []sourceType{id3v1Source, id3v2Source}
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
	err               []error
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
		err:             make([]error, totalSources),
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
	v1, id3v1Err := internalReadID3V1Metadata(path, fileReader)
	d := rawReadID3V2Tag(path)
	tM := newTrackMetadata()
	switch {
	case id3v1Err != nil && d.err != nil:
		tM.err[id3v1Source] = id3v1Err
		tM.err[id3v2Source] = d.err
	case id3v1Err != nil:
		tM.err[id3v1Source] = id3v1Err
		tM.setID3v2Values(d)
		tM.canonicalType = id3v2Source
	case d.err != nil:
		tM.err[id3v2Source] = d.err
		tM.setID3v1Values(v1)
		tM.canonicalType = id3v1Source
	default:
		tM.setID3v2Values(d)
		tM.setID3v1Values(v1)
		tM.canonicalType = id3v2Source
	}
	return tM
}

func (tM *trackMetadata) setID3v2Values(d *id3v2TaggedTrackData) {
	i := id3v2Source
	tM.album[i] = d.album
	tM.artist[i] = d.artist
	tM.title[i] = d.title
	tM.genre[i] = d.genre
	tM.year[i] = d.year
	tM.track[i] = d.track
	tM.musicCDIdentifier = d.musicCDIdentifier
}

func (tM *trackMetadata) setID3v1Values(v1 *id3v1Metadata) {
	index := id3v1Source
	tM.album[index] = v1.album()
	tM.artist[index] = v1.artist()
	tM.title[index] = v1.title()
	if genre, ok := v1.genre(); ok {
		tM.genre[index] = genre
	}
	tM.year[index] = v1.year()
	if track, ok := v1.track(); ok {
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

func (tM *trackMetadata) errors() (errs []error) {
	for _, e := range tM.err {
		if e != nil {
			errs = append(errs, e)
		}
	}
	return
}

type comparableStrings struct {
	externalName string
	metadataName string
}

func (tM *trackMetadata) trackDiffers(track int) (differs bool) {
	for _, sT := range sourceTypes {
		if tM.err[sT] == nil && tM.track[sT] != track {
			differs = true
			tM.requiresEdit[sT] = true
			tM.correctedTrack[sT] = track
		}
	}
	return
}

func (tM *trackMetadata) trackTitleDiffers(title string) (differs bool) {
	for _, sT := range sourceTypes {
		comparison := comparableStrings{externalName: title, metadataName: tM.title[sT]}
		if tM.err[sT] == nil && nameComparators[sT](comparison) {
			differs = true
			tM.requiresEdit[sT] = true
			tM.correctedTitle[sT] = title
		}
	}
	return
}

func (tM *trackMetadata) albumTitleDiffers(albumTitle string) (differs bool) {
	for _, sT := range sourceTypes {
		comparison := comparableStrings{externalName: albumTitle, metadataName: tM.album[sT]}
		if tM.err[sT] == nil && nameComparators[sT](comparison) {
			differs = true
			tM.requiresEdit[sT] = true
			tM.correctedAlbum[sT] = albumTitle
		}
	}
	return
}

func (tM *trackMetadata) artistNameDiffers(artistName string) (differs bool) {
	for _, sT := range sourceTypes {
		comparison := comparableStrings{externalName: artistName, metadataName: tM.artist[sT]}
		if tM.err[sT] == nil && nameComparators[sT](comparison) {
			differs = true
			tM.requiresEdit[sT] = true
			tM.correctedArtist[sT] = artistName
		}
	}
	return
}

func (tM *trackMetadata) genreDiffers(genre string) (differs bool) {
	for _, sT := range sourceTypes {
		comparison := comparableStrings{externalName: genre, metadataName: tM.genre[sT]}
		if tM.err[sT] == nil && genreComparators[sT](comparison) {
			differs = true
			tM.requiresEdit[sT] = true
			tM.correctedGenre[sT] = genre
		}
	}
	return
}

func (tM *trackMetadata) yearDiffers(year string) (differs bool) {
	for _, sT := range sourceTypes {
		if tM.err[sT] == nil && tM.year[sT] != year {
			differs = true
			tM.requiresEdit[sT] = true
			tM.correctedYear[sT] = year
		}
	}
	return
}

func (tM *trackMetadata) mcdiDiffers(f id3v2.UnknownFrame) (differs bool) {
	if tM.err[id3v2Source] == nil && !bytes.Equal(tM.musicCDIdentifier.Body, f.Body) {
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

func editTags(tM *trackMetadata, path string) (e []error) {
	for _, source := range sourceTypes {
		if err := tagEditors[source](tM, path, source); err != nil {
			e = append(e, err)
		}
	}
	return
}
