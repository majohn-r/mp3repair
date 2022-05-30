package subcommands

import (
	"bytes"
	"flag"
	"mp3/internal"
	"mp3/internal/files"
	"os"
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

func Test_newRepairSubCommand(t *testing.T) {
	topDir := "loadTest"
	fnName := "newRepairSubCommand()"
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, topDir, err)
	}
	if err := internal.PopulateTopDirForTesting(topDir); err != nil {
		t.Errorf("%s error populating %s: %v", fnName, topDir, err)
	}
	if err := internal.CreateDefaultYamlFile(); err != nil {
		t.Errorf("error creating defaults.yaml: %v", err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDir)
		internal.DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	type args struct {
		v *viper.Viper
	}
	tests := []struct {
		name       string
		args       args
		wantDryRun bool
	}{
		{
			name:       "ordinary defaults",
			args:       args{v: nil},
			wantDryRun: false,
		},
		{
			name:       "overridden defaults",
			args:       args{v: internal.ReadDefaultsYaml("./mp3")},
			wantDryRun: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repair := newRepairSubCommand(tt.args.v, flag.NewFlagSet("ls", flag.ContinueOnError))
			if s := repair.sf.ProcessArgs(os.Stdout, []string{"-topDir", topDir, "-ext", ".mp3"}); s != nil {
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
			args: args{
				artists: []*files.Artist{
					{
						Name: "artist1",
						Albums: []*files.Album{
							{
								Name: "album1",
								Tracks: []*files.Track{
									{
										TrackNumber:  1,
										Name:         "track1",
										TaggedTrack:  1,
										TaggedTitle:  "track1",
										TaggedAlbum:  "album1",
										TaggedArtist: "artist1",
										ContainingAlbum: &files.Album{
											Name:            "album1",
											RecordingArtist: &files.Artist{Name: "artist1"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "problems",
			args: args{
				artists: []*files.Artist{
					{
						Name: "artist1",
						Albums: []*files.Album{
							{
								Name: "album1",
								Tracks: []*files.Track{
									{
										TrackNumber:  1,
										Name:         "track1",
										TaggedTrack:  1,
										TaggedTitle:  "track3",
										TaggedAlbum:  "album1",
										TaggedArtist: "artist1",
										ContainingAlbum: &files.Album{
											Name:            "album1",
											RecordingArtist: &files.Artist{Name: "artist1"},
										},
									},
								},
							},
						},
					},
				},
			},
			want: []*files.Track{
				{
					TrackNumber:  1,
					Name:         "track1",
					TaggedTrack:  1,
					TaggedTitle:  "track3",
					TaggedAlbum:  "album1",
					TaggedArtist: "artist1",
					ContainingAlbum: &files.Album{
						Name:            "album1",
						RecordingArtist: &files.Artist{Name: "artist1"},
					},
				},
			},
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
	type args struct {
		tracks []*files.Track
	}
	tests := []struct {
		name  string
		args  args
		wantW string
	}{
		{name: "no tracks", args: args{}, wantW: noProblemsFound + "\n"},
		{
			name: "multiple tracks",
			args: args{
				tracks: []*files.Track{
					{
						TrackNumber:  1,
						Name:         "track1",
						TaggedAlbum:  "no album known",
						TaggedArtist: "no artist known",
						TaggedTitle:  "no track name",
						TaggedTrack:  1,
						ContainingAlbum: &files.Album{
							Name: "album1",
							RecordingArtist: &files.Artist{
								Name: "artist1",
							},
						},
					},
					{
						TrackNumber:  2,
						Name:         "track2",
						TaggedAlbum:  "no album known",
						TaggedArtist: "no artist known",
						TaggedTitle:  "track2",
						TaggedTrack:  1,
						ContainingAlbum: &files.Album{
							Name: "album1",
							RecordingArtist: &files.Artist{
								Name: "artist1",
							},
						},
					},
					{
						TrackNumber:  1,
						Name:         "track1",
						TaggedAlbum:  "no album known",
						TaggedArtist: "no artist known",
						TaggedTitle:  "no track name",
						TaggedTrack:  1,
						ContainingAlbum: &files.Album{
							Name: "album2",
							RecordingArtist: &files.Artist{
								Name: "artist1",
							},
						},
					},
					{
						TrackNumber:  1,
						Name:         "track1",
						TaggedAlbum:  "no album known",
						TaggedArtist: "no artist known",
						TaggedTitle:  "no track name",
						TaggedTrack:  1,
						ContainingAlbum: &files.Album{
							Name: "album1",
							RecordingArtist: &files.Artist{
								Name: "artist2",
							},
						},
					},
				},
			},
			wantW: `"artist1"
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
	topDirName := "repairExec"
	if err := internal.Mkdir(topDirName); err != nil {
		t.Errorf("error creating directory %q", topDirName)
	}
	fnName := "repair.Exec()"
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDirName)
	}()
	if err := internal.PopulateTopDirForTesting(topDirName); err != nil {
		t.Errorf("error populating directory %q", topDirName)
	}
	type args struct {
		args []string
	}
	tests := []struct {
		name  string
		r     *repair
		args  args
		wantW string
	}{
		{
			name:  "dry run",
			r:     newRepairSubCommand(nil, flag.NewFlagSet("repair", flag.ContinueOnError)),
			args:  args{[]string{"-topDir", topDirName, "-dryRun"}},
			wantW: noProblemsFound + "\n",
		},
		// NEEDTEST: need test case for real repair
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			tt.r.Exec(w, tt.args.args)
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("%s = %v, want %v", fnName, gotW, tt.wantW)
			}
		})
	}
}
