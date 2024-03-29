/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd

import (
	"io/fs"
	"mp3repair/internal/files"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
	SearchFlags = NewSectionFlags().WithSectionName("search").WithFlags(
		map[string]*FlagDetails{
			SearchAlbumFilter: NewFlagDetails().WithUsage(
				"regular expression specifying which albums to select").WithExpectedType(
				StringType).WithDefaultValue(".*"),
			SearchArtistFilter: NewFlagDetails().WithUsage(
				"regular expression specifying which artists to select").WithExpectedType(
				StringType).WithDefaultValue(".*"),
			SearchTrackFilter: NewFlagDetails().WithUsage(
				"regular expression specifying which tracks to select").WithExpectedType(
				StringType).WithDefaultValue(".*"),
			SearchTopDir: NewFlagDetails().WithUsage(
				"top directory specifying where to find mp3 files").WithExpectedType(
				StringType).WithDefaultValue(filepath.Join("%HOMEPATH%", "Music")),
			SearchFileExtensions: NewFlagDetails().WithUsage(
				"comma-delimited list of file extensions used by mp3 files").WithExpectedType(
				StringType).WithDefaultValue(".mp3"),
		},
	)
)

type SearchSettings struct {
	albumFilter    *regexp.Regexp
	artistFilter   *regexp.Regexp
	fileExtensions []string
	topDirectory   string
	trackFilter    *regexp.Regexp
}

func NewSearchSettings() *SearchSettings {
	return &SearchSettings{}
}

func (ss *SearchSettings) Values() map[string]any {
	return map[string]any{
		SearchAlbumFilterFlag:    ss.albumFilter,
		SearchArtistFilterFlag:   ss.artistFilter,
		SearchTrackFilterFlag:    ss.trackFilter,
		SearchTopDirFlag:         ss.topDirectory,
		SearchFileExtensionsFlag: ss.fileExtensions,
	}
}

func (ss *SearchSettings) WithAlbumFilter(r *regexp.Regexp) *SearchSettings {
	ss.albumFilter = r
	return ss
}

func (ss *SearchSettings) WithArtistFilter(r *regexp.Regexp) *SearchSettings {
	ss.artistFilter = r
	return ss
}

func (ss *SearchSettings) WithFileExtensions(s []string) *SearchSettings {
	ss.fileExtensions = s
	return ss
}

func (ss *SearchSettings) WithTopDirectory(s string) *SearchSettings {
	ss.topDirectory = s
	return ss
}

func (ss *SearchSettings) WithTrackFilter(r *regexp.Regexp) *SearchSettings {
	ss.trackFilter = r
	return ss
}

func EvaluateSearchFlags(o output.Bus, producer FlagProducer) (*SearchSettings, bool) {
	values, eSlice := ReadFlags(producer, SearchFlags)
	if ProcessFlagErrors(o, eSlice) {
		return ProcessSearchFlags(o, values)
	}
	return &SearchSettings{}, false
}

func ProcessSearchFlags(o output.Bus, values map[string]*FlagValue) (settings *SearchSettings, ok bool) {
	ok = true // optimistic
	regexOk := true
	settings = &SearchSettings{}
	// process the filters first, so we can attempt to guide the user to better
	// choice(s)
	if albumFilter, _ok, _regexOk := EvaluateFilter(o, values, SearchAlbumFilter,
		SearchAlbumFilterFlag); _ok {
		settings.albumFilter = albumFilter
	} else {
		if !_regexOk {
			regexOk = false
		}
		ok = false
	}
	if artistFilter, _ok, _regexOk := EvaluateFilter(o, values, SearchArtistFilter,
		SearchArtistFilterFlag); _ok {
		settings.artistFilter = artistFilter
	} else {
		if !_regexOk {
			regexOk = false
		}
		ok = false
	}
	if trackFilter, _ok, _regexOk := EvaluateFilter(o, values, SearchTrackFilter,
		SearchTrackFilterFlag); _ok {
		settings.trackFilter = trackFilter
	} else {
		if !_regexOk {
			regexOk = false
		}
		ok = false
	}
	if !regexOk {
		// user has attempted to use filters that don't compile
		o.WriteCanonicalError(searchRegexInstructions)
	}
	if topDir, _ok := EvaluateTopDir(o, values); _ok {
		settings.topDirectory = topDir
	} else {
		ok = false
	}
	if extensions, _ok := EvaluateFileExtensions(o, values); _ok {
		settings.fileExtensions = extensions
	} else {
		ok = false
	}
	return
}

