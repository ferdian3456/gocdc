package route

import (
	"github.com/julienschmidt/httprouter"
	"gocdc/internal/delivery/http"
	"gocdc/internal/delivery/http/middleware"
)

type RouteConfig struct {
	Router            *httprouter.Router
	ProductController *http.ProductController
	AuthMiddleware    *middleware.AuthMiddleware
}

func (c *RouteConfig) SetupRoute() {
	c.Router.GET("/producthomepage", c.ProductController.FindProductHomePage)
	c.Router.GET("/product", c.ProductController.FindAllProduct)
	c.Router.GET("/product/:productID", c.ProductController.FindProductInfo)
	c.Router.POST("/product", c.AuthMiddleware.ServeExternalService(c.ProductController.Create))
	c.Router.PATCH("/product/:productID", c.AuthMiddleware.ServeHTTP(c.ProductController.Update))
	c.Router.DELETE("/product/:productID", c.AuthMiddleware.ServeHTTP(c.ProductController.Delete))
}
