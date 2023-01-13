package commands

import (
	"flag"
	"fmt"
	"mp3/internal"
	"mp3/internal/files"
	"sort"

	"github.com/majohn-r/output"
)

func init() {
	addCommandData(checkCommandName, commandData{isDefault: false, init: newCheck})
	addDefaultMapping(checkCommandName, map[string]any{
		emptyFolders:       defaultEmptyFolders,
		trackNumberingGaps: defaultTrackNumberingGaps,
		integrity:          defaultIntegrity,
	})
}

type check struct {
	emptyFolders       *bool
	trackNumberingGaps *bool
	integrity          *bool
	sf                 *files.SearchFlags
}

func newCheck(o output.Bus, c *internal.Configuration, fSet *flag.FlagSet) (CommandProcessor, bool) {
	return newCheckCommand(o, c, fSet)
}

const (
	checkCommandName = "check"

	defaultEmptyFolders       = false
	defaultTrackNumberingGaps = false
	defaultIntegrity          = true

	emptyFolders       = "empty"
	trackNumberingGaps = "gaps"
	integrity          = "integrity"
)

type checkDefaults struct {
	empty     bool
	gaps      bool
	integrity bool
}

func newCheckCommand(o output.Bus, c *internal.Configuration, fSet *flag.FlagSet) (*check, bool) {
	defaults, defaultsOk := evaluateCheckDefaults(o, c.SubConfiguration(checkCommandName))
	sFlags, sFlagsOk := files.NewSearchFlags(o, c, fSet)
	if defaultsOk && sFlagsOk {
		return &check{
			emptyFolders:       fSet.Bool(emptyFolders, defaults.empty, internal.DecorateBoolFlagUsage("check for empty artist and album folders", defaults.empty)),
			trackNumberingGaps: fSet.Bool(trackNumberingGaps, defaults.gaps, internal.DecorateBoolFlagUsage("check for gaps in track numbers", defaults.gaps)),
			integrity:          fSet.Bool(integrity, defaults.integrity, internal.DecorateBoolFlagUsage("check for disagreement between the file system and audio file metadata", defaults.integrity)),
			sf:                 sFlags,
		}, true
	}
	return nil, false
}

func evaluateCheckDefaults(o output.Bus, c *internal.Configuration) (defaults checkDefaults, ok bool) {
	ok = true
	var err error
	defaults = checkDefaults{}
	defaults.empty, err = c.BoolDefault(emptyFolders, defaultEmptyFolders)
	if err != nil {
		reportBadDefault(o, checkCommandName, err)
		ok = false
	}
	defaults.gaps, err = c.BoolDefault(trackNumberingGaps, defaultTrackNumberingGaps)
	if err != nil {
		reportBadDefault(o, checkCommandName, err)
		ok = false
	}
	defaults.integrity, err = c.BoolDefault(integrity, defaultIntegrity)
	if err != nil {
		reportBadDefault(o, checkCommandName, err)
		ok = false
	}
	return
}

func (c *check) Exec(o output.Bus, args []string) (ok bool) {
	if s, argsOk := c.sf.ProcessArgs(o, args); argsOk {
		ok = c.runCommand(o, s)
	}
	return
}

func (c *check) logFields() map[string]any {
	return map[string]any{
		"command":                checkCommandName,
		"-" + emptyFolders:       *c.emptyFolders,
		"-" + trackNumberingGaps: *c.trackNumberingGaps,
		"-" + integrity:          *c.integrity,
	}
}

type checkedTrack struct {
	issues  []string
	backing *files.Track
}

func (cT *checkedTrack) hasIssues() bool {
	return len(cT.issues) > 0
}

type checkedAlbum struct {
	issues  []string
	tracks  []*checkedTrack
	backing *files.Album
}

func (cAl *checkedAlbum) hasIssues() bool {
	if len(cAl.issues) > 0 {
		return true
	}
	for _, cT := range cAl.tracks {
		if cT.hasIssues() {
			return true
		}
	}
	return false
}

type checkedArtist struct {
	issues  []string
	albums  []*checkedAlbum
	backing *files.Artist
}

func (cAr *checkedArtist) hasIssues() bool {
	if len(cAr.issues) > 0 {
		return true
	}
	for _, cAl := range cAr.albums {
		if cAl.hasIssues() {
			return true
		}
	}
	return false
}

func (c *check) runCommand(o output.Bus, s *files.Search) (ok bool) {
	if !*c.emptyFolders && !*c.trackNumberingGaps && !*c.integrity {
		reportNothingToDo(o, checkCommandName, c.logFields())
	} else {
		logStart(o, checkCommandName, c.logFields())
		artists, artistsWithEmptyFolders, analysisOk := c.analyzeEmptyFolders(o, s)
		if analysisOk {
			artists, ok = c.filterArtists(o, s, artists)
			if ok {
				artistsWithGaps := c.analyzeGaps(o, artists)
				artistsWithIntegrityIssues := c.analyzeIntegrity(o, artists)
				reportResults(o, artistsWithEmptyFolders, artistsWithGaps, artistsWithIntegrityIssues)
			}
		}
	}
	return
}

