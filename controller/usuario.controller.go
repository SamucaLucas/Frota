package controller

import (
	"html/template"
	"log"
	"net/http"
	"strings"
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
// CadastrarUsuario gerencia a página de criar conta
func CadastrarUsuario(w http.ResponseWriter, r *http.Request) {

	// --- CENÁRIO 1: Acessar a Página (GET) ---
	if r.Method == http.MethodGet {
		err := temp.ExecuteTemplate(w, "Cadastro", nil)
		if err != nil {
			log.Println("Erro ao renderizar tela de cadastro:", err)
		}
		return
	}

	// --- CENÁRIO 2: Enviar o Formulário (POST) ---
	if r.Method == http.MethodPost {
		// 1. Coleta os dados do HTML
		nome := r.FormValue("nome")
		email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
		senha := r.FormValue("senha")
		whatsapp := r.FormValue("whatsapp")
		// Melhoria: Aceita tanto "true" quanto "on" do checkbox HTML
		aceitou := r.FormValue("aceitou") == "true" || r.FormValue("aceitou") == "on"

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
			Nome:          nome,
			Email:         email,
			Senha:         senhaHash,
			Whatsapp:      whatsapp,
			Papel:         "passageiro", // Todo cadastro novo nasce como passageiro
			AceitouTermos: aceitou,
		}

		// 5. O Controller delega a gravação para o Model
		err = models.CriarUsuario(&usuario)
		if err != nil {
			// CORREÇÃO: Em vez de ir para a tela de Construção, volta para a tela de Cadastro com o aviso!
			dados := struct{ Erro string }{Erro: "Este e-mail já está cadastrado em nosso sistema."}
			temp.ExecuteTemplate(w, "Cadastro", dados)
			return
		}

		// 6. AUTO-LOGIN: Cria o Cookie de sessão (JWT)
		// Obs: Verifique no seu arquivo JWT.go se o nome da função é GerarToken ou GerarTokenJWT
		tokenString, errToken := services.GerarToken(usuario.ID, usuario.Papel)
		if errToken == nil {
			http.SetCookie(w, &http.Cookie{
				Name:     "token",
				Value:    tokenString,
				Path:     "/",
				HttpOnly: true,
			})
		} else {
			log.Println("Erro ao gerar token no auto-login:", errToken)
		}

		// 7. REDIRECIONAMENTO INTELIGENTE
		// Em vez de forçar a renderização falha, enviamos o usuário limpo para a rota oficial da Home
		http.Redirect(w, r, "/passageiro/home", http.StatusSeeOther)
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

	if r.Method == "POST" {
		// 1. Pega o e-mail, remove espaços vazios (comuns no celular) e força TUDO para minúsculo
		email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
		senha := r.FormValue("senha")
		lembrar := r.FormValue("lembrar")

		// 2. Validação simples
		if email == "" || senha == "" {
			dados := map[string]interface{}{"Erro": "Por favor, preencha o e-mail e a palavra-passe."}
			temp.ExecuteTemplate(w, "Login", dados)
			return
		}

		// 3. Procurar o utilizador na base de dados
		usuario, err :=  models.BuscarUsuarioPorEmail(email) // Agora vai buscar sempre minúsculo!
		if err != nil {
			dados := map[string]interface{}{"Erro": "E-mail ou palavra-passe inválidos."}
			temp.ExecuteTemplate(w, "Login", dados)
			return
		}

		// 3. Comparar a palavra-passe digitada com o Hash
		if !services.CompararSenha(usuario.Senha, senha) {
			dados := map[string]interface{}{"Erro": "E-mail ou palavra-passe inválidos."}
			temp.ExecuteTemplate(w, "Login", dados)
			return
		}

		// 4. Gerar o Token JWT
		token, err := services.GerarToken(usuario.ID, usuario.Papel)
		if err != nil {
			dados := map[string]interface{}{"Erro": "Erro interno ao iniciar a sessão."}
			temp.ExecuteTemplate(w, "Login", dados)
			return
		}

		// 5. Configurar a expiração do Cookie com base na escolha do utilizador
		tempoExpiracao := time.Hour * 24 // Padrão: 24 horas (se NÃO marcar a caixa)

		if lembrar == "on" {
			tempoExpiracao = time.Hour * 24 * 30 // 30 dias (se MARCAR a caixa)
		}

		// Gravar o Cookie no navegador do utilizador
		http.SetCookie(w, &http.Cookie{
			Name:     "token", // O nome do seu cookie de autenticação
			Value:    token,
			Expires:  time.Now().Add(tempoExpiracao),
			HttpOnly: true,  // Impede que scripts roubem o cookie
			Secure:   false, // Mude para 'true' quando colocar em produção com HTTPS
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})

		// 6. Redirecionar para a Home correta baseada no Papel do Utilizador
		if usuario.Papel == "admin" {
			http.Redirect(w, r, "/admin/home", http.StatusSeeOther)
		} else if usuario.Papel == "motorista" {
			http.Redirect(w, r, "/motorista/home", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/passageiro/home", http.StatusSeeOther)
		}
		return
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Sobrescreve o cookie JWT definindo o seu tempo de vida no passado
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-7 * 24 * time.Hour), // Define a expiração para 7 dias atrás
		MaxAge:   -1,                                  // Força a remoção imediata no navegador
		HttpOnly: true,
		Secure:   false, // Mudar para true quando usar HTTPS em produção
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	// Redireciona o utilizador de volta para a tela de login com status 303 (See Other)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
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
