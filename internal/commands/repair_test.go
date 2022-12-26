package commands

import (
	"flag"
	"fmt"
	"mp3/internal"
	"mp3/internal/files"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/majohn-r/output"
)

func Test_newRepairCommand(t *testing.T) {
	fnName := "newRepairCommand()"
	savedState := internal.SaveEnvVarForTesting("APPDATA")
	os.Setenv("APPDATA", internal.SecureAbsolutePathForTesting("."))
	defer func() {
		savedState.RestoreForTesting()
	}()
	topDir := "loadTest"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, topDir, err)
	}
	if err := internal.PopulateTopDirForTesting(topDir); err != nil {
		t.Errorf("%s error populating %q: %v", fnName, topDir, err)
	}
	if err := internal.CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("%s error creating defaults.yaml: %v", fnName, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
		internal.DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	defaultConfig, _ := internal.ReadConfigurationFile(output.NewNilBus())
	type args struct {
		c *internal.Configuration
	}
	tests := []struct {
		name string
		args
		wantDryRun bool
		wantOk     bool
		output.WantedRecording
	}{
		{
			name:       "ordinary defaults",
			args:       args{c: internal.EmptyConfiguration()},
			wantDryRun: false,
			wantOk:     true,
		},
		{
			name:       "overridden defaults",
			args:       args{c: defaultConfig},
			wantDryRun: true,
			wantOk:     true,
		},
		{
			name: "bad dryRun default",
			args: args{
				c: internal.CreateConfiguration(output.NewNilBus(), map[string]any{
					"repair": map[string]any{
						"dryRun": 42,
					},
				}),
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"repair\": invalid boolean value \"42\" for -dryRun: parse error.\n",
				Log:   "level='error' error='invalid boolean value \"42\" for -dryRun: parse error' section='repair' msg='invalid content in configuration file'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			repair, gotOk := newRepairCommand(o, tt.args.c, flag.NewFlagSet("repair", flag.ContinueOnError))
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk %t wantOk %t", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
			if repair != nil {
				if _, ok := repair.sf.ProcessArgs(output.NewNilBus(), []string{
					"-topDir", topDir,
					"-ext", ".mp3",
				}); ok {
					if *repair.dryRun != tt.wantDryRun {
						t.Errorf("%s %q: got dryRun %t want %t", fnName, tt.name, *repair.dryRun, tt.wantDryRun)
					}
				} else {
					t.Errorf("%s %q: error processing arguments", fnName, tt.name)
				}
			}
		})
	}
}

func newRepairForTesting() *repair {
	r, _ := newRepairCommand(output.NewNilBus(), internal.EmptyConfiguration(), flag.NewFlagSet("repair", flag.ContinueOnError))
	return r
}

