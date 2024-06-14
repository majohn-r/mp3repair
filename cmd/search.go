package cmd

import (
	"io/fs"
	"mp3repair/internal/files"
	"path/filepath"
	"regexp"
	"strings"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
)

const (
	SearchAlbumFilter        = "albumFilter"
	SearchAlbumFilterFlag    = "--" + SearchAlbumFilter
	SearchArtistFilter       = "artistFilter"
	SearchArtistFilterFlag   = "--" + SearchArtistFilter
	SearchFileExtensions     = "extensions"
	SearchFileExtensionsFlag = "--" + SearchFileExtensions
	SearchTopDir             = "topDir"
	SearchTopDirFlag         = "--" + SearchTopDir
	SearchTrackFilter        = "trackFilter"
	SearchTrackFilterFlag    = "--" + SearchTrackFilter
	searchUsage              = "[" + SearchAlbumFilterFlag + " regex] [" +
		SearchArtistFilterFlag + " regex] [" + SearchTrackFilterFlag + " regex] [" +
		SearchTopDirFlag + " dir] [" + SearchFileExtensionsFlag + " extensions]"
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
	// SearchFlags is a common set of flags that many commands need to use
	SearchFlags = &SectionFlags{
		SectionName: "search",
		Details: map[string]*FlagDetails{
			SearchAlbumFilter: {
				Usage:        "regular expression specifying which albums to select",
				ExpectedType: StringType,
				DefaultValue: ".*",
			},
			SearchArtistFilter: {
				Usage:        "regular expression specifying which artists to select",
				ExpectedType: StringType,
				DefaultValue: ".*",
			},
			SearchTrackFilter: {
				Usage:        "regular expression specifying which tracks to select",
				ExpectedType: StringType,
				DefaultValue: ".*",
			},
			SearchTopDir: {
				Usage:        "top directory specifying where to find mp3 files",
				ExpectedType: StringType,
				DefaultValue: filepath.Join("%HOMEPATH%", "Music"),
			},
			SearchFileExtensions: {
				Usage:        "comma-delimited list of file extensions used by mp3 files",
				ExpectedType: StringType,
				DefaultValue: ".mp3",
			},
		},
	}
)

type SearchSettings struct {
	ArtistFilter   *regexp.Regexp
	AlbumFilter    *regexp.Regexp
	TrackFilter    *regexp.Regexp
	FileExtensions []string
	TopDirectory   string
}

func (ss *SearchSettings) Values() map[string]any {
	return map[string]any{
		SearchAlbumFilterFlag:    ss.AlbumFilter,
		SearchArtistFilterFlag:   ss.ArtistFilter,
		SearchTrackFilterFlag:    ss.TrackFilter,
		SearchTopDirFlag:         ss.TopDirectory,
		SearchFileExtensionsFlag: ss.FileExtensions,
	}
}

func EvaluateSearchFlags(o output.Bus, producer FlagProducer) (*SearchSettings, bool) {
	values, eSlice := ReadFlags(producer, SearchFlags)
	if ProcessFlagErrors(o, eSlice) {
		return ProcessSearchFlags(o, values)
	}
	return &SearchSettings{}, false
}

func ProcessSearchFlags(o output.Bus, values map[string]*CommandFlag[any]) (settings *SearchSettings, flagsOk bool) {
	flagsOk = true // optimistic
	regexOk := true
	settings = &SearchSettings{}
	// process the filters first, so we can attempt to guide the user to better
	// choice(s)
	albumFilter := EvaluateFilter(o, FilterFlag{
		Values:             values,
		FlagName:           SearchAlbumFilter,
		FlagRepresentation: SearchAlbumFilterFlag,
	})
	switch {
	case albumFilter.FilterOk:
		settings.AlbumFilter = albumFilter.Regex
	default:
		if !albumFilter.RegexOk {
			regexOk = false
		}
		flagsOk = false
	}
	artistFilter := EvaluateFilter(o, FilterFlag{
		Values:             values,
		FlagName:           SearchArtistFilter,
		FlagRepresentation: SearchArtistFilterFlag,
	})
	switch {
	case artistFilter.FilterOk:
		settings.ArtistFilter = artistFilter.Regex
	default:
		if !artistFilter.RegexOk {
			regexOk = false
		}
		flagsOk = false
	}
	trackFilter := EvaluateFilter(o, FilterFlag{
		Values:             values,
		FlagName:           SearchTrackFilter,
		FlagRepresentation: SearchTrackFilterFlag,
	})
	switch {
	case trackFilter.FilterOk:
		settings.TrackFilter = trackFilter.Regex
	default:
		if !trackFilter.RegexOk {
			regexOk = false
		}
		flagsOk = false
	}
	if !regexOk {
		// user has attempted to use filters that don't compile
		o.WriteCanonicalError(searchRegexInstructions)
	}
	topDir, topDirFilterOk := EvaluateTopDir(o, values)
	switch {
	case topDirFilterOk:
		settings.TopDirectory = topDir
	default:
		flagsOk = false
	}
	extensions, extensionsFilterOk := EvaluateFileExtensions(o, values)
	switch {
	case extensionsFilterOk:
		settings.FileExtensions = extensions
	default:
		flagsOk = false
	}
	return
}

