package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"secure-api-gateway/internal/cache"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// JWTAuthMiddleware - middleware для проверки JWT токенов
func JWTAuthMiddleware(secretKey []byte, rds *cache.Redis) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//Получаем токен из заголовка
			authHeader := r.Header.Get("Authorization")
			//проверка доступа
			if authHeader == "" {
				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			// Проверяем формат: "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			//Парсим и проверяем токен
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
				// Проверяем алгоритм подписи
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return secretKey, nil
			})

			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			//Валидация токена
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				// Проверка истечения срока действия (expiry)
				if exp, ok := claims["exp"].(float64); ok {
					if time.Now().Unix() > int64(exp) {
						http.Error(w, "Token expired", http.StatusUnauthorized)
						return
					}
				}

				// Проверка issuer (если нужно)
				if iss, ok := claims["iss"].(string); ok {
					if iss != "cyber-bro" { // Замени на свой issuer
						http.Error(w, "Invalid issuer", http.StatusUnauthorized)
						return
					}
				}

				//Извлечение user_id
				userID, ok := claims["user_id"].(float64) // JWT хранит числа как float64
				if !ok {
					http.Error(w, "Invalid user_id claim", http.StatusUnauthorized)
					return
				}

				// Сохраняем user_id в контексте запроса
				ctx := context.WithValue(r.Context(), UserIDKey, userID)
				r = r.WithContext(ctx)

				//Защита от повторного использования (basic replay protection)
				if jti, ok := claims["jti"].(string); ok {

					isUsed, err := rds.GetEngine().Exists(context.Background(), "used_tokens:"+jti).Result()

					if err != nil {
						http.Error(w, "Internal error", http.StatusInternalServerError)
						return
					}

					if isUsed == 1 {
						http.Error(w, "Token already used", http.StatusUnauthorized)
						return
					}

					// Помечаем токен как использованный
					expTime := int64(claims["exp"].(float64)) // Приводим exp к int64
					nowTime := time.Now().Unix()              // Уже int64
					ttl := time.Duration(expTime-nowTime) * time.Second
					_, err = rds.GetEngine().Set(context.Background(), "used_tokens:"+jti, true, ttl).Result()
					if err != nil {
						http.Error(w, "Internal error", http.StatusInternalServerError)
						return
					}
				} else {
					http.Error(w, "Invalid jti", http.StatusUnauthorized)
					return
				}
			}
		})
	}
}
