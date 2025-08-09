package router

import (
	"github.com/cristiano-pacheco/pingo/internal/shared/modules/http/httpserver"
	"github.com/gofiber/fiber/v2"
)

type Router struct {
	server *httpserver.FiberHTTPServer
}

func NewRouter(server *httpserver.FiberHTTPServer) *Router {
	return &Router{server: server}
}

func (r *Router) Router() *fiber.App {
	return r.server.App()
}
