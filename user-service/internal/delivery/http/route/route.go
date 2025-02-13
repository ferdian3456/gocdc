package route

import (
	"github.com/julienschmidt/httprouter"
	"gocdc/internal/delivery/http"
	"gocdc/internal/delivery/http/middleware"
)

type RouteConfig struct {
	Router         *httprouter.Router
	UserController *http.UserController
	AuthMiddleware *middleware.AuthMiddleware
}

func (c *RouteConfig) SetupRoute() {
	c.Router.POST("/refresh", c.UserController.TokenRenewal)
	c.Router.GET("/user", c.AuthMiddleware.ServeHTTP(c.UserController.FindUserInfo))
	c.Router.POST("/register", c.UserController.Register)
	c.Router.POST("/login", c.UserController.Login)
	c.Router.PATCH("/user", c.AuthMiddleware.ServeHTTP(c.UserController.Update))
	c.Router.DELETE("/user", c.AuthMiddleware.ServeHTTP(c.UserController.Delete))
	c.Router.GET("/user/:userUUID", c.UserController.FindByUUId)
}
