package files

import (
	"flag"
	"io/fs"
	"mp3/internal"
	"regexp"
)

// Search encapsulates the parameters used to find mp3 files and filter them by
// artist and album
type Search struct {
	topDirectory    string
	targetExtension string
	albumFilter     *regexp.Regexp
	artistFilter    *regexp.Regexp
}

func (s *Search) contents(o internal.OutputBus) ([]fs.FileInfo, bool) {
	return internal.ReadDirectory(o, s.topDirectory)
}

// LoadUnfilteredData loads artists, albums, and tracks from the specified top
// directory, honoring the specified track extension, but ignoring the album and
// artist filter expressions.
func (s *Search) LoadUnfilteredData(o internal.OutputBus) (artists []*Artist, ok bool) {
	o.LogWriter().Info(internal.LI_READING_UNFILTERED_FILES, s.LogFields(false))
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
		o.LogWriter().Error(internal.LE_NO_ARTIST_DIRECTORIES, s.LogFields(false))
		o.WriteError(internal.USER_NO_MUSIC_FILES_FOUND)
	}
	return
}

// LogFields returns an appropriate set of fields for logging
func (s *Search) LogFields(includeFilters bool) map[string]interface{} {
	m := map[string]interface{}{
		fkTopDirFlag:          s.topDirectory,
		fkTargetExtensionFlag: s.targetExtension,
	}
	if includeFilters {
		m[fkAlbumFilterFlag] = s.albumFilter
		m[fkArtistFilterFlag] = s.artistFilter
	}
	return m
}

// FilterArtists filters out the unwanted artists and albums from the input. The
// result is a new, filtered, copy of the original slice of Artists.
func (s *Search) FilterArtists(o internal.OutputBus, unfilteredArtists []*Artist) (artists []*Artist, ok bool) {
	o.LogWriter().Info(internal.LI_FILTERING_FILES, s.LogFields(true))
	for _, unfilteredArtist := range unfilteredArtists {
		if s.artistFilter.MatchString(unfilteredArtist.Name()) {
			artist := copyArtist(unfilteredArtist)
			for _, album := range unfilteredArtist.Albums() {
				if s.albumFilter.MatchString(album.Name()) {
					if album.HasTracks() {
						newAlbum := copyAlbum(album, artist)
						artist.AddAlbum(newAlbum)
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
		o.LogWriter().Error(internal.LE_NO_ARTIST_DIRECTORIES, s.LogFields(true))
		o.WriteError(internal.USER_NO_MUSIC_FILES_FOUND)
	}
	return
}

// LoadData collects the artists, albums, and mp3 tracks, honoring all the
// search parameters.
func (s *Search) LoadData(o internal.OutputBus) (artists []*Artist, ok bool) {
	o.LogWriter().Info(internal.LI_READING_FILTERED_FILES, s.LogFields(true))
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
		o.LogWriter().Error(internal.LE_NO_ARTIST_DIRECTORIES, s.LogFields(true))
		o.WriteError(internal.USER_NO_MUSIC_FILES_FOUND)
	}
	return
}

// used for testing only!
func CreateSearchForTesting(topDir string) *Search {
	realFlagSet := flag.NewFlagSet("testing", flag.ContinueOnError)
	o := internal.NewOutputDeviceForTesting()
	sf, _ := NewSearchFlags(o, internal.EmptyConfiguration(), realFlagSet)
	s, _ := sf.ProcessArgs(o, []string{"-topDir", topDir})
	return s
}

func CreateFilteredSearchForTesting(topDir string, artistFilter string, albumFilter string) *Search {
	realFlagSet := flag.NewFlagSet("testing", flag.ContinueOnError)
	o := internal.NewOutputDeviceForTesting()
	sf, _ := NewSearchFlags(o, internal.EmptyConfiguration(), realFlagSet)
	s, _ := sf.ProcessArgs(o, []string{
		"-topDir", topDir,
		"-artistFilter", artistFilter,
		"-albumFilter", albumFilter,
	})
	return s
}
