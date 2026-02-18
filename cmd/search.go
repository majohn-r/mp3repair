package cmd

import (
	"io/fs"
	"mp3repair/internal/files"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/adrg/xdg"
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
)
/*
Copyright © 2026 Marc Johnson (marc.johnson27591@gmail.com)
*/

const (
	searchAlbumFilter        = "albumFilter"
	searchAlbumFilterFlag    = "--" + searchAlbumFilter
	searchArtistFilter       = "artistFilter"
	searchArtistFilterFlag   = "--" + searchArtistFilter
	searchFileExtensions     = "extensions"
	searchFileExtensionsFlag = "--" + searchFileExtensions
	searchTrackFilter        = "trackFilter"
	searchTrackFilterFlag    = "--" + searchTrackFilter
	searchUsage              = "[" + searchAlbumFilterFlag + " regex] [" +
		searchArtistFilterFlag + " regex] [" + searchTrackFilterFlag + " regex] [" +
		searchFileExtensionsFlag + " extensions]"
	searchRegexInstructions = "" +
		`Here are some common errors in filter expressions and what to do:
Character class problems
Character classes are sets of 1 or more characters, enclosed in square brackets: []
A common error is to forget the final ] bracket.
Character classes can include a range of characters, like this: [a-z], which means
any character between a and z. Order is important - one might think that [z-a] would
mean the same thing, but it doesn't; z comes after a. Do an internet search for ASCII
table; that's the expected order for ranges of characters. And that means [A-z] means
any letter, and [a-Z] is an error.
Repetition problems
The characters '+' and '*' specify repetition: a+ means "exactly one a" and a* means
"0 or more a's". You can also put a count in curly braces - a{2} means "exactly two a's".
Repetition can only be used once for a character or character class. 'a++', 'a+*',
and so on, are not allowed.
For more (too much, often, you are warned) information, do a web search for
"golang regexp".`
)

var (
	searchFlags = &cmdtoolkit.FlagSet{
		Name: "search",
		Details: map[string]*cmdtoolkit.FlagDetails{
			searchAlbumFilter: {
				Usage:        "regular expression specifying which albums to select",
				ExpectedType: cmdtoolkit.StringType,
				DefaultValue: ".*",
			},
			searchArtistFilter: {
				Usage:        "regular expression specifying which artists to select",
				ExpectedType: cmdtoolkit.StringType,
				DefaultValue: ".*",
			},
			searchTrackFilter: {
				Usage:        "regular expression specifying which tracks to select",
				ExpectedType: cmdtoolkit.StringType,
				DefaultValue: ".*",
			},
			searchFileExtensions: {
				Usage:        "comma-delimited list of file extensions used by mp3 files",
				ExpectedType: cmdtoolkit.StringType,
				DefaultValue: ".mp3",
			},
		},
	}
)

type searchSettings struct {
	artistFilter   *regexp.Regexp
	albumFilter    *regexp.Regexp
	trackFilter    *regexp.Regexp
	fileExtensions []string
	musicDir       string
}

func evaluateSearchFlags(o output.Bus, producer cmdtoolkit.FlagProducer) (*searchSettings, bool) {
	values, eSlice := cmdtoolkit.ReadFlags(producer, searchFlags)
	if cmdtoolkit.ProcessFlagErrors(o, eSlice) {
		return processSearchFlags(o, values)
	}
	return &searchSettings{}, false
}

func processSearchFlags(
	o output.Bus,
	values map[string]*cmdtoolkit.CommandFlag[any],
) (settings *searchSettings, flagsOk bool) {
	flagsOk = true // optimistic
	regexOk := true
	settings = &searchSettings{}
	// process the filters first, so we can attempt to guide the user to better
	// choice(s)
	albumFilter := evaluateFilter(o, filterFlag{
		values:         values,
		name:           searchAlbumFilter,
		representation: searchAlbumFilterFlag,
	})
	switch {
	case albumFilter.filterOk:
		settings.albumFilter = albumFilter.regex
	default:
		if !albumFilter.regexOk {
			regexOk = false
		}
		flagsOk = false
	}
	artistFilter := evaluateFilter(o, filterFlag{
		values:         values,
		name:           searchArtistFilter,
		representation: searchArtistFilterFlag,
	})
	switch {
	case artistFilter.filterOk:
		settings.artistFilter = artistFilter.regex
	default:
		if !artistFilter.regexOk {
			regexOk = false
		}
		flagsOk = false
	}
	trackFilter := evaluateFilter(o, filterFlag{
		values:         values,
		name:           searchTrackFilter,
		representation: searchTrackFilterFlag,
	})
	switch {
	case trackFilter.filterOk:
		settings.trackFilter = trackFilter.regex
	default:
		if !trackFilter.regexOk {
			regexOk = false
		}
		flagsOk = false
	}
	if !regexOk {
		// user has attempted to use filters that don't compile
		o.ErrorPrintln(searchRegexInstructions)
	}
	musicDir, musicDirOk := evaluateMusicDir(o)
	switch {
	case musicDirOk:
		settings.musicDir = musicDir
	default:
		flagsOk = false
	}
	extensions, extensionsFilterOk := evaluateFileExtensions(o, values)
	switch {
	case extensionsFilterOk:
		settings.fileExtensions = extensions
	default:
		flagsOk = false
	}
	return
}

