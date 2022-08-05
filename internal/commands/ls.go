package commands

import (
	"flag"
	"fmt"
	"mp3/internal"
	"mp3/internal/files"
	"sort"
	"strings"
)

type ls struct {
	n                string
	includeAlbums    *bool
	includeArtists   *bool
	includeTracks    *bool
	trackSorting     *string
	annotateListings *bool
	diagnostics      *bool
	sf               *files.SearchFlags
}

func (l *ls) name() string {
	return l.n
}

func newLs(o internal.OutputBus, c *internal.Configuration, fSet *flag.FlagSet) (CommandProcessor, bool) {
	return newLsCommand(o, c, fSet)
}

const (
	defaultAnnotateListings  = false
	defaultDiagnosticListing = false
	defaultIncludeAlbums     = true
	defaultIncludeArtists    = true
	defaultIncludeTracks     = false
	alphabeticSorting        = "alpha"
	numericSorting           = "numeric"
	defaultTrackSorting      = numericSorting
	annotateListingsFlag     = "annotate"
	diagnosticListingFlag    = "diagnostic"
	includeAlbumsFlag        = "includeAlbums"
	includeArtistsFlag       = "includeArtists"
	includeTracksFlag        = "includeTracks"
	trackSortingFlag         = "sort"
	fkAnnotateListingsFlag   = "-" + annotateListingsFlag
	fkDiagnosticListingFlag  = "-" + diagnosticListingFlag
	fkIncludeAlbumsFlag      = "-" + includeAlbumsFlag
	fkIncludeArtistsFlag     = "-" + includeArtistsFlag
	fkIncludeTracksFlag      = "-" + includeTracksFlag
	fkTrackSortingFlag       = "-" + trackSortingFlag
	fkTrack                  = "track"
)

type lsDefaults struct {
	includeAlbums  bool
	includeArtists bool
	includeTracks  bool
	annotateTracks bool
	diagnostics    bool
	sorting        string
}

func newLsCommand(o internal.OutputBus, c *internal.Configuration, fSet *flag.FlagSet) (*ls, bool) {
	name := fSet.Name()
	defaults, defaultsOk := evaluateLsDefaults(o, c.SubConfiguration(name), name)
	sFlags, sFlagsOk := files.NewSearchFlags(o, c, fSet)
	if defaultsOk && sFlagsOk {
		return &ls{
			n:                name,
			includeAlbums:    fSet.Bool(includeAlbumsFlag, defaults.includeAlbums, "include album names in listing"),
			includeArtists:   fSet.Bool(includeArtistsFlag, defaults.includeArtists, "include artist names in listing"),
			includeTracks:    fSet.Bool(includeTracksFlag, defaults.includeTracks, "include track names in listing"),
			trackSorting:     fSet.String(trackSortingFlag, defaults.sorting, "track sorting, 'numeric' in track number order, or 'alpha' in track name order"),
			annotateListings: fSet.Bool(annotateListingsFlag, defaults.annotateTracks, "annotate listings with album and artist data"),
			diagnostics:      fSet.Bool(diagnosticListingFlag, defaults.diagnostics, "include diagnostic information with tracks"),
			sf:               sFlags,
		}, true
	}
	return nil, false
}

func evaluateLsDefaults(o internal.OutputBus, c *internal.Configuration, name string) (defaults lsDefaults, ok bool) {
	ok = true
	defaults = lsDefaults{}
	var err error
	defaults.includeAlbums, err = c.BoolDefault(includeAlbumsFlag, defaultIncludeAlbums)
	if err != nil {
		reportBadDefault(o, name, err)
		ok = false
	}
	defaults.includeArtists, err = c.BoolDefault(includeArtistsFlag, defaultIncludeArtists)
	if err != nil {
		reportBadDefault(o, name, err)
		ok = false
	}
	defaults.includeTracks, err = c.BoolDefault(includeTracksFlag, defaultIncludeTracks)
	if err != nil {
		reportBadDefault(o, name, err)
		ok = false
	}
	defaults.annotateTracks, err = c.BoolDefault(annotateListingsFlag, defaultAnnotateListings)
	if err != nil {
		reportBadDefault(o, name, err)
		ok = false
	}
	defaults.diagnostics, err = c.BoolDefault(diagnosticListingFlag, defaultDiagnosticListing)
	if err != nil {
		reportBadDefault(o, name, err)
		ok = false
	}
	defaults.sorting, err = c.StringDefault(trackSortingFlag, defaultTrackSorting)
	if err != nil {
		reportBadDefault(o, name, err)
		ok = false
	}
	return
}

func (l *ls) Exec(o internal.OutputBus, args []string) (ok bool) {
	if s, argsOk := l.sf.ProcessArgs(o, args); argsOk {
		ok = l.runCommand(o, s)
	}
	return
}

func (l *ls) logFields() map[string]interface{} {
	return map[string]interface{}{
		fkCommandName:           l.name(),
		fkIncludeAlbumsFlag:     *l.includeAlbums,
		fkIncludeArtistsFlag:    *l.includeArtists,
		fkIncludeTracksFlag:     *l.includeTracks,
		fkTrackSortingFlag:      *l.trackSorting,
		fkAnnotateListingsFlag:  *l.annotateListings,
		fkDiagnosticListingFlag: *l.diagnostics,
	}
}

