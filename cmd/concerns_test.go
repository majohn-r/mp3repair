/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd_test

import (
	"mp3repair/cmd"
	"mp3repair/internal/files"
	"testing"

	"github.com/majohn-r/output"
)

func TestConcernName(t *testing.T) {
	tests := map[string]struct {
		i    cmd.ConcernType
		want string
	}{
		"unspecified": {i: cmd.UnspecifiedConcern, want: "concern 0"},
		"empty":       {i: cmd.EmptyConcern, want: "empty"},
		"files":       {i: cmd.FilesConcern, want: "files"},
		"numbering":   {i: cmd.NumberingConcern, want: "numbering"},
		"metadata":    {i: cmd.ConflictConcern, want: "metadata conflict"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := cmd.ConcernName(tt.i); got != tt.want {
				t.Errorf("ConcernName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConcerns_AddConcern(t *testing.T) {
	type args struct {
		source  cmd.ConcernType
		concern string
	}
	tests := map[string]struct {
		c    cmd.Concerns
		args args
	}{
		"add empty concern": {
			c: cmd.NewConcerns(),
			args: args{
				source:  cmd.EmptyConcern,
				concern: "no albums",
			},
		},
		"add files concern": {
			c: cmd.NewConcerns(),
			args: args{
				source:  cmd.FilesConcern,
				concern: "genre mismatch",
			},
		},
		"add numbering concern": {
			c: cmd.NewConcerns(),
			args: args{
				source:  cmd.NumberingConcern,
				concern: "missing track 3",
			},
		},
		"add metadata conflict concern": {
			c: cmd.NewConcerns(),
			args: args{
				source:  cmd.ConflictConcern,
				concern: "id3v1 and id3v2 metadata disagree on the album name",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.c.IsConcerned() {
				t.Errorf("Concerns.AddConcern() has concerns from the start")
			}
			tt.c.AddConcern(tt.args.source, tt.args.concern)
			if !tt.c.IsConcerned() {
				t.Errorf("Concerns.AddConcern() did not add a concern")
			}
		})
	}
}

func TestNewConcernedTrack(t *testing.T) {
	tests := map[string]struct {
		track          *files.Track
		wantValidValue bool
	}{
		"nil":  {track: nil, wantValidValue: false},
		"real": {track: sampleTrack, wantValidValue: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := cmd.NewConcernedTrack(tt.track)
			if tt.wantValidValue {
				if got == nil {
					t.Errorf("NewConcernedTrack() = %v, want non-nil", got)
				} else {
					if got.IsConcerned() {
						t.Errorf("NewConcernedTrack() has concerns")
					}
					if got.Track() != tt.track {
						t.Errorf("NewConcernedTrack() has the wrong track")
					}
					got.AddConcern(cmd.FilesConcern, "no metadata")
					if !got.IsConcerned() {
						t.Errorf("NewConcernedTrack() does not reflect added concern")
					}
				}
			} else {
				if got != nil {
					t.Errorf("NewConcernedTrack() = %v, want nil", got)
				}
			}
		})
	}
}

func TestNewConcernedAlbum(t *testing.T) {
	var testAlbum *files.Album
	if albums := generateAlbums(1, 5); len(albums) > 0 {
		testAlbum = albums[0]
	}
	tests := map[string]struct {
		album          *files.Album
		wantValidAlbum bool
	}{
		"nil":  {album: nil, wantValidAlbum: false},
		"real": {album: testAlbum, wantValidAlbum: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := cmd.NewConcernedAlbum(tt.album)
			if tt.wantValidAlbum {
				if got == nil {
					t.Errorf("NewConcernedAlbum() = %v, want non-nil", got)
				} else {
					if got.IsConcerned() {
						t.Errorf("NewConcernedAlbum() created with concerns")
					}
					if got.Album() != tt.album {
						t.Errorf(
							"NewConcernedAlbum() created with wrong album: got %v, want %v",
							got.Album(), tt.album)
					}
					if len(got.Tracks()) != len(tt.album.Tracks) {
						t.Errorf("NewConcernedAlbum() created with %d tracks, want %d",
							len(got.Tracks()), len(tt.album.Tracks))
					}
					got.AddConcern(cmd.NumberingConcern, "missing track 1")
					if !got.IsConcerned() {
						t.Errorf("NewConcernedAlbum() cannot add concern")
					} else {
						got.Concerns = cmd.NewConcerns()
						if got.IsConcerned() {
							t.Errorf("NewConcernedAlbum() has concerns with clean map")
						}
						for _, track := range got.Tracks() {
							track.AddConcern(cmd.FilesConcern, "missing metadata")
							break
						}
						if !got.IsConcerned() {
							t.Errorf("NewConcernedAlbum() does not show concern assigned" +
								" to track")
						}
					}
				}
			} else {
				if got != nil {
					t.Errorf("NewConcernedAlbum() = %v, want nil", got)
				}
			}
		})
	}
}

func TestNewConcernedArtist(t *testing.T) {
	tests := map[string]struct {
		artist          *files.Artist
		wantValidArtist bool
	}{
		"nil":  {artist: nil, wantValidArtist: false},
		"real": {artist: generateArtists(1, 4, 5)[0], wantValidArtist: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := cmd.NewConcernedArtist(tt.artist)
			if tt.wantValidArtist {
				if got == nil {
					t.Errorf("NewConcernedArtist() = %v, want non-nil", got)
				} else {
					if got.IsConcerned() {
						t.Errorf("NewConcernedArtist() created with concerns")
					}
					if got.Artist() != tt.artist {
						t.Errorf("NewConcernedArtist() created with wrong artist:"+
							" got %v, want %v", got.Artist(), tt.artist)
					}
					if len(got.Albums()) != len(tt.artist.Albums) {
						t.Errorf("NewConcernedArtist() created with %d albums, want %d",
							len(got.Albums()), len(tt.artist.Albums))
					}
					got.AddConcern(cmd.EmptyConcern, "no albums!")
					if !got.IsConcerned() {
						t.Errorf("NewConcernedArtist()) cannot add concern")
					} else {
						got.Concerns = cmd.NewConcerns()
						if got.IsConcerned() {
							t.Errorf("NewConcernedArtist() has concerns with clean map")
						}
						for _, track := range got.Albums() {
							track.AddConcern(cmd.NumberingConcern, "missing track 909")
							break
						}
						if !got.IsConcerned() {
							t.Errorf("NewConcernedArtist() does not show concern" +
								" assigned to track")
						}
					}
				}
			} else {
				if got != nil {
					t.Errorf("NewConcernedArtist() = %v, want nil", got)
				}
			}
		})
	}
}

func TestConcerns_ToConsole(t *testing.T) {
	tests := map[string]struct {
		tab      int
		concerns map[cmd.ConcernType][]string
		output.WantedRecording
	}{
		"no concerns": {tab: 0, concerns: nil, WantedRecording: output.WantedRecording{}},
		"lots of concerns, untabbed": {
			tab: 0,
			concerns: map[cmd.ConcernType][]string{
				cmd.EmptyConcern:     {"no albums", "no tracks"},
				cmd.FilesConcern:     {"track 1 no data", "track 0 no data"},
				cmd.NumberingConcern: {"missing track 4", "missing track 1"},
			},
			WantedRecording: output.WantedRecording{
				Console: "" +
					"* [empty] no albums\n" +
					"* [empty] no tracks\n" +
					"* [files] track 0 no data\n" +
					"* [files] track 1 no data\n" +
					"* [numbering] missing track 1\n" +
					"* [numbering] missing track 4\n",
			},
		},
		"lots of concerns, indented": {
			tab: 2,
			concerns: map[cmd.ConcernType][]string{
				cmd.EmptyConcern:     {"no albums", "no tracks"},
				cmd.FilesConcern:     {"track 1 no data", "track 0 no data"},
				cmd.NumberingConcern: {"missing track 4", "missing track 1"},
			},
			WantedRecording: output.WantedRecording{
				Console: "" +
					"  * [empty] no albums\n" +
					"  * [empty] no tracks\n" +
					"  * [files] track 0 no data\n" +
					"  * [files] track 1 no data\n" +
					"  * [numbering] missing track 1\n" +
					"  * [numbering] missing track 4\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cI := cmd.NewConcerns()
			o := output.NewRecorder()
			o.IncrementTab(uint8(tt.tab))
			for k, v := range tt.concerns {
				for _, s := range v {
					cI.AddConcern(k, s)
				}
			}
			cI.ToConsole(o)
			o.Report(t, "Concerns.ToConsole()", tt.WantedRecording)
		})
	}
}

func TestConcernedTrack_ToConsole(t *testing.T) {
	tests := map[string]struct {
		cT       *cmd.ConcernedTrack
		concerns map[cmd.ConcernType][]string
		output.WantedRecording
	}{
		"no concerns": {
			cT:              cmd.NewConcernedTrack(sampleTrack),
			concerns:        nil,
			WantedRecording: output.WantedRecording{},
		},
		"some concerns": {
			cT: cmd.NewConcernedTrack(sampleTrack),
			concerns: map[cmd.ConcernType][]string{
				cmd.FilesConcern: {"missing ID3V1 metadata", "missing ID3V2 metadata"},
			},
			WantedRecording: output.WantedRecording{
				Console: "" +
					"    Track \"track 10\"\n" +
					"    * [files] missing ID3V1 metadata\n" +
					"    * [files] missing ID3V2 metadata\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			for k, v := range tt.concerns {
				for _, s := range v {
					tt.cT.AddConcern(k, s)
				}
			}
			o := output.NewRecorder()
			o.IncrementTab(4)
			tt.cT.ToConsole(o)
			o.Report(t, "ConcernedTrack.ToConsole()", tt.WantedRecording)
		})
	}
}

