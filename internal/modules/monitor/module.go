package monitor

import (
	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/http/fiber/handler"
	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/http/fiber/router"
	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/repository"
	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/usecase"
	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/validator"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"monitor",
	fx.Provide(
		handler.NewContactHandler,

		fx.Annotate(
			repository.NewContactRepository,
			fx.As(new(repository.ContactRepositoryI)),
		),
		fx.Annotate(
			repository.NewHTTPMonitorRepository,
			fx.As(new(repository.HTTPMonitorRepositoryI)),
		),
		fx.Annotate(
			repository.NewHTTPMonitorCheckRepository,
			fx.As(new(repository.HTTPMonitorCheckRepositoryI)),
		),
		fx.Annotate(
			repository.NewNotificationRepository,
			fx.As(new(repository.NotificationRepositoryI)),
		),

		fx.Annotate(
			validator.NewContactValidator,
			fx.As(new(validator.ContactValidatorI)),
		),

		usecase.NewContactCreateUseCase,
		usecase.NewContactListUseCase,
		usecase.NewContactUpdateUseCase,
		usecase.NewContactDeleteUseCase,

		fx.Invoke(router.SetupContactRoutes),
	),
)
