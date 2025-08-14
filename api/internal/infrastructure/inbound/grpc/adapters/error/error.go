/*
* this error package will convert domain errors
* into the correct connect rpc error for the handler
 */
package grpcerror

import (
	"errors"

	"connectrpc.com/connect"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
)

func ToConnectError(err error) error {
	if err == nil {
		return nil
	}

	var domainError *domain.Error
	if errors.As(err, &domainError) {
		switch domainError.Type() {
		case domain.ErrorTypeUnauthenticated:
			return connect.NewError(connect.CodeUnauthenticated, err)
		case domain.ErrorTypePermissionDenied:
			return connect.NewError(connect.CodePermissionDenied, err)
		case domain.ErrorTypeInvalid:
			return connect.NewError(connect.CodeInvalidArgument, err)
		case domain.ErrorTypeNotFound:
			return connect.NewError(connect.CodeNotFound, err)
		case domain.ErrorTypeAlreadyExists:
			return connect.NewError(connect.CodeAlreadyExists, err)
		case domain.ErrorTypeInternal:
			return connect.NewError(connect.CodeInternal, err)
		case domain.ErrorTypeUnimplemented:
			return connect.NewError(connect.CodeUnimplemented, err)
		case domain.ErrorTypeUnavailable:
			return connect.NewError(connect.CodeUnavailable, err)
		case domain.ErrorTypeTimeout:
			return connect.NewError(connect.CodeDeadlineExceeded, err)
		default:
			return connect.NewError(connect.CodeUnknown, err)
		}
	}

	return connect.NewError(connect.CodeUnknown, err)
}
