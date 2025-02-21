package auth

import (
	"github.com/Z3DRP/lessor-service/config"
	"golang.org/x/crypto/bcrypt"
)

func HashString(s string) (string, error) {
	salty, err := config.GetSalty()
	if err != nil {
		return "", err
	}
	saltedNuts := salty + s
	bytes, err := bcrypt.GenerateFromPassword([]byte(saltedNuts), 14)
	return string(bytes), err
}

func VerifyHash(hash string, plainTxt string) (bool, error) {
	salty, err := config.GetSalty()
	if err != nil {
		return false, err
	}
	saltedNuts := salty + plainTxt
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(saltedNuts))
	return err == nil, nil
}
