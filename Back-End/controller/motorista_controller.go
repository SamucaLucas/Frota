package controller // ou o seu package atual

import (
	"Frota/db"
	"Frota/services"
	"Frota/structs"
	"encoding/json"
	"net/http"
	"time"

	// "Frota/db"
	// "Frota/services"
	// "Frota/structs"
	"gorm.io/gorm"
)

// ApiHomeMotorista (GET) devolve os dados do painel do motorista
func ApiHomeMotorista(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	usuarioID, err := services.ExtrairUsuarioID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Sessão expirada. Faça login novamente."})
		return
	}

	var motorista structs.Usuario
	if err := db.DB.First(&motorista, usuarioID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Usuário não encontrado."})
		return
	}

	// Proteção: Apenas motoristas entram aqui
	if motorista.Papel != "motorista" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Acesso negado. Área restrita a motoristas."})
		return
	}

	// 1. Busca as corridas que o Dudu despachou para este motorista
	var atribuidas []structs.Corrida
	db.DB.Preload("Usuario").
		Where("status = ? AND motorista_id = ?", "Aprovada", usuarioID).
		Order("data_hora_agendada ASC").
		Find(&atribuidas)

	// 2. Busca o histórico de corridas que ele já finalizou
	var concluidas []structs.Corrida
	db.DB.Preload("Usuario").
		Where("status = ? AND motorista_id = ?", "Concluida", usuarioID).
		Order("data_hora_agendada DESC").
		Limit(10).
		Find(&concluidas)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sucesso":    true,
		"usuario":    motorista,
		"atribuidas": atribuidas,
		"concluidas": concluidas,
	})
}

// ReqConcluir mapeia o JSON enviado pelo celular
type ReqConcluir struct {
	CorridaID int `json:"corrida_id"`
}

// ApiPostConcluirCorrida (POST) acionada quando o motorista finaliza a viagem
func ApiPostConcluirCorrida(w http.ResponseWriter, r *http.Request) {
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

	// Lê o JSON enviado pelo App {"corrida_id": X}
	var req ReqConcluir
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Dados inválidos."})
		return
	}

	if req.CorridaID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "ID da corrida não informado."})
		return
	}

	// 1. Busca a corrida e verifica segurança
	var corrida structs.Corrida
	if err := db.DB.First(&corrida, req.CorridaID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Corrida não encontrada."})
		return
	}

	// TRAVA DE SEGURANÇA: O motorista só pode concluir a própria corrida
	if corrida.MotoristaID == nil || *corrida.MotoristaID != usuarioID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Você não tem permissão para alterar esta corrida."})
		return
	}

	// 2. Atualiza o status da corrida para "Concluida"
	errDb := db.DB.Model(&structs.Corrida{}).Where("id = ?", req.CorridaID).Update("status", "Concluida").Error
	if errDb != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Erro ao salvar no banco."})
		return
	}

	// ==========================================
	// 🏆 MÁGICA DOS TOKENS: Recompensar o Passageiro
	// ==========================================
	if corrida.UsuarioID != 0 {
		// Adiciona +1 Token no saldo do Passageiro
		db.DB.Model(&structs.Usuario{}).Where("id = ?", corrida.UsuarioID).UpdateColumn("tokens", gorm.Expr("tokens + ?", 1))

		// Salva o recibo no Histórico de Tokens para o extrato do passageiro
		historico := structs.HistoricoToken{
			UsuarioID:     corrida.UsuarioID,
			Quantidade:    1,
			TipoTransacao: "corrida",
		}
		db.DB.Create(&historico)
	}

	// Responde sucesso para o Front-end!
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sucesso":  true,
		"mensagem": "Corrida finalizada com sucesso!",
	})

}

// Estrutura temporária para ler o JSON que vem do celular
type PayloadLocalizacao struct {
	MotoristaID uint    `json:"motorista_id"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Status      string  `json:"status"`
}

// AtualizarLocalizacao recebe as coordenadas do Capacitor e salva no banco
func AtualizarLocalizacao(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload PayloadLocalizacao

	// Lê o JSON do corpo da requisição
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"erro": "Dados inválidos"}`, http.StatusBadRequest)
		return
	}

	// Monta o objeto para o banco de dados
	localizacao := structs.LocalizacaoMotorista{
		MotoristaID: payload.MotoristaID,
		Latitude:    payload.Latitude,
		Longitude:   payload.Longitude,
		Status:      payload.Status,
		UpdatedAt:   time.Now(),
	}

	// Save() no GORM com a PK preenchida faz um "Upsert" (Salva se não existir, Atualiza se existir)
	if err := db.DB.Save(&localizacao).Error; err != nil {
		http.Error(w, `{"erro": "Falha ao salvar localização"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"mensagem": "Localização atualizada com sucesso!"})
}
