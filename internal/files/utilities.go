package files

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bogem/id3v2/v2"
	"github.com/sirupsen/logrus"
)

const (
	rawExtension            string = "mp3"
	DefaultFileExtension    string = "." + rawExtension
	defaultTrackNamePattern string = "^\\d+ .+\\." + rawExtension + "$"
)

type Track struct {
	fullPath        string
	fileName        string
	Name            string
	TrackNumber     int
	ContainingAlbum *Album
}

type Album struct {
	Name            string
	Tracks          []*Track
	RecordingArtist *Artist
}

type Artist struct {
	Name   string
	Albums []*Album
}

type DirectorySearchParams struct {
	topDirectory    string
	targetExtension string
	albumFilter     *regexp.Regexp
	artistFilter    *regexp.Regexp
}

var trackNameRegex *regexp.Regexp = regexp.MustCompile(defaultTrackNamePattern)

func NewDirectorySearchParams(dir, ext, albums, artists string) (params *DirectorySearchParams) {
	albumsFilter, artistsFilter, problemsExist := validateSearchParameters(dir, ext, albums, artists)
	if !problemsExist {
		params = &DirectorySearchParams{
			topDirectory:    dir,
			targetExtension: ext,
			albumFilter:     albumsFilter,
			artistFilter:    artistsFilter,
		}
	}
	return
}

func validateSearchParameters(dir string, ext string, albums string, artists string) (albumsFilter *regexp.Regexp, artistsFilter *regexp.Regexp, problemsExist bool) {
	if !validateTopLevelDirectory(dir) {
		problemsExist = true
	}
	if !validateExtension(ext) {
		problemsExist = true
	}
	if filter, b := validateRegexp(albums, "album"); b {
		problemsExist = true
	} else {
		albumsFilter = filter
	}
	if filter, b := validateRegexp(artists, "artist"); b {
		problemsExist = true
	} else {
		artistsFilter = filter
	}
	return
}

func validateTopLevelDirectory(dir string) bool {
	if file, err := os.Stat(dir); err != nil {
		fmt.Fprintf(os.Stderr, "error checking top directory %q: %v\n", dir, err)
		logrus.WithFields(logrus.Fields{"directory": dir, "error": err}).Error("error checking top directory")
		return false
	} else {
		if file.IsDir() {
			return true
		} else {
			fmt.Fprintf(os.Stderr, "top directory %q is not actually a directory\n", dir)
			logrus.WithFields(logrus.Fields{"directory": dir}).Error("top directory is not a directory")
			return false
		}
	}
}

func validateExtension(ext string) (valid bool) {
	valid = true
	if !strings.HasPrefix(ext, ".") || strings.Contains(strings.TrimPrefix(ext, "."), ".") {
		valid = false
		fmt.Fprintf(os.Stderr, "the extension %q must contain exactly one '.' and '.' must be the first character\n", ext)
		logrus.WithFields(logrus.Fields{"extension": ext}).Error("the file extension must contain exactly one '.' and '.' must be the first character")
	}
	var e error
	trackNameRegex, e = regexp.Compile("^\\d+ .+\\." + strings.TrimPrefix(ext, ".") + "$")
	if e != nil {
		valid = false
		fmt.Fprintf(os.Stderr, "%q is not a valid extension: %v\n", ext, e)
		logrus.WithFields(logrus.Fields{"extension": ext, "error": e}).Error("the extension is not valid")
	}
	return
}

func validateRegexp(pattern, name string) (filter *regexp.Regexp, badRegex bool) {
	if f, err := regexp.Compile(pattern); err != nil {
		fmt.Fprintf(os.Stderr, "%s filter is invalid: %v\n", name, err)
		logrus.WithFields(logrus.Fields{"filterName": name, "error": err}).Error("the filter is invalid")
		badRegex = true
	} else {
		filter = f
	}
	return
}

