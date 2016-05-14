package routes

import (
	"github.com/egobie/egobie-server/controllers"
)

func initCarRoutes() {
	carRouter.POST("/maker", controllers.GetCarMaker)

	carRouter.POST("/model", controllers.GetCarModel)

//	carRouter.GET("/model/:makerId", controllers.GetCarModelForMaker)

	carRouter.POST("", controllers.GetCarById)

	carRouter.POST("/user", controllers.GetCarForUser)

	carRouter.POST("/new", controllers.CreateCar)

	carRouter.POST("/update", controllers.UpdateCar)

	carRouter.POST("/delete", controllers.DeleteCar)
}
