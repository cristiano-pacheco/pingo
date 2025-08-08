package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/cristiano-pacheco/pingo/internal/shared/modules/database"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/errs"
	internal_jwt "github.com/cristiano-pacheco/pingo/internal/shared/modules/jwt"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/registry"
	"github.com/cristiano-pacheco/pingo/internal/shared/sdk/http/request"
	"github.com/cristiano-pacheco/pingo/internal/shared/sdk/http/response"
)

type AuthMiddleware struct {
	jwtParser          *jwt.Parser
	db                 *database.GoflixDB
	logger             logger.Logger
	errorMapper        errs.ErrorMapper
	privateKeyRegistry registry.PrivateKeyRegistry
}

func NewAuthMiddleware(
	jwtParser *jwt.Parser,
	db *database.GoflixDB,
	logger logger.Logger,
	errorMapper errs.ErrorMapper,
	privateKeyRegistry registry.PrivateKeyRegistry,
) *AuthMiddleware {
	return &AuthMiddleware{jwtParser, db, logger, errorMapper, privateKeyRegistry}
}

func (m *AuthMiddleware) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract and validate token
		bearerToken := r.Header.Get("Authorization")
		if !strings.HasPrefix(bearerToken, "Bearer ") {
			m.handleError(w, errs.ErrInvalidToken)
			return
		}

		jwtToken := strings.TrimSpace(bearerToken[7:])
		pk := m.privateKeyRegistry.Get()

		tokenKeyFunc := func(_ *jwt.Token) (interface{}, error) {
			return &pk.PublicKey, nil
		}

		var claims internal_jwt.Claims
		token, err := m.jwtParser.ParseWithClaims(jwtToken, &claims, tokenKeyFunc)
		if err != nil {
			m.handleError(w, errs.ErrInvalidToken)
			return
		}

		if !token.Valid {
			m.handleError(w, errs.ErrInvalidToken)
			return
		}

		userID, err := strconv.ParseUint(claims.Subject, 10, 64)
		if err != nil {
			m.handleError(w, errs.ErrInvalidToken)
			return
		}

		ctx := r.Context()
		isActivated, err := m.isUserActivated(ctx, userID)
		if err != nil {
			m.handleError(w, errs.ErrInternalServer)
			return
		}

		if !isActivated {
			mError := m.errorMapper.MapCustomError(http.StatusUnauthorized, errs.ErrUserIsNotActivated.Error())
			response.Error(w, mError)
			return
		}

		// Store user ID in context
		ctx = context.WithValue(ctx, request.UserIDKey, userID)

		// Call next handler with updated context
		next(w, r.WithContext(ctx))
	}
}

func (m *AuthMiddleware) isUserActivated(ctx context.Context, userID uint64) (bool, error) {
	var isActivated bool
	err := m.db.WithContext(ctx).
		Table("users").
		Select("is_activated").
		Where("id = ?", userID).
		Scan(&isActivated).Error

	if err != nil {
		m.logger.Error("error checking if user is activated", "error", err)
		return false, err
	}

	return isActivated, nil
}

func (m *AuthMiddleware) handleError(w http.ResponseWriter, err error) {
	rError := m.errorMapper.Map(err)
	response.Error(w, rError)
}
