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
	fs               *flag.FlagSet
	includeAlbums    *bool
	includeArtists   *bool
	includeTracks    *bool
	trackSorting     *string
	topDirectory     *string
	fileExtension    *string
	albumRegex       *string
	artistRegex      *string
	annotateListings *bool
}

func (l *ls) Name() string {
	return l.fs.Name()
}

func NewLsCommand() *ls {
	defaultTopDir, _ := files.DefaultDirectory()
	fSet := flag.NewFlagSet("ls", flag.ExitOnError)
	return &ls{
		fs:               fSet,
		includeAlbums:    fSet.Bool("album", true, "include album names in listing"),
		includeArtists:   fSet.Bool("artist", true, "include artist names in listing"),
		includeTracks:    fSet.Bool("track", false, "include track names in listing"),
		trackSorting:     fSet.String("sort", "numeric", "track sorting, 'numeric' in track number order, or 'alpha' in track name order"),
		topDirectory:     fSet.String("topDir", defaultTopDir, "top directory in which to look for music files"),
		fileExtension:    fSet.String("ext", files.DefaultFileExtension, "extension for music files"),
		annotateListings: fSet.Bool("annotate", false, "annotate listings with album and artist data"),
		albumRegex:       fSet.String("albums", ".*", "regular expression of albums to repair"),
		artistRegex:      fSet.String("artists", ".*", "regular epxression of artists to repair"),
	}
}

func (l *ls) Exec(args []string) {
	err := l.fs.Parse(args)
	switch err {
	case nil:
		l.runSubcommand()
	default:
		fmt.Printf("%v\n", err)
	}
}

func (l *ls) runSubcommand() {
	if !*l.includeArtists && !*l.includeAlbums && !*l.includeTracks {
		log.Printf("%s: nothing to do!", l.Name())
		return
	}
	var output []string
	if *l.includeAlbums {
		output = append(output, " include albums")
	}
	if *l.includeArtists {
		output = append(output, " include artists")
	}
	if *l.includeTracks {
		output = append(output, " include tracks")
	}
	log.Printf("%s:%s", l.Name(), strings.Join(output, ";"))
	log.Printf("search %s for files with extension %s; for artists '%s' and albums '%s'", *l.topDirectory, *l.fileExtension, *l.artistRegex, *l.albumRegex)
	if *l.includeTracks {
		l.validateTrackSorting()
		log.Printf("track order: %s", *l.trackSorting)
	}
	artists := files.GetMusic(*l.topDirectory, *l.fileExtension)
	l.outputArtists(artists)
}

func (l *ls) outputArtists(artists []*files.Artist) {
	switch *l.includeArtists {
	case true:
		artistsByArtistNames := make(map[string]*files.Artist)
		var artistNames []string
		for _, artist := range artists {
			artistsByArtistNames[artist.Name()] = artist
			artistNames = append(artistNames, artist.Name())
		}
		sort.Strings(artistNames)
		for _, artistName := range artistNames {
			fmt.Printf("Artist: %s\n", artistName)
			artist := artistsByArtistNames[artistName]
			l.outputAlbums(artist.Albums, "  ")
		}
	case false:
		var albums []*files.Album
		for _, artist := range artists {
			albums = append(albums, artist.Albums...)
		}
		l.outputAlbums(albums, "")
	}
}

func (l *ls) outputAlbums(albums []*files.Album, prefix string) {
	switch *l.includeAlbums {
	case true:
		albumsByAlbumName := make(map[string]*files.Album)
		var albumNames []string
		for _, album := range albums {
			var name string
			switch {
			case !*l.includeArtists && *l.annotateListings:
				name = album.Name() + " by " + album.RecordingArtist.Name()
			default:
				name = album.Name()
			}
			albumsByAlbumName[name] = album
			albumNames = append(albumNames, name)
		}
		sort.Strings(albumNames)
		for _, albumName := range albumNames {
			fmt.Printf("%sAlbum: %s\n", prefix, albumName)
			album := albumsByAlbumName[albumName]
			l.outputTracks(album.Tracks, prefix+"  ")
		}
	case false:
		var tracks []*files.Track
		for _, album := range albums {
			tracks = append(tracks, album.Tracks...)
		}
		l.outputTracks(tracks, prefix)
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
		var preferredValue string
		switch *l.includeAlbums {
		case true:
			preferredValue = "numeric"
		case false:
			preferredValue = "alpha"
		}
		l.trackSorting = &preferredValue
	}
}

func (l *ls) outputTracks(tracks []*files.Track, prefix string) {
	if !*l.includeTracks {
		return
	}
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
			var components []string
			components = append(components, track.Name)
			if (*l.annotateListings) {
				if !*l.includeAlbums {
					components = append(components, fmt.Sprintf("on %s", track.ContainingAlbum.Name()))
					if !*l.includeArtists {
						components = append(components, fmt.Sprintf("by %s",track.ContainingAlbum.RecordingArtist.Name()))
					}
				}
			}
			trackNames = append(trackNames, strings.Join(components, " "))
		}
		sort.Strings(trackNames)
		for _, trackName := range trackNames {
			fmt.Printf("%s%s\n", prefix, trackName)
		}
	}
}
