package cmd

import (
	"fmt"
	"mp3repair/internal/files"
	"reflect"
	"slices"

	"github.com/majohn-r/output"
)

type concernType int32

const (
	unspecifiedConcern concernType = iota
	emptyConcern
	filesConcern
	numberingConcern
	conflictConcern
)

var concernNames = map[concernType]string{
	emptyConcern:     "empty",
	filesConcern:     "files",
	numberingConcern: "numbering",
	conflictConcern:  "metadata conflict",
}

func concernName(i concernType) string {
	if s, found := concernNames[i]; found {
		return s
	}
	return fmt.Sprintf("concern %d", i)
}

type concerns struct {
	concernsCollection map[concernType][]string
}

func newConcerns() concerns {
	return concerns{concernsCollection: map[concernType][]string{}}
}

func (c concerns) addConcern(source concernType, concern string) {
	c.concernsCollection[source] = append(c.concernsCollection[source], concern)
}

func (c concerns) isConcerned() bool {
	for _, list := range c.concernsCollection {
		if len(list) > 0 {
			return true
		}
	}
	return false
}

func (c concerns) toConsole(o output.Bus) {
	if c.isConcerned() {
		cStrings := make([]string, 0, len(c.concernsCollection))
		for key, value := range c.concernsCollection {
			for _, s := range value {
				cStrings = append(cStrings, fmt.Sprintf("* [%s] %s", concernName(key), s))
			}
		}
		slices.Sort(cStrings)
		for _, s := range cStrings {
			o.ConsolePrintln(s)
		}
	}
}

type concernedTrack struct {
	concerns
	backing *files.Track
}

func newConcernedTrack(track *files.Track) *concernedTrack {
	if track == nil {
		return nil
	}
	return &concernedTrack{
		concerns: newConcerns(),
		backing:  track,
	}
}

func (cT *concernedTrack) addConcern(source concernType, concern string) {
	cT.concerns.addConcern(source, concern)
}

func (cT *concernedTrack) isConcerned() bool {
	return cT.concerns.isConcerned()
}

func (cT *concernedTrack) name() string {
	return cT.backing.Name()
}

func (cT *concernedTrack) toConsole(o output.Bus) {
	if cT.isConcerned() {
		o.ConsolePrintf("Track %q\n", cT.name())
		cT.concerns.toConsole(o)
	}
}

func (cT *concernedTrack) backingTrack() *files.Track {
	return cT.backing
}

type concernedAlbum struct {
	concerns
	concernedTracks []*concernedTrack
	backing         *files.Album
	trackMap        map[string]*concernedTrack
}

func newConcernedAlbum(album *files.Album) *concernedAlbum {
	if album == nil {
		return nil
	}
	cAl := &concernedAlbum{
		concerns:        newConcerns(),
		concernedTracks: make([]*concernedTrack, 0, len(album.Tracks())),
		backing:         album,
		trackMap:        map[string]*concernedTrack{},
	}
	for _, track := range album.Tracks() {
		cAl.addTrack(track)
	}
	return cAl
}

func (cAl *concernedAlbum) addConcern(source concernType, concern string) {
	cAl.concerns.addConcern(source, concern)
}

func (cAl *concernedAlbum) addTrack(track *files.Track) {
	if cT := newConcernedTrack(track); cT != nil {
		cAl.concernedTracks = append(cAl.concernedTracks, cT)
		cAl.trackMap[cT.backing.FileName()] = cT
	}
}

func (cAl *concernedAlbum) backingAlbum() *files.Album {
	return cAl.backing
}

func (cAl *concernedAlbum) isConcerned() bool {
	if cAl.concerns.isConcerned() {
		return true
	}
	for _, cT := range cAl.concernedTracks {
		if cT.isConcerned() {
			return true
		}
	}
	return false
}

func (cAl *concernedAlbum) name() string {
	return cAl.backing.Title()
}

func (cAl *concernedAlbum) lookup(track *files.Track) *concernedTrack {
	var cT *concernedTrack
	if track, found := cAl.trackMap[track.FileName()]; found {
		cT = track
	}
	return cT
}

func (cAl *concernedAlbum) rollup() bool {
	if len(cAl.concernedTracks) <= 1 {
		return false
	}
	if !cAl.isConcerned() {
		return false
	}
	initialConcerns := cAl.concernedTracks[0].concernsCollection
	for _, cT := range cAl.concernedTracks[1:] {
		if !reflect.DeepEqual(initialConcerns, cT.concernsCollection) {
			return false
		}
	}
	mergeConcerns(cAl.concernsCollection, initialConcerns, "for all tracks:")
	for _, cT := range cAl.concernedTracks {
		cT.concerns = newConcerns()
	}
	return true
}

