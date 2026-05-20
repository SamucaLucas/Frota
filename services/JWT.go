package services

import (
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
		"usuario_id": usuarioID,
		"papel":      papel,
		"exp":        time.Now().Add(time.Hour * 24).Unix(), // Token expira em 24 horas
	}

	// Gera o token assinado
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(chaveSecreta)
}
