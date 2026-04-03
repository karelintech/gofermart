// Package auth нужен для авторизации
package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
)

// Claims для JWT
type Claims struct {
	UserID int `json:"id"`
	jwt.RegisteredClaims
}

var jwtKey []byte

func init() {
	godotenv.Load()
	jwtKey = []byte(os.Getenv("jwtKey"))
}

// HashFunc - хэширует исходную строку
func HashFunc(rowString string) string {
	rowBytes := sha256.Sum256([]byte(rowString))
	return hex.EncodeToString(rowBytes[:])
}

// GetJWTToken генерирует токен
func GetJWTToken(userID int) (string, error) {
	expirationTime := time.Now().Add(time.Hour)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Генерация токена
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подпись токена
	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// Middleware - проверка токена
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("Token")
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
			} else {
				http.Error(w, "Invalid cookie", http.StatusBadRequest)
			}
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrInvalidKey
			}
			return jwtKey, nil
		})
		if err != nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID - получение UserID из Cookie
func GetUserID(r *http.Request) (int, error) {
	cookie, err := r.Cookie("UserID")
	if err != nil {
		return -1, err
	}
	cookieUserID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		return -1, err
	}
	return cookieUserID, nil
}
