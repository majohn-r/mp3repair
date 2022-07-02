package files

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"mp3/internal"
	"regexp"

	"github.com/sirupsen/logrus"
)

// Search encapsulates the parameters used to find mp3 files and filter them by
// artist and album
type Search struct {
	topDirectory    string
	targetExtension string
	albumFilter     *regexp.Regexp
	artistFilter    *regexp.Regexp
}

// TODO [#77] use OutputBus
func (s *Search) contents(wErr io.Writer) ([]fs.FileInfo, bool) {
	return readDirectory(wErr, s.topDirectory)
}

// LoadUnfilteredData loads artists, albums, and tracks from the specified top
// directory, honoring the specified track extension, but ignoring the album and
// artist filter expressions.
func (s *Search) LoadUnfilteredData(o internal.OutputBus) (artists []*Artist, ok bool) {
	logrus.WithFields(s.LogFields(false)).Info(internal.LI_READING_UNFILTERED_FILES)
	if artistFiles, ok := s.contents(o.ErrorWriter()); ok {
		for _, artistFile := range artistFiles {
			if artistFile.IsDir() {
				artist := newArtistFromFile(artistFile, s.topDirectory)
				if albumFiles, ok := artist.contents(o.ErrorWriter()); ok {
					for _, albumFile := range albumFiles {
						if !albumFile.IsDir() {
							continue
						}
						album := newAlbumFromFile(albumFile, artist)
						if trackFiles, ok := album.contents(o.ErrorWriter()); ok {
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
		o.LogWriter().Warn(internal.LW_NO_ARTIST_DIRECTORIES, s.LogFields(false))
		fmt.Fprintf(o.ErrorWriter(), internal.USER_NO_MUSIC_FILES_FOUND)
	}
	return
}

// LogFields returns an appropriate set of logrus fields
// TODO [#77] return map[string]interface{}
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
	logrus.WithFields(s.LogFields(true)).Info(internal.LI_FILTERING_FILES)
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
		o.LogWriter().Warn(internal.LW_NO_ARTIST_DIRECTORIES, s.LogFields(true))
		fmt.Fprintf(o.ErrorWriter(), internal.USER_NO_MUSIC_FILES_FOUND)
	}
	return
}

// LoadData collects the artists, albums, and mp3 tracks, honoring all the
// search parameters.
// TODO [#77] need OutputBus
func (s *Search) LoadData(o internal.OutputBus, logger internal.Logger, wErr io.Writer) (artists []*Artist, ok bool) {
	logrus.WithFields(s.LogFields(true)).Info(internal.LI_READING_FILTERED_FILES)
	if artistFiles, ok := s.contents(wErr); ok {
		for _, artistFile := range artistFiles {
			if !artistFile.IsDir() || !s.artistFilter.MatchString(artistFile.Name()) {
				continue
			}
			artist := newArtistFromFile(artistFile, s.topDirectory)
			if albumFiles, ok := artist.contents(wErr); ok {
				for _, albumFile := range albumFiles {
					if !albumFile.IsDir() || !s.albumFilter.MatchString(albumFile.Name()) {
						continue
					}
					album := newAlbumFromFile(albumFile, artist)
					if trackFiles, ok := album.contents(wErr); ok {
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
		o.LogWriter().Warn(internal.LW_NO_ARTIST_DIRECTORIES, s.LogFields(true))
		fmt.Fprintf(o.ErrorWriter(), internal.USER_NO_MUSIC_FILES_FOUND)
	}
	return
}

// used for testing only!
func CreateSearchForTesting(topDir string) *Search {
	realFlagSet := flag.NewFlagSet("testing", flag.ContinueOnError)
	s, _ := NewSearchFlags(internal.EmptyConfiguration(), realFlagSet).ProcessArgs(
		internal.NewOutputDeviceForTesting(), []string{"-topDir", topDir})
	return s
}

func CreateFilteredSearchForTesting(topDir string, artistFilter string, albumFilter string) *Search {
	realFlagSet := flag.NewFlagSet("testing", flag.ContinueOnError)
	s, _ := NewSearchFlags(internal.EmptyConfiguration(), realFlagSet).ProcessArgs(
		internal.NewOutputDeviceForTesting(), []string{
			"-topDir", topDir,
			"-artistFilter", artistFilter,
			"-albumFilter", albumFilter,
		})
	return s
}

// TODO [#77] use OutputBus
func readDirectory(wErr io.Writer, dir string) (files []fs.FileInfo, ok bool) {
	var err error
	if files, err = ioutil.ReadDir(dir); err != nil {
		logrus.WithFields(logrus.Fields{
			internal.FK_DIRECTORY: dir,
			internal.FK_ERROR:     err,
		}).Warn(internal.LW_CANNOT_READ_DIRECTORY)
		fmt.Fprintf(wErr, internal.USER_CANNOT_READ_DIRECTORY, dir, err)
		return
	}
	ok = true
	return
}
