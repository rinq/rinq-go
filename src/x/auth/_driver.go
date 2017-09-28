package auth

// Driver is really cool
// type Driver interface {
// ContextWithIdentity(ctx context.Context, id interface{}) (context.Context, error)
// IdentityFromContext(ctx context.Context, id interface{}) error
//
// ReadIdentity(io.Reader) (interface{}, error)
// WriteIdentity(io.Writer, interface{}) error
// }

// type authN struct {
// 	next rinq.CommandHandler
// }
//
// func (a *authN) Handle(ctx context.Context, req rinq.Request, res rinq.Response) {
//     context.WithValue(parent context.Context, key interface{}, val interface{})
// }

// type customer struct{}
//
// func requireIdentity(t interface{}, next rinq.CommandHandler) rinq.CommandHandler {
// 	var required = reflect.TypeOf(t)
//
// 	return func(ctx context.Context, req rinq.Request, res rinq.Response) {
// 		var ident interface{}
//
// 		if authn.IdentityFromContext(ctx, ident) {
// 			actual := reflect.TypeOf(ident)
//
// 			if actual.AssignableTo(required) {
// 				next(ctx, req, res)
// 				return
// 			}
// 		}
//
// 		req.Payload.Close()
// 		res.Fail("auth-failure", "")
// 	}
// }
//
// var purchaseSquare = requireIdentity(
// 	customer{},
// 	func (ctx context.Context, req rinq.Request, res rinq.Response) {
// 		cust := GetCustomer(ctx)
// 	}
// )
