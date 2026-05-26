package controller

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"Frota/models"
	"Frota/services"
	"Frota/structs"
)

// O padrão "views/*/*.html" diz ao Go para ler todos os arquivos HTML dentro de qualquer subpasta de views
var temp = template.Must(template.ParseGlob("views/*/*.html"))

// EmDesenvolvimento renderiza a tela de aviso de funcionalidades futuras
func EmDesenvolvimento(w http.ResponseWriter, r *http.Request) {
	err := temp.ExecuteTemplate(w, "Construcao", nil)
	if err != nil {
		log.Println("Erro ao renderizar tela de construção:", err)
	}
}

// CadastrarUsuario gerencia a página de criar conta
func CadastrarUsuario(w http.ResponseWriter, r *http.Request) {

	// --- CENÁRIO 1: Acessar a Página (GET) ---
	if r.Method == "GET" {
		// Agora chamamos o nome do template "Cadastro" que definimos no HTML
		err := temp.ExecuteTemplate(w, "Cadastro", nil)
		if err != nil {
			log.Println("Erro ao renderizar tela de cadastro:", err)
		}
		return
	}

	// --- CENÁRIO 2: Enviar o Formulário (POST) ---
	if r.Method == "POST" {
		// 1. Coleta os dados do HTML
		nome := r.FormValue("nome")
		email := r.FormValue("email")
		senha := r.FormValue("senha")
		whatsapp := r.FormValue("whatsapp")
		aceitou := r.FormValue("aceitou") == "true"

		// 2. Validação simples
		if nome == "" || email == "" || senha == "" || whatsapp == "" {
			dados := struct{ Erro string }{Erro: "Por favor, preencha todos os campos obrigatórios."}
			temp.ExecuteTemplate(w, "Cadastro", dados)
			return
		}

		// 3. Criptografar a senha
		senhaHash, err := services.HashSenha(senha)
		if err != nil {
			dados := struct{ Erro string }{Erro: "Erro interno ao processar a senha."}
			temp.ExecuteTemplate(w, "Cadastro", dados)
			return
		}

		// 4. Montar a Struct
		usuario := structs.Usuario{
			Nome:     nome,
			Email:    email,
			Senha:    senhaHash,
			Whatsapp: whatsapp,
			Papel:    "passageiro", // Papel padrão
			AceitouTermos: aceitou,   // Converte "on" para true
		}

		// 5. O Controller delega a gravação para o Model
		err = models.CriarUsuario(&usuario)
		if err != nil {
			dados := struct{ Erro string }{Erro: "Este e-mail já está cadastrado em nosso sistema."}
			temp.ExecuteTemplate(w, "Construcao", dados)
			return
		}

		// 6. Retorno de Sucesso para a View
		dados := struct{ Sucesso string }{Sucesso: "Conta criada com sucesso! Você já pode fazer login."}
		temp.ExecuteTemplate(w, "Construcao", dados)
		return
	}
}

func LoginUsuario(w http.ResponseWriter, r *http.Request) {

	// --- CENÁRIO 1: Aceder à Página (GET) ---
	if r.Method == "GET" {
		// Garanta que este nome "Login" bate com o {{define "Login"}} do HTML
		err := temp.ExecuteTemplate(w, "Login", nil)
		if err != nil {
			log.Println("Erro ao renderizar ecrã de login:", err)
		}
		return
	}

	// --- CENÁRIO 2: Enviar o Formulário (POST) ---
	if r.Method == "POST" {
		email := r.FormValue("email")
		senha := r.FormValue("senha")

		// 1. Validação simples
		if email == "" || senha == "" {
			dados := struct{ Erro string }{Erro: "Por favor, preencha o e-mail e a palavra-passe."}
			temp.ExecuteTemplate(w, "Login", dados)
			return
		}

		// 2. Procurar o utilizador na base de dados (através do Model)
		usuario, err := models.BuscarUsuarioPorEmail(email)
		if err != nil {
			// Não dizemos se o erro foi no e-mail ou na palavra-passe por questões de segurança
			dados := struct{ Erro string }{Erro: "E-mail ou palavra-passe inválidos."}
			temp.ExecuteTemplate(w, "Login", dados)
			return
		}

		// 3. Comparar a palavra-passe digitada com o Hash guardado no PostgreSQL
		if !services.CompararSenha(usuario.Senha, senha) {
			dados := struct{ Erro string }{Erro: "E-mail ou palavra-passe inválidos."}
			temp.ExecuteTemplate(w, "Login", dados)
			return
		}

		// 4. Gerar o Token JWT com o ID e o Papel do utilizador (Passageiro, Motorista ou Admin)
		token, err := services.GerarToken(usuario.ID, usuario.Papel)
		if err != nil {
			dados := struct{ Erro string }{Erro: "Erro interno ao iniciar a sessão."}
			temp.ExecuteTemplate(w, "Login", dados)
			return
		}

		// 5. Guardar o Token de forma segura nos Cookies do navegador
		http.SetCookie(w, &http.Cookie{
			Name:     "jwt_frota",                    // Nome do cookie
			Value:    token,                          // O nosso token gerado
			Expires:  time.Now().Add(24 * time.Hour), // Expira em 24 horas
			HttpOnly: true,                           // Fundamental: Impede acesso via JavaScript (Segurança extra)
			Secure:   false,                          // Defina como 'true' quando for para produção com HTTPS
			Path:     "/",                            // Disponível em todo o sistema
		})

		// 6. Retorno de Sucesso (Futuramente, faremos um http.Redirect para o painel de agendamento)
		//dados := struct{ Sucesso string }{Sucesso: "Login efetuado com sucesso! Bem-vindo(a), " + usuario.Nome}
		http.Redirect(w, r, "/construcao", http.StatusSeeOther)
		return
	}
}

// TermosUsuario renderiza a tela de políticas de uso do sistema
func TermosUsuario(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		err := temp.ExecuteTemplate(w, "Termos", nil)
		if err != nil {
			log.Println("Erro ao renderizar tela de termos:", err)
		}
		return
	}
}
