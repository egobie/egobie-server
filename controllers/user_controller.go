package controllers

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/eGobie/egobie-server/config"
	"github.com/eGobie/egobie-server/modules"
	"github.com/eGobie/egobie-server/secures"

	"github.com/gin-gonic/gin"
)

func getUser(condition string, args ...interface{}) (user modules.User, err error) {
	query := `
		select id, type, username, password,
			email, phone_number,
			first_name, last_name, middle_name,
			home_address_state, home_address_zip,
			home_address_city, home_address_street,
			work_address_state, work_address_zip,
			work_address_city, work_address_street
		from user where
	`
	var (
		stmt *sql.Stmt
	)

	if stmt, err = config.DB.Prepare(query + " " + condition); err != nil {
		return
	}

	if err = stmt.QueryRow(args...).Scan(
		&user.Id, &user.Type, &user.Username, &user.Password,
		&user.Email, &user.PhoneNumber,
		&user.FirstName, &user.LastName, &user.MiddleName,
		&user.HomeAddressState, &user.HomeAddressZip,
		&user.HomeAddressCity, &user.HomeAddressStreet,
		&user.WorkAddressState, &user.WorkAddressZip,
		&user.WorkAddressCity, &user.WorkAddressStreet,
	); err != nil {
		return
	}

	return user, nil
}

func getUserById(id int32) (user modules.User, err error) {
	return getUser("id = ?", id)
}

func getUserByUsername(username string) (user modules.User, err error) {
	return getUser("username = ?", username)
}

func getUserToken(password string) string {
	return password[:4]
}

func updateAddress(body []byte, setClause string) (err error) {
	query := "update user set " + setClause + " where id = ? and password like ?"
	request := modules.UpdateAddress{}
	var (
		stmt *sql.Stmt
	)

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	if stmt, err = config.DB.Prepare(query); err != nil {
		return
	}
	defer stmt.Close()

	if _, err = stmt.Exec(
		request.State, request.Zip, request.City, request.Street,
		request.UserId, request.UserToken+"%",
	); err != nil {
		return
	}

	return nil
}

func GetUser(c *gin.Context) {
	request := modules.UserRequest{}
	var (
		body []byte
		user modules.User
		err  error
	)

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if user, err = getUser(
		"id = ? and password like ?", request.UserId, request.UserToken+"%",
	); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	user.Password = ""

	c.IndentedJSON(http.StatusOK, user)
}

func UpdateUser(c *gin.Context) {
	query := `
		update user
		set first_name = ?, last_name = ?, middle_name = ?,
			email = ?, phone_number = ?
		where id = ? and password like ?
	`
	request := modules.UpdateUser{}
	var (
		stmt *sql.Stmt
		body []byte
		err  error
	)

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if stmt, err = config.DB.Prepare(query); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}
	defer stmt.Close()

	if _, err = stmt.Exec(
		request.FirstName, request.LastName, request.MiddleName, request.Email,
		request.PhoneNumber, request.UserId, request.UserToken+"%",
	); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if user, err := getUserById(request.UserId); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
	} else {
		user.Password = getUserToken(user.Password)
		c.IndentedJSON(http.StatusOK, user)
	}
}

func UpdatePassword(c *gin.Context) {
	query := "update user set password = ? where id = ?"
	request := modules.UpdatePassword{}
	var (
		stmt       *sql.Stmt
		user       modules.User
		dePassword string
		enPassword string
		body       []byte
		err        error
	)

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if user, err = getUserById(request.UserId); err != nil {
		c.IndentedJSON(http.StatusBadRequest, "User not found")
		c.Abort()
		return
	}

	if dePassword, err = secures.DecryptPassword(user.Password); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if dePassword != request.Password {
		c.IndentedJSON(http.StatusBadRequest, "Password is not correct!")
		c.Abort()
		return
	}

	if enPassword, err = secures.EncryptPassword(request.NewPassword); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if stmt, err = config.DB.Prepare(query); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}
	defer stmt.Close()

	if _, err = stmt.Exec(
		enPassword, request.UserId,
	); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	c.IndentedJSON(http.StatusOK, modules.UserInfo{request.UserId, getUserToken(enPassword)})
}

func UpdateHome(c *gin.Context) {
	var (
		body []byte
		err  error
	)

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if err = updateAddress(
		body, `home_address_state = ?, home_address_zip = ?, home_address_city = ?, home_address_street = ?`,
	); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	c.IndentedJSON(http.StatusOK, "OK")
}

func UpdateWork(c *gin.Context) {
	var (
		body []byte
		err  error
	)

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	if err = updateAddress(
		body, `work_address_state = ?, work_address_zip = ?, work_address_city = ?, work_address_street = ?`,
	); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	c.IndentedJSON(http.StatusOK, "OK")
}
