package localsession

import (
	"github.com/over-pass/overpass-go/src/overpass"
	"github.com/over-pass/overpass-go/src/overpass/internal/attrmeta"
	"github.com/over-pass/overpass-go/src/overpass/internal/bufferpool"
)

func logSessionClose(
	logger overpass.Logger,
	cat Catalog,
	traceID string,
) {
	ref, attrs := cat.Attrs()

	buffer := bufferpool.Get()
	defer bufferpool.Put(buffer)
	attrmeta.WriteTable(buffer, attrs)

	if traceID == "" {
		logger.Log(
			"%s session destroyed {%s}",
			ref.ShortString(),
			buffer,
		)
	} else {
		logger.Log(
			"%s session destroyed {%s} [%s]",
			ref.ShortString(),
			buffer,
			traceID,
		)
	}
}
