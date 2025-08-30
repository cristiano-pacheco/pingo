package identity

import (
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event/producer"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/http/fiber/handler"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/http/fiber/middleware"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/http/fiber/router"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/usecase"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/validator"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"identity",
	fx.Provide(
		handler.NewAuthHandler,
		handler.NewUserHandler,

		repository.NewUserRepository,
		repository.NewOneTimeTokenRepository,

		service.NewSendEmailConfirmationService,
		service.NewSendEmailVerificationCodeService,
		service.NewEmailTemplateService,
		service.NewTokenService,
		service.NewHashService,

		validator.NewPasswordValidator,

		usecase.NewUserActivateUseCase,
		usecase.NewUserCreateUseCase,
		usecase.NewAuthLoginUseCase,
		usecase.NewAuthGenerateTokenUseCase,
		usecase.NewUserUpdateUseCase,

		middleware.NewAuthMiddleware,

		producer.NewUserAuthenticatedProducer,
		producer.NewUserCreatedProducer,
		producer.NewUserUpdatedProducer,
	),
	fx.Invoke(
		router.SetupUserRoutes,
		router.SetupAuthRoutes,
	),
)
