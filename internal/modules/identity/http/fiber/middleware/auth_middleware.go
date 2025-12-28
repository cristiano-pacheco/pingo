package middleware

import (
	"context"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	internal_jwt "github.com/cristiano-pacheco/pingo/internal/shared/modules/jwt"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/registry"
	"github.com/cristiano-pacheco/pingo/internal/shared/sdk/http/request"
)

type AuthMiddleware struct {
	privateKeyRegistry    registry.PrivateKeyRegistryI
	userActivationService service.UserActivationServiceI
	jwtParser             *jwt.Parser
	logger                logger.Logger
}

func NewAuthMiddleware(
	privateKeyRegistry registry.PrivateKeyRegistryI,
	userActivationService service.UserActivationServiceI,
	jwtParser *jwt.Parser,
	logger logger.Logger,
) *AuthMiddleware {
	return &AuthMiddleware{
		privateKeyRegistry,
		userActivationService,
		jwtParser,
		logger,
	}
}

func (m *AuthMiddleware) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		bearerToken := c.Get("Authorization")
		if !strings.HasPrefix(bearerToken, "Bearer ") {
			return fiber.ErrUnauthorized
		}

		jwtToken := strings.TrimSpace(bearerToken[7:])
		pk := m.privateKeyRegistry.Get()

		tokenKeyFunc := func(_ *jwt.Token) (interface{}, error) {
			return &pk.PublicKey, nil
		}

		var claims internal_jwt.Claims
		token, err := m.jwtParser.ParseWithClaims(jwtToken, &claims, tokenKeyFunc)
		if err != nil || !token.Valid {
			return errs.ErrInvalidToken
		}

		userID, err := strconv.ParseUint(claims.Subject, 10, 64)
		if err != nil {
			return errs.ErrInvalidToken
		}

		ctx := c.UserContext()
		isActivated, err := m.userActivationService.IsUserActivated(ctx, userID)
		if err != nil {
			return err
		}

		if !isActivated {
			return errs.ErrUserIsNotActive
		}

		newCtx := context.WithValue(ctx, request.UserIDKey, userID)
		c.SetUserContext(newCtx)

		return c.Next()
	}
}