func (l *ls) runCommand(o internal.OutputBus, s *files.Search) (ok bool) {
	if !*l.includeArtists && !*l.includeAlbums && !*l.includeTracks {
		o.WriteError(internal.USER_SPECIFIED_NO_WORK, l.name())
		o.LogWriter().Error(internal.LE_NOTHING_TO_DO, l.logFields())
		return
	}
	o.LogWriter().Info(internal.LI_EXECUTING_COMMAND, l.logFields())
	if *l.includeTracks {
		if l.validateTrackSorting(o) {
			o.LogWriter().Info(internal.LI_PARAMETERS_OVERRIDDEN, l.logFields())
		}
	}
	artists, ok := s.LoadData(o)
	if ok {
		l.outputArtists(o, artists)
	}
	return
}

func (l *ls) outputArtists(o internal.OutputBus, artists []*files.Artist) {
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
			o.WriteConsole(false, "Artist: %s\n", artistName)
			artist := artistsByArtistNames[artistName]
			l.outputAlbums(o, artist.Albums(), "  ")
		}
	case false:
		var albums []*files.Album
		for _, artist := range artists {
			albums = append(albums, artist.Albums()...)
		}
		l.outputAlbums(o, albums, "")
	}
}

func (l *ls) outputAlbums(o internal.OutputBus, albums []*files.Album, prefix string) {
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
			o.WriteConsole(false, "%sAlbum: %s\n", prefix, albumName)
			album := albumsByAlbumName[albumName]
			l.outputTracks(o, album.Tracks(), prefix+"  ")
		}
	case false:
		var tracks []*files.Track
		for _, album := range albums {
			tracks = append(tracks, album.Tracks()...)
		}
		l.outputTracks(o, tracks, prefix)
	}
}

func (l *ls) validateTrackSorting(o internal.OutputBus) (ok bool) {
	switch *l.trackSorting {
	case numericSorting:
		if !*l.includeAlbums {
			o.WriteError(internal.USER_INVALID_SORTING_APPLIED, fkTrackSortingFlag, *l.trackSorting, fkIncludeAlbumsFlag)
			o.LogWriter().Error(internal.LE_SORTING_OPTION_UNACCEPTABLE, map[string]interface{}{
				fkTrackSortingFlag:  *l.trackSorting,
				fkIncludeAlbumsFlag: *l.includeAlbums,
			})
			preferredValue := alphabeticSorting
			l.trackSorting = &preferredValue
		}
	case alphabeticSorting:
		ok = true
	default:
		o.WriteError(internal.USER_UNRECOGNIZED_VALUE, fkTrackSortingFlag, *l.trackSorting)
		o.LogWriter().Error(internal.LE_INVALID_FLAG_SETTING, map[string]interface{}{
			fkCommandName:      l.name(),
			fkTrackSortingFlag: *l.trackSorting,
		})
		var preferredValue string
		switch *l.includeAlbums {
		case true:
			preferredValue = numericSorting
		case false:
			preferredValue = alphabeticSorting
		}
		l.trackSorting = &preferredValue
	}
	return
}

func (l *ls) outputTracks(o internal.OutputBus, tracks []*files.Track, prefix string) {
	if !*l.includeTracks {
		return
	}
	switch *l.trackSorting {
	case numericSorting:
		trackNamesNumeric := make(map[int]string)
		tracksNumeric := make(map[int]*files.Track)
		var trackNumbers []int
		for _, track := range tracks {
			trackNumbers = append(trackNumbers, track.Number())
			trackNamesNumeric[track.Number()] = track.Name()
			tracksNumeric[track.Number()] = track
		}
		sort.Ints(trackNumbers)
		for _, trackNumber := range trackNumbers {
			o.WriteConsole(false, "%s%2d. %s\n", prefix, trackNumber, trackNamesNumeric[trackNumber])
			l.outputTrackDiagnostics(o, tracksNumeric[trackNumber], prefix+"  ")
		}
	case alphabeticSorting:
		var trackNames []string
		tracksByName := make(map[string]*files.Track)
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
			var trackName string
			if len(components) > 1 {
				var c2 []string
				c2 = append(c2, fmt.Sprintf("%q", components[0]))
				for k := 1; k < len(components); k += 2 {
					c2 = append(c2, components[k])
					c2 = append(c2, fmt.Sprintf("%q", components[k+1]))
				}
				trackName = strings.Join(c2, " ")
			} else {
				trackName = components[0]
			}
			trackNames = append(trackNames, trackName)
			tracksByName[trackName] = track
		}
		sort.Strings(trackNames)
		for _, trackName := range trackNames {
			o.WriteConsole(false, "%s%s\n", prefix, trackName)
			l.outputTrackDiagnostics(o, tracksByName[trackName], prefix+"  ")
		}
	}
}

func (l *ls) outputTrackDiagnostics(o internal.OutputBus, t *files.Track, prefix string) {
	if *l.diagnostics {
		if version, enc, frames, err := t.Diagnostics(); err != nil {
			o.LogWriter().Error(internal.LE_TAG_ERROR, map[string]interface{}{
				internal.FK_ERROR: err,
				fkTrack:           t.String(),
			})
			o.WriteError(internal.USER_TAG_ERROR, t.Name(), t.AlbumName(), t.RecordingArtist(), fmt.Sprintf("%v", err))
		} else {
			o.WriteConsole(false, "%sVersion: %v\n", prefix, version)
			o.WriteConsole(false, "%sEncoding: %q\n", prefix, enc)
			for _, frame := range frames {
				o.WriteConsole(false, "%s%s\n", prefix, frame.String())
			}
		}
	}
}
