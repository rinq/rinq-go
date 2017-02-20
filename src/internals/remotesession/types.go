package remotesession

import (
	"github.com/over-pass/overpass-go/src/internals/attrmeta"
	"github.com/over-pass/overpass-go/src/overpass"
)

const (
	sessionNamespace = "_sess"
)

const (
	fetchCommand  = "fetch"
	updateCommand = "update"
)

const (
	notFoundFailure         = "not-found"
	staleUpdateFailure      = "stale"
	frozenAttributesFailure = "frozen"
)

type fetchRequest struct {
	Seq  uint32
	Keys []string
}

type fetchResponse struct {
	Rev   overpass.RevisionNumber
	Attrs []attrmeta.Attr
}

type updateRequest struct {
	Seq   uint32
	Rev   overpass.RevisionNumber
	Attrs []overpass.Attr
}

type updateResponse overpass.RevisionNumber
