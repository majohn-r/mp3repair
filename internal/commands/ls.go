package commands

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

func newLs(c *internal.Configuration, fSet *flag.FlagSet) CommandProcessor {
	return newLsCommand(c, fSet)
}

const (
	defaultAnnotateListings = false
	defaultIncludeAlbums    = true
	defaultIncludeArtists   = true
	defaultIncludeTracks    = false
	defaultTrackSorting     = "numeric"
	annotateListingsFlag    = "annotate"
	fkAnnotateListingsFlag  = "-" + annotateListingsFlag
	fkIncludeAlbumsFlag     = "-" + includeAlbumsFlag
	fkIncludeArtistsFlag    = "-" + includeArtistsFlag
	fkIncludeTracksFlag     = "-" + includeTracksFlag
	fkTrackSortingFlag      = "-" + trackSortingFlag
	includeAlbumsFlag       = "includeAlbums"
	includeArtistsFlag      = "includeArtists"
	includeTracksFlag       = "includeTracks"
	trackSortingFlag        = "sort"
)

func newLsCommand(c *internal.Configuration, fSet *flag.FlagSet) *ls {
	name := fSet.Name()
	configuration := c.SubConfiguration(name)
	return &ls{
		n: name,
		includeAlbums: fSet.Bool(includeAlbumsFlag,
			configuration.BoolDefault(includeAlbumsFlag, defaultIncludeAlbums),
			"include album names in listing"),
		includeArtists: fSet.Bool(includeArtistsFlag,
			configuration.BoolDefault(includeArtistsFlag, defaultIncludeArtists),
			"include artist names in listing"),
		includeTracks: fSet.Bool(includeTracksFlag,
			configuration.BoolDefault(includeTracksFlag, defaultIncludeTracks),
			"include track names in listing"),
		trackSorting: fSet.String(trackSortingFlag,
			configuration.StringDefault(trackSortingFlag, defaultTrackSorting),
			"track sorting, 'numeric' in track number order, or 'alpha' in track name order"),
		annotateListings: fSet.Bool(annotateListingsFlag,
			configuration.BoolDefault(annotateListingsFlag, defaultAnnotateListings),
			"annotate listings with album and artist data"),
		sf: files.NewSearchFlags(c, fSet),
	}
}

func (l *ls) Exec(o internal.OutputBus, args []string) (ok bool) {
	if s, argsOk := l.sf.ProcessArgs(o, args); argsOk {
		// TODO [#77] replace o.OutputWriter() with o
		ok = l.runCommand(o.OutputWriter(), s)
	}
	return
}

func (l *ls) logFields() logrus.Fields {
	return logrus.Fields{
		fkCommandName:          l.name(),
		fkIncludeAlbumsFlag:    *l.includeAlbums,
		fkIncludeArtistsFlag:   *l.includeArtists,
		fkIncludeTracksFlag:    *l.includeTracks,
		fkTrackSortingFlag:     *l.trackSorting,
		fkAnnotateListingsFlag: *l.annotateListings,
	}
}

// TODO [#77] should use 2nd writer for error output
func (l *ls) runCommand(w io.Writer, s *files.Search) (ok bool) {
	if !*l.includeArtists && !*l.includeAlbums && !*l.includeTracks {
		fmt.Fprintf(os.Stderr, internal.USER_SPECIFIED_NO_WORK, l.name())
		logrus.WithFields(l.logFields()).Warn(internal.LW_NOTHING_TO_DO)
		return
	}
	logrus.WithFields(l.logFields()).Info(internal.LI_EXECUTING_COMMAND)
	if *l.includeTracks {
		if l.validateTrackSorting() {
			logrus.WithFields(l.logFields()).Info(internal.LI_PARAMETERS_OVERRIDDEN)
		}
	}
	artists, ok := s.LoadData(os.Stderr)
	if ok {
		l.outputArtists(w, artists)
	}
	return
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
				name = fmt.Sprintf("%q by %q", album.Name(), album.RecordingArtistName())
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

// TODO [#77] writer should be used for error output
func (l *ls) validateTrackSorting() (ok bool) {
	switch *l.trackSorting {
	case "numeric":
		if !*l.includeAlbums {
			fmt.Fprintf(os.Stderr, internal.USER_INVALID_SORTING_APPLIED,
				fkTrackSortingFlag, *l.trackSorting, fkIncludeAlbumsFlag)
			logrus.WithFields(logrus.Fields{
				fkTrackSortingFlag:  *l.trackSorting,
				fkIncludeAlbumsFlag: *l.includeAlbums,
			}).Warn(internal.LW_SORTING_OPTION_UNACCEPTABLE)
			preferredValue := "alpha"
			l.trackSorting = &preferredValue
		}
	case "alpha":
		ok = true
	default:
		fmt.Fprintf(os.Stderr, internal.USER_UNRECOGNIZED_VALUE, fkTrackSortingFlag, *l.trackSorting)
		logrus.WithFields(logrus.Fields{
			fkCommandName:      l.name(),
			fkTrackSortingFlag: *l.trackSorting,
		}).Warn(internal.LW_INVALID_FLAG_SETTING)
		var preferredValue string
		switch *l.includeAlbums {
		case true:
			preferredValue = "numeric"
		case false:
			preferredValue = "alpha"
		}
		l.trackSorting = &preferredValue
	}
	return
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
			trackNumbers = append(trackNumbers, track.Number())
			tracksNumeric[track.Number()] = track.Name()
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
			if *l.annotateListings && !(*l.includeAlbums && *l.includeArtists) {
				if !*l.includeAlbums {
					components = append(components, []string{"on", track.AlbumName()}...)
					if !*l.includeArtists {
						components = append(components, []string{"by", track.RecordingArtist()}...)
					}
				}
			}
			if len(components) > 1 {
				var c2 []string
				c2 = append(c2, fmt.Sprintf("%q", components[0]))
				for k := 1; k < len(components); k += 2 {
					c2 = append(c2, components[k])
					c2 = append(c2, fmt.Sprintf("%q", components[k+1]))
				}
				trackNames = append(trackNames, strings.Join(c2, " "))
			} else {
				trackNames = append(trackNames, components[0])
			}
		}
		sort.Strings(trackNames)
		for _, trackName := range trackNames {
			fmt.Fprintf(w, "%s%s\n", prefix, trackName)
		}
	}
}
