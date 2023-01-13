package files

import (
	"reflect"
	"testing"
)

func TestCreateID3V1TaggedDataForTesting(t *testing.T) {
	const fnName = "CreateID3V1TaggedDataForTesting()"
	type args struct {
		m map[string]any
	}
	tests := map[string]struct {
		args
		want []byte
	}{
		"full exercise": {
			args: args{
				m: map[string]any{
					"artist": "Artist Name",
					"album":  "Album Name",
					"title":  "Track Title",
					"genre":  "Classic Rock",
					"year":   "2022",
					"track":  2,
				},
			},
			want: []byte{
				'T', 'A', 'G',
				'T', 'r', 'a', 'c', 'k', ' ', 'T', 'i', 't', 'l', 'e', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'A', 'r', 't', 'i', 's', 't', ' ', 'N', 'a', 'm', 'e', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'A', 'l', 'b', 'u', 'm', ' ', 'N', 'a', 'm', 'e', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'2', '0', '2', '2',
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0,
				2,
				1,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := createID3V1TaggedDataForTesting(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestCreateConsistentlyTaggedDataForTesting(t *testing.T) {
	const fnName = "CreateConsistentlyTaggedDataForTesting()"
	type args struct {
		payload []byte
		m       map[string]any
	}
	tests := map[string]struct {
		args
		want []byte
	}{
		"thorough test": {
			args: args{
				payload: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
				m: map[string]any{
					"artist": "Artist Name",
					"album":  "Album Name",
					"title":  "Track Title",
					"genre":  "Classic Rock",
					"year":   "2022",
					"track":  2,
				},
			},
			want: []byte{
				'I', 'D', '3', 3, 0, 0, 0, 0, 0, 115,
				'T', 'A', 'L', 'B', 0, 0, 0, 11, 0, 0, 0, 'A', 'l', 'b', 'u', 'm', ' ', 'N', 'a', 'm', 'e',
				'T', 'C', 'O', 'N', 0, 0, 0, 13, 0, 0, 0, 'C', 'l', 'a', 's', 's', 'i', 'c', ' ', 'R', 'o', 'c', 'k',
				'T', 'I', 'T', '2', 0, 0, 0, 12, 0, 0, 0, 'T', 'r', 'a', 'c', 'k', ' ', 'T', 'i', 't', 'l', 'e',
				'T', 'P', 'E', '1', 0, 0, 0, 12, 0, 0, 0, 'A', 'r', 't', 'i', 's', 't', ' ', 'N', 'a', 'm', 'e',
				'T', 'R', 'C', 'K', 0, 0, 0, 2, 0, 0, 0, '2',
				'T', 'Y', 'E', 'R', 0, 0, 0, 5, 0, 0, 0, '2', '0', '2', '2',
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				'T', 'A', 'G',
				'T', 'r', 'a', 'c', 'k', ' ', 'T', 'i', 't', 'l', 'e', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'A', 'r', 't', 'i', 's', 't', ' ', 'N', 'a', 'm', 'e', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'A', 'l', 'b', 'u', 'm', ' ', 'N', 'a', 'm', 'e', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				'2', '0', '2', '2',
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0,
				2,
				1,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := CreateConsistentlyTaggedDataForTesting(tt.args.payload, tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}
