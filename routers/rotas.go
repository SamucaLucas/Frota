package routers

import (
	"net/http"

	"Frota/controller"
)

func ConfigurarRotas() *http.ServeMux {

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/construcao", controller.EmDesenvolvimento)
	// --- Rotas de Usuário (Passageiro) ---

	mux.HandleFunc("/cadastrar", controller.CadastrarUsuario)
	mux.HandleFunc("/login", controller.LoginUsuario)
	mux.HandleFunc("/termos", controller.TermosUsuario)

	// Rotas do Google
	mux.HandleFunc("/auth/google/login", controller.GoogleLogin)
	mux.HandleFunc("/auth/google/callback", controller.GoogleCallback)
	mux.HandleFunc("/auth/google/completar", controller.CompletarCadastroGoogle)


	return mux
}