func Test_repair_Exec(t *testing.T) {
	fnName := "repair.Exec()"
	topDirName := "repairExec"
	topDirWithContent := "realContent"
	if err := internal.Mkdir(topDirName); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, topDirName, err)
	}
	if err := internal.Mkdir(topDirWithContent); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, topDirWithContent, err)
	}
	appFolder := filepath.Join(topDirName, "mp3")
	if err := internal.Mkdir(appFolder); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, appFolder, err)
	}
	savedHome := internal.SaveEnvVarForTesting("HOMEPATH")
	home := internal.SavedEnvVar{
		Name:  "HOMEPATH",
		Value: "C:\\Users\\The User",
		Set:   true,
	}
	home.RestoreForTesting()
	savedDirtyFolderFound := dirtyFolderFound
	savedDirtyFolder := dirtyFolder
	savedDirtyFolderValid := dirtyFolderValid
	savedMarkDirtyAttempted := markDirtyAttempted
	defer func() {
		savedHome.RestoreForTesting()
		dirtyFolderFound = savedDirtyFolderFound
		dirtyFolder = savedDirtyFolder
		dirtyFolderValid = savedDirtyFolderValid
		markDirtyAttempted = savedMarkDirtyAttempted
		internal.DestroyDirectoryForTesting(fnName, topDirName)
		internal.DestroyDirectoryForTesting(fnName, topDirWithContent)
	}()
	if err := internal.PopulateTopDirForTesting(topDirName); err != nil {
		t.Errorf("%s error populating directory %q: %v", fnName, topDirName, err)
	}
	artist := "new artist"
	if err := internal.Mkdir(filepath.Join(topDirWithContent, artist)); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, filepath.Join(topDirWithContent, artist), err)
	}
	album := "new album"
	if err := internal.Mkdir(filepath.Join(topDirWithContent, artist, album)); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, filepath.Join(topDirWithContent, artist, album), err)
	}
	frames := map[string]string{
		"TYER": "2022",
		"TALB": "unknown album",
		"TRCK": "2",
		"TCON": "dance music",
		"TCOM": "a couple of idiots",
		"TIT2": "unknown track",
		"TPE1": "unknown artist",
		"TLEN": "1000",
	}
	content := createTaggedContent(frames)
	trackName := "01 new track.mp3"
	if err := internal.CreateFileForTestingWithContent(filepath.Join(topDirWithContent, artist, album), trackName, content); err != nil {
		t.Errorf("%s error creating file %q: %v", fnName, filepath.Join(topDirWithContent, artist, album, trackName), err)
	}
	type args struct {
		args []string
	}
	tests := []struct {
		name string
		r    *repair
		args
		output.WantedRecording
	}{
		{
			name: "help",
			r:    newRepairForTesting(),
			args: args{[]string{"--help"}},
			WantedRecording: output.WantedRecording{
				Error: "Usage of repair:\n" +
					"  -albumFilter regular expression\n" +
					"    \tregular expression specifying which albums to select (default \".*\")\n" +
					"  -artistFilter regular expression\n" +
					"    \tregular expression specifying which artists to select (default \".*\")\n" +
					"  -dryRun\n" +
					"    \toutput what would have been repaired, but make no repairs (default false)\n" +
					"  -ext extension\n" +
					"    \textension identifying music files (default \".mp3\")\n" +
					"  -topDir directory\n" +
					"    \ttop directory specifying where to find music files (default \"C:\\\\Users\\\\The User\\\\Music\")\n",
				Log: "level='error' arguments='[--help]' msg='flag: help requested'\n",
			},
		},
		{
			name: "dry run, no usable content",
			r:    newRepairForTesting(),
			args: args{[]string{"-topDir", topDirName, "-dryRun"}},
			WantedRecording: output.WantedRecording{
				Console: "No repairable track defects found.\n",
				Error:   "Reading track metadata.\n" + generateStandardTrackErrorReport(),
				Log: "level='info' -dryRun='true' command='repair' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='repairExec' msg='reading filtered music files'\n" +
					generateStandardTrackLogReport(),
			},
		},
		{
			name: "real repair, no usable content",
			r:    newRepairForTesting(),
			args: args{[]string{"-topDir", topDirName, "-dryRun=false"}},
			WantedRecording: output.WantedRecording{
				Console: "No repairable track defects found.\n",
				Error:   "Reading track metadata.\n" + generateStandardTrackErrorReport(),
				Log: "level='info' -dryRun='false' command='repair' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='repairExec' msg='reading filtered music files'\n" +
					generateStandardTrackLogReport(),
			},
		},
		{
			name: "dry run, usable content",
			r:    newRepairForTesting(),
			args: args{[]string{"-topDir", topDirWithContent, "-dryRun"}},
			WantedRecording: output.WantedRecording{
				Console: strings.Join([]string{
					"\"new artist\"",
					"    \"new album\"",
					"         1 \"new track\" need to repair track numbering; track name; album name; artist name;\n",
				}, "\n"),
				Error: "Reading track metadata.\n" +
					"An error occurred when trying to read ID3V1 tag information for track \"new track\" on album \"new album\" by artist \"new artist\": \"no id3v1 tag found in file \\\"realContent\\\\\\\\new artist\\\\\\\\new album\\\\\\\\01 new track.mp3\\\"\".\n",
				Log: "level='info' -dryRun='true' command='repair' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='realContent' msg='reading filtered music files'\n" +
					"level='error' error='no id3v1 tag found in file \"realContent\\\\new artist\\\\new album\\\\01 new track.mp3\"' track='realContent\\new artist\\new album\\01 new track.mp3' msg='id3v1 tag error'\n",
			},
		},
		{
			name: "real repair, usable content",
			r:    newRepairForTesting(),
			args: args{[]string{"-topDir", topDirWithContent, "-dryRun=false"}},
			WantedRecording: output.WantedRecording{
				Console: "The track \"realContent\\\\new artist\\\\new album\\\\01 new track.mp3\" has been backed up to \"realContent\\\\new artist\\\\new album\\\\pre-repair-backup\\\\1.mp3\".\n" +
					"\"realContent\\\\new artist\\\\new album\\\\01 new track.mp3\" repaired.\n",
				Error: "Reading track metadata.\n" +
					"An error occurred when trying to read ID3V1 tag information for track \"new track\" on album \"new album\" by artist \"new artist\": \"no id3v1 tag found in file \\\"realContent\\\\\\\\new artist\\\\\\\\new album\\\\\\\\01 new track.mp3\\\"\".\n",
				Log: "level='info' -dryRun='false' command='repair' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='realContent' msg='reading filtered music files'\n" +
					"level='error' error='no id3v1 tag found in file \"realContent\\\\new artist\\\\new album\\\\01 new track.mp3\"' track='realContent\\new artist\\new album\\01 new track.mp3' msg='id3v1 tag error'\n" +
					"level='info' fileName='repairExec\\mp3\\metadata.dirty' msg='metadata dirty file written'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// fake out code for MarkDirty()
			markDirtyAttempted = false
			dirtyFolder = appFolder
			dirtyFolderFound = true
			dirtyFolderValid = true
			o := output.NewRecorder()
			tt.r.Exec(o, tt.args.args)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func generateStandardTrackErrorReport() string {
	var result []string
	for artist := 0; artist < 10; artist++ {
		for album := 0; album < 10; album++ {
			for track := 0; track < 10; track++ {
				var sep string
				if track%2 == 0 {
					sep = "-"
				} else {
					sep = " "
				}
				result = append(result,
					fmt.Sprintf("An error occurred when trying to read ID3V1 tag information for track \"Test Track[%02d]\" on album \"Test Album %d\" by artist \"Test Artist %d\": \"seek repairExec\\\\Test Artist %d\\\\Test Album %d\\\\%02d%sTest Track[%02d].mp3: An attempt was made to move the file pointer before the beginning of the file.\".\n", track, album, artist, artist, album, track, sep, track),
					fmt.Sprintf("An error occurred when trying to read ID3V2 tag information for track \"Test Track[%02d]\" on album \"Test Album %d\" by artist \"Test Artist %d\": \"zero length\".\n", track, album, artist))
			}
		}
	}
	return strings.Join(result, "")
}

func generateStandardTrackLogReport() string {
	var result []string
	for artist := 0; artist < 10; artist++ {
		for album := 0; album < 10; album++ {
			for track := 0; track < 10; track++ {
				var sep string
				if track%2 == 0 {
					sep = "-"
				} else {
					sep = " "
				}
				result = append(result,
					fmt.Sprintf("level='error' error='seek repairExec\\Test Artist %d\\Test Album %d\\%02d%sTest Track[%02d].mp3: An attempt was made to move the file pointer before the beginning of the file.' track='repairExec\\Test Artist %d\\Test Album %d\\%02d%sTest Track[%02d].mp3' msg='id3v1 tag error'\n", artist, album, track, sep, track, artist, album, track, sep, track),
					fmt.Sprintf("level='error' error='zero length' track='repairExec\\Test Artist %d\\Test Album %d\\%02d%sTest Track[%02d].mp3' msg='id3v2 tag error'\n", artist, album, track, sep, track))
			}
		}
	}
	return strings.Join(result, "")
}

func Test_albumPaths(t *testing.T) {
	fnName := "albumPaths()"
	topDir := "albumPaths"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, topDir, err)
	}
	if err := internal.PopulateTopDirForTesting(topDir); err != nil {
		t.Errorf("%s error populating %q: %v", fnName, topDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	s := files.CreateFilteredSearchForTesting(topDir, "^.*$", "^.*$")
	a, _ := s.LoadData(output.NewNilBus())
	var tSlice []*files.Track
	for _, artist := range a {
		for _, album := range artist.Albums() {
			tSlice = append(tSlice, album.Tracks()...)
		}
	}
	type args struct {
		tracks []*files.Track
	}
	tests := []struct {
		name string
		args
		want []string
	}{
		{name: "degenerate case", args: args{}},
		{name: "full blown case", args: args{tracks: tSlice}, want: []string{
			"albumPaths\\Test Artist 0\\Test Album 0",
			"albumPaths\\Test Artist 0\\Test Album 1",
			"albumPaths\\Test Artist 0\\Test Album 2",
			"albumPaths\\Test Artist 0\\Test Album 3",
			"albumPaths\\Test Artist 0\\Test Album 4",
			"albumPaths\\Test Artist 0\\Test Album 5",
			"albumPaths\\Test Artist 0\\Test Album 6",
			"albumPaths\\Test Artist 0\\Test Album 7",
			"albumPaths\\Test Artist 0\\Test Album 8",
			"albumPaths\\Test Artist 0\\Test Album 9",
			"albumPaths\\Test Artist 1\\Test Album 0",
			"albumPaths\\Test Artist 1\\Test Album 1",
			"albumPaths\\Test Artist 1\\Test Album 2",
			"albumPaths\\Test Artist 1\\Test Album 3",
			"albumPaths\\Test Artist 1\\Test Album 4",
			"albumPaths\\Test Artist 1\\Test Album 5",
			"albumPaths\\Test Artist 1\\Test Album 6",
			"albumPaths\\Test Artist 1\\Test Album 7",
			"albumPaths\\Test Artist 1\\Test Album 8",
			"albumPaths\\Test Artist 1\\Test Album 9",
			"albumPaths\\Test Artist 2\\Test Album 0",
			"albumPaths\\Test Artist 2\\Test Album 1",
			"albumPaths\\Test Artist 2\\Test Album 2",
			"albumPaths\\Test Artist 2\\Test Album 3",
			"albumPaths\\Test Artist 2\\Test Album 4",
			"albumPaths\\Test Artist 2\\Test Album 5",
			"albumPaths\\Test Artist 2\\Test Album 6",
			"albumPaths\\Test Artist 2\\Test Album 7",
			"albumPaths\\Test Artist 2\\Test Album 8",
			"albumPaths\\Test Artist 2\\Test Album 9",
			"albumPaths\\Test Artist 3\\Test Album 0",
			"albumPaths\\Test Artist 3\\Test Album 1",
			"albumPaths\\Test Artist 3\\Test Album 2",
			"albumPaths\\Test Artist 3\\Test Album 3",
			"albumPaths\\Test Artist 3\\Test Album 4",
			"albumPaths\\Test Artist 3\\Test Album 5",
			"albumPaths\\Test Artist 3\\Test Album 6",
			"albumPaths\\Test Artist 3\\Test Album 7",
			"albumPaths\\Test Artist 3\\Test Album 8",
			"albumPaths\\Test Artist 3\\Test Album 9",
			"albumPaths\\Test Artist 4\\Test Album 0",
			"albumPaths\\Test Artist 4\\Test Album 1",
			"albumPaths\\Test Artist 4\\Test Album 2",
			"albumPaths\\Test Artist 4\\Test Album 3",
			"albumPaths\\Test Artist 4\\Test Album 4",
			"albumPaths\\Test Artist 4\\Test Album 5",
			"albumPaths\\Test Artist 4\\Test Album 6",
			"albumPaths\\Test Artist 4\\Test Album 7",
			"albumPaths\\Test Artist 4\\Test Album 8",
			"albumPaths\\Test Artist 4\\Test Album 9",
			"albumPaths\\Test Artist 5\\Test Album 0",
			"albumPaths\\Test Artist 5\\Test Album 1",
			"albumPaths\\Test Artist 5\\Test Album 2",
			"albumPaths\\Test Artist 5\\Test Album 3",
			"albumPaths\\Test Artist 5\\Test Album 4",
			"albumPaths\\Test Artist 5\\Test Album 5",
			"albumPaths\\Test Artist 5\\Test Album 6",
			"albumPaths\\Test Artist 5\\Test Album 7",
			"albumPaths\\Test Artist 5\\Test Album 8",
			"albumPaths\\Test Artist 5\\Test Album 9",
			"albumPaths\\Test Artist 6\\Test Album 0",
			"albumPaths\\Test Artist 6\\Test Album 1",
			"albumPaths\\Test Artist 6\\Test Album 2",
			"albumPaths\\Test Artist 6\\Test Album 3",
			"albumPaths\\Test Artist 6\\Test Album 4",
			"albumPaths\\Test Artist 6\\Test Album 5",
			"albumPaths\\Test Artist 6\\Test Album 6",
			"albumPaths\\Test Artist 6\\Test Album 7",
			"albumPaths\\Test Artist 6\\Test Album 8",
			"albumPaths\\Test Artist 6\\Test Album 9",
			"albumPaths\\Test Artist 7\\Test Album 0",
			"albumPaths\\Test Artist 7\\Test Album 1",
			"albumPaths\\Test Artist 7\\Test Album 2",
			"albumPaths\\Test Artist 7\\Test Album 3",
			"albumPaths\\Test Artist 7\\Test Album 4",
			"albumPaths\\Test Artist 7\\Test Album 5",
			"albumPaths\\Test Artist 7\\Test Album 6",
			"albumPaths\\Test Artist 7\\Test Album 7",
			"albumPaths\\Test Artist 7\\Test Album 8",
			"albumPaths\\Test Artist 7\\Test Album 9",
			"albumPaths\\Test Artist 8\\Test Album 0",
			"albumPaths\\Test Artist 8\\Test Album 1",
			"albumPaths\\Test Artist 8\\Test Album 2",
			"albumPaths\\Test Artist 8\\Test Album 3",
			"albumPaths\\Test Artist 8\\Test Album 4",
			"albumPaths\\Test Artist 8\\Test Album 5",
			"albumPaths\\Test Artist 8\\Test Album 6",
			"albumPaths\\Test Artist 8\\Test Album 7",
			"albumPaths\\Test Artist 8\\Test Album 8",
			"albumPaths\\Test Artist 8\\Test Album 9",
			"albumPaths\\Test Artist 9\\Test Album 0",
			"albumPaths\\Test Artist 9\\Test Album 1",
			"albumPaths\\Test Artist 9\\Test Album 2",
			"albumPaths\\Test Artist 9\\Test Album 3",
			"albumPaths\\Test Artist 9\\Test Album 4",
			"albumPaths\\Test Artist 9\\Test Album 5",
			"albumPaths\\Test Artist 9\\Test Album 6",
			"albumPaths\\Test Artist 9\\Test Album 7",
			"albumPaths\\Test Artist 9\\Test Album 8",
			"albumPaths\\Test Artist 9\\Test Album 9",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := albumPaths(tt.args.tracks); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_repair_makeBackupDirectories(t *testing.T) {
	fnName := "repair.makeBackupDirectories()"
	topDir := "makeBackupDirectories"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, topDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	backupDir := files.CreateBackupPath(topDir)
	if err := internal.Mkdir(backupDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, backupDir, err)
	}
	albumDir := filepath.Join(topDir, "album")
	if err := internal.Mkdir(albumDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, albumDir, err)
	}
	if err := internal.CreateNamedFileForTesting(files.CreateBackupPath(albumDir), []byte("nonsense content")); err != nil {
		t.Errorf("%s error creating file %q in %q: %v", fnName, files.CreateBackupPath(albumDir), albumDir, err)
	}
	albumDir2 := filepath.Join(topDir, "album2")
	if err := internal.Mkdir(albumDir2); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, albumDir2, err)
	}
	fFlag := false
	type args struct {
		paths []string
	}
	tests := []struct {
		name string
		r    *repair
		args
		output.WantedRecording
	}{
		{name: "degenerate case", r: &repair{dryRun: &fFlag}, args: args{paths: nil}},
		{
			name: "useful case",
			r:    &repair{dryRun: &fFlag},
			args: args{paths: []string{topDir, albumDir, albumDir2}},
			WantedRecording: output.WantedRecording{
				Error: "The directory \"makeBackupDirectories\\\\album\\\\pre-repair-backup\" cannot be created: file exists and is not a directory.\n",
				Log:   "level='error' command='repair' directory='makeBackupDirectories\\album\\pre-repair-backup' error='file exists and is not a directory' msg='cannot create directory'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			makeBackupDirectories(o, tt.args.paths)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_repair_backupTracks(t *testing.T) {
	fnName := "repair.backupTracks()"
	topDir := "backupTracks"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, topDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	goodTrackName := "1 good track.mp3"
	if err := internal.CreateFileForTesting(topDir, goodTrackName); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, goodTrackName, err)
	}
	if err := internal.Mkdir(files.CreateBackupPath(topDir)); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, files.CreateBackupPath(topDir), err)
	}
	if err := internal.Mkdir(filepath.Join(files.CreateBackupPath(topDir), "2.mp3")); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, "2.mp3", err)
	}
	fFlag := false
	type args struct {
		tracks []*files.Track
	}
	tests := []struct {
		name string
		r    *repair
		args
		output.WantedRecording
	}{
		{name: "degenerate case", r: &repair{dryRun: &fFlag}, args: args{tracks: nil}},
		{
			name: "real tests",
			r:    &repair{dryRun: &fFlag},
			args: args{
				tracks: []*files.Track{
					files.NewTrack(files.NewAlbum("", nil, topDir), goodTrackName, "", 1),
					files.NewTrack(files.NewAlbum("", nil, topDir), "dup track", "", 1),
					files.NewTrack(files.NewAlbum("", nil, topDir), goodTrackName, "", 2),
				},
			},
			WantedRecording: output.WantedRecording{
				Console: fmt.Sprintf("The track %q has been backed up to %q.\n", filepath.Join(topDir, goodTrackName), filepath.Join(files.CreateBackupPath(topDir), "1.mp3")),
				Log:     "level='error' command='repair' destination='backupTracks\\pre-repair-backup\\2.mp3' error='open backupTracks\\pre-repair-backup\\2.mp3: is a directory' source='backupTracks\\1 good track.mp3' msg='error copying file'\n",
				Error:   fmt.Sprintf("The track %q cannot be backed up.\n", filepath.Join(topDir, goodTrackName)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			backupTracks(o, tt.args.tracks)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func createTaggedContent(frames map[string]string) []byte {
	payload := make([]byte, 0)
	for k := 0; k < 256; k++ {
		payload = append(payload, byte(k))
	}
	content := files.CreateID3V2TaggedDataForTesting(payload, frames)
	return content
}

func Test_repair_fixTracks(t *testing.T) {
	fnName := "repair.fixTracks()"
	fFlag := false
	topDir := "fixTracks"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, topDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	frames := map[string]string{
		"TYER": "2022",
		"TALB": "unknown album",
		"TRCK": "2",
		"TCON": "dance music",
		"TCOM": "a couple of idiots",
		"TIT2": "unknown track",
		"TPE1": "unknown artist",
		"TLEN": "1000",
	}
	content := createTaggedContent(frames)
	trackName := "repairable track"
	goodFileName := "01 " + trackName + ".mp3"
	if err := internal.CreateFileForTestingWithContent(topDir, goodFileName, content); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, filepath.Join(topDir, goodFileName), err)
	}
	trackWithData := files.NewTrack(files.NewAlbum("ok album", files.NewArtist("beautiful singer", ""), topDir), goodFileName, trackName, 1)
	type args struct {
		tracks []*files.Track
	}
	tests := []struct {
		name string
		r    *repair
		args
		output.WantedRecording
	}{
		{name: "degenerate case", r: &repair{dryRun: &fFlag}, args: args{tracks: nil}},
		{
			name: "actual tracks",
			r:    &repair{dryRun: &fFlag},
			args: args{tracks: []*files.Track{
				files.NewTrack(
					files.NewAlbum("ok album", files.NewArtist("beautiful singer", ""), topDir),
					"non-existent-track", "", 0),
				trackWithData,
			}},
			WantedRecording: output.WantedRecording{
				Error: "An error occurred repairing track \"fixTracks\\\\non-existent-track\".\n" +
					"An error occurred repairing track \"fixTracks\\\\01 repairable track.mp3\".\n",
				Log: "level='error' command='repair' directory='fixTracks' error='[no edit required]' fileName='non-existent-track' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='fixTracks' error='[no edit required]' fileName='01 repairable track.mp3' msg='cannot edit track'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			fixTracks(o, tt.args.tracks)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}
