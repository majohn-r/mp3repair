package files_test

import (
	"fmt"
	"mp3/internal/files"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
)

var (
	// id3v1DataSet1 is a sample ID3V1 tag from an existing file
	id3v1DataSet1 = []byte{
		'T', 'A', 'G',
		'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r', 'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i', 'e', 'w',
		'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A', 't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l', 'u', 'm',
		'2', '0', '1', '3',
		' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		0,
		29,
		12,
	}
	// id3v1DataSet2 is a sample ID3V1 tag from an existing file
	id3v1DataSet2 = []byte{
		'T', 'A', 'G',
		'J', 'u', 'l', 'i', 'a', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		'T', 'h', 'e', ' ', 'W', 'h', 'i', 't', 'e', ' ', 'A', 'l', 'b', 'u', 'm', ' ', '[', 'D', 'i', 's', 'c', ' ', '1', ']', ' ', ' ', ' ', ' ', ' ', ' ',
		'1', '9', '6', '8',
		' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		0,
		17,
		17,
	}
)

func Test_Trim(t *testing.T) {
	const fnName = "Trim()"
	type args struct {
		s string
	}
	tests := map[string]struct {
		args
		want string
	}{
		"no trailing data": {args: args{s: "foo"}, want: "foo"},
		"trailing space":   {args: args{s: "foo            "}, want: "foo"},
		"trailing nulls":   {args: args{s: string([]byte{'f', 'o', 'o', 0, 0, 0, 0, 0})}, want: "foo"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.Trim(tt.args.s); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func NewID3v1MetadataWithData(b []byte) *files.Id3v1Metadata {
	v1 := files.NewID3v1Metadata()
	if len(b) >= files.Id3v1Length {
		for k := 0; k < files.Id3v1Length; k++ {
			v1.Data[k] = b[k]
		}
	} else {
		copy(v1.Data, b)
		for k := len(b); k < files.Id3v1Length; k++ {
			v1.Data[k] = 0
		}
	}
	return v1
}

func Test_newId3v1MetadataWithData(t *testing.T) {
	const fnName = "newId3v1MetadataWithData()"
	type args struct {
		b []byte
	}
	tests := map[string]struct {
		args
		want *files.Id3v1Metadata
	}{
		"short data": {
			args: args{b: []byte{1, 2, 3, 4}},
			want: &files.Id3v1Metadata{
				Data: []byte{
					1, 2, 3, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				},
			},
		},
		"just right": {
			args: args{b: id3v1DataSet1},
			want: &files.Id3v1Metadata{
				Data: id3v1DataSet1,
			},
		},
		"too much data": {
			args: args{b: []byte{
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				10, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				20, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				30, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				40, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				50, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				60, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				70, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				80, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
			}},
			want: &files.Id3v1Metadata{
				Data: []byte{
					0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
					10, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
					20, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
					30, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
					40, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
					50, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
					60, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
					70, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := NewID3v1MetadataWithData(tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestId3v1Metadata_isValid(t *testing.T) {
	const fnName = "Id3v1Metadata.isValid()"
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
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestId3v1Metadata_getTitle(t *testing.T) {
	const fnName = "Id3v1Metadata.getTitle()"
	tests := map[string]struct {
		v1   *files.Id3v1Metadata
		want string
	}{
		"ringo": {v1: NewID3v1MetadataWithData(id3v1DataSet1), want: "Ringo - Pop Profile [Interview"},
		"julia": {v1: NewID3v1MetadataWithData(id3v1DataSet2), want: "Julia"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.v1.Title(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestId3v1Metadata_setTitle(t *testing.T) {
	const fnName = "Id3v1Metadata.setTitle()"
	type args struct {
		s string
	}
	tests := map[string]struct {
		v1 *files.Id3v1Metadata
		args
		want *files.Id3v1Metadata
	}{
		"short title": {
			v1:   NewID3v1MetadataWithData(id3v1DataSet1),
			args: args{s: "short title"},
			want: NewID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				's', 'h', 'o', 'r', 't', ' ', 't', 'i', 't', 'l', 'e', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A', 't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l', 'u', 'm',
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				12,
			}),
		},
		"long title": {
			v1:   NewID3v1MetadataWithData(id3v1DataSet1),
			args: args{s: "very long title, so long it cannot be copied intact"},
			want: NewID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g', ' ', 't', 'i', 't', 'l', 'e', ',', ' ', 's', 'o', ' ', 'l', 'o', 'n', 'g', ' ', 'i', 't', ' ', 'c', 'a',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A', 't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l', 'u', 'm',
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				12,
			}),
		},
		"non-ASCII title": {
			v1:   NewID3v1MetadataWithData(id3v1DataSet1),
			args: args{s: "Grohg - Cortège Macabre"},
			want: NewID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'G', 'r', 'o', 'h', 'g', ' ', '-', ' ', 'C', 'o', 'r', 't', 0xE8, 'g', 'e', ' ', 'M', 'a', 'c', 'a', 'b', 'r', 'e', 0, 0, 0, 0, 0, 0, 0,
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A', 't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l', 'u', 'm',
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				12,
			}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.v1.SetTitle(tt.args.s)
			if !reflect.DeepEqual(tt.v1, tt.want) {
				t.Errorf("%s got %v want %v", fnName, tt.v1, tt.want)
			}
		})
	}
}

func Test_Id3v1Metadata_getArtist(t *testing.T) {
	const fnName = "Id3v1Metadata.getArtist()"
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
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestId3v1Metadata_setArtist(t *testing.T) {
	const fnName = "Id3v1Metadata.setArtist()"
	type args struct {
		s string
	}
	tests := map[string]struct {
		v1 *files.Id3v1Metadata
		args
		want *files.Id3v1Metadata
	}{
		"short name": {
			v1:   NewID3v1MetadataWithData(id3v1DataSet1),
			args: args{s: "shorties"},
			want: NewID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r', 'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i', 'e', 'w',
				's', 'h', 'o', 'r', 't', 'i', 'e', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A', 't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l', 'u', 'm',
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				12,
			}),
		},
		"long name": {
			v1:   NewID3v1MetadataWithData(id3v1DataSet1),
			args: args{s: "The greatest band ever known, bar none"},
			want: NewID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r', 'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i', 'e', 'w',
				'T', 'h', 'e', ' ', 'g', 'r', 'e', 'a', 't', 'e', 's', 't', ' ', 'b', 'a', 'n', 'd', ' ', 'e', 'v', 'e', 'r', ' ', 'k', 'n', 'o', 'w', 'n', ',', ' ',
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A', 't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l', 'u', 'm',
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				12,
			}),
		},
		"non-ASCII name": {
			v1:   NewID3v1MetadataWithData(id3v1DataSet1),
			args: args{s: "Antonín Dvořák"},
			want: NewID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r', 'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i', 'e', 'w',
				'A', 'n', 't', 'o', 'n', 0xED, 'n', ' ', 'D', 'v', 'o', 'r', 0xE1, 'k', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A', 't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l', 'u', 'm',
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				12,
			}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.v1.SetArtist(tt.args.s)
			if !reflect.DeepEqual(tt.v1, tt.want) {
				t.Errorf("%s got %v want %v", fnName, tt.v1, tt.want)
			}
		})
	}
}

func TestId3v1Metadata_getAlbum(t *testing.T) {
	const fnName = "Id3v1Metadata.getAlbum()"
	tests := map[string]struct {
		v1   *files.Id3v1Metadata
		want string
	}{
		"BBC":         {v1: NewID3v1MetadataWithData(id3v1DataSet1), want: "On Air: Live At The BBC, Volum"},
		"White Album": {v1: NewID3v1MetadataWithData(id3v1DataSet2), want: "The White Album [Disc 1]"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.v1.Album(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestId3v1MetadataSetAlbum(t *testing.T) {
	const fnName = "Id3v1Metadata.SetAlbum()"
	type args struct {
		s string
	}
	tests := map[string]struct {
		v1 *files.Id3v1Metadata
		args
		want *files.Id3v1Metadata
	}{
		"short name": {
			v1:   NewID3v1MetadataWithData(id3v1DataSet1),
			args: args{s: "!"},
			want: NewID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r', 'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i', 'e', 'w',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'!', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				12,
			}),
		},
		"long name": {
			v1:   NewID3v1MetadataWithData(id3v1DataSet1),
			args: args{s: "The Most Amazing Album Ever Released"},
			want: NewID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r', 'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i', 'e', 'w',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'T', 'h', 'e', ' ', 'M', 'o', 's', 't', ' ', 'A', 'm', 'a', 'z', 'i', 'n', 'g', ' ', 'A', 'l', 'b', 'u', 'm', ' ', 'E', 'v', 'e', 'r', ' ', 'R', 'e',
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				12,
			}),
		},
		"non-ASCII name": {
			v1:   NewID3v1MetadataWithData(id3v1DataSet1),
			args: args{s: "Déjà Vu"},
			want: NewID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r', 'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i', 'e', 'w',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'D', 0xE9, 'j', 0xE0, ' ', 'V', 'u', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				12,
			}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.v1.SetAlbum(tt.args.s)
			if !reflect.DeepEqual(tt.v1, tt.want) {
				t.Errorf("%s got %v want %v", fnName, tt.v1, tt.want)
			}
		})
	}
}

func TestId3v1Metadata_getYear(t *testing.T) {
	const fnName = "Id3v1Metadata.getYear()"
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
				t.Errorf("%s gotY = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestId3v1MetadataSetYear(t *testing.T) {
	const fnName = "Id3v1Metadata.SetYear()"
	type args struct {
		s string
	}
	tests := map[string]struct {
		v1 *files.Id3v1Metadata
		args
		wantv1 *files.Id3v1Metadata
	}{
		"realistic": {
			v1:   NewID3v1MetadataWithData(id3v1DataSet1),
			args: args{s: "2022"},
			wantv1: NewID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r', 'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i', 'e', 'w',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A', 't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l', 'u', 'm',
				'2', '0', '2', '2',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				12,
			}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.v1.SetYear(tt.args.s)
			if !reflect.DeepEqual(tt.v1, tt.wantv1) {
				t.Errorf("%s got %v want %v", fnName, tt.v1, tt.wantv1)
			}
		})
	}
}