func evaluateFileExtensions(o output.Bus, values map[string]*cmdtoolkit.CommandFlag[any]) ([]string, bool) {
	rawValue, flagErr := cmdtoolkit.GetString(o, values, searchFileExtensions)
	if flagErr != nil {
		return []string{}, false
	}
	candidates := strings.Split(rawValue.Value, ",")
	var failedCandidates []string
	var extensions []string
	extensionsValid := true
	for _, candidate := range candidates {
		switch {
		case strings.HasPrefix(candidate, ".") && len(candidate) >= 2:
			extensions = append(extensions, candidate)
		default:
			o.ErrorPrintf("The extension %q cannot be used.\n", candidate)
			failedCandidates = append(failedCandidates, candidate)
			extensionsValid = false
		}
	}
	if !extensionsValid {
		o.ErrorPrintln("Why?")
		o.ErrorPrintln("Extensions must be at least two characters long and begin with '.'.")
		o.ErrorPrintln("What to do:")
		o.ErrorPrintln("Provide appropriate extensions.")
		o.Log(output.Error, "invalid file extensions", map[string]any{
			"rejected":               failedCandidates,
			searchFileExtensionsFlag: rawValue.Value,
		})
	}
	return extensions, extensionsValid
}

func evaluateMusicDir(o output.Bus) (string, bool) {
	musicDir := xdg.UserDirs.Music
	file, fileErr := cmdtoolkit.FileSystem().Stat(musicDir)
	if fileErr != nil {
		o.Log(output.Error, "invalid directory", map[string]any{
			"error":     fileErr,
			"directory": musicDir,
		})
		reportBadMusicDir(o, musicDir)
		return "", false
	}
	if !file.IsDir() {
		o.Log(output.Error, "the file is not a directory", map[string]any{"directory": musicDir})
		reportBadMusicDir(o, musicDir)
		return "", false
	}
	return musicDir, true
}

func reportBadMusicDir(o output.Bus, musicDir string) {
	o.ErrorPrintf("The music directory value, %q, cannot be used.\n", musicDir)
	o.ErrorPrintln("Why?")
	o.ErrorPrintln("The value is not a readable folder.")
	o.ErrorPrintln("What to do:")
	o.ErrorPrintln("Set XDG_MUSIC_DIR to a value that is a readable folder.")
}

type filterFlag struct {
	values         map[string]*cmdtoolkit.CommandFlag[any]
	name           string
	representation string
}

type evaluatedFilter struct {
	regex    *regexp.Regexp
	filterOk bool
	regexOk  bool
}

func evaluateFilter(o output.Bus, filtering filterFlag) evaluatedFilter {
	result := evaluatedFilter{regexOk: true}
	rawValue, flagErr := cmdtoolkit.GetString(o, filtering.values, filtering.name)
	if flagErr != nil {
		return result
	}
	f, regexErr := regexp.Compile(rawValue.Value)
	if regexErr != nil {
		o.Log(output.Error, "the filter cannot be parsed as a regular expression",
			map[string]any{
				filtering.representation: rawValue.Value,
				"user-set":               rawValue.UserSet,
				"error":                  regexErr,
			})
		o.ErrorPrintf("The %s value %q cannot be used.\n", filtering.representation, rawValue.Value)
		switch {
		case rawValue.UserSet:
			o.ErrorPrintln("Why?")
			o.ErrorPrintf(
				"The value of %s that you specified is not a valid regular expression: %s.\n",
				filtering.representation,
				cmdtoolkit.ErrorToString(regexErr),
			)
			o.ErrorPrintln("What to do:")
			o.BeginErrorList(false)
			o.ErrorPrintln("Try a different setting, or")
			o.ErrorPrintf("Omit setting %s and try the default value.\n", filtering.representation)
			o.EndErrorList()
		default:
			o.ErrorPrintln("Why?")
			o.ErrorPrintf(
				"The configured default value of %s is not a valid regular expression: %v.\n",
				filtering.representation,
				cmdtoolkit.ErrorToString(regexErr),
			)
			o.ErrorPrintln("What to do:")
			o.BeginErrorList(false)
			o.ErrorPrintln("Edit the defaults.yaml file containing the settings, or")
			o.ErrorPrintf("Explicitly set %s to a better value.\n", filtering.representation)
			o.EndErrorList()
		}
		result.regexOk = false
		return result
	}
	result.regex = f
	result.filterOk = true
	return result
}

