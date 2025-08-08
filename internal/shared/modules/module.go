package modules

import (
	"go.uber.org/fx"

	"github.com/cristiano-pacheco/pingo/internal/shared/modules/config"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/database"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/errs"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/http/httpserver"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/http/middleware"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/http/router"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/jwt"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/mailer"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/redis"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/registry"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/translator"
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/validator"
)

var Module = fx.Module(
	"shared/modules",
	config.Module,
	database.Module,
	validator.Module,
	translator.Module,
	logger.Module,
	registry.Module,
	jwt.Module,
	mailer.Module,
	errs.Module,
	redis.Module,
	router.Module,
	middleware.Module,
	httpserver.Module,
)
