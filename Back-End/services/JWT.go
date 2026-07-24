package services

import (
	"errors"
	"net/http"
	"os"
	"strings" 
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GerarToken cria o crachá digital contendo o ID e o Papel do usuário
func GerarToken(usuarioID uint, papel string) (string, error) {
	// Puxa a chave secreta que criamos lá no nosso arquivo .env
	chaveSecreta := []byte(os.Getenv("JWT_SECRET"))

	// Cria os dados que vão dentro do token (Payload)
	claims := jwt.MapClaims{
		"id":    usuarioID,
		"papel": papel,
		"exp":   time.Now().Add(time.Hour * 24 * 30).Unix(), // Token expira em 30 dias
	}

	// Gera o token assinado
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(chaveSecreta)
}

// ExtrairUsuarioID lê o Header Authorization ou o Cookie JWT e devolve o ID do usuário logado
func ExtrairUsuarioID(r *http.Request) (uint, error) {
	tokenString := ""

	// 1. TENTA LER DO CABEÇALHO (Padrão para App Nativo/API REST)
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenString = strings.TrimPrefix(authHeader, "Bearer ")
	}

	// 2. SE NÃO ACHOU, TENTA LER DO COOKIE (Padrão Web Antigo)
	if tokenString == "" {
		cookie, err := r.Cookie("jwt_frota")
		if err == nil {
			tokenString = cookie.Value
		}
	}

	// Se não achou em nenhum dos dois lugares, barra a requisição
	if tokenString == "" {
		return 0, errors.New("token de autenticação não encontrado")
	}

	// 3. Valida e faz o parse da string do Token que encontramos
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil // Use a sua variável secreta aqui
	})
	if err != nil {
		return 0, err
	}

	// 4. Extrai o ID se o token for válido
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		idFloat, ok := claims["id"].(float64) // O JSON converte números para float64
		if !ok {
			return 0, errors.New("ID inválido no token")
		}
		return uint(idFloat), nil
	}

	return 0, errors.New("token inválido")
}