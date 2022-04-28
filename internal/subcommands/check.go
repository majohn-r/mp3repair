package subcommands

import (
	"flag"
	"fmt"
	"io"
	"mp3/internal"
	"mp3/internal/files"
	"os"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
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

func newCheck(fSet *flag.FlagSet) CommandProcessor {
	return newCheckSubCommand(fSet)
}

func newCheckSubCommand(fSet *flag.FlagSet) *check {
	return &check{
		n:                         fSet.Name(),
		checkEmptyFolders:         fSet.Bool("empty", false, "check for empty artist and album folders"),
		checkGapsInTrackNumbering: fSet.Bool("gaps", false, "check for gaps in track numbers"),
		checkIntegrity:            fSet.Bool("integrity", true, "check for disagreement between the file system and audio file metadata"),
		sf:                        files.NewSearchFlags(fSet),
	}
}

func (c *check) Exec(w io.Writer, args []string) {
	if s := c.sf.ProcessArgs(os.Stderr, args); s != nil {
		c.runSubcommand(w, s)
	}
}

const (
	logEmptyFoldersFlag string = "emptyFolders"
	logIntegrityFlag    string = "integrityAnalysis"
	logTrackGapFlag     string = "gapAnalysis"
)

func (c *check) logFields() logrus.Fields {
	return logrus.Fields{
		internal.LOG_COMMAND_NAME: c.name(),
		logEmptyFoldersFlag:       *c.checkEmptyFolders,
		logTrackGapFlag:           *c.checkGapsInTrackNumbering,
		logIntegrityFlag:          *c.checkIntegrity,
	}
}

func (c *check) runSubcommand(w io.Writer, s *files.Search) {
	if !*c.checkEmptyFolders && !*c.checkGapsInTrackNumbering && !*c.checkIntegrity {
		fmt.Fprintf(os.Stderr, internal.USER_SPECIFIED_NO_WORK, c.name())
		logrus.WithFields(c.logFields()).Error(internal.LOG_NOTHING_TO_DO)
	} else {
		logrus.WithFields(c.logFields()).Info(internal.LOG_EXECUTING_COMMAND)
		artists := c.performEmptyFolderAnalysis(w, s)
		artists = c.filterArtists(s, artists)
		c.performGapAnalysis(w, artists)
		c.performIntegrityCheck(w, artists)
	}
}

func (c *check) filterArtists(s *files.Search, artists []*files.Artist) (filteredArtists []*files.Artist) {
	if *c.checkGapsInTrackNumbering || *c.checkIntegrity {
		if len(artists) == 0 {
			filteredArtists = s.LoadData()
		} else {
			filteredArtists = s.FilterArtists(artists)
		}
	} else {
		filteredArtists = artists
	}
	return
}

func (c *check) performEmptyFolderAnalysis(w io.Writer, s *files.Search) (artists []*files.Artist) {
	if *c.checkEmptyFolders {
		artists = s.LoadUnfilteredData()
		if len(artists) == 0 {
			logrus.WithFields(logrus.Fields{
				internal.LOG_DIRECTORY: s.TopDirectory(),
				internal.LOG_EXTENSION: s.TargetExtension(),
			}).Error(internal.LOG_NO_ARTIST_DIRECTORIES)
		}
		var complaints []string
		for _, artist := range artists {
			if len(artist.Albums) == 0 {
				complaints = append(complaints, fmt.Sprintf("Artist %q: no albums found", artist.Name))
			} else {
				for _, album := range artist.Albums {
					if len(album.Tracks) == 0 {
						complaint := fmt.Sprintf("Artist %q album %q: no tracks found", artist.Name, album.Name)
						complaints = append(complaints, complaint)
					}
				}
			}
		}
		if len(complaints) > 0 {
			sort.Strings(complaints)
			fmt.Fprintf(w, "Empty Folder Analysis\n%s\n", strings.Join(complaints, "\n"))
		} else {
			fmt.Fprintln(w, "Empty Folder Analysis: no empty folders found")
		}
	}
	return
}

func (c *check) performIntegrityCheck(w io.Writer, artists []*files.Artist) {
	if *c.checkIntegrity {
		files.UpdateTracks(artists, files.RawReadTags)
		for _, artist := range artists {
			for _, album := range artist.Albums{
				for _, track := range album.Tracks {
					differences := track.FindDifferences()
					if len(differences) > 0 {
						fmt.Fprintf(w, "%q: %q: %q\n%s\n\n", artist.Name, album.Name, track.Name, differences)
					}
				}
			}
		}
	}
}

func (c *check) performGapAnalysis(w io.Writer, artists []*files.Artist) {
	if *c.checkGapsInTrackNumbering {
		var complaints []string
		for _, artist := range artists {
			for _, album := range artist.Albums {
				albumId := fmt.Sprintf("Artist: %q album %q", artist.Name, album.Name)
				m := make(map[int]*files.Track)
				for _, track := range album.Tracks {
					if t, ok := m[track.TrackNumber]; ok {
						complaint := fmt.Sprintf("%s: track %d used by %q and %q", albumId, track.TrackNumber, t.Name, track.Name)
						complaints = append(complaints, complaint)
					} else {
						m[track.TrackNumber] = track
					}
				}
				missingTracks := 0
				for trackNumber := 1; trackNumber <= len(album.Tracks); trackNumber++ {
					if _, ok := m[trackNumber]; !ok {
						missingTracks++
						complaints = append(complaints, fmt.Sprintf("%s: missing track %d", albumId, trackNumber))
					}
				}
				expectedTrackCount := len(album.Tracks) + missingTracks
				validTracks := fmt.Sprintf("valid tracks are 1..%d", expectedTrackCount)
				for trackNumber, track := range m {
					switch {
					case trackNumber < 1:
						complaint := fmt.Sprintf("%s: track %d (%q) is not a valid track number; %s", albumId, trackNumber, track.Name, validTracks)
						complaints = append(complaints, complaint)
					case trackNumber > expectedTrackCount:
						complaint := fmt.Sprintf("%s: track %d (%q) is not a valid track number; %s", albumId, trackNumber, track.Name, validTracks)
						complaints = append(complaints, complaint)
					}
				}
			}
		}
		if len(complaints) > 0 {
			sort.Strings(complaints)
			fmt.Fprintf(w, "Check Gaps\n%s\n", strings.Join(complaints, "\n"))
		} else {
			fmt.Fprintln(w, "Check Gaps: no gaps found")
		}
	}
}
