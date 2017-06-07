package routes

import (
	"github.com/egobie/egobie-server/controllers"
)

func initCarRoutes() {
	carRouter.POST("", controllers.GetCarById)

	carRouter.POST("/make", controllers.GetCarMake)

	carRouter.POST("/model", controllers.GetCarModel)

	carRouter.POST("/user", controllers.GetCarForUser)

	carRouter.POST("/new", controllers.CreateCar)

	carRouter.POST("/update", controllers.UpdateCar)

	carRouter.POST("/delete", controllers.DeleteCar)
}