func (ss *searchSettings) filter(o output.Bus, originalArtists []*files.Artist) []*files.Artist {
	filteredArtists := make([]*files.Artist, 0, len(originalArtists))
	if len(originalArtists) > 0 {
		for _, originalArtist := range originalArtists {
			if ss.artistFilter.MatchString(originalArtist.Name()) && originalArtist.HasAlbums() {
				filteredArtist := originalArtist.Copy()
				for _, originalAlbum := range originalArtist.Albums() {
					if ss.albumFilter.MatchString(originalAlbum.Title()) && originalAlbum.HasTracks() {
						var willHaveTracks bool
						for _, originalTrack := range originalAlbum.Tracks() {
							if ss.trackFilter.MatchString(originalTrack.Name()) {
								willHaveTracks = true
							}
						}
						filteredAlbum := originalAlbum.Copy(filteredArtist, false, willHaveTracks)
						for _, originalTrack := range originalAlbum.Tracks() {
							if ss.trackFilter.MatchString(originalTrack.Name()) {
								originalTrack.Copy(filteredAlbum, true)
							}
						}
					}
				}
				if filteredArtist.HasAlbums() {
					filteredArtists = append(filteredArtists, filteredArtist)
				}
			}
		}
		if len(filteredArtists) == 0 {
			o.ErrorPrintln("No mp3 files remain after filtering.")
			o.ErrorPrintln("Why?")
			o.ErrorPrintf(
				"After applying %s=%q, %s=%q, and %s=%q, no files remained.\n",
				searchArtistFilterFlag,
				ss.artistFilter,
				searchAlbumFilterFlag,
				ss.albumFilter,
				searchTrackFilterFlag,
				ss.trackFilter,
			)
			o.ErrorPrintln("What to do:")
			o.ErrorPrintln("Use less restrictive filter settings.")
			o.Log(output.Error, "no files remain after filtering", map[string]any{
				searchArtistFilterFlag: ss.artistFilter,
				searchAlbumFilterFlag:  ss.albumFilter,
				searchTrackFilterFlag:  ss.trackFilter,
			})
		}
	}
	return filteredArtists
}

func (ss *searchSettings) load(o output.Bus) []*files.Artist {
	artistFiles, dirRead := readDirectory(o, ss.musicDir)
	artists := make([]*files.Artist, 0, len(artistFiles))
	if dirRead {
		for _, artistFile := range artistFiles {
			if artistFile.IsDir() {
				artist := files.NewArtistFromFile(artistFile, ss.musicDir)
				ss.addAlbums(o, artist)
				artists = append(artists, artist)
			}
		}
	}
	if len(artists) == 0 {
		o.ErrorPrintln("No mp3 files could be found using the specified parameters.")
		o.ErrorPrintln("Why?")
		o.ErrorPrintf("There were no directories found in %q.\n", ss.musicDir)
		o.ErrorPrintln("What to do:")
		o.ErrorPrintln("Set XDG_MUSIC_DIR to the path of a directory that contains artist directories.")
		o.Log(output.Error, "cannot find any artist directories", map[string]any{
			"$XDG_MUSIC_DIR": ss.musicDir,
		})
	}
	return artists
}

func (ss *searchSettings) addAlbums(o output.Bus, artist *files.Artist) {
	if albumFiles, artistDirRead := readDirectory(o, artist.Directory()); artistDirRead {
		for _, albumFile := range albumFiles {
			if albumFile.IsDir() {
				album := files.NewAlbumFromFile(albumFile, artist)
				ss.addTracks(o, album)
			}
		}
	}
}

func (ss *searchSettings) addTracks(o output.Bus, album *files.Album) {
	if trackFiles, filesAvailable := readDirectory(o, album.Directory()); filesAvailable {
		for _, trackFile := range trackFiles {
			if extension, isTrack := ss.isValidTrackFile(trackFile); isTrack {
				var parsedName *files.ParsedTrackName
				var valid bool
				parsedName, valid = files.TrackNameParser{
					FileName:  trackFile.Name(),
					Album:     album,
					Extension: extension,
				}.Parse(o)
				if valid {
					files.TrackMaker{
						Album:      album,
						FileName:   trackFile.Name(),
						SimpleName: parsedName.SimpleName,
						Number:     parsedName.Number,
					}.NewTrack(true)
				}
			}
		}
	}
}

func (ss *searchSettings) isValidTrackFile(file fs.FileInfo) (string, bool) {
	extension := filepath.Ext(file.Name())
	if !file.IsDir() {
		for _, expectedExtension := range ss.fileExtensions {
			if expectedExtension == extension {
				return extension, true
			}
		}
	}
	return extension, false
}

func init() {
	cmdtoolkit.AddDefaults(searchFlags)
}
