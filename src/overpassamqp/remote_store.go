package overpassamqp

import (
	"github.com/over-pass/overpass-go/src/internals"
	"github.com/over-pass/overpass-go/src/overpass"
)

type remoteStore struct {
	// cmdsvc services.CommandService
}

// func (s *remoteStore) Refresh(ctx context.Context, ref overpass.SessionRef) (overpass.Revision, error) {
// 	payload := overpass.NewPayload(ref)
// 	defer payload.Close()
//
// 	response, err := s.cmdsvc.CallInternal(ctx, ref.ID.Peer, "session", "GetSessionRev", payload)
// 	defer response.Close()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	err = response.Decode(&ref)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return s.GetRevision(ref)
// }
//
// func (s *remoteStore) Fetch(ref overpass.SessionRef, keys []string) ([]attrEntry, error) {
// 	payload := overpass.NewPayload(ref)
// 	defer payload.Close()
//
// 	response, err := s.cmdsvc.CallInternal(ctx, ref.ID.Peer, "session", "GetSessionRev", payload)
// 	defer response.Close()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	err = response.Decode(&ref)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return s.GetRevision(ref)
// }
//
// func (s *remoteStore) Update(
// 	ref overpass.SessionRef,
// 	attrs []overpass.Attr,
// ) (overpass.Revision, error) {
// 	return nil, errors.New("not implemented")
// }
//
// func (s *remoteStore) Close(ref overpass.SessionRef) error {
// 	return errors.New("not implemented")
// }

func (s *remoteStore) GetRevision(ref overpass.SessionRef) (overpass.Revision, error) {
	return internals.NewClosedRevision(ref), nil
}
