package commands

import (
	"bytes"
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

func Test_newRepairSubCommand(t *testing.T) {
	savedState := internal.SaveEnvVarForTesting("APPDATA")
	os.Setenv("APPDATA", internal.SecureAbsolutePathForTesting("."))
	defer func() {
		savedState.RestoreForTesting()
	}()
	topDir := "loadTest"
	fnName := "newRepairSubCommand()"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, topDir, err)
	}
	if err := internal.PopulateTopDirForTesting(topDir); err != nil {
		t.Errorf("%s error populating %s: %v", fnName, topDir, err)
	}
	if err := internal.CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("error creating defaults.yaml: %v", err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
		internal.DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	defaultConfig, _ := internal.ReadConfigurationFile(os.Stderr)
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
			repair := newRepairSubCommand(tt.args.c, flag.NewFlagSet("repair", flag.ContinueOnError))
			if _, ok := repair.sf.ProcessArgs(os.Stdout, []string{"-topDir", topDir, "-ext", ".mp3"}); ok {
				if *repair.dryRun != tt.wantDryRun {
					t.Errorf("%s %s: got dryRun %t want %t", fnName, tt.name, *repair.dryRun, tt.wantDryRun)
				}
			} else {
				t.Errorf("%s %s: error processing arguments", fnName, tt.name)
			}
		})
	}
}

