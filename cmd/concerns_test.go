/*
Copyright © 2026 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd

import (
	"mp3repair/internal/files"
	"testing"

	"github.com/majohn-r/output"
)

func Test_concernName(t *testing.T) {
	tests := map[string]struct {
		i    concernType
		want string
	}{
		"unspecified": {i: unspecifiedConcern, want: "concern 0"},
		"empty":       {i: emptyConcern, want: "empty"},
		"files":       {i: filesConcern, want: "files"},
		"numbering":   {i: numberingConcern, want: "numbering"},
		"metadata":    {i: conflictConcern, want: "metadata conflict"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := concernName(tt.i); got != tt.want {
				t.Errorf("concernName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_concerns_addConcern(t *testing.T) {
	type args struct {
		source  concernType
		concern string
	}
	tests := map[string]struct {
		c    concerns
		args args
	}{
		"add empty concern": {
			c: newConcerns(),
			args: args{
				source:  emptyConcern,
				concern: "no albums",
			},
		},
		"add files concern": {
			c: newConcerns(),
			args: args{
				source:  filesConcern,
				concern: "genre mismatch",
			},
		},
		"add numbering concern": {
			c: newConcerns(),
			args: args{
				source:  numberingConcern,
				concern: "missing track 3",
			},
		},
		"add metadata conflict concern": {
			c: newConcerns(),
			args: args{
				source:  conflictConcern,
				concern: "id3v1 and id3v2 metadata disagree on the album name",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.c.isConcerned() {
				t.Errorf("concerns.addConcern() has concerns from the start")
			}
			tt.c.addConcern(tt.args.source, tt.args.concern)
			if !tt.c.isConcerned() {
				t.Errorf("concerns.addConcern() did not add a concern")
			}
		})
	}
}

func Test_newConcernedTrack(t *testing.T) {
	tests := map[string]struct {
		track          *files.Track
		wantValidValue bool
	}{
		"nil":  {track: nil, wantValidValue: false},
		"real": {track: sampleTrack, wantValidValue: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := newConcernedTrack(tt.track)
			if tt.wantValidValue {
				if got == nil {
					t.Errorf("newConcernedTrack() = %v, want non-nil", got)
				} else {
					if got.isConcerned() {
						t.Errorf("newConcernedTrack() has concerns")
					}
					if got.backingTrack() != tt.track {
						t.Errorf("newConcernedTrack() has the wrong track")
					}
					got.addConcern(filesConcern, "no metadata")
					if !got.isConcerned() {
						t.Errorf("newConcernedTrack() does not reflect added concern")
					}
				}
			} else {
				if got != nil {
					t.Errorf("newConcernedTrack() = %v, want nil", got)
				}
			}
		})
	}
}

func Test_newConcernedAlbum(t *testing.T) {
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
			got := newConcernedAlbum(tt.album)
			if tt.wantValidAlbum {
				if got == nil {
					t.Errorf("newConcernedAlbum() = %v, want non-nil", got)
				} else {
					if got.isConcerned() {
						t.Errorf("newConcernedAlbum() created with concerns")
					}
					if got.backingAlbum() != tt.album {
						t.Errorf("newConcernedAlbum() created with wrong album: got %v, want %v",
							got.backingAlbum(), tt.album)
					}
					if len(got.tracks()) != len(tt.album.Tracks()) {
						t.Errorf("newConcernedAlbum() created with %d tracks, want %d",
							len(got.tracks()), len(tt.album.Tracks()))
					}
					got.addConcern(numberingConcern, "missing track 1")
					if !got.isConcerned() {
						t.Errorf("newConcernedAlbum() cannot add concern")
					} else {
						got.concerns = newConcerns()
						if got.isConcerned() {
							t.Errorf("newConcernedAlbum() has concerns with clean map")
						}
						for _, track := range got.tracks() {
							track.addConcern(filesConcern, "missing metadata")
							break
						}
						if !got.isConcerned() {
							t.Errorf("newConcernedAlbum() does not show concern assigned" +
								" to track")
						}
					}
				}
			} else {
				if got != nil {
					t.Errorf("newConcernedAlbum() = %v, want nil", got)
				}
			}
		})
	}
}

func Test_newConcernedArtist(t *testing.T) {
	tests := map[string]struct {
		artist          *files.Artist
		wantValidArtist bool
	}{
		"nil":  {artist: nil, wantValidArtist: false},
		"real": {artist: generateArtists(1, 4, 5, nil)[0], wantValidArtist: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := newConcernedArtist(tt.artist)
			if tt.wantValidArtist {
				if got == nil {
					t.Errorf("newConcernedArtist() = %v, want non-nil", got)
				} else {
					if got.isConcerned() {
						t.Errorf("newConcernedArtist() created with concerns")
					}
					if got.backingArtist() != tt.artist {
						t.Errorf("newConcernedArtist() created with wrong artist:"+
							" got %v, want %v", got.backingArtist(), tt.artist)
					}
					if len(got.albums()) != len(tt.artist.Albums()) {
						t.Errorf("newConcernedArtist() created with %d albums, want %d",
							len(got.albums()), len(tt.artist.Albums()))
					}
					got.addConcern(emptyConcern, "no albums!")
					if !got.isConcerned() {
						t.Errorf("newConcernedArtist()) cannot add concern")
					} else {
						got.concerns = newConcerns()
						if got.isConcerned() {
							t.Errorf("newConcernedArtist() has concerns with clean map")
						}
						for _, track := range got.albums() {
							track.addConcern(numberingConcern, "missing track 909")
							break
						}
						if !got.isConcerned() {
							t.Errorf("newConcernedArtist() does not show concern" +
								" assigned to track")
						}
					}
				}
			} else {
				if got != nil {
					t.Errorf("newConcernedArtist() = %v, want nil", got)
				}
			}
		})
	}
}

func Test_concerns_toConsole(t *testing.T) {
	tests := map[string]struct {
		tab      int
		concerns map[concernType][]string
		output.WantedRecording
	}{
		"no concerns": {tab: 0, concerns: nil, WantedRecording: output.WantedRecording{}},
		"lots of concerns, no indent": {
			tab: 0,
			concerns: map[concernType][]string{
				emptyConcern:     {"no albums", "no tracks"},
				filesConcern:     {"track 1 no data", "track 0 no data"},
				numberingConcern: {"missing track 4", "missing track 1"},
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
			concerns: map[concernType][]string{
				emptyConcern:     {"no albums", "no tracks"},
				filesConcern:     {"track 1 no data", "track 0 no data"},
				numberingConcern: {"missing track 4", "missing track 1"},
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
			cI := newConcerns()
			o := output.NewRecorder()
			o.IncrementTab(uint8(tt.tab))
			for k, v := range tt.concerns {
				for _, s := range v {
					cI.addConcern(k, s)
				}
			}
			cI.toConsole(o)
			o.Report(t, "concerns.toConsole()", tt.WantedRecording)
		})
	}
}

func Test_concernedTrack_toConsole(t *testing.T) {
	tests := map[string]struct {
		cT       *concernedTrack
		concerns map[concernType][]string
		output.WantedRecording
	}{
		"no concerns": {
			cT:              newConcernedTrack(sampleTrack),
			concerns:        nil,
			WantedRecording: output.WantedRecording{},
		},
		"some concerns": {
			cT: newConcernedTrack(sampleTrack),
			concerns: map[concernType][]string{
				filesConcern: {"missing ID3V1 metadata", "missing ID3V2 metadata"},
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
					tt.cT.addConcern(k, s)
				}
			}
			o := output.NewRecorder()
			o.IncrementTab(4)
			tt.cT.toConsole(o)
			o.Report(t, "concernedTrack.toConsole()", tt.WantedRecording)
		})
	}
}

func Test_concernedAlbum_toConsole(t *testing.T) {
	var album1 *files.Album
	if albums := generateAlbums(1, 1); len(albums) > 0 {
		album1 = albums[0]
	}
	albumWithConcerns := newConcernedAlbum(album1)
	if albumWithConcerns != nil {
		albumWithConcerns.addConcern(numberingConcern, "missing track 2")
	}
	var album2 *files.Album
	if albums := generateAlbums(1, 4); len(albums) > 0 {
		album2 = albums[0]
	}
	albumWithTrackConcerns := newConcernedAlbum(album2)
	if albumWithTrackConcerns != nil {
		albumWithTrackConcerns.tracks()[3].addConcern(filesConcern,
			"no metadata detected")
	}
	var nilAlbum *files.Album
	if albums := generateAlbums(1, 2); len(albums) > 0 {
		nilAlbum = albums[0]
	}
	tests := map[string]struct {
		cAl *concernedAlbum
		output.WantedRecording
	}{
		"nil": {cAl: newConcernedAlbum(nilAlbum)},
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
			tt.cAl.toConsole(o)
			o.Report(t, "concernedAlbum.toConsole()", tt.WantedRecording)
		})
	}
}

func Test_concernedArtist_toConsole(t *testing.T) {
	// artist without concerns
	var artist1 *files.Artist
	if artists := generateArtists(1, 1, 1, nil); len(artists) > 0 {
		artist1 = artists[0]
	}
	cAr000 := newConcernedArtist(artist1)
	// artist with artist concerns
	var artist2 *files.Artist
	if artists := generateArtists(1, 0, 0, nil); len(artists) > 0 {
		artist2 = artists[0]
	}
	cAr001 := newConcernedArtist(artist2)
	if cAr001 != nil {
		cAr001.addConcern(emptyConcern, "no albums")
	}
	// artist with artist and album concerns
	var artist3 *files.Artist
	if artists := generateArtists(1, 1, 0, nil); len(artists) > 0 {
		artist3 = artists[0]
	}
	cAr011 := newConcernedArtist(artist3)
	if cAr011 != nil {
		cAr011.addConcern(emptyConcern, "expected no albums")
		cAr011.albums()[0].addConcern(emptyConcern, "no tracks")
	}
	// artist with artist, album, and track concerns
	var artist4 *files.Artist
	if artists := generateArtists(1, 1, 1, nil); len(artists) > 0 {
		artist4 = artists[0]
	}
	cAr111 := newConcernedArtist(artist4)
	if cAr111 != nil {
		cAr111.addConcern(emptyConcern, "expected no albums")
		cAr111.albums()[0].addConcern(emptyConcern, "expected no tracks")
		cAr111.albums()[0].tracks()[0].addConcern(filesConcern, "no metadata")
	}
	// artist with artist and track concerns
	var artist5 *files.Artist
	if artists := generateArtists(1, 1, 1, nil); len(artists) > 0 {
		artist5 = artists[0]
	}
	cAr101 := newConcernedArtist(artist5)
	if cAr101 != nil {
		cAr101.addConcern(emptyConcern, "expected no albums")
		cAr101.albums()[0].tracks()[0].addConcern(filesConcern, "no metadata")
	}
	// artist with album concerns
	var artist6 *files.Artist
	if artists := generateArtists(1, 1, 1, nil); len(artists) > 0 {
		artist6 = artists[0]
	}
	cAr010 := newConcernedArtist(artist6)
	if cAr010 != nil {
		cAr010.albums()[0].addConcern(emptyConcern, "expected no tracks")
	}
	// artist with album and track concerns
	var artist7 *files.Artist
	if artists := generateArtists(1, 1, 1, nil); len(artists) > 0 {
		artist7 = artists[0]
	}
	cAr110 := newConcernedArtist(artist7)
	if cAr110 != nil {
		cAr110.albums()[0].addConcern(emptyConcern, "expected no tracks")
		cAr110.albums()[0].tracks()[0].addConcern(filesConcern, "no metadata")
	}
	// artist with track concerns
	var artist8 *files.Artist
	if artists := generateArtists(1, 1, 1, nil); len(artists) > 0 {
		artist8 = artists[0]
	}
	cAr100 := newConcernedArtist(artist8)
	if cAr100 != nil {
		cAr100.albums()[0].tracks()[0].addConcern(filesConcern, "no metadata")
	}
	tests := map[string]struct {
		cAr *concernedArtist
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
			tt.cAr.toConsole(o)
			o.Report(t, "concernedArtist.toConsole()", tt.WantedRecording)
		})
	}
}

func Test_createConcernedArtists(t *testing.T) {
	tests := map[string]struct {
		artists    []*files.Artist
		want       int
		wantAlbums int
		wantTracks int
	}{
		"empty": {},
		"plenty": {
			artists:    generateArtists(15, 16, 17, nil),
			want:       15,
			wantAlbums: 15 * 16,
			wantTracks: 15 * 16 * 17,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := createConcernedArtists(tt.artists); len(got) != tt.want {
				t.Errorf("createConcernedArtists() = %d, want %v", len(got), tt.want)
			} else {
				albums := 0
				tracks := 0
				var collectedTracks []*files.Track
				for _, artist := range got {
					albums += len(artist.albums())
					for _, album := range artist.albums() {
						tracks += len(album.tracks())
						for _, cT := range album.tracks() {
							collectedTracks = append(collectedTracks, cT.backingTrack())
						}
					}
				}
				if albums != tt.wantAlbums {
					t.Errorf("createConcernedArtists() = %d albums, want %v", albums,
						tt.wantAlbums)
				}
				if tracks != tt.wantTracks {
					t.Errorf("createConcernedArtists() = %d tracks, want %v", tracks,
						tt.wantTracks)
				}
				for _, track := range collectedTracks {
					found := false
					for _, cAr := range got {
						if cAr.lookup(track) != nil {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("createConcernedArtists() cannot find track %q on %q by %q",
							track.FileName(), track.AlbumName(), track.RecordingArtist())
					}
				}
				var copiedTracks []*files.Track
				for _, artist := range tt.artists {
					copiedAr := artist.Copy()
					for _, album := range artist.Albums() {
						copiedAl := album.Copy(copiedAr, false, true)
						for _, track := range album.Tracks() {
							copiedTr := track.Copy(copiedAl, true)
							copiedTracks = append(copiedTracks, copiedTr)
						}
					}
				}
				for _, track := range copiedTracks {
					found := false
					for _, cAr := range got {
						if cAr.lookup(track) != nil {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("createConcernedArtists() cannot find copied track %q on"+
							" %q by %q", track.FileName(), track.AlbumName(),
							track.RecordingArtist())
					}
				}
			}
		})
	}
}

func Test_concernedArtist_rollup(t *testing.T) {
	unconcernedArtist := newConcernedArtist(files.NewArtist("artist name", "artist"))
	concernedArtist1 := newConcernedArtist(files.NewArtist("artist name", "artist"))
	if concernedArtist1 != nil {
		concernedArtist1.addConcern(emptyConcern, "no albums found")
	}
	artist1 := files.NewArtist("artist name", "artist")
	files.AlbumMaker{
		Title:     "album1",
		Artist:    artist1,
		Directory: "album1",
	}.NewAlbum(true)
	files.AlbumMaker{
		Title:     "album2",
		Artist:    artist1,
		Directory: "album2",
	}.NewAlbum(true)
	concernedArtistMixedAlbums := newConcernedArtist(artist1)
	if concernedArtistMixedAlbums != nil {
		concernedArtistMixedAlbums.albums()[0].addConcern(emptyConcern, "no tracks found")
	}
	artist2 := files.NewArtist("artist name", "artist")
	files.AlbumMaker{
		Title:     "album1",
		Artist:    artist2,
		Directory: "album1",
	}.NewAlbum(true)
	files.AlbumMaker{
		Title:     "album2",
		Artist:    artist2,
		Directory: "album2",
	}.NewAlbum(true)
	concernedArtistIdenticalAlbums := newConcernedArtist(artist2)
	if concernedArtistIdenticalAlbums != nil {
		for _, cAl := range concernedArtistIdenticalAlbums.albums() {
			cAl.addConcern(emptyConcern, "no tracks")
		}
	}
	tests := map[string]struct {
		cAr  *concernedArtist
		want bool
	}{
		"unconcerned": {
			cAr:  unconcernedArtist,
			want: false,
		},
		"concerned, no albums": {
			cAr:  concernedArtist1,
			want: false,
		},
		"concerned, mixed album concerns": {
			cAr:  concernedArtistMixedAlbums,
			want: false,
		},
		"concerned, same album concerns": {
			cAr:  concernedArtistIdenticalAlbums,
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.cAr.rollup(); got != tt.want {
				t.Errorf("concernedArtist.rollup() got %t want %t", got, tt.want)
			}
			if tt.want {
				if !tt.cAr.concerns.isConcerned() {
					t.Errorf("concernedArtist.rollup() after rollup has no concerns")
				}
				for _, cAl := range tt.cAr.albums() {
					if cAl.concerns.isConcerned() {
						t.Errorf("concernedArtist.rollup() after rollup, album has concerns")
					}
				}
			}
		})
	}
}

func Test_concernedAlbum_rollup(t *testing.T) {
	albumNoTracks := newConcernedAlbum(files.AlbumMaker{
		Title:     "album",
		Directory: "album",
	}.NewAlbum(false))
	album1 := files.AlbumMaker{Title: "album1", Directory: "album1"}.NewAlbum(false)
	files.TrackMaker{Album: album1}.NewTrack(true)
	files.TrackMaker{Album: album1}.NewTrack(true)
	albumWithTracksNoConcerns := newConcernedAlbum(album1)
	album2 := files.AlbumMaker{Title: "album2", Directory: "album2"}.NewAlbum(false)
	files.TrackMaker{Album: album2}.NewTrack(true)
	files.TrackMaker{Album: album2}.NewTrack(true)
	albumWithMixedConcerns := newConcernedAlbum(album2)
	if albumWithMixedConcerns != nil {
		albumWithMixedConcerns.tracks()[0].addConcern(filesConcern, "no metadata")
	}
	album3 := files.AlbumMaker{Title: "album3", Directory: "album3"}.NewAlbum(false)
	files.TrackMaker{Album: album3}.NewTrack(true)
	files.TrackMaker{Album: album3}.NewTrack(true)
	albumWithIdenticalConcerns := newConcernedAlbum(album3)
	if albumWithIdenticalConcerns != nil {
		for _, cT := range albumWithIdenticalConcerns.tracks() {
			cT.addConcern(filesConcern, "no metadata")
		}
	}
	tests := map[string]struct {
		cAl  *concernedAlbum
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
			if got := tt.cAl.rollup(); got != tt.want {
				t.Errorf("concernedAlbum.rollup() = %v, want %v", got, tt.want)
			}
			if tt.want {
				if !tt.cAl.concerns.isConcerned() {
					t.Errorf("concernedAlbum.rollup() after rollup has no concerns")
				}
				for _, cT := range tt.cAl.tracks() {
					if cT.concerns.isConcerned() {
						t.Errorf("concernedAlbum.rollup() after rollup, track has concerns")
					}
				}
			}
		})
	}
}