func EvaluateFileExtensions(o output.Bus, values map[string]*FlagValue) ([]string, bool) {
	extensions := []string{}
	ok := false
	if rawValue, _, err := GetString(o, values, SearchFileExtensions); err == nil {
		candidates := strings.Split(rawValue, ",")
		failedCandidates := []string{}
		ok = true
		for _, candidate := range candidates {
			if strings.HasPrefix(candidate, ".") && len(candidate) >= 2 {
				extensions = append(extensions, candidate)
			} else {
				o.WriteCanonicalError("The extension %q cannot be used.", candidate)
				failedCandidates = append(failedCandidates, candidate)
				ok = false
			}
		}
		if !ok {
			o.WriteCanonicalError("Why?")
			o.WriteCanonicalError(
				"Extensions must be at least two characters long and begin with '.'")
			o.WriteCanonicalError("What to do:\nProvide appropriate extensions.")
			o.Log(output.Error, "invalid file extensions", map[string]any{
				"rejected":               failedCandidates,
				SearchFileExtensionsFlag: rawValue,
			})
		}
	}
	return extensions, ok
}

func EvaluateTopDir(o output.Bus, values map[string]*FlagValue) (dir string, ok bool) {
	if rawValue, userSet, err := GetString(o, values, SearchTopDir); err == nil {
		if file, err := os.Stat(rawValue); err != nil {
			o.WriteCanonicalError("The %s value, %q, cannot be used", SearchTopDirFlag,
				rawValue)
			o.Log(output.Error, "invalid directory", map[string]any{
				"error":          err,
				SearchTopDirFlag: rawValue,
				"user-set":       userSet,
			})
			o.WriteCanonicalError("Why?")
			if userSet {
				o.WriteCanonicalError("The value you specified is not a readable file.")
				o.WriteCanonicalError(
					"What to do:\nSpecify a value that is a readable file.")
			} else {
				o.WriteCanonicalError(
					"The currently configured value is not a readable file.")
				o.WriteCanonicalError("What to do:\n"+
					"Edit the configuration file or specify %s with a value that is a"+
					" readable file.", SearchTopDirFlag)
			}
		} else {
			if file.IsDir() {
				dir = rawValue
				ok = true
			} else {
				o.WriteCanonicalError("The %s value, %q, cannot be used", SearchTopDirFlag,
					rawValue)
				o.Log(output.Error, "the file is not a directory", map[string]any{
					SearchTopDirFlag: rawValue,
					"user-set":       userSet,
				})
				o.WriteCanonicalError("Why?")
				if userSet {
					o.WriteCanonicalError(
						"The value you specified is not the name of a directory.")
					o.WriteCanonicalError("What to do:\n" +
						"Specify a value that is the name of a directory.")
				} else {
					o.WriteCanonicalError(
						"The currently configured value is not the name of a directory.")
					o.WriteCanonicalError("What to do:\n"+
						"Edit the configuration file or specify %s with a value that is the"+
						" name of a directory.", SearchTopDirFlag)
				}
			}
		}
	}
	return
}

func EvaluateFilter(o output.Bus, values map[string]*FlagValue, flagName,
	nameAsFlag string) (filter *regexp.Regexp, ok, regexOk bool) {
	regexOk = true
	if rawValue, userSet, err := GetString(o, values, flagName); err == nil {
		if f, err := regexp.Compile(rawValue); err != nil {
			o.Log(output.Error, "the filter cannot be parsed as a regular expression",
				map[string]any{
					nameAsFlag: rawValue,
					"user-set": userSet,
					"error":    err,
				})
			o.WriteCanonicalError("the %s value %q cannot be used", nameAsFlag, rawValue)
			if userSet {
				o.WriteCanonicalError("Why?\n"+
					"The value of %s that you specified is not a valid regular expression: %v.",
					nameAsFlag, err)
				o.WriteCanonicalError("What to do:\n"+
					"Either try a different setting,"+
					" or omit setting %s and try the default value.", nameAsFlag)
			} else {
				o.WriteCanonicalError("Why?\n"+
					"The configured default value of %s is not a valid regular expression: %v.",
					nameAsFlag, err)
				o.WriteCanonicalError("What to do:\n"+
					"Either edit the defaults.yaml file containing the settings,"+
					" or explicitly set %s to a better value.", nameAsFlag)
			}
			regexOk = false
		} else {
			filter = f
			ok = true
		}
	}
	return
}

