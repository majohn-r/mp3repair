package cmd

import (
	"fmt"
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"mp3repair/internal/files"
	"slices"
	"strings"

	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

// Details:

// A mp3 file usually contains ID3V1 metadata (gory details: https://id3.org/ID3v1),
// ID3V2 metadata (gory details: https://id3.org/id3v2.3.0), often both, in addition to
// the audio data. The integrity check reads each mp3 file's metadata and does the
// following:

// * Verify that the file name begins with the track number encoded in the TRCK (track
//   number/position in set) ID3V2 frame and the ID3V1 track field, and that the rest of
//   the file name matches the value encoded in the TIT2 (title/song name/content
//   description) ID3V2 frame and the ID3V1 song title field.
// * Verify that the containing album directory's name matches the TALB (album/movie/
//   show title) ID3V2 frame and the ID3V1 album field, and that all mp3 files in the
//   same album directory use the same album name in their ID3V2 and ID3V1 metadata.
// * Verify that the containing artist directory's name matches the TPE1 (lead artist/
//   lead performer/soloist/performing group) ID3V2 frame and the ID3V1 artist field, and
//   that all mp3 files within the same artist directory use the same artist name in their
//   ID3V2 and ID3V1 metadata.
// * Verify that all the mp3 files in the same album directory:
//   - contain the same TYER (year) ID3V2 frame and the same ID3V1 year field, and that
//     both agree.
//   - contain the same TCON (content type, aka genre) ID3V2 frame and the same ID3V1
//     genre field, and that the ID3V1 and ID3V2 genre agree as closely as possible.
//   - contain the same MCDI (music CD identifier) ID3V2 frame.

// About name matching:

//   File names and their corresponding metadata values cannot always be identical, as
//   some characters in the metadata may not be legal file name characters and end up
//   being replaced with, typically, punctuation characters. The check code takes those
//   differences into account. The following characters are known to be illegal in Windows
//   file names:
//   asterisk (*)      backward slash (\) colon (:)
//   forward slash (/) greater than (>)   less than (<)
//   question mark (?) quotation mark (") vertical bar (|)

// About ID3V1 and ID3V2 consistency:

//   The ID3V1 format is older (more primitive) than the ID3V2 format, and the check code
//   takes into account:
//   - ID3V1 fields can not encode multibyte characters; similar 8-bit characters are
//     used as needed.
//   - ID3V2 frames are variable-length; corresponding ID3V1 fields are fixed-length.
//   - ID3V1 encodes genre as a numeric code that indexes a table of genre names; ID3V2
//     encodes genre as free-form text.

const (
	CheckCommand       = "check"
	CheckEmpty         = "empty"
	CheckEmptyAbbr     = "e"
	CheckEmptyFlag     = "--" + CheckEmpty
	CheckFiles         = "files"
	CheckFilesAbbr     = "f"
	CheckFilesFlag     = "--" + CheckFiles
	CheckNumbering     = "numbering"
	CheckNumberingAbbr = "n"
	CheckNumberingFlag = "--" + CheckNumbering
)

var (
	// CheckCmd represents the check command
	CheckCmd = &cobra.Command{
		Use: CheckCommand + " [" + CheckEmptyFlag + "] [" +
			CheckFilesFlag + "] [" + CheckNumberingFlag + "] " + searchUsage,
		DisableFlagsInUseLine: true,
		Short: "" +
			"Inspects mp3 files and their directories and reports" + " problems",
		Long: fmt.Sprintf(
			"%q inspects mp3 files and their containing directories and reports any"+
				" problems detected", CheckCommand),
		Example: "" +
			CheckCommand + " " + CheckEmptyFlag + "\n" +
			"  reports empty artist and album directories\n" +
			CheckCommand + " " + CheckFilesFlag + "\n" +
			"  reads each mp3 file's metadata and reports any inconsistencies found\n" +
			CheckCommand + " " + CheckNumberingFlag + "\n" +
			"  reports errors in the track numbers of mp3 files",
		RunE: CheckRun,
	}
	CheckFlags = &SectionFlags{
		SectionName: CheckCommand,
		Details: map[string]*FlagDetails{
			CheckEmpty: {
				AbbreviatedName: CheckEmptyAbbr,
				Usage:           "report empty album and artist directories",
				ExpectedType:    BoolType,
				DefaultValue:    false,
			},
			CheckFiles: {
				AbbreviatedName: CheckFilesAbbr,
				Usage:           "report metadata/file inconsistencies",
				ExpectedType:    BoolType,
				DefaultValue:    false,
			},
			CheckNumbering: {
				AbbreviatedName: CheckNumberingAbbr,
				Usage:           "report missing track numbers and duplicated track numbering",
				ExpectedType:    BoolType,
				DefaultValue:    false,
			},
		},
	}
)

func CheckRun(cmd *cobra.Command, _ []string) error {
	exitError := cmdtoolkit.NewExitProgrammingError(CheckCommand)
	o := getBus()
	producer := cmd.Flags()
	values, eSlice := ReadFlags(producer, CheckFlags)
	searchSettings, searchFlagsOk := EvaluateSearchFlags(o, producer)
	if ProcessFlagErrors(o, eSlice) && searchFlagsOk {
		if cs, flagsOk := ProcessCheckFlags(o, values); flagsOk {
			details := map[string]any{
				CheckEmptyFlag:       cs.Empty.Value,
				"empty-user-set":     cs.Empty.UserSet,
				CheckFilesFlag:       cs.Files.Value,
				"files-user-set":     cs.Files.UserSet,
				CheckNumberingFlag:   cs.Numbering.Value,
				"numbering-user-set": cs.Numbering.UserSet,
			}
			for k, v := range searchSettings.Values() {
				details[k] = v
			}
			LogCommandStart(o, CheckCommand, details)
			exitError = cs.MaybeDoWork(o, searchSettings)
		}
	}
	return cmdtoolkit.ToErrorInterface(exitError)
}

type CheckSettings struct {
	Empty     CommandFlag[bool]
	Files     CommandFlag[bool]
	Numbering CommandFlag[bool]
}

func (cs *CheckSettings) MaybeDoWork(o output.Bus, ss *SearchSettings) (err *cmdtoolkit.ExitError) {
	err = cmdtoolkit.NewExitUserError(CheckCommand)
	if cs.HasWorkToDo(o) {
		err = cs.PerformChecks(o, ss.Load(o), ss)
	}
	return
}

func (cs *CheckSettings) PerformChecks(o output.Bus, artists []*files.Artist, ss *SearchSettings) (err *cmdtoolkit.ExitError) {
	err = cmdtoolkit.NewExitUserError(CheckCommand)
	if len(artists) != 0 {
		err = nil
		requests := CheckReportRequests{}
		concernedArtists := CreateConcernedArtists(artists)
		requests.ReportEmptyCheckResults = cs.PerformEmptyAnalysis(concernedArtists)
		requests.ReportNumberingCheckResults = cs.PerformNumberingAnalysis(concernedArtists)
		requests.ReportFilesCheckResults = cs.PerformFileAnalysis(o, concernedArtists, ss)
		for _, artist := range concernedArtists {
			artist.Rollup()
			artist.ToConsole(o)
		}
		cs.MaybeReportCleanResults(o, requests)
	}
	return
}

type CheckReportRequests struct {
	ReportEmptyCheckResults     bool
	ReportFilesCheckResults     bool
	ReportNumberingCheckResults bool
}

func (cs *CheckSettings) MaybeReportCleanResults(o output.Bus, requests CheckReportRequests) {
	if !requests.ReportEmptyCheckResults && cs.Empty.Value {
		o.WriteCanonicalConsole("Empty Folder Analysis: no empty folders found")
	}
	if !requests.ReportNumberingCheckResults && cs.Numbering.Value {
		o.WriteCanonicalConsole("Numbering Analysis: no missing or duplicate tracks found")
	}
	if !requests.ReportFilesCheckResults && cs.Files.Value {
		o.WriteCanonicalConsole("File Analysis: no inconsistencies found")
	}
}

func (cs *CheckSettings) PerformFileAnalysis(o output.Bus,
	concernedArtists []*ConcernedArtist, ss *SearchSettings) bool {
	foundConcerns := false
	if cs.Files.Value {
		artists := make([]*files.Artist, 0, len(concernedArtists))
		for _, cAr := range concernedArtists {
			artists = append(artists, cAr.Artist())
		}
		if filteredArtists := ss.Filter(o, artists); len(filteredArtists) != 0 {
			ReadMetadata(o, filteredArtists)
			for _, artist := range filteredArtists {
				for _, album := range artist.Albums {
					for _, track := range album.Tracks {
						concerns := track.ReportMetadataProblems()
						if found := RecordTrackFileConcerns(concernedArtists, track, concerns); found {
							foundConcerns = true
						}
					}
				}
			}
		}
	}
	return foundConcerns
}

func RecordTrackFileConcerns(artists []*ConcernedArtist, track *files.Track, concerns []string) (foundConcerns bool) {
	if len(concerns) > 0 {
		foundConcerns = true
		for _, cAr := range artists {
			if cT := cAr.Lookup(track); cT != nil {
				for _, s := range concerns {
					cT.AddConcern(FilesConcern, s)
				}
				break
			}
		}
	}
	return foundConcerns
}

func (cs *CheckSettings) PerformNumberingAnalysis(
	concernedArtists []*ConcernedArtist) bool {
	foundConcerns := false
	if cs.Numbering.Value {
		for _, cAr := range concernedArtists {
			for _, cAl := range cAr.Albums() {
				trackMap := map[int][]string{}
				maxTrack := len(cAl.Tracks())
				for _, cT := range cAl.Tracks() {
					track := cT.Track()
					trackNumber := track.Number
					trackMap[trackNumber] = append(trackMap[trackNumber], cT.name())
					if trackNumber > maxTrack {
						maxTrack = trackNumber
					}
				}
				concerns := GenerateNumberingConcerns(trackMap, maxTrack)
				if len(concerns) > 0 {
					foundConcerns = true
					for _, s := range concerns {
						cAl.AddConcern(NumberingConcern, s)
					}
				}
			}
		}
	}
	return foundConcerns
}

func GenerateNumberingConcerns(m map[int][]string, maxTrack int) []string {
	concerns := make([]string, 0, len(m)+1)
	var numbers []int
	// find duplicates
	for k, v := range m {
		if len(v) != 0 {
			numbers = append(numbers, k)
		}
		if len(v) > 1 {
			slices.Sort(v)
			formattedTracks := make([]string, 0, len(v)-1)
			for j := 0; j < len(v)-1; j++ {
				formattedTracks = append(formattedTracks, fmt.Sprintf("%q", v[j]))
			}
			finalTrack := fmt.Sprintf("%q", v[len(v)-1])
			concern := fmt.Sprintf("multiple tracks identified as track %d: %s and %s", k,
				strings.Join(formattedTracks, ", "), finalTrack)
			concerns = append(concerns, concern)
		}
	}
	// find missing track numbers
	slices.Sort(numbers)
	var missingNumbers []string
	if len(numbers) != 0 {
		if numbers[0] > 1 {
			missingNumbers = append(missingNumbers,
				NumberGap{Value1: 1, Value2: numbers[0] - 1}.GenerateMissingTrackNumbers())
		}
		for k := 0; k < len(numbers)-1; k++ {
			if numbers[k+1]-numbers[k] != 1 {
				missingNumbers = append(missingNumbers,
					NumberGap{
						Value1: numbers[k] + 1,
						Value2: numbers[k+1] - 1,
					}.GenerateMissingTrackNumbers())
			}
		}
		if numbers[len(numbers)-1] != maxTrack {
			missingNumbers = append(missingNumbers,
				NumberGap{
					Value1: numbers[len(numbers)-1] + 1,
					Value2: maxTrack,
				}.GenerateMissingTrackNumbers())
		}
	}
	if len(missingNumbers) != 0 {
		concern := fmt.Sprintf("missing tracks identified: %s",
			strings.Join(missingNumbers, ", "))
		concerns = append(concerns, concern)
	}
	return concerns
}

type NumberGap struct {
	Value1 int
	Value2 int
}

func (gap NumberGap) GenerateMissingTrackNumbers() string {
	if gap.Value1 == gap.Value2 {
		return fmt.Sprintf("%d", gap.Value1)
	}
	return fmt.Sprintf("%d-%d", min(gap.Value1, gap.Value2), max(gap.Value1, gap.Value2))
}

func (cs *CheckSettings) PerformEmptyAnalysis(concernedArtists []*ConcernedArtist) bool {
	emptyFoldersFound := false
	if cs.Empty.Value {
		for _, concernedArtist := range concernedArtists {
			if !concernedArtist.Artist().HasAlbums() {
				concernedArtist.AddConcern(EmptyConcern, "no albums found")
				emptyFoldersFound = true
				continue // next artist, please
			}
			for _, concernedAlbum := range concernedArtist.Albums() {
				if !concernedAlbum.Album().HasTracks() {
					concernedAlbum.AddConcern(EmptyConcern, "no tracks found")
					emptyFoldersFound = true
				}
			}
		}
	}
	return emptyFoldersFound
}

func (cs *CheckSettings) HasWorkToDo(o output.Bus) bool {
	if cs.Empty.Value || cs.Files.Value || cs.Numbering.Value {
		return true
	}
	userPartiallyAtFault := cs.Empty.UserSet || cs.Files.UserSet || cs.Numbering.UserSet
	o.WriteCanonicalError("No checks will be executed.\nWhy?\n")
	switch userPartiallyAtFault {
	case true:
		flagsUserSet := make([]string, 0, 3)
		flagsFromConfig := make([]string, 0, 3)
		switch cs.Empty.UserSet {
		case true:
			flagsUserSet = append(flagsUserSet, CheckEmptyFlag)
		case false:
			flagsFromConfig = append(flagsFromConfig, CheckEmptyFlag)
		}
		switch cs.Files.UserSet {
		case true:
			flagsUserSet = append(flagsUserSet, CheckFilesFlag)
		case false:
			flagsFromConfig = append(flagsFromConfig, CheckFilesFlag)
		}
		switch cs.Numbering.UserSet {
		case true:
			flagsUserSet = append(flagsUserSet, CheckNumberingFlag)
		case false:
			flagsFromConfig = append(flagsFromConfig, CheckNumberingFlag)
		}
		switch len(flagsFromConfig) {
		case 0:
			o.WriteCanonicalError("You explicitly set %s, %s, and %s false",
				CheckEmptyFlag, CheckFilesFlag, CheckNumberingFlag)
		default:
			o.WriteCanonicalError(
				"In addition to %s configured false, you explicitly set %s false",
				strings.Join(flagsFromConfig, " and "),
				strings.Join(flagsUserSet, " and "))
		}
	default:
		o.WriteCanonicalError("The flags %s, %s, and %s are all configured false",
			CheckEmptyFlag, CheckFilesFlag, CheckNumberingFlag)
	}
	o.WriteError("What to do:\n")
	o.WriteCanonicalError("Either:\n[1] Edit the configuration file so that at least one" +
		" of these flags is true, or\n[2] explicitly set at least one of these flags true on" +
		" the command line")
	return false
}

func ProcessCheckFlags(o output.Bus, values map[string]*CommandFlag[any]) (*CheckSettings, bool) {
	settings := &CheckSettings{}
	flagsOk := true // optimistic
	var flagErr error
	if settings.Empty, flagErr = GetBool(o, values, CheckEmpty); flagErr != nil {
		flagsOk = false
	}
	if settings.Files, flagErr = GetBool(o, values, CheckFiles); flagErr != nil {
		flagsOk = false
	}
	if settings.Numbering, flagErr = GetBool(o, values, CheckNumbering); flagErr != nil {
		flagsOk = false
	}
	return settings, flagsOk
}

func init() {
	RootCmd.AddCommand(CheckCmd)
	addDefaults(CheckFlags)
	o := getBus()
	c := getConfiguration()
	AddFlags(o, c, CheckCmd.Flags(), CheckFlags, SearchFlags)
}
