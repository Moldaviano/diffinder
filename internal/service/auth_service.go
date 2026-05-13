package service

import (
	"context"

	"github.com/alloy/diffinder/internal/auth"
	"github.com/alloy/diffinder/internal/httpx"
	"github.com/alloy/diffinder/internal/model"
	"github.com/alloy/diffinder/internal/repository"
)

type AuthService struct {
	users  *repository.UsersRepo
	issuer *auth.Issuer
}

func NewAuthService(users *repository.UsersRepo, issuer *auth.Issuer) *AuthService {
	return &AuthService{users: users, issuer: issuer}
}

type TokenPair struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         *model.User `json:"user"`
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, httpx.ErrUnauthorized("invalid credentials")
		}
		return nil, err
	}
	if !auth.CheckPassword(u.PasswordHash, password) {
		return nil, httpx.ErrUnauthorized("invalid credentials")
	}
	return s.tokensFor(u)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.issuer.Parse(refreshToken)
	if err != nil {
		return nil, httpx.ErrUnauthorized("invalid refresh token")
	}
	if claims.Type != auth.TokenRefresh {
		return nil, httpx.ErrUnauthorized("not a refresh token")
	}
	u, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, httpx.ErrUnauthorized("user not found")
		}
		return nil, err
	}
	return s.tokensFor(u)
}

func (s *AuthService) tokensFor(u *model.User) (*TokenPair, error) {
	access, err := s.issuer.IssueAccess(u)
	if err != nil {
		return nil, err
	}
	refresh, err := s.issuer.IssueRefresh(u)
	if err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: refresh, User: u}, nil
}