func (ss *SearchSettings) Filter(o output.Bus,
	originalArtists []*files.Artist) ([]*files.Artist, bool) {
	filteredArtists := make([]*files.Artist, 0, len(originalArtists))
	for _, originalArtist := range originalArtists {
		if ss.artistFilter.MatchString(originalArtist.Name()) && originalArtist.HasAlbums() {
			filteredArtist := originalArtist.Copy()
			for _, originalAlbum := range originalArtist.Albums() {
				if ss.albumFilter.MatchString(originalAlbum.Name()) &&
					originalAlbum.HasTracks() {
					filteredAlbum := originalAlbum.Copy(filteredArtist, false)
					for _, originalTrack := range originalAlbum.Tracks() {
						if ss.trackFilter.MatchString(originalTrack.CommonName()) {
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
	ok := len(filteredArtists) > 0
	if !ok {
		o.WriteCanonicalError("No music files remain after filtering.")
		o.WriteCanonicalError("Why?")
		o.WriteCanonicalError("After applying %s=%q, %s=%q, and %s=%q, no files remained",
			SearchArtistFilterFlag, ss.artistFilter, SearchAlbumFilterFlag, ss.albumFilter,
			SearchTrackFilterFlag, ss.trackFilter)
		o.WriteCanonicalError("What to do:\nUse less restrictive filter settings.")
		o.Log(output.Error, "no files remain after filtering", map[string]any{
			SearchArtistFilterFlag: ss.artistFilter,
			SearchAlbumFilterFlag:  ss.albumFilter,
			SearchTrackFilterFlag:  ss.trackFilter,
		})
	}
	return filteredArtists, ok
}

func (ss *SearchSettings) Load(o output.Bus) ([]*files.Artist, bool) {
	artistFiles, dirRead := ReadDirectory(o, ss.topDirectory)
	artists := make([]*files.Artist, 0, len(artistFiles))
	if dirRead {
		for _, artistFile := range artistFiles {
			if artistFile.IsDir() {
				artist := files.NewArtistFromFile(artistFile, ss.topDirectory)
				ss.addAlbums(o, artist)
				artists = append(artists, artist)
			}
		}
	}
	ok := len(artists) > 0
	if !ok {
		o.WriteCanonicalError(
			"No music files could be found using the specified parameters.")
		o.WriteCanonicalError("Why?")
		o.WriteCanonicalError("There were no directories found in %q (the %s value)",
			ss.topDirectory, SearchTopDirFlag)
		o.WriteCanonicalError("What to do:\n"+
			"Set %s to the path of a directory that contains artist directories",
			SearchTopDirFlag)
		o.Log(output.Error, "cannot find any artist directories", map[string]any{
			SearchTopDirFlag: ss.topDirectory,
		})
	}
	return artists, ok
}

func (ss *SearchSettings) addAlbums(o output.Bus, artist *files.Artist) {
	if albumFiles, artistDirRead := ReadDirectory(o, artist.Path()); artistDirRead {
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
	if trackFiles, ok := ReadDirectory(o, album.Path()); ok {
		for _, trackFile := range trackFiles {
			if extension, isTrack := ss.isValidTrackFile(trackFile); isTrack {
				if simpleName, trackNumber, valid := files.ParseTrackName(o,
					trackFile.Name(), album, extension); valid {
					album.AddTrack(files.NewTrack(album, trackFile.Name(), simpleName,
						trackNumber))
				}
			}
		}
	}

}

func (ss *SearchSettings) isValidTrackFile(file fs.DirEntry) (string, bool) {
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
