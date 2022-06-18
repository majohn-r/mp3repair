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
	albumRegexFlag    = "albumFilter"
	artistRegexFlag   = "artistFilter"
	defaultRegex      = ".*"
)

// NewSearchFlags are used by subCommands to use the common top directory,
// target extension, and album and artist filter regular expressions.
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
// TODO: return bool on error [#65]
func (sf *SearchFlags) ProcessArgs(writer io.Writer, args []string) *Search {
	dereferencedArgs := make([]string, len(args))
	for i, arg := range args {
		dereferencedArgs[i] = internal.InterpretEnvVarReferences(arg)
	}
	sf.f.SetOutput(writer)
	if err := sf.f.Parse(dereferencedArgs); err != nil {
		// TODO: [#65] should also present the error to the user!
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

// TODO: [#66] should use writer for error output
func (sf *SearchFlags) validateTopLevelDirectory() bool {
	if file, err := os.Stat(*sf.topDirectory); err != nil {
		fmt.Fprintf(os.Stderr, internal.USER_CANNOT_READ_TOPDIR, *sf.topDirectory, err)
		logrus.WithFields(logrus.Fields{
			internal.FK_TOP_DIR_FLAG: sf.topDirectory,
			internal.FK_ERROR:        err,
		}).Warn(internal.LW_CANNOT_READ_DIRECTORY)
		return false
	} else {
		if file.IsDir() {
			return true
		} else {
			fmt.Fprintf(os.Stderr, internal.USER_TOPDIR_NOT_A_DIRECTORY, *sf.topDirectory)
			logrus.WithFields(logrus.Fields{
				internal.FK_TOP_DIR_FLAG: sf.topDirectory,
			}).Warn(internal.LW_NOT_A_DIRECTORY)
			return false
		}
	}
}

// TODO: [#66] should use writer for error output
func (sf *SearchFlags) validateExtension() (valid bool) {
	valid = true
	if !strings.HasPrefix(*sf.fileExtension, ".") || strings.Contains(strings.TrimPrefix(*sf.fileExtension, "."), ".") {
		valid = false
		fmt.Fprintf(os.Stderr, internal.USER_EXTENSION_INVALID_FORMAT, *sf.fileExtension)
		logrus.WithFields(logrus.Fields{
			internal.FK_FILE_EXTENSION_FLAG: sf.fileExtension,
		}).Warn(internal.LW_INVALID_EXTENSION_FORMAT)
	}
	var e error
	trackNameRegex, e = regexp.Compile("^\\d+[\\s-].+\\." + strings.TrimPrefix(*sf.fileExtension, ".") + "$")
	if e != nil {
		valid = false
		fmt.Fprintf(os.Stderr, internal.USER_EXTENSION_GARBLED, *sf.fileExtension, e)
		logrus.WithFields(logrus.Fields{
			internal.FK_FILE_EXTENSION_FLAG: sf.fileExtension,
			internal.FK_ERROR:               e,
		}).Warn(internal.LW_GARBLED_EXTENSION)
	}
	return
}

// TODO: [#66] should use writer for error output
// TODO: [#67] badRegex is an anti-pattern - use ok instead
func validateRegexp(pattern, name string) (filter *regexp.Regexp, badRegex bool) {
	if f, err := regexp.Compile(pattern); err != nil {
		fmt.Fprintf(os.Stderr, internal.USER_FILTER_GARBLED, name, pattern, err)
		logrus.WithFields(logrus.Fields{
			name:              pattern,
			internal.FK_ERROR: err,
		}).Warn(internal.LW_GARBLED_FILTER)
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
	if filter, b := validateRegexp(*sf.albumRegex, internal.FK_ALBUM_FILTER_FLAG); b {
		problemsExist = true
	} else {
		albumsFilter = filter
	}
	if filter, b := validateRegexp(*sf.artistRegex, internal.FK_ARTIST_FILTER_FLAG); b {
		problemsExist = true
	} else {
		artistsFilter = filter
	}
	return
}
