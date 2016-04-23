package routes

import (
	"github.com/egobie/egobie-server/controllers"
)

func initServiceRoutes() {
	// Get all services
	serviceRouter.GET("", controllers.GetService)

	serviceRouter.POST("/user", controllers.GetUserService)

	serviceRouter.GET("/opening", controllers.GetOpening)

	serviceRouter.POST("/order", controllers.PlaceOrder)

	serviceRouter.GET("/demand/:id", controllers.Demand)
}
