/*
* Client Request → Interceptor → Original Handler → Response back through Interceptor → Client
 */
package logging

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"connectrpc.com/connect"
)

// TODO pass in env config (local, dev, prod) and add slog.warn based on env
func LoggerInterceptor() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// pre handler logic
			start := time.Now()

			// call handler
			resp, err := next(ctx, req)

			// post handler logic
			duration := time.Since(start)
			if err != nil {
				var connectErr *connect.Error
				var code connect.Code
				if errors.As(err, &connectErr) {
					code = connectErr.Code()
				} else {
					code = connect.CodeUnknown
				}

				slog.Error(
					"req failed",
					"procedure", req.Spec().Procedure,
					"req", fmt.Sprintf("%+v", req.Any()),
					"code", code,
					"error", err,
					"duration", duration.Milliseconds(),
				)
				return nil, err
			}

			slog.Info(
				"req succeeded",
				"procedure", req.Spec().Procedure,
				"req", fmt.Sprintf("%+v", req.Any()),
				"resp", fmt.Sprintf("%+v", resp.Any()),
				"duration", duration.Milliseconds(),
			)

			return resp, err
		})
	}
	return connect.UnaryInterceptorFunc(interceptor)
}