func TestConcernedAlbum_ToConsole(t *testing.T) {
	var album1 *files.Album
	if albums := generateAlbums(1, 1); len(albums) > 0 {
		album1 = albums[0]
	}
	albumWithConcerns := cmd.NewConcernedAlbum(album1)
	if albumWithConcerns != nil {
		albumWithConcerns.AddConcern(cmd.NumberingConcern, "missing track 2")
	}
	var album2 *files.Album
	if albums := generateAlbums(1, 4); len(albums) > 0 {
		album2 = albums[0]
	}
	albumWithTrackConcerns := cmd.NewConcernedAlbum(album2)
	if albumWithTrackConcerns != nil {
		albumWithTrackConcerns.Tracks()[3].AddConcern(cmd.FilesConcern,
			"no metadata detected")
	}
	var nilAlbum *files.Album
	if albums := generateAlbums(1, 2); len(albums) > 0 {
		nilAlbum = albums[0]
	}
	tests := map[string]struct {
		cAl *cmd.ConcernedAlbum
		output.WantedRecording
	}{
		"nil": {cAl: cmd.NewConcernedAlbum(nilAlbum)},
		"album with concerns itself": {
			cAl: albumWithConcerns,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"  Album \"my album 00\"\n" +
					"  * [numbering] missing track 2\n",
			},
		},
		"album with track concerns": {
			cAl: albumWithTrackConcerns,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"  Album \"my album 00\"\n" +
					"    Track \"my track 004\"\n" +
					"    * [files] no metadata detected\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			o.IncrementTab(2)
			tt.cAl.ToConsole(o)
			o.Report(t, "ConcernedAlbum.ToConsole()", tt.WantedRecording)
		})
	}
}

