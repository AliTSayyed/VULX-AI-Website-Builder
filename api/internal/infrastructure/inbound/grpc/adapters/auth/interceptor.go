package auth

/*
this file is used to intercept all requests, extract the user from the token,
and refresh the token if needed. Done by modifying context and the req struct
*/
import (
	"context"

	"connectrpc.com/connect"
)

func (h *HTTPAuthAdapter) HTTPAuthInterceptor() connect.UnaryInterceptorFunc {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			// pre handler call, if user is nil then no user is added to context, this is fine for public routes, will get errors for protected routes
			if user, refresh := h.AuthenticateWithJWT(ctx, req); user != nil {

				// this is the actual handler call with user passed into context
				res, err := next(context.WithValue(ctx, UserContextKey{}, user), req)
				// post handler, modify response with new token if needed
				if err != nil {
					return nil, connect.NewError(connect.CodeInvalidArgument, err)
				}
				if refresh {
					h.RefreshJWTCookie(ctx, res, user.ID())
				}

				return res, nil
			}
			// this calls the handler if no auth is needed
			return next(ctx, req)
		})
	})
}
