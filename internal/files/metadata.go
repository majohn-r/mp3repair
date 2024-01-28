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
	metadataUpdaters = map[SourceType]func(tM *TrackMetadata, path string, src SourceType) error{
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
type TrackMetadata struct {
	Album             []string           // public so unit tests can set it and force a difference
	Artist            []string           // public so unit tests can set it and force a difference
	Title             []string           // public so unit tests can set it and force a difference
	Genre             []string           // public so unit tests can set it and force a difference
	Year              []string           // public so unit tests can set it and force a difference
	Track             []int              // public so unit tests can set it and force a difference
	MusicCDIdentifier id3v2.UnknownFrame // public so unit tests can set it and force a difference
	CanonicalType     SourceType
	ErrCause          []string
	// these fields are set by the various xDiffers methods
	CorrectedAlbum             []string
	CorrectedArtist            []string
	CorrectedTitle             []string
	CorrectedGenre             []string
	CorrectedYear              []string
	CorrectedTrack             []int
	CorrectedMusicCDIdentifier id3v2.UnknownFrame
	RequiresEdit               []bool
}

func newTrackMetadata() *TrackMetadata {
	return &TrackMetadata{
		Album:           make([]string, TotalSources),
		Artist:          make([]string, TotalSources),
		Title:           make([]string, TotalSources),
		Genre:           make([]string, TotalSources),
		Year:            make([]string, TotalSources),
		Track:           make([]int, TotalSources),
		ErrCause:        make([]string, TotalSources),
		CorrectedAlbum:  make([]string, TotalSources),
		CorrectedArtist: make([]string, TotalSources),
		CorrectedTitle:  make([]string, TotalSources),
		CorrectedGenre:  make([]string, TotalSources),
		CorrectedYear:   make([]string, TotalSources),
		CorrectedTrack:  make([]int, TotalSources),
		RequiresEdit:    make([]bool, TotalSources),
	}
}

func readMetadata(path string) *TrackMetadata {
	v1, id3v1Err := internalReadID3V1Metadata(path, fileReader)
	d := rawReadID3V2Metadata(path)
	tM := newTrackMetadata()
	switch {
	case id3v1Err != nil && d.err != nil:
		tM.ErrCause[ID3V1] = id3v1Err.Error()
		tM.ErrCause[ID3V2] = d.err.Error()
	case id3v1Err != nil:
		tM.ErrCause[ID3V1] = id3v1Err.Error()
		tM.setID3v2Values(d)
		tM.CanonicalType = ID3V2
	case d.err != nil:
		tM.ErrCause[ID3V2] = d.err.Error()
		tM.setID3v1Values(v1)
		tM.CanonicalType = ID3V1
	default:
		tM.setID3v2Values(d)
		tM.setID3v1Values(v1)
		tM.CanonicalType = ID3V2
	}
	return tM
}

func (tM *TrackMetadata) setID3v2Values(d *id3v2Metadata) {
	i := ID3V2
	tM.Album[i] = d.album
	tM.Artist[i] = d.artist
	tM.Title[i] = d.title
	tM.Genre[i] = d.genre
	tM.Year[i] = d.year
	tM.Track[i] = d.track
	tM.MusicCDIdentifier = d.musicCDIdentifier
}

func (tM *TrackMetadata) setID3v1Values(v1 *id3v1Metadata) {
	index := ID3V1
	tM.Album[index] = v1.album()
	tM.Artist[index] = v1.artist()
	tM.Title[index] = v1.title()
	if genre, ok := v1.genre(); ok {
		tM.Genre[index] = genre
	}
	tM.Year[index] = v1.year()
	if track, ok := v1.track(); ok {
		tM.Track[index] = track
	}
}

func (tM *TrackMetadata) isValid() bool {
	return tM.CanonicalType == ID3V1 || tM.CanonicalType == ID3V2
}

func (tM *TrackMetadata) canonicalArtist() string {
	return tM.Artist[tM.CanonicalType]
}

func (tM *TrackMetadata) canonicalAlbum() string {
	return tM.Album[tM.CanonicalType]
}

func (tM *TrackMetadata) canonicalGenre() string {
	return tM.Genre[tM.CanonicalType]
}

func (tM *TrackMetadata) canonicalYear() string {
	return tM.Year[tM.CanonicalType]
}

func (tM *TrackMetadata) canonicalMusicCDIdentifier() id3v2.UnknownFrame {
	return tM.MusicCDIdentifier
}

func (tM *TrackMetadata) errorCauses() []string {
	errCauses := []string{}
	for _, e := range tM.ErrCause {
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

func (tM *TrackMetadata) trackDiffers(track int) (differs bool) {
	for _, sT := range sourceTypes {
		if tM.ErrCause[sT] == "" && tM.Track[sT] != track {
			differs = true
			tM.RequiresEdit[sT] = true
			tM.CorrectedTrack[sT] = track
		}
	}
	return
}

func (tM *TrackMetadata) trackTitleDiffers(title string) (differs bool) {
	for _, sT := range sourceTypes {
		comparison := comparableStrings{externalName: title, metadataName: tM.Title[sT]}
		if tM.ErrCause[sT] == "" && nameComparators[sT](comparison) {
			differs = true
			tM.RequiresEdit[sT] = true
			tM.CorrectedTitle[sT] = title
		}
	}
	return
}

func (tM *TrackMetadata) albumTitleDiffers(albumTitle string) (differs bool) {
	for _, sT := range sourceTypes {
		comparison := comparableStrings{externalName: albumTitle, metadataName: tM.Album[sT]}
		if tM.ErrCause[sT] == "" && nameComparators[sT](comparison) {
			differs = true
			tM.RequiresEdit[sT] = true
			tM.CorrectedAlbum[sT] = albumTitle
		}
	}
	return
}

func (tM *TrackMetadata) artistNameDiffers(artistName string) (differs bool) {
	for _, sT := range sourceTypes {
		comparison := comparableStrings{externalName: artistName, metadataName: tM.Artist[sT]}
		if tM.ErrCause[sT] == "" && nameComparators[sT](comparison) {
			differs = true
			tM.RequiresEdit[sT] = true
			tM.CorrectedArtist[sT] = artistName
		}
	}
	return
}

func (tM *TrackMetadata) genreDiffers(genre string) (differs bool) {
	for _, sT := range sourceTypes {
		comparison := comparableStrings{externalName: genre, metadataName: tM.Genre[sT]}
		if tM.ErrCause[sT] == "" && genreComparators[sT](comparison) {
			differs = true
			tM.RequiresEdit[sT] = true
			tM.CorrectedGenre[sT] = genre
		}
	}
	return
}

func (tM *TrackMetadata) yearDiffers(year string) (differs bool) {
	for _, sT := range sourceTypes {
		if tM.ErrCause[sT] == "" && tM.Year[sT] != year {
			differs = true
			tM.RequiresEdit[sT] = true
			tM.CorrectedYear[sT] = year
		}
	}
	return
}

func (tM *TrackMetadata) mcdiDiffers(f id3v2.UnknownFrame) (differs bool) {
	if tM.ErrCause[ID3V2] == "" && !bytes.Equal(tM.MusicCDIdentifier.Body, f.Body) {
		differs = true
		tM.RequiresEdit[ID3V2] = true
		tM.CorrectedMusicCDIdentifier = f
	}
	return
}

func (tM *TrackMetadata) canonicalAlbumTitleMatches(albumTitle string) bool {
	comparison := comparableStrings{externalName: albumTitle, metadataName: tM.canonicalAlbum()}
	return !nameComparators[tM.CanonicalType](comparison)
}

func (tM *TrackMetadata) canonicalArtistNameMatches(artistName string) bool {
	comparison := comparableStrings{externalName: artistName, metadataName: tM.canonicalArtist()}
	return !nameComparators[tM.CanonicalType](comparison)
}

func updateMetadata(tM *TrackMetadata, path string) (e []error) {
	for _, source := range sourceTypes {
		if err := metadataUpdaters[source](tM, path, source); err != nil {
			e = append(e, err)
		}
	}
	return
}