func EvaluateFileExtensions(o output.Bus, values map[string]*CommandFlag[any]) ([]string, bool) {
	rawValue, flagErr := GetString(o, values, SearchFileExtensions)
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
			o.WriteCanonicalError("The extension %q cannot be used.", candidate)
			failedCandidates = append(failedCandidates, candidate)
			extensionsValid = false
		}
	}
	if !extensionsValid {
		o.WriteCanonicalError("Why?")
		o.WriteCanonicalError(
			"Extensions must be at least two characters long and begin with '.'")
		o.WriteCanonicalError("What to do:\nProvide appropriate extensions.")
		o.Log(output.Error, "invalid file extensions", map[string]any{
			"rejected":               failedCandidates,
			SearchFileExtensionsFlag: rawValue.Value,
		})
	}
	return extensions, extensionsValid
}

func EvaluateTopDir(o output.Bus, values map[string]*CommandFlag[any]) (dir string, topDirValid bool) {
	rawValue, flagErr := GetString(o, values, SearchTopDir)
	if flagErr != nil {
		return
	}
	file, fileErr := cmdtoolkit.FileSystem().Stat(rawValue.Value)
	if fileErr != nil {
		o.WriteCanonicalError("The %s value, %q, cannot be used", SearchTopDirFlag,
			rawValue.Value)
		o.Log(output.Error, "invalid directory", map[string]any{
			"error":          fileErr,
			SearchTopDirFlag: rawValue.Value,
			"user-set":       rawValue.UserSet,
		})
		o.WriteCanonicalError("Why?")
		switch rawValue.UserSet {
		case true:
			o.WriteCanonicalError("The value you specified is not a readable file.")
			o.WriteCanonicalError(
				"What to do:\nSpecify a value that is a readable file.")
		case false:
			o.WriteCanonicalError(
				"The currently configured value is not a readable file.")
			o.WriteCanonicalError("What to do:\n"+
				"Edit the configuration file or specify %s with a value that is a"+
				" readable file.", SearchTopDirFlag)
		}
		return
	}
	if !file.IsDir() {
		o.WriteCanonicalError("The %s value, %q, cannot be used", SearchTopDirFlag,
			rawValue.Value)
		o.Log(output.Error, "the file is not a directory", map[string]any{
			SearchTopDirFlag: rawValue.Value,
			"user-set":       rawValue.UserSet,
		})
		o.WriteCanonicalError("Why?")
		switch rawValue.UserSet {
		case true:
			o.WriteCanonicalError(
				"The value you specified is not the name of a directory.")
			o.WriteCanonicalError("What to do:\n" +
				"Specify a value that is the name of a directory.")
		default:
			o.WriteCanonicalError(
				"The currently configured value is not the name of a directory.")
			o.WriteCanonicalError("What to do:\n"+
				"Edit the configuration file or specify %s with a value that is the"+
				" name of a directory.", SearchTopDirFlag)
		}
		return
	}
	dir = rawValue.Value
	topDirValid = true
	return
}

type FilterFlag struct {
	Values             map[string]*CommandFlag[any]
	FlagName           string
	FlagRepresentation string
}

type EvaluatedFilter struct {
	Regex    *regexp.Regexp
	FilterOk bool
	RegexOk  bool
}

func EvaluateFilter(o output.Bus, filtering FilterFlag) EvaluatedFilter {
	result := EvaluatedFilter{RegexOk: true}
	rawValue, flagErr := GetString(o, filtering.Values, filtering.FlagName)
	if flagErr != nil {
		return result
	}
	f, regexErr := regexp.Compile(rawValue.Value)
	if regexErr != nil {
		o.Log(output.Error, "the filter cannot be parsed as a regular expression",
			map[string]any{
				filtering.FlagRepresentation: rawValue.Value,
				"user-set":                   rawValue.UserSet,
				"error":                      regexErr,
			})
		o.WriteCanonicalError("the %s value %q cannot be used", filtering.FlagRepresentation, rawValue.Value)
		switch {
		case rawValue.UserSet:
			o.WriteCanonicalError("Why?\n"+
				"The value of %s that you specified is not a valid regular expression: %v.",
				filtering.FlagRepresentation, regexErr)
			o.WriteCanonicalError("What to do:\n"+
				"Either try a different setting,"+
				" or omit setting %s and try the default value.", filtering.FlagRepresentation)
		default:
			o.WriteCanonicalError("Why?\n"+
				"The configured default value of %s is not a valid regular expression: %v.",
				filtering.FlagRepresentation, regexErr)
			o.WriteCanonicalError("What to do:\n"+
				"Either edit the defaults.yaml file containing the settings,"+
				" or explicitly set %s to a better value.", filtering.FlagRepresentation)
		}
		result.RegexOk = false
		return result
	}
	result.Regex = f
	result.FilterOk = true
	return result
}

