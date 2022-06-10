package files

import (
	"flag"
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

func (s *Search) contents() ([]fs.FileInfo, error) {
	return readDirectory(s.topDirectory)
}

// LoadUnfilteredData loads artists, albums, and tracks from the specified top
// directory, honoring the specified track extension, but ignoring the album and
// artist filter expressions.
func (s *Search) LoadUnfilteredData() (artists []*Artist) {
	logrus.WithFields(s.LogFields(false)).Info(internal.LOG_READING_UNFILTERED_FILES)
	artistFiles, err := s.contents()
	if err == nil {
		for _, artistFile := range artistFiles {
			if artistFile.IsDir() {
				artist := newArtistFromFile(artistFile, s.topDirectory)
				// artistDir := filepath.Join(s.topDirectory, artistFile.Name())
				albumFiles, err := artist.contents()
				if err == nil {
					for _, albumFile := range albumFiles {
						if !albumFile.IsDir() {
							continue
						}
						album := newAlbumFromFile(albumFile, artist)
						trackFiles, err := album.contents()
						if err == nil {
							for _, trackFile := range trackFiles {
								if trackFile.IsDir() || !trackNameRegex.MatchString(trackFile.Name()) {
									continue
								}
								if simpleName, trackNumber, valid := ParseTrackName(trackFile.Name(), album.Name(), artist.Name(), s.targetExtension); valid {
									track := &Track{
										Path:            album.subDirectory(trackFile.Name()),
										Name:            simpleName,
										TrackNumber:     trackNumber,
										TaggedTrack:     trackUnknownTagsNotRead,
										ContainingAlbum: album,
									}
									album.AddTrack(track)
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
			internal.LOG_DIRECTORY:     s.topDirectory,
			internal.LOG_EXTENSION:     s.targetExtension,
			internal.LOG_ALBUM_FILTER:  s.albumFilter,
			internal.LOG_ARTIST_FILTER: s.artistFilter,
		}
	} else {
		return logrus.Fields{
			internal.LOG_DIRECTORY: s.topDirectory,
			internal.LOG_EXTENSION: s.targetExtension,
		}
	}
}

// FilterArtists filters out the unwanted artists and albums from the input. The
// result is a new, filtered, copy of the original slice of Artists.
func (s *Search) FilterArtists(unfilteredArtists []*Artist) (artists []*Artist) {
	logrus.WithFields(s.LogFields(true)).Info(internal.LOG_FILTERING_FILES)
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
func (s *Search) LoadData() (artists []*Artist) {
	logrus.WithFields(s.LogFields(true)).Info(internal.LOG_READING_FILTERED_FILES)
	artistFiles, err := s.contents()
	if err == nil {
		for _, artistFile := range artistFiles {
			if !artistFile.IsDir() || !s.artistFilter.MatchString(artistFile.Name()) {
				continue
			}
			artist := newArtistFromFile(artistFile, s.topDirectory)
			albumFiles, err := artist.contents()
			if err == nil {
				for _, albumFile := range albumFiles {
					if !albumFile.IsDir() || !s.albumFilter.MatchString(albumFile.Name()) {
						continue
					}
					album := newAlbumFromFile(albumFile, artist)
					trackFiles, err := album.contents()
					if err == nil {
						for _, trackFile := range trackFiles {
							if trackFile.IsDir() || !trackNameRegex.MatchString(trackFile.Name()) {
								continue
							}
							if simpleName, trackNumber, valid := ParseTrackName(trackFile.Name(), album.Name(), artist.Name(), s.targetExtension); valid {
								track := &Track{
									Path:            album.subDirectory(trackFile.Name()),
									Name:            simpleName,
									TrackNumber:     trackNumber,
									TaggedTrack:     trackUnknownTagsNotRead,
									ContainingAlbum: album,
								}
								album.AddTrack(track)
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
	return NewSearchFlags(nil, realFlagSet).ProcessArgs(os.Stdout, []string{"-topDir", topDir})
}

func CreateFilteredSearchForTesting(topDir string, artistFilter string, albumFilter string) *Search {
	realFlagSet := flag.NewFlagSet("testing", flag.ContinueOnError)
	return NewSearchFlags(nil, realFlagSet).ProcessArgs(os.Stdout, []string{
		"-topDir", topDir,
		"-artists", artistFilter,
		"-albums", albumFilter,
	})
}

func readDirectory(dir string) (files []fs.FileInfo, err error) {
	files, err = ioutil.ReadDir(dir)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			internal.LOG_DIRECTORY: dir,
			internal.LOG_ERROR:     err,
		}).Error(internal.LOG_CANNOT_READ_DIRECTORY)
	}
	return
}
