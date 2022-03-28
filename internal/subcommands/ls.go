package subcommands

import (
	"flag"
	"fmt"
	"io"
	"mp3/internal/files"
	"os"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

type ls struct {
	n                string
	includeAlbums    *bool
	includeArtists   *bool
	includeTracks    *bool
	trackSorting     *string
	annotateListings *bool
	sf               *files.SearchFlags
}

func (l *ls) name() string {
	return l.n
}

func newLs(fSet *flag.FlagSet) CommandProcessor {
	return newLsSubCommand(fSet)
}

func newLsSubCommand(fSet *flag.FlagSet) *ls {
	return &ls{
		n:                fSet.Name(),
		includeAlbums:    fSet.Bool("album", true, "include album names in listing"),
		includeArtists:   fSet.Bool("artist", true, "include artist names in listing"),
		includeTracks:    fSet.Bool("track", false, "include track names in listing"),
		trackSorting:     fSet.String("sort", "numeric", "track sorting, 'numeric' in track number order, or 'alpha' in track name order"),
		annotateListings: fSet.Bool("annotate", false, "annotate listings with album and artist data"),
		sf:               files.NewSearchFlags(fSet),
	}
}

func (l *ls) Exec(args []string) {
	if params := l.sf.ProcessArgs(os.Stderr, args); params != nil {
		l.runSubcommand(os.Stdout, params)
	}
}

func (l *ls) runSubcommand(w io.Writer, s *files.Search) {
	if !*l.includeArtists && !*l.includeAlbums && !*l.includeTracks {
		fmt.Fprintf(os.Stderr, "%s: nothing to do!", l.name())
		logrus.WithFields(logrus.Fields{"subcommand name": l.name()}).Error("nothing to do")
	} else {
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
		artists := s.LoadData()
		l.outputArtists(w, artists)
	}
}

func (l *ls) outputArtists(w io.Writer, artists []*files.Artist) {
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
			fmt.Fprintf(w, "Artist: %s\n", artistName)
			artist := artistsByArtistNames[artistName]
			l.outputAlbums(w, artist.Albums, "  ")
		}
	case false:
		var albums []*files.Album
		for _, artist := range artists {
			albums = append(albums, artist.Albums...)
		}
		l.outputAlbums(w, albums, "")
	}
}

func (l *ls) outputAlbums(w io.Writer, albums []*files.Album, prefix string) {
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
			fmt.Fprintf(w, "%sAlbum: %s\n", prefix, albumName)
			album := albumsByAlbumName[albumName]
			l.outputTracks(w, album.Tracks, prefix+"  ")
		}
	case false:
		var tracks []*files.Track
		for _, album := range albums {
			tracks = append(tracks, album.Tracks...)
		}
		l.outputTracks(w, tracks, prefix)
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
		fmt.Fprintf(os.Stderr, "unexpected track sorting '%s'", *l.trackSorting)
		logrus.WithFields(logrus.Fields{"subcommand": l.name(), "setting": "-sort", "value": *l.trackSorting}).Warn("unexpected setting")
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

func (l *ls) outputTracks(w io.Writer, tracks []*files.Track, prefix string) {
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
			fmt.Fprintf(w, "%s%2d. %s\n", prefix, trackNumber, tracksNumeric[trackNumber])
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
			fmt.Fprintf(w, "%s%s\n", prefix, trackName)
		}
	}
}
