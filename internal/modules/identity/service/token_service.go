package service

import (
	"context"
	"strconv"
	"time"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository/model"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/otel"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/registry"
	"github.com/golang-jwt/jwt/v5"
)

type TokenService interface {
	GenerateJWT(ctx context.Context, user model.UserModel) (string, error)
}

type tokenService struct {
	privateKeyRegistry registry.PrivateKeyRegistry
	conf               config.Config
	logger             logger.Logger
}

func NewTokenService(
	conf config.Config,
	privateKeyRegistry registry.PrivateKeyRegistry,
	logger logger.Logger,
) TokenService {
	return &tokenService{privateKeyRegistry, conf, logger}
}

func (s *tokenService) GenerateJWT(ctx context.Context, user model.UserModel) (string, error) {
	_, span := otel.Trace().StartSpan(ctx, "TokenService.GenerateJWT")
	defer span.End()

	now := time.Now()
	duration := time.Duration(s.conf.JWT.ExpirationInSeconds) * time.Second
	expires := now.Add(duration)
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expires),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    s.conf.JWT.Issuer,
		Subject:   strconv.FormatUint(user.ID, 10),
	}

	method := jwt.GetSigningMethod(jwt.SigningMethodRS256.Name)
	token := jwt.NewWithClaims(method, claims)

	pk := s.privateKeyRegistry.Get()
	signedToken, err := token.SignedString(pk)
	if err != nil {
		message := "[generate_token] error signing token"
		s.logger.Error(message, "error", err)
		return "", err
	}

	return signedToken, nil
}
