package auth

import (
	"sync"
	"time"

	"github.com/Z3DRP/lessor-service/config"
	"github.com/golang-jwt/jwt/v4"
)

var (
	authToken []byte
	authErr   error
	once      sync.Once
)

var Expirey = time.Now().Add((24 * time.Hour) * 365)

func GetJwtKey() ([]byte, error) {
	return getKey()
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(id string, username string, role string) (string, error) {
	expirationTime := time.Now().Add(2 * time.Hour)
	//devTime := time.Now().Add((24 * time.Hour) * 365)

	claims := &Claims{
		Id:       id,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	key, err := getKey()

	if err != nil {
		return "", err

	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func getKey() ([]byte, error) {
	once.Do(func() {
		authToken, authErr = config.GetAuthToken()
	})
	return authToken, authErr
}