func (cAl *concernedAlbum) toConsole(o output.Bus) {
	if cAl.isConcerned() {
		o.ConsolePrintf("Album %q\n", cAl.name())
		cAl.concerns.toConsole(o)
		m := map[string]*concernedTrack{}
		names := make([]string, 0, len(cAl.concernedTracks))
		for _, cT := range cAl.concernedTracks {
			trackName := cT.name()
			m[trackName] = cT
			names = append(names, trackName)
		}
		slices.Sort(names)
		o.IncrementTab(2)
		for _, name := range names {
			if cT := m[name]; cT != nil {
				cT.toConsole(o)
			}
		}
		o.DecrementTab(2)
	}
}

func (cAl *concernedAlbum) tracks() []*concernedTrack {
	return cAl.concernedTracks
}

type concernedArtist struct {
	concerns
	concernedAlbums []*concernedAlbum
	backing         *files.Artist
	albumMap        map[string]*concernedAlbum
}

func newConcernedArtist(artist *files.Artist) *concernedArtist {
	if artist == nil {
		return nil
	}
	cAr := &concernedArtist{
		concerns:        newConcerns(),
		concernedAlbums: make([]*concernedAlbum, 0, len(artist.Albums())),
		backing:         artist,
		albumMap:        map[string]*concernedAlbum{},
	}
	for _, album := range artist.Albums() {
		cAr.addAlbum(album)
	}
	return cAr
}

func (cAr *concernedArtist) addAlbum(album *files.Album) {
	if cAl := newConcernedAlbum(album); cAl != nil {
		cAr.concernedAlbums = append(cAr.concernedAlbums, cAl)
		cAr.albumMap[cAl.name()] = cAl
	}
}

func (cAr *concernedArtist) addConcern(source concernType, concern string) {
	cAr.concerns.addConcern(source, concern)
}

func (cAr *concernedArtist) albums() []*concernedAlbum {
	return cAr.concernedAlbums
}

func (cAr *concernedArtist) backingArtist() *files.Artist {
	return cAr.backing
}

func (cAr *concernedArtist) isConcerned() bool {
	if cAr.concerns.isConcerned() {
		return true
	}
	for _, cAl := range cAr.concernedAlbums {
		if cAl.isConcerned() {
			return true
		}
	}
	return false
}

func (cAr *concernedArtist) lookup(track *files.Track) *concernedTrack {
	albumKey := track.AlbumName()
	if cAl, found := cAr.albumMap[albumKey]; found {
		return cAl.lookup(track)
	}
	return nil
}

func (cAr *concernedArtist) name() string {
	return cAr.backing.Name()
}

func (cAr *concernedArtist) rollup() bool {
	if !cAr.isConcerned() {
		return false
	}
	for _, cAl := range cAr.concernedAlbums {
		cAl.rollup()
	}
	if len(cAr.concernedAlbums) <= 1 {
		return false
	}
	initialConcerns := cAr.concernedAlbums[0].concernsCollection
	for _, cAl := range cAr.concernedAlbums[1:] {
		if !reflect.DeepEqual(initialConcerns, cAl.concernsCollection) {
			return false
		}
	}
	mergeConcerns(cAr.concernsCollection, initialConcerns, "for all albums:")
	for _, cAl := range cAr.concernedAlbums {
		cAl.concerns = newConcerns()
	}
	return true
}

func (cAr *concernedArtist) toConsole(o output.Bus) {
	if cAr.isConcerned() {
		o.ConsolePrintf("Artist %q\n", cAr.name())
		cAr.concerns.toConsole(o)
		m := map[string]*concernedAlbum{}
		names := make([]string, 0, len(cAr.concernedAlbums))
		for _, cT := range cAr.concernedAlbums {
			albumName := cT.name()
			m[albumName] = cT
			names = append(names, albumName)
		}
		slices.Sort(names)
		o.IncrementTab(2)
		for _, name := range names {
			if cAl := m[name]; cAl != nil {
				cAl.toConsole(o)
			}
		}
		o.DecrementTab(2)
	}
}

func mergeConcerns(initial, addition map[concernType][]string, prefix string) {
	for concern, issues := range addition {
		for _, issue := range issues {
			initial[concern] = append(initial[concern], fmt.Sprintf("%s %s", prefix, issue))
		}
	}
}

func createConcernedArtists(artists []*files.Artist) []*concernedArtist {
	concernedArtists := make([]*concernedArtist, 0, len(artists))
	for _, artist := range artists {
		if cAr := newConcernedArtist(artist); cAr != nil {
			concernedArtists = append(concernedArtists, cAr)
		}
	}
	return concernedArtists
}
