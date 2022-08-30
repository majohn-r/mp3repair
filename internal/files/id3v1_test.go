package files

import (
	"fmt"
	"mp3/internal"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func Test_trim(t *testing.T) {
	fnName := "trim()"
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "no trailing data",
			args: args{s: "foo"},
			want: "foo",
		},
		{
			name: "trailing space",
			args: args{s: "foo            "},
			want: "foo",
		},
		{
			name: "trailing nulls",
			args: args{s: string([]byte{'f', 'o', 'o', 0, 0, 0, 0, 0})},
			want: "foo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trim(tt.args.s); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_newId3v1MetadataWithData(t *testing.T) {
	fnName := "newId3v1MetadataWithData()"
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args
		want *id3v1Metadata
	}{
		{
			name: "short data",
			args: args{b: []byte{1, 2, 3, 4}},
			want: &id3v1Metadata{
				data: []byte{
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
		{
			name: "just right",
			args: args{b: internal.ID3V1DataSet1},
			want: &id3v1Metadata{
				data: internal.ID3V1DataSet1,
			},
		},
		{
			name: "too much data",
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
			want: &id3v1Metadata{
				data: []byte{
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newId3v1MetadataWithData(tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_id3v1Metadata_isValid(t *testing.T) {
	fnName := "id3v1Metadata.isValid()"
	tests := []struct {
		name string
		v1   *id3v1Metadata
		want bool
	}{
		{
			name: "expected",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet1),
			want: true,
		},
		{
			name: "bad",
			v1:   newId3v1MetadataWithData([]byte{0, 1, 2}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v1.isValid(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_id3v1Metadata_getTitle(t *testing.T) {
	fnName := "id3v1Metadata.getTitle()"
	tests := []struct {
		name string
		v1   *id3v1Metadata
		want string
	}{
		{
			name: "ringo",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet1),
			want: "Ringo - Pop Profile [Interview",
		},
		{
			name: "julia",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet2),
			want: "Julia",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v1.getTitle(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_id3v1Metadata_setTitle(t *testing.T) {
	fnName := "id3v1Metadata.setTitle()"
	type args struct {
		s string
	}
	tests := []struct {
		name string
		v1   *id3v1Metadata
		args
		want *id3v1Metadata
	}{
		{
			name: "short title",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet1),
			args: args{s: "short title"},
			want: newId3v1MetadataWithData([]byte{
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
		{
			name: "long title",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet1),
			args: args{s: "very long title, so long it cannot be copied intact"},
			want: newId3v1MetadataWithData([]byte{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.v1.setTitle(tt.args.s)
			if !reflect.DeepEqual(tt.v1, tt.want) {
				t.Errorf("%s got %v want %v", fnName, tt.v1, tt.want)
			}
		})
	}
}

func Test_id3v1Metadata_getArtist(t *testing.T) {
	fnName := "id3v1Metadata.getArtist()"
	tests := []struct {
		name string
		v1   *id3v1Metadata
		want string
	}{
		{
			name: "beatles1",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet1),
			want: "The Beatles",
		},
		{
			name: "beatles2",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet2),
			want: "The Beatles",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v1.getArtist(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_id3v1Metadata_setArtist(t *testing.T) {
	fnName := "id3v1Metadata.setArtist()"
	type args struct {
		s string
	}
	tests := []struct {
		name string
		v1   *id3v1Metadata
		args
		want *id3v1Metadata
	}{
		{
			name: "short name",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet1),
			args: args{s: "shorties"},
			want: newId3v1MetadataWithData([]byte{
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
		{
			name: "long name",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet1),
			args: args{s: "The greatest band ever known, bar none"},
			want: newId3v1MetadataWithData([]byte{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.v1.setArtist(tt.args.s)
			if !reflect.DeepEqual(tt.v1, tt.want) {
				t.Errorf("%s got %v want %v", fnName, tt.v1, tt.want)
			}
		})
	}
}

func Test_id3v1Metadata_getAlbum(t *testing.T) {
	fnName := "id3v1Metadata.getAlbum()"
	tests := []struct {
		name string
		v1   *id3v1Metadata
		want string
	}{
		{
			name: "BBC",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet1),
			want: "On Air: Live At The BBC, Volum",
		},
		{
			name: "White Album",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet2),
			want: "The White Album [Disc 1]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v1.getAlbum(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_id3v1Metadata_setAlbum(t *testing.T) {
	fnName := "id3v1Metadata.setAlbum()"
	type args struct {
		s string
	}
	tests := []struct {
		name string
		v1   *id3v1Metadata
		args
		want *id3v1Metadata
	}{
		{
			name: "short name",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet1),
			args: args{s: "!"},
			want: newId3v1MetadataWithData([]byte{
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
		{
			name: "long name",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet1),
			args: args{s: "The Most Amazing Album Ever Released"},
			want: newId3v1MetadataWithData([]byte{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.v1.setAlbum(tt.args.s)
			if !reflect.DeepEqual(tt.v1, tt.want) {
				t.Errorf("%s got %v want %v", fnName, tt.v1, tt.want)
			}
		})
	}
}

func Test_id3v1Metadata_getYear(t *testing.T) {
	fnName := "id3v1Metadata.getYear()"
	tests := []struct {
		name   string
		v1     *id3v1Metadata
		wantY  int
		wantOk bool
	}{
		{
			name:   "BBC",
			v1:     newId3v1MetadataWithData(internal.ID3V1DataSet1),
			wantY:  2013,
			wantOk: true,
		},
		{
			name:   "White Album",
			v1:     newId3v1MetadataWithData(internal.ID3V1DataSet2),
			wantY:  1968,
			wantOk: true,
		},
		{
			name: "no date",
			v1:   newId3v1Metadata(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotY, gotOk := tt.v1.getYear()
			if gotY != tt.wantY {
				t.Errorf("%s gotY = %v, want %v", fnName, gotY, tt.wantY)
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk = %v, want %v", fnName, gotOk, tt.wantOk)
			}
		})
	}
}

func Test_id3v1Metadata_setYear(t *testing.T) {
	fnName := "id3v1Metadata.setYear()"
	type args struct {
		y int
	}
	tests := []struct {
		name string
		v1   *id3v1Metadata
		args
		want   bool
		wantv1 *id3v1Metadata
	}{
		{
			name:   "prehistoric",
			v1:     newId3v1MetadataWithData(internal.ID3V1DataSet1),
			args:   args{y: 999},
			want:   false,
			wantv1: newId3v1MetadataWithData(internal.ID3V1DataSet1),
		},
		{
			name:   "futuristic",
			v1:     newId3v1MetadataWithData(internal.ID3V1DataSet1),
			args:   args{y: 10000},
			want:   false,
			wantv1: newId3v1MetadataWithData(internal.ID3V1DataSet1),
		},
		{
			name: "realistic",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet1),
			args: args{y: 2022},
			want: true,
			wantv1: newId3v1MetadataWithData([]byte{
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v1.setYear(tt.args.y); got != tt.want {
				t.Errorf("%s = %t, want %t", fnName, got, tt.want)
			}
			if !reflect.DeepEqual(tt.v1, tt.wantv1) {
				t.Errorf("%s got %v want %v", fnName, tt.v1, tt.wantv1)
			}
		})
	}
}

func Test_id3v1Metadata_getComment(t *testing.T) {
	fnName := "id3v1Metadata.getComment()"
	tests := []struct {
		name string
		v1   *id3v1Metadata
		want string
	}{
		{
			name: "BBC",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet1),
		},
		{
			name: "White Album",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet2),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v1.getComment(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_id3v1Metadata_setComment(t *testing.T) {
	fnName := "id3v1Metadata.setComment()"
	type args struct {
		s string
	}
	tests := []struct {
		name string
		v1   *id3v1Metadata
		args
		want *id3v1Metadata
	}{
		{
			name: "typical comment",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet1),
			args: args{s: ""},
			want: newId3v1MetadataWithData([]byte{
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
		{
			name: "long winded",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet1),
			args: args{s: "This track is genuinely insightful"},
			want: newId3v1MetadataWithData([]byte{
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.v1.setComment(tt.args.s)
			if !reflect.DeepEqual(tt.v1, tt.want) {
				t.Errorf("%s got %v want %v", fnName, tt.v1, tt.want)
			}
		})
	}
}

func Test_id3v1Metadata_getTrack(t *testing.T) {
	fnName := "id3v1Metadata.getTrack()"
	tests := []struct {
		name   string
		v1     *id3v1Metadata
		wantI  int
		wantOk bool
	}{
		{
			name:   "BBC",
			v1:     newId3v1MetadataWithData(internal.ID3V1DataSet1),
			wantI:  29,
			wantOk: true,
		},
		{
			name:   "White Album",
			v1:     newId3v1MetadataWithData(internal.ID3V1DataSet2),
			wantI:  17,
			wantOk: true,
		},
		{
			name: "bad zero byte",
			v1: newId3v1MetadataWithData([]byte{
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotI, gotOk := tt.v1.getTrack()
			if gotI != tt.wantI {
				t.Errorf("%s gotI = %v, want %v", fnName, gotI, tt.wantI)
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk = %v, want %v", fnName, gotOk, tt.wantOk)
			}
		})
	}
}

func Test_id3v1Metadata_setTrack(t *testing.T) {
	fnName := "id3v1Metadata.setTrack()"
	type args struct {
		t int
	}
	tests := []struct {
		name string
		v1   *id3v1Metadata
		args
		want   bool
		wantv1 *id3v1Metadata
	}{
		{
			name:   "low",
			v1:     newId3v1MetadataWithData(internal.ID3V1DataSet1),
			args:   args{t: 0},
			want:   false,
			wantv1: newId3v1MetadataWithData(internal.ID3V1DataSet1),
		},
		{
			name:   "high",
			v1:     newId3v1MetadataWithData(internal.ID3V1DataSet1),
			args:   args{t: 256},
			want:   false,
			wantv1: newId3v1MetadataWithData(internal.ID3V1DataSet1),
		},
		{
			name: "ok",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet1),
			args: args{t: 45},
			want: true,
			wantv1: newId3v1MetadataWithData([]byte{
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v1.setTrack(tt.args.t); got != tt.want {
				t.Errorf("%s = %t, want %t", fnName, got, tt.want)
			}
			if !reflect.DeepEqual(tt.v1, tt.wantv1) {
				t.Errorf("%s got %v want %v", fnName, tt.v1, tt.wantv1)
			}
		})
	}
}

func Test_id3v1Metadata_getGenre(t *testing.T) {
	fnName := "id3v1Metadata.getGenre()"
	tests := []struct {
		name   string
		v1     *id3v1Metadata
		wantS  string
		wantOk bool
	}{
		{
			name:   "BBC",
			v1:     newId3v1MetadataWithData(internal.ID3V1DataSet1),
			wantS:  "Other",
			wantOk: true,
		},
		{
			name:   "White Album",
			v1:     newId3v1MetadataWithData(internal.ID3V1DataSet2),
			wantS:  "Rock",
			wantOk: true,
		},
		{
			name: "bad zero byte",
			v1: newId3v1MetadataWithData([]byte{
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotS, gotOk := tt.v1.getGenre()
			if gotS != tt.wantS {
				t.Errorf("%s gotS = %v, want %v", fnName, gotS, tt.wantS)
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk = %v, want %v", fnName, gotOk, tt.wantOk)
			}
		})
	}
}

func Test_id3v1Metadata_setGenre(t *testing.T) {
	fnName := "id3v1Metadata.setGenre()"
	type args struct {
		s string
	}
	tests := []struct {
		name string
		v1   *id3v1Metadata
		args
		want   bool
		wantv1 *id3v1Metadata
	}{
		{
			name:   "no such genre",
			v1:     newId3v1MetadataWithData(internal.ID3V1DataSet1),
			args:   args{s: "Subspace Radio"},
			want:   false,
			wantv1: newId3v1MetadataWithData(internal.ID3V1DataSet1),
		},
		{
			name: "known genre",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet1),
			args: args{s: genreMap[37]},
			want: true,
			wantv1: newId3v1MetadataWithData([]byte{
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v1.setGenre(tt.args.s); got != tt.want {
				t.Errorf("%s = %t, want %t", fnName, got, tt.want)
			}
			if !reflect.DeepEqual(tt.v1, tt.wantv1) {
				t.Errorf("%s got %v want %v", fnName, tt.v1, tt.wantv1)
			}
		})
	}
}

func Test_initGenreIndices(t *testing.T) {
	fnName := "initGenreIndices()"
	initGenreIndices()
	if len(genreIndicesMap) != len(genreMap) {
		t.Errorf("%s size of genreIndicesMap is %d, genreMap is %d", fnName, len(genreIndicesMap), len(genreMap))
	} else {
		for k, v := range genreMap {
			if k2 := genreIndicesMap[strings.ToLower(v)]; k2 != k {
				t.Errorf("%s index for %q got %d want %d", fnName, v, k2, k)
			}
		}
	}
}

func Test_internalReadId3V1Metadata(t *testing.T) {
	fnName := "internalReadId3V1Metadata()"
	testDir := "id3v1read"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	shortFile := "short.mp3"
	if err := internal.CreateFileForTestingWithContent(testDir, shortFile, []byte{0, 1, 2}); err != nil {
		t.Errorf("%s error creating %q: %v", testDir, shortFile, err)
	}
	badFile := "bad.mp3"
	if err := internal.CreateFileForTestingWithContent(testDir, badFile, []byte{
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
	payload = append(payload, internal.ID3V1DataSet1...)
	if err := internal.CreateFileForTestingWithContent(testDir, goodFile, payload); err != nil {
		t.Errorf("%s error creating %q: %v", testDir, goodFile, err)
	}
	type args struct {
		path     string
		readFunc func(f *os.File, b []byte) (int, error)
	}
	tests := []struct {
		name string
		args
		want    *id3v1Metadata
		wantErr bool
	}{
		{
			name: "non-existent file",
			args: args{
				path:     "./non-existent",
				readFunc: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "short file",
			args: args{
				path:     filepath.Join(testDir, shortFile),
				readFunc: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "read with error",
			args: args{
				path: filepath.Join(testDir, badFile),
				readFunc: func(f *os.File, b []byte) (int, error) {
					return 0, fmt.Errorf("oops")
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "short read",
			args: args{
				path: filepath.Join(testDir, badFile),
				readFunc: func(f *os.File, b []byte) (int, error) {
					return 127, nil
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "bad file",
			args: args{
				path:     filepath.Join(testDir, badFile),
				readFunc: readFromFile,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "good file",
			args: args{
				path:     filepath.Join(testDir, goodFile),
				readFunc: readFromFile,
			},
			want:    newId3v1MetadataWithData(internal.ID3V1DataSet1),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := internalReadId3V1Metadata(tt.args.path, tt.args.readFunc)
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
	fnName := "readId3v1Metadata()"
	testDir := "id3v1read"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
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
	payload = append(payload, internal.ID3V1DataSet1...)
	if err := internal.CreateFileForTestingWithContent(testDir, goodFile, payload); err != nil {
		t.Errorf("%s error creating %q: %v", testDir, goodFile, err)
	}
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args
		want    []string
		wantErr bool
	}{
		// only testing good path ... all the error paths are handled in the
		// internal read test
		{
			name:    "good file",
			args:    args{path: filepath.Join(testDir, goodFile)},
			want:    []string{
				"Artist: \"The Beatles\"",
				"Album: \"On Air: Live At The BBC, Volum\"",
				"Title: \"Ringo - Pop Profile [Interview\"",
				"Track: 29",
				"Year: 2013",
				"Genre: \"Other\"",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readId3v1Metadata(tt.args.path)
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

func Test_id3v1Metadata_internalWrite(t *testing.T) {
	fnName := "id3v1Metadata.internalWrite()"
	testDir := "id3v1write"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	shortFile := "short.mp3"
	if err := internal.CreateFileForTestingWithContent(testDir, shortFile, []byte{0, 1, 2}); err != nil {
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
	payload = append(payload, internal.ID3V1DataSet1...)
	if err := internal.CreateFileForTestingWithContent(testDir, goodFile, payload); err != nil {
		t.Errorf("%s error creating %q: %v", testDir, goodFile, err)
	}
	type args struct {
		oldPath   string
		writeFunc func(f *os.File, b []byte) (int, error)
	}
	tests := []struct {
		name string
		v1   *id3v1Metadata
		args
		wantErr  bool
		wantData []byte
	}{
		{
			name:    "non-existent file",
			args:    args{oldPath: "./no such file"},
			wantErr: true,
		},
		{
			name:    "short file",
			args:    args{oldPath: filepath.Join(testDir, shortFile)},
			wantErr: true,
		},
		{
			name: "error on write",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet1),
			args: args{
				oldPath: filepath.Join(testDir, goodFile),
				writeFunc: func(f *os.File, b []byte) (int, error) {
					return 0, fmt.Errorf("ruh-roh!")
				},
			},
			wantErr: true,
		},
		{
			name: "short write",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet1),
			args: args{
				oldPath: filepath.Join(testDir, goodFile),
				writeFunc: func(f *os.File, b []byte) (int, error) {
					return 127, nil
				},
			},
			wantErr: true,
		},
		{
			name: "good write",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet2),
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if err = tt.v1.internalWrite(tt.args.oldPath, tt.args.writeFunc); (err != nil) != tt.wantErr {
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

func Test_id3v1Metadata_write(t *testing.T) {
	fnName := "id3v1Metadata.write()"
	testDir := "id3v1write"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
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
	payload = append(payload, internal.ID3V1DataSet1...)
	if err := internal.CreateFileForTestingWithContent(testDir, goodFile, payload); err != nil {
		t.Errorf("%s error creating %q: %v", testDir, goodFile, err)
	}
	type args struct {
		path string
	}
	tests := []struct {
		name string
		v1   *id3v1Metadata
		args
		wantErr  bool
		wantData []byte
	}{
		{
			name: "happy place",
			v1:   newId3v1MetadataWithData(internal.ID3V1DataSet2),
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if err = tt.v1.write(tt.args.path); (err != nil) != tt.wantErr {
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
