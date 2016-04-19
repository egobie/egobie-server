package routes

import (
	"github.com/eGobie/egobie-server/controllers"
)

func initServiceRoutes() {
	// Get all services
	serviceRouter.POST("/", controllers.GetService)

	serviceRouter.POST("/user", controllers.GetUserService)

	serviceRouter.GET("/opening", controllers.GetOpening)
}