func TestConcernedArtist_ToConsole(t *testing.T) {
	// artist without concerns
	var artist1 *files.Artist
	if artists := generateArtists(1, 1, 1); len(artists) > 0 {
		artist1 = artists[0]
	}
	cAr000 := cmd.NewConcernedArtist(artist1)
	// artist with artist concerns
	var artist2 *files.Artist
	if artists := generateArtists(1, 0, 0); len(artists) > 0 {
		artist2 = artists[0]
	}
	cAr001 := cmd.NewConcernedArtist(artist2)
	if cAr001 != nil {
		cAr001.AddConcern(cmd.EmptyConcern, "no albums")
	}
	// artist with artist and album concerns
	var artist3 *files.Artist
	if artists := generateArtists(1, 1, 0); len(artists) > 0 {
		artist3 = artists[0]
	}
	cAr011 := cmd.NewConcernedArtist(artist3)
	if cAr011 != nil {
		cAr011.AddConcern(cmd.EmptyConcern, "expected no albums")
		cAr011.Albums()[0].AddConcern(cmd.EmptyConcern, "no tracks")
	}
	// artist with artist, album, and track concerns
	var artist4 *files.Artist
	if artists := generateArtists(1, 1, 1); len(artists) > 0 {
		artist4 = artists[0]
	}
	cAr111 := cmd.NewConcernedArtist(artist4)
	if cAr111 != nil {
		cAr111.AddConcern(cmd.EmptyConcern, "expected no albums")
		cAr111.Albums()[0].AddConcern(cmd.EmptyConcern, "expected no tracks")
		cAr111.Albums()[0].Tracks()[0].AddConcern(cmd.FilesConcern, "no metadata")
	}
	// artist with artist and track concerns
	var artist5 *files.Artist
	if artists := generateArtists(1, 1, 1); len(artists) > 0 {
		artist5 = artists[0]
	}
	cAr101 := cmd.NewConcernedArtist(artist5)
	if cAr101 != nil {
		cAr101.AddConcern(cmd.EmptyConcern, "expected no albums")
		cAr101.Albums()[0].Tracks()[0].AddConcern(cmd.FilesConcern, "no metadata")
	}
	// artist with album concerns
	var artist6 *files.Artist
	if artists := generateArtists(1, 1, 1); len(artists) > 0 {
		artist6 = artists[0]
	}
	cAr010 := cmd.NewConcernedArtist(artist6)
	if cAr010 != nil {
		cAr010.Albums()[0].AddConcern(cmd.EmptyConcern, "expected no tracks")
	}
	// artist with album and track concerns
	var artist7 *files.Artist
	if artists := generateArtists(1, 1, 1); len(artists) > 0 {
		artist7 = artists[0]
	}
	cAr110 := cmd.NewConcernedArtist(artist7)
	if cAr110 != nil {
		cAr110.Albums()[0].AddConcern(cmd.EmptyConcern, "expected no tracks")
		cAr110.Albums()[0].Tracks()[0].AddConcern(cmd.FilesConcern, "no metadata")
	}
	// artist with track concerns
	var artist8 *files.Artist
	if artists := generateArtists(1, 1, 1); len(artists) > 0 {
		artist8 = artists[0]
	}
	cAr100 := cmd.NewConcernedArtist(artist8)
	if cAr100 != nil {
		cAr100.Albums()[0].Tracks()[0].AddConcern(cmd.FilesConcern, "no metadata")
	}
	tests := map[string]struct {
		cAr *cmd.ConcernedArtist
		output.WantedRecording
	}{
		"nothing": {cAr: cAr000},
		"bad artist": {
			cAr: cAr001,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Artist \"my artist 0\"\n" +
					"* [empty] no albums\n",
			},
		},
		"bad artist, bad album": {
			cAr: cAr011,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Artist \"my artist 0\"\n" +
					"* [empty] expected no albums\n" +
					"  Album \"my album 00\"\n" +
					"  * [empty] no tracks\n",
			},
		},
		"bad artist, bad album, bad track": {
			cAr: cAr111,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Artist \"my artist 0\"\n" +
					"* [empty] expected no albums\n" +
					"  Album \"my album 00\"\n" +
					"  * [empty] expected no tracks\n" +
					"    Track \"my track 001\"\n" +
					"    * [files] no metadata\n",
			},
		},
		"bad artist, bad track": {
			cAr: cAr101,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Artist \"my artist 0\"\n" +
					"* [empty] expected no albums\n" +
					"  Album \"my album 00\"\n" +
					"    Track \"my track 001\"\n" +
					"    * [files] no metadata\n",
			},
		},
		"bad album": {
			cAr: cAr010,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Artist \"my artist 0\"\n" +
					"  Album \"my album 00\"\n" +
					"  * [empty] expected no tracks\n",
			},
		},
		"bad album, bad track": {
			cAr: cAr110,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Artist \"my artist 0\"\n" +
					"  Album \"my album 00\"\n" +
					"  * [empty] expected no tracks\n" +
					"    Track \"my track 001\"\n" +
					"    * [files] no metadata\n",
			},
		},
		"bad track": {
			cAr: cAr100,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Artist \"my artist 0\"\n" +
					"  Album \"my album 00\"\n" +
					"    Track \"my track 001\"\n" +
					"    * [files] no metadata\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.cAr.ToConsole(o)
			o.Report(t, "ConcernedArtist.ToConsole()", tt.WantedRecording)
		})
	}
}

