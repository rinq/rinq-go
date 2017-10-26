package remotesession

import (
	"github.com/rinq/rinq-go/src/rinq"
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/attributes"
	"github.com/rinq/rinq-go/src/rinq/internal/attrmeta"
)

const (
	sessionNamespace = "_sess"
)

const (
	fetchCommand   = "fetch"
	updateCommand  = "update"
	clearCommand   = "clear"
	destroyCommand = "destroy"
)

type fetchRequest struct {
	Seq       uint32   `json:"s"`
	Namespace string   `json:"ns,omitempty"`
	Keys      []string `json:"k,omitempty"`
}

type fetchResponse struct {
	Rev   ident.Revision `json:"r"`
	Attrs attrmeta.List  `json:"a,omitempty"`
}

type updateRequest struct {
	Seq       uint32          `json:"s"`
	Rev       ident.Revision  `json:"r"`
	Namespace string          `json:"ns"`
	Attrs     attributes.List `json:"a,omitempty"` // omitted for "clear" command
}

type updateResponse struct {
	Rev         ident.Revision   `json:"r"`
	CreatedRevs []ident.Revision `json:"cr,omitempty"`
}

type destroyRequest struct {
	Seq uint32         `json:"s"`
	Rev ident.Revision `json:"r"`
}

const (
	notFoundFailure         = "not-found"
	staleUpdateFailure      = "stale"
	frozenAttributesFailure = "frozen"
)

// errorToFailure returns the appropriate failure type based on the type of err.
func errorToFailure(err error) error {
	switch err.(type) {
	case rinq.NotFoundError:
		return rinq.Failure{Type: notFoundFailure}
	case rinq.StaleUpdateError:
		return rinq.Failure{Type: staleUpdateFailure}
	case rinq.FrozenAttributesError:
		return rinq.Failure{Type: frozenAttributesFailure}
	default:
		return err
	}
}

// failureToError returns the appropriate error based on the failure type of err.
func failureToError(ref ident.Ref, err error) error {
	switch rinq.FailureType(err) {
	case notFoundFailure:
		return rinq.NotFoundError{ID: ref.ID}
	case staleUpdateFailure:
		return rinq.StaleUpdateError{Ref: ref}
	case frozenAttributesFailure:
		return rinq.FrozenAttributesError{Ref: ref}
	}

	return err
}
