package handlers

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/gommon/log"
	"golang.org/x/crypto/bcrypt"
)

func isCredValid(givenPwd, storedPwd string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(storedPwd), []byte(givenPwd)); err != nil {
		return false
	}

	return true
}

func (u User) createToken() (string, error) {
	if err := cleanenv.ReadEnv(&prop); err != nil {
		log.Errorf("Configuration cannot be read : %v", err)
	}
	claims := jwt.MapClaims{}
	claims["authorized"] = u.IsAdmin
	claims["user_id"] = u.Email
	claims["exp"] = time.Now().Add(time.Minute * 15).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := at.SignedString([]byte(prop.JwtTokenSecret))
	if err != nil {
		log.Errorf("Unable to generate the token :%v", err)
		return "", err
	}
	return token, nil
}
