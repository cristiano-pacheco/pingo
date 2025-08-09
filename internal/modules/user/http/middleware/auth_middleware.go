package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/cristiano-pacheco/pingo/internal/modules/user/repository"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/errs"
	shared_errs "github.com/cristiano-pacheco/pingo/internal/shared/modules/errs"
	internal_jwt "github.com/cristiano-pacheco/pingo/internal/shared/modules/jwt"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/registry"
	"github.com/cristiano-pacheco/pingo/internal/shared/sdk/http/request"
)

type AuthMiddleware struct {
	jwtParser          *jwt.Parser
	logger             logger.Logger
	errorMapper        shared_errs.ErrorMapper
	privateKeyRegistry registry.PrivateKeyRegistry
	userRepository     repository.UserRepository
}

func NewAuthMiddleware(
	jwtParser *jwt.Parser,
	logger logger.Logger,
	errorMapper errs.ErrorMapper,
	privateKeyRegistry registry.PrivateKeyRegistry,
	userRepository repository.UserRepository,
) *AuthMiddleware {
	return &AuthMiddleware{
		jwtParser,
		logger,
		errorMapper,
		privateKeyRegistry,
		userRepository,
	}
}

func (m *AuthMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		bearerToken := c.GetHeader("Authorization")
		if !strings.HasPrefix(bearerToken, "Bearer ") {
			m.handleGinError(c, errs.ErrInvalidToken)
			c.Abort()
			return
		}

		jwtToken := strings.TrimSpace(bearerToken[7:])
		pk := m.privateKeyRegistry.Get()

		tokenKeyFunc := func(_ *jwt.Token) (interface{}, error) {
			return &pk.PublicKey, nil
		}

		var claims internal_jwt.Claims
		token, err := m.jwtParser.ParseWithClaims(jwtToken, &claims, tokenKeyFunc)
		if err != nil || !token.Valid {
			m.handleGinError(c, errs.ErrInvalidToken)
			c.Abort()
			return
		}

		userID, err := strconv.ParseUint(claims.Subject, 10, 64)
		if err != nil {
			m.handleGinError(c, errs.ErrInvalidToken)
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		isActivated, err := m.userRepository.IsUserActivated(ctx, userID)
		if err != nil {
			m.handleGinError(c, shared_errs.ErrInternalServer)
			c.Abort()
			return
		}

		if !isActivated {
			mError := m.errorMapper.MapCustomError(http.StatusUnauthorized, errs.ErrUserIsNotActivated.Error())
			c.JSON(http.StatusUnauthorized, mError)
			c.Abort()
			return
		}

		// Store user ID in context
		c.Set(string(request.UserIDKey), userID)

		c.Next()
	}
}

func (m *AuthMiddleware) handleGinError(c *gin.Context, err error) {
	rError := m.errorMapper.Map(err)
	c.JSON(http.StatusUnauthorized, rError)
}
