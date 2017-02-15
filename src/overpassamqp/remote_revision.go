package overpassamqp

// type remoteRevision struct {
// 	session *remoteSession
// 	ref     overpass.SessionRef
// 	attrs   attrTable
// }
//
// func (r *remoteRevision) Ref() overpass.SessionRef {
// 	return r.ref
// }
//
// func (r *remoteRevision) Refresh(ctx context.Context) (overpass.Revision, error) {
// 	return r.session.CurrentRevision(ctx)
// }
//
// func (r *remoteRevision) Get(ctx context.Context, key string) (overpass.Attr, error) {
// 	entry, ok := r.attrs[key]
//
// 	// The attribute doesn't exist at all, or was created after this revision.
// 	if !ok || entry.CreatedAt > r.ref.Rev {
// 		return overpass.Attr{Key: key}, nil
// 	}
//
// 	// The attribute exists, but has been updated since this revision. The value
// 	// at this revision is no longer known.
// 	if entry.UpdatedAt > r.ref.Rev {
// 		return overpass.Attr{}, overpass.StaleFetchError{Ref: r.ref}
// 	}
//
// 	return entry.Attr, nil
// }
//
// func (r *remoteRevision) GetMany(ctx context.Context, keys ...string) (overpass.AttrTable, error) {
// 	attrs := overpass.AttrTable{}
//
// 	for _, key := range keys {
// 		entry, ok := r.attrs[key]
//
// 		// The attribute doesn't exist at all, or was created after this revision.
// 		if !ok || entry.CreatedAt > r.ref.Rev {
// 			continue
// 		}
//
// 		// The attribute exists, but has been updated since this revision. The
// 		// value at this revision is no longer known.
// 		if entry.UpdatedAt > r.ref.Rev {
// 			return nil, overpass.StaleFetchError{Ref: r.ref}
// 		}
//
// 		if entry.Attr.Value != "" && entry.Attr.IsFrozen {
// 			attrs[key] = entry.Attr
// 		}
// 	}
//
// 	return attrs, nil
// }
//
// func (r *remoteRevision) Update(ctx context.Context, attrs ...overpass.Attr) (overpass.Revision, error) {
// 	return r, overpass.StaleUpdateError{Ref: r.ref}
// }
//
// func (r *remoteRevision) Close(ctx context.Context) error {
// 	return overpass.StaleUpdateError{Ref: r.ref}
// }