func reportResults(o output.Bus, checkedArtists ...[]*checkedArtist) {
	var filteredArtistSets [][]*checkedArtist
	for _, artists := range checkedArtists {
		filteredArtistSets = append(filteredArtistSets, filterAndSortCheckedArtists(artists))
	}
	filteredArtists := merge(filteredArtistSets)
	if len(filteredArtists) > 0 {
		for _, cAr := range filteredArtists {
			o.WriteConsole("%s\n", cAr.backing.Name())
			for _, issue := range cAr.issues {
				o.WriteConsole("  %s\n", issue)
			}
			for _, cAl := range cAr.albums {
				o.WriteConsole("    %s\n", cAl.backing.Name())
				for _, issue := range cAl.issues {
					o.WriteConsole("      %s\n", issue)
				}
				for _, cT := range cAl.tracks {
					o.WriteConsole("        %2d %s\n", cT.backing.Number(), cT.backing.Name())
					for _, issue := range cT.issues {
						o.WriteConsole("          %s\n", issue)
					}
				}
			}
		}
	}
}

func merge(sets [][]*checkedArtist) []*checkedArtist {
	m := make(map[string]*checkedArtist)
	for _, set := range sets {
		for _, instance := range set {
			if cAr, ok := m[instance.backing.Name()]; !ok {
				m[instance.backing.Name()] = instance
			} else {
				// merge instance into artist
				cAr.issues = append(cAr.issues, instance.issues...)
				for _, cAl := range instance.albums {
					mergedAlbum := false
					for _, existingAlbum := range cAr.albums {
						if existingAlbum.backing.Name() == cAl.backing.Name() {
							// merge album into existingAlbum
							existingAlbum.issues = append(existingAlbum.issues, cAl.issues...)
							for _, cT := range cAl.tracks {
								mergedTrack := false
								for _, existingTrack := range existingAlbum.tracks {
									if existingTrack.backing.Number() == cT.backing.Number() {
										// merge track into existingTrack
										existingTrack.issues = append(existingTrack.issues, cT.issues...)
										mergedTrack = true
										break
									}
								}
								if !mergedTrack {
									existingAlbum.tracks = append(existingAlbum.tracks, cT)
								}
							}
							mergedAlbum = true
							break
						}
					}
					if !mergedAlbum {
						cAr.albums = append(cAr.albums, cAl)
					}
				}
			}
		}
	}
	var checked []*checkedArtist
	for _, artist := range m {
		checked = append(checked, artist)
	}
	sortCheckedArtists(checked)
	return checked
}

func (c *check) filterArtists(o output.Bus, s *files.Search, artists []*files.Artist) (filtered []*files.Artist, ok bool) {
	if *c.trackNumberingGaps || *c.integrity {
		if len(artists) == 0 {
			filtered, ok = s.Load(o)
		} else {
			filtered, ok = s.FilterArtists(o, artists)
		}
	} else {
		filtered = artists
		ok = true
	}
	return
}

type checkedArtistSlice []*checkedArtist

func (cArS checkedArtistSlice) Len() int {
	return len(cArS)
}

func (cArS checkedArtistSlice) Less(i, j int) bool {
	return cArS[i].backing.Name() < cArS[j].backing.Name()
}

func (cArS checkedArtistSlice) Swap(i, j int) {
	cArS[i], cArS[j] = cArS[j], cArS[i]
}

type checkedAlbumSlice []*checkedAlbum

func (cAlS checkedAlbumSlice) Len() int {
	return len(cAlS)
}

func (cAlS checkedAlbumSlice) Less(i, j int) bool {
	return cAlS[i].backing.Name() < cAlS[j].backing.Name()
}

func (cAlS checkedAlbumSlice) Swap(i, j int) {
	cAlS[i], cAlS[j] = cAlS[j], cAlS[i]
}

type checkedTrackSlice []*checkedTrack

func (cTS checkedTrackSlice) Len() int {
	return len(cTS)
}

func (cTS checkedTrackSlice) Less(i, j int) bool {
	return cTS[i].backing.Number() < cTS[j].backing.Number()
}

func (cTS checkedTrackSlice) Swap(i, j int) {
	cTS[i], cTS[j] = cTS[j], cTS[i]
}

func filterAndSortCheckedArtists(checkedArtists []*checkedArtist) []*checkedArtist {
	var filtered []*checkedArtist
	for _, cAr := range checkedArtists {
		if cAr.hasIssues() {
			fAr := checkedArtist{
				issues:  cAr.issues,
				backing: cAr.backing,
			}
			for _, cAl := range cAr.albums {
				if cAl.hasIssues() {
					fAl := checkedAlbum{
						issues:  cAl.issues,
						backing: cAl.backing,
					}
					for _, cT := range cAl.tracks {
						if cT.hasIssues() {
							fT := checkedTrack{
								issues:  cT.issues,
								backing: cT.backing,
							}
							fAl.tracks = append(fAl.tracks, &fT)
						}
					}
					fAr.albums = append(fAr.albums, &fAl)
				}
			}
			filtered = append(filtered, &fAr)
		}
	}
	sortCheckedArtists(filtered)
	return filtered
}

