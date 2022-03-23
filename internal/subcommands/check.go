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
	checkEmptyFolders         *bool
	checkGapsInTrackNumbering *bool
	checkIntegrity            *bool
	commons                   *CommonCommandFlags
}

func (c *check) name() string {
	return c.commons.name()
}

func newCheck(fSet *flag.FlagSet) CommandProcessor {
	return &check{
		checkEmptyFolders:         fSet.Bool("empty", false, "check for empty artist and album folders"),
		checkGapsInTrackNumbering: fSet.Bool("gaps", false, "check for gaps in track numbers"),
		checkIntegrity:            fSet.Bool("integrity", true, "check for disagreement between the file system and audio file metadata"),
		commons:                   newCommonCommandFlags(fSet),
	}
}

func (c *check) Exec(args []string) {
	if params := c.commons.processArgs(os.Stderr, args); params != nil {
		c.runSubcommand(os.Stdout)
	}
}

func (c *check) runSubcommand(w io.Writer) {
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
		var artists []*files.Artist
		if *c.checkEmptyFolders {
			artists = files.LoadUnfilteredData(*c.commons.topDirectory, *c.commons.fileExtension)
			if len(artists) == 0 {
				logrus.WithFields(
					logrus.Fields{
						"topDirectory":  *c.commons.topDirectory,
						"fileExtension": *c.commons.fileExtension,
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
		if *c.checkGapsInTrackNumbering || *c.checkIntegrity {
			searchParams := files.NewDirectorySearchParams(*c.commons.topDirectory, *c.commons.fileExtension, *c.commons.albumRegex, *c.commons.artistRegex)
			if len(artists) == 0 {
				artists = files.LoadData(searchParams)
			} else {
				// filter existing artists using provided filters
				artists = files.FilterArtists(artists, searchParams)
			}
			if *c.checkGapsInTrackNumbering {
				var complaints []string
				for _, artist := range artists{
					for _, album := range artist.Albums {
						albumId := fmt.Sprintf("Artist: %q album %q", artist.Name, album.Name )
						m := make(map[int]*files.Track)
						for _, track := range album.Tracks {
							if t, ok := m[track.TrackNumber]; ok {
								complaints = append(complaints, fmt.Sprintf("%s: track %d used by %q and %q", albumId, track.TrackNumber, t.Name, track.Name))
							} else {
								m[track.TrackNumber] = track
							}
						}
						validTracks := fmt.Sprintf("valid tracks are 1..%d", len(album.Tracks))
						for trackNumber, track := range m {
							switch {
								case trackNumber < 1:								
								complaints = append(complaints, fmt.Sprintf("%s: track %d (%q) is not a valid track number; %s", albumId, trackNumber, track.Name, validTracks))
							case trackNumber > len(album.Tracks):
								complaints = append(complaints, fmt.Sprintf("%s: track %d (%q) is not a valid track number; %s", albumId, trackNumber, track.Name, validTracks))
							}
						}
						for trackNumber := 1; trackNumber <= len(album.Tracks); trackNumber++ {
							if _, ok := m[trackNumber]; !ok {
								complaints = append(complaints, fmt.Sprintf("%s: missing track %d", albumId, trackNumber))
							}
						}
					} // each album
				} // each artist
				if len(complaints) > 0 {
					sort.Strings(complaints)
					fmt.Fprintf(w, "Check Gaps\n%s\n", strings.Join(complaints, "\n"))
				} else {
					fmt.Fprintln(w, "Check Gaps: no gaps found")
				}
			}
		}
	}
}
