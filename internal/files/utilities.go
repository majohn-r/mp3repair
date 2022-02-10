package files

import (
	"fmt"
	"io/ioutil"
	"mp3/internal"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/bogem/id3v2/v2"
)

const (
	DefaultFileExtension string = ".mp3"
)

func DefaultDirectory() string {
	return filepath.Join(internal.HomePath, "Music")
}

type File struct {
	parentPath string
	name       string
	dirFlag    bool
	contents   []*File
}

type Track struct {
	internalRep     *File
	Name            string
	TrackNumber     int
	ContainingAlbum *Album
}

func (t *Track) FullName() string {
	return t.internalRep.name
}

type Album struct {
	internalRep     *File
	Tracks          []*Track
	RecordingArtist *Artist
}

func (al *Album) Name() string {
	return al.internalRep.name
}

type Artist struct {
	internalRep *File
	Albums      []*Album
}

func (ar *Artist) Name() string {
	return ar.internalRep.name
}

type DirectorySearchParams struct {
	topDirectory    string
	targetExtension string
	albumFilter     *regexp.Regexp
	artistFilter    *regexp.Regexp
}

func NewDirectorySearchParams(dir, ext, albums, artists string) *DirectorySearchParams {
	var badRegex bool
	var albumsFilter *regexp.Regexp
	var artistsFilter *regexp.Regexp
	if filter, b := validateRegexp(albums, "album"); b {
		badRegex = true
	} else {
		albumsFilter = filter
	}
	if filter, b := validateRegexp(artists, "artist"); b {
		badRegex = true
	} else {
		artistsFilter = filter
	}
	if badRegex {
		os.Exit(1)
	}
	return &DirectorySearchParams{
		topDirectory:    dir,
		targetExtension: ext,
		albumFilter:     albumsFilter,
		artistFilter:    artistsFilter,
	}
}

func validateRegexp(pattern, name string) (filter *regexp.Regexp, badRegex bool) {
	if f, err := regexp.Compile(pattern); err != nil {
		fmt.Printf("%s filter is invalid: %v\n", name, err)
		badRegex = true
	} else {
		filter = f
	}
	return
}

func GetMusic(params *DirectorySearchParams) (artists []*Artist) {
	log.WithFields(log.Fields{
		"topDirectory":  params.topDirectory,
		"fileExtension": params.targetExtension,
		"albumFilter":   params.albumFilter,
		"artistFilter":  params.artistFilter,
	}).Info("read data from file system")
	tree := ReadDirectory(params.topDirectory)
	var filteredAlbums bool
	for _, file := range tree.contents {
		if file.dirFlag {
			// got an artist!
			if !params.artistFilter.MatchString(file.name) {
				continue
			}
			artist := &Artist{
				internalRep: file,
			}
			artists = append(artists, artist)
			for _, albumFile := range file.contents {
				if albumFile.dirFlag {
					// got an album!
					if !params.albumFilter.MatchString(albumFile.name) {
						filteredAlbums = true
						continue
					}
					album := &Album{
						internalRep:     albumFile,
						RecordingArtist: artist,
					}
					artist.Albums = append(artist.Albums, album)
					for _, trackFile := range albumFile.contents {
						if !trackFile.dirFlag && strings.HasSuffix(trackFile.name, params.targetExtension) {
							// got a track!
							name, trackNumber := parseTrackName(trackFile.name, params.targetExtension)
							track := &Track{
								internalRep:     trackFile,
								Name:            name,
								TrackNumber:     trackNumber,
								ContainingAlbum: album,
							}
							// test mp3 read?
							tag, err := id3v2.Open(filepath.Join(trackFile.parentPath, trackFile.name), id3v2.Options{Parse: true})
							if err != nil {
								log.WithFields(log.Fields{
									"filename": filepath.Join(trackFile.parentPath, trackFile.name),
									"error":    err,
								}).Warn("cannot open mp3 file")
							} else {
								defer tag.Close()

								// Read tags.
								log.WithFields(log.Fields{
									"fileSystemTrackName":   track.Name,
									"fileSystemTrackNumber": track.TrackNumber,
									"fileSystemArtistName":  track.ContainingAlbum.RecordingArtist.Name(),
									"fileSystemAlbumName":   track.ContainingAlbum.Name(),
									"metadataTrackName":     tag.Title(),
									"metadataTrackNumber":   tag.GetTextFrame("TRCK").Text,
									"metadataArtistName":    tag.Artist(),
									"metadataAlbumName":     tag.Album(),
								}).Info("track date")
							}
							album.Tracks = append(album.Tracks, track)
						}
					}
				}
			}
		}
	}
	// purge artists with all albums filtered out
	if filteredAlbums {
		var filteredArtists []*Artist
		for _, artist := range artists {
			if len(artist.Albums) > 0 {
				filteredArtists = append(filteredArtists, artist)
			}
		}
		artists = filteredArtists
	}
	return artists
}

func parseTrackName(name string, ext string) (simple string, track int) {
	var rawTrackName string
	fmt.Sscanf(name, "%s ", &rawTrackName)
	simple = strings.TrimPrefix(name, rawTrackName)
	simple = strings.TrimPrefix(simple, " ")
	simple = strings.TrimSuffix(simple, ext)
	fmt.Sscanf(rawTrackName, "%d", &track)
	return
}

func ReadDirectory(dir string) (f *File) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("reading directory")
	}
	parentDirName, dirName := filepath.Split(dir)
	f = &File{
		parentPath: parentDirName,
		name:       dirName,
		dirFlag:    true,
	}
	for _, file := range files {
		if file.IsDir() {
			subdir := ReadDirectory(filepath.Join(dir, file.Name()))
			f.contents = append(f.contents, subdir)
		} else {
			plainFile := &File{
				parentPath: dir,
				name:       file.Name(),
				dirFlag:    false,
			}
			f.contents = append(f.contents, plainFile)
		}
	}
	return
}
