/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd

import (
	"fmt"
	"mp3/internal/files"
	"slices"
	"strings"

	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

// Details:

// An mp3 file usually contains ID3V1 metadata (gory details: https://id3.org/ID3v1),
// ID3V2 metadata (gory details: https://id3.org/id3v2.3.0), often both, in addition to
// the audio data. The integrity check reads each mp3 file's metadata and does the
// following:

// * Verify that the file name begins with the track number encoded in the TRCK (track
//   number/position in set) ID3V2 frame and the ID3V1 track field, and that the rest of
//   the file name matches the value encoded in the TIT2 (title/songname/content
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
//   - ID3V1 fields can not encode multi-byte characters; similar 8-bit characters are
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
		Use:                   CheckCommand + " [" + CheckEmptyFlag + "] [" + CheckFilesFlag + "] [" + CheckNumberingFlag + "] " + searchUsage,
		DisableFlagsInUseLine: true,
		Short:                 "Runs checks on mp3 files and their directories and reports problems",
		Long:                  fmt.Sprintf("%q runs checks on mp3 files and their containing directories and reports any problems detected", CheckCommand),
		Example: "" +
			CheckCommand + " " + CheckEmptyFlag + "\n" +
			"  reports empty artist and album directories\n" +
			CheckCommand + " " + CheckFilesFlag + "\n" +
			"  reads each mp3 file's metadata and reports any inconsistencies found\n" +
			CheckCommand + " " + CheckNumberingFlag + "\n" +
			"  reports errors in the track numbers of mp3 files",
		Run: CheckRun,
	}
	CheckFlags = SectionFlags{
		SectionName: CheckCommand,
		Flags: map[string]*FlagDetails{
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

func CheckRun(cmd *cobra.Command, _ []string) {
	commandStatus := ProgramError
	o := getBus()
	producer := cmd.Flags()
	values, eSlice := ReadFlags(producer, CheckFlags)
	searchSettings, searchFlagsOk := EvaluateSearchFlags(o, producer)
	if ProcessFlagErrors(o, eSlice) && searchFlagsOk {
		if cs, ok := ProcessCheckFlags(o, values); ok {
			LogCommandStart(o, CheckCommand, map[string]any{
				CheckEmptyFlag:         cs.empty,
				"empty-user-set":       cs.emptyUserSet,
				CheckFilesFlag:         cs.files,
				"files-user-set":       cs.filesUserSet,
				CheckNumberingFlag:     cs.numbering,
				"numbering-user-set":   cs.numberingUserSet,
				SearchAlbumFilterFlag:  searchSettings.AlbumFilter,
				SearchArtistFilterFlag: searchSettings.ArtistFilter,
				SearchTrackFilterFlag:  searchSettings.TrackFilter,
				SearchTopDirFlag:       searchSettings.TopDirectory,
			})
			commandStatus = cs.MaybeDoWork(o, searchSettings)
		}
	}
	Exit(commandStatus)
}

type CheckSettings struct {
	empty            bool
	emptyUserSet     bool
	files            bool
	filesUserSet     bool
	numbering        bool
	numberingUserSet bool
}

func NewCheckSettings() *CheckSettings {
	return &CheckSettings{}
}

func (cs *CheckSettings) WithEmpty(b bool) *CheckSettings {
	cs.empty = b
	return cs
}

func (cs *CheckSettings) WithEmptyUserSet(b bool) *CheckSettings {
	cs.emptyUserSet = b
	return cs
}

func (cs *CheckSettings) WithFiles(b bool) *CheckSettings {
	cs.files = b
	return cs
}

func (cs *CheckSettings) WithFilesUserSet(b bool) *CheckSettings {
	cs.filesUserSet = b
	return cs
}

func (cs *CheckSettings) WithNumbering(b bool) *CheckSettings {
	cs.numbering = b
	return cs
}

func (cs *CheckSettings) WithNumberingUserSet(b bool) *CheckSettings {
	cs.numberingUserSet = b
	return cs
}

func (cs *CheckSettings) MaybeDoWork(o output.Bus, ss *SearchSettings) int {
	status := UserError
	if cs.HasWorkToDo(o) {
		allArtists, loaded := ss.Load(o)
		status = cs.PerformChecks(o, allArtists, loaded, ss)
	}
	return status
}

type CheckIssueType int32

const (
	CheckUnspecifiedIssue CheckIssueType = iota
	CheckEmptyIssue
	CheckFilesIssue
	CheckNumberingIssue
	CheckConflictIssue
)

func IssueTypeAsString(i CheckIssueType) string {
	switch i {
	case CheckEmptyIssue:
		return CheckEmpty
	case CheckFilesIssue:
		return CheckFiles
	case CheckNumberingIssue:
		return CheckNumbering
	case CheckConflictIssue:
		return "metadata conflict"
	}
	return fmt.Sprintf("unspecified issue %d", i)
}

type CheckedIssues struct {
	issues map[CheckIssueType][]string
}

func NewCheckedIssues() CheckedIssues {
	return CheckedIssues{issues: map[CheckIssueType][]string{}}
}

func (cI CheckedIssues) AddIssue(source CheckIssueType, issue string) {
	cI.issues[source] = append(cI.issues[source], issue)
}

func (cI CheckedIssues) HasIssues() bool {
	for _, list := range cI.issues {
		if len(list) > 0 {
			return true
		}
	}
	return false
}

func (cI CheckedIssues) OutputIssues(o output.Bus, tab int) {
	if cI.HasIssues() {
		iStrings := []string{}
		for key, value := range cI.issues {
			for _, s := range value {
				iStrings = append(iStrings, fmt.Sprintf("* [%s] %s", IssueTypeAsString(key), s))
			}
		}
		slices.Sort(iStrings)
		for _, s := range iStrings {
			o.WriteConsole("%*s%s\n", tab, "", s)
		}
	}
}

type CheckedTrack struct {
	CheckedIssues
	backing *files.Track
}

func NewCheckedTrack(track *files.Track) *CheckedTrack {
	if track == nil {
		return nil
	}
	return &CheckedTrack{
		CheckedIssues: NewCheckedIssues(),
		backing:       track,
	}
}

func (cT *CheckedTrack) AddIssue(source CheckIssueType, issue string) {
	cT.CheckedIssues.AddIssue(source, issue)
}

func (cT *CheckedTrack) HasIssues() bool {
	return cT.CheckedIssues.HasIssues()
}

func (cT *CheckedTrack) name() string {
	return cT.backing.CommonName()
}

func (cT *CheckedTrack) OutputIssues(o output.Bus) {
	if cT.HasIssues() {
		o.WriteConsole("    Track %q\n", cT.name())
		cT.CheckedIssues.OutputIssues(o, 4)
	}
}

func (cT *CheckedTrack) Track() *files.Track {
	return cT.backing
}

type CheckedAlbum struct {
	CheckedIssues
	tracks   []*CheckedTrack
	backing  *files.Album
	trackMap map[string]*CheckedTrack
}

func NewCheckedAlbum(album *files.Album) *CheckedAlbum {
	if album == nil {
		return nil
	}
	cAl := &CheckedAlbum{
		CheckedIssues: NewCheckedIssues(),
		tracks:        []*CheckedTrack{},
		backing:       album,
		trackMap:      map[string]*CheckedTrack{},
	}
	for _, track := range album.Tracks() {
		cAl.AddTrack(track)
	}
	return cAl
}

func (cAl *CheckedAlbum) AddIssue(source CheckIssueType, issue string) {
	cAl.CheckedIssues.AddIssue(source, issue)
}

func (cAl *CheckedAlbum) AddTrack(track *files.Track) {
	if cT := NewCheckedTrack(track); cT != nil {
		cAl.tracks = append(cAl.tracks, cT)
		cAl.trackMap[cT.backing.FileName()] = cT
	}
}

func (cAl *CheckedAlbum) Album() *files.Album {
	return cAl.backing
}

func (cAl *CheckedAlbum) HasIssues() bool {
	if cAl.CheckedIssues.HasIssues() {
		return true
	}
	for _, cT := range cAl.tracks {
		if cT.HasIssues() {
			return true
		}
	}
	return false
}

func (cAl *CheckedAlbum) name() string {
	return cAl.backing.Name()
}

func (cAl *CheckedAlbum) Lookup(track *files.Track) *CheckedTrack {
	var cT *CheckedTrack
	if found, ok := cAl.trackMap[track.FileName()]; ok {
		cT = found
	}
	return cT
}

func (cAl *CheckedAlbum) OutputIssues(o output.Bus) {
	if cAl.HasIssues() {
		o.WriteConsole("  Album %q\n", cAl.name())
		cAl.CheckedIssues.OutputIssues(o, 2)
		m := map[string]*CheckedTrack{}
		names := []string{}
		for _, cT := range cAl.tracks {
			trackName := cT.name()
			m[trackName] = cT
			names = append(names, trackName)
		}
		slices.Sort(names)
		for _, name := range names {
			if cT := m[name]; cT != nil {
				cT.OutputIssues(o)
			}
		}
	}
}

func (cAl *CheckedAlbum) Tracks() []*CheckedTrack {
	return cAl.tracks
}

type CheckedArtist struct {
	CheckedIssues
	albums   []*CheckedAlbum
	backing  *files.Artist
	albumMap map[string]*CheckedAlbum
}

func NewCheckedArtist(artist *files.Artist) *CheckedArtist {
	if artist == nil {
		return nil
	}
	cAr := &CheckedArtist{
		CheckedIssues: NewCheckedIssues(),
		albums:        []*CheckedAlbum{},
		backing:       artist,
		albumMap:      map[string]*CheckedAlbum{},
	}
	for _, album := range artist.Albums() {
		cAr.AddAlbum(album)
	}
	return cAr
}

func (cAr *CheckedArtist) AddAlbum(album *files.Album) {
	if cAl := NewCheckedAlbum(album); cAl != nil {
		cAr.albums = append(cAr.albums, cAl)
		cAr.albumMap[cAl.name()] = cAl
	}
}

func (cAr *CheckedArtist) AddIssue(source CheckIssueType, issue string) {
	cAr.CheckedIssues.AddIssue(source, issue)
}

func (cAr *CheckedArtist) Albums() []*CheckedAlbum {
	return cAr.albums
}

func (cAr *CheckedArtist) Artist() *files.Artist {
	return cAr.backing
}

func (cAr *CheckedArtist) HasIssues() bool {
	if cAr.CheckedIssues.HasIssues() {
		return true
	}
	for _, cAl := range cAr.albums {
		if cAl.HasIssues() {
			return true
		}
	}
	return false
}

func (cAr *CheckedArtist) Lookup(track *files.Track) *CheckedTrack {
	albumKey := track.AlbumName()
	if cAl, ok := cAr.albumMap[albumKey]; ok {
		return cAl.Lookup(track)
	}
	return nil
}

func (cAr *CheckedArtist) name() string {
	return cAr.backing.Name()
}

func (cAr *CheckedArtist) OutputIssues(o output.Bus) {
	if cAr.HasIssues() {
		o.WriteConsole("Artist %q\n", cAr.name())
		cAr.CheckedIssues.OutputIssues(o, 0)
		m := map[string]*CheckedAlbum{}
		names := []string{}
		for _, cT := range cAr.albums {
			albumName := cT.name()
			m[albumName] = cT
			names = append(names, albumName)
		}
		slices.Sort(names)
		for _, name := range names {
			if cAl := m[name]; cAl != nil {
				cAl.OutputIssues(o)
			}
		}
	}
}

func PrepareCheckedArtists(artists []*files.Artist) []*CheckedArtist {
	checkedArtists := []*CheckedArtist{}
	for _, artist := range artists {
		if cAr := NewCheckedArtist(artist); cAr != nil {
			checkedArtists = append(checkedArtists, cAr)
		}
	}
	return checkedArtists
}

func (cs *CheckSettings) PerformChecks(o output.Bus, artists []*files.Artist, artistsLoaded bool, ss *SearchSettings) int {
	status := UserError
	if artistsLoaded && len(artists) > 0 {
		status = Success
		checkedArtists := PrepareCheckedArtists(artists)
		emptyFoldersFound := cs.PerformEmptyAnalysis(checkedArtists)
		numberingIssuesFound := cs.PerformNumberingAnalysis(checkedArtists)
		fileIssuesFound := cs.PerformFileAnalysis(o, checkedArtists, ss)
		for _, artist := range checkedArtists {
			artist.OutputIssues(o)
		}
		cs.MaybeReportCleanResults(o, emptyFoldersFound, numberingIssuesFound, fileIssuesFound)
	}
	return status
}

func (cs *CheckSettings) MaybeReportCleanResults(o output.Bus, emptyIssues, numberingIssues, fileIssues bool) {
	if !emptyIssues && cs.empty {
		o.WriteCanonicalConsole("Empty Folder Analysis: no empty folders found")
	}
	if !numberingIssues && cs.numbering {
		o.WriteCanonicalConsole("Numbering Analysis: no missing or duplicate tracks found")
	}
	if !fileIssues && cs.files {
		o.WriteCanonicalConsole("File Analysis: no inconsistencies found")
	}
}

func (cs *CheckSettings) PerformFileAnalysis(o output.Bus, checkedArtists []*CheckedArtist, ss *SearchSettings) bool {
	foundIssues := false
	if cs.files {
		artists := []*files.Artist{}
		for _, cAr := range checkedArtists {
			artists = append(artists, cAr.Artist())
		}
		if filteredArtists, filtered := ss.Filter(o, artists); filtered {
			ReadMetadata(o, filteredArtists)
			for _, artist := range filteredArtists {
				for _, album := range artist.Albums() {
					for _, track := range album.Tracks() {
						issues := track.ReportMetadataProblems()
						if found := RecordFileIssues(checkedArtists, track, issues); found {
							foundIssues = true
						}
					}
				}
			}
		}
	}
	return foundIssues
}

func RecordFileIssues(checkedArtists []*CheckedArtist, track *files.Track, issues []string) (foundIssues bool) {
	if len(issues) > 0 {
		foundIssues = true
		for _, cAr := range checkedArtists {
			if cT := cAr.Lookup(track); cT != nil {
				for _, s := range issues {
					cT.AddIssue(CheckFilesIssue, s)
				}
				break
			}
		}
	}
	return foundIssues
}

func (cs *CheckSettings) PerformNumberingAnalysis(checkedArtists []*CheckedArtist) bool {
	foundIssues := false
	if cs.numbering {
		for _, cAr := range checkedArtists {
			for _, cAl := range cAr.Albums() {
				trackMap := map[int][]string{}
				maxTrack := len(cAl.Tracks())
				for _, cT := range cAl.Tracks() {
					track := cT.Track()
					trackNumber := track.Number()
					trackMap[trackNumber] = append(trackMap[trackNumber], cT.name())
					if trackNumber > maxTrack {
						maxTrack = trackNumber
					}
				}
				issues := GenerateNumberingIssues(trackMap, maxTrack)
				if len(issues) > 0 {
					foundIssues = true
					for _, s := range issues {
						cAl.AddIssue(CheckNumberingIssue, s)
					}
				}
			}
		}
	}
	return foundIssues
}

func GenerateNumberingIssues(m map[int][]string, maxTrack int) []string {
	issues := []string{}
	numbers := []int{}
	// find duplicates
	for k, v := range m {
		if len(v) != 0 {
			numbers = append(numbers, k)
		}
		if len(v) > 1 {
			slices.Sort(v)
			formattedTracks := []string{}
			for j := 0; j < len(v)-1; j++ {
				formattedTracks = append(formattedTracks, fmt.Sprintf("%q", v[j]))
			}
			finalTrack := fmt.Sprintf("%q", v[len(v)-1])
			issue := fmt.Sprintf("multiple tracks identified as track %d: %s and %s", k, strings.Join(formattedTracks, ", "), finalTrack)
			issues = append(issues, issue)
		}
	}
	// find missing track numbers
	slices.Sort(numbers)
	missingNumbers := []string{}
	if len(numbers) != 0 {
		if numbers[0] > 1 {
			missingNumbers = append(missingNumbers, GenerateMissingNumbers(1, numbers[0]-1))
		}
		for k := 0; k < len(numbers)-1; k++ {
			if numbers[k+1]-numbers[k] != 1 {
				missingNumbers = append(missingNumbers, GenerateMissingNumbers(numbers[k]+1, numbers[k+1]-1))
			}
		}
		if numbers[len(numbers)-1] != maxTrack {
			missingNumbers = append(missingNumbers, GenerateMissingNumbers(numbers[len(numbers)-1]+1, maxTrack))
		}
	}
	if len(missingNumbers) != 0 {
		issue := fmt.Sprintf("missing tracks identified: %s", strings.Join(missingNumbers, ", "))
		issues = append(issues, issue)
	}
	return issues
}

func GenerateMissingNumbers(low, high int) string {
	if low == high {
		return fmt.Sprintf("%d", low)
	} else {
		return fmt.Sprintf("%d-%d", low, high)
	}
}

func (cs *CheckSettings) PerformEmptyAnalysis(checkedArtists []*CheckedArtist) bool {
	emptyFoldersFound := false
	if cs.empty {
		for _, checkedArtist := range checkedArtists {
			if !checkedArtist.Artist().HasAlbums() {
				checkedArtist.AddIssue(CheckEmptyIssue, "no albums found")
				emptyFoldersFound = true
			} else {
				for _, checkedAlbum := range checkedArtist.Albums() {
					if !checkedAlbum.Album().HasTracks() {
						checkedAlbum.AddIssue(CheckEmptyIssue, "no tracks found")
						emptyFoldersFound = true
					}
				}
			}
		}
	}
	return emptyFoldersFound
}

func (cs *CheckSettings) HasWorkToDo(o output.Bus) bool {
	if cs.empty || cs.files || cs.numbering {
		return true
	}
	userPartiallyAtFault := cs.emptyUserSet || cs.filesUserSet || cs.numberingUserSet
	o.WriteCanonicalError("No checks will be executed.\nWhy?\n")
	if userPartiallyAtFault {
		flagsUserSet := []string{}
		flagsFromConfig := []string{}
		if cs.emptyUserSet {
			flagsUserSet = append(flagsUserSet, CheckEmptyFlag)
		} else {
			flagsFromConfig = append(flagsFromConfig, CheckEmptyFlag)
		}
		if cs.filesUserSet {
			flagsUserSet = append(flagsUserSet, CheckFilesFlag)
		} else {
			flagsFromConfig = append(flagsFromConfig, CheckFilesFlag)
		}
		if cs.numberingUserSet {
			flagsUserSet = append(flagsUserSet, CheckNumberingFlag)
		} else {
			flagsFromConfig = append(flagsFromConfig, CheckNumberingFlag)
		}
		if len(flagsFromConfig) == 0 {
			o.WriteCanonicalError("You explicitly set %s, %s, and %s false", CheckEmptyFlag, CheckFilesFlag, CheckNumberingFlag)
		} else {
			o.WriteCanonicalError("In addition to %s configured false, you explicitly set %s false", strings.Join(flagsFromConfig, " and "), strings.Join(flagsUserSet, " and "))
		}
	} else {
		o.WriteCanonicalError("The flags %s, %s, and %s are all configured false", CheckEmptyFlag, CheckFilesFlag, CheckNumberingFlag)
	}
	o.WriteError("What to do:\n")
	o.WriteCanonicalError("Either:\n[1] Edit the configuration file so that at least one of these flags is true, or\n[2] explicitly set at least one of these flags true on the command line")
	return false
}

func ProcessCheckFlags(o output.Bus, values map[string]*FlagValue) (*CheckSettings, bool) {
	settings := &CheckSettings{}
	ok := true // optimistic
	var err error
	if settings.empty, settings.emptyUserSet, err = GetBool(o, values, CheckEmpty); err != nil {
		ok = false
	}
	if settings.files, settings.filesUserSet, err = GetBool(o, values, CheckFiles); err != nil {
		ok = false
	}
	if settings.numbering, settings.numberingUserSet, err = GetBool(o, values, CheckNumbering); err != nil {
		ok = false
	}
	return settings, ok
}

func init() {
	RootCmd.AddCommand(CheckCmd)
	addDefaults(CheckFlags)
	o := getBus()
	c := getConfiguration()
	AddFlags(o, c, CheckCmd.Flags(), CheckFlags, true)
}
