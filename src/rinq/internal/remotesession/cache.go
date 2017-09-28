package remotesession

import (
	"github.com/rinq/rinq-go/src/rinq/ident"
	"github.com/rinq/rinq-go/src/rinq/internal/attrmeta"
)

// attrTableCache is a namespaced local cache of session attributes.
type attrTableCache map[string]attrNamespaceCache

// attrNamespaceCache is an entry in an attrTableCache, representing a single namespace.
type attrNamespaceCache map[string]cachedAttr

// cachedAttr is an entry in a attrNamespaceCache, representing a single attribute.
type cachedAttr struct {
	Attr      attrmeta.Attr
	FetchedAt ident.Revision
}
