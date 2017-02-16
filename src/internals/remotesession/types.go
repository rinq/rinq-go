package remotesession

import (
	"github.com/over-pass/overpass-go/src/internals/attrmeta"
	"github.com/over-pass/overpass-go/src/overpass"
)

const (
	sessionNamespace = "_sess"
	fetchCommand     = "f"
	notFoundFailure  = "nf"
)

type fetchRequest struct {
	Seq  uint32
	Keys []string
}

type fetchResponse struct {
	Rev   overpass.RevisionNumber
	Attrs []attrmeta.Attr
}
