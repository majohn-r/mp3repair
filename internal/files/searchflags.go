package files

import (
	"flag"
	"fmt"
	"io"
	"mp3/internal"
	"os"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
	topDirectoryFlag  = "topDir"
	fileExtensionFlag = "ext"
	albumRegexFlag    = "albums"
	artistRegexFlag   = "artists"
	defaultRegex      = ".*"
)

// NewSearchFlags are used by subCommands to use the common top directory,
// target extension, and album and artist filter regular expressions.
func NewSearchFlags(v *viper.Viper, fSet *flag.FlagSet) *SearchFlags {
	subViper := internal.SafeSubViper(v, "common")
	return &SearchFlags{
		f: fSet,
		topDirectory: fSet.String(topDirectoryFlag,
			internal.GetStringDefault(subViper, topDirectoryFlag, "$HOMEPATH/Music"),
			"top directory in which to look for music files"),
		fileExtension: fSet.String(fileExtensionFlag,
			internal.GetStringDefault(subViper, fileExtensionFlag, defaultFileExtension),
			"extension for music files"),
		albumRegex: fSet.String(albumRegexFlag,
			internal.GetStringDefault(subViper, albumRegexFlag, defaultRegex),
			"regular expression of albums to select"),
		artistRegex: fSet.String(artistRegexFlag,
			internal.GetStringDefault(subViper, artistRegexFlag, defaultRegex),
			"regular expression of artists to select"),
	}
}

// ProcessArgs consumes the command line arguments.
func (sf *SearchFlags) ProcessArgs(writer io.Writer, args []string) *Search {
	dereferencedArgs := make([]string, len(args))
	for i, arg := range args {
		dereferencedArgs[i] = internal.InterpretEnvVarReferences(arg)
	}
	sf.f.SetOutput(writer)
	if err := sf.f.Parse(dereferencedArgs); err != nil {
		logrus.Error(err)
		return nil
	}
	return sf.NewSearch()
}

// NewSearch validates the common search parameters and creates a Search
// instance based on them.
func (sf *SearchFlags) NewSearch() (s *Search) {
	albumsFilter, artistsFilter, problemsExist := sf.validate()
	if !problemsExist {
		s = &Search{
			topDirectory:    *sf.topDirectory,
			targetExtension: *sf.fileExtension,
			albumFilter:     albumsFilter,
			artistFilter:    artistsFilter,
		}
	}
	return
}

func (sf *SearchFlags) validateTopLevelDirectory() bool {
	if file, err := os.Stat(*sf.topDirectory); err != nil {
		fmt.Fprintf(os.Stderr, internal.USER_CANNOT_READ_TOPDIR, *sf.topDirectory, err)
		logrus.WithFields(logrus.Fields{internal.LOG_DIRECTORY: sf.topDirectory, internal.LOG_ERROR: err}).Error(internal.LOG_CANNOT_READ_DIRECTORY)
		return false
	} else {
		if file.IsDir() {
			return true
		} else {
			fmt.Fprintf(os.Stderr, internal.USER_TOPDIR_NOT_A_DIRECTORY, *sf.topDirectory)
			logrus.WithFields(logrus.Fields{internal.LOG_DIRECTORY: sf.topDirectory, internal.LOG_FLAG: "-topDir"}).Error(internal.LOG_NOT_A_DIRECTORY)
			return false
		}
	}
}

func (sf *SearchFlags) validateExtension() (valid bool) {
	valid = true
	if !strings.HasPrefix(*sf.fileExtension, ".") || strings.Contains(strings.TrimPrefix(*sf.fileExtension, "."), ".") {
		valid = false
		fmt.Fprintf(os.Stderr, internal.USER_EXTENSION_INVALID_FORMAT, *sf.fileExtension)
		logrus.WithFields(logrus.Fields{internal.LOG_EXTENSION: sf.fileExtension, internal.LOG_FLAG: "-ext"}).Error(internal.LOG_INVALID_EXTENSION_FORMAT)
	}
	var e error
	trackNameRegex, e = regexp.Compile("^\\d+[\\s-].+\\." + strings.TrimPrefix(*sf.fileExtension, ".") + "$")
	if e != nil {
		valid = false
		fmt.Fprintf(os.Stderr, internal.USER_EXTENSION_GARBLED, *sf.fileExtension, e)
		logrus.WithFields(logrus.Fields{internal.LOG_EXTENSION: sf.fileExtension, internal.LOG_FLAG: "-ext", internal.LOG_ERROR: e}).Error(internal.LOG_GARBLED_EXTENSION)
	}
	return
}

func validateRegexp(pattern, name string) (filter *regexp.Regexp, badRegex bool) {
	if f, err := regexp.Compile(pattern); err != nil {
		fmt.Fprintf(os.Stderr, internal.USER_FILTER_GARBLED, name, pattern, err)
		logrus.WithFields(logrus.Fields{internal.LOG_FILTER: name, internal.LOG_VALUE: pattern, internal.LOG_ERROR: err}).Error(internal.LOG_GARBLED_FILTER)
		badRegex = true
	} else {
		filter = f
	}
	return
}

func (sf *SearchFlags) validate() (albumsFilter *regexp.Regexp, artistsFilter *regexp.Regexp, problemsExist bool) {
	if !sf.validateTopLevelDirectory() {
		problemsExist = true
	}
	if !sf.validateExtension() {
		problemsExist = true
	}
	if filter, b := validateRegexp(*sf.albumRegex, "-albums"); b {
		problemsExist = true
	} else {
		albumsFilter = filter
	}
	if filter, b := validateRegexp(*sf.artistRegex, "-artists"); b {
		problemsExist = true
	} else {
		artistsFilter = filter
	}
	return
}
