package files

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	DefaultFileExtension string = ".mp3"
	pathSeparator        string = string(os.PathSeparator)
)

func DefaultDirectory() (s string, found bool) {
	var homePath string
	homePath, found = os.LookupEnv("HOMEPATH")
	if !found {
		return
	}
	s = fmt.Sprintf("%s%s%s", homePath, pathSeparator, "Music")
	return
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
		log.Println(err)
	}
	parts := strings.Split(dir, pathSeparator)
	var parentSlice []string
	for k := 0; k < len(parts)-1; k++ {
		parentSlice = append(parentSlice, parts[k])
	}
	f = &File{
		parentPath: strings.Join(parentSlice, pathSeparator),
		name:       parts[len(parts)-1],
		dirFlag:    true,
	}
	for _, file := range files {
		if file.IsDir() {
			subdir := ReadDirectory(dir + string(os.PathSeparator) + file.Name())
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
