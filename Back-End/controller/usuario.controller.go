package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"Frota/db"
	"Frota/models"
	"Frota/services"
	"Frota/structs"

	"github.com/golang-jwt/jwt/v5"
)

//API Rest

type RequisicaoLogin struct {
	Email   string `json:"email"`
	Senha   string `json:"senha"`
	Lembrar bool   `json:"lembrar"` // Opcional
}

// ApiVerificarToken valida o Token JWT do usuário (Web ou App) e devolve seu perfil em JSON
func ApiVerificarToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tokenString := ""

	// 1. Tenta pegar o Token do Header "Authorization: Bearer <TOKEN>" (Padrão para App Nativo/API)
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenString = strings.TrimPrefix(authHeader, "Bearer ")
	}

	// 2. Fallback: Se não veio no Header, tenta ler do Cookie (Para fluxo Web tradicional)
	if tokenString == "" {
		cookie, err := r.Cookie("jwt_frota")
		if err == nil {
			tokenString = cookie.Value
		}
	}

	// Se não encontrou o token em lugar nenhum:
	if tokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"valido": false, "erro": "Token não fornecido"})
		return
	}

	// 3. Valida o Token JWT
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err == nil && token.Valid {
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if papel, okPapel := claims["papel"].(string); okPapel {
				usuarioID := claims["id"] // ou o campo que usa no seu JWT

				// Responde com sucesso, ID e o Papel em JSON
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"valido":     true,
					"usuario_id": usuarioID,
					"papel":      strings.ToLower(strings.TrimSpace(papel)),
				})
				return
			}
		}
	}

	// 4. Token inválido ou expirado
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]interface{}{"valido": false, "erro": "Token inválido ou expirado"})
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

// Criamos uma estrutura para receber os dados JSON enviados pelo celular
type RequisicaoCadastro struct {
	Nome     string `json:"nome"`
	Email    string `json:"email"`
	Whatsapp string `json:"whatsapp"`
	Senha    string `json:"senha"`
}

func CadastrarUsuario(w http.ResponseWriter, r *http.Request) {
	// 1. Avisar ao Front-end que a resposta será em formato JSON
	w.Header().Set("Content-Type", "application/json")

	// 2. Garantir que é um POST (O GET já não existe, pois a tela abre no próprio celular)
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Método não permitido"})
		return
	}

	// 3. Ler o JSON enviado pelo App
	var req RequisicaoCadastro
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Formato de dados inválido."})
		return
	}

	// 4. Limpeza de dados (TrimSpace e ToLower)
	nome := strings.TrimSpace(req.Nome)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	senha := req.Senha
	whatsapp := strings.TrimSpace(req.Whatsapp)

	// 5. Validação simples
	if nome == "" || email == "" || senha == "" || whatsapp == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Por favor, preencha todos os campos obrigatórios."})
		return
	}

	// 6. Criptografar a senha
	senhaHash, err := services.HashSenha(senha)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Erro interno ao processar a senha."})
		return
	}

	// 7. Montar a Struct do Banco de Dados
	usuario := structs.Usuario{
		Nome:          nome,
		Email:         email,
		Senha:         senhaHash,
		Whatsapp:      whatsapp,
		Papel:         "passageiro", // Todo cadastro novo nasce como passageiro
		AceitouTermos: true,         // Como já validamos no front, podemos assumir true aqui
	}

	// 8. O Controller delega a gravação para o Model
	err = models.CriarUsuario(&usuario)
	if err != nil {
		// Retornamos 409 Conflict se o e-mail já existir
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Este e-mail já está cadastrado em nosso sistema."})
		return
	}

	// 9. AUTO-LOGIN: Cria o token JWT para já logar o usuário automaticamente
	tokenString, errToken := services.GerarToken(usuario.ID, usuario.Papel)
	if errToken != nil {
		log.Println("Erro ao gerar token no auto-login:", errToken)
		w.WriteHeader(http.StatusCreated)
		// Devolvemos sucesso no cadastro, mas pedimos para logar manual se o token falhar
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": true, "erro": "Conta criada! Por favor, faça login manualmente."})
		return
	}

	// 10. Sucesso total! Devolvemos o Token para o Celular
	// Não gravamos Cookie, enviamos no JSON. O Javascript lá no HTML vai salvar no LocalStorage!
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sucesso": true,
		"token":   tokenString,
	})
}

// ApiLogout limpa os cookies do servidor (Web) e confirma o encerramento da sessão
func ApiLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Sobrescreve e destrói o cookie no navegador (se existir)
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_frota",
		Value:    "",
		Expires:  time.Now().Add(-7 * 24 * time.Hour),
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	// Retorna confirmação em JSON para o Front-end
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sucesso":  true,
		"mensagem": "Logout realizado com sucesso",
	})
}

func ObterConfiguracoes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	
	var config structs.ConfiguracaoApp

	// Busca sempre a configuração de ID 1
	err := db.DB.First(&config, 1).Error
	if err != nil {
		// Se não existir (primeira vez rodando), cria com os valores padrão
		config = structs.ConfiguracaoApp{ID: 1, BaseUrbana: 10.00, KmUrbano: 2.50, LimiteUrbano: 20.00, KmInter: 4.00}
		db.DB.Create(&config)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// DeletarUsuario remove um usuário do sistema, com proteção para o ID 1
func DeletarUsuario(w http.ResponseWriter, r *http.Request) {

	// Se estiver usando rotas padrões do Go com Query Param (ex: /api/usuario?id=1):
	usuarioID := r.URL.Query().Get("id")

	if usuarioID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"erro": "ID do usuário não informado"}`))
		return
	}

	// 2. A BLINDAGEM: Protege o Passageiro Avulso
	if usuarioID == "1" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden) // Erro 403 - Proibido
		w.Write([]byte(`{"erro": "Acesso Negado: O Passageiro Avulso é estrutural e não pode ser excluído."}`))
		return
	}

	// 3. Executa a exclusão com GORM
	// Passamos o ponteiro vazio da Struct de Usuário e o ID
	resultado := db.DB.Delete(&structs.Usuario{}, usuarioID)

	// Verifica se ocorreu algum erro na conexão ou sintaxe do banco
	if resultado.Error != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"erro": "Falha ao processar exclusão no banco de dados"}`))
		return
	}

	// Verifica se realmente achou e apagou a linha (evita dar sucesso se o ID não existia)
	if resultado.RowsAffected == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"erro": "Usuário não encontrado"}`))
		return
	}

	// 4. Retorna sucesso para o Front-End
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"mensagem": "Usuário excluído com sucesso"}`))
}

// AtualizarFCMToken salva o token de push do aparelho do usuário
func AtualizarFCMToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"erro": "Método não permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	usuarioID, err := services.ExtrairUsuarioID(r)
	if err != nil {
		http.Error(w, `{"erro": "Sessão expirada"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
		http.Error(w, `{"erro": "Token inválido"}`, http.StatusBadRequest)
		return
	}

	db.DB.Model(&structs.Usuario{}).Where("id = ?", usuarioID).Update("fcm_token", req.Token)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"sucesso": true})
}
