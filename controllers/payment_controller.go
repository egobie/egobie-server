package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/egobie/egobie-server/config"
	"github.com/egobie/egobie-server/modules"
	"github.com/egobie/egobie-server/secures"

	"github.com/gin-gonic/gin"
	"github.com/lionelbarrow/braintree-go"
)

func encryptAccount(accountType, accountNumber, accountCode string) (enNumber, enCode string, err error) {
	switch {
	case accountType == config.CREDIT_CARD:
		if enNumber, err = secures.EncryptCredit(accountNumber); err != nil {
			return
		}

		if enCode, err = secures.EncryptCreditCVV(accountCode); err != nil {
			return
		}
	case accountType == config.DEBIT_CARD:
		if enNumber, err = secures.EncryptDebit(accountNumber); err != nil {
			return
		}

		if enCode, err = secures.EncryptDebitPin(accountCode); err != nil {
			return
		}
	default:
		return "", "", errors.New("Unknown payment type - " + accountType)
	}

	return enNumber, enCode, nil
}

func decryptAccount(accountType, accountNumber, accountCode string) (deNumber, deCode string, err error) {
	switch {
	case accountType == config.CREDIT_CARD:
		if deNumber, err = secures.DecryptCredit(accountNumber); err != nil {
			return
		}

		if deCode, err = secures.DecryptCreditCVV(accountCode); err != nil {
			return
		}
	case accountType == config.DEBIT_CARD:
		if deNumber, err = secures.DecryptDebit(accountNumber); err != nil {
			return
		}

		if deCode, err = secures.DecryptDebitPin(accountCode); err != nil {
			return
		}
	default:
		return "", "", errors.New("Unknown payment type - " + accountType)
	}

	return deNumber, deCode, nil
}

func decryptAccountNumber(accountType, accountNumber string) (deNumber string, err error) {
	switch {
	case accountType == config.CREDIT_CARD:
		if deNumber, err = secures.DecryptCredit(accountNumber); err != nil {
			return
		}
	case accountType == config.DEBIT_CARD:
		if deNumber, err = secures.DecryptDebit(accountNumber); err != nil {
			return
		}
	default:
		return "", errors.New("Unknown payment type - " + accountType)
	}

	return deNumber, nil
}

func decryptAccountCode(accountType, code string) (deNumber string, err error) {
	switch {
	case accountType == config.CREDIT_CARD:
		if deNumber, err = secures.DecryptCreditCVV(code); err != nil {
			return
		}
	case accountType == config.DEBIT_CARD:
		if deNumber, err = secures.DecryptDebitPin(code); err != nil {
			return
		}
	default:
		return "", errors.New("Unknown payment type - " + accountType)
	}

	return deNumber, nil
}

func getPaymentLastFour(accountNumber string) string {
	return accountNumber[len(accountNumber)-4:]
}

func getPaymentByIdAndUserId(id, userId int32) (payment modules.Payment, err error) {
	query := `
		select id, user_id, account_name, account_number,
				account_type, account_zip, code, card_type,
				expire_month, expire_year, reserved, card_type
		from user_payment
		where id = ? and user_id = ?
	`

	if err = config.DB.QueryRow(query, id, userId).Scan(
		&payment.Id, &payment.UserId, &payment.AccountName,
		&payment.AccountNumber, &payment.AccountType, &payment.AccountZip,
		&payment.Code, &payment.CardType, &payment.ExpireMonth,
		&payment.ExpireYear, &payment.Reserved, &payment.CardType,
	); err != nil {
		return
	}

	if payment.AccountNumber, err = decryptAccountNumber(
		payment.AccountType, payment.AccountNumber,
	); err != nil {
		return
	}

	if payment.Code, err = decryptAccountCode(
		payment.AccountType, payment.Code,
	); err != nil {
		return
	}

	return payment, nil
}

func getPaymentForUser(userId int32) (payments []modules.Payment, err error) {
	query := `
		select id, user_id, account_name, account_number, card_type,
				account_type, account_zip, expire_month, expire_year, reserved
		from user_payment
		where user_id = ?
	`
	var (
		rows     *sql.Rows
		deNumber string
	)

	if rows, err = config.DB.Query(query, userId); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		payment := modules.Payment{}

		if err = rows.Scan(
			&payment.Id, &payment.UserId, &payment.AccountName,
			&payment.AccountNumber, &payment.CardType, &payment.AccountType,
			&payment.AccountZip, &payment.ExpireMonth, &payment.ExpireYear,
			&payment.Reserved,
		); err != nil {
			return
		}

		if deNumber, err = decryptAccountNumber(
			payment.AccountType, payment.AccountNumber,
		); err != nil {
			return
		}

		payment.AccountNumber = getPaymentLastFour(deNumber)

		payments = append(payments, payment)
	}

	return payments, nil
}

