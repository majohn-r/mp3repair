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

type FileFlags struct {
	f             *flag.FlagSet
	topDirectory  *string
	fileExtension *string
	albumRegex    *string
	artistRegex   *string
}

func NewFileFlags(fSet *flag.FlagSet) *FileFlags {
	return &FileFlags{
		f:             fSet,
		topDirectory:  fSet.String("topDir", filepath.Join(internal.HomePath, "Music"), "top directory in which to look for music files"),
		fileExtension: fSet.String("ext", DefaultFileExtension, "extension for music files"),
		albumRegex:    fSet.String("albums", ".*", "regular expression of albums to select"),
		artistRegex:   fSet.String("artists", ".*", "regular expression of artists to select"),
	}
}

func (ff *FileFlags) ProcessArgs(writer io.Writer, args []string) *Search {
	ff.f.SetOutput(writer)
	if err := ff.f.Parse(args); err != nil {
		logrus.Error(err)
		return nil
	}
	return ff.NewSearch()
}

func (ff *FileFlags) NewSearch() (s *Search) {
	albumsFilter, artistsFilter, problemsExist := ff.validateSearchParameters()
	if !problemsExist {
		s = &Search{
			topDirectory:    *ff.topDirectory,
			targetExtension: *ff.fileExtension,
			albumFilter:     albumsFilter,
			artistFilter:    artistsFilter,
		}
	}
	return
}

func validateTopLevelDirectory(dir string) bool {
	if file, err := os.Stat(dir); err != nil {
		fmt.Fprintf(os.Stderr, "error checking top directory %q: %v\n", dir, err)
		logrus.WithFields(logrus.Fields{"directory": dir, "error": err}).Error("error checking top directory")
		return false
	} else {
		if file.IsDir() {
			return true
		} else {
			fmt.Fprintf(os.Stderr, "top directory %q is not actually a directory\n", dir)
			logrus.WithFields(logrus.Fields{"directory": dir}).Error("top directory is not a directory")
			return false
		}
	}
}

func validateExtension(ext string) (valid bool) {
	valid = true
	if !strings.HasPrefix(ext, ".") || strings.Contains(strings.TrimPrefix(ext, "."), ".") {
		valid = false
		fmt.Fprintf(os.Stderr, "the extension %q must contain exactly one '.' and '.' must be the first character\n", ext)
		logrus.WithFields(logrus.Fields{"extension": ext}).Error("the file extension must contain exactly one '.' and '.' must be the first character")
	}
	var e error
	trackNameRegex, e = regexp.Compile("^\\d+[\\s-].+\\." + strings.TrimPrefix(ext, ".") + "$")
	if e != nil {
		valid = false
		fmt.Fprintf(os.Stderr, "%q is not a valid extension: %v\n", ext, e)
		logrus.WithFields(logrus.Fields{"extension": ext, "error": e}).Error("the extension is not valid")
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

func (ff *FileFlags) validateSearchParameters() (albumsFilter *regexp.Regexp, artistsFilter *regexp.Regexp, problemsExist bool) {
	if !validateTopLevelDirectory(*ff.topDirectory) {
		problemsExist = true
	}
	if !validateExtension(*ff.fileExtension) {
		problemsExist = true
	}
	if filter, b := validateRegexp(*ff.albumRegex, "album"); b {
		problemsExist = true
	} else {
		albumsFilter = filter
	}
	if filter, b := validateRegexp(*ff.artistRegex, "artist"); b {
		problemsExist = true
	} else {
		artistsFilter = filter
	}
	return
}