func TestCreateConcernedArtists(t *testing.T) {
	tests := map[string]struct {
		artists    []*files.Artist
		want       int
		wantAlbums int
		wantTracks int
	}{
		"empty": {},
		"plenty": {
			artists:    generateArtists(15, 16, 17),
			want:       15,
			wantAlbums: 15 * 16,
			wantTracks: 15 * 16 * 17,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := cmd.CreateConcernedArtists(tt.artists); len(got) != tt.want {
				t.Errorf("CreateConcernedArtists() = %d, want %v", len(got), tt.want)
			} else {
				albums := 0
				tracks := 0
				collectedTracks := []*files.Track{}
				for _, artist := range got {
					albums += len(artist.Albums())
					for _, album := range artist.Albums() {
						tracks += len(album.Tracks())
						for _, cT := range album.Tracks() {
							collectedTracks = append(collectedTracks, cT.Track())
						}
					}
				}
				if albums != tt.wantAlbums {
					t.Errorf("CreateConcernedArtists() = %d albums, want %v", albums,
						tt.wantAlbums)
				}
				if tracks != tt.wantTracks {
					t.Errorf("PrepareConcernedArtists() = %d tracks, want %v", tracks,
						tt.wantTracks)
				}
				for _, track := range collectedTracks {
					found := false
					for _, cAr := range got {
						if cAr.Lookup(track) != nil {
							found = true
							break
						}
					}
					if !found {
						t.Errorf(
							"PrepareConcernedArtists() cannot find track %q on %q by %q",
							track.FileName(), track.AlbumName(), track.RecordingArtist())
					}
				}
				copiedTracks := []*files.Track{}
				for _, artist := range tt.artists {
					copiedAr := artist.Copy()
					for _, album := range artist.Albums {
						copiedAl := album.Copy(copiedAr, false)
						copiedAr.AddAlbum(copiedAl)
						for _, track := range album.Tracks {
							copiedTr := track.Copy(copiedAl)
							copiedAl.AddTrack(copiedTr)
							copiedTracks = append(copiedTracks, copiedTr)
						}
					}
				}
				for _, track := range copiedTracks {
					found := false
					for _, cAr := range got {
						if cAr.Lookup(track) != nil {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("PrepareConcernedArtists() cannot find copied track %q on"+
							" %q by %q", track.FileName(), track.AlbumName(),
							track.RecordingArtist())
					}
				}
			}
		})
	}
}

