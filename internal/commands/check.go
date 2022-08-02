package commands

import (
	"flag"
	"fmt"
	"mp3/internal"
	"mp3/internal/files"
	"sort"
)

type check struct {
	n                         string
	checkEmptyFolders         *bool
	checkGapsInTrackNumbering *bool
	checkIntegrity            *bool
	sf                        *files.SearchFlags
}

func (c *check) name() string {
	return c.n
}

func newCheck(o internal.OutputBus, c *internal.Configuration, fSet *flag.FlagSet) (CommandProcessor, bool) {
	return newCheckCommand(o, c, fSet)
}

const (
	defaultEmptyFolders         = false
	defaultGapsInTrackNumbering = false
	defaultIntegrity            = true
	emptyFoldersFlag            = "empty"
	fkEmptyFoldersFlag          = "-" + emptyFoldersFlag
	fkGapAnalysisFlag           = "-" + gapsInTrackNumberingFlag
	fkIntegrityAnalysisFlag     = "-" + integrityFlag
	gapsInTrackNumberingFlag    = "gaps"
	integrityFlag               = "integrity"
)

func newCheckCommand(o internal.OutputBus, c *internal.Configuration, fSet *flag.FlagSet) (*check, bool) {
	ok := true
	name := fSet.Name()
	configuration := c.SubConfiguration(name)
	defaultEmpty, err := configuration.BoolDefault(emptyFoldersFlag, defaultEmptyFolders)
	if err != nil {
		reportBadDefault(o, name, err)
		ok = false
	}
	defaultGaps, err := configuration.BoolDefault(gapsInTrackNumberingFlag, defaultGapsInTrackNumbering)
	if err != nil {
		reportBadDefault(o, name, err)
		ok = false
	}
	defaultIntegritySetting, err := configuration.BoolDefault(integrityFlag, defaultIntegrity)
	if err != nil {
		reportBadDefault(o, name, err)
		ok = false
	}
	if ok {
		return &check{
			n:                         name,
			checkEmptyFolders:         fSet.Bool(emptyFoldersFlag, defaultEmpty, "check for empty artist and album folders"),
			checkGapsInTrackNumbering: fSet.Bool(gapsInTrackNumberingFlag, defaultGaps, "check for gaps in track numbers"),
			checkIntegrity:            fSet.Bool(integrityFlag, defaultIntegritySetting, "check for disagreement between the file system and audio file metadata"),
			sf:                        files.NewSearchFlags(c, fSet),
		}, ok
	}
	return nil, ok
}

func (c *check) Exec(o internal.OutputBus, args []string) (ok bool) {
	if s, argsOk := c.sf.ProcessArgs(o, args); argsOk {
		ok = c.runCommand(o, s)
	}
	return
}

func (c *check) logFields() map[string]interface{} {
	return map[string]interface{}{
		fkCommandName:           c.name(),
		fkEmptyFoldersFlag:      *c.checkEmptyFolders,
		fkGapAnalysisFlag:       *c.checkGapsInTrackNumbering,
		fkIntegrityAnalysisFlag: *c.checkIntegrity,
	}
}

type trackWithIssues struct {
	number int
	name   string
	issues []string
	track  *files.Track
}

func (t *trackWithIssues) hasIssues() bool {
	return len(t.issues) > 0
}

type albumWithIssues struct {
	name   string
	issues []string
	tracks []*trackWithIssues
	album  *files.Album
}

func (a *albumWithIssues) hasIssues() bool {
	if len(a.issues) > 0 {
		return true
	}
	for _, t := range a.tracks {
		if t.hasIssues() {
			return true
		}
	}
	return false
}

type artistWithIssues struct {
	name   string
	issues []string
	albums []*albumWithIssues
	artist *files.Artist
}

func (a *artistWithIssues) hasIssues() bool {
	if len(a.issues) > 0 {
		return true
	}
	for _, album := range a.albums {
		if album.hasIssues() {
			return true
		}
	}
	return false
}

func (c *check) runCommand(o internal.OutputBus, s *files.Search) (ok bool) {
	if !*c.checkEmptyFolders && !*c.checkGapsInTrackNumbering && !*c.checkIntegrity {
		o.WriteError(internal.USER_SPECIFIED_NO_WORK, c.name())
		o.LogWriter().Warn(internal.LW_NOTHING_TO_DO, c.logFields())
	} else {
		o.LogWriter().Info(internal.LI_EXECUTING_COMMAND, c.logFields())
		artists, artistsWithEmptyIssues, analysisOk := c.performEmptyFolderAnalysis(o, s)
		if analysisOk {
			artists, ok = c.filterArtists(o, s, artists)
			if ok {
				artistsWithGaps := c.performGapAnalysis(o, artists)
				artistsWithIntegrityIssues := c.performIntegrityCheck(o, artists)
				reportResults(o, artistsWithEmptyIssues, artistsWithGaps, artistsWithIntegrityIssues)
			}
		}
	}
	return
}

