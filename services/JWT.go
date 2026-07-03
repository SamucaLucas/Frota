package services

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GerarToken cria o crachá digital contendo o ID e o Papel do usuário
func GerarToken(usuarioID uint, papel string) (string, error) {
	// Puxa a chave secreta que criamos lá no nosso arquivo .env
	chaveSecreta := []byte(os.Getenv("JWT_SECRET"))

	// Cria os dados que vão dentro do token (Payload)
	claims := jwt.MapClaims{
		"id": usuarioID,
		"papel":      papel,
		"exp":        time.Now().Add(time.Hour * 24).Unix(), // Token expira em 24 horas
	}

	// Gera o token assinado
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(chaveSecreta)
}

// ExtrairUsuarioID lê o cookie JWT e devolve o ID do usuário logado
func ExtrairUsuarioID(r *http.Request) (uint, error) {
	cookie, err := r.Cookie("jwt_frota")
	if err != nil {
		return 0, err
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil // Use a sua variável secreta aqui
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		idFloat, ok := claims["id"].(float64) // O JSON converte números para float64
		if !ok {
			return 0, errors.New("ID inválido no token")
		}
		return uint(idFloat), nil
	}
	return 0, errors.New("Token inválido")
}
