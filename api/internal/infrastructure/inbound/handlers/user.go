package handlers

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	apiv1 "github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/gen/api/v1"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/infrastructure/inbound/grpc/gen/api/v1/apiv1connect"
)

type UserService struct {
	apiv1connect.UnimplementedUserServiceHandler
}

func (s *UserService) GetUser(ctx context.Context, req *connect.Request[apiv1.GetUserRequest]) (*connect.Response[apiv1.GetUserResponse], error) {
	fmt.Printf("GetUser called with request: %+v\n", req.Msg)
	fmt.Printf("User ID from request: %s\n", req.Msg.Id)
	// 1. Extract the user ID from the request
	userID := req.Msg.Id

	// 2. Get the user (from database, API, etc.)
	// this will be of type of User which we convert into a proto var.
	user, err := s.getUserFromSomewhere(userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found: %w", err))
	}

	// 3. Convert to protobuf response
	response := &apiv1.GetUserResponse{
		User: &apiv1.User{
			Id:   user.id,
			Name: user.name,
		},
	}

	// 4. Return the response
	return connect.NewResponse(response), nil
}

func (s *UserService) CreateUser(context.Context, *connect.Request[apiv1.CreateUserRequest]) (*connect.Response[apiv1.CreateUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (s *UserService) getUserFromSomewhere(id string) (*User, error) {
	return &User{
		id:   id,
		name: "Bob smith",
	}, nil
}

type User struct {
	id   string
	name string
}