func TestId3v1Metadata_getComment(t *testing.T) {
	const fnName = "Id3v1Metadata.getComment()"
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
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestId3v1MetadataSetComment(t *testing.T) {
	const fnName = "Id3v1Metadata.SetComment()"
	type args struct {
		s string
	}
	tests := map[string]struct {
		v1 *files.Id3v1Metadata
		args
		want *files.Id3v1Metadata
	}{
		"typical comment": {
			v1:   NewID3v1MetadataWithData(id3v1DataSet1),
			args: args{s: ""},
			want: NewID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r', 'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i', 'e', 'w',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A', 't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l', 'u', 'm',
				'2', '0', '1', '3',
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0,
				29,
				12,
			}),
		},
		"long winded": {
			v1:   NewID3v1MetadataWithData(id3v1DataSet1),
			args: args{s: "This track is genuinely insightful"},
			want: NewID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r', 'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i', 'e', 'w',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A', 't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l', 'u', 'm',
				'2', '0', '1', '3',
				'T', 'h', 'i', 's', ' ', 't', 'r', 'a', 'c', 'k', ' ', 'i', 's', ' ', 'g', 'e', 'n', 'u', 'i', 'n', 'e', 'l', 'y', ' ', 'i', 'n', 's', 'i',
				0,
				29,
				12,
			}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.v1.SetComment(tt.args.s)
			if !reflect.DeepEqual(tt.v1, tt.want) {
				t.Errorf("%s got %v want %v", fnName, tt.v1, tt.want)
			}
		})
	}
}

