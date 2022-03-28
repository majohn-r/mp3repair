package files

import (
	"flag"
	"fmt"
	"io"
	"mp3/internal"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

type SearchFlags struct {
	f             *flag.FlagSet
	topDirectory  *string
	fileExtension *string
	albumRegex    *string
	artistRegex   *string
}

func NewSearchFlags(fSet *flag.FlagSet) *SearchFlags {
	return &SearchFlags{
		f:             fSet,
		topDirectory:  fSet.String("topDir", filepath.Join(internal.HomePath, "Music"), "top directory in which to look for music files"),
		fileExtension: fSet.String("ext", DefaultFileExtension, "extension for music files"),
		albumRegex:    fSet.String("albums", ".*", "regular expression of albums to select"),
		artistRegex:   fSet.String("artists", ".*", "regular expression of artists to select"),
	}
}

func (sf *SearchFlags) ProcessArgs(writer io.Writer, args []string) *Search {
	sf.f.SetOutput(writer)
	if err := sf.f.Parse(args); err != nil {
		logrus.Error(err)
		return nil
	}
	return sf.NewSearch()
}

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
		fmt.Fprintf(os.Stderr, "error checking top directory %q: %v\n", *sf.topDirectory, err)
		logrus.WithFields(logrus.Fields{"directory": sf.topDirectory, "error": err}).Error("error checking top directory")
		return false
	} else {
		if file.IsDir() {
			return true
		} else {
			fmt.Fprintf(os.Stderr, "top directory %q is not actually a directory\n", *sf.topDirectory)
			logrus.WithFields(logrus.Fields{"directory": sf.topDirectory}).Error("top directory is not a directory")
			return false
		}
	}
}

func (sf *SearchFlags) validateExtension() (valid bool) {
	valid = true
	if !strings.HasPrefix(*sf.fileExtension, ".") || strings.Contains(strings.TrimPrefix(*sf.fileExtension, "."), ".") {
		valid = false
		fmt.Fprintf(os.Stderr, "the extension %q must contain exactly one '.' and '.' must be the first character\n", *sf.fileExtension)
		logrus.WithFields(logrus.Fields{"extension": sf.fileExtension}).Error("the file extension must contain exactly one '.' and '.' must be the first character")
	}
	var e error
	trackNameRegex, e = regexp.Compile("^\\d+[\\s-].+\\." + strings.TrimPrefix(*sf.fileExtension, ".") + "$")
	if e != nil {
		valid = false
		fmt.Fprintf(os.Stderr, "%q is not a valid extension: %v\n", *sf.fileExtension, e)
		logrus.WithFields(logrus.Fields{"extension": sf.fileExtension, "error": e}).Error("the extension is not valid")
	}
	return
}

func validateRegexp(pattern, name string) (filter *regexp.Regexp, badRegex bool) {
	if f, err := regexp.Compile(pattern); err != nil {
		fmt.Fprintf(os.Stderr, "%s filter is invalid: %v\n", name, err)
		logrus.WithFields(logrus.Fields{"filterName": name, "error": err}).Error("the filter is invalid")
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
	if filter, b := validateRegexp(*sf.albumRegex, "album"); b {
		problemsExist = true
	} else {
		albumsFilter = filter
	}
	if filter, b := validateRegexp(*sf.artistRegex, "artist"); b {
		problemsExist = true
	} else {
		artistsFilter = filter
	}
	return
}
