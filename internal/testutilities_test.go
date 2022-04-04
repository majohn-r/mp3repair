package internal

import (
	"testing"
)

func TestCreateAlbumNameForTesting(t *testing.T) {
	fnName := "CreateAlbumNameForTesting()"
	type args struct {
		k int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "negative value", args: args{k: -1}, want: "Test Album -1"},
		{name: "zero", args: args{k: 0}, want: "Test Album 0"},
		{name: "positive value", args: args{k: 1}, want: "Test Album 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateAlbumNameForTesting(tt.args.k); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestCreateArtistNameForTesting(t *testing.T) {
	fnName := "CreateArtistNameForTesting()"
	type args struct {
		k int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "negative value", args: args{k: -1}, want: "Test Artist -1"},
		{name: "zero", args: args{k: 0}, want: "Test Artist 0"},
		{name: "positive value", args: args{k: 1}, want: "Test Artist 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateArtistNameForTesting(tt.args.k); got != tt.want {
				t.Errorf("%q = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestCreateTrackNameForTesting(t *testing.T) {
	fnName := "CreateTrackNameForTesting()"
	type args struct {
		k int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "zero", args: args{k: 0}, want: "00-Test Track[00].mp3"},
		{name: "odd positive value", args: args{k: 1}, want: "01 Test Track[01].mp3"},
		{name: "even positive value", args: args{k: 2}, want: "02-Test Track[02].mp3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateTrackNameForTesting(tt.args.k); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}
