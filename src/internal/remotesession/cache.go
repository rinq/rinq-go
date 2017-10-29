package remotesession

import (
	"github.com/rinq/rinq-go/src/internal/attributes"
	"github.com/rinq/rinq-go/src/rinq/ident"
)

// attrTableCache is a namespaced local cache of session attributes.
type attrTableCache map[string]attrNamespaceCache

// attrNamespaceCache is an entry in an attrTableCache, representing a single namespace.
type attrNamespaceCache map[string]cachedAttr

// cachedAttr is an entry in a attrNamespaceCache, representing a single attribute.
type cachedAttr struct {
	Attr      attributes.VAttr
	FetchedAt ident.Revision
}
