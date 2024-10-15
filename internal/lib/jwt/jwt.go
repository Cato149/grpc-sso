package jwt

import (
	"awesomeProject/internal/domain/model"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// TODO: Покрыть тестами
func NewToken(user model.User, app model.App, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(duration).Unix()
	claims["app_id"] = app.ID

	tokenStr, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}
