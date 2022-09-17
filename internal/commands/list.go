package commands

import (
	"flag"
	"fmt"
	"mp3/internal"
	"mp3/internal/files"
	"sort"
	"strings"
)

type list struct {
	n                string
	includeAlbums    *bool
	includeArtists   *bool
	includeTracks    *bool
	trackSorting     *string
	annotateListings *bool
	diagnostics      *bool
	details          *bool
	sf               *files.SearchFlags
}

func (l *list) name() string {
	return l.n
}

func newList(o internal.OutputBus, c *internal.Configuration, fSet *flag.FlagSet) (CommandProcessor, bool) {
	return newListCommand(o, c, fSet)
}

const (
	alphabeticSorting = "alpha"
	numericSorting    = "numeric"

	defaultAnnotateListings  = false
	defaultDetailsListing    = false
	defaultDiagnosticListing = false
	defaultIncludeAlbums     = true
	defaultIncludeArtists    = true
	defaultIncludeTracks     = false
	defaultTrackSorting      = numericSorting

	annotateListingsFlag  = "annotate"
	detailsListingFlag    = "details"
	diagnosticListingFlag = "diagnostic"
	includeAlbumsFlag     = "includeAlbums"
	includeArtistsFlag    = "includeArtists"
	includeTracksFlag     = "includeTracks"
	trackSortingFlag      = "sort"

	fkAnnotateListingsFlag  = "-" + annotateListingsFlag
	fkDetailsListingFlag    = "-" + detailsListingFlag
	fkDiagnosticListingFlag = "-" + diagnosticListingFlag
	fkIncludeAlbumsFlag     = "-" + includeAlbumsFlag
	fkIncludeArtistsFlag    = "-" + includeArtistsFlag
	fkIncludeTracksFlag     = "-" + includeTracksFlag
	fkTrackSortingFlag      = "-" + trackSortingFlag
	fkTrack                 = "track"
)

type listDefaults struct {
	includeAlbums  bool
	includeArtists bool
	includeTracks  bool
	annotateTracks bool
	diagnostics    bool
	details        bool
	sorting        string
}

func newListCommand(o internal.OutputBus, c *internal.Configuration, fSet *flag.FlagSet) (*list, bool) {
	name := fSet.Name()
	defaults, defaultsOk := evaluateListDefaults(o, c.SubConfiguration(name), name)
	sFlags, sFlagsOk := files.NewSearchFlags(o, c, fSet)
	if defaultsOk && sFlagsOk {
		albumUsage := internal.DecorateBoolFlagUsage("include album names in listing", defaults.includeAlbums)
		artistUsage := internal.DecorateBoolFlagUsage("include artist names in listing", defaults.includeArtists)
		trackUsage := internal.DecorateBoolFlagUsage("include track names in listing", defaults.includeTracks)
		sortingUsage := internal.DecorateStringFlagUsage("track `sorting`, 'numeric' in track number order, or 'alpha' in track name order", defaults.sorting)
		annotateUsage := internal.DecorateBoolFlagUsage("annotate listings with album and artist data", defaults.annotateTracks)
		diagnosticUsage := internal.DecorateBoolFlagUsage("include diagnostic information with tracks", defaults.diagnostics)
		detailsUsage := internal.DecorateBoolFlagUsage("include details with tracks", defaults.details)
		return &list{
			n:                name,
			includeAlbums:    fSet.Bool(includeAlbumsFlag, defaults.includeAlbums, albumUsage),
			includeArtists:   fSet.Bool(includeArtistsFlag, defaults.includeArtists, artistUsage),
			includeTracks:    fSet.Bool(includeTracksFlag, defaults.includeTracks, trackUsage),
			trackSorting:     fSet.String(trackSortingFlag, defaults.sorting, sortingUsage),
			annotateListings: fSet.Bool(annotateListingsFlag, defaults.annotateTracks, annotateUsage),
			diagnostics:      fSet.Bool(diagnosticListingFlag, defaults.diagnostics, diagnosticUsage),
			details:          fSet.Bool(detailsListingFlag, defaults.details, detailsUsage),
			sf:               sFlags,
		}, true
	}
	return nil, false
}

