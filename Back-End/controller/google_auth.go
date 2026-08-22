package controller

import (
	"Frota/models"
	"Frota/services"
	"Frota/structs"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOauthConfig *oauth2.Config
var FRONTEND_WEB_URL = "http://127.0.0.1:5500/my-app/www"
var WEB_URL = "http://127.0.0.1:5500"

// ConfigurarGoogleOAuth deve ser chamado no main.go
func ConfigurarGoogleOAuth() {
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_CALLBACK_URL"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

// ====================================================================
// FLUXO 1: WEB TRADICIONAL (Com redirecionamentos e Cookies)
// ====================================================================

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL("estado-aleatorio")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// URL base do seu Live Server (ou do site em produção)
	frontendWebURL := os.Getenv("FRONTEND_WEB_URL")
	if frontendWebURL == "" {
		frontendWebURL = "http://127.0.0.1:5500/my-app/www" // Endereço do seu Live Server
	}

	estado := r.FormValue("state")
	if estado != "estado-aleatorio" {
		http.Redirect(w, r, frontendWebURL+"/Usuario/login.html?erro=Estado+Invalido", http.StatusSeeOther)
		return
	}

	codigo := r.FormValue("code")
	tokenGo, err := googleOauthConfig.Exchange(context.Background(), codigo)
	if err != nil {
		http.Redirect(w, r, frontendWebURL+"/Usuario/login.html?erro=Falha+no+Google", http.StatusSeeOther)
		return
	}

	// Busca os dados do usuário usando o token do Google
	resposta, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + tokenGo.AccessToken)
	if err != nil {
		http.Redirect(w, r, frontendWebURL+"/Usuario/login.html?erro=Falha+ao+obter+dados", http.StatusSeeOther)
		return
	}
	defer resposta.Body.Close()

	var dadosGoogle map[string]interface{}
	json.NewDecoder(resposta.Body).Decode(&dadosGoogle)

	email := strings.ToLower(strings.TrimSpace(dadosGoogle["email"].(string)))
	nome := dadosGoogle["name"].(string)
	
	foto := ""
	if p, ok := dadosGoogle["picture"].(string); ok {
		foto = p
	}

	// 1. Tenta buscar o usuário no banco de dados
	usuario, err := models.BuscarUsuarioPorEmail(email)

	// 2. Se NÃO encontrou (Usuário Novo): manda para a tela de Completar Cadastro no Live Server!
	if err != nil {
		// Passa o e-mail, nome e foto via URL para a tela estática capturar
		urlCompletar := fmt.Sprintf("%s/Usuario/completar_cadastro.html?email=%s&nome=%s&foto=%s", frontendWebURL, url.QueryEscape(email), url.QueryEscape(nome), url.QueryEscape(foto))
		http.Redirect(w, r, urlCompletar, http.StatusSeeOther)
		return
	}

	// 3. Se ENCONTROU (Usuário Antigo): Gera o Token JWT
	tokenJWT, err := services.GerarToken(usuario.ID, usuario.Papel)
	if err != nil {
		http.Redirect(w, r, frontendWebURL+"/Usuario/login.html?erro=Erro+ao+gerar+token", http.StatusSeeOther)
		return
	}

	// 4. Redireciona para a Home correta no Live Server passando o token na URL para o JS salvar!
	var homeURL string
	if usuario.Papel == "admin" {
		homeURL = frontendWebURL + "/Admin/home_admin.html"
	} else if usuario.Papel == "motorista" {
		homeURL = frontendWebURL + "/Motorista/home_motorista.html"
	} else {
		homeURL = frontendWebURL + "/Passageiro/home_Passageiro.html"
	}

	// Redireciona com o token na URL
	http.Redirect(w, r, fmt.Sprintf("%s?token=%s", homeURL, tokenJWT), http.StatusSeeOther)
}

func CompletarCadastroGoogle(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Se estiver usando templates na web, renderiza aqui
		// temp.ExecuteTemplate(w, "CompletarCadastro", nil)
		return
	}

	aceitou := r.FormValue("aceitou") == "true" || r.FormValue("aceitou") == "on"

	cookie, err := r.Cookie("temp_google_data")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	partes := strings.Split(cookie.Value, "|")
	email := partes[0]
	nome := partes[1]
	whatsapp := r.FormValue("whatsapp")

	novoUsuario := structs.Usuario{
		Nome:          nome,
		Email:         email,
		Whatsapp:      whatsapp,
		Papel:         "passageiro",
		AceitouTermos: aceitou,
	}
	models.CriarUsuario(&novoUsuario)

	http.SetCookie(w, &http.Cookie{Name: "temp_google_data", MaxAge: -1, Path: "/"})

	tokenJWT, _ := services.GerarToken(novoUsuario.ID, novoUsuario.Papel)
	http.SetCookie(w, &http.Cookie{Name: "jwt_frota", Value: tokenJWT, Expires: time.Now().Add(24 * time.Hour), HttpOnly: true, Path: "/"})

	http.Redirect(w, r, "/passageiro/home", http.StatusSeeOther)
}

// ====================================================================
// FLUXO 2: APP NATIVO / API REST (JSON sem redirecionamento)
// ====================================================================

// Estruturas auxiliares para ler o JSON do App
type ReqGoogleLogin struct {
	Email string `json:"email"`
	Nome  string `json:"nome"`
	Foto  string `json:"foto"`
}

type ReqGoogleCompletar struct {
	Email    string `json:"email"`
	Nome     string `json:"nome"`
	Whatsapp string `json:"whatsapp"`
	Genero   string `json:"genero"`
	Foto     string `json:"foto"`
}

// ApiGoogleLogin recebe os dados do Google capturados pelo celular
func ApiGoogleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req ReqGoogleLogin
	json.NewDecoder(r.Body).Decode(&req)

	// Tenta achar o usuário
	usuario, err := models.BuscarUsuarioPorEmail(req.Email)

	// Se NÃO encontrou, avisa o Front-end para abrir a tela de completar WhatsApp
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"sucesso":           true,
			"precisa_completar": true, // A MÁGICA: O JS vai ler isso e trocar de tela
			"email":             req.Email,
			"nome":              req.Nome,
			"foto":              req.Foto,
		})
		return
	}

	// Se encontrou, gera o JWT e já faz o login automático!
	tokenJWT, _ := services.GerarToken(usuario.ID, usuario.Papel)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sucesso":           true,
		"precisa_completar": false,
		"token":             tokenJWT,
		"papel":             usuario.Papel,
	})
}

// ApiCompletarCadastroGoogle é chamada quando o usuário digita o WhatsApp na tela do App
func ApiCompletarCadastroGoogle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req ReqGoogleCompletar
	json.NewDecoder(r.Body).Decode(&req)

	if req.Whatsapp == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "WhatsApp é obrigatório"})
		return
	}

	genero := strings.TrimSpace(req.Genero)
	if genero == "" {
		genero = "Não Informado"
	}

	novoUsuario := structs.Usuario{
		Nome:          req.Nome,
		Email:         req.Email,
		Whatsapp:      req.Whatsapp,
		Papel:         "passageiro",
		Genero:        genero,
		FotoPerfil:    req.Foto,
		AceitouTermos: true,
	}

	err := models.CriarUsuario(&novoUsuario)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Erro ao criar conta."})
		return
	}

	// Gera o token
	tokenJWT, _ := services.GerarToken(novoUsuario.ID, novoUsuario.Papel)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sucesso": true,
		"token":   tokenJWT,
		"papel":   novoUsuario.Papel,
	})
}
