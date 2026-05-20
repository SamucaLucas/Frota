package models

import (
	"Frota/db"
	"Frota/structs"
)

// CriarUsuario recebe a struct já preenchida do Controller e salva no banco de dados
func CriarUsuario(usuario *structs.Usuario) error {
	// Apenas o Model tem o direito de chamar db.DB
	resultado := db.DB.Create(usuario)
	return resultado.Error
}

// BuscarUsuarioPorEmail busca um usuário no banco de dados para validar o login
func BuscarUsuarioPorEmail(email string) (structs.Usuario, error) {
	var usuario structs.Usuario
	// O GORM busca o primeiro registro que bater com o e-mail informado
	resultado := db.DB.Where("email = ?", email).First(&usuario)
	return usuario, resultado.Error
}
