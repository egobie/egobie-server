package routes

import (
	"encoding/json"
	"fmt"
	"database/sql"
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"

	"github.com/gin-gonic/gin"
)

var (
	router        = gin.New()
	userRouter    = router.Group("/user")
	carRouter     = router.Group("/car")
	paymentRouter = router.Group("/payment")
	serviceRouter = router.Group("/service")
	historyRouter = router.Group("/history")

	egobieRouter = router.Group("/egobie")
)

func init() {
	// CORS
	router.Use(cors, request, sleep)

	// CORS, Authorize User
	userRouter.Use(cors, request, authorizeResidentialUser)
	carRouter.Use(cors, request, authorizeResidentialUser)
	paymentRouter.Use(cors, request, authorizeResidentialUser)
	serviceRouter.Use(cors, request, authorizeResidentialUser)
	historyRouter.Use(cors, request, authorizeResidentialUser)

	egobieRouter.Use(cors, request, authorizeEgobieUser)

	initSignRoutes()
	initUserRoutes()
	initServiceRoutes()
	initCarRoutes()
	initPaymentRoutes()
	initHistoryRoutes()

	initEgobieRoutes()
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

func authorizeResidentialUser(c *gin.Context) {
	var (
		err      error
		token    string
		userId   int32
		userType string
	)

	if userId, token, err = parseUser(c); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	} else if int32(len(token)) != modules.USER_RESIDENTIAL_TOKEN {
		c.IndentedJSON(http.StatusBadRequest, "Invalid user")
		c.Abort()
		return
	}

	if userType, err = readUser(userId, token); err != nil {
		if err == sql.ErrNoRows {
			c.IndentedJSON(http.StatusBadRequest, "Invalid user")
		} else {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
		}

		c.Abort()
		return
	} else if !modules.IsResidential(userType) {
		c.IndentedJSON(http.StatusBadRequest, "Invalid user")
		c.Abort()
		return
	}

	c.Next()
}

func authorizeEgobieUser(c *gin.Context) {
	var (
		err      error
		token    string
		userId   int32
		userType string
	)

	if userId, token, err = parseUser(c); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	} else if int32(len(token)) != modules.USER_EGOBIE_TOKEN {
		c.IndentedJSON(http.StatusBadRequest, "Invalid user")
		c.Abort()
		return
	}

	if userType, err = readUser(userId, token); err != nil {
		if err == sql.ErrNoRows {
			c.IndentedJSON(http.StatusBadRequest, "Invalid user")
		} else {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
		}

		c.Abort()
		return
	} else if !modules.IsEgobie(userType) {
		c.IndentedJSON(http.StatusBadRequest, "Invalid user")
		c.Abort()
		return
	}

	c.Next()
}

func parseUser(c *gin.Context) (int32, string, error) {
	request := modules.BaseRequest{}
	var (
		data   []byte
		err    error
	)

	if data, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return -1, "", err
	}

	if err = json.Unmarshal(data, &request); err != nil {
		return -1, "", err
	}

	// Put request body back
	c.Request.Body = ioutil.NopCloser(bytes.NewReader(data))

	return request.UserId, request.UserToken, nil
}

func readUser(userId int32, token string) (userType string, err error) {
	if err = config.DB.QueryRow(
		"select type from user where id = ? and password like ?",
		userId, token+"%",
	).Scan(&userType); err != nil {
		return "", err
	}

	return userType, nil
}

func request(c *gin.Context) {
	fmt.Println(c.Request.URL)

	c.Next()
}

func sleep(c *gin.Context) {
	//	time.Sleep(500 * time.Millisecond)

	c.Next()
}

func Serve(port string) {
	router.Run(port)
}
