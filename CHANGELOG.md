# Changelog

## 0.6.0 (2017-10-27)

- **[BC]** Rename the `amqp` package to `rinqamqp`
- **[BC]** Removed `rinq.ValidateNamespace()`
- **[BC]** Panic, rather than return an error, when a "programmer error" is made, such as using an invalid namespace
- **[BC]** Change `rinq.AttrTable` from a map to an interface
- **[BC]** Expand notification constraints to a fully-fledged expression system
- **[NEW]** Add `Revision.Clear()` which sets all attributes in a single namespace to the empty string
- **[NEW]** Add message IDs to `rinq.Request` and `rinq.Notification`
- **[NEW]** Add `trace.WithRoot()` which adds a custom trace ID to a context only if there is no existing trace ID
- **[NEW]** Add `ident.MustValidate()`
- **[FIX]** Payload now enocdes `nil` maps and slices as empty objects and arrays, respectively, instead of `null`

## 0.5.0 (2017-10-02)

- **[BC]** Remove `Config` in favour of "functional options" in the `options` package
- **[BC]** Session attributes are now namespaced
- **[NEW]** Add support for [OpenTracing](https://opentracing.io) via new `options.Tracer()` option

## 0.4.0 (2017-09-18)

Please note that this release includes changes to the definition of AMQP
exchanges. The `ntf.uc` and `ntf.mc` exchanges will need to be deleted on the
broker before starting a peer.

- **[BC]** Remove `Session.ExecuteMany()`
- **[BC]** Add namespaces to session notifications
- **[BC]** Rename `Config.CommandPreFetch` to `CommandWorkers`
- **[BC]** Rename `Config.SessionPreFetch` to `SessionWorkers`
- **[FIX]** Fix race-condition caused by payload buffer "double-free"
- **[NEW]** Add `amqp.DialEnv()` which connects to an AMQP Rinq network described by environment variables
- **[NEW]** Add `NewConfigFromEnv()` which returns a Rinq configuration described by environment variables
- **[NEW]** Add `Config.Product` which is passed to the broker in the AMQP handshake
- **[FIX]** Honour the context deadline when dialing an AMQP broker
- **[IMPROVED]** `Session.NotifyMany()` and `NotifyMany()` now return `context.Canceled` when the peer is stopping

## 0.3.0 (2017-04-07)

- **[BC]** `AsyncResponseHandler` is now passed the session
- **[NEW]** `Session.Execute[Many]` now supports context deadlines
- **[NEW]** Promote `trace` module to public API
- **[FIX]** Allow empty messages in failure responses
- **[FIX]** Fix panic when when stopping a peer repeatedly

## 0.2.0 (2017-03-04)

- **[BC]** Rename from "Overpass" to "Rinq"
- **[BC]** Require Go 1.8
- **[BC]** Rename `Session.Close()` and `Revision.Close()` to `Destroy()`
- **[BC]** Move AMQP implementation into `amqp` sub-package
- **[BC]** Move identifier types into `ident` sub-package
- **[BC]** Renamed `Command` to `Request`
- **[BC]** Renamed `Responder` to `Response`
- **[BC]** `Peer.Stop()` and `GracefulStop()` no longer block
- **[NEW]** Add `Session.CallAsync()`
- **[IMPROVED]** AMQP broker capabilities are checked on connect
- **[IMPROVED]** `Response.Fail()` accepts sprintf-style format specifier
- **[IMPROVED]** Log all payload values when debug logging is enabled

## 0.1.0 (2017-02-24)

- Initial release
