package routes

import (
	"github.com/egobie/egobie-server/controllers"
)

func initEgobieRoutes() {

	egobieRouter.POST("/service/finish", controllers.MakeServiceFinish)

	egobieRouter.POST("/service/reserved", controllers.MakeServiceReserved)

	egobieRouter.POST("/service/progress", controllers.MakeServiceInProgress)
}