func (ss *SearchSettings) Filter(o output.Bus, originalArtists []*files.Artist) []*files.Artist {
	filteredArtists := make([]*files.Artist, 0, len(originalArtists))
	for _, originalArtist := range originalArtists {
		if ss.ArtistFilter.MatchString(originalArtist.Name) && originalArtist.HasAlbums() {
			filteredArtist := originalArtist.Copy()
			for _, originalAlbum := range originalArtist.Albums {
				if ss.AlbumFilter.MatchString(originalAlbum.Title) &&
					originalAlbum.HasTracks() {
					filteredAlbum := originalAlbum.Copy(filteredArtist, false)
					for _, originalTrack := range originalAlbum.Tracks {
						if ss.TrackFilter.MatchString(originalTrack.SimpleName) {
							filteredTrack := originalTrack.Copy(filteredAlbum)
							filteredAlbum.AddTrack(filteredTrack)
						}
					}
					if filteredAlbum.HasTracks() {
						filteredArtist.AddAlbum(filteredAlbum)
					}
				}
			}
			if filteredArtist.HasAlbums() {
				filteredArtists = append(filteredArtists, filteredArtist)
			}
		}
	}
	if len(filteredArtists) == 0 {
		o.WriteCanonicalError("No mp3 files remain after filtering.")
		o.WriteCanonicalError("Why?")
		o.WriteCanonicalError("After applying %s=%q, %s=%q, and %s=%q, no files remained",
			SearchArtistFilterFlag, ss.ArtistFilter, SearchAlbumFilterFlag, ss.AlbumFilter,
			SearchTrackFilterFlag, ss.TrackFilter)
		o.WriteCanonicalError("What to do:\nUse less restrictive filter settings.")
		o.Log(output.Error, "no files remain after filtering", map[string]any{
			SearchArtistFilterFlag: ss.ArtistFilter,
			SearchAlbumFilterFlag:  ss.AlbumFilter,
			SearchTrackFilterFlag:  ss.TrackFilter,
		})
	}
	return filteredArtists
}

func (ss *SearchSettings) Load(o output.Bus) []*files.Artist {
	artistFiles, dirRead := ReadDirectory(o, ss.TopDirectory)
	artists := make([]*files.Artist, 0, len(artistFiles))
	if dirRead {
		for _, artistFile := range artistFiles {
			if artistFile.IsDir() {
				artist := files.NewArtistFromFile(artistFile, ss.TopDirectory)
				ss.addAlbums(o, artist)
				artists = append(artists, artist)
			}
		}
	}
	if len(artists) == 0 {
		o.WriteCanonicalError(
			"No mp3 files could be found using the specified parameters.")
		o.WriteCanonicalError("Why?")
		o.WriteCanonicalError("There were no directories found in %q (the %s value)",
			ss.TopDirectory, SearchTopDirFlag)
		o.WriteCanonicalError("What to do:\n"+
			"Set %s to the path of a directory that contains artist directories",
			SearchTopDirFlag)
		o.Log(output.Error, "cannot find any artist directories", map[string]any{
			SearchTopDirFlag: ss.TopDirectory,
		})
	}
	return artists
}

func (ss *SearchSettings) addAlbums(o output.Bus, artist *files.Artist) {
	if albumFiles, artistDirRead := ReadDirectory(o, artist.FilePath); artistDirRead {
		for _, albumFile := range albumFiles {
			if albumFile.IsDir() {
				album := files.NewAlbumFromFile(albumFile, artist)
				ss.addTracks(o, album)
				artist.AddAlbum(album)
			}
		}
	}
}

func (ss *SearchSettings) addTracks(o output.Bus, album *files.Album) {
	if trackFiles, filesAvailable := ReadDirectory(o, album.FilePath); filesAvailable {
		for _, trackFile := range trackFiles {
			if extension, isTrack := ss.isValidTrackFile(trackFile); isTrack {
				parsedName, valid := files.TrackNameParser{
					FileName:  trackFile.Name(),
					Album:     album,
					Extension: extension,
				}.Parse(o)
				if valid {
					album.AddTrack(files.TrackMaker{
						Album:      album,
						FileName:   trackFile.Name(),
						SimpleName: parsedName.SimpleName,
						Number:     parsedName.Number,
					}.NewTrack())
				}
			}
		}
	}

}

func (ss *SearchSettings) isValidTrackFile(file fs.FileInfo) (string, bool) {
	extension := filepath.Ext(file.Name())
	if !file.IsDir() {
		for _, expectedExtension := range ss.FileExtensions {
			if expectedExtension == extension {
				return extension, true
			}
		}
	}
	return extension, false
}
