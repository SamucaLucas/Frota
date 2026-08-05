package routers

import (
	"net/http"
	"strings"

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

	r.HandleFunc("/api/corrida/detalhes", controller.ObterDetalhesCorrida)
	r.HandleFunc("/api/configuracoes", controller.ObterConfiguracoes)
	r.HandleFunc("/api/usuario/deletar", controller.DeletarUsuario)

	//Rotas de API - Passageiro
	r.HandleFunc("/api/passageiro/home", controller.HomePassageiro)
	r.HandleFunc("/api/passageiro/agendar", controller.AgendarViagem)

	// Rotas de API - Admin
	r.HandleFunc("/api/admin/home", controller.HomeAdmin)
	r.HandleFunc("/api/admin/despachar/", controller.ApiGetDespachar) // GET com o ID no final
	r.HandleFunc("/api/admin/atribuir", controller.ApiPostAtribuir)   // POST salvando os dados
	r.HandleFunc("/api/admin/motoristas/localizacao", controller.BuscarLocalizacoesAdmin)
	r.HandleFunc("/api/admin/nova-chamada", controller.NovaChamada)

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

	//Rota de anuncio do admin

	// Rotas do Admin (Abaixo dos seus middlewares de proteção de Admin)
	r.HandleFunc("/api/admin/anuncios", controller.GerenciarAnunciosAdmin)  // GET e POST
	r.HandleFunc("/api/admin/anuncios/", controller.GerenciarAnunciosAdmin) // GET e POST

	// Rotas do Passageiro / Públicas
	r.HandleFunc("/api/anuncios/ativos", controller.ListarAnunciosAtivos)

	r.HandleFunc("/api/anuncios/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/clique") {
			controller.RegistrarClique(w, r)
		}
	})

	// Libera o acesso público às imagens dos anúncios
	fsUploads := http.FileServer(http.Dir("./uploads"))
	r.Handle("/uploads/", http.StripPrefix("/uploads/", fsUploads))

	return r
}
