package router

import (
	"github.com/gin-gonic/gin"

	"github.com/cristiano-pacheco/pingo/internal/shared/modules/http/httpserver"
)

type Router struct {
	server *httpserver.GinHTTPServer
}

func NewRouter(server *httpserver.GinHTTPServer) *Router {
	return &Router{server: server}
}

func (r *Router) Router() *gin.Engine {
	return r.server.Engine()
}
