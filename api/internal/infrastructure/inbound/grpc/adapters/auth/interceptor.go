package auth

/*
this file is used to intercept all requests, extract the user from the token,
and refresh the token if needed. Done by modifying context and the req struct
*/
import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/utils"
)

// codes describing something about the request itself are safe to return verbatim;
// anything else can carry infrastructure detail in its message (a dropped db
// connection, a redis timeout) and gets collapsed to a generic error instead
var clientFacingCodes = map[connect.Code]bool{
	connect.CodeUnauthenticated:  true,
	connect.CodePermissionDenied: true,
	connect.CodeInvalidArgument:  true,
	connect.CodeNotFound:         true,
	connect.CodeAlreadyExists:    true,
}

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
					var connectErr *connect.Error
					if errors.As(err, &connectErr) && clientFacingCodes[connectErr.Code()] {
						return nil, err
					}
					// server-caused failure — keep the real detail in logs, not on the wire
					utils.Logger.Error("authenticated request failed", "procedure", req.Spec().Procedure, "error", err)
					return nil, connect.NewError(connect.CodeInternal, errors.New("internal error"))
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