func sortCheckedArtists(checkedArtists []*checkedArtist) {
	sort.Sort(checkedArtistSlice(checkedArtists))
	for _, cAr := range checkedArtists {
		sort.Sort(checkedAlbumSlice(cAr.albums))
		sort.Strings(cAr.issues)
		for _, cAl := range cAr.albums {
			sort.Sort(checkedTrackSlice(cAl.tracks))
			sort.Strings(cAl.issues)
			for _, cT := range cAl.tracks {
				sort.Strings(cT.issues)
			}
		}
	}
}

func (c *check) analyzeEmptyFolders(o output.Bus, s *files.Search) (artists []*files.Artist, emptyArtists []*checkedArtist, ok bool) {
	if !*c.emptyFolders {
		ok = true
		return
	}
	var loadedOk bool
	artists, loadedOk = s.LoadUnfiltered(o)
	if !loadedOk {
		return
	}
	emptyArtists = toCheckedArtists(artists)
	emptyFoldersFound := false
	for _, cAr := range emptyArtists {
		if !cAr.backing.HasAlbums() {
			cAr.issues = append(cAr.issues, "no albums found")
			emptyFoldersFound = true
		} else {
			for _, cAl := range cAr.albums {
				if !cAl.backing.HasTracks() {
					cAl.issues = append(cAl.issues, "no tracks found")
					emptyFoldersFound = true
				}
			}
		}
	}
	if !emptyFoldersFound {
		o.WriteCanonicalConsole("Empty Folder Analysis: no empty folders found")
	}
	ok = true
	return
}

func toCheckedArtists(artists []*files.Artist) (checkedArtists []*checkedArtist) {
	for _, artist := range artists {
		cAr := checkedArtist{backing: artist}
		checkedArtists = append(checkedArtists, &cAr)
		for _, album := range artist.Albums() {
			cAl := checkedAlbum{backing: album}
			cAr.albums = append(cAr.albums, &cAl)
			for _, track := range album.Tracks() {
				cAl.tracks = append(cAl.tracks, &checkedTrack{backing: track})
			}
		}
	}
	return
}

func (c *check) analyzeIntegrity(o output.Bus, artists []*files.Artist) []*checkedArtist {
	checkedArtists := make([]*checkedArtist, 0)
	if *c.integrity {
		files.ReadMetadata(o, artists)
		checkedArtists = toCheckedArtists(artists)
		issuesFound := false
		for _, cAr := range checkedArtists {
			for _, cAl := range cAr.albums {
				for _, cT := range cAl.tracks {
					issues := cT.backing.ReportMetadataProblems()
					if len(issues) > 0 {
						cT.issues = append(cT.issues, issues...)
						issuesFound = true
					}
				}
			}
		}
		if !issuesFound {
			o.WriteCanonicalConsole("Integrity Analysis: no issues found")
		}
	}
	return checkedArtists
}

func (c *check) analyzeGaps(o output.Bus, artists []*files.Artist) []*checkedArtist {
	checkedArtists := make([]*checkedArtist, 0)
	if *c.trackNumberingGaps {
		checkedArtists = toCheckedArtists(artists)
		gapsFound := false
		for _, cAr := range checkedArtists {
			for _, cAl := range cAr.albums {
				m := make(map[int]*checkedTrack)
				for _, cT := range cAl.tracks {
					if priorCT, ok := m[cT.backing.Number()]; ok {
						cAl.issues = append(cAl.issues, fmt.Sprintf("track %d used by %q and %q", cT.backing.Number(), priorCT.backing.Name(), cT.backing.Name()))
						gapsFound = true
					} else {
						m[cT.backing.Number()] = cT
					}
				}
				c := 0
				for n := 1; n <= len(cAl.tracks); n++ {
					if _, ok := m[n]; !ok {
						c++
						cAl.issues = append(cAl.issues, fmt.Sprintf("missing track %d", n))
						gapsFound = true
					}
				}
				maxNumber := len(cAl.tracks) + c
				validTracks := fmt.Sprintf("valid tracks are 1..%d", maxNumber)
				for n, t := range m {
					switch {
					case n < 1:
						cAl.issues = append(cAl.issues, fmt.Sprintf("track %d (%q) is not a valid track number; %s", n, t.backing.Name(), validTracks))
						gapsFound = true
					case n > maxNumber:
						cAl.issues = append(cAl.issues, fmt.Sprintf("track %d (%q) is not a valid track number; %s", n, t.backing.Name(), validTracks))
						gapsFound = true
					}
				}
			}
		}
		if !gapsFound {
			o.WriteCanonicalConsole("Check Gaps: no gaps found")
		}
	}
	return checkedArtists
}
