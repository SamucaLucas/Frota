package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	"Frota/models"
	"Frota/services"
)

//API Rest

type RequisicaoLogin struct {
	Email   string `json:"email"`
	Senha   string `json:"senha"`
	Lembrar bool   `json:"lembrar"` // Opcional
}

func LoginUsuario(w http.ResponseWriter, r *http.Request) {
	// 1. Avisar ao Front-end que a resposta será em formato JSON
	w.Header().Set("Content-Type", "application/json")

	// 2. Garantir que é um POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Método não permitido"})
		return
	}

	// 3. Ler o JSON enviado pelo aplicativo (Celular)
	var req RequisicaoLogin
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Formato de dados inválido."})
		return
	}

	// 4. Limpar o e-mail (Tudo minúsculo e sem espaços)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	senha := req.Senha

	// 5. Validação simples
	if email == "" || senha == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Por favor, preencha o e-mail e a palavra-passe."})
		return
	}

	// 6. Procurar o utilizador na base de dados
	usuario, err := models.BuscarUsuarioPorEmail(email)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "E-mail ou palavra-passe inválidos."})
		return
	}

	// 7. Comparar a palavra-passe digitada com o Hash
	if !services.CompararSenha(usuario.Senha, senha) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "E-mail ou palavra-passe inválidos."})
		return
	}

	// 8. Gerar o Token JWT
	token, err := services.GerarToken(usuario.ID, usuario.Papel)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Erro interno ao iniciar a sessão."})
		return
	}

	// 9. Devolver o Sucesso, o Token e o Papel para o Celular!
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sucesso": true,
		"token":   token,
		"papel":   usuario.Papel,
	})
}

/*
//Golang

// O padrão "views.html" diz ao Go para ler todos os arquivos HTML dentro de qualquer subpasta de views
var temp = template.Must(template.ParseGlob("views/*.html"))

// EmDesenvolvimento renderiza a tela de aviso de funcionalidades futuras
func EmDesenvolvimento(w http.ResponseWriter, r *http.Request) {
	err := temp.ExecuteTemplate(w, "Construcao", nil)
	if err != nil {
		log.Println("Erro ao renderizar tela de construção:", err)
	}
}

// IndexHandler cuida da rota raiz "/" e redireciona o usuário de forma inteligente
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// Garante que só vai tratar o "/" exato e não caminhos inexistentes
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// 1. Tenta ler o cookie de sessão
	cookie, err := r.Cookie("jwt_frota")
	if err != nil {
		// Se o cookie não existir, redireciona direto para o login
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// 2. Se o cookie existe, vamos validar o JWT para saber o papel (role) do usuário
	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	// 3. Se o token for válido, extrai o papel e joga o usuário para a sua respectiva Home
	if err == nil && token.Valid {
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			papel, okPapel := claims["papel"].(string)
			if okPapel {
				// Usamos strings.ToLower para ignorar maiúsculas/minúsculas vindas do banco
				switch strings.ToLower(strings.TrimSpace(papel)) {
				case "admin":
					http.Redirect(w, r, "/admin/home", http.StatusSeeOther)
					return
				case "motorista":
					http.Redirect(w, r, "/motorista/home", http.StatusSeeOther)
					return
				default:
					http.Redirect(w, r, "/passageiro/home", http.StatusSeeOther)
					return
				}
			}
		}
	}

	// 4. Se o token estiver expirado ou corrompido, limpa e manda para o login
	http.Redirect(w, r, "/login", http.StatusSeeOther)
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
				Name:     "jwt_frota",
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
		usuario, err := models.BuscarUsuarioPorEmail(email) // Agora vai buscar sempre minúsculo!
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
		var cookie *http.Cookie

		if lembrar == "on" {
			// SE MARCOU: Definimos Expires para daqui a 30 dias (Cookie persistente)
			cookie = &http.Cookie{
				Name:     "jwt_frota",
				Value:    token,
				Expires:  time.Now().Add(time.Hour * 24 * 30),
				HttpOnly: true,
				Secure:   false, // Defina como true em produção (com HTTPS)
				Path:     "/",
				SameSite: http.SameSiteLaxMode,
			}
		} else {
			// SE NÃO MARCOU: Omitimos o campo Expires (Cookie de Sessão)
			// Este cookie será destruído automaticamente assim que o App/PWA for fechado
			cookie = &http.Cookie{
				Name:     "jwt_frota",
				Value:    token,
				HttpOnly: true,
				Secure:   false, // Defina como true em produção (com HTTPS)
				Path:     "/",
				SameSite: http.SameSiteLaxMode,
			}
		}

		// Gravar o Cookie estruturado no navegador/PWA do utilizador
		http.SetCookie(w, cookie)

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
		Name:     "jwt_frota",
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
*/
