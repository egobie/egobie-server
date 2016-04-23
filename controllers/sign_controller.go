package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"
	"github.com/egobie/egobie-server/secures"

	"github.com/gin-gonic/gin"
)

func updateUserSignIn(userId int32) {
	query := `
		update user set sign_in = CURRENT_TIMESTAMP where id = ?
	`

	if stmt, err := config.DB.Prepare(query); err == nil {
		defer stmt.Close()

		if _, err = stmt.Exec(userId); err != nil {
			fmt.Println("fail to update sign-in - ", err.Error())
		}
	} else {
		fmt.Println("fail to update sign-in - ", err.Error())
	}
}

func SignUp(c *gin.Context) {
	query := `
		insert into user (type, username, password, email, phone_number)
		values ('RESIDENTIAL', ?, ?, ?, ?)
	`
	request := modules.SignUp{}
	var (
		stmt         *sql.Stmt
		result       sql.Result
		enPassword   string
		lastInsertId int64
		body         []byte
		err          error
	)

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if stmt, err = config.DB.Prepare(query); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	defer stmt.Close()

	if enPassword, err = secures.EncryptPassword(request.Password); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if result, err = stmt.Exec(
		request.Username, enPassword, request.Email, request.PhoneNumber,
	); err != nil {
		message := ""

		if isDuplicateEntryError(err) {
			message = "Username or email already exists!"
		} else {
			message = err.Error()
		}

		c.IndentedJSON(http.StatusBadRequest, message)
		return
	}

	if lastInsertId, err = result.LastInsertId(); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if user, err := getUserById(int32(lastInsertId)); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
	} else {
		user.Password = getUserToken(user.Password);
		c.IndentedJSON(http.StatusOK, user);
	}
}

func SignIn(c *gin.Context) {
	request := modules.SignIn{}
	var (
		dePassword string
		user       modules.User
		body       []byte
		err        error
	)

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if user, err = getUser(
		"username = ?", request.Username,
	); err != nil {
		switch {
			case err == sql.ErrNoRows:
				c.IndentedJSON(http.StatusBadRequest, "User not found")
			default:
				c.IndentedJSON(http.StatusBadRequest, err.Error())
		}
		return
	}

	if dePassword, err = secures.DecryptPassword(user.Password); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if dePassword != request.Password {
		c.IndentedJSON(http.StatusBadRequest, "Password not match")
		return
	}

	updateUserSignIn(user.Id)

	user.Password = getUserToken(user.Password)

	c.IndentedJSON(http.StatusOK, user)
}

func Secure(c *gin.Context) {
	if code, err := secures.EncryptPassword(c.Param("code")); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	} else {
		c.IndentedJSON(http.StatusOK, code)
	}
}
