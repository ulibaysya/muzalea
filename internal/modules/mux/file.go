package mux

import "io"

type File struct {
	Name     string
	Playlist bool
	Contents io.Reader
}
