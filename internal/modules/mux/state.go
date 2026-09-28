package mux

import (
	"io"
)

type State interface {
	io.Closer
	State()
}
