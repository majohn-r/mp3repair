package files

import (
	"flag"
	"io/fs"
	"io/ioutil"
	"mp3/internal"
	"os"
	"path/filepath"
	"regexp"

	"github.com/sirupsen/logrus"
)

type Search struct {
	topDirectory    string
	targetExtension string
	albumFilter     *regexp.Regexp
	artistFilter    *regexp.Regexp
}

func (s *Search) TopDirectory() string {
	return s.topDirectory
}

func (s *Search) TargetExtension() string {
	return s.targetExtension
}

func (s *Search) LoadUnfilteredData() (artists []*Artist) {
	logrus.WithFields(logrus.Fields{
		internal.LOG_DIRECTORY: s.topDirectory,
		internal.LOG_EXTENSION: s.targetExtension,
	}).Info(internal.LOG_READING_UNFILTERED_FILES)
	artistFiles, err := readDirectory(s.topDirectory)
	if err == nil {
		for _, artistFile := range artistFiles {
			if artistFile.IsDir() {
				artist := newArtist(artistFile, s.topDirectory)
				// artistDir := filepath.Join(s.topDirectory, artistFile.Name())
				albumFiles, err := readDirectory(artist.Path)
				if err == nil {
					for _, albumFile := range albumFiles {
						if !albumFile.IsDir() {
							continue
						}
						album := newAlbumFromFile(albumFile, artist)
						trackFiles, err := readDirectory(album.Path())
						if err == nil {
							for _, trackFile := range trackFiles {
								if trackFile.IsDir() || !trackNameRegex.MatchString(trackFile.Name()) {
									continue
								}
								if simpleName, trackNumber, valid := ParseTrackName(trackFile.Name(), album.Name(), artist.Name, s.targetExtension); valid {
									track := &Track{
										Path:            filepath.Join(album.Path(), trackFile.Name()),
										Name:            simpleName,
										TrackNumber:     trackNumber,
										TaggedTrack:     trackUnknownTagsNotRead,
										ContainingAlbum: album,
									}
									album.AddTrack(track)
								}
							}
						}
						artist.Albums = append(artist.Albums, album)
					}
				}
				artists = append(artists, artist)
			}
		}
	}
	return
}

func (s *Search) logFields() logrus.Fields {
	return logrus.Fields{
		internal.LOG_DIRECTORY:     s.topDirectory,
		internal.LOG_EXTENSION:     s.targetExtension,
		internal.LOG_ALBUM_FILTER:  s.albumFilter,
		internal.LOG_ARTIST_FILTER: s.artistFilter,
	}
}

func (s *Search) FilterArtists(unfilteredArtists []*Artist) (artists []*Artist) {
	logrus.WithFields(s.logFields()).Info(internal.LOG_FILTERING_FILES)
	for _, unfilteredArtist := range unfilteredArtists {
		if s.artistFilter.MatchString(unfilteredArtist.Name) {
			artist := copyArtist(unfilteredArtist)
			for _, album := range unfilteredArtist.Albums {
				if s.albumFilter.MatchString(album.Name()) {
					if album.HasTracks() {
						newAlbum := copyAlbum(album, artist)
						artist.Albums = append(artist.Albums, newAlbum)
					}
				}
			}
			if len(artist.Albums) != 0 {
				artists = append(artists, artist)
			}
		}
	}
	return
}

func (s *Search) LoadData() (artists []*Artist) {
	logrus.WithFields(s.logFields()).Info(internal.LOG_READING_FILTERED_FILES)
	artistFiles, err := readDirectory(s.topDirectory)
	if err == nil {
		for _, artistFile := range artistFiles {
			if !artistFile.IsDir() || !s.artistFilter.MatchString(artistFile.Name()) {
				continue
			}
			artist := newArtist(artistFile, s.topDirectory)
			albumFiles, err := readDirectory(artist.Path)
			if err == nil {
				for _, albumFile := range albumFiles {
					if !albumFile.IsDir() || !s.albumFilter.MatchString(albumFile.Name()) {
						continue
					}
					album := newAlbumFromFile(albumFile, artist)
					trackFiles, err := readDirectory(album.Path())
					if err == nil {
						for _, trackFile := range trackFiles {
							if trackFile.IsDir() || !trackNameRegex.MatchString(trackFile.Name()) {
								continue
							}
							if simpleName, trackNumber, valid := ParseTrackName(trackFile.Name(), album.Name(), artist.Name, s.targetExtension); valid {
								track := &Track{
									Path:            filepath.Join(album.Path(), trackFile.Name()),
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
						artist.Albums = append(artist.Albums, album)
					}
				}
			}
			if len(artist.Albums) != 0 {
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
