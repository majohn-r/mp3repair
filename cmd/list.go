package cmd

import (
	"fmt"
	"mp3repair/internal/files"
	"sort"
	"strings"

	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

const (
	ListAlbums           = "albums"
	ListAlbumsFlag       = "--" + ListAlbums
	ListAnnotate         = "annotate"
	ListAnnotateFlag     = "--" + ListAnnotate
	ListArtists          = "artists"
	ListArtistsFlag      = "--" + ListArtists
	ListCommand          = "list"
	ListDetails          = "details"
	ListDetailsFlag      = "--" + ListDetails
	ListDiagnostic       = "diagnostic"
	ListDiagnosticFlag   = "--" + ListDiagnostic
	ListSortByNumber     = "byNumber"
	ListSortByNumberFlag = "--" + ListSortByNumber
	ListSortByTitle      = "byTitle"
	ListSortByTitleFlag  = "--" + ListSortByTitle
	ListTracks           = "tracks"
	ListTracksFlag       = "--" + ListTracks
)

var (
	// ListCmd represents the list command
	ListCmd = &cobra.Command{
		Use: ListCommand + " [" + ListAlbumsFlag + "] [" + ListArtistsFlag + "] " +
			"[" + ListTracksFlag + "] [" + ListAnnotateFlag + "] [" + ListDetailsFlag + "] " +
			"[" + ListDiagnosticFlag + "] [" + ListSortByNumberFlag + " | " +
			ListSortByTitleFlag + "] " + searchUsage,
		DisableFlagsInUseLine: true,
		Short:                 "Lists mp3 files and containing album and artist directories",
		Long: fmt.Sprintf(
			"%q lists mp3 files and containing album and artist directories", ListCommand),
		Example: ListCommand + " " + ListAnnotateFlag + "\n" +
			"  Annotate tracks with album and artist data and albums with artist data\n" +
			ListCommand + " " + ListDetailsFlag + "\n" +
			"  Include detailed information, if available, for each track. This includes" +
			" composer,\n" +
			"  conductor, key, lyricist, orchestra/band, and subtitle\n" +
			ListCommand + " " + ListAlbumsFlag + "\n" +
			"  Include the album names in the output\n" +
			ListCommand + " " + ListArtistsFlag + "\n" +
			"  Include the artist names in the output\n" +
			ListCommand + " " + ListTracksFlag + "\n" +
			"  Include the track names in the output\n" +
			ListCommand + " " + ListSortByTitleFlag + "\n" +
			"  Sort tracks by name, ignoring track numbers\n" +
			ListCommand + " " + ListSortByNumberFlag + "\n" +
			"  Sort tracks by track number",
		RunE: ListRun,
	}
	ListFlags = &SectionFlags{
		SectionName: ListCommand,
		Details: map[string]*FlagDetails{
			ListAlbums: {
				AbbreviatedName: "l",
				Usage:           "include album names in listing",
				ExpectedType:    BoolType,
				DefaultValue:    false,
			},
			ListArtists: {
				AbbreviatedName: "r",
				Usage:           "include artist names in listing",
				ExpectedType:    BoolType,
				DefaultValue:    false,
			},
			ListTracks: {
				AbbreviatedName: "t",
				Usage:           "include track names in listing",
				ExpectedType:    BoolType,
				DefaultValue:    false,
			},
			ListSortByNumber: {
				Usage:        "sort tracks by track number",
				ExpectedType: BoolType,
				DefaultValue: false,
			},
			ListSortByTitle: {
				Usage:        "sort tracks by track title",
				ExpectedType: BoolType,
				DefaultValue: false,
			},
			ListAnnotate: {
				Usage:        "annotate listings with album and artist names",
				ExpectedType: BoolType,
				DefaultValue: false,
			},
			ListDetails: {
				Usage:        "include details with tracks",
				ExpectedType: BoolType,
				DefaultValue: false,
			},
			ListDiagnostic: {
				Usage:        "include diagnostic information with tracks",
				ExpectedType: BoolType,
				DefaultValue: false,
			},
		},
	}
)

func ListRun(cmd *cobra.Command, _ []string) error {
	exitError := NewExitProgrammingError(ListCommand)
	o := getBus()
	producer := cmd.Flags()
	values, eSlice := ReadFlags(producer, ListFlags)
	searchSettings, searchFlagsOk := EvaluateSearchFlags(o, producer)
	if ProcessFlagErrors(o, eSlice) && searchFlagsOk {
		if ls, flagsOk := ProcessListFlags(o, values); flagsOk {
			details := map[string]any{
				ListAlbumsFlag:       ls.Albums.Value,
				"albums-user-set":    ls.Albums.UserSet,
				ListAnnotateFlag:     ls.Annotate.Value,
				ListArtistsFlag:      ls.Artists.Value,
				"artists-user-set":   ls.Artists.UserSet,
				ListSortByNumberFlag: ls.SortByNumber.Value,
				"byNumber-user-set":  ls.SortByNumber.UserSet,
				ListSortByTitleFlag:  ls.SortByTitle.Value,
				"byTitle-user-set":   ls.SortByTitle.UserSet,
				ListDetailsFlag:      ls.Details.Value,
				ListDiagnosticFlag:   ls.Diagnostic.Value,
				ListTracksFlag:       ls.Tracks.Value,
				"tracks-user-set":    ls.Tracks.UserSet,
			}
			for k, v := range searchSettings.Values() {
				details[k] = v
			}
			LogCommandStart(o, ListCommand, details)
			switch ls.HasWorkToDo(o) {
			case true:
				switch ls.TracksSortable(o) {
				case true:
					allArtists := searchSettings.Load(o)
					exitError = ls.ListArtists(o, allArtists, searchSettings)
				case false:
					exitError = NewExitUserError(ListCommand)
				}
			case false:
				exitError = NewExitUserError(ListCommand)
			}
		}
	}
	return ToErrorInterface(exitError)
}

type ListSettings struct {
	Albums       CommandFlag[bool]
	Annotate     CommandFlag[bool]
	Artists      CommandFlag[bool]
	Details      CommandFlag[bool]
	Diagnostic   CommandFlag[bool]
	SortByNumber CommandFlag[bool]
	SortByTitle  CommandFlag[bool]
	Tracks       CommandFlag[bool]
}

func (ls *ListSettings) ListArtists(o output.Bus, allArtists []*files.Artist, ss *SearchSettings) (err *ExitError) {
	err = NewExitUserError(ListCommand)
	if len(allArtists) != 0 {
		if filteredArtists := ss.Filter(o, allArtists); len(filteredArtists) != 0 {
			ls.ListFilteredArtists(o, filteredArtists)
			err = nil
		}
	}
	return err
}

func (ls *ListSettings) ListFilteredArtists(o output.Bus, artists []*files.Artist) {
	if ls.Artists.Value {
		m := map[string]*files.Artist{}
		names := make([]string, 0, len(artists))
		for _, a := range artists {
			m[a.Name] = a
			names = append(names, a.Name)
		}
		sort.Strings(names)
		for _, s := range names {
			o.WriteConsole("Artist: %s\n", s)
			artist := m[s]
			if artist != nil {
				o.IncrementTab(2)
				ls.ListAlbums(o, artist.Albums)
				o.DecrementTab(2)
			}
		}
		return
	}
	albumCount := 0
	for _, a := range artists {
		albumCount += len(a.Albums)
	}
	albums := make([]*files.Album, 0, albumCount)
	for _, a := range artists {
		albums = append(albums, a.Albums...)
	}
	ls.ListAlbums(o, albums)
}

type AlbumSlice []*files.Album

func (as AlbumSlice) Len() int {
	return len(as)
}

func (as AlbumSlice) Less(i, j int) bool {
	if as[i].Title == as[j].Title {
		return as[i].RecordingArtistName() < as[j].RecordingArtistName()
	}
	return as[i].Title < as[j].Title
}

func (as AlbumSlice) Swap(i, j int) {
	as[i], as[j] = as[j], as[i]
}

func (ls *ListSettings) ListAlbums(o output.Bus, albums []*files.Album) {
	if ls.Albums.Value {
		sort.Sort(AlbumSlice(albums))
		for _, album := range albums {
			o.WriteConsole("Album: %s\n", ls.AnnotateAlbumName(album))
			o.IncrementTab(2)
			ls.ListTracks(o, album.Tracks)
			o.DecrementTab(2)
		}
		return
	}
	trackCount := 0
	for _, album := range albums {
		trackCount += len(album.Tracks)
	}
	tracks := make([]*files.Track, 0, trackCount)
	for _, album := range albums {
		tracks = append(tracks, album.Tracks...)
	}
	ls.ListTracks(o, tracks)
}

func (ls *ListSettings) AnnotateAlbumName(album *files.Album) string {
	switch {
	case !ls.Artists.Value && ls.Annotate.Value:
		return strings.Join([]string{quote(album.Title), "by",
			quote(album.RecordingArtistName())}, " ")
	default:
		return album.Title
	}
}

func (ls *ListSettings) ListTracks(o output.Bus, tracks []*files.Track) {
	if !ls.Tracks.Value {
		return
	}
	if ls.SortByNumber.Value {
		ls.ListTracksByNumber(o, tracks)
		return
	}
	if ls.SortByTitle.Value {
		ls.ListTracksByName(o, tracks)
	}
}

func (ls *ListSettings) ListTracksByNumber(o output.Bus, tracks []*files.Track) {
	m := map[int]*files.Track{}
	numbers := make([]int, 0, len(tracks))
	for _, track := range tracks {
		numbers = append(numbers, track.Number)
		m[track.Number] = track
	}
	sort.Ints(numbers)
	for _, n := range numbers {
		track := m[n]
		if track != nil {
			o.WriteConsole("%2d. %s\n", n, track.SimpleName)
			o.IncrementTab(2)
			ls.ListTrackDetails(o, track)
			ls.ListTrackDiagnostics(o, track)
			o.DecrementTab(2)
		}
	}
}

type TrackSlice []*files.Track

func (ts TrackSlice) Len() int {
	return len(ts)
}

func (ts TrackSlice) Less(i, j int) bool {
	if ts[i].SimpleName == ts[j].SimpleName {
		album1 := ts[i].Album
		album2 := ts[j].Album
		if album1.Title == album2.Title {
			return album1.RecordingArtistName() < album2.RecordingArtistName()
		}
		return album1.Title < album2.Title
	}
	return ts[i].SimpleName < ts[j].SimpleName
}

func (ts TrackSlice) Swap(i, j int) {
	ts[i], ts[j] = ts[j], ts[i]
}

func (ls *ListSettings) ListTracksByName(o output.Bus, tracks []*files.Track) {
	sort.Sort(TrackSlice(tracks))
	for _, track := range tracks {
		o.WriteConsole("%s\n", ls.AnnotateTrackName(track))
		o.IncrementTab(2)
		ls.ListTrackDetails(o, track)
		ls.ListTrackDiagnostics(o, track)
		o.DecrementTab(2)
	}
}

func quote(s string) string {
	return fmt.Sprintf("%q", s)
}

func (ls *ListSettings) AnnotateTrackName(track *files.Track) string {
	commonName := track.SimpleName
	if !ls.Annotate.Value || ls.Albums.Value {
		return commonName
	}
	trackNameParts := []string{quote(commonName), "on", quote(track.AlbumName())}
	if !ls.Artists.Value {
		trackNameParts = append(trackNameParts, "by", quote(track.RecordingArtist()))
	}
	return strings.Join(trackNameParts, " ")
}

func (ls *ListSettings) ListTrackDetails(o output.Bus, track *files.Track) {
	if ls.Details.Value {
		// go get information from track and display it
		m, readErr := track.Details()
		ShowDetails(o, track, m, readErr)
	}
}

// ShowDetails outputs track details; this functionality is split out for testing
func ShowDetails(o output.Bus, track *files.Track, details map[string]string, detailsError error) {
	if detailsError != nil {
		o.Log(output.Error, "cannot get details", map[string]any{
			"error": detailsError,
			"track": track.String(),
		})
		o.WriteCanonicalError(
			"The details are not available for track %q on album %q by artist %q: %q",
			track.SimpleName, track.AlbumName(), track.RecordingArtist(),
			detailsError.Error())
		return
	}
	if len(details) == 0 {
		return
	}
	keys := make([]string, 0, len(details))
	for k := range details {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	o.WriteConsole("Details:\n")
	o.IncrementTab(2)
	for _, k := range keys {
		o.WriteConsole("%s = %q\n", k, details[k])
	}
	o.DecrementTab(2)
}

func (ls *ListSettings) ListTrackDiagnostics(o output.Bus, track *files.Track) {
	if ls.Diagnostic.Value {
		info, ID3V2readErr := track.ID3V2Diagnostics()
		ShowID3V2Diagnostics(o, track, info, ID3V2readErr)
		tags, ID3V1readErr := track.ID3V1Diagnostics()
		ShowID3V1Diagnostics(o, track, tags, ID3V1readErr)
	}
}

// ShowID3V1Diagnostics outputs ID3V1 diagnostic data; this function is split out for testing
func ShowID3V1Diagnostics(o output.Bus, track *files.Track, tags []string, readErr error) {
	if readErr != nil {
		track.ReportMetadataReadError(o, files.ID3V1, readErr.Error())
		return
	}
	for _, s := range tags {
		o.WriteConsole("ID3V1 %s\n", s)
	}
}

// ShowID3V2Diagnostics outputs ID3V2 diagnostic data; this functionality is split out for
// testing
func ShowID3V2Diagnostics(o output.Bus, track *files.Track, info *files.ID3V2Info, readErr error) {
	if readErr != nil {
		track.ReportMetadataReadError(o, files.ID3V2, readErr.Error())
		return
	}
	if info != nil {
		o.WriteConsole("ID3V2 Version: %v\n", info.Version)
		o.WriteConsole("ID3V2 Encoding: %q\n", info.Encoding)
		for _, frame := range info.FrameStrings {
			o.WriteConsole("ID3V2 %s\n", frame)
		}
	}
}

func (ls *ListSettings) TracksSortable(o output.Bus) bool {
	bothSortingOptionsSet := ls.SortByNumber.Value && ls.SortByTitle.Value
	neitherSortingOptionSet := !ls.SortByNumber.Value && !ls.SortByTitle.Value
	if ls.Tracks.Value {
		switch {
		case bothSortingOptionsSet:
			o.WriteCanonicalError("Track sorting cannot be done")
			o.WriteCanonicalError("Why?")
			switch ls.SortByNumber.UserSet {
			case true:
				switch ls.SortByTitle.UserSet {
				case true:
					o.WriteCanonicalError("You explicitly set %s and %s true",
						ListSortByNumberFlag, ListSortByTitleFlag)
				case false:
					o.WriteCanonicalError(
						"The %s flag is configured true and you explicitly set %s true",
						ListSortByTitleFlag, ListSortByNumberFlag)
				}
			case false:
				switch ls.SortByTitle.UserSet {
				case true:
					o.WriteCanonicalError(
						"The %s flag is configured true and you explicitly set %s true",
						ListSortByNumberFlag, ListSortByTitleFlag)
				case false:
					o.WriteCanonicalError("The %s and %s flags are both configured true",
						ListSortByNumberFlag, ListSortByTitleFlag)
				}
			}
			o.WriteCanonicalError("What to do:\nEither edit the configuration file and use" +
				" those default values, or use appropriate command line values")
			return false
		case ls.SortByNumber.Value && !ls.Albums.Value:
			o.WriteCanonicalError("Sorting tracks by number not possible.")
			o.WriteCanonicalError("Why?")
			o.WriteCanonicalError("Track numbers are only relevant if albums are also output.")
			switch ls.SortByNumber.UserSet {
			case true:
				switch ls.Albums.UserSet {
				case true:
					o.WriteCanonicalError("You set %s true and %s false.", ListSortByNumberFlag, ListAlbumsFlag)
				case false:
					o.WriteCanonicalError("You set %s true and %s is configured as false", ListSortByNumberFlag, ListAlbumsFlag)
				}
			case false:
				switch ls.Albums.UserSet {
				case true:
					o.WriteCanonicalError("You set %s false and %s is configured as true", ListAlbumsFlag, ListSortByNumberFlag)
				case false:
					o.WriteCanonicalError("%s is configured as false, and %s is configured as true", ListAlbumsFlag, ListSortByNumberFlag)
				}
			}
			o.WriteCanonicalError("What to do:\nEither edit the configuration file or change which flags you set on the command line.")
			return false
		case neitherSortingOptionSet:
			if ls.SortByNumber.UserSet && ls.SortByTitle.UserSet {
				o.WriteCanonicalError("A listing of tracks is not possible.")
				o.WriteCanonicalError("Why?")
				o.WriteCanonicalError("Tracks are enabled, but you set both %s and %s false", ListSortByNumberFlag, ListSortByTitleFlag)
				o.WriteCanonicalError("What to do:\nEnable one of the sorting flags")
				return false
			}
			// pick a sensible option
			switch {
			case ls.SortByNumber.UserSet:
				ls.SortByTitle.Value = true // pick the other setting
			case ls.SortByTitle.UserSet:
				ls.SortByNumber.Value = true // pick the other setting
			default: // ok, pick something sensible, user does not care
				switch ls.Albums.Value {
				case true:
					ls.SortByNumber.Value = true
				case false:
					ls.SortByTitle.Value = true
				}
			}
			o.Log(output.Info, "no track sorting set, providing a sensible value", map[string]any{
				ListAlbumsFlag:      ls.Albums.Value,
				ListSortByNumber:    ls.SortByNumber.Value,
				ListSortByTitleFlag: ls.SortByTitle.Value,
			})
			return true
		default: // https://github.com/majohn-r/mp3repair/issues/170
			return true
		}
	}
	if (ls.SortByNumber.Value && ls.SortByNumber.UserSet) || (ls.SortByTitle.Value && ls.SortByTitle.UserSet) {
		o.WriteCanonicalError("Your sorting preferences are not relevant")
		o.WriteCanonicalError("Why?")
		o.WriteCanonicalError(
			"Tracks are not included in the output, but you explicitly set %s or %s true.",
			ListSortByNumberFlag, ListSortByTitleFlag)
		o.WriteCanonicalError("What to do:\nEither set %s true or remove the sorting flags"+
			" from the command line.", ListTracksFlag)
		return false
	}
	return true
}

func (ls *ListSettings) HasWorkToDo(o output.Bus) bool {
	if ls.Albums.Value || ls.Artists.Value || ls.Tracks.Value {
		return true
	}
	o.WriteCanonicalError("No listing will be output.\nWhy?\n")
	switch {
	case ls.Albums.UserSet || ls.Artists.UserSet || ls.Tracks.UserSet:
		flagsUserSet := make([]string, 0, 3)
		flagsFromConfig := make([]string, 0, 3)
		switch {
		case ls.Albums.UserSet:
			flagsUserSet = append(flagsUserSet, ListAlbumsFlag)
		default:
			flagsFromConfig = append(flagsFromConfig, ListAlbumsFlag)
		}
		switch {
		case ls.Artists.UserSet:
			flagsUserSet = append(flagsUserSet, ListArtistsFlag)
		default:
			flagsFromConfig = append(flagsFromConfig, ListArtistsFlag)
		}
		switch {
		case ls.Tracks.UserSet:
			flagsUserSet = append(flagsUserSet, ListTracksFlag)
		default:
			flagsFromConfig = append(flagsFromConfig, ListTracksFlag)
		}
		switch len(flagsFromConfig) {
		case 0:
			o.WriteCanonicalError("You explicitly set %s, %s, and %s false",
				ListAlbumsFlag, ListArtistsFlag, ListTracksFlag)
		default:
			o.WriteCanonicalError(
				"In addition to %s configured false, you explicitly set %s false",
				strings.Join(flagsFromConfig, " and "), strings.Join(flagsUserSet, " and "))
		}
	default: // user did not set any of the relevant flags that way
		o.WriteCanonicalError("The flags %s, %s, and %s are all configured false",
			ListAlbumsFlag, ListArtistsFlag, ListTracksFlag)
	}
	o.WriteError("What to do:\n")
	o.WriteCanonicalError("Either:\n[1] Edit the configuration file so that at least one" +
		" of these flags is true, or\n[2] explicitly set at least one of these flags true on" +
		" the command line")
	return false
}

func ProcessListFlags(o output.Bus, values map[string]*CommandFlag[any]) (*ListSettings, bool) {
	settings := &ListSettings{}
	flagsOk := true // optimistic
	var flagErr error
	if settings.Albums, flagErr = GetBool(o, values, ListAlbums); flagErr != nil {
		flagsOk = false
	}
	if settings.Annotate, flagErr = GetBool(o, values, ListAnnotate); flagErr != nil {
		flagsOk = false
	}
	if settings.Artists, flagErr = GetBool(o, values, ListArtists); flagErr != nil {
		flagsOk = false
	}
	if settings.Details, flagErr = GetBool(o, values, ListDetails); flagErr != nil {
		flagsOk = false
	}
	if settings.Diagnostic, flagErr = GetBool(o, values, ListDiagnostic); flagErr != nil {
		flagsOk = false
	}
	if settings.SortByNumber, flagErr = GetBool(o, values, ListSortByNumber); flagErr != nil {
		flagsOk = false
	}
	if settings.SortByTitle, flagErr = GetBool(o, values, ListSortByTitle); flagErr != nil {
		flagsOk = false
	}
	if settings.Tracks, flagErr = GetBool(o, values, ListTracks); flagErr != nil {
		flagsOk = false
	}
	return settings, flagsOk
}

func init() {
	RootCmd.AddCommand(ListCmd)
	addDefaults(ListFlags)
	c := getConfiguration()
	o := getBus()
	AddFlags(o, c, ListCmd.Flags(), ListFlags, SearchFlags)
}
