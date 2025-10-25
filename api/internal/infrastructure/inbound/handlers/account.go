package handlers

import (
	"context"

	"connectrpc.com/connect"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/application/services"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	authAdapter "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/adapters/auth"
	errorAdapter "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/adapters/error"
	apiv1 "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/gen/api/v1"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/gen/api/v1/apiv1connect"
)

type AccountServiceHandler struct {
	apiv1connect.AccountServiceHandler

	accountService *services.AccountService
	authAdapter    *authAdapter.HTTPAuthAdapter
}

func NewAccountServiceHandler(accountService *services.AccountService, authAdapter *authAdapter.HTTPAuthAdapter) *AccountServiceHandler {
	return &AccountServiceHandler{
		accountService: accountService,
		authAdapter:    authAdapter,
	}
}

func (a *AccountServiceHandler) BeginAccountAuth(ctx context.Context, req *connect.Request[apiv1.BeginAccountAuthRequest]) (*connect.Response[apiv1.BeginAccountAuthResponse], error) {
	provider := loginProviderToDomain(req.Msg.GetLoginProvider())

	loginUrl, err := a.accountService.BeginAuth(ctx, provider)
	if err != nil {
		return nil, errorAdapter.ToConnectError(err)
	}

	return connect.NewResponse(&apiv1.BeginAccountAuthResponse{
		LoginUrl: loginUrl,
	}), nil
}

func (a *AccountServiceHandler) FinishAccountAuth(ctx context.Context, req *connect.Request[apiv1.FinishAccountAuthRequest]) (*connect.Response[apiv1.FinishAccountAuthResponse], error) {
	authResult, err := a.accountService.FinishAuth(ctx, req.Msg.GetCode(), req.Msg.GetState())
	if err != nil {
		return nil, errorAdapter.ToConnectError(err)
	}

	res := connect.NewResponse(&apiv1.FinishAccountAuthResponse{
		Profile: profileToProto(authResult.Profile),
	})
	a.authAdapter.SetJWTCookie(ctx, res, authResult.Token)

	return res, nil
}

func (a *AccountServiceHandler) AccountLogout(ctx context.Context, req *connect.Request[apiv1.AccountLogoutRequest]) (*connect.Response[apiv1.AccountLogoutResponse], error) {
	token := a.authAdapter.GetJWTCookie(ctx, req)
	if token == "" {
		return connect.NewResponse(&apiv1.AccountLogoutResponse{}), nil
	}

	if err := a.accountService.Logout(ctx, token); err != nil {
		return nil, errorAdapter.ToConnectError(err)
	}

	res := connect.NewResponse(&apiv1.AccountLogoutResponse{})
	a.authAdapter.ClearJWTCookie(ctx, res)

	return res, nil
}

func (a *AccountServiceHandler) GetUserProfile(ctx context.Context, req *connect.Request[apiv1.GetUserProfileRequest]) (*connect.Response[apiv1.GetUserProfileResponse], error) {
	user, err := authAdapter.User(ctx)
	if err != nil {
		return nil, err
	}

	profile, err := a.accountService.GetProfile(ctx, user.ID())
	if err != nil {
		return nil, errorAdapter.ToConnectError(err)
	}

	return connect.NewResponse(&apiv1.GetUserProfileResponse{
		Profile: profileToProto(profile),
	}), nil
}

func loginProviderToDomain(provider apiv1.LoginProvider) domain.LoginProvider {
	switch provider {
	case apiv1.LoginProvider_LOGIN_PROVIDER_GOOGLE:
		return domain.LoginProviderGoogle
	case apiv1.LoginProvider_LOGIN_PROVIDER_UNSPECIFIED:
		fallthrough
	default:
		return domain.LoginProviderUnspecified
	}
}

func profileToProto(profile *domain.Profile) *apiv1.Profile {
	return &apiv1.Profile{
		FirstName: profile.FirstName(),
		LastName:  profile.LastName(),
		Email:     profile.Email(),
		Credits:   int64(profile.Credits()),
	}
}
