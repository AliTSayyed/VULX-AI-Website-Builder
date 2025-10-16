/*
* This file will handle all /user endpoint requests
* it holds a reference to the userService so it can call on the user service methods
* verfied data will be converted back to the "proto" defined structure and sent back
 */
package handlers

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/application/services"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	authAdapter "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/adapters/auth"
	errorAdapter "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/adapters/error"
	apiv1 "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/gen/api/v1"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/gen/api/v1/apiv1connect"
	"github.com/google/uuid"
)

type UserServiceHandler struct {
	apiv1connect.UnimplementedUserServiceHandler

	userService *services.UserService
	authAdapter *authAdapter.HTTPAuthAdapter
}

func NewUserServiceHandler(userService *services.UserService, authAdapter *authAdapter.HTTPAuthAdapter) *UserServiceHandler {
	return &UserServiceHandler{
		userService: userService,
		authAdapter: authAdapter,
	}
}

func (u *UserServiceHandler) GetUser(ctx context.Context, req *connect.Request[apiv1.GetUserRequest]) (*connect.Response[apiv1.GetUserResponse], error) {
	user, err := authAdapter.User(ctx)
	if err != nil {
		return nil, err
	}

	if user.Email() != "alitsayyed@gmail.com" {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("only the super user can access this endpoint"))
	}

	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, errorAdapter.ToConnectError(err)
	}

	foundUser, err := u.userService.Get(ctx, id)
	if err != nil {
		return nil, errorAdapter.ToConnectError(err)
	}

	return connect.NewResponse(&apiv1.GetUserResponse{
		User: userToProto(foundUser),
	}), nil
}

// func (u *UserServiceHandler) ListUsers(ctx context.Context, req *connect.Request[apiv1.ListUsersRequest]) (*connect.Response[apiv1.ListUsersResponse], error) {
// 	user, err := authAdapter.User(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if user.Email() != "alitsayyed@gmail.com" {
// 		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("only the super user can access this endpoint"))
// 	}
// }

func (u *UserServiceHandler) CreateUser(ctx context.Context, req *connect.Request[apiv1.CreateUserRequest]) (*connect.Response[apiv1.CreateUserResponse], error) {
	user, err := authAdapter.User(ctx)
	if err != nil {
		return nil, err
	}

	if user.Email() != "alitsayyed@gmail.com" {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("only the super user can access this endpoint"))
	}

	createdUser, err := u.userService.Add(ctx, req.Msg.GetFirstName(), req.Msg.GetLastName(), req.Msg.GetEmail())
	if err != nil {
		return nil, errorAdapter.ToConnectError(err)
	}
	return connect.NewResponse(&apiv1.CreateUserResponse{
		User: userToProto(createdUser),
	}), nil
}

func userToProto(user *domain.User) *apiv1.User {
	if user == nil {
		return nil
	}
	return &apiv1.User{
		Id:        user.ID().String(),
		FirstName: user.FirstName(),
		LastName:  user.LastName(),
		Email:     user.Email(),
		Credits:   int64(user.Credits()),
		IsActive:  user.IsActive(),
	}
}