func Test_findConflictedTracks(t *testing.T) {
	goodArtist := files.NewArtist("artist1", "")
	goodAlbum := files.NewAlbum("album1", goodArtist, "")
	goodArtist.AddAlbum(goodAlbum)
	goodTrack := files.NewTrack(goodAlbum, "", "track1", 1)
	goodTrack.SetTags(files.NewTaggedTrackData("album1", "artist1", "track1", "1"))
	goodAlbum.AddTrack(goodTrack)
	badArtist := files.NewArtist("artist1", "")
	badAlbum := files.NewAlbum("album1", badArtist, "")
	badArtist.AddAlbum(badAlbum)
	badTrack := files.NewTrack(badAlbum, "", "track1", 1)
	badTrack.SetTags(files.NewTaggedTrackData("album1", "artist1", "track3", "1"))
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
				t.Errorf("findConflictedTracks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_reportTracks(t *testing.T) {
	t1 := files.NewTrack(files.NewAlbum("album1", files.NewArtist("artist1", ""), ""), "", "track1", 1)
	t1.SetTags(files.NewTaggedTrackData("no album known", "no artist known", "no track name", "1"))
	t2 := files.NewTrack(files.NewAlbum("album1", files.NewArtist("artist1", ""), ""), "", "track2", 2)
	t2.SetTags(files.NewTaggedTrackData("no album known", "no artist known", "track2", "1"))
	t3 := files.NewTrack(files.NewAlbum("album2", files.NewArtist("artist1", ""), ""), "", "track1", 1)
	t3.SetTags(files.NewTaggedTrackData("no album known", "no artist known", "no track name", "1"))
	t4 := files.NewTrack(files.NewAlbum("album1", files.NewArtist("artist2", ""), ""), "", "track1", 1)
	t4.SetTags(files.NewTaggedTrackData("no album known", "no artist known", "no track name", "1"))
	type args struct {
		tracks []*files.Track
	}
	tests := []struct {
		name  string
		args  args
		wantW string
	}{
		{name: "no tracks", args: args{}},
		{name: "multiple tracks", args: args{tracks: []*files.Track{t1, t2, t3, t4}}, wantW: `"artist1"
    "album1"
         1 "track1" need to fix track name; album name; artist name;
         2 "track2" need to fix track numbering; album name; artist name;
    "album2"
         1 "track1" need to fix track name; album name; artist name;
"artist2"
    "album1"
         1 "track1" need to fix track name; album name; artist name;
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			reportTracks(w, tt.args.tracks)
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("reportTracks() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func Test_repair_Exec(t *testing.T) {
	fnName := "repair.Exec()"
	topDirName := "repairExec"
	topDirWithContent := "realContent"
	if err := internal.Mkdir(topDirName); err != nil {
		t.Errorf("%s error creating directory %q", fnName, topDirName)
	}
	if err := internal.Mkdir(topDirWithContent); err != nil {
		t.Errorf("%s error creating directory %q", fnName, topDirWithContent)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDirName)
		internal.DestroyDirectoryForTesting(fnName, topDirWithContent)
	}()
	if err := internal.PopulateTopDirForTesting(topDirName); err != nil {
		t.Errorf("%s error populating directory %q", fnName, topDirName)
	}
	artist := "new artist"
	if err := internal.Mkdir(filepath.Join(topDirWithContent, artist)); err != nil {
		t.Errorf("%s error creating directory %q", filepath.Join(topDirWithContent, artist), err)
	}
	album := "new album"
	if err := internal.Mkdir(filepath.Join(topDirWithContent, artist, album)); err != nil {
		t.Errorf("%s error creating directory %q", filepath.Join(topDirWithContent, artist, album), err)
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
	if err := internal.CreateFileForTestingWithContent(filepath.Join(topDirWithContent, artist, album), trackName, string(content)); err != nil {
		t.Errorf("%s error creating file %q", filepath.Join(topDirWithContent, artist, album, trackName), err)
	}
	type args struct {
		args []string
	}
	tests := []struct {
		name    string
		r       *repair
		args    args
		wantOut string
		wantErr string
		wantLog string
	}{
		{
			name:    "dry run, no usable content",
			r:       newRepairSubCommand(internal.EmptyConfiguration(), flag.NewFlagSet("repair", flag.ContinueOnError)),
			args:    args{[]string{"-topDir", topDirName, "-dryRun"}},
			wantOut: noProblemsFound + "\n",
		},
		{
			name:    "real repair, no usable content",
			r:       newRepairSubCommand(internal.EmptyConfiguration(), flag.NewFlagSet("repair", flag.ContinueOnError)),
			args:    args{[]string{"-topDir", topDirName, "-dryRun=false"}},
			wantOut: noProblemsFound + "\n",
		},
		{
			name: "dry run, usable content",
			r:    newRepairSubCommand(internal.EmptyConfiguration(), flag.NewFlagSet("repair", flag.ContinueOnError)),
			args: args{[]string{"-topDir", topDirWithContent, "-dryRun"}},
			wantOut: strings.Join([]string{
				"\"new artist\"",
				"    \"new album\"",
				"         1 \"new track\" need to fix track numbering; track name; album name; artist name;\n",
			}, "\n"),
		},
		{
			name: "real repair, usable content",
			r:    newRepairSubCommand(internal.EmptyConfiguration(), flag.NewFlagSet("repair", flag.ContinueOnError)),
			args: args{[]string{"-topDir", topDirWithContent, "-dryRun=false"}},
			wantOut: strings.Join([]string{
				"The track \"realContent\\\\new artist\\\\new album\\\\01 new track.mp3\" has been backed up to \"realContent\\\\new artist\\\\new album\\\\pre-repair-backup\\\\1.mp3\".",
				"\"realContent\\\\new artist\\\\new album\\\\01 new track.mp3\" fixed\n",
			}, "\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			tt.r.Exec(o, tt.args.args)
			if gotOut := o.Stdout(); gotOut != tt.wantOut {
				t.Errorf("%s console output = %v, want %v", fnName, gotOut, tt.wantOut)
			}
			if gotErr := o.Stderr(); gotErr != tt.wantErr {
				t.Errorf("%s error output = %v, want %v", fnName, gotErr, tt.wantErr)
			}
			if gotLog := o.LogOutput(); gotLog != tt.wantLog {
				t.Errorf("%s log output = %v, want %v", fnName, gotLog, tt.wantLog)
			}
		})
	}
}

func Test_getAlbumPaths(t *testing.T) {
	fnName := "getAlbumPaths()"
	topDir := "getAlbumPaths"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, topDir, err)
	}
	if err := internal.PopulateTopDirForTesting(topDir); err != nil {
		t.Errorf("%s error populating %s: %v", fnName, topDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	s := files.CreateFilteredSearchForTesting(topDir, "^.*$", "^.*$")
	a := s.LoadData(os.Stderr)
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
				t.Errorf("getAlbumPaths() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_repair_makeBackupDirectories(t *testing.T) {
	fnName := "repair.makeBackupDirectories()"
	topDir := "makeBackupDirectories"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, topDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	backupDir := files.CreateBackupPath(topDir)
	if err := internal.Mkdir(backupDir); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, backupDir, err)
	}
	albumDir := filepath.Join(topDir, "album")
	if err := internal.Mkdir(albumDir); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, albumDir, err)
	}
	if err := internal.CreateNamedFileForTesting(files.CreateBackupPath(albumDir), "nonsense content"); err != nil {
		t.Errorf("%s error creating file %s in %s: %v", fnName, files.CreateBackupPath(albumDir), albumDir, err)
	}
	albumDir2 := filepath.Join(topDir, "album2")
	if err := internal.Mkdir(albumDir2); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, albumDir2, err)
	}
	fFlag := false
	type args struct {
		paths []string
	}
	tests := []struct {
		name  string
		r     *repair
		args  args
		wantW string
	}{
		{name: "degenerate case", r: &repair{dryRun: &fFlag}, args: args{paths: nil}, wantW: ""},
		{
			name: "useful case",
			r:    &repair{dryRun: &fFlag},
			args: args{paths: []string{topDir, albumDir, albumDir2}},
			wantW: `The directory "makeBackupDirectories\\album\\pre-repair-backup" cannot be created: file exists and is not a directory.
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			tt.r.makeBackupDirectories(w, tt.args.paths)
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("%s = %v, want %v", fnName, gotW, tt.wantW)
			}
		})
	}
}

