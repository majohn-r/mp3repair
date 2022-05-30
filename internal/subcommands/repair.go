package subcommands

import (
	"flag"
	"fmt"
	"io"
	"mp3/internal"
	"mp3/internal/files"
	"os"
	"sort"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type repair struct {
	n      string
	dryRun *bool
	sf     *files.SearchFlags
}

func (r *repair) name() string {
	return r.n
}

func newRepair(v *viper.Viper, fSet *flag.FlagSet) CommandProcessor {
	return newRepairSubCommand(v, fSet)
}

const (
	dryRunFlag    = "dryRun"
	defaultDryRun = false
)

func newRepairSubCommand(v *viper.Viper, fSet *flag.FlagSet) *repair {
	subViper := internal.SafeSubViper(v, "repair")
	return &repair{
		n: fSet.Name(),
		dryRun: fSet.Bool(dryRunFlag,
			internal.GetBoolDefault(subViper, dryRunFlag, defaultDryRun),
			"if true, output what would have repaired, but make no repairs"),
		sf: files.NewSearchFlags(v, fSet),
	}
}

func (r *repair) Exec(w io.Writer, args []string) {
	if s := r.sf.ProcessArgs(os.Stderr, args); s != nil {
		r.runSubcommand(w, s)
	}
}

func (r *repair) logFields() logrus.Fields {
	return logrus.Fields{internal.LOG_COMMAND_NAME: r.name(), dryRunFlag: *r.dryRun}
}

func (r *repair) runSubcommand(w io.Writer, s *files.Search) {
	logrus.WithFields(r.logFields()).Info(internal.LOG_EXECUTING_COMMAND)
	artists := s.LoadData()
	files.UpdateTracks(artists, files.RawReadTags)
	tracksWithConflicts := findConflictedTracks(artists)
	if *r.dryRun {
		reportTracks(w, tracksWithConflicts)
	} else {
		fixTracks(w, tracksWithConflicts)
	}
}

func findConflictedTracks(artists []*files.Artist) []*files.Track {
	var t []*files.Track
	for _, artist := range artists {
		for _, album := range artist.Albums {
			for _, track := range album.Tracks {
				if state := track.AnalyzeIssues(); state.HasTaggingConflicts() {
					t = append(t, track)
				}
			}
		}
	}
	sort.Sort(files.Tracks(t))
	return t
}

const noProblemsFound = "No repairable track defects found"

func reportTracks(w io.Writer, tracks []*files.Track) {
	if len(tracks) == 0 {
		fmt.Fprintln(w, noProblemsFound)
	} else {
		lastArtistName := ""
		lastAlbumName := ""
		for _, t := range tracks {
			albumName := t.ContainingAlbum.Name
			artistName := t.ContainingAlbum.RecordingArtist.Name
			if lastArtistName != artistName {
				fmt.Fprintf(w, "%q\n", artistName)
				lastArtistName = artistName
				lastAlbumName = ""
			}
			if albumName != lastAlbumName {
				fmt.Fprintf(w, "    %q\n", albumName)
				lastAlbumName = albumName
			}
			s := t.AnalyzeIssues()
			fmt.Fprintf(w, "        %2d %q need to fix%s%s%s%s\n",
				t.TrackNumber, t.Name,
				reportProblem(s.HasNumberingConflict(), " track numbering;"),
				reportProblem(s.HasTrackNameConflict(), " track name;"),
				reportProblem(s.HasAlbumNameConflict(), " album name;"),
				reportProblem(s.HasArtistNameConflict(), " artist name;"))
		}
	}
}

func reportProblem(b bool, problem string) string {
	if !b {
		return ""
	}
	return problem
}

func fixTracks(w io.Writer, tracks []*files.Track) {

}
