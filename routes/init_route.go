package routes

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/eGobie/egobie-server/config"

	"github.com/gin-gonic/gin"
)

var (
	router        = gin.New()
	userRouter    = router.Group("/user")
	carRouter     = router.Group("/car")
	paymentRouter = router.Group("/payment")
	serviceRouter = router.Group("/service")
	historyRouter = router.Group("/history")
)

func init() {
	// CORS
	router.Use(cors, request)

	// CORS, Authorize User
	userRouter.Use(cors, request, authorizeUser)
	carRouter.Use(cors, request, authorizeUser)
	paymentRouter.Use(cors, request, authorizeUser)
	serviceRouter.Use(cors, request, authorizeUser)
	historyRouter.Use(cors, request, authorizeUser)

	initSignRoutes()
	initUserRoutes()
	initServiceRoutes()
	initCarRoutes()
	initPaymentRoutes()
	initHistoryRoutes()
}

func cors(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set(
		"Access-Control-Allow-Headers",
		"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With"+
			", "+config.EGOBIE_HEADER_TOKEN+
			", "+config.EGOBIE_HEADER_USERID,
	)
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, DELETE, GET, PUT")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.Next()
}

func authorizeUser(c *gin.Context) {
	var (
		err    error
		stmt   *sql.Stmt
		id     int64
		userId int64
	)

	token, ok := c.Request.Header[config.EGOBIE_HEADER_TOKEN]

	if !ok {
		fmt.Println(config.EGOBIE_HEADER_TOKEN)
		c.IndentedJSON(http.StatusBadRequest, "Token Header")
		c.Abort()
		return
	}

	userIds, ok := c.Request.Header[config.EGOBIE_HEADER_USERID]

	if !ok {
		fmt.Println(config.EGOBIE_HEADER_USERID)
		c.IndentedJSON(http.StatusBadRequest, "Id Header")
		c.Abort()
		return
	}

	if userId, err = strconv.ParseInt(userIds[0], 10, 32); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	query := `
		select id from user where id = ? and password like ?
	`

	if stmt, err = config.DB.Prepare(query); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}
	defer stmt.Close()

	if err := stmt.QueryRow(int32(userId), token[0]+"%").Scan(&id); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	c.Next()
}

func request(c *gin.Context) {
	fmt.Println("New Request - start")
	fmt.Println(c.Request.URL)
	fmt.Println("New Request - end")

	c.Next()
}

func Serve(port string) {
	router.Run(port)
}
