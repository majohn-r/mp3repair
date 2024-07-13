package files

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/spf13/afero"
)

var (
	// id3v1DataSet1 is a sample ID3V1 tag from an existing file
	id3v1DataSet1 = []byte{
		'T', 'A', 'G',
		'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r', 'o', 'f',
		'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i', 'e', 'w',
		'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A', 't', ' ',
		'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l', 'u', 'm',
		'2', '0', '1', '3',
		' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		0,
		29,
		12,
	}
	// id3v1DataSet2 is a sample ID3V1 tag from an existing file
	id3v1DataSet2 = []byte{
		'T', 'A', 'G',
		'J', 'u', 'l', 'i', 'a', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', ' ', ' ', ' ', ' ', ' ',
		' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		'T', 'h', 'e', ' ', 'W', 'h', 'i', 't', 'e', ' ', 'A', 'l', 'b', 'u', 'm', ' ',
		'[', 'D', 'i', 's', 'c', ' ', '1', ']', ' ', ' ', ' ', ' ', ' ', ' ',
		'1', '9', '6', '8',
		' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		0,
		17,
		17,
	}
)

func Test_Trim(t *testing.T) {
	tests := map[string]struct {
		s    string
		want string
	}{
		"no trailing data": {s: "foo", want: "foo"},
		"trailing space":   {s: "foo            ", want: "foo"},
		"trailing nulls": {
			s:    string([]byte{'f', 'o', 'o', 0, 0, 0, 0, 0}),
			want: "foo",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := trim(tt.s); got != tt.want {
				t.Errorf("trim() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newID3v1MetadataWithData(b []byte) *id3v1Metadata {
	return newID3v1Metadata().withData(b)
}

func TestNewId3v1MetadataWithData(t *testing.T) {
	tests := map[string]struct {
		b    []byte
		want *id3v1Metadata
	}{
		"short data": {
			b: []byte{1, 2, 3, 4},
			want: newID3v1Metadata().withData([]byte{
				1, 2, 3, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
			),
		},
		"just right": {
			b:    id3v1DataSet1,
			want: newID3v1Metadata().withData(id3v1DataSet1),
		},
		"too much data": {
			b: []byte{
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				10, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				20, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				30, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				40, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				50, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				60, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				70, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				80, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
			},
			want: newID3v1Metadata().withData([]byte{
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				10, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				20, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				30, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				40, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				50, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				60, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				70, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
			},
			),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := newID3v1MetadataWithData(tt.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newID3v1MetadataWithData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestId3v1MetadataIsValid(t *testing.T) {
	tests := map[string]struct {
		v1   *id3v1Metadata
		want bool
	}{
		"expected": {v1: newID3v1MetadataWithData(id3v1DataSet1), want: true},
		"bad":      {v1: newID3v1MetadataWithData([]byte{0, 1, 2})},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.v1.IsValid(); got != tt.want {
				t.Errorf("id3v1Metadata.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestId3v1MetadataTitle(t *testing.T) {
	tests := map[string]struct {
		v1   *id3v1Metadata
		want string
	}{
		"ringo": {
			v1:   newID3v1MetadataWithData(id3v1DataSet1),
			want: "Ringo - Pop Profile [Interview",
		},
		"julia": {v1: newID3v1MetadataWithData(id3v1DataSet2), want: "Julia"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.v1.title(); got != tt.want {
				t.Errorf("id3v1Metadata.title() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestId3v1MetadataSetTitle(t *testing.T) {
	tests := map[string]struct {
		v1   *id3v1Metadata
		s    string
		want *id3v1Metadata
	}{
		"short title": {
			v1: newID3v1MetadataWithData(id3v1DataSet1),
			s:  "short title",
			want: newID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				's', 'h', 'o', 'r', 't', ' ', 't', 'i', 't', 'l', 'e', 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A', 't',
				' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l', 'u', 'm',
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				12,
			}),
		},
		"long title": {
			v1: newID3v1MetadataWithData(id3v1DataSet1),
			s:  "very long title, so long it cannot be copied intact",
			want: newID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g', ' ', 't', 'i', 't', 'l', 'e',
				',', ' ', 's', 'o', ' ', 'l', 'o', 'n', 'g', ' ', 'i', 't', ' ', 'c', 'a',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A', 't',
				' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l', 'u', 'm',
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				12,
			}),
		},
		"non-ASCII title": {
			v1: newID3v1MetadataWithData(id3v1DataSet1),
			s:  "Grohg - Cortège Macabre",
			want: newID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'G', 'r', 'o', 'h', 'g', ' ', '-', ' ', 'C', 'o', 'r', 't', 0xE8, 'g',
				'e', ' ', 'M', 'a', 'c', 'a', 'b', 'r', 'e', 0, 0, 0, 0, 0, 0, 0,
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A', 't',
				' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l', 'u', 'm',
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				12,
			}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.v1.setTitle(tt.s)
			if !reflect.DeepEqual(tt.v1, tt.want) {
				t.Errorf("id3v1Metadata.setTitle() got %v want %v", tt.v1, tt.want)
			}
		})
	}
}

func Test_Id3v1MetadataArtist(t *testing.T) {
	tests := map[string]struct {
		v1   *id3v1Metadata
		want string
	}{
		"beatles1": {v1: newID3v1MetadataWithData(id3v1DataSet1), want: "The Beatles"},
		"beatles2": {v1: newID3v1MetadataWithData(id3v1DataSet2), want: "The Beatles"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.v1.artist(); got != tt.want {
				t.Errorf("id3v1Metadata.artist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestId3v1MetadataSetArtist(t *testing.T) {
	tests := map[string]struct {
		v1   *id3v1Metadata
		s    string
		want *id3v1Metadata
	}{
		"short name": {
			v1: newID3v1MetadataWithData(id3v1DataSet1),
			s:  "shorties",
			want: newID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r',
				'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i',
				'e', 'w',
				's', 'h', 'o', 'r', 't', 'i', 'e', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A',
				't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l',
				'u', 'm',
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				12,
			}),
		},
		"long name": {
			v1: newID3v1MetadataWithData(id3v1DataSet1),
			s:  "The greatest band ever known, bar none",
			want: newID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r',
				'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i',
				'e', 'w',
				'T', 'h', 'e', ' ', 'g', 'r', 'e', 'a', 't', 'e', 's', 't', ' ', 'b',
				'a', 'n', 'd', ' ', 'e', 'v', 'e', 'r', ' ', 'k', 'n', 'o', 'w', 'n',
				',', ' ',
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A',
				't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l',
				'u', 'm',
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				12,
			}),
		},
		"non-ASCII name": {
			v1: newID3v1MetadataWithData(id3v1DataSet1),
			s:  "Antonín Dvořák",
			want: newID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r',
				'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i',
				'e', 'w',
				'A', 'n', 't', 'o', 'n', 0xED, 'n', ' ', 'D', 'v', 'o', 'r', 0xE1, 'k',
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A',
				't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l',
				'u', 'm',
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				12,
			}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.v1.setArtist(tt.s)
			if !reflect.DeepEqual(tt.v1, tt.want) {
				t.Errorf("id3v1Metadata.setArtist() got %v want %v", tt.v1, tt.want)
			}
		})
	}
}

func TestId3v1MetadataAlbum(t *testing.T) {
	tests := map[string]struct {
		v1   *id3v1Metadata
		want string
	}{
		"BBC": {
			v1:   newID3v1MetadataWithData(id3v1DataSet1),
			want: "On Air: Live At The BBC, Volum",
		},
		"White Album": {
			v1:   newID3v1MetadataWithData(id3v1DataSet2),
			want: "The White Album [Disc 1]",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.v1.album(); got != tt.want {
				t.Errorf("id3v1Metadata.album() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestId3v1MetadataSetAlbum(t *testing.T) {
	tests := map[string]struct {
		v1   *id3v1Metadata
		s    string
		want *id3v1Metadata
	}{
		"short name": {
			v1: newID3v1MetadataWithData(id3v1DataSet1),
			s:  "!",
			want: newID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r',
				'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i',
				'e', 'w',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'!', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0,
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				12,
			}),
		},
		"long name": {
			v1: newID3v1MetadataWithData(id3v1DataSet1),
			s:  "The Most Amazing Album Ever Released",
			want: newID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r',
				'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i',
				'e', 'w',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'T', 'h', 'e', ' ', 'M', 'o', 's', 't', ' ', 'A', 'm', 'a', 'z', 'i',
				'n', 'g', ' ', 'A', 'l', 'b', 'u', 'm', ' ', 'E', 'v', 'e', 'r', ' ',
				'R', 'e',
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				12,
			}),
		},
		"non-ASCII name": {
			v1: newID3v1MetadataWithData(id3v1DataSet1),
			s:  "Déjà Vu",
			want: newID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r',
				'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i',
				'e', 'w',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'D', 0xE9, 'j', 0xE0, ' ', 'V', 'u', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				12,
			}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.v1.setAlbum(tt.s)
			if !reflect.DeepEqual(tt.v1, tt.want) {
				t.Errorf("id3v1Metadata.setAlbum() got %v want %v", tt.v1, tt.want)
			}
		})
	}
}

func TestId3v1MetadataYear(t *testing.T) {
	tests := map[string]struct {
		v1   *id3v1Metadata
		want string
	}{
		"BBC":         {v1: newID3v1MetadataWithData(id3v1DataSet1), want: "2013"},
		"White Album": {v1: newID3v1MetadataWithData(id3v1DataSet2), want: "1968"},
		"no date":     {v1: newID3v1Metadata(), want: ""},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.v1.year()
			if got != tt.want {
				t.Errorf("id3v1Metadata.year() gotY = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestId3v1MetadataSetYear(t *testing.T) {
	tests := map[string]struct {
		v1     *id3v1Metadata
		s      string
		wantV1 *id3v1Metadata
	}{
		"realistic": {
			v1: newID3v1MetadataWithData(id3v1DataSet1),
			s:  "2022",
			wantV1: newID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r',
				'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i',
				'e', 'w',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A',
				't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l',
				'u', 'm',
				'2', '0', '2', '2',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				12,
			}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.v1.setYear(tt.s)
			if !reflect.DeepEqual(tt.v1, tt.wantV1) {
				t.Errorf("id3v1Metadata.setYear() got %v want %v", tt.v1, tt.wantV1)
			}
		})
	}
}

func TestId3v1MetadataComment(t *testing.T) {
	tests := map[string]struct {
		v1   *id3v1Metadata
		want string
	}{
		"BBC":         {v1: newID3v1MetadataWithData(id3v1DataSet1)},
		"White Album": {v1: newID3v1MetadataWithData(id3v1DataSet2)},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.v1.comment(); got != tt.want {
				t.Errorf("id3v1Metadata.comment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestId3v1MetadataSetComment(t *testing.T) {
	tests := map[string]struct {
		v1   *id3v1Metadata
		s    string
		want *id3v1Metadata
	}{
		"typical comment": {
			v1: newID3v1MetadataWithData(id3v1DataSet1),
			s:  "",
			want: newID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r',
				'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i',
				'e', 'w',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A',
				't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l',
				'u', 'm',
				'2', '0', '1', '3',
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0,
				0,
				29,
				12,
			}),
		},
		"long winded": {
			v1: newID3v1MetadataWithData(id3v1DataSet1),
			s:  "This track is genuinely insightful",
			want: newID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r',
				'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i',
				'e', 'w',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A',
				't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l',
				'u', 'm',
				'2', '0', '1', '3',
				'T', 'h', 'i', 's', ' ', 't', 'r', 'a', 'c', 'k', ' ', 'i', 's', ' ',
				'g', 'e', 'n', 'u', 'i', 'n', 'e', 'l', 'y', ' ', 'i', 'n', 's', 'i',
				0,
				29,
				12,
			}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.v1.setComment(tt.s)
			if !reflect.DeepEqual(tt.v1, tt.want) {
				t.Errorf("id3v1Metadata.setComment() got %v want %v", tt.v1, tt.want)
			}
		})
	}
}

func TestId3v1MetadataTrack(t *testing.T) {
	tests := map[string]struct {
		v1     *id3v1Metadata
		wantI  int
		wantOk bool
	}{
		"BBC": {
			v1:     newID3v1MetadataWithData(id3v1DataSet1),
			wantI:  29,
			wantOk: true,
		},
		"White Album": {
			v1:     newID3v1MetadataWithData(id3v1DataSet2),
			wantI:  17,
			wantOk: true,
		},
		"bad zero byte": {
			v1: newID3v1MetadataWithData([]byte{
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
			}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotI, gotOk := tt.v1.track()
			if gotI != tt.wantI {
				t.Errorf("id3v1Metadata.track() gotI = %v, want %v", gotI, tt.wantI)
			}
			if gotOk != tt.wantOk {
				t.Errorf("id3v1Metadata.track() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestId3v1MetadataSetTrack(t *testing.T) {
	tests := map[string]struct {
		v1     *id3v1Metadata
		t      int
		want   bool
		wantV1 *id3v1Metadata
	}{
		"low": {
			v1:     newID3v1MetadataWithData(id3v1DataSet1),
			t:      0,
			want:   false,
			wantV1: newID3v1MetadataWithData(id3v1DataSet1),
		},
		"high": {
			v1:     newID3v1MetadataWithData(id3v1DataSet1),
			t:      256,
			want:   false,
			wantV1: newID3v1MetadataWithData(id3v1DataSet1),
		},
		"ok": {
			v1:   newID3v1MetadataWithData(id3v1DataSet1),
			t:    45,
			want: true,
			wantV1: newID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r',
				'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i',
				'e', 'w',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A',
				't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l',
				'u', 'm',
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				45,
				12,
			}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.v1.setTrack(tt.t); got != tt.want {
				t.Errorf("id3v1Metadata.setTrack() = %t, want %t", got, tt.want)
			}
			if !reflect.DeepEqual(tt.v1, tt.wantV1) {
				t.Errorf("id3v1Metadata.setTrack() got %v want %v", tt.v1, tt.wantV1)
			}
		})
	}
}

func TestId3v1MetadataGenre(t *testing.T) {
	tests := map[string]struct {
		v1     *id3v1Metadata
		wantS  string
		wantOk bool
	}{
		"BBC": {
			v1:     newID3v1MetadataWithData(id3v1DataSet1),
			wantS:  "other",
			wantOk: true,
		},
		"White Album": {
			v1:     newID3v1MetadataWithData(id3v1DataSet2),
			wantS:  "rock",
			wantOk: true,
		},
		"bad zero byte": {
			v1: newID3v1MetadataWithData([]byte{
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 254,
			}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotS, gotOk := tt.v1.genre()
			if gotS != tt.wantS {
				t.Errorf("id3v1Metadata.genre() gotS = %v, want %v", gotS, tt.wantS)
			}
			if gotOk != tt.wantOk {
				t.Errorf("id3v1Metadata.genre() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestId3v1MetadataSetGenre(t *testing.T) {
	tests := map[string]struct {
		v1     *id3v1Metadata
		s      string
		wantV1 *id3v1Metadata
	}{
		"no such genre": {
			v1: newID3v1MetadataWithData(id3v1DataSet1),
			s:  "Subspace Radio",
			wantV1: newID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r',
				'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i',
				'e', 'w',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A',
				't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l',
				'u', 'm',
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				12,
			}),
		},
		"known genre": {
			v1: newID3v1MetadataWithData(id3v1DataSet1),
			s:  "Sound clip",
			wantV1: newID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r',
				'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i',
				'e', 'w',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A',
				't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l',
				'u', 'm',
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				37,
			}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.v1.setGenre(tt.s)
			if !reflect.DeepEqual(tt.v1, tt.wantV1) {
				t.Errorf("id3v1Metadata.setGenre() got %v want %v", tt.v1, tt.wantV1)
			}
		})
	}
}

func TestInternalReadId3V1Metadata(t *testing.T) {
	originalFileSystem := cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	testDir := "id3v1read"
	_ = cmdtoolkit.Mkdir(testDir)
	defer func() {
		cmdtoolkit.AssignFileSystem(originalFileSystem)
	}()
	shortFile := "short.mp3"
	_ = createFileWithContent(testDir, shortFile, []byte{0, 1, 2})
	badFile := "bad.mp3"
	_ = createFileWithContent(testDir, badFile, []byte{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	})
	goodFile := "good.mp3"
	payload := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	}
	payload = append(payload, id3v1DataSet1...)
	_ = createFileWithContent(testDir, goodFile, payload)
	type args struct {
		path     string
		readFunc func(f afero.File, b []byte) (int, error)
	}
	tests := map[string]struct {
		args
		want    *id3v1Metadata
		wantErr bool
	}{
		"non-existent file": {
			args:    args{path: "./non-existent", readFunc: nil},
			want:    nil,
			wantErr: true,
		},
		"short file": {
			args:    args{path: filepath.Join(testDir, shortFile), readFunc: nil},
			want:    nil,
			wantErr: true,
		},
		"read with error": {
			args: args{
				path: filepath.Join(testDir, badFile),
				readFunc: func(f afero.File, b []byte) (int, error) {
					return 0, fmt.Errorf("oops")
				},
			},
			want:    nil,
			wantErr: true,
		},
		"short read": {
			args: args{
				path: filepath.Join(testDir, badFile),
				readFunc: func(f afero.File, b []byte) (int, error) {
					return 127, nil
				},
			},
			want:    nil,
			wantErr: true,
		},
		"bad file": {
			args:    args{path: filepath.Join(testDir, badFile), readFunc: fileReader},
			want:    nil,
			wantErr: true,
		},
		"good file": {
			args:    args{path: filepath.Join(testDir, goodFile), readFunc: fileReader},
			want:    newID3v1MetadataWithData(id3v1DataSet1),
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotErr := internalReadID3V1Metadata(tt.args.path, tt.args.readFunc)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("internalReadID3V1Metadata error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("internalReadID3V1Metadata = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadId3v1Metadata(t *testing.T) {
	originalFileSystem := cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmdtoolkit.AssignFileSystem(originalFileSystem)
	}()
	testDir := "id3v1read"
	_ = cmdtoolkit.Mkdir(testDir)
	goodFile := "good.mp3"
	payload := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	}
	payload = append(payload, id3v1DataSet1...)
	_ = createFileWithContent(testDir, goodFile, payload)
	tests := map[string]struct {
		path    string
		want    []string
		wantErr bool
	}{
		// only testing good path ... all the error paths are handled in the
		// internal read test
		"good file": {
			path: filepath.Join(testDir, goodFile),
			want: []string{
				`Artist: "The Beatles"`,
				`Album: "On Air: Live At The BBC, Volum"`,
				`Title: "Ringo - Pop Profile [Interview"`,
				"Track: 29",
				`Year: "2013"`,
				`Genre: "other"`,
			},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotErr := readID3v1Metadata(tt.path)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("readID3v1Metadata error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readID3v1Metadata = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestId3v1MetadataInternalWrite(t *testing.T) {
	originalFileSystem := cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmdtoolkit.AssignFileSystem(originalFileSystem)
	}()
	testDir := "id3v1write"
	_ = cmdtoolkit.Mkdir(testDir)
	shortFile := "short.mp3"
	_ = createFileWithContent(testDir, shortFile, []byte{0, 1, 2})
	goodFile := "good.mp3"
	payload := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	}
	payload = append(payload, id3v1DataSet1...)
	_ = createFileWithContent(testDir, goodFile, payload)
	type args struct {
		oldPath   string
		writeFunc func(f afero.File, b []byte) (int, error)
	}
	tests := map[string]struct {
		v1 *id3v1Metadata
		args
		wantErr  bool
		wantData []byte
	}{
		"non-existent file": {args: args{oldPath: "./no such file"}, wantErr: true},
		"short file": {
			args:    args{oldPath: filepath.Join(testDir, shortFile)},
			wantErr: true,
		},
		"error on write": {
			v1: newID3v1MetadataWithData(id3v1DataSet1),
			args: args{
				oldPath: filepath.Join(testDir, goodFile),
				writeFunc: func(f afero.File, b []byte) (int, error) {
					return 0, fmt.Errorf("ruh-roh")
				},
			},
			wantErr: true,
		},
		"short write": {
			v1: newID3v1MetadataWithData(id3v1DataSet1),
			args: args{
				oldPath: filepath.Join(testDir, goodFile),
				writeFunc: func(f afero.File, b []byte) (int, error) {
					return 127, nil
				},
			},
			wantErr: true,
		},
		"good write": {
			v1: newID3v1MetadataWithData(id3v1DataSet2),
			args: args{
				oldPath:   filepath.Join(testDir, goodFile),
				writeFunc: writeToFile,
			},
			wantData: []byte{
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				'T', 'A', 'G',
				'J', 'u', 'l', 'i', 'a', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				' ', ' ',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', ' ', ' ', ' ',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				' ', ' ',
				'T', 'h', 'e', ' ', 'W', 'h', 'i', 't', 'e', ' ', 'A', 'l', 'b', 'u',
				'm', ' ', '[', 'D', 'i', 's', 'c', ' ', '1', ']', ' ', ' ', ' ', ' ',
				' ', ' ',
				'1', '9', '6', '8',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				17,
				17,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var gotErr error
			if gotErr = tt.v1.internalWrite(tt.args.oldPath, tt.args.writeFunc); (gotErr != nil) != tt.wantErr {
				t.Errorf("id3v1Metadata.internalWrite() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
			if gotErr == nil && tt.wantErr == false {
				got, _ := afero.ReadFile(cmdtoolkit.FileSystem(), tt.args.oldPath)
				if !reflect.DeepEqual(got, tt.wantData) {
					t.Errorf("id3v1Metadata.internalWrite() got %v want %v", got, tt.wantData)
				}
			}
		})
	}
}

func TestId3v1MetadataWrite(t *testing.T) {
	originalFileSystem := cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmdtoolkit.AssignFileSystem(originalFileSystem)
	}()
	testDir := "id3v1write"
	_ = cmdtoolkit.Mkdir(testDir)
	goodFile := "good.mp3"
	payload := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	}
	payload = append(payload, id3v1DataSet1...)
	_ = createFileWithContent(testDir, goodFile, payload)
	tests := map[string]struct {
		v1       *id3v1Metadata
		path     string
		wantErr  bool
		wantData []byte
	}{
		"happy place": {
			v1:   newID3v1MetadataWithData(id3v1DataSet2),
			path: filepath.Join(testDir, goodFile),
			wantData: []byte{
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				'T', 'A', 'G',
				'J', 'u', 'l', 'i', 'a', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				' ', ' ',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', ' ', ' ', ' ',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				' ', ' ',
				'T', 'h', 'e', ' ', 'W', 'h', 'i', 't', 'e', ' ', 'A', 'l', 'b', 'u',
				'm', ' ', '[', 'D', 'i', 's', 'c', ' ', '1', ']', ' ', ' ', ' ', ' ',
				' ', ' ',
				'1', '9', '6', '8',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				17,
				17,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var gotErr error
			if gotErr = tt.v1.write(tt.path); (gotErr != nil) != tt.wantErr {
				t.Errorf("id3v1Metadata.write() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
			if gotErr == nil && tt.wantErr == false {
				got, _ := afero.ReadFile(cmdtoolkit.FileSystem(), tt.path)
				if !reflect.DeepEqual(got, tt.wantData) {
					t.Errorf("id3v1Metadata.write() got %v want %v", got, tt.wantData)
				}
			}
		})
	}
}

func TestId3v1NameDiffers(t *testing.T) {
	tests := map[string]struct {
		cS   *ComparableStrings
		want bool
	}{
		"identical strings": {
			cS: &ComparableStrings{
				External: "Fiddler On The Roof",
				Metadata: "Fiddler On The Roof",
			}, want: false,
		},
		"unusable characters in metadata": {
			cS: &ComparableStrings{
				External: "Theme From M-A-S-H",
				Metadata: "Theme From M*A*S*H",
			},
			want: false,
		},
		"really long name": {
			cS: &ComparableStrings{
				External: "A Funny Thing Happened On The Way To The Forum 1996 Broadway Revival Cast",
				Metadata: "A Funny Thing Happened On The",
			},
			want: false,
		},
		"non-ASCII values": {
			cS: &ComparableStrings{
				External: "Grohg - Cortège Macabre",
				Metadata: "Grohg - Cort\xe8ge Macabre",
			},
			want: false,
		},
		"larger non-ASCII values": {
			cS: &ComparableStrings{
				External: "Dvořák",
				Metadata: "Dvor\xe1k",
			},
			want: false,
		},
		"identical strings with case differences": {
			cS: &ComparableStrings{
				External: "SIMPLE name",
				Metadata: "simple NAME",
			},
			want: false,
		},
		"strings of different length within name length limit": {
			cS: &ComparableStrings{
				External: "simple name",
				Metadata: "artist: simple name",
			},
			want: true,
		},
		"use of runes that are illegal for file names": {
			cS: &ComparableStrings{
				External: "simple_name",
				Metadata: "simple:name",
			},
			want: false,
		},
		"complex mismatch": {
			cS: &ComparableStrings{
				External: "simple_name",
				Metadata: "simple: nam",
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := id3v1NameDiffers(tt.cS); got != tt.want {
				t.Errorf("id3v1NameDiffers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestId3v1GenreDiffers(t *testing.T) {
	tests := map[string]struct {
		cS   *ComparableStrings
		want bool
	}{
		"match": {
			cS: &ComparableStrings{
				External: "Classic Rock",
				Metadata: "Classic Rock",
			},
			want: false,
		},
		"case does not match": {
			cS: &ComparableStrings{
				External: "Classic Rock",
				Metadata: "classic rock",
			},
			want: false,
		},
		"other": {
			cS: &ComparableStrings{
				External: "Prog Rock",
				Metadata: "other",
			},
			want: false,
		},
		"known genre": {
			// known id3v1 genre - "Other" will not match
			cS: &ComparableStrings{
				External: "Classic Rock",
				Metadata: "Other",
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := id3v1GenreDiffers(tt.cS); got != tt.want {
				t.Errorf("id3v1GenreDiffers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBiDirectionalMap_AddPair(t *testing.T) {
	tests := map[string]struct {
		m       *biDirectionalMap[string, string]
		input   map[string]string
		wantErr bool
	}{
		"redundant key": {
			m: newBiDirectionalMap[string, string](strings.ToLower, nil),
			input: map[string]string{
				"k1": "v1",
				"K1": "v2",
			},
			wantErr: true,
		},
		"redundant value": {
			m: newBiDirectionalMap[string, string](nil, nil),
			input: map[string]string{
				"k1": "v1",
				"k2": "v1",
			},
			wantErr: true,
		},
		"good": {
			m: newBiDirectionalMap[string, string](nil, nil),
			input: map[string]string{
				"k1": "v1",
				"k2": "v2",
			},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var got error
			for k, v := range tt.input {
				if got = tt.m.addPair(k, v); got != nil {
					break
				}
			}
			if (got != nil) != tt.wantErr {
				t.Errorf("BiDirectionalMap.addPair() error = %v, wantErr %v", got, tt.wantErr)
			}
		})
	}
}

func Test_populate(t *testing.T) {
	_ = initGenres()
	tests := map[string]struct {
		m       map[int]string
		wantMap *biDirectionalMap[int, string]
		wantErr bool
	}{
		"error": {
			m: map[int]string{
				1: "foo",
				2: "foo",
			},
			wantMap: nil,
			wantErr: true,
		},
		"normal": {
			m:       lcGenres,
			wantMap: genres,
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := populate(tt.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("populate() error = %v, wantErr %v", err, tt.wantErr)
			}
			switch {
			case got == nil && tt.wantMap == nil:
				// good
			case got == nil || tt.wantMap == nil:
				t.Errorf("populate() map = %v, wantMap %v", got, tt.wantMap)
			default:
				if !reflect.DeepEqual(got.keyMap(), tt.wantMap.keyMap()) ||
					!reflect.DeepEqual(got.valueMap(), tt.wantMap.valueMap()) {
					t.Errorf("populate() map = %v, wantMap %v", got, tt.wantMap)
				}
			}
		})
	}
}
