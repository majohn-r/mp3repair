package files

import (
	"flag"
	"io/fs"
	"mp3/internal"
	"regexp"

	"github.com/majohn-r/output"
)

// Search encapsulates the parameters used to find mp3 files and filter them by
// artist and album
type Search struct {
	topDirectory    string
	targetExtension string
	albumFilter     *regexp.Regexp
	artistFilter    *regexp.Regexp
}

func (s *Search) contents(o output.Bus) ([]fs.DirEntry, bool) {
	return internal.ReadDirectory(o, s.topDirectory)
}

// LoadUnfiltered loads artists, albums, and tracks from the specified top
// directory, honoring the specified track extension, but ignoring the album and
// artist filter expressions.
func (s *Search) LoadUnfiltered(o output.Bus) (artists []*Artist, ok bool) {
	o.Log(output.Info, "reading unfiltered music files", s.logFields(false))
	if artistFiles, ok := s.contents(o); ok {
		for _, artistFile := range artistFiles {
			if artistFile.IsDir() {
				artist := newArtistFromFile(artistFile, s.topDirectory)
				if albumFiles, ok := artist.contents(o); ok {
					for _, albumFile := range albumFiles {
						if !albumFile.IsDir() {
							continue
						}
						album := newAlbumFromFile(albumFile, artist)
						if trackFiles, ok := album.contents(o); ok {
							for _, trackFile := range trackFiles {
								if trackFile.IsDir() || !trackNameRegex.MatchString(trackFile.Name()) {
									continue
								}
								if simpleName, trackNumber, valid := parseTrackName(o, trackFile.Name(), album, s.targetExtension); valid {
									album.AddTrack(newTrackFromFile(album, trackFile, simpleName, trackNumber))
								}
							}
						}
						artist.AddAlbum(album)
					}
				}
				artists = append(artists, artist)
			}
		}
	}
	ok = len(artists) != 0
	if !ok {
		s.reportNoArtistDirectories(o, false)
	}
	return
}

func (s *Search) reportNoArtistDirectories(o output.Bus, filtered bool) {
	o.WriteCanonicalError("No music files could be found using the specified parameters.")
	o.Log(output.Error, "cannot find any artist directories", s.logFields(filtered))
}

func (s *Search) logFields(filtered bool) map[string]any {
	m := map[string]any{
		"-" + topDirectoryFlag:  s.topDirectory,
		"-" + fileExtensionFlag: s.targetExtension,
	}
	if filtered {
		m["-"+albumRegexFlag] = s.albumFilter
		m["-"+artistRegexFlag] = s.artistFilter
	}
	return m
}

// FilterArtists filters out the unwanted artists and albums from the input. The
// result is a new, filtered, copy of the original slice of Artists.
func (s *Search) FilterArtists(o output.Bus, unfiltered []*Artist) (filtered []*Artist, ok bool) {
	o.Log(output.Info, "filtering music files", s.logFields(true))
	for _, originalArtist := range unfiltered {
		if s.artistFilter.MatchString(originalArtist.Name()) {
			artist := copyArtist(originalArtist)
			for _, originalAlbum := range originalArtist.Albums() {
				if s.albumFilter.MatchString(originalAlbum.Name()) && originalAlbum.HasTracks() {
					artist.AddAlbum(copyAlbum(originalAlbum, artist))
				}
			}
			if artist.HasAlbums() {
				filtered = append(filtered, artist)
			}
		}
	}
	ok = len(filtered) != 0
	if !ok {
		s.reportNoArtistDirectories(o, true)
	}
	return
}

// Load collects the artists, albums, and mp3 tracks, honoring all the search
// parameters.
func (s *Search) Load(o output.Bus) (artists []*Artist, ok bool) {
	o.Log(output.Info, "reading filtered music files", s.logFields(true))
	if artistFiles, ok := s.contents(o); ok {
		for _, artistFile := range artistFiles {
			if !artistFile.IsDir() || !s.artistFilter.MatchString(artistFile.Name()) {
				continue
			}
			artist := newArtistFromFile(artistFile, s.topDirectory)
			if albumFiles, ok := artist.contents(o); ok {
				for _, albumFile := range albumFiles {
					if !albumFile.IsDir() || !s.albumFilter.MatchString(albumFile.Name()) {
						continue
					}
					album := newAlbumFromFile(albumFile, artist)
					if trackFiles, ok := album.contents(o); ok {
						for _, trackFile := range trackFiles {
							if trackFile.IsDir() || !trackNameRegex.MatchString(trackFile.Name()) {
								continue
							}
							if simpleName, trackNumber, valid := parseTrackName(o, trackFile.Name(), album, s.targetExtension); valid {
								album.AddTrack(newTrackFromFile(album, trackFile, simpleName, trackNumber))
							}
						}
					}
					if album.HasTracks() {
						artist.AddAlbum(album)
					}
				}
			}
			if artist.HasAlbums() {
				artists = append(artists, artist)
			}
		}
	}
	ok = len(artists) != 0
	if !ok {
		s.reportNoArtistDirectories(o, true)
	}
	return
}

// CreateSearchForTesting creates a minimal Search instance; used for testing
// only!
func CreateSearchForTesting(topDir string) *Search {
	o := output.NewNilBus()
	sf, _ := NewSearchFlags(o, internal.EmptyConfiguration(), flag.NewFlagSet("testing", flag.ContinueOnError))
	s, _ := sf.ProcessArgs(o, []string{"-topDir", topDir})
	return s
}

// CreateFilteredSearchForTesting creates a Search instance configured for
// specified search parameters
func CreateFilteredSearchForTesting(topDir, artistFilter, albumFilter string) *Search {
	o := output.NewNilBus()
	sf, _ := NewSearchFlags(o, internal.EmptyConfiguration(), flag.NewFlagSet("testing", flag.ContinueOnError))
	s, _ := sf.ProcessArgs(o, []string{"-topDir", topDir, "-artistFilter", artistFilter, "-albumFilter", albumFilter})
	return s
}