func GetPaymentById(c *gin.Context) {
	request := modules.PaymentRequest{}
	var (
		payment modules.Payment
		body    []byte
		err     error
	)

	defer func() {
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	if payment, err = getPaymentByIdAndUserId(
		request.Id, request.UserId,
	); err == nil {
		payment.AccountNumber = getPaymentLastFour(payment.AccountNumber)
		payment.Code = ""

		c.IndentedJSON(http.StatusOK, payment)
	}
}

func GetPaymentByUserId(c *gin.Context) {
	request := modules.BaseRequest{}
	var (
		body     []byte
		payments []modules.Payment
		err      error
	)

	defer func() {
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	if payments, err = getPaymentForUser(request.UserId); err == nil {
		c.IndentedJSON(http.StatusOK, payments)
	}
}

func CreatePayment(c *gin.Context) {
	query := `
		insert into user_payment (user_id, account_name, account_number,
			account_type, account_zip, code, expire_month, expire_year, card_type)
		values (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	request := modules.PaymentNew{}
	var (
		result   sql.Result
		insertId int64
		enNumber string
		enCode   string
		body     []byte
		err      error
	)

	defer func() {
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	if enNumber, enCode, err = encryptAccount(
		request.AccountType, request.AccountNumber, request.Code,
	); err != nil {
		return
	}

	if result, err = config.DB.Exec(query,
		request.UserId, request.AccountName, enNumber, request.AccountType,
		request.AccountZip, enCode, request.ExpireMonth, request.ExpireYear,
		request.CardType,
	); err != nil {
		return
	}

	if insertId, err = result.LastInsertId(); err != nil {
		return
	}

	if payment, err := getPaymentByIdAndUserId(
		int32(insertId), request.UserId,
	); err == nil {
		c.IndentedJSON(http.StatusOK, payment)
	}
}

func UpdatePayment(c *gin.Context) {
	query := `
		update user_payment set account_name = ?, account_number = ?, card_type = ?,
		account_type = ?, account_zip = ?, code = ?, expire_month = ?, expire_year = ?
		where id = ? and user_id = ?
	`
	request := modules.UpdatePayment{}
	var (
		result      sql.Result
		affectedRow int64
		enNumber    string
		enCode      string
		body        []byte
		err         error
	)

	defer func() {
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	if enNumber, enCode, err = encryptAccount(
		request.AccountType, request.AccountNumber, request.Code,
	); err != nil {
		return
	}

	if result, err = config.DB.Exec(query,
		request.AccountName, enNumber, request.CardType, request.AccountType,
		request.AccountZip, enCode, request.ExpireMonth, request.ExpireYear,
		request.Id, request.UserId,
	); err != nil {
		return
	} else if affectedRow, err = result.RowsAffected(); err != nil {
		return
	} else if affectedRow <= 0 {
		err = errors.New("Payment not found")
		return
	}

	if payment, err := getPaymentByIdAndUserId(
		request.Id, request.UserId,
	); err == nil {
		c.IndentedJSON(http.StatusOK, payment)
	}
}

func DeletePayment(c *gin.Context) {
	query := `
		delete from user_payment where id = ? and user_id = ?
	`
	request := modules.PaymentRequest{}
	var (
		result      sql.Result
		affectedRow int64
		body        []byte
		err         error
	)

	defer func() {
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if body, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		return
	}

	if status, msg := checkPaymentStatus(
		request.Id, request.UserId,
	); status {
		err = errors.New(msg)
		return
	}

	if result, err = config.DB.Exec(
		query, request.Id, request.UserId,
	); err != nil {
		return
	} else if affectedRow, err = result.RowsAffected(); err != nil {
		return
	} else if affectedRow <= 0 {
		err = errors.New("Payment not found")
		c.Abort()
	}

	c.IndentedJSON(http.StatusOK, "OK")
}

func checkPaymentStatus(id, userId int32) (bool, string) {
	var temp int32

	query := `
		select reserved from user_payment
		where id = ? and user_id = ?
	`

	if err := config.DB.QueryRow(
		query, id, userId,
	).Scan(&temp); err != nil {
		fmt.Println("Check Payment Status - Error - ", err.Error())
		return true, err.Error()
	} else if temp > 0 {
		return true, `
			This payment method cannot be deleted since you have one reservation on it.
		`
	}

	query = `
		select count(*)
		from user_service us
		inner join user_payment up on up.id = us.user_payment_id and up.user_id = us.user_id
		where up.id = ? and up.user_id = ? and us.paid = 0 and us.status = 'DONE'
	`

	if err := config.DB.QueryRow(
		query, id, userId,
	).Scan(&temp); err != nil {
		fmt.Println("Check Payment Status - Error - ", err.Error())
		return true, err.Error()
	} else if temp > 0 {
		return true, `
			This payment method cannot be deleted since you need to process your payment.
		`
	}

	return false, ""
}

func lockPayment(tx *sql.Tx, id, userId int32) (err error) {
	query := `
		update user_payment set reserved = reserved + 1 where id = ? and user_id = ?
	`

	if _, err = config.DB.Exec(
		query, id, userId,
	); err != nil {
		fmt.Println("Lock Pyment - Error - ", err.Error())
	}

	return
}

func unlockPayment(tx *sql.Tx, id, userId int32) (err error) {
	query := `
		update user_payment set reserved = reserved - 1 where id = ? and user_id = ?
	`

	if _, err = tx.Exec(
		query, id, userId,
	); err != nil {
		fmt.Println("Unlock Pyment - Error - ", err.Error())
	}

	return
}

func ProcessPayment(c *gin.Context) {
	query := `
		select up.id, us.estimated_price, up.account_number, up.account_zip,
				up.code, up.expire_month, up.expire_year, up.account_type
		from user_service us
		inner join user_payment up on up.id = us.user_payment_id and up.user_id = us.user_id
		where us.id = ? and up.id = ? and us.user_id = ? and us.status = 'DONE'
	`
	request := modules.ProcessRequest{}
	process := struct {
		PaymentId     int32
		Price         float32
		Code          string
		Zip           string
		Year          string
		Month         string
		AccountNumber string
		AccountType   string
	}{}
	var (
		data []byte
		err  error
		tx   *braintree.Transaction
	)

	defer func() {
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			c.Abort()
		}
	}()

	if data, err = ioutil.ReadAll(c.Request.Body); err != nil {
		return
	}

	if err = json.Unmarshal(data, &request); err != nil {
		return
	}

	if err = config.DB.QueryRow(
		query, request.ServiceId, request.PaymentId, request.UserId,
	).Scan(
		&process.PaymentId, &process.Price, &process.AccountNumber,
		&process.Zip, &process.Code, &process.Month, &process.Year,
		&process.AccountType,
	); err != nil {
		if err == sql.ErrNoRows {
			err = errors.New("UserService (payment) not found")
		}

		return
	}

	if process.AccountNumber, process.Code, err =  decryptAccount(
		process.AccountType, process.AccountNumber, process.Code,
	); err != nil {
		err = errors.New("Cannot process payment now " + err.Error())
		return
	}

	fmt.Println("Process Payment ------ Start")
	fmt.Println("Decimal - ", int64(process.Price*100))
	fmt.Println("Number - ", process.AccountNumber)
	fmt.Println("CVV - ", process.Code)
	fmt.Println("ExpirationMonth - ", process.Month)
	fmt.Println("ExpirationYear - ", process.Year[2:])
	fmt.Println("Process Payment ------ End\n")

	if tx, err = config.BT.Transaction().Create(
		&braintree.Transaction{
			Type:   "sale",
			Amount: braintree.NewDecimal(int64(process.Price*100), 2),
			CreditCard: &braintree.CreditCard{
				Number:          process.AccountNumber,
				CVV:             process.Code,
				ExpirationMonth: process.Month,
				ExpirationYear:  process.Year[2:],
			},
		},
	); err != nil {
		err = errors.New("Cannot process payment now - " + err.Error())
		return
	}

	makeServicePaid(request.UserId, request.ServiceId, request.PaymentId)

	fmt.Println("Transaction Info - ", tx)

	c.IndentedJSON(http.StatusOK, "OK")
}
