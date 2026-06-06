package routers

import (
	"net/http"

	"Frota/controller"
)

func ConfigurarRotas() *http.ServeMux {

	r := http.NewServeMux()

	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/", http.StripPrefix("/static/", fs))

	r.HandleFunc("/construcao", controller.EmDesenvolvimento)
	// --- Rotas de Usuário (Passageiro) ---

	r.HandleFunc("/cadastrar", controller.CadastrarUsuario)
	r.HandleFunc("/login", controller.LoginUsuario)
	r.HandleFunc("/termos", controller.TermosUsuario)

	// Rotas do Google
	r.HandleFunc("/auth/google/login", controller.GoogleLogin)
	r.HandleFunc("/auth/google/callback", controller.GoogleCallback)
	r.HandleFunc("/auth/google/completar", controller.CompletarCadastroGoogle)

	// Rotas do Passageiro
	r.HandleFunc("/passageiro/home", controller.HomePassageiro)
	r.HandleFunc("/passageiro/agendar", controller.AgendarViagem)

	// Rotas do Motorista
	r.HandleFunc("/motorista/home", controller.HomeMotorista)
	// Rotas do Admin
	r.HandleFunc("/admin/home", controller.HomeAdmin)
	r.HandleFunc("/admin/atribuir", controller.AtribuirCorrida)

	return r
}
