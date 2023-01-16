package commands

import (
	"flag"
	"fmt"
	"mp3/internal"
	"mp3/internal/files"
	"sort"
	"strings"

	"github.com/majohn-r/output"
)

func init() {
	addCommandData(listCommandName, commandData{isDefault: true, init: newList})
	addDefaultMapping(listCommandName, map[string]any{
		annotateListingsFlag:  defaultAnnotateListings,
		detailsListingFlag:    defaultDetailsListing,
		diagnosticListingFlag: defaultDiagnosticListing,
		includeAlbumsFlag:     defaultIncludeAlbums,
		includeArtistsFlag:    defaultIncludeArtists,
		includeTracksFlag:     defaultIncludeTracks,
		trackSortingFlag:      defaultTrackSorting,
	})
	addDefaultMapping("command", map[string]any{"default": listCommandName})
}

type list struct {
	includeAlbums    *bool
	includeArtists   *bool
	includeTracks    *bool
	trackSorting     *string
	annotateListings *bool
	diagnostics      *bool
	details          *bool
	sf               *files.SearchFlags
}

func newList(o output.Bus, c *internal.Configuration, fSet *flag.FlagSet) (CommandProcessor, bool) {
	return newListCommand(o, c, fSet)
}

const (
	listCommandName = "list"

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

func newListCommand(o output.Bus, c *internal.Configuration, fSet *flag.FlagSet) (*list, bool) {
	defaults, ok := evaluateListDefaults(o, c.SubConfiguration(listCommandName))
	sFlags, sFlagsOk := files.NewSearchFlags(o, c, fSet)
	if ok && sFlagsOk {
		albumUsage := internal.DecorateBoolFlagUsage("include album names in listing", defaults.includeAlbums)
		artistUsage := internal.DecorateBoolFlagUsage("include artist names in listing", defaults.includeArtists)
		trackUsage := internal.DecorateBoolFlagUsage("include track names in listing", defaults.includeTracks)
		sortingUsage := internal.DecorateStringFlagUsage("track `sorting`, 'numeric' in track number order, or 'alpha' in track name order", defaults.sorting)
		annotateUsage := internal.DecorateBoolFlagUsage("annotate listings with album and artist data", defaults.annotateTracks)
		diagnosticUsage := internal.DecorateBoolFlagUsage("include diagnostic information with tracks", defaults.diagnostics)
		detailsUsage := internal.DecorateBoolFlagUsage("include details with tracks", defaults.details)
		return &list{
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

func evaluateListDefaults(o output.Bus, c *internal.Configuration) (defaults listDefaults, ok bool) {
	ok = true
	defaults = listDefaults{}
	var err error
	defaults.includeAlbums, err = c.BoolDefault(includeAlbumsFlag, defaultIncludeAlbums)
	if err != nil {
		reportBadDefault(o, listCommandName, err)
		ok = false
	}
	defaults.includeArtists, err = c.BoolDefault(includeArtistsFlag, defaultIncludeArtists)
	if err != nil {
		reportBadDefault(o, listCommandName, err)
		ok = false
	}
	defaults.includeTracks, err = c.BoolDefault(includeTracksFlag, defaultIncludeTracks)
	if err != nil {
		reportBadDefault(o, listCommandName, err)
		ok = false
	}
	defaults.annotateTracks, err = c.BoolDefault(annotateListingsFlag, defaultAnnotateListings)
	if err != nil {
		reportBadDefault(o, listCommandName, err)
		ok = false
	}
	defaults.diagnostics, err = c.BoolDefault(diagnosticListingFlag, defaultDiagnosticListing)
	if err != nil {
		reportBadDefault(o, listCommandName, err)
		ok = false
	}
	defaults.details, err = c.BoolDefault(detailsListingFlag, defaultDetailsListing)
	if err != nil {
		reportBadDefault(o, listCommandName, err)
		ok = false
	}
	defaults.sorting, err = c.StringDefault(trackSortingFlag, defaultTrackSorting)
	if err != nil {
		reportBadDefault(o, listCommandName, err)
		ok = false
	}
	return
}

func (l *list) Exec(o output.Bus, args []string) (ok bool) {
	if s, argsOk := l.sf.ProcessArgs(o, args); argsOk {
		ok = l.runCommand(o, s)
	}
	return
}

func (l *list) logFields() map[string]any {
	return map[string]any{
		"command":                   listCommandName,
		"-" + includeAlbumsFlag:     *l.includeAlbums,
		"-" + includeArtistsFlag:    *l.includeArtists,
		"-" + includeTracksFlag:     *l.includeTracks,
		"-" + trackSortingFlag:      *l.trackSorting,
		"-" + annotateListingsFlag:  *l.annotateListings,
		"-" + diagnosticListingFlag: *l.diagnostics,
		"-" + detailsListingFlag:    *l.details,
	}
}

func (l *list) runCommand(o output.Bus, s *files.Search) (ok bool) {
	if !*l.includeArtists && !*l.includeAlbums && !*l.includeTracks {
		reportNothingToDo(o, listCommandName, l.logFields())
		return
	}
	logStart(o, listCommandName, l.logFields())
	if *l.includeTracks {
		if l.validateTrackSorting(o) {
			o.Log(output.Info, "one or more flags were overridden", l.logFields())
		}
	}
	if artists, artistsLoaded := s.Load(o); artistsLoaded {
		l.outputArtists(o, artists)
		ok = true
	}
	return
}

func (l *list) outputArtists(o output.Bus, artists []*files.Artist) {
	switch *l.includeArtists {
	case true:
		m := make(map[string]*files.Artist)
		var names []string
		for _, a := range artists {
			m[a.Name()] = a
			names = append(names, a.Name())
		}
		sort.Strings(names)
		for _, s := range names {
			o.WriteConsole("Artist: %s\n", s)
			l.outputAlbums(o, m[s].Albums(), "  ")
		}
	case false:
		var albums []*files.Album
		for _, a := range artists {
			albums = append(albums, a.Albums()...)
		}
		l.outputAlbums(o, albums, "")
	}
}

func (l *list) outputAlbums(o output.Bus, albums []*files.Album, prefix string) {
	switch *l.includeAlbums {
	case true:
		m := make(map[string]*files.Album)
		var names []string
		for _, a := range albums {
			var s string
			switch {
			case !*l.includeArtists && *l.annotateListings:
				s = fmt.Sprintf("%q by %q", a.Name(), a.RecordingArtistName())
			default:
				s = a.Name()
			}
			m[s] = a
			names = append(names, s)
		}
		sort.Strings(names)
		for _, s := range names {
			o.WriteConsole("%sAlbum: %s\n", prefix, s)
			l.outputTracks(o, m[s].Tracks(), prefix+"  ")
		}
	case false:
		var tracks []*files.Track
		for _, a := range albums {
			tracks = append(tracks, a.Tracks()...)
		}
		l.outputTracks(o, tracks, prefix)
	}
}

func (l *list) validateTrackSorting(o output.Bus) (ok bool) {
	switch *l.trackSorting {
	case numericSorting:
		if !*l.includeAlbums {
			o.WriteCanonicalError(`The "-%s" value you specified, %q, is not valid unless "-%s" is true; track sorting will be alphabetic`, trackSortingFlag, *l.trackSorting, includeAlbumsFlag)
			o.Log(output.Error, "numeric track sorting is not applicable", map[string]any{
				"-" + trackSortingFlag:  *l.trackSorting,
				"-" + includeAlbumsFlag: *l.includeAlbums,
			})
			v := alphabeticSorting
			l.trackSorting = &v
		}
	case alphabeticSorting:
		ok = true
	default:
		o.WriteCanonicalError(`The "-%s" value you specified, %q, is not valid`, trackSortingFlag, *l.trackSorting)
		o.Log(output.Error, "flag value is not valid", map[string]any{
			"command":              listCommandName,
			"-" + trackSortingFlag: *l.trackSorting,
		})
		var v string
		switch *l.includeAlbums {
		case true:
			v = numericSorting
		case false:
			v = alphabeticSorting
		}
		l.trackSorting = &v
	}
	return
}

func (l *list) outputTracks(o output.Bus, tracks []*files.Track, prefix string) {
	if !*l.includeTracks {
		return
	}
	switch *l.trackSorting {
	case numericSorting:
		m := make(map[int]*files.Track)
		var numbers []int
		for _, t := range tracks {
			numbers = append(numbers, t.Number())
			m[t.Number()] = t
		}
		sort.Ints(numbers)
		for _, n := range numbers {
			o.WriteConsole("%s%2d. %s\n", prefix, n, m[n].CommonName())
			l.outputTrackDetails(o, m[n], prefix+"  ")
			l.outputTrackDiagnostics(o, m[n], prefix+"  ")
		}
	case alphabeticSorting:
		var annotatedNames []string
		m := make(map[string]*files.Track)
		for _, t := range tracks {
			var s []string
			s = append(s, t.CommonName())
			if *l.annotateListings && !(*l.includeAlbums && *l.includeArtists) {
				if !*l.includeAlbums {
					s = append(s, []string{"on", t.AlbumName()}...)
					if !*l.includeArtists {
						s = append(s, []string{"by", t.RecordingArtist()}...)
					}
				}
			}
			var name string
			if len(s) > 1 {
				var c2 []string
				c2 = append(c2, fmt.Sprintf("%q", s[0]))
				for k := 1; k < len(s); k += 2 {
					c2 = append(c2, s[k], fmt.Sprintf("%q", s[k+1]))
				}
				name = strings.Join(c2, " ")
			} else {
				name = s[0]
			}
			annotatedNames = append(annotatedNames, name)
			m[name] = t
		}
		sort.Strings(annotatedNames)
		for _, s := range annotatedNames {
			o.WriteConsole("%s%s\n", prefix, s)
			l.outputTrackDetails(o, m[s], prefix+"  ")
			l.outputTrackDiagnostics(o, m[s], prefix+"  ")
		}
	}
}

func (l *list) outputTrackDetails(o output.Bus, t *files.Track, prefix string) {
	if *l.details {
		// go get information from track and display it
		if m, err := t.Details(); err != nil {
			o.Log(output.Error, "cannot get details", map[string]any{
				"error": err,
				"track": t.String(),
			})
			o.WriteCanonicalError("The details are not available for track %q on album %q by artist %q: %q", t.CommonName(), t.AlbumName(), t.RecordingArtist(), err.Error())
		} else if len(m) != 0 {
			var keys []string
			for k := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			o.WriteConsole("%sDetails:\n", prefix)
			for _, k := range keys {
				o.WriteConsole("%s  %s = %q\n", prefix, k, m[k])
			}
		}
	}
}

func (l *list) outputTrackDiagnostics(o output.Bus, t *files.Track, prefix string) {
	if *l.diagnostics {
		if version, enc, frames, err := t.ID3V2Diagnostics(); err != nil {
			t.ReportMetadataReadError(o, files.ID3V2, err)
		} else {
			o.WriteConsole("%sID3V2 Version: %v\n", prefix, version)
			o.WriteConsole("%sID3V2 Encoding: %q\n", prefix, enc)
			for _, frame := range frames {
				o.WriteConsole("%sID3V2 %s\n", prefix, frame)
			}
		}
		if v1, err := t.ID3V1Diagnostics(); err != nil {
			t.ReportMetadataReadError(o, files.ID3V1, err)
		} else {
			for _, s := range v1 {
				o.WriteConsole("%sID3V1 %s\n", prefix, s)
			}
		}
	}
}
