package subcommands

import (
	"flag"
	"fmt"
	"mp3/internal/files"
	"os"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

type ls struct {
	includeAlbums    *bool
	includeArtists   *bool
	includeTracks    *bool
	trackSorting     *string
	annotateListings *bool
	commons          *CommonCommandFlags
}

func (l *ls) name() string {
	return l.commons.name()
}

func newLs(fSet *flag.FlagSet) CommandProcessor {
	processor := &ls{
		includeAlbums:    fSet.Bool("album", true, "include album names in listing"),
		includeArtists:   fSet.Bool("artist", true, "include artist names in listing"),
		includeTracks:    fSet.Bool("track", false, "include track names in listing"),
		trackSorting:     fSet.String("sort", "numeric", "track sorting, 'numeric' in track number order, or 'alpha' in track name order"),
		annotateListings: fSet.Bool("annotate", false, "annotate listings with album and artist data"),
		commons:          newCommonCommandFlags(fSet),
	}
	return processor
}

func (l *ls) Exec(args []string) error {
	if params, err := l.commons.processArgs(args); err == nil {
		l.runSubcommand(params)
		return nil
	} else {
		return err
	}
}

func (l *ls) runSubcommand(params *files.DirectorySearchParams) {
	if !*l.includeArtists && !*l.includeAlbums && !*l.includeTracks {
		fmt.Printf("%s: nothing to do!", l.name())
		os.Exit(0)
	}
	logrus.WithFields(logrus.Fields{
		"subcommandName": l.name(),
		"includeAlbums":  *l.includeAlbums,
		"includeArtists": *l.includeArtists,
		"includeTracks":  *l.includeTracks,
	}).Info("subcommand")
	if *l.includeTracks {
		l.validateTrackSorting()
		logrus.WithFields(logrus.Fields{
			"trackSorting": *l.trackSorting,
		}).Infof("track sorting")
	}
	// artists := files.GetMusic(params)
	artists := files.LoadData(params)
	l.outputArtists(artists)
}

func (l *ls) outputArtists(artists []*files.Artist) {
	switch *l.includeArtists {
	case true:
		artistsByArtistNames := make(map[string]*files.Artist)
		var artistNames []string
		for _, artist := range artists {
			artistsByArtistNames[artist.Name] = artist
			artistNames = append(artistNames, artist.Name)
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
				name = album.Name + " by " + album.RecordingArtist.Name
			default:
				name = album.Name
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
			logrus.Warn("numeric track sorting does not make sense without listing albums")
			preferredValue := "alpha"
			l.trackSorting = &preferredValue
		}
	case "alpha":
	default:
		fmt.Printf("unexpected track sorting '%s'", *l.trackSorting)
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
			if *l.annotateListings {
				if !*l.includeAlbums {
					components = append(components, fmt.Sprintf("on %s", track.ContainingAlbum.Name))
					if !*l.includeArtists {
						components = append(components, fmt.Sprintf("by %s", track.ContainingAlbum.RecordingArtist.Name))
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