func TestId3v1Metadata_getTrack(t *testing.T) {
	const fnName = "Id3v1Metadata.getTrack()"
	tests := map[string]struct {
		v1     *files.Id3v1Metadata
		wantI  int
		wantOk bool
	}{
		"BBC":         {v1: NewID3v1MetadataWithData(id3v1DataSet1), wantI: 29, wantOk: true},
		"White Album": {v1: NewID3v1MetadataWithData(id3v1DataSet2), wantI: 17, wantOk: true},
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
				t.Errorf("%s gotI = %v, want %v", fnName, gotI, tt.wantI)
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk = %v, want %v", fnName, gotOk, tt.wantOk)
			}
		})
	}
}

func TestId3v1Metadata_setTrack(t *testing.T) {
	const fnName = "Id3v1Metadata.setTrack()"
	type args struct {
		t int
	}
	tests := map[string]struct {
		v1 *files.Id3v1Metadata
		args
		want   bool
		wantv1 *files.Id3v1Metadata
	}{
		"low":  {v1: NewID3v1MetadataWithData(id3v1DataSet1), args: args{t: 0}, want: false, wantv1: NewID3v1MetadataWithData(id3v1DataSet1)},
		"high": {v1: NewID3v1MetadataWithData(id3v1DataSet1), args: args{t: 256}, want: false, wantv1: NewID3v1MetadataWithData(id3v1DataSet1)},
		"ok": {
			v1:   NewID3v1MetadataWithData(id3v1DataSet1),
			args: args{t: 45},
			want: true,
			wantv1: NewID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r', 'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i', 'e', 'w',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A', 't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l', 'u', 'm',
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				45,
				12,
			}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.v1.SetTrack(tt.args.t); got != tt.want {
				t.Errorf("%s = %t, want %t", fnName, got, tt.want)
			}
			if !reflect.DeepEqual(tt.v1, tt.wantv1) {
				t.Errorf("%s got %v want %v", fnName, tt.v1, tt.wantv1)
			}
		})
	}
}

