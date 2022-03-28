package subcommands

import (
	"flag"
	"fmt"
	"io"
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
	ff                        *files.FileFlags
}

func (c *check) name() string {
	return c.n
}

func newCheck(fSet *flag.FlagSet) CommandProcessor {
	return &check{
		n:                         fSet.Name(),
		checkEmptyFolders:         fSet.Bool("empty", false, "check for empty artist and album folders"),
		checkGapsInTrackNumbering: fSet.Bool("gaps", false, "check for gaps in track numbers"),
		checkIntegrity:            fSet.Bool("integrity", true, "check for disagreement between the file system and audio file metadata"),
		ff:                        files.NewFileFlags(fSet),
	}
}

func (c *check) Exec(args []string) {
	if s := c.ff.ProcessArgs(os.Stderr, args); s != nil {
		c.runSubcommand(os.Stdout, s)
	}
}

func (c *check) runSubcommand(w io.Writer, s *files.Search) {
	if !*c.checkEmptyFolders && !*c.checkGapsInTrackNumbering && !*c.checkIntegrity {
		fmt.Fprintf(os.Stderr, "%s: nothing to do!", c.name())
		logrus.WithFields(logrus.Fields{"subcommand name": c.name()}).Error("nothing to do")
	} else {
		logrus.WithFields(logrus.Fields{
			"subcommandName":    c.name(),
			"checkEmptyFolders": *c.checkEmptyFolders,
			"checkTrackGaps":    *c.checkGapsInTrackNumbering,
			"checkIntegrity":    *c.checkIntegrity,
		}).Info("subcommand")
		artists := performEmptyFolderAnalysis(w, c, s)
		// filter existing artists using provided filters
		artists = filterArtists(c, s, artists)
		c.performGapAnalysis(w, c, artists)
	}
}

func filterArtists(c *check, s *files.Search, artists []*files.Artist) (filteredArtists []*files.Artist) {
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

func performEmptyFolderAnalysis(w io.Writer, c *check, s *files.Search) (artists []*files.Artist) {
	if *c.checkEmptyFolders {
		artists = s.LoadUnfilteredData()
		if len(artists) == 0 {
			logrus.WithFields(
				logrus.Fields{
					"topDirectory":  s.TopDirectory(),
					"fileExtension": s.TargetExtension(),
				}).Error("checking empty folders, no artists found")
		}
		var complaints []string
		for _, artist := range artists {
			if len(artist.Albums) == 0 {
				complaints = append(complaints, fmt.Sprintf("Artist %q: no albums found", artist.Name))
			} else {
				for _, album := range artist.Albums {
					if len(album.Tracks) == 0 {
						complaints = append(complaints, fmt.Sprintf("Artist %q album %q: no tracks found", artist.Name, album.Name))
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

func (*check) performGapAnalysis(w io.Writer, c *check, artists []*files.Artist) {
	if *c.checkGapsInTrackNumbering {
		var complaints []string
		for _, artist := range artists {
			for _, album := range artist.Albums {
				albumId := fmt.Sprintf("Artist: %q album %q", artist.Name, album.Name)
				m := make(map[int]*files.Track)
				for _, track := range album.Tracks {
					if t, ok := m[track.TrackNumber]; ok {
						complaints = append(complaints, fmt.Sprintf("%s: track %d used by %q and %q", albumId, track.TrackNumber, t.Name, track.Name))
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
						complaints = append(complaints, fmt.Sprintf("%s: track %d (%q) is not a valid track number; %s", albumId, trackNumber, track.Name, validTracks))
					case trackNumber > expectedTrackCount:
						complaints = append(complaints, fmt.Sprintf("%s: track %d (%q) is not a valid track number; %s", albumId, trackNumber, track.Name, validTracks))
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