func reportResults(o internal.OutputBus, artistsWithIssues ...[]*artistWithIssues) {
	var filteredArtistSets [][]*artistWithIssues
	for _, artists := range artistsWithIssues {
		filteredArtistSets = append(filteredArtistSets, filterAndSortArtists(artists))
	}
	filteredArtists := merge(filteredArtistSets)
	if len(filteredArtists) > 0 {
		for _, artist := range filteredArtists {
			o.WriteConsole(false, "%s\n", artist.name)
			for _, issue := range artist.issues {
				o.WriteConsole(false, "  %s\n", issue)
			}
			for _, album := range artist.albums {
				o.WriteConsole(false, "    %s\n", album.name)
				for _, issue := range album.issues {
					o.WriteConsole(false, "      %s\n", issue)
				}
				for _, track := range album.tracks {
					o.WriteConsole(false, "        %2d %s\n", track.number, track.name)
					for _, issue := range track.issues {
						o.WriteConsole(false, "          %s\n", issue)
					}
				}
			}
		}
	}
}

func merge(sets [][]*artistWithIssues) []*artistWithIssues {
	m := make(map[string]*artistWithIssues)
	for _, set := range sets {
		for _, instance := range set {
			if artist, ok := m[instance.name]; !ok {
				m[instance.name] = instance
			} else {
				// merge instance into artist
				artist.issues = append(artist.issues, instance.issues...)
				for _, album := range instance.albums {
					mergedAlbum := false
					for _, existingAlbum := range artist.albums {
						if existingAlbum.name == album.name {
							// merge album into existingAlbum
							existingAlbum.issues = append(existingAlbum.issues, album.issues...)
							for _, track := range album.tracks {
								mergedTrack := false
								for _, existingTrack := range existingAlbum.tracks {
									if existingTrack.number == track.number {
										// merge track into existingTrack
										existingTrack.issues = append(existingTrack.issues, track.issues...)
										mergedTrack = true
										break
									}
								}
								if !mergedTrack {
									existingAlbum.tracks = append(existingAlbum.tracks, track)
								}
							}
							mergedAlbum = true
							break
						}
					}
					if !mergedAlbum {
						artist.albums = append(artist.albums, album)
					}
				}
			}
		}
	}
	var results []*artistWithIssues
	for _, artist := range m {
		results = append(results, artist)
	}
	sortArtists(results)
	return results
}

func (c *check) filterArtists(o internal.OutputBus, s *files.Search, artists []*files.Artist) (filteredArtists []*files.Artist, ok bool) {
	if *c.checkGapsInTrackNumbering || *c.checkIntegrity {
		if len(artists) == 0 {
			filteredArtists, ok = s.LoadData(o)
		} else {
			filteredArtists, ok = s.FilterArtists(o, artists)
		}
	} else {
		filteredArtists = artists
		ok = true
	}
	return
}

type artistSlice []*artistWithIssues

func (a artistSlice) Len() int {
	return len(a)
}

func (a artistSlice) Less(i, j int) bool {
	return a[i].name < a[j].name
}