func TestConcernedArtist_Rollup(t *testing.T) {
	unconcernedArtist := cmd.NewConcernedArtist(files.NewArtist("artist name", "artist"))
	concernedArtist := cmd.NewConcernedArtist(files.NewArtist("artist name", "artist"))
	if concernedArtist != nil {
		concernedArtist.AddConcern(cmd.EmptyConcern, "no albums found")
	}
	artist1 := files.NewArtist("artist name", "artist")
	album1 := files.AlbumMaker{
		Title:  "album1",
		Artist: artist1,
		Path:   "album1",
	}.NewAlbum()
	album2 := files.AlbumMaker{
		Title:  "album2",
		Artist: artist1,
		Path:   "album2",
	}.NewAlbum()
	artist1.AddAlbum(album1)
	artist1.AddAlbum(album2)
	concernedArtistHeterogenousAlbums := cmd.NewConcernedArtist(artist1)
	if concernedArtistHeterogenousAlbums != nil {
		concernedArtistHeterogenousAlbums.Albums()[0].AddConcern(cmd.EmptyConcern, "no tracks found")
	}
	artist2 := files.NewArtist("artist name", "artist")
	album2a := files.AlbumMaker{
		Title:  "album1",
		Artist: artist2,
		Path:   "album1",
	}.NewAlbum()
	album2b := files.AlbumMaker{
		Title:  "album2",
		Artist: artist2,
		Path:   "album2",
	}.NewAlbum()
	artist2.AddAlbum(album2a)
	artist2.AddAlbum(album2b)
	concernedArtistIdenticalAlbums := cmd.NewConcernedArtist(artist2)
	if concernedArtistIdenticalAlbums != nil {
		for _, cAl := range concernedArtistIdenticalAlbums.Albums() {
			cAl.AddConcern(cmd.EmptyConcern, "no tracks")
		}
	}
	tests := map[string]struct {
		cAr  *cmd.ConcernedArtist
		want bool
	}{
		"unconcerned": {
			cAr:  unconcernedArtist,
			want: false,
		},
		"concerned, no albums": {
			cAr:  concernedArtist,
			want: false,
		},
		"concerned, mixed album concerns": {
			cAr:  concernedArtistHeterogenousAlbums,
			want: false,
		},
		"concerned, same album concerns": {
			cAr:  concernedArtistIdenticalAlbums,
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.cAr.Rollup(); got != tt.want {
				t.Errorf("ConcernedArtist.Rollup() got %t want %t", got, tt.want)
			}
			if tt.want {
				if !tt.cAr.Concerns.IsConcerned() {
					t.Errorf("ConcernedArtist.Rollup() after rollup has no concerns")
				}
				for _, cAl := range tt.cAr.Albums() {
					if cAl.Concerns.IsConcerned() {
						t.Errorf("ConcernedArtist.Rollup() after rollup, album has concerns")
					}
				}
			}
		})
	}
}

