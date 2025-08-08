package router

import (
	"github.com/go-chi/chi/v5"

	"github.com/cristiano-pacheco/pingo/internal/shared/modules/http/httpserver"
)

type Router struct {
	server *httpserver.HTTPServer
}

func NewRouter(server *httpserver.HTTPServer) *Router {
	return &Router{server: server}
}

func (r *Router) Router() *chi.Mux {
	return r.server.Router()
}
