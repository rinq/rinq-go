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
	closeCommand  = "close"
)

const (
	notFoundFailure         = "not-found"
	staleUpdateFailure      = "stale"
	frozenAttributesFailure = "frozen"
)

type fetchRequest struct {
	Seq  uint32   `json:"s"`
	Keys []string `json:"k,omitempty"`
}

type fetchResponse struct {
	Rev   overpass.RevisionNumber `json:"r"`
	Attrs []attrmeta.Attr         `json:"a,omitempty"`
}

type updateRequest struct {
	Seq   uint32                  `json:"s"`
	Rev   overpass.RevisionNumber `json:"r"`
	Attrs []overpass.Attr         `json:"a,omitempty"`
}

type updateResponse struct {
	Rev         overpass.RevisionNumber   `json:"r"`
	CreatedRevs []overpass.RevisionNumber `json:"cr"`
}

type closeRequest struct {
	Seq uint32                  `json:"s"`
	Rev overpass.RevisionNumber `json:"r"`
}
