package files

import "testing"

func TestAlbum_RecordingArtistName(t *testing.T) {
	const fnName = "Album.RecordingArtistName()"
	tests := map[string]struct {
		a    *Album
		want string
	}{
		"with recording artist": {a: NewAlbum("album1", NewArtist("artist1", ""), ""), want: "artist1"},
		"no recording artist":   {a: NewAlbum("album1", nil, ""), want: ""},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.a.RecordingArtistName(); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}
