package cmd

import (
	"fmt"
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"mp3repair/internal/files"
	"sort"
	"strings"

	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

const (
	listAlbums           = "albums"
	listAlbumsFlag       = "--" + listAlbums
	listAnnotate         = "annotate"
	listAnnotateFlag     = "--" + listAnnotate
	listArtists          = "artists"
	listArtistsFlag      = "--" + listArtists
	listCommand          = "list"
	listDetails          = "details"
	listDetailsFlag      = "--" + listDetails
	listDiagnostic       = "diagnostic"
	listDiagnosticFlag   = "--" + listDiagnostic
	listSortByNumber     = "byNumber"
	listSortByNumberFlag = "--" + listSortByNumber
	listSortByTitle      = "byTitle"
	listSortByTitleFlag  = "--" + listSortByTitle
	listTracks           = "tracks"
	listTracksFlag       = "--" + listTracks
)

var (
	listCmd = &cobra.Command{
		Use: listCommand + " [" + listAlbumsFlag + "] [" + listArtistsFlag + "] " +
			"[" + listTracksFlag + "] [" + listAnnotateFlag + "] [" + listDetailsFlag + "] " +
			"[" + listDiagnosticFlag + "] [" + listSortByNumberFlag + " | " +
			listSortByTitleFlag + "] " + searchUsage,
		DisableFlagsInUseLine: true,
		Short:                 "Lists mp3 files and containing album and artist directories",
		Long: fmt.Sprintf(
			"%q lists mp3 files and containing album and artist directories", listCommand),
		Example: listCommand + " " + listAnnotateFlag + "\n" +
			"  Annotate tracks with album and artist data and albums with artist data\n" +
			listCommand + " " + listDetailsFlag + "\n" +
			"  Include detailed information, if available, for each track. This includes" +
			" composer,\n" +
			"  conductor, key, lyricist, orchestra/band, and subtitle\n" +
			listCommand + " " + listAlbumsFlag + "\n" +
			"  Include the album names in the output\n" +
			listCommand + " " + listArtistsFlag + "\n" +
			"  Include the artist names in the output\n" +
			listCommand + " " + listTracksFlag + "\n" +
			"  Include the track names in the output\n" +
			listCommand + " " + listSortByTitleFlag + "\n" +
			"  Sort tracks by name, ignoring track numbers\n" +
			listCommand + " " + listSortByNumberFlag + "\n" +
			"  Sort tracks by track number",
		RunE: listRun,
	}
	listFlags = &cmdtoolkit.FlagSet{
		Name: listCommand,
		Details: map[string]*cmdtoolkit.FlagDetails{
			listAlbums: {
				AbbreviatedName: "l",
				Usage:           "include album names in listing",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
			listArtists: {
				AbbreviatedName: "r",
				Usage:           "include artist names in listing",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
			listTracks: {
				AbbreviatedName: "t",
				Usage:           "include track names in listing",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
			listSortByNumber: {
				Usage:        "sort tracks by track number",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: false,
			},
			listSortByTitle: {
				Usage:        "sort tracks by track title",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: false,
			},
			listAnnotate: {
				Usage:        "annotate listings with album and artist names",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: false,
			},
			listDetails: {
				Usage:        "include details with tracks",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: false,
			},
			listDiagnostic: {
				Usage:        "include diagnostic information with tracks",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: false,
			},
		},
	}
)

func listRun(cmd *cobra.Command, _ []string) error {
	exitError := cmdtoolkit.NewExitProgrammingError(listCommand)
	o := getBus()
	producer := cmd.Flags()
	values, eSlice := cmdtoolkit.ReadFlags(producer, listFlags)
	searchSettings, searchFlagsOk := evaluateSearchFlags(o, producer)
	if cmdtoolkit.ProcessFlagErrors(o, eSlice) && searchFlagsOk {
		if ls, flagsOk := processListFlags(o, values); flagsOk {
			switch ls.hasWorkToDo(o) {
			case true:
				switch ls.tracksSortable(o) {
				case true:
					allArtists := searchSettings.load(o)
					exitError = ls.listArtists(o, allArtists, searchSettings)
				case false:
					exitError = cmdtoolkit.NewExitUserError(listCommand)
				}
			case false:
				exitError = cmdtoolkit.NewExitUserError(listCommand)
			}
		}
	}
	return cmdtoolkit.ToErrorInterface(exitError)
}

type listSettings struct {
	albums       cmdtoolkit.CommandFlag[bool]
	annotate     cmdtoolkit.CommandFlag[bool]
	artists      cmdtoolkit.CommandFlag[bool]
	details      cmdtoolkit.CommandFlag[bool]
	diagnostic   cmdtoolkit.CommandFlag[bool]
	sortByNumber cmdtoolkit.CommandFlag[bool]
	sortByTitle  cmdtoolkit.CommandFlag[bool]
	tracks       cmdtoolkit.CommandFlag[bool]
}

func (ls *listSettings) listArtists(o output.Bus, allArtists []*files.Artist, ss *searchSettings) (err *cmdtoolkit.ExitError) {
	err = cmdtoolkit.NewExitUserError(listCommand)
	if len(allArtists) != 0 {
		if filteredArtists := ss.filter(o, allArtists); len(filteredArtists) != 0 {
			ls.listFilteredArtists(o, filteredArtists)
			err = nil
		}
	}
	return err
}

func (ls *listSettings) listFilteredArtists(o output.Bus, artists []*files.Artist) {
	if ls.artists.Value {
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
				ls.listAlbums(o, artist.Albums)
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
	ls.listAlbums(o, albums)
}

type albumSlice []*files.Album

func (as albumSlice) Len() int {
	return len(as)
}

func (as albumSlice) Less(i, j int) bool {
	if as[i].Title == as[j].Title {
		return as[i].RecordingArtistName() < as[j].RecordingArtistName()
	}
	return as[i].Title < as[j].Title
}

func (as albumSlice) Swap(i, j int) {
	as[i], as[j] = as[j], as[i]
}

func (ls *listSettings) listAlbums(o output.Bus, albums []*files.Album) {
	if ls.albums.Value {
		sort.Sort(albumSlice(albums))
		for _, album := range albums {
			o.WriteConsole("Album: %s\n", ls.annotateAlbumName(album))
			o.IncrementTab(2)
			ls.listTracks(o, album.Tracks)
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
	ls.listTracks(o, tracks)
}

func (ls *listSettings) annotateAlbumName(album *files.Album) string {
	switch {
	case !ls.artists.Value && ls.annotate.Value:
		return strings.Join([]string{quote(album.Title), "by",
			quote(album.RecordingArtistName())}, " ")
	default:
		return album.Title
	}
}

func (ls *listSettings) listTracks(o output.Bus, tracks []*files.Track) {
	if !ls.tracks.Value {
		return
	}
	if ls.sortByNumber.Value {
		ls.listTracksByNumber(o, tracks)
		return
	}
	if ls.sortByTitle.Value {
		ls.listTracksByName(o, tracks)
	}
}

func (ls *listSettings) listTracksByNumber(o output.Bus, tracks []*files.Track) {
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
			ls.listTrackDetails(o, track)
			ls.listTrackDiagnostics(o, track)
			o.DecrementTab(2)
		}
	}
}

type trackSlice []*files.Track

func (ts trackSlice) Len() int {
	return len(ts)
}

func (ts trackSlice) Less(i, j int) bool {
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

func (ts trackSlice) Swap(i, j int) {
	ts[i], ts[j] = ts[j], ts[i]
}

func (ls *listSettings) listTracksByName(o output.Bus, tracks []*files.Track) {
	sort.Sort(trackSlice(tracks))
	for _, track := range tracks {
		o.WriteConsole("%s\n", ls.annotateTrackName(track))
		o.IncrementTab(2)
		ls.listTrackDetails(o, track)
		ls.listTrackDiagnostics(o, track)
		o.DecrementTab(2)
	}
}

func quote(s string) string {
	return fmt.Sprintf("%q", s)
}

func (ls *listSettings) annotateTrackName(track *files.Track) string {
	commonName := track.SimpleName
	if !ls.annotate.Value || ls.albums.Value {
		return commonName
	}
	trackNameParts := []string{quote(commonName), "on", quote(track.AlbumName())}
	if !ls.artists.Value {
		trackNameParts = append(trackNameParts, "by", quote(track.RecordingArtist()))
	}
	return strings.Join(trackNameParts, " ")
}

func (ls *listSettings) listTrackDetails(o output.Bus, track *files.Track) {
	if ls.details.Value {
		// go get information from track and display it
		m, readErr := track.Details()
		showDetails(o, track, m, readErr)
	}
}

func showDetails(o output.Bus, track *files.Track, details map[string]string, detailsError error) {
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

func (ls *listSettings) listTrackDiagnostics(o output.Bus, track *files.Track) {
	if ls.diagnostic.Value {
		info, ID3V2readErr := track.ID3V2Diagnostics()
		showID3V2Diagnostics(o, track, info, ID3V2readErr)
		tags, ID3V1readErr := track.ID3V1Diagnostics()
		showID3V1Diagnostics(o, track, tags, ID3V1readErr)
	}
}

func showID3V1Diagnostics(o output.Bus, track *files.Track, tags []string, readErr error) {
	if readErr != nil {
		track.ReportMetadataReadError(o, files.ID3V1, readErr.Error())
		return
	}
	for _, s := range tags {
		o.WriteConsole("ID3V1 %s\n", s)
	}
}

func showID3V2Diagnostics(o output.Bus, track *files.Track, info *files.ID3V2Info, readErr error) {
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

func (ls *listSettings) tracksSortable(o output.Bus) bool {
	bothSortingOptionsSet := ls.sortByNumber.Value && ls.sortByTitle.Value
	neitherSortingOptionSet := !ls.sortByNumber.Value && !ls.sortByTitle.Value
	if ls.tracks.Value {
		switch {
		case bothSortingOptionsSet:
			o.WriteCanonicalError("Track sorting cannot be done")
			o.WriteCanonicalError("Why?")
			switch ls.sortByNumber.UserSet {
			case true:
				switch ls.sortByTitle.UserSet {
				case true:
					o.WriteCanonicalError("You explicitly set %s and %s true",
						listSortByNumberFlag, listSortByTitleFlag)
				case false:
					o.WriteCanonicalError(
						"The %s flag is configured true and you explicitly set %s true",
						listSortByTitleFlag, listSortByNumberFlag)
				}
			case false:
				switch ls.sortByTitle.UserSet {
				case true:
					o.WriteCanonicalError(
						"The %s flag is configured true and you explicitly set %s true",
						listSortByNumberFlag, listSortByTitleFlag)
				case false:
					o.WriteCanonicalError("The %s and %s flags are both configured true",
						listSortByNumberFlag, listSortByTitleFlag)
				}
			}
			o.WriteCanonicalError("What to do:\nEither edit the configuration file and use" +
				" those default values, or use appropriate command line values")
			return false
		case ls.sortByNumber.Value && !ls.albums.Value:
			o.WriteCanonicalError("Sorting tracks by number not possible.")
			o.WriteCanonicalError("Why?")
			o.WriteCanonicalError("Track numbers are only relevant if albums are also output.")
			switch ls.sortByNumber.UserSet {
			case true:
				switch ls.albums.UserSet {
				case true:
					o.WriteCanonicalError("You set %s true and %s false.", listSortByNumberFlag, listAlbumsFlag)
				case false:
					o.WriteCanonicalError("You set %s true and %s is configured as false", listSortByNumberFlag, listAlbumsFlag)
				}
			case false:
				switch ls.albums.UserSet {
				case true:
					o.WriteCanonicalError("You set %s false and %s is configured as true", listAlbumsFlag, listSortByNumberFlag)
				case false:
					o.WriteCanonicalError("%s is configured as false, and %s is configured as true", listAlbumsFlag, listSortByNumberFlag)
				}
			}
			o.WriteCanonicalError("What to do:\nEither edit the configuration file or change which flags you set on the command line.")
			return false
		case neitherSortingOptionSet:
			if ls.sortByNumber.UserSet && ls.sortByTitle.UserSet {
				o.WriteCanonicalError("A listing of tracks is not possible.")
				o.WriteCanonicalError("Why?")
				o.WriteCanonicalError("Tracks are enabled, but you set both %s and %s false", listSortByNumberFlag, listSortByTitleFlag)
				o.WriteCanonicalError("What to do:\nEnable one of the sorting flags")
				return false
			}
			// pick a sensible option
			switch {
			case ls.sortByNumber.UserSet:
				ls.sortByTitle.Value = true // pick the other setting
			case ls.sortByTitle.UserSet:
				ls.sortByNumber.Value = true // pick the other setting
			default: // ok, pick something sensible, user does not care
				switch ls.albums.Value {
				case true:
					ls.sortByNumber.Value = true
				case false:
					ls.sortByTitle.Value = true
				}
			}
			o.Log(output.Info, "no track sorting set, providing a sensible value", map[string]any{
				listAlbumsFlag:      ls.albums.Value,
				listSortByNumber:    ls.sortByNumber.Value,
				listSortByTitleFlag: ls.sortByTitle.Value,
			})
			return true
		default: // https://github.com/majohn-r/mp3repair/issues/170
			return true
		}
	}
	if (ls.sortByNumber.Value && ls.sortByNumber.UserSet) || (ls.sortByTitle.Value && ls.sortByTitle.UserSet) {
		o.WriteCanonicalError("Your sorting preferences are not relevant")
		o.WriteCanonicalError("Why?")
		o.WriteCanonicalError(
			"Tracks are not included in the output, but you explicitly set %s or %s true.",
			listSortByNumberFlag, listSortByTitleFlag)
		o.WriteCanonicalError("What to do:\nEither set %s true or remove the sorting flags"+
			" from the command line.", listTracksFlag)
		return false
	}
	return true
}

func (ls *listSettings) hasWorkToDo(o output.Bus) bool {
	if ls.albums.Value || ls.artists.Value || ls.tracks.Value {
		return true
	}
	o.WriteCanonicalError("No listing will be output.\nWhy?\n")
	switch {
	case ls.albums.UserSet || ls.artists.UserSet || ls.tracks.UserSet:
		flagsUserSet := make([]string, 0, 3)
		flagsFromConfig := make([]string, 0, 3)
		switch {
		case ls.albums.UserSet:
			flagsUserSet = append(flagsUserSet, listAlbumsFlag)
		default:
			flagsFromConfig = append(flagsFromConfig, listAlbumsFlag)
		}
		switch {
		case ls.artists.UserSet:
			flagsUserSet = append(flagsUserSet, listArtistsFlag)
		default:
			flagsFromConfig = append(flagsFromConfig, listArtistsFlag)
		}
		switch {
		case ls.tracks.UserSet:
			flagsUserSet = append(flagsUserSet, listTracksFlag)
		default:
			flagsFromConfig = append(flagsFromConfig, listTracksFlag)
		}
		switch len(flagsFromConfig) {
		case 0:
			o.WriteCanonicalError("You explicitly set %s, %s, and %s false",
				listAlbumsFlag, listArtistsFlag, listTracksFlag)
		default:
			o.WriteCanonicalError(
				"In addition to %s configured false, you explicitly set %s false",
				strings.Join(flagsFromConfig, " and "), strings.Join(flagsUserSet, " and "))
		}
	default: // user did not set any of the relevant flags that way
		o.WriteCanonicalError("The flags %s, %s, and %s are all configured false",
			listAlbumsFlag, listArtistsFlag, listTracksFlag)
	}
	o.WriteError("What to do:\n")
	o.WriteCanonicalError("Either:\n[1] Edit the configuration file so that at least one" +
		" of these flags is true, or\n[2] explicitly set at least one of these flags true on" +
		" the command line")
	return false
}

func processListFlags(o output.Bus, values map[string]*cmdtoolkit.CommandFlag[any]) (*listSettings, bool) {
	settings := &listSettings{}
	flagsOk := true // optimistic
	var flagErr error
	if settings.albums, flagErr = cmdtoolkit.GetBool(o, values, listAlbums); flagErr != nil {
		flagsOk = false
	}
	if settings.annotate, flagErr = cmdtoolkit.GetBool(o, values, listAnnotate); flagErr != nil {
		flagsOk = false
	}
	if settings.artists, flagErr = cmdtoolkit.GetBool(o, values, listArtists); flagErr != nil {
		flagsOk = false
	}
	if settings.details, flagErr = cmdtoolkit.GetBool(o, values, listDetails); flagErr != nil {
		flagsOk = false
	}
	if settings.diagnostic, flagErr = cmdtoolkit.GetBool(o, values, listDiagnostic); flagErr != nil {
		flagsOk = false
	}
	if settings.sortByNumber, flagErr = cmdtoolkit.GetBool(o, values, listSortByNumber); flagErr != nil {
		flagsOk = false
	}
	if settings.sortByTitle, flagErr = cmdtoolkit.GetBool(o, values, listSortByTitle); flagErr != nil {
		flagsOk = false
	}
	if settings.tracks, flagErr = cmdtoolkit.GetBool(o, values, listTracks); flagErr != nil {
		flagsOk = false
	}
	return settings, flagsOk
}

func init() {
	rootCmd.AddCommand(listCmd)
	cmdtoolkit.AddDefaults(listFlags)
	cmdtoolkit.AddFlags(getBus(), getConfiguration(), listCmd.Flags(), listFlags, searchFlags)
}
