package controller // ou o seu package atual

import (
	"Frota/db"
	"Frota/services"
	"Frota/structs"
	"encoding/json"
	"net/http"
	"strings"
)

// HomeAdmin retorna os dados do painel do administrador em formato JSON
func HomeAdmin(w http.ResponseWriter, r *http.Request) {
	// 1. Avisa ao Front-end que a resposta é um JSON
	w.Header().Set("Content-Type", "application/json")

	// 2. Extrai o ID do usuário pelo Token Híbrido (Header ou Cookie)
	usuarioID, err := services.ExtrairUsuarioID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Sessão expirada. Faça login novamente."})
		return
	}

	// 3. Busca o usuário no banco de dados
	var usuario structs.Usuario
	if err := db.DB.First(&usuario, usuarioID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Usuário não encontrado."})
		return
	}

	// 4. PROTEÇÃO DE ROTA: Apenas o ADMIN (Dudu) pode acessar essa API
	if usuario.Papel != "admin" {
		w.WriteHeader(http.StatusForbidden) // 403: Acesso Negado
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Acesso negado. Área restrita à Central."})
		return
	}

	// 5. Busca todos os motoristas cadastrados na empresa
	var motoristas []structs.Usuario
	db.DB.Where("papel = ?", "motorista").Find(&motoristas)

	// 6. Busca solicitações aguardando atribuição (Pendentes)
	// O Preload("Usuario") garante que os dados de quem pediu a corrida venham no JSON
	var pendentes []structs.Corrida
	db.DB.Preload("Usuario").
		Where("status = ?", "Aguardando Confirmacao").
		Order("data_hora_agendada ASC").
		Find(&pendentes)

	// 7. Busca a agenda geral da frota (Corridas já atribuídas a algum motorista)
	// O Preload pega os dados do Passageiro e do Motorista
	var aprovadas []structs.Corrida
	db.DB.Preload("Usuario").Preload("Motorista").
		Where("status = ?", "Aprovada").
		Order("data_hora_agendada ASC").
		Find(&aprovadas)

	// 8. Devolve tudo formatado em JSON para o App
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sucesso":    true,
		"usuario":    usuario,
		"motoristas": motoristas,
		"pendentes":  pendentes,
		"aprovadas":  aprovadas,
	})
}

// ApiGetDespachar (GET) devolve os dados da corrida e a lista de motoristas
func ApiGetDespachar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	usuarioID, err := services.ExtrairUsuarioID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Sessão expirada."})
		return
	}

	var admin structs.Usuario
	db.DB.First(&admin, usuarioID)

	if admin.Papel != "admin" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Acesso negado."})
		return
	}

	// Extrai o ID da corrida da URL (ex: /api/admin/despachar/5)
	pathParts := strings.Split(r.URL.Path, "/")
	corridaID := pathParts[len(pathParts)-1]

	var corrida structs.Corrida
	if err := db.DB.Preload("Usuario").First(&corrida, corridaID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Corrida não encontrada."})
		return
	}

	var motoristas []structs.Usuario
	db.DB.Where("papel = ?", "motorista").Find(&motoristas)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sucesso":    true,
		"corrida":    corrida,
		"motoristas": motoristas,
		"admin_id":   admin.ID, // Enviamos o ID do Admin para ele poder assumir a corrida se quiser
	})
}

// Estrutura para ler o JSON enviado pelo celular na hora de atribuir
type ReqAtribuir struct {
	CorridaID   int `json:"corrida_id"`
	MotoristaID int `json:"motorista_id"`
}

// ApiPostAtribuir (POST) salva o motorista na corrida e muda o status para "Aprovada"
func ApiPostAtribuir(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	usuarioID, err := services.ExtrairUsuarioID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Sessão expirada."})
		return
	}

	var admin structs.Usuario
	db.DB.First(&admin, usuarioID)

	if admin.Papel != "admin" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Acesso negado."})
		return
	}

	// Lê o JSON enviado pelo Front-end
	var req ReqAtribuir
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Dados inválidos."})
		return
	}

	if req.MotoristaID == 0 || req.CorridaID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Motorista ou Corrida não informados."})
		return
	}

	// Atualiza no banco de dados
	errDb := db.DB.Model(&structs.Corrida{}).Where("id = ?", req.CorridaID).Updates(map[string]interface{}{
		"motorista_id": req.MotoristaID,
		"status":       "Aprovada",
	}).Error

	if errDb != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Erro ao salvar no banco."})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sucesso": true,
	})
}

/*

// MapaAdmin renderiza a página exclusiva do mapa de monitoramento
func MapaAdmin(w http.ResponseWriter, r *http.Request) {
	usuarioID, err := services.ExtrairUsuarioID(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var admin structs.Usuario
	db.DB.First(&admin, usuarioID)

	// Proteção: Apenas o ADMIN (Dudu) pode acessar essa tela
	if admin.Papel != "admin" {
		http.Redirect(w, r, "/passageiro/home", http.StatusSeeOther)
		return
	}

	dados := struct {
		Usuario structs.Usuario
	}{
		Usuario: admin,
	}

	err = temp.ExecuteTemplate(w, "AdminMapa", dados)
	if err != nil {
		log.Println("Erro na renderização do Mapa do Admin:", err)
	}
}

// ApiCorridasPendentes devolve APENAS os dados em JSON, sem tela HTML
func ApiCorridasPendentes(w http.ResponseWriter, r *http.Request) {

	// 1. Busca todos os motoristas cadastrados na empresa
	var motoristas []structs.Usuario
	db.DB.Where("papel = ?", "motorista").Find(&motoristas)

	// 2. Busca solicitações aguardando atribuição
	var pendentes []structs.Corrida
	db.DB.Preload("Usuario").Where("status = ?", "Aguardando Confirmacao").Order("data_hora_agendada ASC").Find(&pendentes)

	// 3. Busca a agenda geral da frota (Corridas já atribuídas a algum motorista)
	var aprovadas []structs.Corrida
	db.DB.Preload("Usuario").Preload("Motorista").Where("status = ?", "Aprovada").Order("data_hora_agendada ASC").Find(&aprovadas)


	// A mágica do REST:
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pendentes)
}
*/
