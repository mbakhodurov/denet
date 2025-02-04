package middlewares

import (
	"denet/internal/lib/api/response"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/render"
)

var jwtSecret = []byte("secret_1234")

func GenerateJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,                             // Идентификатор пользователя
		"exp": time.Now().Add(time.Minute).Unix(), // Время жизни токена 1 минута
		"nbf": time.Now().Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func ValidateJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.uSERInfo.New"

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			render.JSON(w, r, response.Error("No Authorization header"))
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			render.JSON(w, r, response.Error("Invalid token"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
