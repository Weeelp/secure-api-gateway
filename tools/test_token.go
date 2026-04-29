package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func main() {
	generateToken()
}

func generateToken() {
	secretKey := []byte("f3ff0b04e44eb99cefab7c2862440c8cd6581a4b07e798f420e0fa212c4c7964") // Тот же ключ, что в main.go

	claims := jwt.MapClaims{
		"user_id": float64(123),
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Токен на 24 часа
		"iss":     "cyber-bro",                           // Тот же issuer, что в middleware
		"jti":     uuid.New().String(),                   // Уникальный ID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	fmt.Println(tokenString)
}
