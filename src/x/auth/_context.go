package authn

import (
	"context"
	"reflect"
)

type keyType string

const (
	identityKey keyType = "identity"
)

// IdentityFromContext reads an identity from ctx, into id.
// It panics if id can not be set.
// It returns true only if the identity in ctx is assigned to id. It returns
// false if there is no identity in ctx, or the identity is not assignable to id.
func IdentityFromContext(ctx context.Context, id interface{}) bool {
	target := reflect.ValueOf(id)

	if !target.CanSet() {
		panic("id can not be set, try passing a pointer")
	}

	ident := ctx.Value(identityKey)

	if ident == nil {
		return false
	}

	source := reflect.ValueOf(ident)

	if !source.Type().AssignableTo(target.Type()) {
		return false
	}

	target.Set(source)

	return true
}

// WithIdentity returns a new context containing the identity id, which can be
// extracted with IdentityFromContext()
func WithIdentity(ctx context.Context, id interface{}) context.Context {
	if id == nil {
		return ctx
	}

	return context.WithValue(ctx, identityKey, id)
}
