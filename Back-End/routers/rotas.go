package routers

import (
	"net/http"

	"Frota/controller"
)

func ConfigurarRotas() *http.ServeMux {

	r := http.NewServeMux()

	// Rota pública para Health Check / UptimeRobot
	r.HandleFunc("/api/ping", controller.Ping)

	// Rotas de API - Usuario
	r.HandleFunc("/api/login", controller.LoginUsuario)
	r.HandleFunc("/api/cadastrar", controller.CadastrarUsuario)
	r.HandleFunc("/api/google/login", controller.ApiGoogleLogin)
	r.HandleFunc("/api/google/completar", controller.ApiCompletarCadastroGoogle)
	r.HandleFunc("/api/verificar-token", controller.ApiVerificarToken)
	r.HandleFunc("/api/logout", controller.ApiLogout)

	//Rotas de API - Passageiro
	r.HandleFunc("/api/passageiro/home", controller.HomePassageiro)
	r.HandleFunc("/api/passageiro/agendar", controller.AgendarViagem)
	r.HandleFunc("/api/config-precos", controller.ApiConfigPrecos)

	// Rotas de API - Admin
	r.HandleFunc("/api/admin/home", controller.HomeAdmin)
	r.HandleFunc("/api/admin/despachar/", controller.ApiGetDespachar) // GET com o ID no final
	r.HandleFunc("/api/admin/atribuir", controller.ApiPostAtribuir)   // POST salvando os dados
	r.HandleFunc("/api/admin/motoristas/localizacao", controller.BuscarLocalizacoesAdmin)
	
	// Rotas de API - Motorista
	r.HandleFunc("/api/motorista/home", controller.ApiHomeMotorista)
	r.HandleFunc("/api/motorista/concluir", controller.ApiPostConcluirCorrida)
	r.HandleFunc("/api/motorista/localizacao", controller.AtualizarLocalizacao)
	r.HandleFunc("/api/motorista/corrida-livre", controller.SalvarCorridaLivre)
	r.HandleFunc("/api/motorista/buscar-passageiro", controller.BuscarPassageiroAvulso)
	// Rotas do Google
	r.HandleFunc("/auth/google/login", controller.GoogleLogin)
	r.HandleFunc("/auth/google/callback", controller.GoogleCallback)
	r.HandleFunc("/auth/google/completar", controller.CompletarCadastroGoogle)

	return r
}
