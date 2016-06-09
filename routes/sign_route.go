package routes

import (
	"github.com/egobie/egobie-server/controllers"
)

func initSignRoutes() {
	router.POST("/check/email", controllers.CheckEmail)

	router.POST("/check/name", controllers.CheckUsername)

	router.POST("/signup", controllers.SignUp)

	router.POST("/signin", controllers.SignIn)

	router.POST("/signup/fleet", controllers.SignUpFleet)

	router.GET("/secure/:code", controllers.Secure)
}
