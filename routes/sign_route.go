package routes

import (
	"github.com/eGobie/server/controllers"
)

func initSignRoutes() {
	router.POST("/signup", controllers.SignUp)

	router.POST("/signin", controllers.SignIn)

	router.GET("/secure/:code", controllers.Secure)
}
