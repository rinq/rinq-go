package remotesession

import "github.com/over-pass/overpass-go/src/overpass"

func logCacheAdd(
	logger overpass.Logger,
	peerID overpass.PeerID,
	sessID overpass.SessionID,
) {
	logger.Log(
		"%s discovered remote session %s ",
		peerID.ShortString(),
		sessID.ShortString(),
	)
}

func logCacheMark(
	logger overpass.Logger,
	peerID overpass.PeerID,
	sessID overpass.SessionID,
) {
	logger.Log(
		"%s marked remote session %s for removal from the store",
		peerID.ShortString(),
		sessID.ShortString(),
	)
}

func logCacheRemove(
	logger overpass.Logger,
	peerID overpass.PeerID,
	sessID overpass.SessionID,
) {
	logger.Log(
		"%s removed remote session %s from the store",
		peerID.ShortString(),
		sessID.ShortString(),
	)
}
