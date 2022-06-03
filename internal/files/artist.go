package files

import (
	"io/fs"
	"path/filepath"
)

type Artist struct {
	Name   string
	Albums []*Album
	Path   string
}

func newArtist(f fs.FileInfo, dir string) *Artist {
	name := f.Name()
	return &Artist{Name: name, Path: filepath.Join(dir, name)}
}

func copyArtist(a *Artist) *Artist {
	return &Artist{Name: a.Name, Path: a.Path}
}
