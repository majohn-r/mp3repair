package files_test

import (
	"fmt"
	"mp3repair/internal/files"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
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
			if got := files.Trim(tt.s); got != tt.want {
				t.Errorf("Trim() = %v, want %v", got, tt.want)
			}
		})
	}
}

func NewID3v1MetadataWithData(b []byte) *files.Id3v1Metadata {
	return files.NewID3v1Metadata().WithData(b)
}

func TestNewId3v1MetadataWithData(t *testing.T) {
	tests := map[string]struct {
		b    []byte
		want *files.Id3v1Metadata
	}{
		"short data": {
			b: []byte{1, 2, 3, 4},
			want: files.NewID3v1Metadata().WithData([]byte{
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
			want: files.NewID3v1Metadata().WithData(id3v1DataSet1),
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
			want: files.NewID3v1Metadata().WithData([]byte{
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
			if got := NewID3v1MetadataWithData(tt.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewID3v1MetadataWithData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestId3v1MetadataIsValid(t *testing.T) {
	tests := map[string]struct {
		v1   *files.Id3v1Metadata
		want bool
	}{
		"expected": {v1: NewID3v1MetadataWithData(id3v1DataSet1), want: true},
		"bad":      {v1: NewID3v1MetadataWithData([]byte{0, 1, 2})},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.v1.IsValid(); got != tt.want {
				t.Errorf("Id3v1Metadata.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestId3v1MetadataTitle(t *testing.T) {
	tests := map[string]struct {
		v1   *files.Id3v1Metadata
		want string
	}{
		"ringo": {
			v1:   NewID3v1MetadataWithData(id3v1DataSet1),
			want: "Ringo - Pop Profile [Interview",
		},
		"julia": {v1: NewID3v1MetadataWithData(id3v1DataSet2), want: "Julia"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.v1.Title(); got != tt.want {
				t.Errorf("Id3v1Metadata.Title() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestId3v1MetadataSetTitle(t *testing.T) {
	tests := map[string]struct {
		v1   *files.Id3v1Metadata
		s    string
		want *files.Id3v1Metadata
	}{
		"short title": {
			v1: NewID3v1MetadataWithData(id3v1DataSet1),
			s:  "short title",
			want: NewID3v1MetadataWithData([]byte{
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
			v1: NewID3v1MetadataWithData(id3v1DataSet1),
			s:  "very long title, so long it cannot be copied intact",
			want: NewID3v1MetadataWithData([]byte{
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
			v1: NewID3v1MetadataWithData(id3v1DataSet1),
			s:  "Grohg - Cortège Macabre",
			want: NewID3v1MetadataWithData([]byte{
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
			tt.v1.SetTitle(tt.s)
			if !reflect.DeepEqual(tt.v1, tt.want) {
				t.Errorf("Id3v1Metadata.SetTitle() got %v want %v", tt.v1, tt.want)
			}
		})
	}
}

func Test_Id3v1MetadataArtist(t *testing.T) {
	tests := map[string]struct {
		v1   *files.Id3v1Metadata
		want string
	}{
		"beatles1": {v1: NewID3v1MetadataWithData(id3v1DataSet1), want: "The Beatles"},
		"beatles2": {v1: NewID3v1MetadataWithData(id3v1DataSet2), want: "The Beatles"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.v1.Artist(); got != tt.want {
				t.Errorf("Id3v1Metadata.Artist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestId3v1MetadataSetArtist(t *testing.T) {
	tests := map[string]struct {
		v1   *files.Id3v1Metadata
		s    string
		want *files.Id3v1Metadata
	}{
		"short name": {
			v1: NewID3v1MetadataWithData(id3v1DataSet1),
			s:  "shorties",
			want: NewID3v1MetadataWithData([]byte{
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
			v1: NewID3v1MetadataWithData(id3v1DataSet1),
			s:  "The greatest band ever known, bar none",
			want: NewID3v1MetadataWithData([]byte{
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
			v1: NewID3v1MetadataWithData(id3v1DataSet1),
			s:  "Antonín Dvořák",
			want: NewID3v1MetadataWithData([]byte{
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
			tt.v1.SetArtist(tt.s)
			if !reflect.DeepEqual(tt.v1, tt.want) {
				t.Errorf("Id3v1Metadata.SetArtist() got %v want %v", tt.v1, tt.want)
			}
		})
	}
}

func TestId3v1MetadataAlbum(t *testing.T) {
	tests := map[string]struct {
		v1   *files.Id3v1Metadata
		want string
	}{
		"BBC": {
			v1:   NewID3v1MetadataWithData(id3v1DataSet1),
			want: "On Air: Live At The BBC, Volum",
		},
		"White Album": {
			v1:   NewID3v1MetadataWithData(id3v1DataSet2),
			want: "The White Album [Disc 1]",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.v1.Album(); got != tt.want {
				t.Errorf("Id3v1Metadata.Album() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestId3v1MetadataSetAlbum(t *testing.T) {
	tests := map[string]struct {
		v1   *files.Id3v1Metadata
		s    string
		want *files.Id3v1Metadata
	}{
		"short name": {
			v1: NewID3v1MetadataWithData(id3v1DataSet1),
			s:  "!",
			want: NewID3v1MetadataWithData([]byte{
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
			v1: NewID3v1MetadataWithData(id3v1DataSet1),
			s:  "The Most Amazing Album Ever Released",
			want: NewID3v1MetadataWithData([]byte{
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
			v1: NewID3v1MetadataWithData(id3v1DataSet1),
			s:  "Déjà Vu",
			want: NewID3v1MetadataWithData([]byte{
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
			tt.v1.SetAlbum(tt.s)
			if !reflect.DeepEqual(tt.v1, tt.want) {
				t.Errorf("Id3v1Metadata.SetAlbum() got %v want %v", tt.v1, tt.want)
			}
		})
	}
}

func TestId3v1MetadataYear(t *testing.T) {
	tests := map[string]struct {
		v1   *files.Id3v1Metadata
		want string
	}{
		"BBC":         {v1: NewID3v1MetadataWithData(id3v1DataSet1), want: "2013"},
		"White Album": {v1: NewID3v1MetadataWithData(id3v1DataSet2), want: "1968"},
		"no date":     {v1: files.NewID3v1Metadata(), want: ""},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.v1.Year()
			if got != tt.want {
				t.Errorf("Id3v1Metadata.Year() gotY = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestId3v1MetadataSetYear(t *testing.T) {
	tests := map[string]struct {
		v1     *files.Id3v1Metadata
		s      string
		wantv1 *files.Id3v1Metadata
	}{
		"realistic": {
			v1: NewID3v1MetadataWithData(id3v1DataSet1),
			s:  "2022",
			wantv1: NewID3v1MetadataWithData([]byte{
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
			tt.v1.SetYear(tt.s)
			if !reflect.DeepEqual(tt.v1, tt.wantv1) {
				t.Errorf("Id3v1Metadata.SetYear() got %v want %v", tt.v1, tt.wantv1)
			}
		})
	}
}

func TestId3v1MetadataComment(t *testing.T) {
	tests := map[string]struct {
		v1   *files.Id3v1Metadata
		want string
	}{
		"BBC":         {v1: NewID3v1MetadataWithData(id3v1DataSet1)},
		"White Album": {v1: NewID3v1MetadataWithData(id3v1DataSet2)},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.v1.Comment(); got != tt.want {
				t.Errorf("Id3v1Metadata.Comment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestId3v1MetadataSetComment(t *testing.T) {
	tests := map[string]struct {
		v1   *files.Id3v1Metadata
		s    string
		want *files.Id3v1Metadata
	}{
		"typical comment": {
			v1: NewID3v1MetadataWithData(id3v1DataSet1),
			s:  "",
			want: NewID3v1MetadataWithData([]byte{
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
			v1: NewID3v1MetadataWithData(id3v1DataSet1),
			s:  "This track is genuinely insightful",
			want: NewID3v1MetadataWithData([]byte{
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
			tt.v1.SetComment(tt.s)
			if !reflect.DeepEqual(tt.v1, tt.want) {
				t.Errorf("Id3v1Metadata.SetComment() got %v want %v", tt.v1, tt.want)
			}
		})
	}
}

func TestId3v1MetadataTrack(t *testing.T) {
	tests := map[string]struct {
		v1     *files.Id3v1Metadata
		wantI  int
		wantOk bool
	}{
		"BBC": {
			v1:     NewID3v1MetadataWithData(id3v1DataSet1),
			wantI:  29,
			wantOk: true,
		},
		"White Album": {
			v1:     NewID3v1MetadataWithData(id3v1DataSet2),
			wantI:  17,
			wantOk: true,
		},
		"bad zero byte": {
			v1: NewID3v1MetadataWithData([]byte{
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
			gotI, gotOk := tt.v1.Track()
			if gotI != tt.wantI {
				t.Errorf("Id3v1Metadata.Track() gotI = %v, want %v", gotI, tt.wantI)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Id3v1Metadata.Track() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestId3v1MetadataSetTrack(t *testing.T) {
	tests := map[string]struct {
		v1     *files.Id3v1Metadata
		t      int
		want   bool
		wantv1 *files.Id3v1Metadata
	}{
		"low": {
			v1:     NewID3v1MetadataWithData(id3v1DataSet1),
			t:      0,
			want:   false,
			wantv1: NewID3v1MetadataWithData(id3v1DataSet1),
		},
		"high": {
			v1:     NewID3v1MetadataWithData(id3v1DataSet1),
			t:      256,
			want:   false,
			wantv1: NewID3v1MetadataWithData(id3v1DataSet1),
		},
		"ok": {
			v1:   NewID3v1MetadataWithData(id3v1DataSet1),
			t:    45,
			want: true,
			wantv1: NewID3v1MetadataWithData([]byte{
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
			if got := tt.v1.SetTrack(tt.t); got != tt.want {
				t.Errorf("Id3v1Metadata.SetTrack() = %t, want %t", got, tt.want)
			}
			if !reflect.DeepEqual(tt.v1, tt.wantv1) {
				t.Errorf("Id3v1Metadata.SetTrack() got %v want %v", tt.v1, tt.wantv1)
			}
		})
	}
}

func TestId3v1MetadataGenre(t *testing.T) {
	tests := map[string]struct {
		v1     *files.Id3v1Metadata
		wantS  string
		wantOk bool
	}{
		"BBC": {
			v1:     NewID3v1MetadataWithData(id3v1DataSet1),
			wantS:  "Other",
			wantOk: true,
		},
		"White Album": {
			v1:     NewID3v1MetadataWithData(id3v1DataSet2),
			wantS:  "Rock",
			wantOk: true,
		},
		"bad zero byte": {
			v1: NewID3v1MetadataWithData([]byte{
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
			gotS, gotOk := tt.v1.Genre()
			if gotS != tt.wantS {
				t.Errorf("Id3v1Metadata.Genre() gotS = %v, want %v", gotS, tt.wantS)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Id3v1Metadata.Genre() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestId3v1MetadataSetGenre(t *testing.T) {
	tests := map[string]struct {
		v1     *files.Id3v1Metadata
		s      string
		wantv1 *files.Id3v1Metadata
	}{
		"no such genre": {
			v1: NewID3v1MetadataWithData(id3v1DataSet1),
			s:  "Subspace Radio",
			wantv1: NewID3v1MetadataWithData([]byte{
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
			v1: NewID3v1MetadataWithData(id3v1DataSet1),
			s:  files.GenreMap[37],
			wantv1: NewID3v1MetadataWithData([]byte{
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
			tt.v1.SetGenre(tt.s)
			if !reflect.DeepEqual(tt.v1, tt.wantv1) {
				t.Errorf("Id3v1Metadata.SetGenre() got %v want %v", tt.v1, tt.wantv1)
			}
		})
	}
}

func TestInitGenreIndices(t *testing.T) {
	files.InitGenreIndices()
	if len(files.GenreIndicesMap) != len(files.GenreMap) {
		t.Errorf("InitGenreIndices() size of genreIndicesMap is %d, genreMap is %d",
			len(files.GenreIndicesMap), len(files.GenreMap))
	} else {
		for k, v := range files.GenreMap {
			if k2 := files.GenreIndicesMap[strings.ToLower(v)]; k2 != k {
				t.Errorf("InitGenreIndices() index for %q got %d want %d", v, k2, k)
			}
		}
	}
}

func TestInternalReadId3V1Metadata(t *testing.T) {
	originalFileSystem := cmd_toolkit.AssignFileSystem(afero.NewMemMapFs())
	testDir := "id3v1read"
	cmd_toolkit.Mkdir(testDir)
	defer func() {
		cmd_toolkit.AssignFileSystem(originalFileSystem)
	}()
	shortFile := "short.mp3"
	createFileWithContent(testDir, shortFile, []byte{0, 1, 2})
	badFile := "bad.mp3"
	createFileWithContent(testDir, badFile, []byte{
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
	createFileWithContent(testDir, goodFile, payload)
	type args struct {
		path     string
		readFunc func(f afero.File, b []byte) (int, error)
	}
	tests := map[string]struct {
		args
		want    *files.Id3v1Metadata
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
			args:    args{path: filepath.Join(testDir, badFile), readFunc: files.FileReader},
			want:    nil,
			wantErr: true,
		},
		"good file": {
			args:    args{path: filepath.Join(testDir, goodFile), readFunc: files.FileReader},
			want:    NewID3v1MetadataWithData(id3v1DataSet1),
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotErr := files.InternalReadID3V1Metadata(tt.args.path, tt.args.readFunc)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("InternalReadID3V1Metadata error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InternalReadID3V1Metadata = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadId3v1Metadata(t *testing.T) {
	originalFileSystem := cmd_toolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmd_toolkit.AssignFileSystem(originalFileSystem)
	}()
	testDir := "id3v1read"
	cmd_toolkit.Mkdir(testDir)
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
	createFileWithContent(testDir, goodFile, payload)
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
				`Genre: "Other"`,
			},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotErr := files.ReadID3v1Metadata(tt.path)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("ReadID3v1Metadata error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadID3v1Metadata = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestId3v1MetadataInternalWrite(t *testing.T) {
	originalFileSystem := cmd_toolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmd_toolkit.AssignFileSystem(originalFileSystem)
	}()
	testDir := "id3v1write"
	cmd_toolkit.Mkdir(testDir)
	shortFile := "short.mp3"
	createFileWithContent(testDir, shortFile, []byte{0, 1, 2})
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
	createFileWithContent(testDir, goodFile, payload)
	type args struct {
		oldPath   string
		writeFunc func(f afero.File, b []byte) (int, error)
	}
	tests := map[string]struct {
		v1 *files.Id3v1Metadata
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
			v1: NewID3v1MetadataWithData(id3v1DataSet1),
			args: args{
				oldPath: filepath.Join(testDir, goodFile),
				writeFunc: func(f afero.File, b []byte) (int, error) {
					return 0, fmt.Errorf("ruh-roh")
				},
			},
			wantErr: true,
		},
		"short write": {
			v1: NewID3v1MetadataWithData(id3v1DataSet1),
			args: args{
				oldPath: filepath.Join(testDir, goodFile),
				writeFunc: func(f afero.File, b []byte) (int, error) {
					return 127, nil
				},
			},
			wantErr: true,
		},
		"good write": {
			v1: NewID3v1MetadataWithData(id3v1DataSet2),
			args: args{
				oldPath:   filepath.Join(testDir, goodFile),
				writeFunc: files.WriteToFile,
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
			if gotErr = tt.v1.InternalWrite(tt.args.oldPath, tt.args.writeFunc); (gotErr != nil) != tt.wantErr {
				t.Errorf("Id3v1Metadata.InternalWrite() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
			if gotErr == nil && tt.wantErr == false {
				got, _ := afero.ReadFile(cmd_toolkit.FileSystem(), tt.args.oldPath)
				if !reflect.DeepEqual(got, tt.wantData) {
					t.Errorf("Id3v1Metadata.InternalWrite() got %v want %v", got, tt.wantData)
				}
			}
		})
	}
}

func TestId3v1MetadataWrite(t *testing.T) {
	originalFileSystem := cmd_toolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmd_toolkit.AssignFileSystem(originalFileSystem)
	}()
	testDir := "id3v1write"
	cmd_toolkit.Mkdir(testDir)
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
	createFileWithContent(testDir, goodFile, payload)
	tests := map[string]struct {
		v1       *files.Id3v1Metadata
		path     string
		wantErr  bool
		wantData []byte
	}{
		"happy place": {
			v1:   NewID3v1MetadataWithData(id3v1DataSet2),
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
			if gotErr = tt.v1.Write(tt.path); (gotErr != nil) != tt.wantErr {
				t.Errorf("Id3v1Metadata.Write() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
			if gotErr == nil && tt.wantErr == false {
				got, _ := afero.ReadFile(cmd_toolkit.FileSystem(), tt.path)
				if !reflect.DeepEqual(got, tt.wantData) {
					t.Errorf("Id3v1Metadata.Write() got %v want %v", got, tt.wantData)
				}
			}
		})
	}
}

func TestId3v1NameDiffers(t *testing.T) {
	tests := map[string]struct {
		cS   *files.ComparableStrings
		want bool
	}{
		"identical strings": {
			cS: &files.ComparableStrings{
				External: "Fiddler On The Roof",
				Metadata: "Fiddler On The Roof",
			}, want: false,
		},
		"unusable characters in metadata": {
			cS: &files.ComparableStrings{
				External: "Theme From M-A-S-H",
				Metadata: "Theme From M*A*S*H",
			},
			want: false,
		},
		"really long name": {
			cS: &files.ComparableStrings{
				External: "A Funny Thing Happened On The Way To The Forum 1996 Broadway Revival Cast",
				Metadata: "A Funny Thing Happened On The",
			},
			want: false,
		},
		"non-ASCII values": {
			cS: &files.ComparableStrings{
				External: "Grohg - Cortège Macabre",
				Metadata: "Grohg - Cort\xe8ge Macabre",
			},
			want: false,
		},
		"larger non-ASCII values": {
			cS: &files.ComparableStrings{
				External: "Dvořák",
				Metadata: "Dvor\xe1k",
			},
			want: false,
		},
		"identical strings with case differences": {
			cS: &files.ComparableStrings{
				External: "SIMPLE name",
				Metadata: "simple NAME",
			},
			want: false,
		},
		"strings of different length within name length limit": {
			cS: &files.ComparableStrings{
				External: "simple name",
				Metadata: "artist: simple name",
			},
			want: true,
		},
		"use of runes that are illegal for file names": {
			cS: &files.ComparableStrings{
				External: "simple_name",
				Metadata: "simple:name",
			},
			want: false,
		},
		"complex mismatch": {
			cS: &files.ComparableStrings{
				External: "simple_name",
				Metadata: "simple: nam",
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.Id3v1NameDiffers(tt.cS); got != tt.want {
				t.Errorf("Id3v1NameDiffers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestId3v1GenreDiffers(t *testing.T) {
	tests := map[string]struct {
		cS   *files.ComparableStrings
		want bool
	}{
		"match": {
			cS: &files.ComparableStrings{
				External: "Classic Rock",
				Metadata: "Classic Rock",
			},
			want: false,
		},
		"no match": {
			cS: &files.ComparableStrings{
				External: "Classic Rock",
				Metadata: "classic rock",
			},
			want: true,
		},
		"other": {
			cS: &files.ComparableStrings{
				External: "Prog Rock",
				Metadata: "Other",
			},
			want: false,
		},
		"known genre": {
			// known id3v1 genre - "Other" will not match
			cS: &files.ComparableStrings{
				External: "Classic Rock",
				Metadata: "Other",
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.Id3v1GenreDiffers(tt.cS); got != tt.want {
				t.Errorf("Id3v1GenreDiffers() = %v, want %v", got, tt.want)
			}
		})
	}
}
