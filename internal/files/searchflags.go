package files

import (
	"flag"
	"mp3/internal"
	"os"
	"regexp"
	"strings"
)

// SearchFlags defines the common flags used to specify how the top directory,
// file extension, and album and artist filters are defined and validated.
type SearchFlags struct {
	f             *flag.FlagSet
	topDirectory  *string
	fileExtension *string
	albumRegex    *string
	artistRegex   *string
}

const (
	albumRegexFlag        = "albumFilter"
	artistRegexFlag       = "artistFilter"
	defaultRegex          = ".*"
	fileExtensionFlag     = "ext"
	fkAlbumFilterFlag     = "-" + albumRegexFlag
	fkArtistFilterFlag    = "-" + artistRegexFlag
	fkTargetExtensionFlag = "-" + fileExtensionFlag
	fkTopDirFlag          = "-" + topDirectoryFlag
	topDirectoryFlag      = "topDir"
)

// NewSearchFlags are used by commands to use the common top directory, target
// extension, and album and artist filter regular expressions.
func NewSearchFlags(c *internal.Configuration, fSet *flag.FlagSet) *SearchFlags {
	configuration := c.SubConfiguration("common")
	return &SearchFlags{
		f: fSet,
		topDirectory: fSet.String(topDirectoryFlag,
			configuration.StringDefault(topDirectoryFlag, "$HOMEPATH/Music"),
			"top directory in which to look for music files"),
		fileExtension: fSet.String(fileExtensionFlag,
			configuration.StringDefault(fileExtensionFlag, defaultFileExtension),
			"extension for music files"),
		albumRegex: fSet.String(albumRegexFlag,
			configuration.StringDefault(albumRegexFlag, defaultRegex),
			"regular expression of albums to select"),
		artistRegex: fSet.String(artistRegexFlag,
			configuration.StringDefault(artistRegexFlag, defaultRegex),
			"regular expression of artists to select"),
	}
}

// ProcessArgs consumes the command line arguments.
func (sf *SearchFlags) ProcessArgs(o internal.OutputBus, args []string) (s *Search, ok bool) {
	if ok = internal.ProcessArgs(o, sf.f, args); ok {
		s, ok = sf.NewSearch(o)
	}
	return
}

// NewSearch validates the common search parameters and creates a Search
// instance based on them.
func (sf *SearchFlags) NewSearch(o internal.OutputBus) (s *Search, ok bool) {
	albumsFilter, artistsFilter, validated := sf.validate(o)
	if validated {
		s = &Search{
			topDirectory:    *sf.topDirectory,
			targetExtension: *sf.fileExtension,
			albumFilter:     albumsFilter,
			artistFilter:    artistsFilter,
		}
		ok = true
	}
	return
}

func (sf *SearchFlags) validateTopLevelDirectory(o internal.OutputBus) bool {
	if file, err := os.Stat(*sf.topDirectory); err != nil {
		o.WriteError(internal.USER_CANNOT_READ_TOPDIR, *sf.topDirectory, err)
		o.LogWriter().Warn(internal.LW_CANNOT_READ_DIRECTORY, map[string]interface{}{
			fkTopDirFlag:      *sf.topDirectory,
			internal.FK_ERROR: err,
		})
		return false
	} else {
		if file.IsDir() {
			return true
		} else {
			o.WriteError(internal.USER_TOPDIR_NOT_A_DIRECTORY, *sf.topDirectory)
			o.LogWriter().Warn(internal.LW_NOT_A_DIRECTORY, map[string]interface{}{
				fkTopDirFlag: *sf.topDirectory,
			})
			return false
		}
	}
}

func (sf *SearchFlags) validateExtension(o internal.OutputBus) (ok bool) {
	ok = true
	if !strings.HasPrefix(*sf.fileExtension, ".") || strings.Contains(strings.TrimPrefix(*sf.fileExtension, "."), ".") {
		ok = false
		o.WriteError(internal.USER_EXTENSION_INVALID_FORMAT, *sf.fileExtension)
		o.LogWriter().Warn(internal.LW_INVALID_EXTENSION_FORMAT, map[string]interface{}{
			fkTargetExtensionFlag: *sf.fileExtension,
		})
	}
	var e error
	trackNameRegex, e = regexp.Compile("^\\d+[\\s-].+\\." + strings.TrimPrefix(*sf.fileExtension, ".") + "$")
	if e != nil {
		ok = false
		o.WriteError(internal.USER_EXTENSION_GARBLED, *sf.fileExtension, e)
		o.LogWriter().Warn(internal.LW_GARBLED_EXTENSION, map[string]interface{}{
			fkTargetExtensionFlag: *sf.fileExtension,
			internal.FK_ERROR:     e,
		})
	}
	return
}

func validateRegexp(o internal.OutputBus, pattern string, name string) (filter *regexp.Regexp, ok bool) {
	if f, err := regexp.Compile(pattern); err != nil {
		o.WriteError(internal.USER_FILTER_GARBLED, name, pattern, err)
		o.LogWriter().Warn(internal.LW_GARBLED_FILTER, map[string]interface{}{
			name:              pattern,
			internal.FK_ERROR: err,
		})
	} else {
		filter = f
		ok = true
	}
	return
}

func (sf *SearchFlags) validate(o internal.OutputBus) (albumsFilter *regexp.Regexp, artistsFilter *regexp.Regexp, ok bool) {
	ok = true
	if !sf.validateTopLevelDirectory(o) {
		ok = false
	}
	if !sf.validateExtension(o) {
		ok = false
	}
	if filter, regexOk := validateRegexp(o, *sf.albumRegex, fkAlbumFilterFlag); !regexOk {
		ok = false
	} else {
		albumsFilter = filter
	}
	if filter, regexOk := validateRegexp(o, *sf.artistRegex, fkArtistFilterFlag); !regexOk {
		ok = false
	} else {
		artistsFilter = filter
	}
	return
}
