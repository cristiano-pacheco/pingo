package monitor

import (
	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/repository"
	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/usecase"
	"github.com/cristiano-pacheco/pingo/internal/modules/monitor/validator"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"monitor",
	fx.Provide(
		repository.NewContactRepository,
		repository.NewHTTPMonitorRepository,
		repository.NewHTTPMonitorCheckRepository,
		repository.NewNotificationRepository,

		validator.NewContactValidator,

		usecase.NewContactCreateUseCase,
	),
)
