package files

import (
	"flag"
	"mp3/internal"
	"os"
	"path/filepath"
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
	defaultSectionName = "common"

	albumRegexFlag    = "albumFilter"
	artistRegexFlag   = "artistFilter"
	fileExtensionFlag = "ext"
	topDirectoryFlag  = "topDir"

	defaultRegex = ".*"

	fkAlbumFilterFlag     = "-" + albumRegexFlag
	fkArtistFilterFlag    = "-" + artistRegexFlag
	fkTargetExtensionFlag = "-" + fileExtensionFlag
	fkTopDirFlag          = "-" + topDirectoryFlag
)

func reportBadDefault(o internal.OutputBus, err error) {
	o.WriteError(internal.USER_CONFIGURATION_FILE_INVALID, internal.DefaultConfigFileName, defaultSectionName, err)
	o.LogWriter().Error(internal.LE_INVALID_CONFIGURATION_DATA, map[string]interface{}{
		internal.FK_SECTION: defaultSectionName,
		internal.FK_ERROR:   err,
	})
}

// NewSearchFlags are used by commands that use the common top directory, target
// extension, and album and artist filter regular expressions.
func NewSearchFlags(o internal.OutputBus, c *internal.Configuration, fSet *flag.FlagSet) (*SearchFlags, bool) {
	return makeSearchFlags(o, c.SubConfiguration(defaultSectionName), fSet)
}

func makeSearchFlags(o internal.OutputBus, configuration *internal.Configuration, fSet *flag.FlagSet) (*SearchFlags, bool) {
	var ok = true
	defTopDirectory, err := configuration.StringDefault(topDirectoryFlag, filepath.Join("$HOMEPATH", "Music"))
	if err != nil {
		reportBadDefault(o, err)
		ok = false
	}
	defFileExtension, err := configuration.StringDefault(fileExtensionFlag, defaultFileExtension)
	if err != nil {
		reportBadDefault(o, err)
		ok = false
	}
	defAlbumRegex, err := configuration.StringDefault(albumRegexFlag, defaultRegex)
	if err != nil {
		reportBadDefault(o, err)
		ok = false
	}
	defArtistRegex, err := configuration.StringDefault(artistRegexFlag, defaultRegex)
	if err != nil {
		reportBadDefault(o, err)
		ok = false
	}
	if ok {
		topDirUsage := internal.DecorateStringFlagUsage("top `directory` specifying where to find music files", defTopDirectory)
		extUsage := internal.DecorateStringFlagUsage("`extension` identifying music files", defFileExtension)
		albumUsage := internal.DecorateStringFlagUsage("`regular expression` specifying which albums to select", defAlbumRegex)
		artistUsage := internal.DecorateStringFlagUsage("`regular expression` specifying which artists to select", defArtistRegex)
		return &SearchFlags{
			f:             fSet,
			topDirectory:  fSet.String(topDirectoryFlag, defTopDirectory, topDirUsage),
			fileExtension: fSet.String(fileExtensionFlag, defFileExtension, extUsage),
			albumRegex:    fSet.String(albumRegexFlag, defAlbumRegex, albumUsage),
			artistRegex:   fSet.String(artistRegexFlag, defArtistRegex, artistUsage),
		}, true
	}
	return nil, false
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
		o.LogWriter().Error(internal.LE_CANNOT_READ_DIRECTORY, map[string]interface{}{
			fkTopDirFlag:      *sf.topDirectory,
			internal.FK_ERROR: err,
		})
		return false
	} else {
		if file.IsDir() {
			return true
		} else {
			o.WriteError(internal.USER_TOPDIR_NOT_A_DIRECTORY, *sf.topDirectory)
			o.LogWriter().Error(internal.LE_NOT_A_DIRECTORY, map[string]interface{}{
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
		o.LogWriter().Error(internal.LE_INVALID_EXTENSION_FORMAT, map[string]interface{}{
			fkTargetExtensionFlag: *sf.fileExtension,
		})
	}
	var e error
	trackNameRegex, e = regexp.Compile("^\\d+[\\s-].+\\." + strings.TrimPrefix(*sf.fileExtension, ".") + "$")
	if e != nil {
		ok = false
		o.WriteError(internal.USER_EXTENSION_GARBLED, *sf.fileExtension, e)
		o.LogWriter().Error(internal.LE_GARBLED_EXTENSION, map[string]interface{}{
			fkTargetExtensionFlag: *sf.fileExtension,
			internal.FK_ERROR:     e,
		})
	}
	return
}

func validateRegexp(o internal.OutputBus, pattern string, name string) (filter *regexp.Regexp, ok bool) {
	if f, err := regexp.Compile(pattern); err != nil {
		o.WriteError(internal.USER_FILTER_GARBLED, name, pattern, err)
		o.LogWriter().Error(internal.LE_GARBLED_FILTER, map[string]interface{}{
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

func SearchDefaults() (string, map[string]any) {
	return defaultSectionName, map[string]any{
		albumRegexFlag:    defaultRegex,
		artistRegexFlag:   defaultRegex,
		fileExtensionFlag: defaultFileExtension,
		topDirectoryFlag:  filepath.Join("$HOMEPATH", "Music"),
	}
}
