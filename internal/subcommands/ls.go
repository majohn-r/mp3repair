package subcommands

import (
	"flag"
	"fmt"
	"io"
	"mp3/internal"
	"mp3/internal/files"
	"os"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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

func newLs(v *viper.Viper, fSet *flag.FlagSet) CommandProcessor {
	return newLsSubCommand(v, fSet)
}

const (
	includeAlbumsFlag       = "album"
	defaultIncludeAlbums    = true
	includeArtistsFlag      = "artist"
	defaultIncludeArtists   = true
	includeTracksFlag       = "track"
	defaultIncludeTracks    = false
	trackSortingFlag        = "sort"
	defaultTrackSorting     = "numeric"
	annotateListingsFlag    = "annotate"
	defaultAnnotateListings = false
)

func newLsSubCommand(v *viper.Viper, fSet *flag.FlagSet) *ls {
	subViper := internal.SafeSubViper(v, "ls")
	return &ls{
		n: fSet.Name(),
		includeAlbums: fSet.Bool(includeAlbumsFlag,
			internal.GetBoolDefault(subViper, includeAlbumsFlag, defaultIncludeAlbums),
			"include album names in listing"),
		includeArtists: fSet.Bool(includeArtistsFlag,
			internal.GetBoolDefault(subViper, includeArtistsFlag, defaultIncludeArtists),
			"include artist names in listing"),
		includeTracks: fSet.Bool(includeTracksFlag,
			internal.GetBoolDefault(subViper, includeTracksFlag, defaultIncludeTracks),
			"include track names in listing"),
		trackSorting: fSet.String(trackSortingFlag,
			internal.GetStringDefault(subViper, trackSortingFlag, defaultTrackSorting),
			"track sorting, 'numeric' in track number order, or 'alpha' in track name order"),
		annotateListings: fSet.Bool(annotateListingsFlag,
			internal.GetBoolDefault(subViper, annotateListingsFlag, defaultAnnotateListings),
			"annotate listings with album and artist data"),
		sf: files.NewSearchFlags(v, fSet),
	}
}

func (l *ls) Exec(w io.Writer, args []string) {
	if s := l.sf.ProcessArgs(os.Stderr, args); s != nil {
		l.runSubcommand(w, s)
	}
}

const (
	logAlbumsFlag  string = "includeAlbums"
	logArtistsFlag string = "includeArtists"
	logSortingFlag string = "trackSorting"
	logTracksFlag  string = "includeTracks"
)

func (l *ls) logFields() logrus.Fields {
	return logrus.Fields{
		internal.LOG_COMMAND_NAME: l.name(),
		logAlbumsFlag:             *l.includeAlbums,
		logArtistsFlag:            *l.includeArtists,
		logTracksFlag:             *l.includeTracks,
		logSortingFlag:            *l.trackSorting,
	}
}

func (l *ls) runSubcommand(w io.Writer, s *files.Search) {
	if !*l.includeArtists && !*l.includeAlbums && !*l.includeTracks {
		fmt.Fprintf(os.Stderr, internal.USER_SPECIFIED_NO_WORK, l.name())
		logrus.WithFields(l.logFields()).Error(internal.LOG_NOTHING_TO_DO)
	} else {
		logrus.WithFields(l.logFields()).Info(internal.LOG_EXECUTING_COMMAND)
		if *l.includeTracks {
			if l.validateTrackSorting() {
				logrus.WithFields(l.logFields()).Info(internal.LOG_PARAMETERS_OVERRIDDEN)
			}
		}
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
			artistsByArtistNames[artist.Name()] = artist
			artistNames = append(artistNames, artist.Name())
		}
		sort.Strings(artistNames)
		for _, artistName := range artistNames {
			fmt.Fprintf(w, "Artist: %s\n", artistName)
			artist := artistsByArtistNames[artistName]
			l.outputAlbums(w, artist.Albums(), "  ")
		}
	case false:
		var albums []*files.Album
		for _, artist := range artists {
			albums = append(albums, artist.Albums()...)
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
				name = album.Name() + " by " + album.RecordingArtistName()
			default:
				name = album.Name()
			}
			albumsByAlbumName[name] = album
			albumNames = append(albumNames, name)
		}
		sort.Strings(albumNames)
		for _, albumName := range albumNames {
			fmt.Fprintf(w, "%sAlbum: %s\n", prefix, albumName)
			album := albumsByAlbumName[albumName]
			l.outputTracks(w, album.Tracks(), prefix+"  ")
		}
	case false:
		var tracks []*files.Track
		for _, album := range albums {
			tracks = append(tracks, album.Tracks()...)
		}
		l.outputTracks(w, tracks, prefix)
	}
}

func (l *ls) validateTrackSorting() bool {
	switch *l.trackSorting {
	case "numeric":
		if !*l.includeAlbums {
			logrus.Warn("numeric track sorting does not make sense without listing albums")
			preferredValue := "alpha"
			l.trackSorting = &preferredValue
			return true
		}
	case "alpha":
	default:
		fmt.Fprintf(os.Stderr, internal.USER_UNRECOGNIZED_VALUE, "-sort", *l.trackSorting)
		logrus.WithFields(logrus.Fields{
			internal.LOG_COMMAND_NAME: l.name(),
			internal.LOG_FLAG:         "-sort",
			internal.LOG_VALUE:        *l.trackSorting,
		}).Warn(internal.LOG_INVALID_FLAG_SETTING)
		var preferredValue string
		switch *l.includeAlbums {
		case true:
			preferredValue = "numeric"
		case false:
			preferredValue = "alpha"
		}
		l.trackSorting = &preferredValue
		return true
	}
	return false
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
			tracksNumeric[track.TrackNumber] = track.Name()
		}
		sort.Ints(trackNumbers)
		for _, trackNumber := range trackNumbers {
			fmt.Fprintf(w, "%s%2d. %s\n", prefix, trackNumber, tracksNumeric[trackNumber])
		}
	case "alpha":
		var trackNames []string
		for _, track := range tracks {
			var components []string
			components = append(components, track.Name())
			if *l.annotateListings {
				if !*l.includeAlbums {
					components = append(components, fmt.Sprintf("on %s", track.ContainingAlbum.Name()))
					if !*l.includeArtists {
						components = append(components, fmt.Sprintf("by %s", track.ContainingAlbum.RecordingArtistName()))
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
