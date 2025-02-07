package route

import (
	"github.com/julienschmidt/httprouter"
	"gocdc/internal/delivery/http"
	"gocdc/internal/delivery/http/middleware"
)

type RouteConfig struct {
	Router            *httprouter.Router
	UserController    *http.UserController
	ProductController *http.ProductController
	AuthMiddleware    *middleware.AuthMiddleware
}

func (c *RouteConfig) SetupRoute() {
	c.Router.POST("/refresh", c.UserController.TokenRenewal)
	c.Router.GET("/user", c.AuthMiddleware.ServeHTTP(c.UserController.FindUserInfo))
	c.Router.POST("/register", c.UserController.Register)
	c.Router.POST("/login", c.UserController.Login)
	c.Router.PATCH("/user", c.AuthMiddleware.ServeHTTP(c.UserController.Update))
	c.Router.DELETE("/user", c.AuthMiddleware.ServeHTTP(c.UserController.Delete))

	c.Router.GET("/producthomepage", c.ProductController.FindProductHomePage)
	c.Router.GET("/product", c.ProductController.FindAllProduct)
	c.Router.GET("/product/:productID", c.ProductController.FindProductInfo)
	c.Router.POST("/product", c.AuthMiddleware.ServeHTTP(c.ProductController.Create))
	c.Router.PATCH("/product/:productID", c.AuthMiddleware.ServeHTTP(c.ProductController.Update))
	c.Router.DELETE("/product/:productID", c.AuthMiddleware.ServeHTTP(c.ProductController.Delete))
}
