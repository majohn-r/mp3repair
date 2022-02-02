package subcommands

import (
	"flag"
	"fmt"
	"log"
	"mp3/internal/files"
	"sort"
	"strings"
)

type ls struct {
	fs             *flag.FlagSet
	includeAlbums  *bool
	includeArtists *bool
	includeTracks  *bool
	trackSorting   *string
	topDirectory   *string
	fileExtension  *string
}

func (l *ls) Name() string {
	return l.fs.Name()
}

func NewLsCommand() *ls {
	defaultTopDir, _ := files.DefaultDirectory()
	fSet := flag.NewFlagSet("ls", flag.ExitOnError)
	return &ls{
		fs:             fSet,
		includeAlbums:  fSet.Bool("album", true, "include album names in listing"),
		includeArtists: fSet.Bool("artist", true, "include artist names in listing"),
		includeTracks:  fSet.Bool("track", false, "include track names in listing"),
		trackSorting:   fSet.String("sort", "numeric", "track sorting, 'numeric' in track number order, or 'alpha' in track name order"),
		topDirectory:   fSet.String("topDir", defaultTopDir, "top directory in which to look for music files"),
		fileExtension:  fSet.String("ext", files.DefaultFileExtension, "extension for music files"),
	}
}

func (l *ls) Exec(args []string) {
	err := l.fs.Parse(args)
	if err == nil {
		l.runSubcommand()
	} else {
		fmt.Printf("%v\n", err)
	}
}

func (l *ls) runSubcommand() {
	var output []string
	if *l.includeAlbums {
		output = append(output, "include albums")
	}
	if *l.includeArtists {
		output = append(output, "include artists")
	}
	if *l.includeTracks {
		output = append(output, "include tracks")
	}
	log.Printf("%s: %s", l.Name(), strings.Join(output, "; "))
	if !*l.includeArtists && !*l.includeAlbums && !*l.includeTracks {
		log.Printf("nothing to do!")
		return
	}
	log.Printf("search %s for files with extension %s", *l.topDirectory, *l.fileExtension)
	if *l.includeTracks {
		l.validateTrackSorting()
		log.Printf("track order: %s", *l.trackSorting)
	}
	artists := files.GetMusic(*l.topDirectory, *l.fileExtension)
	artistsByArtistNames := make(map[string]*files.Artist)
	var artistNames []string
	for _, artist := range artists {
		artistsByArtistNames[artist.Name()] = artist
		artistNames = append(artistNames, artist.Name())
	}
	sort.Strings(artistNames)
	for _, artistName := range artistNames {
		fmt.Printf("Artist: '%s'\n", artistName)
		artist := artistsByArtistNames[artistName]
		albumsByAlbumName := make(map[string]*files.Album)
		var albumNames []string
		for _, album := range artist.Albums {
			albumsByAlbumName[album.Name()] = album
			albumNames = append(albumNames, album.Name())
		}
		sort.Strings(albumNames)
		for _, albumName := range albumNames {
			fmt.Printf("  Album: '%s'\n", albumName)
			album := albumsByAlbumName[albumName]
			l.outputTracks(album.Tracks, "    ")
		}
	}
}

func (l *ls) validateTrackSorting() {
	switch *l.trackSorting {
	case "numeric":
		if !*l.includeAlbums {
			log.Printf("numeric track sorting does not make sense without listing albums")
			preferredValue := "alpha"
			l.trackSorting = &preferredValue
		}
	case "alpha":
	default:
		log.Printf("unexpected track sorting '%s'", *l.trackSorting)
		preferredValue := "numeric"
		l.trackSorting = &preferredValue
	}
}

func (l *ls) outputTracks(tracks []*files.Track, prefix string) {
	switch *l.trackSorting {
	case "numeric":
		tracksNumeric := make(map[int]string)
		var trackNumbers []int
		for _, track := range tracks {
			trackNumbers = append(trackNumbers, track.TrackNumber)
			tracksNumeric[track.TrackNumber] = track.Name
		}
		sort.Ints(trackNumbers)
		for _, trackNumber := range trackNumbers {
			fmt.Printf("%s%2d. %s\n", prefix, trackNumber, tracksNumeric[trackNumber])
		}
	case "alpha":
		var trackNames []string
		for _, track := range tracks {
			trackNames = append(trackNames, track.Name)
		}
		sort.Strings(trackNames)
		for _, trackName := range trackNames {
			fmt.Printf("%s%s\n", prefix, trackName)
		}
	}
}
