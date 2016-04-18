package routes

import (
	"github.com/eGobie/egobie-server/controllers"
)

func initCarRoutes() {
	carRouter.GET("/maker", controllers.GetCarMaker)

	carRouter.GET("/model/:makerId", controllers.GetCarModel)

	carRouter.POST("", controllers.GetCarById)

	carRouter.POST("/user", controllers.GetCarForUser)

	carRouter.POST("/new", controllers.CreateCar)

	carRouter.POST("/update", controllers.UpdateCar)

	carRouter.POST("/delete", controllers.DeleteCar)
}
