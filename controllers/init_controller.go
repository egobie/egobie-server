package controllers

import (
	"strings"
	"database/sql"

	"github.com/egobie/egobie-server/config"

	"gopkg.in/guregu/null.v3"
)

func authorizeUser(userId int32, token string) (valid bool, err error) {
	query := `
		select id from user where user_id = ? and password like ?
	`
	var (
		stmt *sql.Stmt
		id   null.Int
	)

	if stmt, err = config.DB.Prepare(query); err != nil {
		return false, err
	}

	if err = stmt.QueryRow(userId, token+"%").Scan(id); err != nil {
		return false, err
	}

	return true, nil
}

func isDuplicateEntryError(err error) bool {
	return strings.HasPrefix(err.Error(), "Error 1062: Duplicate entry")
}
