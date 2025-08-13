/*
* This file will handle all /user endpoint requests
* it holds a reference to the userService so it can call on the user service methods
* verfied data will be converted back to the "proto" defined structure and sent back
 */
package handlers

import (
	"context"

	"connectrpc.com/connect"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/application/services"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	errorAdapter "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/adapters/error"
	apiv1 "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/gen/api/v1"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/gen/api/v1/apiv1connect"
	"github.com/google/uuid"
)

type UserServiceHandler struct {
	apiv1connect.UnimplementedUserServiceHandler

	userService *services.UserService
}

func NewUserServiceHandler(userService *services.UserService) *UserServiceHandler {
	return &UserServiceHandler{
		userService: userService,
	}
}

func (s *UserServiceHandler) GetUser(ctx context.Context, req *connect.Request[apiv1.GetUserRequest]) (*connect.Response[apiv1.GetUserResponse], error) {
	id, err := uuid.Parse(req.Msg.GetId())
	if err != nil {
		return nil, errorAdapter.ToConnectError(err)
	}
	user, err := s.userService.Get(ctx, id)
	if err != nil {
		return nil, errorAdapter.ToConnectError(err)
	}

	return connect.NewResponse(&apiv1.GetUserResponse{
		User: userToProto(user),
	}), nil
}

func (s *UserServiceHandler) CreateUser(ctx context.Context, req *connect.Request[apiv1.CreateUserRequest]) (*connect.Response[apiv1.CreateUserResponse], error) {
	user, err := s.userService.Add(ctx, req.Msg.GetName())
	if err != nil {
		return nil, errorAdapter.ToConnectError(err)
	}
	return connect.NewResponse(&apiv1.CreateUserResponse{
		User: userToProto(user),
	}), nil
}

func userToProto(user *domain.User) *apiv1.User {
	if user == nil {
		return nil
	}
	return &apiv1.User{
		Id:   user.ID().String(),
		Name: user.Name(),
	}
}
