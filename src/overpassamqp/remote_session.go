package overpassamqp

// import (
// 	"context"
// 	"log"
// 	"sync"
//
// 	"github.com/over-pass/overpass-go/src/amqputil"
// 	"github.com/over-pass/overpass-go/src/overpass"
// 	"github.com/over-pass/overpass-go/src/overpassamqp/services"
// )
//
// // remoteSession represents a session owned by a peer running in another process.
// type remoteSession struct {
// 	id     overpass.SessionID
// 	logger *log.Logger
//
// 	mutex  sync.RWMutex
// 	latest *remoteRevision
// }
//
// func newRemoteSession(
// 	id overpass.SessionID,
// 	cmdsvc services.CommandService,
// 	ntfsvc services.NotifyService,
// 	logger *log.Logger,
// ) *remoteSession {
// 	sess := &remoteSession{
// 		id:     id,
// 		cmdsvc: cmdsvc,
// 		ntfsvc: ntfsvc,
// 		logger: logger,
// 	}
// 	// sess.current = &remoteRevision{
// 	// 	session: sess,
// 	// 	ref:     id.At(0),
// 	// }
//
// 	return sess
// }
//
// func (s *remoteSession) ID() overpass.SessionID {
// 	return s.id
// }
//
// func (s *remoteSession) CurrentRevision() (overpass.Revision, error) {
// 	s.mutex.RLock()
// 	defer s.mutex.RUnlock()
//
// 	if s.current == nil {
// 		return nil, overpass.NotFoundError{ID: s.id}
// 	}
//
// 	return s.current, nil
// }
//
// func (s *remoteSession) ApplyUpdate(ctx context.Context, next *localRevision) error {
// 	s.mutex.Lock()
// 	defer s.mutex.Unlock()
//
// 	if s.current == nil {
// 		return overpass.NotFoundError{ID: s.id}
// 	}
//
// 	expected := s.current.Ref()
// 	expected.Rev++
//
// 	if next.Ref() != expected {
// 		return overpass.StaleUpdateError{Ref: next.Ref()}
// 	}
//
// 	s.current = next
// 	s.messageSeq = 0
//
// 	if corrID := amqputil.GetCorrelationID(ctx); corrID != "" {
// 		s.logger.Printf(
// 			"%s session updated {%s} [%s]",
// 			s.current.Ref().ShortString(),
// 			s.current,
// 			corrID,
// 		)
// 	} else {
// 		s.logger.Printf(
// 			"%s session updated {%s}",
// 			s.current.Ref().ShortString(),
// 			s.current,
// 		)
// 	}
//
// 	return nil
// }
//
// func (s *remoteSession) ApplyClose(ctx context.Context, ref overpass.SessionRef) error {
// 	s.mutex.Lock()
// 	defer s.mutex.Unlock()
//
// 	if s.current == nil {
// 		return nil
// 	}
//
// 	if ref != s.current.Ref() {
// 		return overpass.StaleUpdateError{Ref: ref}
// 	}
//
// 	s.destroy(ctx)
//
// 	return nil
// }
//
// func (s *remoteSession) nextMessageID() (msgID overpass.MessageID, err error) {
// 	s.mutex.RLock()
// 	defer s.mutex.RUnlock()
//
// 	if s.current == nil {
// 		err = overpass.NotFoundError{ID: s.id}
// 	} else {
// 		s.messageSeq++
// 		msgID = overpass.MessageID{
// 			Session: s.current.Ref(),
// 			Seq:     s.messageSeq,
// 		}
// 	}
//
// 	return
// }
//
// // destroy the session. It is assumed that s.mutex is already acquired for writing.
// func (s *remoteSession) destroy(ctx context.Context) {
// 	ref := s.current.Ref()
// 	s.current = nil
// 	close(s.done)
//
// 	s.ntfsvc.Unlisten(s.id) // TODO: log error, probably
//
// 	if corrID := amqputil.GetCorrelationID(ctx); corrID != "" {
// 		s.logger.Printf(
// 			"%s session destroyed [%s]",
// 			ref.ShortString(),
// 			corrID,
// 		)
// 	} else {
// 		s.logger.Printf(
// 			"%s session destroyed",
// 			ref.ShortString(),
// 		)
// 	}
// }