func LoadData(params *DirectorySearchParams) (artists []*Artist) {
	logrus.WithFields(logrus.Fields{
		"topDirectory":  params.topDirectory,
		"fileExtension": params.targetExtension,
		"albumFilter":   params.albumFilter,
		"artistFilter":  params.artistFilter,
	}).Info("load data from file system")
	// read top directory
	artistFiles, err := readDirectory(params.topDirectory)
	if err != nil {
		return
	}
	for _, artistFile := range artistFiles {
		// we only care about directories, which correspond to artists
		if !artistFile.IsDir() || !params.artistFilter.MatchString(artistFile.Name()) {
			continue
		}
		artist := &Artist{
			Name: artistFile.Name(),
		}
		// look for albums for the current artist
		artistDir := filepath.Join(params.topDirectory, artistFile.Name())
		albumFiles, err := readDirectory(artistDir)
		if err == nil {
			for _, albumFile := range albumFiles {
				// skip over non-directories or directories whose name does not match the album filter
				if !albumFile.IsDir() || !params.albumFilter.MatchString(albumFile.Name()) {
					continue
				}
				album := &Album{
					Name:            albumFile.Name(),
					RecordingArtist: artist,
				}
				// look for tracks in the current album
				albumDir := filepath.Join(artistDir, album.Name)
				trackFiles, err := readDirectory(albumDir)
				if err == nil {
					// process tracks
					for _, trackFile := range trackFiles {
						if trackFile.IsDir() || !trackNameRegex.MatchString(trackFile.Name()) {
							continue
						}
						if simpleName, trackNumber, valid := parseTrackName(trackFile.Name(), album.Name, artist.Name, params.targetExtension); valid {
							track := &Track{
								fullPath:        filepath.Join(albumDir, trackFile.Name()),
								fileName:        trackFile.Name(),
								Name:            simpleName,
								TrackNumber:     trackNumber,
								ContainingAlbum: album,
							}
							album.Tracks = append(album.Tracks, track)
						}
					}
				}
				if len(album.Tracks) != 0 {
					artist.Albums = append(artist.Albums, album)
				}
			}
		}
		if len(artist.Albums) != 0 {
			artists = append(artists, artist)
		}
	}
	return
}

func readDirectory(dir string) (files []fs.FileInfo, err error) {
	files, err = ioutil.ReadDir(dir)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"directory": dir,
			"error":     err,
		}).Error("problem reading directory")
	}
	return
}

func ReadMP3Data(track *Track) {
	tag, err := id3v2.Open(track.fullPath, id3v2.Options{Parse: true})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"filename": track.fullPath,
			"error":    err,
		}).Warn("cannot open mp3 file")
	} else {
		defer tag.Close()

		// Read tags.
		logrus.WithFields(logrus.Fields{
			"fileSystemTrackName":   track.Name,
			"fileSystemTrackNumber": track.TrackNumber,
			"fileSystemArtistName":  track.ContainingAlbum.RecordingArtist.Name,
			"fileSystemAlbumName":   track.ContainingAlbum.Name,
			"metadataTrackName":     tag.Title(),
			"metadataTrackNumber":   tag.GetTextFrame("TRCK").Text,
			"metadataArtistName":    tag.Artist(),
			"metadataAlbumName":     tag.Album(),
		}).Info("track data")
	}
}

func parseTrackName(name string, album string, artist string, ext string) (simpleName string, trackNumber int, valid bool) {
	if !trackNameRegex.MatchString(name) {
		logrus.WithFields(logrus.Fields{
			"trackName":  name,
			"albumName":  album,
			"artistName": artist,
		}).Warn("invalid track name")
		return
	}
	var rawTrackNumber string
	fmt.Sscanf(name, "%s ", &rawTrackNumber)
	simpleName = strings.TrimPrefix(name, rawTrackNumber) // trim off leading track number
	simpleName = strings.TrimPrefix(simpleName, " ")      // trim off leading space
	simpleName = strings.TrimSuffix(simpleName, ext)      // trim off extension
	fmt.Sscanf(rawTrackNumber, "%d", &trackNumber)        // read track number as int
	valid = true
	return
}
