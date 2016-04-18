package routes

import (
	"github.com/eGobie/server/controllers"
)

func initServiceRoutes() {
	// Get all services
	serviceRouter.POST("/", controllers.GetService)

	serviceRouter.POST("/user", controllers.GetUserService)
}
