package cmd

import (
	"fmt"
	"mp3/internal/files"
	"slices"

	"github.com/majohn-r/output"
)

type ConcernType int32

const (
	UnspecifiedConcern ConcernType = iota
	EmptyConcern
	FilesConcern
	NumberingConcern
	ConflictConcern
)

var concernNames = map[ConcernType]string{
	EmptyConcern:     "empty",
	FilesConcern:     "files",
	NumberingConcern: "numbering",
	ConflictConcern:  "metadata conflict",
}

func ConcernName(i ConcernType) string {
	if s, ok := concernNames[i]; ok {
		return s
	}
	return fmt.Sprintf("concern %d", i)
}

type Concerns struct {
	concerns map[ConcernType][]string
}

func NewConcerns() Concerns {
	return Concerns{concerns: map[ConcernType][]string{}}
}

func (c Concerns) AddConcern(source ConcernType, concern string) {
	c.concerns[source] = append(c.concerns[source], concern)
}

func (c Concerns) IsConcerned() bool {
	for _, list := range c.concerns {
		if len(list) > 0 {
			return true
		}
	}
	return false
}

func (c Concerns) ToConsole(o output.Bus, tab int) {
	if c.IsConcerned() {
		cStrings := []string{}
		for key, value := range c.concerns {
			for _, s := range value {
				cStrings = append(cStrings, fmt.Sprintf("* [%s] %s", ConcernName(key), s))
			}
		}
		slices.Sort(cStrings)
		for _, s := range cStrings {
			o.WriteConsole("%*s%s\n", tab, "", s)
		}
	}
}

type ConcernedTrack struct {
	Concerns
	backing *files.Track
}

func NewConcernedTrack(track *files.Track) *ConcernedTrack {
	if track == nil {
		return nil
	}
	return &ConcernedTrack{
		Concerns: NewConcerns(),
		backing:  track,
	}
}

func (cT *ConcernedTrack) AddConcern(source ConcernType, concern string) {
	cT.Concerns.AddConcern(source, concern)
}

func (cT *ConcernedTrack) IsConcerned() bool {
	return cT.Concerns.IsConcerned()
}

func (cT *ConcernedTrack) name() string {
	return cT.backing.CommonName()
}

func (cT *ConcernedTrack) ToConsole(o output.Bus) {
	if cT.IsConcerned() {
		o.WriteConsole("    Track %q\n", cT.name())
		cT.Concerns.ToConsole(o, 4)
	}
}

func (cT *ConcernedTrack) Track() *files.Track {
	return cT.backing
}

type ConcernedAlbum struct {
	Concerns
	tracks   []*ConcernedTrack
	backing  *files.Album
	trackMap map[string]*ConcernedTrack
}

func NewConcernedAlbum(album *files.Album) *ConcernedAlbum {
	if album == nil {
		return nil
	}
	cAl := &ConcernedAlbum{
		Concerns: NewConcerns(),
		tracks:   []*ConcernedTrack{},
		backing:  album,
		trackMap: map[string]*ConcernedTrack{},
	}
	for _, track := range album.Tracks() {
		cAl.AddTrack(track)
	}
	return cAl
}

func (cAl *ConcernedAlbum) AddConcern(source ConcernType, concern string) {
	cAl.Concerns.AddConcern(source, concern)
}

func (cAl *ConcernedAlbum) AddTrack(track *files.Track) {
	if cT := NewConcernedTrack(track); cT != nil {
		cAl.tracks = append(cAl.tracks, cT)
		cAl.trackMap[cT.backing.FileName()] = cT
	}
}

func (cAl *ConcernedAlbum) Album() *files.Album {
	return cAl.backing
}

func (cAl *ConcernedAlbum) IsConcerned() bool {
	if cAl.Concerns.IsConcerned() {
		return true
	}
	for _, cT := range cAl.tracks {
		if cT.IsConcerned() {
			return true
		}
	}
	return false
}

func (cAl *ConcernedAlbum) name() string {
	return cAl.backing.Name()
}

func (cAl *ConcernedAlbum) Lookup(track *files.Track) *ConcernedTrack {
	var cT *ConcernedTrack
	if found, ok := cAl.trackMap[track.FileName()]; ok {
		cT = found
	}
	return cT
}

func (cAl *ConcernedAlbum) ToConsole(o output.Bus) {
	if cAl.IsConcerned() {
		o.WriteConsole("  Album %q\n", cAl.name())
		cAl.Concerns.ToConsole(o, 2)
		m := map[string]*ConcernedTrack{}
		names := []string{}
		for _, cT := range cAl.tracks {
			trackName := cT.name()
			m[trackName] = cT
			names = append(names, trackName)
		}
		slices.Sort(names)
		for _, name := range names {
			if cT := m[name]; cT != nil {
				cT.ToConsole(o)
			}
		}
	}
}

func (cAl *ConcernedAlbum) Tracks() []*ConcernedTrack {
	return cAl.tracks
}

type ConcernedArtist struct {
	Concerns
	albums   []*ConcernedAlbum
	backing  *files.Artist
	albumMap map[string]*ConcernedAlbum
}

func NewConcernedArtist(artist *files.Artist) *ConcernedArtist {
	if artist == nil {
		return nil
	}
	cAr := &ConcernedArtist{
		Concerns: NewConcerns(),
		albums:   []*ConcernedAlbum{},
		backing:  artist,
		albumMap: map[string]*ConcernedAlbum{},
	}
	for _, album := range artist.Albums() {
		cAr.AddAlbum(album)
	}
	return cAr
}

func (cAr *ConcernedArtist) AddAlbum(album *files.Album) {
	if cAl := NewConcernedAlbum(album); cAl != nil {
		cAr.albums = append(cAr.albums, cAl)
		cAr.albumMap[cAl.name()] = cAl
	}
}

func (cAr *ConcernedArtist) AddConcern(source ConcernType, concern string) {
	cAr.Concerns.AddConcern(source, concern)
}

func (cAr *ConcernedArtist) Albums() []*ConcernedAlbum {
	return cAr.albums
}

func (cAr *ConcernedArtist) Artist() *files.Artist {
	return cAr.backing
}

func (cAr *ConcernedArtist) IsConcerned() bool {
	if cAr.Concerns.IsConcerned() {
		return true
	}
	for _, cAl := range cAr.albums {
		if cAl.IsConcerned() {
			return true
		}
	}
	return false
}

func (cAr *ConcernedArtist) Lookup(track *files.Track) *ConcernedTrack {
	albumKey := track.AlbumName()
	if cAl, ok := cAr.albumMap[albumKey]; ok {
		return cAl.Lookup(track)
	}
	return nil
}

func (cAr *ConcernedArtist) name() string {
	return cAr.backing.Name()
}

func (cAr *ConcernedArtist) ToConsole(o output.Bus) {
	if cAr.IsConcerned() {
		o.WriteConsole("Artist %q\n", cAr.name())
		cAr.Concerns.ToConsole(o, 0)
		m := map[string]*ConcernedAlbum{}
		names := []string{}
		for _, cT := range cAr.albums {
			albumName := cT.name()
			m[albumName] = cT
			names = append(names, albumName)
		}
		slices.Sort(names)
		for _, name := range names {
			if cAl := m[name]; cAl != nil {
				cAl.ToConsole(o)
			}
		}
	}
}

func PrepareConcernedArtists(artists []*files.Artist) []*ConcernedArtist {
	concernedArtists := []*ConcernedArtist{}
	for _, artist := range artists {
		if cAr := NewConcernedArtist(artist); cAr != nil {
			concernedArtists = append(concernedArtists, cAr)
		}
	}
	return concernedArtists
}
