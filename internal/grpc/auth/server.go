package auth

import (
	"awesomeProject/internal/services/auth"
	"context"
	"errors"
	ssov1 "github.com/Cato149/protos/proto/gen/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)
import "google.golang.org/grpc"

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID int64,
	) (string, error)
	Register(
		ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type serverAPI struct {
	// this impl needs to compile app without NotImplementedError
	ssov1.UnimplementedAuthServer
	auth Auth
}

func RegisterServerAPI(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

const (
	emptyValue = 0
)

func (s *serverAPI) Login(
	ctx context.Context,
	req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if err := s.validateLogin(req); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetAppId())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "Internal Error")
	}

	return &ssov1.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if err := s.validateRegister(req); err != nil {
		return nil, err
	}

	userID, err := s.auth.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserExist) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, "Internal Error")
	}

	return &ssov1.RegisterResponse{UserId: userID}, nil
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	if err := s.validateIsAdmin(req); err != nil {
		return nil, err
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, "Internal Error")
	}
	return &ssov1.IsAdminResponse{IsAdmin: isAdmin}, nil
}

func (s *serverAPI) validateLogin(req *ssov1.LoginRequest) error {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "email and password is required")
	}

	if req.GetAppId() == emptyValue {
		return status.Error(codes.InvalidArgument, "app_id is required")
	}

	return nil
}

func (s *serverAPI) validateRegister(req *ssov1.RegisterRequest) error {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "email and password is required")
	}

	return nil
}

func (s *serverAPI) validateIsAdmin(req *ssov1.IsAdminRequest) error {
	if req.GetUserId() == emptyValue {
		return status.Error(codes.InvalidArgument, "UserID is required")
	}

	return nil
}
