package router

import (
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/http/httpserver"
	"github.com/gofiber/fiber/v2"
)

type FiberRouter struct {
	server *httpserver.FiberHTTPServer
}

func NewFiberRouter(server *httpserver.FiberHTTPServer) *FiberRouter {
	return &FiberRouter{server: server}
}

func (r *FiberRouter) Router() *fiber.App {
	return r.server.App()
}