func Test_repair_backupTracks(t *testing.T) {
	fnName := "repair.backupTracks()"
	topDir := "backupTracks"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, topDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
	}()
	goodTrackName := "1 good track.mp3"
	if err := internal.CreateFileForTesting(topDir, goodTrackName); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, goodTrackName, err)
	}
	if err := internal.Mkdir(files.CreateBackupPath(topDir)); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, files.CreateBackupPath(topDir), err)
	}
	if err := internal.Mkdir(filepath.Join(files.CreateBackupPath(topDir), "2.mp3")); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, "2.mp3", err)
	}
	fFlag := false
	type args struct {
		tracks []*files.Track
	}
	tests := []struct {
		name  string
		r     *repair
		args  args
		wantW string
	}{
		{name: "degenerate case", r: &repair{dryRun: &fFlag}, args: args{tracks: nil}, wantW: ""},
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
			wantW: fmt.Sprintf("The track %q has been backed up to %q.\n", filepath.Join(topDir, goodTrackName), filepath.Join(files.CreateBackupPath(topDir), "1.mp3")) +
				fmt.Sprintf("The track %q cannot be backed up.\n", filepath.Join(topDir, goodTrackName)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			tt.r.backupTracks(w, tt.args.tracks)
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("%s = %v, want %v", fnName, gotW, tt.wantW)
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
	fFlag := false
	fnName := "repair.fixTracks()"
	topDir := "fixTracks"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, topDir, err)
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
	if err := internal.CreateFileForTestingWithContent(topDir, goodFileName, string(content)); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, filepath.Join(topDir, goodFileName), err)
	}
	trackWithData := files.NewTrack(files.NewAlbum("ok album", files.NewArtist("beautiful singer", ""), topDir), goodFileName, trackName, 1)
	trackWithData.SetTags(files.NewTaggedTrackData(frames["TALB"], frames["TPE1"], frames["TIT2"], frames["TRCK"]))
	type args struct {
		tracks []*files.Track
	}
	tests := []struct {
		name  string
		r     *repair
		args  args
		wantW string
	}{
		{name: "degenerate case", r: &repair{dryRun: &fFlag}, args: args{tracks: nil}, wantW: ""},
		{
			name: "actual tracks",
			r:    &repair{dryRun: &fFlag},
			args: args{tracks: []*files.Track{
				files.NewTrack(
					files.NewAlbum("ok album", files.NewArtist("beautiful singer", ""), topDir),
					"non-existent-track", "", 0),
				trackWithData,
			}},
			wantW: fmt.Sprintf("An error occurred fixing track %q\n",
				filepath.Join(topDir, "non-existent-track")) +
				fmt.Sprintf("%q fixed\n", filepath.Join(topDir, goodFileName)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			tt.r.fixTracks(w, tt.args.tracks)
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("%s = %v, want %v", fnName, gotW, tt.wantW)
			}
		})
	}
}
