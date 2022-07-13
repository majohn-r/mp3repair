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
	defaultConfig, _ := internal.ReadConfigurationFile(internal.NewOutputDeviceForTesting())
	type args struct {
		c *internal.Configuration
	}
	tests := []struct {
		name       string
		args       args
		wantDryRun bool
	}{
		{
			name:       "ordinary defaults",
			args:       args{c: internal.EmptyConfiguration()},
			wantDryRun: false,
		},
		{
			name:       "overridden defaults",
			args:       args{c: defaultConfig},
			wantDryRun: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repair := newRepairCommand(tt.args.c, flag.NewFlagSet("repair", flag.ContinueOnError))
			if _, ok := repair.sf.ProcessArgs(internal.NewOutputDeviceForTesting(), []string{
				"-topDir", topDir,
				"-ext", ".mp3",
			}); ok {
				if *repair.dryRun != tt.wantDryRun {
					t.Errorf("%s %q: got dryRun %t want %t", fnName, tt.name, *repair.dryRun, tt.wantDryRun)
				}
			} else {
				t.Errorf("%s %q: error processing arguments", fnName, tt.name)
			}
		})
	}
}

func Test_findConflictedTracks(t *testing.T) {
	fnName := "findConflictedTracks()"
	goodArtist := files.NewArtist("artist1", "")
	goodAlbum := files.NewAlbum("album1", goodArtist, "")
	goodArtist.AddAlbum(goodAlbum)
	goodTrack := files.NewTrack(goodAlbum, "", "track1", 1)
	goodTrack.SetTags(files.NewTaggedTrackData("album1", "artist1", "track1", 1, nil))
	goodAlbum.AddTrack(goodTrack)
	badArtist := files.NewArtist("artist1", "")
	badAlbum := files.NewAlbum("album1", badArtist, "")
	badArtist.AddAlbum(badAlbum)
	badTrack := files.NewTrack(badAlbum, "", "track1", 1)
	badTrack.SetTags(files.NewTaggedTrackData("album1", "artist1", "track3", 1, []byte{0, 1, 2}))
	badAlbum.AddTrack(badTrack)
	type args struct {
		artists []*files.Artist
	}
	tests := []struct {
		name string
		args args
		want []*files.Track
	}{
		{name: "degenerate case", args: args{}},
		{
			name: "no problems",
			args: args{artists: []*files.Artist{goodArtist}},
		},
		{
			name: "problems",
			args: args{artists: []*files.Artist{badArtist}},
			want: []*files.Track{badTrack},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findConflictedTracks(tt.args.artists); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_reportTracks(t *testing.T) {
	fnName := "reportTracks()"
	t1 := files.NewTrack(files.NewAlbum("album1", files.NewArtist("artist1", ""), ""), "", "track1", 1)
	t1.SetTags(files.NewTaggedTrackData("no album known", "no artist known", "no track name", 1, []byte{0, 1, 2}))
	t2 := files.NewTrack(files.NewAlbum("album1", files.NewArtist("artist1", ""), ""), "", "track2", 2)
	t2.SetTags(files.NewTaggedTrackData("no album known", "no artist known", "track2", 1, []byte{0, 1, 2}))
	t3 := files.NewTrack(files.NewAlbum("album2", files.NewArtist("artist1", ""), ""), "", "track1", 1)
	t3.SetTags(files.NewTaggedTrackData("no album known", "no artist known", "no track name", 1, []byte{0, 1, 2}))
	t4 := files.NewTrack(files.NewAlbum("album1", files.NewArtist("artist2", ""), ""), "", "track1", 1)
	t4.SetTags(files.NewTaggedTrackData("no album known", "no artist known", "no track name", 1, []byte{0, 1, 2}))
	type args struct {
		tracks []*files.Track
	}
	tests := []struct {
		name string
		args args
		internal.WantedOutput
	}{
		{name: "no tracks", args: args{}},
		{
			name: "multiple tracks",
			args: args{tracks: []*files.Track{t1, t2, t3, t4}},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: "\"artist1\"\n" +
					"    \"album1\"\n" +
					"         1 \"track1\" need to fix track name; album name; artist name;\n" +
					"         2 \"track2\" need to fix track numbering; album name; artist name;\n" +
					"    \"album2\"\n" +
					"         1 \"track1\" need to fix track name; album name; artist name;\n" +
					"\"artist2\"\n" +
					"    \"album1\"\n" +
					"         1 \"track1\" need to fix track name; album name; artist name;\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			reportTracks(o, tt.args.tracks)
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
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
	defer func() {
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
		args args
		internal.WantedOutput
	}{
		{
			name: "dry run, no usable content",
			r:    newRepairCommand(internal.EmptyConfiguration(), flag.NewFlagSet("repair", flag.ContinueOnError)),
			args: args{[]string{"-topDir", topDirName, "-dryRun"}},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: noProblemsFound + "\n",
				WantErrorOutput:   generateStandardTrackErrorReport(),
				WantLogOutput: "level='info' -dryRun='true' command='repair' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='repairExec' msg='reading filtered music files'\n" +
					generateStandardTrackLogReport(),
			},
		},
		{
			name: "real repair, no usable content",
			r:    newRepairCommand(internal.EmptyConfiguration(), flag.NewFlagSet("repair", flag.ContinueOnError)),
			args: args{[]string{"-topDir", topDirName, "-dryRun=false"}},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: noProblemsFound + "\n",
				WantErrorOutput:   generateStandardTrackErrorReport(),
				WantLogOutput: "level='info' -dryRun='false' command='repair' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='repairExec' msg='reading filtered music files'\n" +
					generateStandardTrackLogReport(),
			},
		},
		{
			name: "dry run, usable content",
			r:    newRepairCommand(internal.EmptyConfiguration(), flag.NewFlagSet("repair", flag.ContinueOnError)),
			args: args{[]string{"-topDir", topDirWithContent, "-dryRun"}},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: strings.Join([]string{
					"\"new artist\"",
					"    \"new album\"",
					"         1 \"new track\" need to fix track numbering; track name; album name; artist name;\n",
				}, "\n"),
				WantLogOutput: "level='info' -dryRun='true' command='repair' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='realContent' msg='reading filtered music files'\n",
			},
		},
		{
			name: "real repair, usable content",
			r:    newRepairCommand(internal.EmptyConfiguration(), flag.NewFlagSet("repair", flag.ContinueOnError)),
			args: args{[]string{"-topDir", topDirWithContent, "-dryRun=false"}},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: strings.Join([]string{
					"The track \"realContent\\\\new artist\\\\new album\\\\01 new track.mp3\" has been backed up to \"realContent\\\\new artist\\\\new album\\\\pre-repair-backup\\\\1.mp3\".",
					"\"realContent\\\\new artist\\\\new album\\\\01 new track.mp3\" fixed\n",
				}, "\n"),
				WantLogOutput: "level='info' -dryRun='false' command='repair' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='realContent' msg='reading filtered music files'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			tt.r.Exec(o, tt.args.args)
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
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
				result = append(result, fmt.Sprintf("An error occurred when trying to read tag information for track \"Test Track[%02d]\" on album \"Test Album %d\" by artist \"Test Artist %d\": \"zero length\"\n", track, album, artist))
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
				result = append(result, fmt.Sprintf("level='warn' albumName='Test Album %d' artistName='Test Artist %d' error='zero length' trackName='Test Track[%02d]' msg='tag error'\n", album, artist, track))
			}
		}
	}
	return strings.Join(result, "")
}

func Test_getAlbumPaths(t *testing.T) {
	fnName := "getAlbumPaths()"
	topDir := "getAlbumPaths"
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
	a, _ := s.LoadData(internal.NewOutputDeviceForTesting())
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
		args args
		want []string
	}{
		{name: "degenerate case", args: args{}},
		{name: "full blown case", args: args{tracks: tSlice}, want: []string{
			"getAlbumPaths\\Test Artist 0\\Test Album 0",
			"getAlbumPaths\\Test Artist 0\\Test Album 1",
			"getAlbumPaths\\Test Artist 0\\Test Album 2",
			"getAlbumPaths\\Test Artist 0\\Test Album 3",
			"getAlbumPaths\\Test Artist 0\\Test Album 4",
			"getAlbumPaths\\Test Artist 0\\Test Album 5",
			"getAlbumPaths\\Test Artist 0\\Test Album 6",
			"getAlbumPaths\\Test Artist 0\\Test Album 7",
			"getAlbumPaths\\Test Artist 0\\Test Album 8",
			"getAlbumPaths\\Test Artist 0\\Test Album 9",
			"getAlbumPaths\\Test Artist 1\\Test Album 0",
			"getAlbumPaths\\Test Artist 1\\Test Album 1",
			"getAlbumPaths\\Test Artist 1\\Test Album 2",
			"getAlbumPaths\\Test Artist 1\\Test Album 3",
			"getAlbumPaths\\Test Artist 1\\Test Album 4",
			"getAlbumPaths\\Test Artist 1\\Test Album 5",
			"getAlbumPaths\\Test Artist 1\\Test Album 6",
			"getAlbumPaths\\Test Artist 1\\Test Album 7",
			"getAlbumPaths\\Test Artist 1\\Test Album 8",
			"getAlbumPaths\\Test Artist 1\\Test Album 9",
			"getAlbumPaths\\Test Artist 2\\Test Album 0",
			"getAlbumPaths\\Test Artist 2\\Test Album 1",
			"getAlbumPaths\\Test Artist 2\\Test Album 2",
			"getAlbumPaths\\Test Artist 2\\Test Album 3",
			"getAlbumPaths\\Test Artist 2\\Test Album 4",
			"getAlbumPaths\\Test Artist 2\\Test Album 5",
			"getAlbumPaths\\Test Artist 2\\Test Album 6",
			"getAlbumPaths\\Test Artist 2\\Test Album 7",
			"getAlbumPaths\\Test Artist 2\\Test Album 8",
			"getAlbumPaths\\Test Artist 2\\Test Album 9",
			"getAlbumPaths\\Test Artist 3\\Test Album 0",
			"getAlbumPaths\\Test Artist 3\\Test Album 1",
			"getAlbumPaths\\Test Artist 3\\Test Album 2",
			"getAlbumPaths\\Test Artist 3\\Test Album 3",
			"getAlbumPaths\\Test Artist 3\\Test Album 4",
			"getAlbumPaths\\Test Artist 3\\Test Album 5",
			"getAlbumPaths\\Test Artist 3\\Test Album 6",
			"getAlbumPaths\\Test Artist 3\\Test Album 7",
			"getAlbumPaths\\Test Artist 3\\Test Album 8",
			"getAlbumPaths\\Test Artist 3\\Test Album 9",
			"getAlbumPaths\\Test Artist 4\\Test Album 0",
			"getAlbumPaths\\Test Artist 4\\Test Album 1",
			"getAlbumPaths\\Test Artist 4\\Test Album 2",
			"getAlbumPaths\\Test Artist 4\\Test Album 3",
			"getAlbumPaths\\Test Artist 4\\Test Album 4",
			"getAlbumPaths\\Test Artist 4\\Test Album 5",
			"getAlbumPaths\\Test Artist 4\\Test Album 6",
			"getAlbumPaths\\Test Artist 4\\Test Album 7",
			"getAlbumPaths\\Test Artist 4\\Test Album 8",
			"getAlbumPaths\\Test Artist 4\\Test Album 9",
			"getAlbumPaths\\Test Artist 5\\Test Album 0",
			"getAlbumPaths\\Test Artist 5\\Test Album 1",
			"getAlbumPaths\\Test Artist 5\\Test Album 2",
			"getAlbumPaths\\Test Artist 5\\Test Album 3",
			"getAlbumPaths\\Test Artist 5\\Test Album 4",
			"getAlbumPaths\\Test Artist 5\\Test Album 5",
			"getAlbumPaths\\Test Artist 5\\Test Album 6",
			"getAlbumPaths\\Test Artist 5\\Test Album 7",
			"getAlbumPaths\\Test Artist 5\\Test Album 8",
			"getAlbumPaths\\Test Artist 5\\Test Album 9",
			"getAlbumPaths\\Test Artist 6\\Test Album 0",
			"getAlbumPaths\\Test Artist 6\\Test Album 1",
			"getAlbumPaths\\Test Artist 6\\Test Album 2",
			"getAlbumPaths\\Test Artist 6\\Test Album 3",
			"getAlbumPaths\\Test Artist 6\\Test Album 4",
			"getAlbumPaths\\Test Artist 6\\Test Album 5",
			"getAlbumPaths\\Test Artist 6\\Test Album 6",
			"getAlbumPaths\\Test Artist 6\\Test Album 7",
			"getAlbumPaths\\Test Artist 6\\Test Album 8",
			"getAlbumPaths\\Test Artist 6\\Test Album 9",
			"getAlbumPaths\\Test Artist 7\\Test Album 0",
			"getAlbumPaths\\Test Artist 7\\Test Album 1",
			"getAlbumPaths\\Test Artist 7\\Test Album 2",
			"getAlbumPaths\\Test Artist 7\\Test Album 3",
			"getAlbumPaths\\Test Artist 7\\Test Album 4",
			"getAlbumPaths\\Test Artist 7\\Test Album 5",
			"getAlbumPaths\\Test Artist 7\\Test Album 6",
			"getAlbumPaths\\Test Artist 7\\Test Album 7",
			"getAlbumPaths\\Test Artist 7\\Test Album 8",
			"getAlbumPaths\\Test Artist 7\\Test Album 9",
			"getAlbumPaths\\Test Artist 8\\Test Album 0",
			"getAlbumPaths\\Test Artist 8\\Test Album 1",
			"getAlbumPaths\\Test Artist 8\\Test Album 2",
			"getAlbumPaths\\Test Artist 8\\Test Album 3",
			"getAlbumPaths\\Test Artist 8\\Test Album 4",
			"getAlbumPaths\\Test Artist 8\\Test Album 5",
			"getAlbumPaths\\Test Artist 8\\Test Album 6",
			"getAlbumPaths\\Test Artist 8\\Test Album 7",
			"getAlbumPaths\\Test Artist 8\\Test Album 8",
			"getAlbumPaths\\Test Artist 8\\Test Album 9",
			"getAlbumPaths\\Test Artist 9\\Test Album 0",
			"getAlbumPaths\\Test Artist 9\\Test Album 1",
			"getAlbumPaths\\Test Artist 9\\Test Album 2",
			"getAlbumPaths\\Test Artist 9\\Test Album 3",
			"getAlbumPaths\\Test Artist 9\\Test Album 4",
			"getAlbumPaths\\Test Artist 9\\Test Album 5",
			"getAlbumPaths\\Test Artist 9\\Test Album 6",
			"getAlbumPaths\\Test Artist 9\\Test Album 7",
			"getAlbumPaths\\Test Artist 9\\Test Album 8",
			"getAlbumPaths\\Test Artist 9\\Test Album 9",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAlbumPaths(tt.args.tracks); !reflect.DeepEqual(got, tt.want) {
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
		args args
		internal.WantedOutput
	}{
		{name: "degenerate case", r: &repair{dryRun: &fFlag}, args: args{paths: nil}},
		{
			name: "useful case",
			r:    &repair{dryRun: &fFlag},
			args: args{paths: []string{topDir, albumDir, albumDir2}},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The directory \"makeBackupDirectories\\\\album\\\\pre-repair-backup\" cannot be created: file exists and is not a directory.\n",
				WantLogOutput:   "level='warn' command='' directory='makeBackupDirectories\\album\\pre-repair-backup' error='file exists and is not a directory' msg='cannot create directory'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			tt.r.makeBackupDirectories(o, tt.args.paths)
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
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
		args args
		internal.WantedOutput
	}{
		{name: "degenerate case", r: &repair{dryRun: &fFlag}, args: args{tracks: nil}},
		{
			name: "real tests",
			r:    &repair{dryRun: &fFlag, n: "repair"},
			args: args{
				tracks: []*files.Track{
					files.NewTrack(files.NewAlbum("", nil, topDir), goodTrackName, "", 1),
					files.NewTrack(files.NewAlbum("", nil, topDir), "dup track", "", 1),
					files.NewTrack(files.NewAlbum("", nil, topDir), goodTrackName, "", 2),
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: fmt.Sprintf("The track %q has been backed up to %q.\n", filepath.Join(topDir, goodTrackName), filepath.Join(files.CreateBackupPath(topDir), "1.mp3")),
				WantLogOutput:     "level='warn' command='repair' destination='backupTracks\\pre-repair-backup\\2.mp3' error='open backupTracks\\pre-repair-backup\\2.mp3: is a directory' source='backupTracks\\1 good track.mp3' msg='error copying file'\n",
				WantErrorOutput:   fmt.Sprintf("The track %q cannot be backed up.\n", filepath.Join(topDir, goodTrackName)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			tt.r.backupTracks(o, tt.args.tracks)
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
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
	content := files.CreateTaggedDataForTesting(payload, frames)
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
	trackWithData.SetTags(files.NewTaggedTrackData(frames["TALB"], frames["TPE1"], frames["TIT2"], 2, []byte{0, 1, 2}))
	type args struct {
		tracks []*files.Track
	}
	tests := []struct {
		name string
		r    *repair
		args args
		internal.WantedOutput
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
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: fmt.Sprintf("%q fixed\n", filepath.Join(topDir, goodFileName)),
				WantErrorOutput:   fmt.Sprintf("An error occurred fixing track %q\n", filepath.Join(topDir, "non-existent-track")),
				WantLogOutput:     "level='warn' directory='fixTracks' error='no edit required' executing command='' fileName='non-existent-track' msg='cannot edit track'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			tt.r.fixTracks(o, tt.args.tracks)
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}
