package middleware

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/cristiano-pacheco/pingo/internal/modules/identity/errs"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository"
	internal_jwt "github.com/cristiano-pacheco/pingo/internal/shared/modules/jwt"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/registry"
	"github.com/cristiano-pacheco/pingo/internal/shared/sdk/http/request"
)

type AuthMiddleware struct {
	jwtParser          *jwt.Parser
	logger             logger.Logger
	privateKeyRegistry registry.PrivateKeyRegistry
	userRepository     repository.UserRepository
}

func NewAuthMiddleware(
	jwtParser *jwt.Parser,
	logger logger.Logger,
	privateKeyRegistry registry.PrivateKeyRegistry,
	userRepository repository.UserRepository,
) *AuthMiddleware {
	return &AuthMiddleware{
		jwtParser,
		logger,
		privateKeyRegistry,
		userRepository,
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

		ctx := c.Context()
		isActivated, err := m.userRepository.IsUserActivated(ctx, userID)
		if err != nil {
			return err
		}

		if !isActivated {
			return errs.ErrUserIsNotActive
		}

		// Store user ID in context
		c.Locals(string(request.UserIDKey), userID)

		return nil
	}
}