func (a artistSlice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

type albumSlice []*albumWithIssues

func (a albumSlice) Len() int {
	return len(a)
}

func (a albumSlice) Less(i, j int) bool {
	return a[i].name < a[j].name
}

func (a albumSlice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

type trackSlice []*trackWithIssues

func (t trackSlice) Len() int {
	return len(t)
}

func (t trackSlice) Less(i, j int) bool {
	return t[i].number < t[j].number
}

func (t trackSlice) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func filterAndSortArtists(artists []*artistWithIssues) []*artistWithIssues {
	var filteredArtists []*artistWithIssues
	for _, artist := range artists {
		if artist.hasIssues() {
			filteredArtist := artistWithIssues{
				name:   artist.name,
				issues: artist.issues,
			}
			for _, album := range artist.albums {
				if album.hasIssues() {
					filteredAlbum := albumWithIssues{
						name:   album.name,
						issues: album.issues,
					}
					for _, track := range album.tracks {
						if track.hasIssues() {
							filteredTrack := trackWithIssues{
								name:   track.name,
								number: track.number,
								issues: track.issues,
							}
							filteredAlbum.tracks = append(filteredAlbum.tracks, &filteredTrack)
						}
					}
					filteredArtist.albums = append(filteredArtist.albums, &filteredAlbum)
				}
			}
			filteredArtists = append(filteredArtists, &filteredArtist)
		}
	}
	sortArtists(filteredArtists)
	return filteredArtists
}

func sortArtists(filteredArtists []*artistWithIssues) {
	sort.Sort(artistSlice(filteredArtists))
	for _, artist := range filteredArtists {
		sort.Sort(albumSlice(artist.albums))
		sort.Strings(artist.issues)
		for _, album := range artist.albums {
			sort.Sort(trackSlice(album.tracks))
			sort.Strings(album.issues)
			for _, track := range album.tracks {
				sort.Strings(track.issues)
			}
		}
	}
}

func (c *check) performEmptyFolderAnalysis(o internal.OutputBus, s *files.Search) (artists []*files.Artist, conflictedArtists []*artistWithIssues, ok bool) {
	if !*c.checkEmptyFolders {
		ok = true
		return
	}
	var loadedOk bool
	artists, loadedOk = s.LoadUnfilteredData(o)
	if !loadedOk {
		return
	}
	conflictedArtists = createBareConflictedIssues(artists)
	issuesFound := false
	for _, conflictedArtist := range conflictedArtists {
		if !conflictedArtist.artist.HasAlbums() {
			conflictedArtist.issues = append(conflictedArtist.issues, "no albums found")
			issuesFound = true
		} else {
			for _, conflictedAlbum := range conflictedArtist.albums {
				if !conflictedAlbum.album.HasTracks() {
					conflictedAlbum.issues = append(conflictedAlbum.issues, "no tracks found")
					issuesFound = true
				}
			}
		}
	}
	if !issuesFound {
		o.WriteConsole(true, "Empty Folder Analysis: no empty folders found")
	}
	ok = true
	return
}

func createBareConflictedIssues(artists []*files.Artist) (conflictedArtists []*artistWithIssues) {
	for _, originalArtist := range artists {
		artistWithIssues := artistWithIssues{name: originalArtist.Name(), artist: originalArtist}
		conflictedArtists = append(conflictedArtists, &artistWithIssues)
		for _, originalAlbum := range originalArtist.Albums() {
			albumWithIssues := albumWithIssues{name: originalAlbum.Name(), album: originalAlbum}
			artistWithIssues.albums = append(artistWithIssues.albums, &albumWithIssues)
			for _, originalTrack := range originalAlbum.Tracks() {
				trackWithIssues := trackWithIssues{number: originalTrack.Number(), name: originalTrack.Name(), track: originalTrack}
				albumWithIssues.tracks = append(albumWithIssues.tracks, &trackWithIssues)
			}
		}
	}
	return
}

func (c *check) performIntegrityCheck(o internal.OutputBus, artists []*files.Artist) []*artistWithIssues {
	conflictedArtists := make([]*artistWithIssues, 0)
	if *c.checkIntegrity {
		files.UpdateTracks(o, artists, files.RawReadTags)
		conflictedArtists = createBareConflictedIssues(artists)
		issuesFound := false
		for _, conflictedArtist := range conflictedArtists {
			for _, conflictedAlbum := range conflictedArtist.albums {
				for _, conflictedTrack := range conflictedAlbum.tracks {
					differences := conflictedTrack.track.FindDifferences()
					if len(differences) > 0 {
						conflictedTrack.issues = append(conflictedTrack.issues, differences...)
						issuesFound = true
					}
				}
			}
		}
		if !issuesFound {
			o.WriteConsole(true, "Integrity Analysis: no issues found")
		}
	}
	return conflictedArtists
}

func (c *check) performGapAnalysis(o internal.OutputBus, artists []*files.Artist) []*artistWithIssues {
	conflictedArtists := make([]*artistWithIssues, 0)
	if *c.checkGapsInTrackNumbering {
		conflictedArtists = createBareConflictedIssues(artists)
		issuesFound := false
		for _, conflictedArtist := range conflictedArtists {
			for _, conflictedAlbum := range conflictedArtist.albums {
				m := make(map[int]*trackWithIssues)
				for _, track := range conflictedAlbum.tracks {
					if t, ok := m[track.number]; ok {
						complaint := fmt.Sprintf("track %d used by %q and %q", track.number, t.name, track.name)
						conflictedAlbum.issues = append(conflictedAlbum.issues, complaint)
						issuesFound = true
					} else {
						m[track.number] = track
					}
				}
				missingTracks := 0
				for trackNumber := 1; trackNumber <= len(conflictedAlbum.tracks); trackNumber++ {
					if _, ok := m[trackNumber]; !ok {
						missingTracks++
						conflictedAlbum.issues = append(conflictedAlbum.issues, fmt.Sprintf("missing track %d", trackNumber))
						issuesFound = true
					}
				}
				expectedTrackCount := len(conflictedAlbum.tracks) + missingTracks
				validTracks := fmt.Sprintf("valid tracks are 1..%d", expectedTrackCount)
				for trackNumber, track := range m {
					switch {
					case trackNumber < 1:
						complaint := fmt.Sprintf("track %d (%q) is not a valid track number; %s", trackNumber, track.name, validTracks)
						conflictedAlbum.issues = append(conflictedAlbum.issues, complaint)
						issuesFound = true
					case trackNumber > expectedTrackCount:
						complaint := fmt.Sprintf("track %d (%q) is not a valid track number; %s", trackNumber, track.name, validTracks)
						conflictedAlbum.issues = append(conflictedAlbum.issues, complaint)
						issuesFound = true
					}
				}
			}
		}
		if !issuesFound {
			o.WriteConsole(true, "Check Gaps: no gaps found")
		}
	}
	return conflictedArtists
}
