package service

import (
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/rahul4019/tasker/internal/server"
)

type AuthService struct {
	server *server.Server
}

func NewAuthService(s *server.Server) *AuthService {
	clerk.SetKey(s.Config.Auth.SecretKey)
	return &AuthService{
		server: s,
	}
}
