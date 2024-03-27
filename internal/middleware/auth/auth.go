package auth

import (
	"context"
	"fmt"
	"github.com/dubrovsky1/url-shortener/internal/models"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID
}

type (
	KeyType string
	Token   string
)

const (
	TokenExp                    = time.Hour * 3
	SecretKey  KeyType          = "supersecretkey"
	CookieName                  = "userid"
	KeyName    models.KeyUserID = "UserID"
)

func Auth(h http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var tokenString Token
		var userID uuid.UUID
		var errToken, errGetUserID error

		//нужно получить токен для расшифровки id пользователя
		//токен лежит в куке, если её нет - создаем новую, в которою записываем новый userID
		cookie, err := req.Cookie(CookieName)

		if err == http.ErrNoCookie {
			tokenString, errToken = BuildJWTString()
			if errToken != nil {
				http.Error(res, errToken.Error(), http.StatusUnauthorized)
				return
			}

			c := &http.Cookie{
				Name:     CookieName,
				Value:    string(tokenString),
				HttpOnly: true,
				Secure:   true,
			}

			http.SetCookie(res, c)
		} else {
			tokenString = Token(cookie.Value)
		}

		userID, errGetUserID = GetUserID(tokenString)
		if errGetUserID != nil {
			http.Error(res, errGetUserID.Error(), http.StatusUnauthorized)
			return
		}

		//Наследуем от контекста запроса новый контекст и записываем в него полученный или новый UserID
		authContext := context.WithValue(req.Context(), KeyName, userID)
		h.ServeHTTP(res, req.WithContext(authContext))
	}
}

func BuildJWTString() (Token, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		// собственное утверждение
		UserID: uuid.New(),
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return Token(tokenString), nil
}

func GetUserID(tokenString Token) (uuid.UUID, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(string(tokenString), claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(SecretKey), nil
		})
	if err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}