func TestId3v1Metadata_getGenre(t *testing.T) {
	const fnName = "Id3v1Metadata.getGenre()"
	tests := map[string]struct {
		v1     *files.Id3v1Metadata
		wantS  string
		wantOk bool
	}{
		"BBC":         {v1: NewID3v1MetadataWithData(id3v1DataSet1), wantS: "Other", wantOk: true},
		"White Album": {v1: NewID3v1MetadataWithData(id3v1DataSet2), wantS: "Rock", wantOk: true},
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
				t.Errorf("%s gotS = %v, want %v", fnName, gotS, tt.wantS)
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk = %v, want %v", fnName, gotOk, tt.wantOk)
			}
		})
	}
}

func TestId3v1Metadata_setGenre(t *testing.T) {
	const fnName = "Id3v1Metadata.setGenre()"
	type args struct {
		s string
	}
	tests := map[string]struct {
		v1 *files.Id3v1Metadata
		args
		wantv1 *files.Id3v1Metadata
	}{
		"no such genre": {
			v1:   NewID3v1MetadataWithData(id3v1DataSet1),
			args: args{s: "Subspace Radio"},
			wantv1: NewID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r', 'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i', 'e', 'w',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A', 't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l', 'u', 'm',
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				12,
			}),
		},
		"known genre": {
			v1:   NewID3v1MetadataWithData(id3v1DataSet1),
			args: args{s: files.GenreMap[37]},
			wantv1: NewID3v1MetadataWithData([]byte{
				'T', 'A', 'G',
				'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r', 'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i', 'e', 'w',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A', 't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l', 'u', 'm',
				'2', '0', '1', '3',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				29,
				37,
			}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.v1.SetGenre(tt.args.s)
			if !reflect.DeepEqual(tt.v1, tt.wantv1) {
				t.Errorf("%s got %v want %v", fnName, tt.v1, tt.wantv1)
			}
		})
	}
}

func TestInitGenreIndices(t *testing.T) {
	const fnName = "InitGenreIndices()"
	files.InitGenreIndices()
	if len(files.GenreIndicesMap) != len(files.GenreMap) {
		t.Errorf("%s size of genreIndicesMap is %d, genreMap is %d", fnName, len(files.GenreIndicesMap), len(files.GenreMap))
	} else {
		for k, v := range files.GenreMap {
			if k2 := files.GenreIndicesMap[strings.ToLower(v)]; k2 != k {
				t.Errorf("%s index for %q got %d want %d", fnName, v, k2, k)
			}
		}
	}
}

