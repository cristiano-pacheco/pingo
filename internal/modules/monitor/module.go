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

		repository.NewContactRepository,
		repository.NewHTTPMonitorRepository,
		repository.NewHTTPMonitorCheckRepository,
		repository.NewNotificationRepository,

		validator.NewContactValidator,

		usecase.NewContactCreateUseCase,
		usecase.NewContactListUseCase,
		usecase.NewContactUpdateUseCase,
		usecase.NewContactDeleteUseCase,

		fx.Invoke(router.SetupContactRoutes),
	),
)
