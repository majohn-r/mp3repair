package cmd

import (
	"fmt"
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"maps"
	"mp3repair/internal/files"
	"slices"
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
			"[" + listTracksFlag + "] [" + listAnnotateFlag + "] [" + listDiagnosticFlag + "] [" +
			listSortByNumberFlag + " | " + listSortByTitleFlag + "] " + searchUsage,
		DisableFlagsInUseLine: true,
		Short:                 "Lists mp3 files and containing album and artist directories",
		Long: fmt.Sprintf(
			"%q lists mp3 files and containing album and artist directories", listCommand),
		Example: listCommand + " " + listAnnotateFlag + "\n" +
			"  Annotate tracks with album and artist data and albums with artist data\n" +
			listCommand + " " + listDiagnosticFlag + "\n" +
			"  Include full listing of ID3V1 and ID3V2 tags for each track\n" +
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
	diagnostic   cmdtoolkit.CommandFlag[bool]
	sortByNumber cmdtoolkit.CommandFlag[bool]
	sortByTitle  cmdtoolkit.CommandFlag[bool]
	tracks       cmdtoolkit.CommandFlag[bool]
}

func (ls *listSettings) listArtists(
	o output.Bus,
	allArtists []*files.Artist,
	ss *searchSettings,
) (err *cmdtoolkit.ExitError) {
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
			m[a.Name()] = a
			names = append(names, a.Name())
		}
		sort.Strings(names)
		for _, s := range names {
			o.ConsolePrintf("Artist: %s\n", s)
			artist := m[s]
			if artist != nil {
				o.IncrementTab(2)
				ls.listAlbums(o, artist.Albums())
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
	ls.listAlbums(o, albums)
}

func (ls *listSettings) listAlbums(o output.Bus, albums []*files.Album) {
	if ls.albums.Value {
		files.SortAlbums(albums)
		for _, album := range albums {
			o.ConsolePrintf("Album: %s\n", ls.annotateAlbumName(album))
			o.IncrementTab(2)
			ls.listTracks(o, album.Tracks())
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
	ls.listTracks(o, tracks)
}

func (ls *listSettings) annotateAlbumName(album *files.Album) string {
	switch {
	case !ls.artists.Value && ls.annotate.Value:
		return strings.Join([]string{quote(album.Title()), "by",
			quote(album.RecordingArtistName())}, " ")
	default:
		return album.Title()
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
		numbers = append(numbers, track.Number())
		m[track.Number()] = track
	}
	sort.Ints(numbers)
	for _, n := range numbers {
		track := m[n]
		if track != nil {
			o.ConsolePrintf("%2d. %s\n", n, track.Name())
			o.IncrementTab(2)
			ls.listTrackDiagnostics(o, track)
			o.DecrementTab(2)
		}
	}
}

func (ls *listSettings) listTracksByName(o output.Bus, tracks []*files.Track) {
	files.SortTracks(tracks)
	for _, track := range tracks {
		o.ConsolePrintln(ls.annotateTrackName(track))
		o.IncrementTab(2)
		ls.listTrackDiagnostics(o, track)
		o.DecrementTab(2)
	}
}

func quote(s string) string {
	return fmt.Sprintf("%q", s)
}

func (ls *listSettings) annotateTrackName(track *files.Track) string {
	commonName := track.Name()
	if !ls.annotate.Value || ls.albums.Value {
		return commonName
	}
	trackNameParts := []string{quote(commonName), "on", quote(track.AlbumName())}
	if !ls.artists.Value {
		trackNameParts = append(trackNameParts, "by", quote(track.RecordingArtist()))
	}
	return strings.Join(trackNameParts, " ")
}

func (ls *listSettings) listTrackDiagnostics(o output.Bus, track *files.Track) {
	if ls.diagnostic.Value {
		tags, ID3V1readErr := track.ID3V1Diagnostics()
		showID3V1Diagnostics(o, track, tags, ID3V1readErr)
		info, ID3V2readErr := track.ID3V2Diagnostics()
		showID3V2Diagnostics(o, track, info, ID3V2readErr)
	}
}

func showID3V1Diagnostics(o output.Bus, track *files.Track, tags []string, readErr error) {
	if readErr != nil {
		track.ReportMetadataReadError(o, files.ID3V1, readErr.Error())
		return
	}
	o.ConsolePrintln("ID3V1 metadata")
	o.IncrementTab(2)
	sort.Strings(tags)
	for _, s := range tags {
		o.ConsolePrintln(s)
	}
	o.DecrementTab(2)
}

func showID3V2Diagnostics(o output.Bus, track *files.Track, info *files.ID3V2Info, readErr error) {
	if readErr != nil {
		track.ReportMetadataReadError(o, files.ID3V2, readErr.Error())
		return
	}
	if info != nil {
		o.ConsolePrintln("ID3V2 metadata")
		o.IncrementTab(2)
		o.ConsolePrintf("Version: %v\n", info.Version())
		o.ConsolePrintf("Encoding: %s\n", info.Encoding())
		tags := slices.Sorted(maps.Keys(info.Frames()))
		for _, tag := range tags {
			values := info.Frames()[tag]
			description := files.FrameDescription(tag)
			switch len(values) {
			case 0:
				o.ConsolePrintf("%s [%s]: <<empty>>\n", tag, description)
			case 1:
				o.ConsolePrintf("%s [%s]: %s\n", tag, description, values[0])
			default:
				header := fmt.Sprintf("%s [%s]: ", tag, description)
				o.ConsolePrintf("%s%s\n", header, values[0])
				tab := uint8(len(header))
				o.IncrementTab(tab)
				for _, value := range values[1:] {
					o.ConsolePrintln(value)
				}
				o.DecrementTab(tab)
			}
		}
		o.DecrementTab(2)
	}
}

func (ls *listSettings) tracksSortable(o output.Bus) bool {
	bothSortingOptionsSet := ls.sortByNumber.Value && ls.sortByTitle.Value
	neitherSortingOptionSet := !ls.sortByNumber.Value && !ls.sortByTitle.Value
	if ls.tracks.Value {
		switch {
		case bothSortingOptionsSet:
			o.ErrorPrintln("Track sorting cannot be done.")
			o.ErrorPrintln("Why?")
			switch ls.sortByNumber.UserSet {
			case true:
				switch ls.sortByTitle.UserSet {
				case true:
					o.ErrorPrintf("You explicitly set %s and %s true.\n", listSortByNumberFlag, listSortByTitleFlag)
				case false:
					o.ErrorPrintf(
						"The %s flag is configured true and you explicitly set %s true.\n",
						listSortByTitleFlag,
						listSortByNumberFlag,
					)
				}
			case false:
				switch ls.sortByTitle.UserSet {
				case true:
					o.ErrorPrintf(
						"The %s flag is configured true and you explicitly set %s true.\n",
						listSortByNumberFlag,
						listSortByTitleFlag,
					)
				case false:
					o.ErrorPrintf(
						"The %s and %s flags are both configured true.\n",
						listSortByNumberFlag,
						listSortByTitleFlag,
					)
				}
			}
			o.ErrorPrintln("What to do:")
			o.BeginErrorList(false)
			o.ErrorPrintln("Edit the configuration file and use its default values, or")
			o.ErrorPrintln("Use more appropriate command line values.")
			o.EndErrorList()
			return false
		case ls.sortByNumber.Value && !ls.albums.Value:
			o.ErrorPrintln("Sorting tracks by number not possible.")
			o.ErrorPrintln("Why?")
			o.ErrorPrintln("Track numbers are only relevant if albums are also output.")
			switch ls.sortByNumber.UserSet {
			case true:
				switch ls.albums.UserSet {
				case true:
					o.ErrorPrintf("You set %s true and %s false.\n", listSortByNumberFlag, listAlbumsFlag)
				case false:
					o.ErrorPrintf(
						"You set %s true and %s is configured as false.\n",
						listSortByNumberFlag,
						listAlbumsFlag,
					)
				}
			case false:
				switch ls.albums.UserSet {
				case true:
					o.ErrorPrintf(
						"You set %s false and %s is configured as true.\n",
						listAlbumsFlag,
						listSortByNumberFlag,
					)
				case false:
					o.ErrorPrintf(
						"%s is configured as false, and %s is configured as true.\n",
						listAlbumsFlag,
						listSortByNumberFlag,
					)
				}
			}
			o.ErrorPrintln("What to do:")
			o.BeginErrorList(false)
			o.ErrorPrintln("Edit the configuration file and use its default values, or")
			o.ErrorPrintln("Change which flags you set on the command line.")
			o.EndErrorList()
			return false
		case neitherSortingOptionSet:
			if ls.sortByNumber.UserSet && ls.sortByTitle.UserSet {
				o.ErrorPrintln("A listing of tracks is not possible.")
				o.ErrorPrintln("Why?")
				o.ErrorPrintf(
					"Tracks are enabled, but you set both %s and %s false.\n",
					listSortByNumberFlag,
					listSortByTitleFlag,
				)
				o.ErrorPrintln("What to do:")
				o.ErrorPrintln("Enable one of the sorting flags.")
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
		o.ErrorPrintln("Your sorting preferences are not relevant.")
		o.ErrorPrintln("Why?")
		o.ErrorPrintf(
			"Tracks are not included in the output, but you explicitly set %s or %s true.\n",
			listSortByNumberFlag,
			listSortByTitleFlag)
		o.ErrorPrintln("What to do:")
		o.BeginErrorList(false)
		o.ErrorPrintf("Set %s true, or\n", listTracksFlag)
		o.ErrorPrintln("Remove the sorting flags from the command line.")
		o.EndErrorList()
		return false
	}
	return true
}

func (ls *listSettings) hasWorkToDo(o output.Bus) bool {
	if ls.albums.Value || ls.artists.Value || ls.tracks.Value {
		return true
	}
	o.ErrorPrintln("No listing will be output.")
	o.ErrorPrintln("Why?")
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
			o.ErrorPrintf("You explicitly set %s, %s, and %s false.\n", listAlbumsFlag, listArtistsFlag, listTracksFlag)
		default:
			o.ErrorPrintf(
				"In addition to %s configured false, you explicitly set %s false.\n",
				strings.Join(flagsFromConfig, " and "),
				strings.Join(flagsUserSet, " and "),
			)
		}
	default: // user did not set any of the relevant flags that way
		o.ErrorPrintf(
			"The flags %s, %s, and %s are all configured false.\n",
			listAlbumsFlag,
			listArtistsFlag,
			listTracksFlag,
		)
	}
	o.ErrorPrintln("What to do:")
	o.ErrorPrintln("Either:")
	o.BeginErrorList(true)
	o.ErrorPrintln("Edit the configuration file so that at least one of these flags is true, or")
	o.ErrorPrintln("Explicitly set at least one of these flags true on the command line.")
	o.EndErrorList()
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
