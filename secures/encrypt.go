package secures

func EncryptPassword(password string) (string, error) {
	return encrypt(password, PASSWORD_KEY)
}

func EncryptCredit(credit string) (string, error) {
	return encrypt(credit, CREDIT_KEY)
}

func EncryptCreditCVV(cvv string) (string, error) {
	return encrypt(cvv, CREDIT_CVV_KEY)
}

func EncryptDebit(debit string) (string, error) {
	return encrypt(debit, DEBIT_KEY)
}

func EncryptDebitPin(pin string) (string, error) {
	return encrypt(pin, DEBIT_PIN_KEY)
}

