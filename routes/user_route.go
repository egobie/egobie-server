package routes

import (
	"github.com/eGobie/egobie-server/controllers"
)

func initUserRoutes() {

	userRouter.POST("/", controllers.GetUser)

	userRouter.POST("/update/password", controllers.UpdatePassword)

	userRouter.POST("/update/user", controllers.UpdateUser)

	userRouter.POST("/update/home", controllers.UpdateHome)

	userRouter.POST("/update/work", controllers.UpdateWork)

}
