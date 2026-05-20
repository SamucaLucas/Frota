package routers

import (
	"net/http"

	"Frota/controller" 
)

func ConfigurarRotas() *http.ServeMux {

	mux := http.NewServeMux()

	// --- Rotas de Usuário (Passageiro) ---

	mux.HandleFunc("/cadastrar", controller.CadastrarUsuario)
	mux.HandleFunc("/login", controller.LoginUsuario)


	return mux
}