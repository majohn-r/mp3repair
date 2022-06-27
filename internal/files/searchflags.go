package files

import (
	"flag"
	"fmt"
	"mp3/internal"
	"os"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
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
	fkArguments           = "arguments"
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
	dereferencedArgs := make([]string, len(args))
	for i, arg := range args {
		dereferencedArgs[i] = internal.InterpretEnvVarReferences(arg)
	}
	sf.f.SetOutput(o.ErrorWriter())
	// note: Parse outputs errors to o.ErrorWriter*()
	if err := sf.f.Parse(dereferencedArgs); err != nil {
		o.Log(internal.ERROR, fmt.Sprintf("%v", err), map[string]interface{}{
			fkArguments: dereferencedArgs,
		})
		return nil, false
	}
	return sf.NewSearch()
}

// NewSearch validates the common search parameters and creates a Search
// instance based on them.
func (sf *SearchFlags) NewSearch() (s *Search, ok bool) {
	albumsFilter, artistsFilter, validated := sf.validate()
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

// TODO [#77] should use OutputBus for output
func (sf *SearchFlags) validateTopLevelDirectory() bool {
	if file, err := os.Stat(*sf.topDirectory); err != nil {
		fmt.Fprintf(os.Stderr, internal.USER_CANNOT_READ_TOPDIR, *sf.topDirectory, err)
		logrus.WithFields(logrus.Fields{
			fkTopDirFlag:      sf.topDirectory,
			internal.FK_ERROR: err,
		}).Warn(internal.LW_CANNOT_READ_DIRECTORY)
		return false
	} else {
		if file.IsDir() {
			return true
		} else {
			fmt.Fprintf(os.Stderr, internal.USER_TOPDIR_NOT_A_DIRECTORY, *sf.topDirectory)
			logrus.WithFields(logrus.Fields{
				fkTopDirFlag: sf.topDirectory,
			}).Warn(internal.LW_NOT_A_DIRECTORY)
			return false
		}
	}
}

// TODO [#77] should use OutputBus for error output
func (sf *SearchFlags) validateExtension() (ok bool) {
	ok = true
	if !strings.HasPrefix(*sf.fileExtension, ".") || strings.Contains(strings.TrimPrefix(*sf.fileExtension, "."), ".") {
		ok = false
		fmt.Fprintf(os.Stderr, internal.USER_EXTENSION_INVALID_FORMAT, *sf.fileExtension)
		logrus.WithFields(logrus.Fields{
			fkTargetExtensionFlag: sf.fileExtension,
		}).Warn(internal.LW_INVALID_EXTENSION_FORMAT)
	}
	var e error
	trackNameRegex, e = regexp.Compile("^\\d+[\\s-].+\\." + strings.TrimPrefix(*sf.fileExtension, ".") + "$")
	if e != nil {
		ok = false
		fmt.Fprintf(os.Stderr, internal.USER_EXTENSION_GARBLED, *sf.fileExtension, e)
		logrus.WithFields(logrus.Fields{
			fkTargetExtensionFlag: sf.fileExtension,
			internal.FK_ERROR:     e,
		}).Warn(internal.LW_GARBLED_EXTENSION)
	}
	return
}

// TODO [#77] should use OutputBus for error output
func validateRegexp(pattern, name string) (filter *regexp.Regexp, ok bool) {
	if f, err := regexp.Compile(pattern); err != nil {
		fmt.Fprintf(os.Stderr, internal.USER_FILTER_GARBLED, name, pattern, err)
		logrus.WithFields(logrus.Fields{
			name:              pattern,
			internal.FK_ERROR: err,
		}).Warn(internal.LW_GARBLED_FILTER)
	} else {
		filter = f
		ok = true
	}
	return
}

func (sf *SearchFlags) validate() (albumsFilter *regexp.Regexp, artistsFilter *regexp.Regexp, ok bool) {
	ok = true
	if !sf.validateTopLevelDirectory() {
		ok = false
	}
	if !sf.validateExtension() {
		ok = false
	}
	if filter, regexOk := validateRegexp(*sf.albumRegex, fkAlbumFilterFlag); !regexOk {
		ok = false
	} else {
		albumsFilter = filter
	}
	if filter, regexOk := validateRegexp(*sf.artistRegex, fkArtistFilterFlag); !regexOk {
		ok = false
	} else {
		artistsFilter = filter
	}
	return
}
