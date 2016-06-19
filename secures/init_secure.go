package secures

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	mr "math/rand"
	"time"
)

const (
	PASSWORD_KEY   = "23nIR19CYqStxfetIske2oRNBMDpFqAq"
	CREDIT_KEY     = "Gd9uoxQksPnM3qiheoPXIFg4R7hNAw9k"
	CREDIT_CVV_KEY = "PBoR6BdFdtJTaaGP51PSzlWvkYl7Qc5l"
	DEBIT_KEY      = "Luo1XtcQWQopSl451wALrp4Iblw2LuuX"
	DEBIT_PIN_KEY  = "eP9535CDBuFqvGZXOQSR5vxaIXB0Ww2O"

	LETTERS = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func encrypt(content, secret string) (code string, err error) {
	text := []byte(content)
	key := []byte(secret)

	block, err := aes.NewCipher(key)

	if err != nil {
		return "", err
	}

	b := base64.StdEncoding.EncodeToString(text)

	ciphertext := make([]byte, aes.BlockSize+len(b))
	prefix := ciphertext[:aes.BlockSize]

	if _, err = io.ReadFull(rand.Reader, prefix); err != nil {
		return "", err
	}

	cfb := cipher.NewCFBEncrypter(block, prefix)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))

	return hex.EncodeToString(ciphertext), nil
}

func decrypt(content, secret string) (text string, err error) {
	code, err := hex.DecodeString(content)

	if err != nil {
		return "", err
	}

	key := []byte(secret)
	block, err := aes.NewCipher(key)

	if err != nil {
		return "", err
	}

	if len(code) < aes.BlockSize {
		return "", errors.New("Text is too short")
	}

	prefix := code[:aes.BlockSize]
	ciphertext := code[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, prefix)
	cfb.XORKeyStream(ciphertext, ciphertext)

	data, err := base64.StdEncoding.DecodeString(string(ciphertext))

	if err != nil {
		return "", err
	}

	return string(data), nil
}

func RandString(n int) string {
	mr.Seed(time.Now().UTC().UnixNano())

	size := len(LETTERS)
	str := make([]byte, n)

	for i := range str {
		str[i] = LETTERS[mr.Intn(size)]
	}

	return string(str)
}
