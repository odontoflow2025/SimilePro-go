package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	privateKeyBytes, err := os.ReadFile("private.pem")
	if err != nil {
		panic(err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		panic(err)
	}

	claims := jwt.MapClaims{
		"sub":         float64(1),
		"clinicaId":   float64(1),
		"tipoUsuario": "ADMIN_TOTAL", // Bypassa a verificação de vínculo
		"exp":         time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		panic(err)
	}

	fmt.Print(tokenString)
}
