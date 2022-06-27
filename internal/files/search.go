package files

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"mp3/internal"
	"os"
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

func (s *Search) contents(wErr io.Writer) ([]fs.FileInfo, bool) {
	return readDirectory(wErr, s.topDirectory)
}

// LoadUnfilteredData loads artists, albums, and tracks from the specified top
// directory, honoring the specified track extension, but ignoring the album and
// artist filter expressions.
// TODO [#81] should return bool, false on returning no artists
func (s *Search) LoadUnfilteredData(wErr io.Writer) (artists []*Artist) {
	logrus.WithFields(s.LogFields(false)).Info(internal.LI_READING_UNFILTERED_FILES)
	if artistFiles, ok := s.contents(wErr); ok {
		for _, artistFile := range artistFiles {
			if artistFile.IsDir() {
				artist := newArtistFromFile(artistFile, s.topDirectory)
				if albumFiles, ok := artist.contents(wErr); ok {
					for _, albumFile := range albumFiles {
						if !albumFile.IsDir() {
							continue
						}
						album := newAlbumFromFile(albumFile, artist)
						if trackFiles, ok := album.contents(wErr); ok {
							for _, trackFile := range trackFiles {
								if trackFile.IsDir() || !trackNameRegex.MatchString(trackFile.Name()) {
									continue
								}
								if simpleName, trackNumber, valid := parseTrackName(trackFile.Name(), album, s.targetExtension); valid {
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
	return
}

// LogFields returns an appropriate set of logrus fields
func (s *Search) LogFields(includeFilters bool) logrus.Fields {
	if includeFilters {
		return logrus.Fields{
			fkTopDirFlag:          s.topDirectory,
			fkTargetExtensionFlag: s.targetExtension,
			fkAlbumFilterFlag:     s.albumFilter,
			fkArtistFilterFlag:    s.artistFilter,
		}
	} else {
		return logrus.Fields{
			fkTopDirFlag:          s.topDirectory,
			fkTargetExtensionFlag: s.targetExtension,
		}
	}
}

// FilterArtists filters out the unwanted artists and albums from the input. The
// result is a new, filtered, copy of the original slice of Artists.
// TODO [#81] should return bool, false on returning no artists
func (s *Search) FilterArtists(unfilteredArtists []*Artist) (artists []*Artist) {
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
	return
}

// LoadData collects the artists, albums, and mp3 tracks, honoring all the
// search parameters.
// TODO [#81] should return bool, false on returning no artists
func (s *Search) LoadData(wErr io.Writer) (artists []*Artist) {
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
							if simpleName, trackNumber, valid := parseTrackName(trackFile.Name(), album, s.targetExtension); valid {
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
	return
}

// used for testing only!
func CreateSearchForTesting(topDir string) *Search {
	realFlagSet := flag.NewFlagSet("testing", flag.ContinueOnError)
	s, _ := NewSearchFlags(internal.EmptyConfiguration(), realFlagSet).ProcessArgs(os.Stdout, []string{
		"-topDir", topDir,
	})
	return s
}

func CreateFilteredSearchForTesting(topDir string, artistFilter string, albumFilter string) *Search {
	realFlagSet := flag.NewFlagSet("testing", flag.ContinueOnError)
	s, _ := NewSearchFlags(internal.EmptyConfiguration(), realFlagSet).ProcessArgs(os.Stdout, []string{
		"-topDir", topDir,
		"-artistFilter", artistFilter,
		"-albumFilter", albumFilter,
	})
	return s
}

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
