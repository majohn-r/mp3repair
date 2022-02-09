package files

import (
	"fmt"
	"io/ioutil"
	"mp3/internal"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/bogem/id3v2"
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

func GetMusic(dir string, ext string) (artists []*Artist) {
	tree := ReadDirectory(dir)
	for _, file := range tree.contents {
		if file.dirFlag {
			// got an artist!
			artist := &Artist{
				internalRep: file,
			}
			artists = append(artists, artist)
			for _, albumFile := range file.contents {
				if albumFile.dirFlag {
					// got an album!
					album := &Album{
						internalRep:     albumFile,
						RecordingArtist: artist,
					}
					artist.Albums = append(artist.Albums, album)
					for _, trackFile := range albumFile.contents {
						if !trackFile.dirFlag && strings.HasSuffix(trackFile.name, ext) {
							// got a track!
							name, trackNumber := parseTrackName(trackFile.name, ext)
							track := &Track{
								internalRep:     trackFile,
								Name:            name,
								TrackNumber:     trackNumber,
								ContainingAlbum: album,
							}
							// test mp3 read?
							tag, err := id3v2.Open(filepath.Join(trackFile.parentPath, trackFile.name), id3v2.Options{Parse: true})
							if err != nil {
								log.Errorf("Error while opening mp3 file %s %v: ", filepath.Join(trackFile.parentPath, trackFile.name), err)
							} else {
								defer tag.Close()

								// Read tags.
								log.Infof("         name: %s by %s on %s\n", track.Name, track.ContainingAlbum.RecordingArtist.Name(), track.ContainingAlbum.Name())
								log.Infof("Metadata says: %s by %s on %s\n", tag.Title(), tag.Artist(), tag.Album())
							}
							album.Tracks = append(album.Tracks, track)
						}
					}
				}
			}
		}
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
		log.Error(err)
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
