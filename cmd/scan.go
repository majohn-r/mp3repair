package cmd

import (
	"fmt"
	"mp3repair/internal/files"
	"slices"
	"strings"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"

	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

// Details:

// A mp3 file usually contains ID3V1 metadata (gory details: https://id3.org/ID3v1), ID3V2 metadata (gory details:
// https://id3.org/id3v2.3.0), often both, in addition to the audio data. The integrity scan reads each mp3 file's
// metadata and does the following:

// * Verify that the file name begins with the track number encoded in the TRCK (track number/position in set) ID3V2
//   frame and the ID3V1 track field
// * Verify that the rest of the file name matches the value encoded in the TIT2 (title/song name/content description)
//   ID3V2 frame and the ID3V1 song title field.
// * Verify that the containing album directory's name matches the TALB (album/movie/ show title) ID3V2 frame and the
//   ID3V1 album field, and that all mp3 files in the same album directory use the same album name in their ID3V2 and
//   ID3V1 metadata.
// * Verify that the containing artist directory's name matches the TPE1 (lead artist/ lead performer/soloist/performing
//   group) ID3V2 frame and the ID3V1 artist field, and that all mp3 files within the same artist directory use the same
//   artist name in their ID3V2 and ID3V1 metadata.
// * Verify that all the mp3 files in the same album directory:
//   - contain the same TYER (year) ID3V2 frame and the same ID3V1 year field, and that both agree.
//   - contain the same TCON (content type, aka genre) ID3V2 frame and the same ID3V1 genre field, and that the ID3V1
//     and ID3V2 genre agree as closely as possible.
//   - contain the same MCDI (music CD identifier) ID3V2 frame.

// About name matching:

//   File names and their corresponding metadata values cannot always be identical, as some characters in the metadata
//   may not be legal file name characters and end up being replaced with, typically, punctuation characters. The scan
//   code takes those differences into account. The following characters are known to be illegal in Windows
//   file names:
//   asterisk (*)      backward slash (\) colon (:)
//   forward slash (/) greater than (>)   less than (<)
//   question mark (?) quotation mark (") vertical bar (|)

// About ID3V1 and ID3V2 consistency:

//   The ID3V1 format is older (more primitive) than the ID3V2 format, and the scan code takes into account:
//   - ID3V1 fields can not encode multibyte characters; similar 8-bit characters are used as needed.
//   - ID3V2 frames are variable-length; corresponding ID3V1 fields are fixed-length.
//   - ID3V1 encodes genre as a numeric code that indexes a table of genre names; ID3V2 encodes genre as free-form text.

const (
	scanCommand       = "scan"
	scanEmpty         = "empty"
	scanEmptyAbbr     = "e"
	scanEmptyFlag     = "--" + scanEmpty
	scanFiles         = "files"
	scanFilesAbbr     = "f"
	scanFilesFlag     = "--" + scanFiles
	scanNumbering     = "numbering"
	scanNumberingAbbr = "n"
	scanNumberingFlag = "--" + scanNumbering
)

var (
	scanCmd = &cobra.Command{
		Use: scanCommand + " [" + scanEmptyFlag + "] [" + scanFilesFlag + "] [" + scanNumberingFlag + "] " +
			searchUsage + " " + ioUsage,
		DisableFlagsInUseLine: true,
		Short: "" +
			"Inspects mp3 files and their directories and reports" + " problems",
		Long: fmt.Sprintf(
			"%q inspects mp3 files and their containing directories and reports any"+
				" problems detected", scanCommand),
		Example: "" +
			scanCommand + " " + scanEmptyFlag + "\n" +
			"  reports empty artist and album directories\n" +
			scanCommand + " " + scanFilesFlag + "\n" +
			"  reads each mp3 file's metadata and reports any inconsistencies found\n" +
			scanCommand + " " + scanNumberingFlag + "\n" +
			"  reports errors in the track numbers of mp3 files",
		RunE: scanRun,
	}
	scanFlags = &cmdtoolkit.FlagSet{
		Name: scanCommand,
		Details: map[string]*cmdtoolkit.FlagDetails{
			scanEmpty: {
				AbbreviatedName: scanEmptyAbbr,
				Usage:           "report empty album and artist directories",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
			scanFiles: {
				AbbreviatedName: scanFilesAbbr,
				Usage:           "report metadata/file inconsistencies",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
			scanNumbering: {
				AbbreviatedName: scanNumberingAbbr,
				Usage:           "report missing track numbers and duplicated track numbering",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
		},
	}
)

func scanRun(cmd *cobra.Command, _ []string) error {
	exitError := cmdtoolkit.NewExitProgrammingError(scanCommand)
	o := getBus()
	producer := cmd.Flags()
	values, eSlice := cmdtoolkit.ReadFlags(producer, scanFlags)
	ss, searchFlagsOk := evaluateSearchFlags(o, producer)
	ios, ioFlagsOk := evaluateIOFlags(o, producer)
	if cmdtoolkit.ProcessFlagErrors(o, eSlice) && searchFlagsOk && ioFlagsOk {
		if cs, flagsOk := processScanFlags(o, values); flagsOk {
			exitError = cs.maybeDoWork(o, ss, ios)
		}
	}
	return cmdtoolkit.ToErrorInterface(exitError)
}

type scanSettings struct {
	empty     cmdtoolkit.CommandFlag[bool]
	files     cmdtoolkit.CommandFlag[bool]
	numbering cmdtoolkit.CommandFlag[bool]
}

func (scanSets *scanSettings) maybeDoWork(o output.Bus, ss *searchSettings, ios *ioSettings) (err *cmdtoolkit.ExitError) {
	err = cmdtoolkit.NewExitUserError(scanCommand)
	if scanSets.hasWorkToDo(o) {
		err = scanSets.performScans(o, ss.load(o), ss, ios)
	}
	return
}

func (scanSets *scanSettings) performScans(
	o output.Bus,
	artists []*files.Artist,
	ss *searchSettings,
	ios *ioSettings,
) (err *cmdtoolkit.ExitError) {
	err = cmdtoolkit.NewExitUserError(scanCommand)
	if len(artists) != 0 {
		err = nil
		requests := scanReportRequests{}
		concernedArtists := createConcernedArtists(artists)
		requests.reportEmptyScanResults = scanSets.performEmptyAnalysis(concernedArtists)
		requests.reportNumberingScanResults = scanSets.performNumberingAnalysis(concernedArtists)
		requests.reportFilesScanResults = scanSets.performFileAnalysis(o, concernedArtists, ss, ios)
		for _, artist := range concernedArtists {
			artist.rollup()
			artist.toConsole(o)
		}
		scanSets.maybeReportCleanResults(o, requests)
	}
	return
}

type scanReportRequests struct {
	reportEmptyScanResults     bool
	reportFilesScanResults     bool
	reportNumberingScanResults bool
}

func (scanSets *scanSettings) maybeReportCleanResults(o output.Bus, requests scanReportRequests) {
	if !requests.reportEmptyScanResults && scanSets.empty.Value {
		o.ConsolePrintln("Empty Folder Analysis: no empty folders found.")
	}
	if !requests.reportNumberingScanResults && scanSets.numbering.Value {
		o.ConsolePrintln("Numbering Analysis: no missing or duplicate tracks found.")
	}
	if !requests.reportFilesScanResults && scanSets.files.Value {
		o.ConsolePrintln("File Analysis: no inconsistencies found.")
	}
}

func (scanSets *scanSettings) performFileAnalysis(
	o output.Bus,
	concernedArtists []*concernedArtist,
	ss *searchSettings,
	ios *ioSettings,
) bool {
	foundConcerns := false
	if scanSets.files.Value {
		artists := make([]*files.Artist, 0, len(concernedArtists))
		for _, cAr := range concernedArtists {
			artists = append(artists, cAr.backingArtist())
		}
		if filteredArtists := ss.filter(o, artists); len(filteredArtists) != 0 {
			readMetadata(o, filteredArtists, ios.openFileLimit)
			for _, artist := range filteredArtists {
				for _, album := range artist.Albums() {
					for _, track := range album.Tracks() {
						concerns := track.ReportMetadataProblems()
						if found := recordTrackFileConcerns(concernedArtists, track, concerns); found {
							foundConcerns = true
						}
					}
				}
			}
		}
	}
	return foundConcerns
}

func recordTrackFileConcerns(artists []*concernedArtist, track *files.Track, concerns []string) (foundConcerns bool) {
	if len(concerns) > 0 {
		foundConcerns = true
		for _, cAr := range artists {
			if cT := cAr.lookup(track); cT != nil {
				for _, s := range concerns {
					cT.addConcern(filesConcern, s)
				}
				break
			}
		}
	}
	return foundConcerns
}

func (scanSets *scanSettings) performNumberingAnalysis(
	concernedArtists []*concernedArtist) bool {
	foundConcerns := false
	if scanSets.numbering.Value {
		for _, cAr := range concernedArtists {
			for _, cAl := range cAr.albums() {
				trackMap := map[int][]string{}
				maxTrack := len(cAl.tracks())
				for _, cT := range cAl.tracks() {
					track := cT.backingTrack()
					trackNumber := track.Number()
					trackMap[trackNumber] = append(trackMap[trackNumber], cT.name())
					if trackNumber > maxTrack {
						maxTrack = trackNumber
					}
				}
				concerns := generateNumberingConcerns(trackMap, maxTrack)
				if len(concerns) > 0 {
					foundConcerns = true
					for _, s := range concerns {
						cAl.addConcern(numberingConcern, s)
					}
				}
			}
		}
	}
	return foundConcerns
}

func generateNumberingConcerns(m map[int][]string, maxTrack int) []string {
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
				numberGap{value1: 1, value2: numbers[0] - 1}.generateMissingTrackNumbers())
		}
		for k := 0; k < len(numbers)-1; k++ {
			if numbers[k+1]-numbers[k] != 1 {
				missingNumbers = append(missingNumbers,
					numberGap{
						value1: numbers[k] + 1,
						value2: numbers[k+1] - 1,
					}.generateMissingTrackNumbers())
			}
		}
		if numbers[len(numbers)-1] != maxTrack {
			missingNumbers = append(missingNumbers,
				numberGap{
					value1: numbers[len(numbers)-1] + 1,
					value2: maxTrack,
				}.generateMissingTrackNumbers())
		}
	}
	if len(missingNumbers) != 0 {
		concern := fmt.Sprintf("missing tracks identified: %s",
			strings.Join(missingNumbers, ", "))
		concerns = append(concerns, concern)
	}
	return concerns
}

type numberGap struct {
	value1 int
	value2 int
}

func (gap numberGap) generateMissingTrackNumbers() string {
	if gap.value1 == gap.value2 {
		return fmt.Sprintf("%d", gap.value1)
	}
	return fmt.Sprintf("%d-%d", min(gap.value1, gap.value2), max(gap.value1, gap.value2))
}

func (scanSets *scanSettings) performEmptyAnalysis(concernedArtists []*concernedArtist) bool {
	emptyFoldersFound := false
	if scanSets.empty.Value {
		for _, concernedArtist := range concernedArtists {
			if !concernedArtist.backingArtist().HasAlbums() {
				concernedArtist.addConcern(emptyConcern, "no albums found")
				emptyFoldersFound = true
				continue // next artist, please
			}
			for _, concernedAlbum := range concernedArtist.albums() {
				if !concernedAlbum.backingAlbum().HasTracks() {
					concernedAlbum.addConcern(emptyConcern, "no tracks found")
					emptyFoldersFound = true
				}
			}
		}
	}
	return emptyFoldersFound
}

func (scanSets *scanSettings) hasWorkToDo(o output.Bus) bool {
	if scanSets.empty.Value || scanSets.files.Value || scanSets.numbering.Value {
		return true
	}
	userPartiallyAtFault := scanSets.empty.UserSet || scanSets.files.UserSet || scanSets.numbering.UserSet
	o.ErrorPrintln("No scans will be performed.")
	o.ErrorPrintln("Why?")
	switch userPartiallyAtFault {
	case true:
		flagsUserSet := make([]string, 0, 3)
		flagsFromConfig := make([]string, 0, 3)
		switch scanSets.empty.UserSet {
		case true:
			flagsUserSet = append(flagsUserSet, scanEmptyFlag)
		case false:
			flagsFromConfig = append(flagsFromConfig, scanEmptyFlag)
		}
		switch scanSets.files.UserSet {
		case true:
			flagsUserSet = append(flagsUserSet, scanFilesFlag)
		case false:
			flagsFromConfig = append(flagsFromConfig, scanFilesFlag)
		}
		switch scanSets.numbering.UserSet {
		case true:
			flagsUserSet = append(flagsUserSet, scanNumberingFlag)
		case false:
			flagsFromConfig = append(flagsFromConfig, scanNumberingFlag)
		}
		switch len(flagsFromConfig) {
		case 0:
			o.ErrorPrintf(
				"You explicitly set %s, %s, and %s false.\n",
				scanEmptyFlag,
				scanFilesFlag,
				scanNumberingFlag,
			)
		default:
			o.ErrorPrintf(
				"In addition to %s configured false, you explicitly set %s false.\n",
				strings.Join(flagsFromConfig, " and "),
				strings.Join(flagsUserSet, " and "))
		}
	default:
		o.ErrorPrintf(
			"The flags %s, %s, and %s are all configured false.\n",
			scanEmptyFlag,
			scanFilesFlag,
			scanNumberingFlag,
		)
	}
	o.ErrorPrintln("What to do:")
	o.ErrorPrintln("Either:")
	o.BeginErrorList(true)
	o.ErrorPrintln("Edit the configuration file so that at least one of these flags is true, or")
	o.ErrorPrintln("Explicitly set at least one of these flags true on the command line.")
	o.EndErrorList()
	return false
}

func processScanFlags(o output.Bus, values map[string]*cmdtoolkit.CommandFlag[any]) (*scanSettings, bool) {
	settings := &scanSettings{}
	flagsOk := true // optimistic
	var flagErr error
	if settings.empty, flagErr = cmdtoolkit.GetBool(o, values, scanEmpty); flagErr != nil {
		flagsOk = false
	}
	if settings.files, flagErr = cmdtoolkit.GetBool(o, values, scanFiles); flagErr != nil {
		flagsOk = false
	}
	if settings.numbering, flagErr = cmdtoolkit.GetBool(o, values, scanNumbering); flagErr != nil {
		flagsOk = false
	}
	return settings, flagsOk
}

func init() {
	rootCmd.AddCommand(scanCmd)
	cmdtoolkit.AddDefaults(scanFlags)
	cmdtoolkit.AddFlags(getBus(), getConfiguration(), scanCmd.Flags(), scanFlags, searchFlags, ioFlags)
}