func TestIntTypeernalReadId3V1Metadata(t *testing.T) {
	const fnName = "internalReadId3V1Metadata()"
	testDir := "id3v1read"
	if err := cmd_toolkit.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	defer func() {
		destroyDirectory(fnName, testDir)
	}()
	shortFile := "short.mp3"
	if err := createFileWithContent(testDir, shortFile, []byte{0, 1, 2}); err != nil {
		t.Errorf("%s error creating %q: %v", testDir, shortFile, err)
	}
	badFile := "bad.mp3"
	if err := createFileWithContent(testDir, badFile, []byte{
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
	}); err != nil {
		t.Errorf("%s error creating %q: %v", testDir, badFile, err)
	}
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
	if err := createFileWithContent(testDir, goodFile, payload); err != nil {
		t.Errorf("%s error creating %q: %v", testDir, goodFile, err)
	}
	type args struct {
		path     string
		readFunc func(f *os.File, b []byte) (int, error)
	}
	tests := map[string]struct {
		args
		want    *files.Id3v1Metadata
		wantErr bool
	}{
		"non-existent file": {args: args{path: "./non-existent", readFunc: nil}, want: nil, wantErr: true},
		"short file":        {args: args{path: filepath.Join(testDir, shortFile), readFunc: nil}, want: nil, wantErr: true},
		"read with error": {
			args: args{
				path: filepath.Join(testDir, badFile),
				readFunc: func(f *os.File, b []byte) (int, error) {
					return 0, fmt.Errorf("oops")
				},
			},
			want:    nil,
			wantErr: true,
		},
		"short read": {
			args: args{
				path: filepath.Join(testDir, badFile),
				readFunc: func(f *os.File, b []byte) (int, error) {
					return 127, nil
				},
			},
			want:    nil,
			wantErr: true,
		},
		"bad file":  {args: args{path: filepath.Join(testDir, badFile), readFunc: files.FileReader}, want: nil, wantErr: true},
		"good file": {args: args{path: filepath.Join(testDir, goodFile), readFunc: files.FileReader}, want: NewID3v1MetadataWithData(id3v1DataSet1), wantErr: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := files.InternalReadID3V1Metadata(tt.args.path, tt.args.readFunc)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_readId3v1Metadata(t *testing.T) {
	const fnName = "readId3v1Metadata()"
	testDir := "id3v1read"
	if err := cmd_toolkit.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	defer func() {
		destroyDirectory(fnName, testDir)
	}()
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
	if err := createFileWithContent(testDir, goodFile, payload); err != nil {
		t.Errorf("%s error creating %q: %v", testDir, goodFile, err)
	}
	type args struct {
		path string
	}
	tests := map[string]struct {
		args
		want    []string
		wantErr bool
	}{
		// only testing good path ... all the error paths are handled in the
		// internal read test
		"good file": {
			args: args{path: filepath.Join(testDir, goodFile)},
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
			got, err := files.ReadID3v1Metadata(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestId3v1MetadataIntTypeernalWrite(t *testing.T) {
	const fnName = "Id3v1Metadata.internalWrite()"
	testDir := "id3v1write"
	if err := cmd_toolkit.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	defer func() {
		destroyDirectory(fnName, testDir)
	}()
	shortFile := "short.mp3"
	if err := createFileWithContent(testDir, shortFile, []byte{0, 1, 2}); err != nil {
		t.Errorf("%s error creating %q: %v", testDir, shortFile, err)
	}
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
	if err := createFileWithContent(testDir, goodFile, payload); err != nil {
		t.Errorf("%s error creating %q: %v", testDir, goodFile, err)
	}
	type args struct {
		oldPath   string
		writeFunc func(f *os.File, b []byte) (int, error)
	}
	tests := map[string]struct {
		v1 *files.Id3v1Metadata
		args
		wantErr  bool
		wantData []byte
	}{
		"non-existent file": {args: args{oldPath: "./no such file"}, wantErr: true},
		"short file":        {args: args{oldPath: filepath.Join(testDir, shortFile)}, wantErr: true},
		"error on write": {
			v1: NewID3v1MetadataWithData(id3v1DataSet1),
			args: args{
				oldPath: filepath.Join(testDir, goodFile),
				writeFunc: func(f *os.File, b []byte) (int, error) {
					return 0, fmt.Errorf("ruh-roh")
				},
			},
			wantErr: true,
		},
		"short write": {
			v1: NewID3v1MetadataWithData(id3v1DataSet1),
			args: args{
				oldPath: filepath.Join(testDir, goodFile),
				writeFunc: func(f *os.File, b []byte) (int, error) {
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
				'J', 'u', 'l', 'i', 'a', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				'T', 'h', 'e', ' ', 'W', 'h', 'i', 't', 'e', ' ', 'A', 'l', 'b', 'u', 'm', ' ', '[', 'D', 'i', 's', 'c', ' ', '1', ']', ' ', ' ', ' ', ' ', ' ', ' ',
				'1', '9', '6', '8',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				17,
				17,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var err error
			if err = tt.v1.InternalWrite(tt.args.oldPath, tt.args.writeFunc); (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
			}
			if err == nil && tt.wantErr == false {
				got, _ := os.ReadFile(tt.args.oldPath)
				if !reflect.DeepEqual(got, tt.wantData) {
					t.Errorf("%s got %v want %v", fnName, got, tt.wantData)
				}
			}
		})
	}
}

func TestId3v1Metadata_write(t *testing.T) {
	const fnName = "Id3v1Metadata.write()"
	testDir := "id3v1write"
	if err := cmd_toolkit.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	defer func() {
		destroyDirectory(fnName, testDir)
	}()
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
	if err := createFileWithContent(testDir, goodFile, payload); err != nil {
		t.Errorf("%s error creating %q: %v", testDir, goodFile, err)
	}
	type args struct {
		path string
	}
	tests := map[string]struct {
		v1 *files.Id3v1Metadata
		args
		wantErr  bool
		wantData []byte
	}{
		"happy place": {
			v1: NewID3v1MetadataWithData(id3v1DataSet2),
			args: args{
				path: filepath.Join(testDir, goodFile),
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
				'J', 'u', 'l', 'i', 'a', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				'T', 'h', 'e', ' ', 'W', 'h', 'i', 't', 'e', ' ', 'A', 'l', 'b', 'u', 'm', ' ', '[', 'D', 'i', 's', 'c', ' ', '1', ']', ' ', ' ', ' ', ' ', ' ', ' ',
				'1', '9', '6', '8',
				' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
				0,
				17,
				17,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var err error
			if err = tt.v1.Write(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
			}
			if err == nil && tt.wantErr == false {
				got, _ := os.ReadFile(tt.args.path)
				if !reflect.DeepEqual(got, tt.wantData) {
					t.Errorf("%s got %v want %v", fnName, got, tt.wantData)
				}
			}
		})
	}
}

func TestId3v1NameDiffers(t *testing.T) {
	const fnName = "Id3v1NameDiffers()"
	type args struct {
		cS files.ComparableStrings
	}
	tests := map[string]struct {
		args
		want bool
	}{
		"identical strings": {
			args: args{files.ComparableStrings{ExternalName: "Fiddler On The Roof", MetadataName: "Fiddler On The Roof"}},
			want: false,
		},
		"unusable characters in metadata": {
			args: args{files.ComparableStrings{ExternalName: "Theme From M-A-S-H", MetadataName: "Theme From M*A*S*H"}},
			want: false,
		},
		"really long name": {
			args: args{files.ComparableStrings{
				ExternalName: "A Funny Thing Happened On The Way To The Forum 1996 Broadway Revival Cast",
				MetadataName: "A Funny Thing Happened On The",
			}},
			want: false,
		},
		"non-ASCII values": {
			args: args{files.ComparableStrings{ExternalName: "Grohg - Cortège Macabre", MetadataName: "Grohg - Cort\xe8ge Macabre"}},
			want: false,
		},
		"larger non-ASCII values": {
			args: args{files.ComparableStrings{ExternalName: "Dvořák", MetadataName: "Dvor\xe1k"}},
			want: false,
		},
		"identical strings with case differences": {
			args: args{files.ComparableStrings{ExternalName: "SIMPLE name", MetadataName: "simple NAME"}},
			want: false,
		},
		"strings of different length within name length limit": {
			args: args{files.ComparableStrings{ExternalName: "simple name", MetadataName: "artist: simple name"}},
			want: true,
		},
		"use of runes that are illegal for file names": {args: args{files.ComparableStrings{ExternalName: "simple_name", MetadataName: "simple:name"}}, want: false},
		"complex mismatch": {args: args{files.ComparableStrings{ExternalName: "simple_name", MetadataName: "simple: nam"}}, want: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.Id3v1NameDiffers(tt.args.cS); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestId3v1GenreDiffers(t *testing.T) {
	const fnName = "Id3v1GenreDiffers()"
	type args struct {
		cS files.ComparableStrings
	}
	tests := map[string]struct {
		args
		want bool
	}{
		"match":    {args: args{cS: files.ComparableStrings{ExternalName: "Classic Rock", MetadataName: "Classic Rock"}}, want: false},
		"no match": {args: args{cS: files.ComparableStrings{ExternalName: "Classic Rock", MetadataName: "classic rock"}}, want: true},
		"other":    {args: args{cS: files.ComparableStrings{ExternalName: "Prog Rock", MetadataName: "Other"}}, want: false},
		"known genre": {
			args: args{
				cS: files.ComparableStrings{
					ExternalName: "Classic Rock", // known id3v1 genre - "Other" will not match
					MetadataName: "Other",
				},
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := files.Id3v1GenreDiffers(tt.args.cS); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}
