package files

import "testing"

func TestAlbum_RecordingArtistName(t *testing.T) {
	tests := []struct {
		name string
		a    *Album
		want string
	}{
		{name: "with recording artist", a: NewAlbum("album1", &Artist{Name: "artist1"}, ""), want: "artist1"},
		{name: "no recording artist", a: NewAlbum("album1", nil, ""), want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.RecordingArtistName(); got != tt.want {
				t.Errorf("Album.RecordingArtistName() = %v, want %v", got, tt.want)
			}
		})
	}
}