func TestConcernedAlbum_Rollup(t *testing.T) {
	albumNoTracks := cmd.NewConcernedAlbum(files.AlbumMaker{
		Title: "album",
		Path:  "album",
	}.NewAlbum())
	album1 := files.AlbumMaker{Title: "album1", Path: "album1"}.NewAlbum()
	album1.AddTrack(&files.Track{})
	album1.AddTrack(&files.Track{})
	albumWithTracksNoConcerns := cmd.NewConcernedAlbum(album1)
	album2 := files.AlbumMaker{Title: "album2", Path: "album2"}.NewAlbum()
	album2.AddTrack(&files.Track{})
	album2.AddTrack(&files.Track{})
	albumWithMixedConcerns := cmd.NewConcernedAlbum(album2)
	if albumWithMixedConcerns != nil {
		albumWithMixedConcerns.Tracks()[0].AddConcern(cmd.FilesConcern, "no metadata")
	}
	album3 := files.AlbumMaker{Title: "album3", Path: "album3"}.NewAlbum()
	album3.AddTrack(&files.Track{})
	album3.AddTrack(&files.Track{})
	albumWithIdenticalConcerns := cmd.NewConcernedAlbum(album3)
	if albumWithIdenticalConcerns != nil {
		for _, cT := range albumWithIdenticalConcerns.Tracks() {
			cT.AddConcern(cmd.FilesConcern, "no metadata")
		}
	}
	tests := map[string]struct {
		cAl  *cmd.ConcernedAlbum
		want bool
	}{
		"no tracks": {
			cAl:  albumNoTracks,
			want: false,
		},
		"tracks, no concerns": {
			cAl:  albumWithTracksNoConcerns,
			want: false,
		},
		"tracks, mixed concerns": {
			cAl:  albumWithMixedConcerns,
			want: false,
		},
		"tracks, same concerns": {
			cAl:  albumWithIdenticalConcerns,
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.cAl.Rollup(); got != tt.want {
				t.Errorf("ConcernedAlbum.Rollup() = %v, want %v", got, tt.want)
			}
			if tt.want {
				if !tt.cAl.Concerns.IsConcerned() {
					t.Errorf("ConcernedAlbum.Rollup() after rollup has no concerns")
				}
				for _, cT := range tt.cAl.Tracks() {
					if cT.Concerns.IsConcerned() {
						t.Errorf("ConcernedAlbum.Rollup() after rollup, track has concerns")
					}
				}
			}
		})
	}
}
