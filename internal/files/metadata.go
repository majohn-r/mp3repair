package files

import (
	"bytes"

	"github.com/bogem/id3v2/v2"
)

// SourceType identifies the source of a particular form of metadata
type SourceType int

const (
	UndefinedSource SourceType = iota
	ID3V1
	ID3V2
	TotalSources
)

var (
	nameComparators = map[SourceType]func(comparableStrings) bool{
		ID3V1: id3v1NameDiffers,
		ID3V2: id3v2NameDiffers,
	}
	genreComparators = map[SourceType]func(comparableStrings) bool{
		ID3V1: id3v1GenreDiffers,
		ID3V2: id3v2GenreDiffers,
	}
	metadataUpdaters = map[SourceType]func(tM *trackMetadata, path string, src SourceType) error{
		ID3V1: updateID3V1Metadata,
		ID3V2: updateID3V2Metadata,
	}
	sourceTypes = []SourceType{ID3V1, ID3V2}
)

func (sT SourceType) name() string {
	switch sT {
	case ID3V1:
		return "ID3V1"
	case ID3V2:
		return "ID3V2"
	case TotalSources:
		return "total"
	default:
		return "undefined"
	}
}

// outside of unit tests
type trackMetadata struct {
	album             []string
	artist            []string
	title             []string
	genre             []string
	year              []string
	track             []int
	musicCDIdentifier id3v2.UnknownFrame
	canonicalType     SourceType
	errCause          []string
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
		album:           make([]string, TotalSources),
		artist:          make([]string, TotalSources),
		title:           make([]string, TotalSources),
		genre:           make([]string, TotalSources),
		year:            make([]string, TotalSources),
		track:           make([]int, TotalSources),
		errCause:        make([]string, TotalSources),
		correctedAlbum:  make([]string, TotalSources),
		correctedArtist: make([]string, TotalSources),
		correctedTitle:  make([]string, TotalSources),
		correctedGenre:  make([]string, TotalSources),
		correctedYear:   make([]string, TotalSources),
		correctedTrack:  make([]int, TotalSources),
		requiresEdit:    make([]bool, TotalSources),
	}
}

func readMetadata(path string) *trackMetadata {
	v1, id3v1Err := internalReadID3V1Metadata(path, fileReader)
	d := rawReadID3V2Metadata(path)
	tM := newTrackMetadata()
	switch {
	case id3v1Err != nil && d.err != nil:
		tM.errCause[ID3V1] = id3v1Err.Error()
		tM.errCause[ID3V2] = d.err.Error()
	case id3v1Err != nil:
		tM.errCause[ID3V1] = id3v1Err.Error()
		tM.setID3v2Values(d)
		tM.canonicalType = ID3V2
	case d.err != nil:
		tM.errCause[ID3V2] = d.err.Error()
		tM.setID3v1Values(v1)
		tM.canonicalType = ID3V1
	default:
		tM.setID3v2Values(d)
		tM.setID3v1Values(v1)
		tM.canonicalType = ID3V2
	}
	return tM
}

func (tM *trackMetadata) setID3v2Values(d *id3v2Metadata) {
	i := ID3V2
	tM.album[i] = d.album
	tM.artist[i] = d.artist
	tM.title[i] = d.title
	tM.genre[i] = d.genre
	tM.year[i] = d.year
	tM.track[i] = d.track
	tM.musicCDIdentifier = d.musicCDIdentifier
}

func (tM *trackMetadata) setID3v1Values(v1 *id3v1Metadata) {
	index := ID3V1
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
	return tM.canonicalType == ID3V1 || tM.canonicalType == ID3V2
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

func (tM *trackMetadata) errorCauses() []string {
	errCauses := []string{}
	for _, e := range tM.errCause {
		if e != "" {
			errCauses = append(errCauses, e)
		}
	}
	return errCauses
}

type comparableStrings struct {
	externalName string
	metadataName string
}

func (tM *trackMetadata) trackDiffers(track int) (differs bool) {
	for _, sT := range sourceTypes {
		if tM.errCause[sT] == "" && tM.track[sT] != track {
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
		if tM.errCause[sT] == "" && nameComparators[sT](comparison) {
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
		if tM.errCause[sT] == "" && nameComparators[sT](comparison) {
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
		if tM.errCause[sT] == "" && nameComparators[sT](comparison) {
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
		if tM.errCause[sT] == "" && genreComparators[sT](comparison) {
			differs = true
			tM.requiresEdit[sT] = true
			tM.correctedGenre[sT] = genre
		}
	}
	return
}

func (tM *trackMetadata) yearDiffers(year string) (differs bool) {
	for _, sT := range sourceTypes {
		if tM.errCause[sT] == "" && tM.year[sT] != year {
			differs = true
			tM.requiresEdit[sT] = true
			tM.correctedYear[sT] = year
		}
	}
	return
}

func (tM *trackMetadata) mcdiDiffers(f id3v2.UnknownFrame) (differs bool) {
	if tM.errCause[ID3V2] == "" && !bytes.Equal(tM.musicCDIdentifier.Body, f.Body) {
		differs = true
		tM.requiresEdit[ID3V2] = true
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

func updateMetadata(tM *trackMetadata, path string) (e []error) {
	for _, source := range sourceTypes {
		if err := metadataUpdaters[source](tM, path, source); err != nil {
			e = append(e, err)
		}
	}
	return
}
