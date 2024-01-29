/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd

import (
	"fmt"
	"mp3/internal/files"
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
		Use:                   ListCommand + " [" + ListAlbumsFlag + "] [" + ListArtistsFlag + "] [" + ListTracksFlag + "] [" + ListAnnotateFlag + "] [" + ListDetailsFlag + "] [" + ListDiagnosticFlag + "] [" + ListSortByNumberFlag + " | " + ListSortByTitleFlag + "] " + searchUsage,
		DisableFlagsInUseLine: true,
		Short:                 "Lists mp3 files and containing album and artist directories",
		Long:                  fmt.Sprintf("%q lists mp3 files and containing album and artist directories", ListCommand),
		Example: ListCommand + " " + ListAnnotateFlag + "\n" +
			"  Annotate tracks with album and artist data and albums with artist data\n" +
			ListCommand + " " + ListDetailsFlag + "\n" +
			"  Include detailed information, if available, for each track. This includes composer,\n" +
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
		Run: ListRun,
	}
	ListFlags = SectionFlags{
		SectionName: ListCommand,
		Flags: map[string]*FlagDetails{
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

func ListRun(cmd *cobra.Command, _ []string) {
	o := getBus()
	producer := cmd.Flags()
	values, eSlice := ReadFlags(producer, ListFlags)
	searchSettings, searchFlagsOk := EvaluateSearchFlags(o, producer)
	if ProcessFlagErrors(o, eSlice) && searchFlagsOk {
		if ls, ok := ProcessListFlags(o, values); ok {
			LogCommandStart(o, ListCommand, map[string]any{
				ListAlbumsFlag:         ls.Albums,
				"albums-user-set":      ls.AlbumsUserSet,
				ListAnnotateFlag:       ls.Annotate,
				ListArtistsFlag:        ls.Artists,
				"artists-user-set":     ls.ArtistsUserSet,
				ListSortByNumberFlag:   ls.SortByNumber,
				"byNumber-user-set":    ls.SortByNumberUserSet,
				ListSortByTitleFlag:    ls.SortByTitle,
				"byTitle-user-set":     ls.SortByTitleUserSet,
				ListDetailsFlag:        ls.Details,
				ListDiagnosticFlag:     ls.Diagnostic,
				ListTracksFlag:         ls.Tracks,
				"tracks-user-set":      ls.TracksUserSet,
				SearchAlbumFilterFlag:  searchSettings.AlbumFilter,
				SearchArtistFilterFlag: searchSettings.ArtistFilter,
				SearchTrackFilterFlag:  searchSettings.TrackFilter,
				SearchTopDirFlag:       searchSettings.TopDirectory,
			})
			if ls.HasWorkToDo(o) {
				if ls.TracksSortable(o) {
					allArtists, loaded := searchSettings.Load(o)
					ls.ProcessArtists(o, allArtists, loaded, searchSettings)
				}
			}
		}
	}
}

type ListSettings struct {
	Albums              bool
	AlbumsUserSet       bool
	Annotate            bool
	Artists             bool
	ArtistsUserSet      bool
	SortByNumber        bool
	SortByNumberUserSet bool
	SortByTitle         bool
	SortByTitleUserSet  bool
	Details             bool
	Diagnostic          bool
	Tracks              bool
	TracksUserSet       bool
}

func (ls *ListSettings) ProcessArtists(o output.Bus, allArtists []*files.Artist, loaded bool, searchSettings *SearchSettings) {
	if loaded {
		if filteredArtists, filtered := searchSettings.Filter(o, allArtists); filtered {
			ls.ListArtists(o, filteredArtists)
		}
	}
}

func (ls *ListSettings) ListArtists(o output.Bus, artists []*files.Artist) {
	if ls.Artists {
		m := map[string]*files.Artist{}
		names := []string{}
		for _, a := range artists {
			m[a.Name()] = a
			names = append(names, a.Name())
		}
		sort.Strings(names)
		for _, s := range names {
			o.WriteConsole("Artist: %s\n", s)
			artist := m[s]
			if artist != nil {
				ls.ListAlbums(o, artist.Albums(), 2)
			}
		}
	} else {
		albums := []*files.Album{}
		for _, a := range artists {
			albums = append(albums, a.Albums()...)
		}
		ls.ListAlbums(o, albums, 0)
	}
}

func (ls *ListSettings) ListAlbums(o output.Bus, albums []*files.Album, tab int) {
	if ls.Albums {
		m := map[string]*files.Album{}
		albumNames := []string{}
		for _, album := range albums {
			annotatedAlbumName := ls.AnnotateAlbumName(album)
			m[annotatedAlbumName] = album
			albumNames = append(albumNames, annotatedAlbumName)
		}
		sort.Strings(albumNames)
		for _, name := range albumNames {
			o.WriteConsole("%*sAlbum: %s\n", tab, "", name)
			album := m[name]
			if album != nil {
				ls.ListTracks(o, album.Tracks(), tab+2)
			}
		}
	} else {
		tracks := []*files.Track{}
		for _, album := range albums {
			tracks = append(tracks, album.Tracks()...)
		}
		ls.ListTracks(o, tracks, tab)
	}
}

func (ls *ListSettings) AnnotateAlbumName(album *files.Album) string {
	switch {
	case !ls.Artists && ls.Annotate:
		return strings.Join([]string{quote(album.Name()), "by", quote(album.RecordingArtistName())}, " ")
	default:
		return album.Name()
	}
}

func (ls *ListSettings) ListTracks(o output.Bus, tracks []*files.Track, tab int) {
	if !ls.Tracks {
		return
	}
	if ls.SortByNumber {
		ls.ListTracksByNumber(o, tracks, tab)
		return
	}
	if ls.SortByTitle {
		ls.ListTracksByName(o, tracks, tab)
	}
}

func (ls *ListSettings) ListTracksByNumber(o output.Bus, tracks []*files.Track, tab int) {
	m := map[int]*files.Track{}
	numbers := []int{}
	for _, track := range tracks {
		numbers = append(numbers, track.Number())
		m[track.Number()] = track
	}
	sort.Ints(numbers)
	for _, n := range numbers {
		track := m[n]
		if track != nil {
			o.WriteConsole("%*s%2d. %s\n", tab, "", n, track.CommonName())
			ls.ListTrackDetails(o, track, tab+2)
			ls.ListTrackDiagnostics(o, track, tab+2)
		}
	}
}

func (ls *ListSettings) ListTracksByName(o output.Bus, tracks []*files.Track, tab int) {
	annotatedNames := []string{}
	m := map[string]*files.Track{}
	for _, track := range tracks {
		annotatedName := ls.AnnotateTrackName(track)
		annotatedNames = append(annotatedNames, annotatedName)
		m[annotatedName] = track
	}
	sort.Strings(annotatedNames)
	for _, s := range annotatedNames {
		o.WriteConsole("%*s%s\n", tab, "", s)
		track := m[s]
		if track != nil {
			ls.ListTrackDetails(o, track, tab+2)
			ls.ListTrackDiagnostics(o, track, tab+2)
		}
	}
}

func quote(s string) string {
	return fmt.Sprintf("%q", s)
}

func (ls *ListSettings) AnnotateTrackName(track *files.Track) string {
	commonName := track.CommonName()
	if !ls.Annotate || ls.Albums {
		return commonName
	}
	trackNameParts := []string{quote(commonName), "on", quote(track.AlbumName())}
	if !ls.Artists {
		trackNameParts = append(trackNameParts, "by", quote(track.RecordingArtist()))
	}
	return strings.Join(trackNameParts, " ")
}

func (ls *ListSettings) ListTrackDetails(o output.Bus, track *files.Track, tab int) {
	if ls.Details {
		// go get information from track and display it
		m, err := track.Details()
		ShowDetails(o, track, m, err, tab)
	}
}

// split out for testing!
func ShowDetails(o output.Bus, track *files.Track, details map[string]string, detailsError error, tab int) {
	if detailsError != nil {
		o.Log(output.Error, "cannot get details", map[string]any{
			"error": detailsError,
			"track": track.String(),
		})
		o.WriteCanonicalError("The details are not available for track %q on album %q by artist %q: %q", track.CommonName(), track.AlbumName(), track.RecordingArtist(), detailsError.Error())
	} else if len(details) != 0 {
		keys := []string{}
		for k := range details {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		o.WriteConsole("%*sDetails:\n", tab, "")
		for _, k := range keys {
			o.WriteConsole("%*s%s = %q\n", tab+2, "", k, details[k])
		}
	}
}

func (ls *ListSettings) ListTrackDiagnostics(o output.Bus, track *files.Track, tab int) {
	if ls.Diagnostic {
		version, encoding, frames, err := track.ID3V2Diagnostics()
		ShowID3V2Diagnostics(o, track, version, encoding, frames, err, tab)
		tags, err := track.ID3V1Diagnostics()
		ShowID3V1Diagnostics(o, track, tags, err, tab)
	}
}

// split out for testing!
func ShowID3V1Diagnostics(o output.Bus, track *files.Track, tags []string, err error, tab int) {
	if err != nil {
		track.ReportMetadataReadError(o, files.ID3V1, err.Error())
	} else {
		for _, s := range tags {
			o.WriteConsole("%*sID3V1 %s\n", tab, "", s)
		}
	}
}

// split out for testing!
func ShowID3V2Diagnostics(o output.Bus, track *files.Track, version byte, encoding string, frames []string, err error, tab int) {
	if err != nil {
		track.ReportMetadataReadError(o, files.ID3V2, err.Error())
	} else {
		o.WriteConsole("%*sID3V2 Version: %v\n", tab, "", version)
		o.WriteConsole("%*sID3V2 Encoding: %q\n", tab, "", encoding)
		for _, frame := range frames {
			o.WriteConsole("%*sID3V2 %s\n", tab, "", frame)
		}
	}
}

func (ls *ListSettings) TracksSortable(o output.Bus) bool {
	bothSortingOptionsSet := ls.SortByNumber && ls.SortByTitle
	neitherSortingOptionSet := !ls.SortByNumber && !ls.SortByTitle
	if ls.Tracks {
		switch {
		case bothSortingOptionsSet:
			o.WriteCanonicalError("Track sorting cannot be done")
			o.WriteCanonicalError("Why?")
			if ls.SortByNumberUserSet {
				if ls.SortByTitleUserSet {
					o.WriteCanonicalError("You explicitly set %s and %s true", ListSortByNumberFlag, ListSortByTitleFlag)
				} else {
					o.WriteCanonicalError("The %s flag is configured true and you explicitly set %s true", ListSortByTitleFlag, ListSortByNumberFlag)
				}
			} else {
				if ls.SortByTitleUserSet {
					o.WriteCanonicalError("The %s flag is configured true and you explicitly set %s true", ListSortByNumberFlag, ListSortByTitleFlag)
				} else {
					o.WriteCanonicalError("The %s and %s flags are both configured true", ListSortByNumberFlag, ListSortByTitleFlag)
				}
			}
			o.WriteCanonicalError("What to do:\nEither edit the configuration file and use those default values, or use appropriate command line values")
			return false
		case ls.SortByNumber && !ls.Albums:
			o.WriteCanonicalError("Sorting tracks by number not possible.")
			o.WriteCanonicalError("Why?")
			o.WriteCanonicalError("Track numbers are only relevant if albums are also output.")
			if ls.SortByNumberUserSet {
				if ls.AlbumsUserSet {
					o.WriteCanonicalError("You set %s true and %s false.", ListSortByNumberFlag, ListAlbumsFlag)
				} else {
					o.WriteCanonicalError("You set %s true and %s is configured as false", ListSortByNumberFlag, ListAlbumsFlag)
				}
			} else {
				if ls.AlbumsUserSet {
					o.WriteCanonicalError("You set %s false and %s is configured as true", ListAlbumsFlag, ListSortByNumberFlag)
				} else {
					o.WriteCanonicalError("%s is configured as false, and %s is configured as true", ListAlbumsFlag, ListSortByNumberFlag)
				}
			}
			o.WriteCanonicalError("What to do:\nEither edit the configuration file or change which flags you set on the command line.")
			return false
		case neitherSortingOptionSet:
			if ls.SortByNumberUserSet && ls.SortByTitleUserSet {
				o.WriteCanonicalError("A listing of tracks is not possible.")
				o.WriteCanonicalError("Why?")
				o.WriteCanonicalError("Tracks are enabled, but you set both %s and %s false", ListSortByNumberFlag, ListSortByTitleFlag)
				o.WriteCanonicalError("What to do:\nEnable one of the sorting flags")
				return false
			}
			// pick a sensible option
			switch {
			case ls.SortByNumberUserSet:
				ls.SortByTitle = true // pick the other setting
			case ls.SortByTitleUserSet:
				ls.SortByNumber = true // pick the other setting
			default: // ok, pick something sensible, user does not care
				if ls.Albums {
					ls.SortByNumber = true
				} else {
					ls.SortByTitle = true
				}
			}
			o.Log(output.Info, "no track sorting set, providing a sensible value", map[string]any{
				ListAlbumsFlag:      ls.Albums,
				ListSortByNumber:    ls.SortByNumber,
				ListSortByTitleFlag: ls.SortByTitle,
			})
		}
	} else if (ls.SortByNumber && ls.SortByNumberUserSet) ||
		(ls.SortByTitle && ls.SortByTitleUserSet) {
		o.WriteCanonicalError("Your sorting preferences are not relevant")
		o.WriteCanonicalError("Why?")
		o.WriteCanonicalError("Tracks are not included in the output, but you explicitly set %s or %s true.", ListSortByNumberFlag, ListSortByTitleFlag)
		o.WriteCanonicalError("What to do:\nEither set %s true or remove the sorting flags from the command line.", ListTracksFlag)
		return false
	}
	return true
}

func (ls *ListSettings) HasWorkToDo(o output.Bus) bool {
	if ls.Albums || ls.Artists || ls.Tracks {
		return true
	}
	userPartiallyAtFault := ls.AlbumsUserSet || ls.ArtistsUserSet || ls.TracksUserSet
	o.WriteCanonicalError("No listing will be output.\nWhy?\n")
	if userPartiallyAtFault {
		flagsUserSet := []string{}
		flagsFromConfig := []string{}
		if ls.AlbumsUserSet {
			flagsUserSet = append(flagsUserSet, ListAlbumsFlag)
		} else {
			flagsFromConfig = append(flagsFromConfig, ListAlbumsFlag)
		}
		if ls.ArtistsUserSet {
			flagsUserSet = append(flagsUserSet, ListArtistsFlag)
		} else {
			flagsFromConfig = append(flagsFromConfig, ListArtistsFlag)
		}
		if ls.TracksUserSet {
			flagsUserSet = append(flagsUserSet, ListTracksFlag)
		} else {
			flagsFromConfig = append(flagsFromConfig, ListTracksFlag)
		}
		if len(flagsFromConfig) == 0 {
			o.WriteCanonicalError("You explicitly set %s, %s, and %s false", ListAlbumsFlag, ListArtistsFlag, ListTracksFlag)
		} else {
			o.WriteCanonicalError("In addition to %s configured false, you explicitly set %s false", strings.Join(flagsFromConfig, " and "), strings.Join(flagsUserSet, " and "))
		}
	} else {
		o.WriteCanonicalError("The flags %s, %s, and %s are all configured false", ListAlbumsFlag, ListArtistsFlag, ListTracksFlag)
	}
	o.WriteError("What to do:\n")
	o.WriteCanonicalError("Either:\n[1] Edit the configuration file so that at least one of these flags is true, or\n[2] explicitly set at least one of these flags true on the command line")
	return false
}

func ProcessListFlags(o output.Bus, values map[string]*FlagValue) (*ListSettings, bool) {
	settings := &ListSettings{}
	ok := true // optimistic
	var err error
	if settings.Albums, settings.AlbumsUserSet, err = GetBool(o, values, ListAlbums); err != nil {
		ok = false
	}
	if settings.Annotate, _, err = GetBool(o, values, ListAnnotate); err != nil {
		ok = false
	}
	if settings.Artists, settings.ArtistsUserSet, err = GetBool(o, values, ListArtists); err != nil {
		ok = false
	}
	if settings.Details, _, err = GetBool(o, values, ListDetails); err != nil {
		ok = false
	}
	if settings.Diagnostic, _, err = GetBool(o, values, ListDiagnostic); err != nil {
		ok = false
	}
	if settings.SortByNumber, settings.SortByNumberUserSet, err = GetBool(o, values, ListSortByNumber); err != nil {
		ok = false
	}
	if settings.SortByTitle, settings.SortByTitleUserSet, err = GetBool(o, values, ListSortByTitle); err != nil {
		ok = false
	}
	if settings.Tracks, settings.TracksUserSet, err = GetBool(o, values, ListTracks); err != nil {
		ok = false
	}
	return settings, ok
}

func init() {
	RootCmd.AddCommand(ListCmd)
	addDefaults(ListFlags)
	c := getConfiguration()
	o := getBus()
	AddFlags(o, c, ListCmd.Flags(), ListFlags, true)
}
