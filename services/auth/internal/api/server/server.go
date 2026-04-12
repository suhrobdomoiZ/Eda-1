package api

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonpb "github.com/suhrobdomoiZ/Eda-1/pkg/api/common"
	pb "github.com/suhrobdomoiZ/Eda-1/services/auth/internal/api/gen"
	service "github.com/suhrobdomoiZ/Eda-1/services/auth/internal/services"
)

type Server struct {
	pb.UnimplementedAuthServiceServer
	svc *service.AuthService
}

func NewServer(svc *service.AuthService) *Server {
	return &Server{svc: svc}
}

func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password required")
	}

	role, err := protoRoleToString(req.Role)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	in := service.RegisterInput{
		Username: req.Username,
		Password: req.Password,
		Role:     role,
	}

	// Заполняем профиль из oneof
	switch p := req.Profile.(type) {
	case *pb.RegisterRequest_Restaurant:
		in.RestaurantName = p.Restaurant.Name
		in.RestaurantAddress = p.Restaurant.Address
		in.RestaurantPhone = p.Restaurant.Phone
	case *pb.RegisterRequest_Courier:
		in.CourierName = p.Courier.Name
		in.CourierPhone = p.Courier.Phone
	}

	result, err := s.svc.Register(ctx, in)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "username already taken")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.RegisterResponse{
		UserId: result.UserID,
		Tokens: &pb.TokenPair{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
		},
	}, nil
}

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password required")
	}

	result, err := s.svc.Login(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.LoginResponse{
		UserId: result.UserID,
		Role:   stringRoleToProto(result.Role),
		Tokens: &pb.TokenPair{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
		},
	}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	claims, err := s.svc.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid: true,
		Claims: &pb.TokenClaims{
			UserId: claims.UserID,
			Role:   stringRoleToProto(claims.Role),
			Exp:    claims.ExpiresAt.Unix(),
		},
	}, nil
}

func (s *Server) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	result, err := s.svc.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "refresh token invalid or expired")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.RefreshTokenResponse{
		Tokens: &pb.TokenPair{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
		},
	}, nil
}

func (s *Server) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.GetProfileResponse, error) {
	user, rp, cp, err := s.svc.GetProfile(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	resp := &pb.GetProfileResponse{
		User: &commonpb.UserProfile{
			Id:       user.ID,
			Username: user.Username,
			Role:     stringRoleToProto(user.Role),
		},
	}

	if rp != nil {
		resp.Extended = &pb.GetProfileResponse_Restaurant{
			Restaurant: &commonpb.RestaurantProfile{
				Id:      rp.UserID,
				Name:    rp.Name,
				Address: rp.Address,
				Phone:   rp.Phone,
			},
		}
	} else if cp != nil {
		resp.Extended = &pb.GetProfileResponse_Courier{
			Courier: &commonpb.CourierProfile{
				Id:    cp.UserID,
				Name:  cp.Name,
				Phone: cp.Phone,
			},
		}
	}

	return resp, nil
}

func (s *Server) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if err := s.svc.Logout(ctx, req.RefreshToken); err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &pb.LogoutResponse{Success: true}, nil
}

// Для конвертации ролей
func protoRoleToString(r commonpb.UserRole) (string, error) {
	switch r {
	case commonpb.UserRole_USER_ROLE_CUSTOMER:
		return "user", nil
	case commonpb.UserRole_USER_ROLE_RESTAURANT:
		return "restaurant", nil
	case commonpb.UserRole_USER_ROLE_COURIER:
		return "courier", nil
	case commonpb.UserRole_USER_ROLE_ADMIN:
		return "admin", nil
	default:
		return "", errors.New("unknown role")
	}
}

func stringRoleToProto(r string) commonpb.UserRole {
	switch r {
	case "restaurant":
		return commonpb.UserRole_USER_ROLE_RESTAURANT
	case "courier":
		return commonpb.UserRole_USER_ROLE_COURIER
	case "admin":
		return commonpb.UserRole_USER_ROLE_ADMIN
	default:
		return commonpb.UserRole_USER_ROLE_CUSTOMER
	}
}