func evaluateListDefaults(o internal.OutputBus, c *internal.Configuration, name string) (defaults listDefaults, ok bool) {
	ok = true
	defaults = listDefaults{}
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
	defaults.details, err = c.BoolDefault(detailsListingFlag, defaultDetailsListing)
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

func (l *list) Exec(o internal.OutputBus, args []string) (ok bool) {
	if s, argsOk := l.sf.ProcessArgs(o, args); argsOk {
		ok = l.runCommand(o, s)
	}
	return
}

func (l *list) logFields() map[string]interface{} {
	return map[string]interface{}{
		fkCommandName:           l.name(),
		fkIncludeAlbumsFlag:     *l.includeAlbums,
		fkIncludeArtistsFlag:    *l.includeArtists,
		fkIncludeTracksFlag:     *l.includeTracks,
		fkTrackSortingFlag:      *l.trackSorting,
		fkAnnotateListingsFlag:  *l.annotateListings,
		fkDiagnosticListingFlag: *l.diagnostics,
		fkDetailsListingFlag:    *l.details,
	}
}

func (l *list) runCommand(o internal.OutputBus, s *files.Search) (ok bool) {
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

func (l *list) outputArtists(o internal.OutputBus, artists []*files.Artist) {
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

func (l *list) outputAlbums(o internal.OutputBus, albums []*files.Album, prefix string) {
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

func (l *list) validateTrackSorting(o internal.OutputBus) (ok bool) {
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

func (l *list) outputTracks(o internal.OutputBus, tracks []*files.Track, prefix string) {
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
			l.outputTrackDetails(o, tracksNumeric[trackNumber], prefix+"  ")
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
			l.outputTrackDetails(o, tracksByName[trackName], prefix+"  ")
			l.outputTrackDiagnostics(o, tracksByName[trackName], prefix+"  ")
		}
	}
}

func (l *list) outputTrackDetails(o internal.OutputBus, t *files.Track, prefix string) {
	if *l.details {
		// go get information from track and display it
		if m, err := t.Details(); err != nil {
			o.LogWriter().Error(internal.LE_CANNOT_GET_TRACK_DETAILS, map[string]any{
				internal.FK_ERROR: err,
				fkTrack:           t.String(),
			})
			o.WriteError(internal.USER_CANNOT_READ_TRACK_DETAILS, t.Name(), t.AlbumName(), t.RecordingArtist(), fmt.Sprintf("%v", err))
		} else {
			if len(m) != 0 {
				var keys []string
				for k := range m {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				o.WriteConsole(false, "%sDetails:\n", prefix)
				for _, k := range keys {
					o.WriteConsole(false, "%s  %s = %q\n", prefix, k, m[k])
				}
			}
		}
	}
}

func (l *list) outputTrackDiagnostics(o internal.OutputBus, t *files.Track, prefix string) {
	if *l.diagnostics {
		if version, enc, frames, err := t.ID3V2Diagnostics(); err != nil {
			o.LogWriter().Error(internal.LE_ID3V2_TAG_ERROR, map[string]interface{}{
				internal.FK_ERROR: err,
				fkTrack:           t.String(),
			})
			o.WriteError(internal.USER_ID3V2_TAG_ERROR, t.Name(), t.AlbumName(), t.RecordingArtist(), fmt.Sprintf("%v", err))
		} else {
			o.WriteConsole(false, "%sID3V2 Version: %v\n", prefix, version)
			o.WriteConsole(false, "%sID3V2 Encoding: %q\n", prefix, enc)
			for _, frame := range frames {
				o.WriteConsole(false, "%sID3V2 %s\n", prefix, frame)
			}
		}
		if id3v1Data, err := t.ID3V1Diagnostics(); err != nil {
			o.LogWriter().Error(internal.LE_ID3V1_TAG_ERROR, map[string]interface{}{
				internal.FK_ERROR: err,
				fkTrack:           t.String(),
			})
			o.WriteError(internal.USER_ID3V1_TAG_ERROR, t.Name(), t.AlbumName(), t.RecordingArtist(), fmt.Sprintf("%v", err))
		} else {
			for _, datum := range id3v1Data {
				o.WriteConsole(false, "%sID3V1 %s\n", prefix, datum)
			}
		}
	}
}
