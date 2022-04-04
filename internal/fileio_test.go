package internal

import (
	"os"
	"testing"
)

func TestMkdir(t *testing.T) {
	fnName := "Mkdir()"
	topDir := "artificalDir"
	defer func() {
		if err := os.RemoveAll(topDir); err != nil {
			t.Errorf("%s error destroying test directory %q: %v", fnName, topDir, err)
		}
	}()
	type args struct {
		dirName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "failure", args: args{dirName: "testutilities_test.go"}, wantErr: true},
		{name: "success", args: args{dirName: topDir}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Mkdir(tt.args.dirName); (err != nil) != tt.wantErr {
				t.Errorf("%q error = %v, wantErr %v", fnName, err, tt.wantErr)
			}
		})
	}
}
