package remotesession

import (
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/attrmeta"
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
	Seq       uint32   `json:"s"`
	Namespace string   `json:"ns,omitempty"`
	Keys      []string `json:"k,omitempty"`
}

type fetchResponse struct {
	Rev   ident.Revision  `json:"r"`
	Attrs []attrmeta.Attr `json:"a,omitempty"`
}

type updateRequest struct {
	Seq       uint32         `json:"s"`
	Rev       ident.Revision `json:"r"`
	Namespace string         `json:"ns"`
	Attrs     []rinq.Attr    `json:"a,omitempty"`
}

type updateResponse struct {
	Rev         ident.Revision   `json:"r"`
	CreatedRevs []ident.Revision `json:"cr"`
}

type closeRequest struct {
	Seq uint32         `json:"s"`
	Rev ident.Revision `json:"r"`
}
