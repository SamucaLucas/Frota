package controller

import (
	"Frota/db"
	"Frota/services"
	"Frota/structs"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

//api

// HomePassageiro agora é uma API REST que devolve os dados em JSON
func HomePassageiro(w http.ResponseWriter, r *http.Request) {
	// 1. Avisa que a resposta é JSON
	w.Header().Set("Content-Type", "application/json")

	// 2. Extrai o ID do usuário (Agora lendo tanto Header quanto Cookie)
	usuarioID, err := services.ExtrairUsuarioID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Sessão expirada ou inválida."})
		return
	}

	// 3. Busca o usuário no banco
	var usuario structs.Usuario
	if err := db.DB.First(&usuario, usuarioID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Usuário não encontrado."})
		return
	}

	// 4. Busca as próximas viagens
	var proximas []structs.Corrida
	db.DB.Preload("Motorista").
		Where("usuario_id = ? AND status IN ?", usuarioID, []string{"Aguardando Confirmacao", "Aprovada"}).
		Order("data_hora_agendada ASC").
		Find(&proximas)

	// 5. Busca o histórico
	var historico []structs.Corrida
	db.DB.Where("usuario_id = ? AND status = ?", usuarioID, "Concluida").
		Order("data_hora_agendada DESC").
		Limit(3).
		Find(&historico)

	// 6. Devolve tudo mastigadinho em JSON para o Front-end
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sucesso":   true,
		"usuario":   usuario,
		"proximas":  proximas,
		"historico": historico,
	})
}

// ReqAgendamento mapeia o JSON exato que o Front-end envia no POST
type ReqAgendamento struct {
	OrigemTexto   string  `json:"origem_texto"`
	OrigemLat     float64 `json:"origem_lat"`
	OrigemLng     float64 `json:"origem_lng"`
	DestinoTexto  string  `json:"destino_texto"`
	DestinoLat    float64 `json:"destino_lat"`
	DestinoLng    float64 `json:"destino_lng"`
	DataHora      string  `json:"data_hora"`
	KMRodado      float64 `json:"km_rodado"`
	ValorEstimado float64 `json:"valor_estimado"`
}

// AgendarViagem recebe o pedido JSON do App e grava no banco
func AgendarViagem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Garante que só aceita POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Método não permitido"})
		return
	}

	// 2. Extrai o ID do usuário pelo Token (Obrigatório estar logado)
	usuarioID, err := services.ExtrairUsuarioID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Sessão expirada. Faça login novamente."})
		return
	}

	// 3. Lê o corpo da requisição JSON enviada pelo celular
	var req ReqAgendamento
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Formato de dados inválido."})
		return
	}

	// 4. Validação Básica
	if strings.TrimSpace(req.OrigemTexto) == "" || strings.TrimSpace(req.DestinoTexto) == "" || req.DataHora == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Preencha origem, destino e data/hora."})
		return
	}

	// 5. Configuração de Fuso Horário e Parse da Data
	loc, errLoc := time.LoadLocation("America/Sao_Paulo")
	if errLoc != nil {
		loc = time.Local // Fallback de segurança
	}

	dataHora, errData := time.ParseInLocation("2006-01-02T15:04", req.DataHora, loc)
	if errData != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Data ou hora inválida."})
		return
	}

	// 6. Monta a Struct do Banco de Dados
	novaCorrida := structs.Corrida{
		UsuarioID:        usuarioID,
		Tipo:             "padrao",
		OrigemTexto:      req.OrigemTexto,
		OrigemLat:        req.OrigemLat,
		OrigemLng:        req.OrigemLng,
		DestinoTexto:     req.DestinoTexto,
		DestinoLat:       req.DestinoLat,
		DestinoLng:       req.DestinoLng,
		KMRodado:         req.KMRodado,
		ValorEstimado:    req.ValorEstimado,
		DataHoraAgendada: dataHora,
		Status:           "Aguardando Confirmacao",
	}

	// 7. Salva a Corrida no Banco de Dados
	if errDb := db.DB.Create(&novaCorrida).Error; errDb != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Falha ao gravar viagem no servidor."})
		return
	}

	// 8. Responde Sucesso para o App
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sucesso":  true,
		"mensagem": "Viagem solicitada com sucesso!",
	})
}


