package files

import (
	"flag"
	"mp3/internal"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/majohn-r/output"
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
)

func reportBadDefault(o output.Bus, err error) {
	internal.ReportInvalidConfigurationData(o, defaultSectionName, err)
}

// NewSearchFlags are used by commands that use the common top directory, target
// extension, and album and artist filter regular expressions.
func NewSearchFlags(o output.Bus, c *internal.Configuration, fSet *flag.FlagSet) (*SearchFlags, bool) {
	return makeSearchFlags(o, c.SubConfiguration(defaultSectionName), fSet)
}

func makeSearchFlags(o output.Bus, configuration *internal.Configuration, fSet *flag.FlagSet) (*SearchFlags, bool) {
	var ok = true
	defTopDirectory, err := configuration.StringDefault(topDirectoryFlag, filepath.Join("%HOMEPATH%", "Music"))
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
func (sf *SearchFlags) ProcessArgs(o output.Bus, args []string) (s *Search, ok bool) {
	if ok = internal.ProcessArgs(o, sf.f, args); ok {
		s, ok = sf.NewSearch(o)
	}
	return
}

// NewSearch validates the common search parameters and creates a Search
// instance based on them.
func (sf *SearchFlags) NewSearch(o output.Bus) (s *Search, ok bool) {
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

func (sf *SearchFlags) validateTopLevelDirectory(o output.Bus) bool {
	file, err := os.Stat(*sf.topDirectory)
	if err != nil {
		o.WriteCanonicalError("The -topDir value you specified, %q, cannot be read: %v", *sf.topDirectory, err)
		internal.LogUnreadableDirectory(o, *sf.topDirectory, err)
		return false
	}
	if file.IsDir() {
		return true
	}
	o.WriteCanonicalError("The -topDir value you specified, %q, is not a directory", *sf.topDirectory)
	o.Log(output.Error, "the file is not a directory", map[string]any{
		"-" + topDirectoryFlag: *sf.topDirectory,
	})
	return false
}

func (sf *SearchFlags) validateExtension(o output.Bus) (ok bool) {
	ok = true
	if !strings.HasPrefix(*sf.fileExtension, ".") || strings.Contains(strings.TrimPrefix(*sf.fileExtension, "."), ".") {
		ok = false
		o.WriteCanonicalError("The -ext value you specified, %q, must contain exactly one '.' and '.' must be the first character", *sf.fileExtension)
		o.Log(output.Error, "the file extension must begin with '.' and contain no other '.' characters", map[string]any{
			"-" + fileExtensionFlag: *sf.fileExtension,
		})
	}
	var e error
	trackNameRegex, e = regexp.Compile("^\\d+[\\s-].+\\." + strings.TrimPrefix(*sf.fileExtension, ".") + "$")
	if e != nil {
		ok = false
		o.WriteCanonicalError("The -ext value you specified, %q, cannot be used for file matching: %v", *sf.fileExtension, e)
		o.Log(output.Error, "the file extension cannot be parsed as a regular expression", map[string]any{
			"-" + fileExtensionFlag: *sf.fileExtension,
			"error":                 e,
		})
	}
	return
}

func validateRegexp(o output.Bus, pattern, name string) (filter *regexp.Regexp, ok bool) {
	if f, err := regexp.Compile(pattern); err != nil {
		o.WriteCanonicalError("The %s filter value you specified, %q, cannot be used: %v", name, pattern, err)
		o.Log(output.Error, "the filter cannot be parsed as a regular expression", map[string]any{
			name:    pattern,
			"error": err,
		})
	} else {
		filter = f
		ok = true
	}
	return
}

func (sf *SearchFlags) validate(o output.Bus) (albumsFilter, artistsFilter *regexp.Regexp, ok bool) {
	ok = true
	if !sf.validateTopLevelDirectory(o) {
		ok = false
	}
	if !sf.validateExtension(o) {
		ok = false
	}
	if filter, regexOk := validateRegexp(o, *sf.albumRegex, "-"+albumRegexFlag); !regexOk {
		ok = false
	} else {
		albumsFilter = filter
	}
	if filter, regexOk := validateRegexp(o, *sf.artistRegex, "-"+artistRegexFlag); !regexOk {
		ok = false
	} else {
		artistsFilter = filter
	}
	return
}

// SearchDefaults returns the defaults for the search parameters
func SearchDefaults() (sectionName string, defaults map[string]any) {
	sectionName = defaultSectionName
	defaults = map[string]any{
		albumRegexFlag:    defaultRegex,
		artistRegexFlag:   defaultRegex,
		fileExtensionFlag: defaultFileExtension,
		topDirectoryFlag:  filepath.Join("%HOMEPATH%", "Music"),
	}
	return
}
