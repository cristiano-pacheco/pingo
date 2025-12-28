package identity

import (
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/cache"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event/consumer"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/event/producer"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/http/fiber/handler"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/http/fiber/middleware"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/http/fiber/router"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/repository"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/service"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/usecase"
	"github.com/cristiano-pacheco/pingo/internal/modules/identity/validator"
	shared_kafka "github.com/cristiano-pacheco/pingo/internal/shared/modules/kafka"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/pkg/kafka"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"identity",
	fx.Provide(
		handler.NewAuthHandler,
		handler.NewUserHandler,

		fx.Annotate(
			repository.NewUserRepository,
			fx.As(new(repository.UserRepositoryI)),
		),
		fx.Annotate(
			repository.NewOneTimeTokenRepository,
			fx.As(new(repository.OneTimeTokenRepositoryI)),
		),

		service.NewSendEmailConfirmationService,
		service.NewSendEmailVerificationCodeService,
		service.NewEmailTemplateService,
		service.NewTokenService,
		service.NewHashService,
		service.NewUserActivationService,

		validator.NewPasswordValidator,

		fx.Annotate(
			cache.NewUserActivatedCache,
			fx.As(new(cache.UserActivatedCacheI)),
		),

		usecase.NewUserActivateUseCase,
		usecase.NewUserCreateUseCase,
		usecase.NewAuthLoginUseCase,
		usecase.NewAuthGenerateTokenUseCase,
		usecase.NewUserUpdateUseCase,

		middleware.NewAuthMiddleware,

		producer.NewUserAuthenticatedProducer,
		producer.NewUserCreatedProducer,
		producer.NewUserUpdatedProducer,

		consumer.NewUserCreatedConsumer,
		consumer.NewUserAuthenticatedConsumer,
	),
	fx.Invoke(
		router.SetupUserRoutes,
		router.SetupAuthRoutes,
		consumer.NewUserCreatedConsumer,
		registerConsumerRunners,
	),
)

func registerConsumerRunners(
	builder kafka.Builder,
	logger logger.Logger,
	lc fx.Lifecycle,
	userCreatedConsumer *consumer.UserCreatedConsumer,
	userAuthenticatedConsumer *consumer.UserAuthenticatedConsumer,
) {
	shared_kafka.NewConsumerRunner(builder, userCreatedConsumer, logger, lc)
	shared_kafka.NewConsumerRunner(builder, userAuthenticatedConsumer, logger, lc)
}
