/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
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
	ListFlags = NewSectionFlags().WithSectionName(ListCommand).WithFlags(
		map[string]*FlagDetails{
			ListAlbums: NewFlagDetails().WithAbbreviatedName("l").WithUsage(
				"include album names in listing").WithExpectedType(
				BoolType).WithDefaultValue(false),
			ListArtists: NewFlagDetails().WithAbbreviatedName("r").WithUsage(
				"include artist names in listing").WithExpectedType(
				BoolType).WithDefaultValue(false),
			ListTracks: NewFlagDetails().WithAbbreviatedName("t").WithUsage(
				"include track names in listing").WithExpectedType(
				BoolType).WithDefaultValue(false),
			ListSortByNumber: NewFlagDetails().WithUsage(
				"sort tracks by track number").WithExpectedType(BoolType).WithDefaultValue(
				false),
			ListSortByTitle: NewFlagDetails().WithUsage(
				"sort tracks by track title").WithExpectedType(BoolType).WithDefaultValue(
				false),
			ListAnnotate: NewFlagDetails().WithUsage(
				"annotate listings with album and artist names").WithExpectedType(
				BoolType).WithDefaultValue(false),
			ListDetails: NewFlagDetails().WithUsage(
				"include details with tracks").WithExpectedType(BoolType).WithDefaultValue(
				false),
			ListDiagnostic: NewFlagDetails().WithUsage(
				"include diagnostic information with tracks").WithExpectedType(
				BoolType).WithDefaultValue(false),
		},
	)
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
				ListAlbumsFlag:       ls.albums,
				"albums-user-set":    ls.albumsUserSet,
				ListAnnotateFlag:     ls.annotate,
				ListArtistsFlag:      ls.artists,
				"artists-user-set":   ls.artistsUserSet,
				ListSortByNumberFlag: ls.sortByNumber,
				"byNumber-user-set":  ls.sortByNumberUserSet,
				ListSortByTitleFlag:  ls.sortByTitle,
				"byTitle-user-set":   ls.sortByTitleUserSet,
				ListDetailsFlag:      ls.details,
				ListDiagnosticFlag:   ls.diagnostic,
				ListTracksFlag:       ls.tracks,
				"tracks-user-set":    ls.tracksUserSet,
			}
			for k, v := range searchSettings.Values() {
				details[k] = v
			}
			LogCommandStart(o, ListCommand, details)
			switch ls.HasWorkToDo(o) {
			case true:
				switch ls.TracksSortable(o) {
				case true:
					allArtists, loaded := searchSettings.Load(o)
					exitError = ls.ProcessArtists(o, allArtists, loaded, searchSettings)
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
	albums              bool
	albumsUserSet       bool
	annotate            bool
	artists             bool
	artistsUserSet      bool
	details             bool
	diagnostic          bool
	sortByNumber        bool
	sortByNumberUserSet bool
	sortByTitle         bool
	sortByTitleUserSet  bool
	tracks              bool
	tracksUserSet       bool
}

func NewListSettings() *ListSettings {
	return &ListSettings{}
}

func (ls *ListSettings) WithAlbums(b bool) *ListSettings {
	ls.albums = b
	return ls
}

func (ls *ListSettings) WithAlbumsUserSet(b bool) *ListSettings {
	ls.albumsUserSet = b
	return ls
}

func (ls *ListSettings) WithAnnotate(b bool) *ListSettings {
	ls.annotate = b
	return ls
}

func (ls *ListSettings) WithArtists(b bool) *ListSettings {
	ls.artists = b
	return ls
}

func (ls *ListSettings) WithArtistsUserSet(b bool) *ListSettings {
	ls.artistsUserSet = b
	return ls
}

func (ls *ListSettings) WithDetails(b bool) *ListSettings {
	ls.details = b
	return ls
}

func (ls *ListSettings) WithDiagnostic(b bool) *ListSettings {
	ls.diagnostic = b
	return ls
}

func (ls *ListSettings) WithSortByNumber(b bool) *ListSettings {
	ls.sortByNumber = b
	return ls
}

func (ls *ListSettings) WithSortByNumberUserSet(b bool) *ListSettings {
	ls.sortByNumberUserSet = b
	return ls
}

func (ls *ListSettings) WithSortByTitle(b bool) *ListSettings {
	ls.sortByTitle = b
	return ls
}

func (ls *ListSettings) WithSortByTitleUserSet(b bool) *ListSettings {
	ls.sortByTitleUserSet = b
	return ls
}

func (ls *ListSettings) WithTracks(b bool) *ListSettings {
	ls.tracks = b
	return ls
}

func (ls *ListSettings) WithTracksUserSet(b bool) *ListSettings {
	ls.tracksUserSet = b
	return ls
}

// TODO: better name: ListArtists
func (ls *ListSettings) ProcessArtists(o output.Bus, allArtists []*files.Artist,
	loaded bool, searchSettings *SearchSettings) (err *ExitError) {
	err = NewExitUserError(ListCommand)
	if loaded {
		if filteredArtists, filtered := searchSettings.Filter(o, allArtists); filtered {
			ls.ListArtists(o, filteredArtists)
			err = nil
		}
	}
	return err
}

// TODO: better name: ListFilteredArtists
func (ls *ListSettings) ListArtists(o output.Bus, artists []*files.Artist) {
	if ls.artists {
		m := map[string]*files.Artist{}
		names := make([]string, 0, len(artists))
		for _, a := range artists {
			m[a.Name()] = a
			names = append(names, a.Name())
		}
		sort.Strings(names)
		for _, s := range names {
			o.WriteConsole("Artist: %s\n", s)
			artist := m[s]
			if artist != nil {
				o.IncrementTab(2)
				ls.ListAlbums(o, artist.Albums())
				o.DecrementTab(2)
			}
		}
		return
	}
	albumCount := 0
	for _, a := range artists {
		albumCount += len(a.Albums())
	}
	albums := make([]*files.Album, 0, albumCount)
	for _, a := range artists {
		albums = append(albums, a.Albums()...)
	}
	ls.ListAlbums(o, albums)
}

type AlbumSlice []*files.Album

func (as AlbumSlice) Len() int {
	return len(as)
}

func (as AlbumSlice) Less(i, j int) bool {
	if as[i].Name() == as[j].Name() {
		return as[i].RecordingArtistName() < as[j].RecordingArtistName()
	}
	return as[i].Name() < as[j].Name()
}

func (as AlbumSlice) Swap(i, j int) {
	as[i], as[j] = as[j], as[i]
}

func (ls *ListSettings) ListAlbums(o output.Bus, albums []*files.Album) {
	if ls.albums {
		sort.Sort(AlbumSlice(albums))
		for _, album := range albums {
			o.WriteConsole("Album: %s\n", ls.AnnotateAlbumName(album))
			o.IncrementTab(2)
			ls.ListTracks(o, album.Tracks())
			o.DecrementTab(2)
		}
		return
	}
	trackCount := 0
	for _, album := range albums {
		trackCount += len(album.Tracks())
	}
	tracks := make([]*files.Track, 0, trackCount)
	for _, album := range albums {
		tracks = append(tracks, album.Tracks()...)
	}
	ls.ListTracks(o, tracks)
}

func (ls *ListSettings) AnnotateAlbumName(album *files.Album) string {
	switch {
	case !ls.artists && ls.annotate:
		return strings.Join([]string{quote(album.Name()), "by",
			quote(album.RecordingArtistName())}, " ")
	default:
		return album.Name()
	}
}

func (ls *ListSettings) ListTracks(o output.Bus, tracks []*files.Track) {
	if !ls.tracks {
		return
	}
	if ls.sortByNumber {
		ls.ListTracksByNumber(o, tracks)
		return
	}
	if ls.sortByTitle {
		ls.ListTracksByName(o, tracks)
	}
}

func (ls *ListSettings) ListTracksByNumber(o output.Bus, tracks []*files.Track) {
	m := map[int]*files.Track{}
	numbers := make([]int, 0, len(tracks))
	for _, track := range tracks {
		numbers = append(numbers, track.Number())
		m[track.Number()] = track
	}
	sort.Ints(numbers)
	for _, n := range numbers {
		track := m[n]
		if track != nil {
			o.WriteConsole("%2d. %s\n", n, track.CommonName())
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
	if ts[i].CommonName() == ts[j].CommonName() {
		album1 := ts[i].Album()
		album2 := ts[j].Album()
		if album1.Name() == album2.Name() {
			return album1.RecordingArtistName() < album2.RecordingArtistName()
		}
		return album1.Name() < album2.Name()
	}
	return ts[i].CommonName() < ts[j].CommonName()
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
	commonName := track.CommonName()
	if !ls.annotate || ls.albums {
		return commonName
	}
	trackNameParts := []string{quote(commonName), "on", quote(track.AlbumName())}
	if !ls.artists {
		trackNameParts = append(trackNameParts, "by", quote(track.RecordingArtist()))
	}
	return strings.Join(trackNameParts, " ")
}

func (ls *ListSettings) ListTrackDetails(o output.Bus, track *files.Track) {
	if ls.details {
		// go get information from track and display it
		m, readErr := track.Details()
		ShowDetails(o, track, m, readErr)
	}
}

// split out for testing!
// TODO: put values in a struct
func ShowDetails(o output.Bus, track *files.Track, details map[string]string, detailsError error) {
	if detailsError != nil {
		o.Log(output.Error, "cannot get details", map[string]any{
			"error": detailsError,
			"track": track.String(),
		})
		o.WriteCanonicalError(
			"The details are not available for track %q on album %q by artist %q: %q",
			track.CommonName(), track.AlbumName(), track.RecordingArtist(),
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
	if ls.diagnostic {
		version, encoding, frames, ID3V2readErr := track.ID3V2Diagnostics()
		ShowID3V2Diagnostics(o, track, version, encoding, frames, ID3V2readErr)
		tags, ID3V1readErr := track.ID3V1Diagnostics()
		ShowID3V1Diagnostics(o, track, tags, ID3V1readErr)
	}
}

// split out for testing!
// TODO: put values in a struct
func ShowID3V1Diagnostics(o output.Bus, track *files.Track, tags []string, readErr error) {
	if readErr != nil {
		track.ReportMetadataReadError(o, files.ID3V1, readErr.Error())
		return
	}
	for _, s := range tags {
		o.WriteConsole("ID3V1 %s\n", s)
	}
}

// split out for testing!
// TODO: put values in a struct
func ShowID3V2Diagnostics(o output.Bus, track *files.Track, version byte, encoding string, frames []string, readErr error) {
	if readErr != nil {
		track.ReportMetadataReadError(o, files.ID3V2, readErr.Error())
		return
	}
	o.WriteConsole("ID3V2 Version: %v\n", version)
	o.WriteConsole("ID3V2 Encoding: %q\n", encoding)
	for _, frame := range frames {
		o.WriteConsole("ID3V2 %s\n", frame)
	}
}

func (ls *ListSettings) TracksSortable(o output.Bus) bool {
	bothSortingOptionsSet := ls.sortByNumber && ls.sortByTitle
	neitherSortingOptionSet := !ls.sortByNumber && !ls.sortByTitle
	if ls.tracks {
		switch {
		case bothSortingOptionsSet:
			o.WriteCanonicalError("Track sorting cannot be done")
			o.WriteCanonicalError("Why?")
			switch ls.sortByNumberUserSet {
			case true:
				switch ls.sortByTitleUserSet {
				case true:
					o.WriteCanonicalError("You explicitly set %s and %s true",
						ListSortByNumberFlag, ListSortByTitleFlag)
				case false:
					o.WriteCanonicalError(
						"The %s flag is configured true and you explicitly set %s true",
						ListSortByTitleFlag, ListSortByNumberFlag)
				}
			case false:
				switch ls.sortByTitleUserSet {
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
		case ls.sortByNumber && !ls.albums:
			o.WriteCanonicalError("Sorting tracks by number not possible.")
			o.WriteCanonicalError("Why?")
			o.WriteCanonicalError("Track numbers are only relevant if albums are also output.")
			switch ls.sortByNumberUserSet {
			case true:
				switch ls.albumsUserSet {
				case true:
					o.WriteCanonicalError("You set %s true and %s false.", ListSortByNumberFlag, ListAlbumsFlag)
				case false:
					o.WriteCanonicalError("You set %s true and %s is configured as false", ListSortByNumberFlag, ListAlbumsFlag)
				}
			case false:
				switch ls.albumsUserSet {
				case true:
					o.WriteCanonicalError("You set %s false and %s is configured as true", ListAlbumsFlag, ListSortByNumberFlag)
				case false:
					o.WriteCanonicalError("%s is configured as false, and %s is configured as true", ListAlbumsFlag, ListSortByNumberFlag)
				}
			}
			o.WriteCanonicalError("What to do:\nEither edit the configuration file or change which flags you set on the command line.")
			return false
		case neitherSortingOptionSet:
			if ls.sortByNumberUserSet && ls.sortByTitleUserSet {
				o.WriteCanonicalError("A listing of tracks is not possible.")
				o.WriteCanonicalError("Why?")
				o.WriteCanonicalError("Tracks are enabled, but you set both %s and %s false", ListSortByNumberFlag, ListSortByTitleFlag)
				o.WriteCanonicalError("What to do:\nEnable one of the sorting flags")
				return false
			}
			// pick a sensible option
			switch {
			case ls.sortByNumberUserSet:
				ls.sortByTitle = true // pick the other setting
			case ls.sortByTitleUserSet:
				ls.sortByNumber = true // pick the other setting
			default: // ok, pick something sensible, user does not care
				switch ls.albums {
				case true:
					ls.sortByNumber = true
				case false:
					ls.sortByTitle = true
				}
			}
			o.Log(output.Info, "no track sorting set, providing a sensible value", map[string]any{
				ListAlbumsFlag:      ls.albums,
				ListSortByNumber:    ls.sortByNumber,
				ListSortByTitleFlag: ls.sortByTitle,
			})
			return true
		default: // https://github.com/majohn-r/mp3repair/issues/170
			return true
		}
	}
	if (ls.sortByNumber && ls.sortByNumberUserSet) || (ls.sortByTitle && ls.sortByTitleUserSet) {
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
	if ls.albums || ls.artists || ls.tracks {
		return true
	}
	o.WriteCanonicalError("No listing will be output.\nWhy?\n")
	switch {
	case ls.albumsUserSet || ls.artistsUserSet || ls.tracksUserSet:
		flagsUserSet := make([]string, 0, 3)
		flagsFromConfig := make([]string, 0, 3)
		switch {
		case ls.albumsUserSet:
			flagsUserSet = append(flagsUserSet, ListAlbumsFlag)
		default:
			flagsFromConfig = append(flagsFromConfig, ListAlbumsFlag)
		}
		switch {
		case ls.artistsUserSet:
			flagsUserSet = append(flagsUserSet, ListArtistsFlag)
		default:
			flagsFromConfig = append(flagsFromConfig, ListArtistsFlag)
		}
		switch {
		case ls.tracksUserSet:
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

func ProcessListFlags(o output.Bus, values map[string]*FlagValue) (*ListSettings, bool) {
	settings := &ListSettings{}
	flagsOk := true // optimistic
	var flagErr error
	if settings.albums, settings.albumsUserSet, flagErr = GetBool(o, values, ListAlbums); flagErr != nil {
		flagsOk = false
	}
	if settings.annotate, _, flagErr = GetBool(o, values, ListAnnotate); flagErr != nil {
		flagsOk = false
	}
	if settings.artists, settings.artistsUserSet, flagErr = GetBool(o, values, ListArtists); flagErr != nil {
		flagsOk = false
	}
	if settings.details, _, flagErr = GetBool(o, values, ListDetails); flagErr != nil {
		flagsOk = false
	}
	if settings.diagnostic, _, flagErr = GetBool(o, values, ListDiagnostic); flagErr != nil {
		flagsOk = false
	}
	if settings.sortByNumber, settings.sortByNumberUserSet, flagErr = GetBool(o, values, ListSortByNumber); flagErr != nil {
		flagsOk = false
	}
	if settings.sortByTitle, settings.sortByTitleUserSet, flagErr = GetBool(o, values, ListSortByTitle); flagErr != nil {
		flagsOk = false
	}
	if settings.tracks, settings.tracksUserSet, flagErr = GetBool(o, values, ListTracks); flagErr != nil {
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
