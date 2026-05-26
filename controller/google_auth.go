package controller

import (
	"Frota/models"
	"Frota/services"
	"Frota/structs"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOauthConfig *oauth2.Config

// ConfigurarGoogleOAuth deve ser chamado no main.go para iniciar as variáveis
func ConfigurarGoogleOAuth() {
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_CALLBACK_URL"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

// GoogleLogin redireciona o usuário para a tela de login do Google
func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	// "estado-aleatorio" deveria ser uma string aleatória real em produção para segurança (CSRF)
	url := googleOauthConfig.AuthCodeURL("estado-aleatorio")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback recebe a resposta do Google após o usuário aceitar
func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	estado := r.FormValue("state")
	if estado != "estado-aleatorio" {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	codigo := r.FormValue("code")
	tokenGo, err := googleOauthConfig.Exchange(context.Background(), codigo)
	if err != nil {
		http.Redirect(w, r, "/login?erro=Falha+no+Google", http.StatusTemporaryRedirect)
		return
	}

	// Busca os dados do usuário usando o token do Google
	resposta, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + tokenGo.AccessToken)
	if err != nil {
		http.Redirect(w, r, "/login?erro=Falha+ao+obter+dados", http.StatusTemporaryRedirect)
		return
	}
	defer resposta.Body.Close()

	// Lê os dados que o Google enviou
	var dadosGoogle map[string]interface{}
	json.NewDecoder(resposta.Body).Decode(&dadosGoogle)

	email := dadosGoogle["email"].(string)
	nome := dadosGoogle["name"].(string)

	// =========================================================
	// A INTELIGÊNCIA DE REGISTRO E LOGIN ENTRA AQUI
	// =========================================================

	// 1. Tenta buscar o usuário no banco de dados
	usuario, err := models.BuscarUsuarioPorEmail(email)

	// 2. Se deu erro (não encontrou), vamos criar o usuário automaticamente!
	if err != nil {
		http.SetCookie(w, &http.Cookie{
			Name:     "temp_google_data",
			Value:    email + "|" + nome,
			Expires:  time.Now().Add(5 * time.Minute),
			Path:     "/",
			HttpOnly: true, // Boa prática de segurança
		})

		// ADICIONA ESTE RETURN!
		// Sem ele, o código continua a correr e tenta gerar o token abaixo
		http.Redirect(w, r, "/auth/google/completar", http.StatusSeeOther)
		return
	}

	// 3. O usuário agora existe (seja antigo ou recém-criado). Vamos gerar o JWT!
	tokenJWT, err := services.GerarToken(usuario.ID, usuario.Papel)
	if err != nil {
		http.Redirect(w, r, "/login?erro=Erro+ao+gerar+token", http.StatusTemporaryRedirect)
		return
	}

	// 4. Salva o JWT nos Cookies (Acesso seguro)
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_frota",
		Value:    tokenJWT,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
	})

	// 5. Manda para a tela de Construção!
	http.Redirect(w, r, "/construcao", http.StatusSeeOther)
}

// CompletarCadastroGoogle finaliza o registro após capturar o WhatsApp
func CompletarCadastroGoogle(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		temp.ExecuteTemplate(w, "CompletarCadastro", nil)
		return
	}
	aceitou := r.FormValue("aceitou") == "true"

	// 1. Recupera o cookie temporário
	cookie, err := r.Cookie("temp_google_data")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// 2. Extrai nome e e-mail (usando strings.Split ou similar)
	// Exemplo simplificado: "email|nome"
	partes := strings.Split(cookie.Value, "|")
	email := partes[0]
	nome := partes[1]
	whatsapp := r.FormValue("whatsapp")

	// 3. Cria o usuário finalmente
	novoUsuario := structs.Usuario{
		Nome:     nome,
		Email:    email,
		Whatsapp: whatsapp,
		Papel:    "passageiro",
		AceitouTermos: aceitou,
	}
	models.CriarUsuario(&novoUsuario)

	// 4. Limpa o cookie temporário
	http.SetCookie(w, &http.Cookie{Name: "temp_google_data", MaxAge: -1, Path: "/"})

	// 5. Gera o token e loga
	tokenJWT, _ := services.GerarToken(novoUsuario.ID, novoUsuario.Papel)
	http.SetCookie(w, &http.Cookie{
		Name: "jwt_frota", Value: tokenJWT, Expires: time.Now().Add(24 * time.Hour), HttpOnly: true, Path: "/",
	})

	http.Redirect(w, r, "/construcao", http.StatusSeeOther)
}
