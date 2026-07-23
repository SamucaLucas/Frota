package services

import "golang.org/x/crypto/bcrypt"

// HashSenha recebe uma senha em texto plano e retorna o hash seguro para o banco
func HashSenha(senha string) (string, error) {
	// O custo 14 é um bom equilíbrio entre segurança e performance
	bytes, err := bcrypt.GenerateFromPassword([]byte(senha), 14)
	return string(bytes), err
}

// CompararSenha será usada na rota de login para verificar se a senha digitada bate com o hash
func CompararSenha(senhaHash, senhaTexto string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(senhaHash), []byte(senhaTexto))
	return err == nil
}
