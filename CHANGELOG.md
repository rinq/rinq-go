# Changelog

## Next release

Please note that this release includes changes to the definition of AMQP
exchanges. The `ntf.uc` and `ntf.mc` exchanges will need to be deleted on the
broker before starting a peer.

- **[BC]** Add namespaces to session notifications

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
