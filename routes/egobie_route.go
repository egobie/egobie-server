package routes

import (
	"github.com/egobie/egobie-server/controllers"
)

func initEgobieRoutes() {

	egobieRouter.POST("/service/task", controllers.GetTask)

	egobieRouter.POST("/service/done", controllers.MakeServiceDone)

	egobieRouter.POST("/service/reserved", controllers.MakeServiceReserved)

	egobieRouter.POST("/service/progress", controllers.MakeServiceInProgress)

	egobieRouter.POST("/fleet/user/new", controllers.NewFleetUser)
}
