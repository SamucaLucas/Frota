package controller // ou o seu package atual

import (
	"Frota/db"
	"Frota/services"
	"Frota/structs"
	"encoding/json"
	"net/http"
	"strings"
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
		Where("status = ? OR status = ? AND motorista_id = ?", "Aprovada", "Em Corrida", usuarioID).
		Order("data_hora_agendada ASC").
		Find(&atribuidas)

	// 2. Busca o histórico de corridas que ele já finalizou
	var concluidas []structs.Corrida
	db.DB.Preload("Usuario").
		Where("status = ? AND motorista_id = ?", "Concluida", usuarioID).
		Order("data_hora_agendada DESC").
		Limit(10).
		Find(&concluidas)

	// 3. Busca os veículos do motorista
	var veiculos []structs.Veiculo
	db.DB.Where("motorista_id = ?", usuarioID).Order("ativo DESC, id ASC").Find(&veiculos)

	var veiculoAtivo structs.Veiculo
	temVeiculoAtivo := false
	if len(veiculos) > 0 && veiculos[0].Ativo {
		veiculoAtivo = veiculos[0]
		temVeiculoAtivo = true
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sucesso":           true,
		"usuario":           motorista,
		"atribuidas":        atribuidas,
		"concluidas":        concluidas,
		"veiculo_ativo":     veiculoAtivo,
		"veiculos":          veiculos,
		"tem_veiculo_ativo": temVeiculoAtivo,
	})
}

// ReqConcluir mapeia o JSON enviado pelo celular
type ReqConcluir struct {
	CorridaID  int     `json:"corrida_id"`
	ValorFinal float64 `json:"valor_final"`
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

	// Lê o JSON enviado pelo App {"corrida_id": X, "valor_final": Y}
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

	// Define qual valor salvar (Fallback: se o front mandar 0, salva o estimado)
	valorParaSalvar := req.ValorFinal
	if valorParaSalvar <= 0 {
		valorParaSalvar = corrida.ValorEstimado
	}

	// 2. A MÁGICA AQUI: Usamos Updates (plural) com um map para salvar os dois campos!
	errDb := db.DB.Model(&structs.Corrida{}).Where("id = ?", req.CorridaID).Updates(map[string]interface{}{
		"status":      "Concluida",
		"valor_final": valorParaSalvar,
	}).Error

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

		// Envia Push Notification para o passageiro
		var passageiro structs.Usuario
		if err := db.DB.First(&passageiro, corrida.UsuarioID).Error; err == nil && passageiro.FCMToken != "" {
			go services.EnviarPushNotification(passageiro.FCMToken, "✅ Corrida Concluída!", "Obrigado por viajar conosco! Você ganhou 1 Token de Fidelidade.")
		}
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
	MotoristaID    uint    `json:"motorista_id"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	Status         string  `json:"status"`
	CorridaAtivaID uint    `json:"corrida_ativa_id"` // NOVO CAMPO PARA SALVAR A CORRIDA ATIVA
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
		MotoristaID:    payload.MotoristaID,
		Latitude:       payload.Latitude,
		Longitude:      payload.Longitude,
		Status:         payload.Status,
		CorridaAtivaID: payload.CorridaAtivaID, // <--- NOVO CAMPO AQUI!
		UpdatedAt:      time.Now(),
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

// Struct temporária para receber o JSON do Taxímetro
type PayloadCorridaLivre struct {
	MotoristaID   uint    `json:"motorista_id"`
	UsuarioID     uint    `json:"usuario_id"`
	OrigemLat     float64 `json:"origem_lat"`
	OrigemLng     float64 `json:"origem_lng"`
	DestinoLat    float64 `json:"destino_lat"`
	DestinoLng    float64 `json:"destino_lng"`
	KMRodado      float64 `json:"km_rodado"`
	ValorEstimado float64 `json:"valor_estimado"` // NOVO CAMPO
	ValorFinal    float64 `json:"valor_final"`
	DataInicio    string  `json:"data_inicio"`
}

func SalvarCorridaLivre(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"erro": "Método não permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	var payload PayloadCorridaLivre
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"erro": "Dados inválidos"}`, http.StatusBadRequest)
		return
	}

	// VALIDAÇÃO DE SEGURANÇA NO BACKEND (Nunca confie apenas no Front)
	if payload.ValorFinal < 10.00 {
		http.Error(w, `{"erro": "Violação de regra: O valor final não pode ser inferior a R$ 10,00"}`, http.StatusBadRequest)
		return
	}

	dataInicio, _ := time.Parse(time.RFC3339, payload.DataInicio)

	novaCorrida := structs.Corrida{
		UsuarioID:        payload.UsuarioID,
		MotoristaID:      &payload.MotoristaID,
		Tipo:             "livre",
		DataHoraAgendada: dataInicio,
		OrigemTexto:      "Embarque Avulso (Rua)",
		OrigemLat:        payload.OrigemLat,
		OrigemLng:        payload.OrigemLng,
		DestinoTexto:     "Desembarque (Corrida Livre)",
		DestinoLat:       payload.DestinoLat,
		DestinoLng:       payload.DestinoLng,
		KMRodado:         payload.KMRodado,
		ValorEstimado:    payload.ValorEstimado, // Recebe do JS
		ValorFinal:       payload.ValorFinal,    // Recebe do JS
		Status:           "Concluida",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := db.DB.Create(&novaCorrida).Error; err != nil {
		http.Error(w, `{"erro": "Falha ao salvar corrida"}`, http.StatusInternalServerError)
		return
	}

	// GATILHO DE RECOMPENSA: Injeta 1 Token para o Passageiro
	if payload.UsuarioID != 1 { // Supondo que 1 seja o usuário genérico
		db.DB.Model(&structs.Usuario{}).Where("id = ?", payload.UsuarioID).UpdateColumn("tokens", gorm.Expr("tokens + ?", 1))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensagem":   "Corrida livre salva com sucesso!",
		"corrida_id": novaCorrida.ID,
	})
}

func BuscarPassageiroAvulso(w http.ResponseWriter, r *http.Request) {
	// Pega o que o motorista digitou e já converte tudo para minúsculo no Go
	q := strings.ToLower(r.URL.Query().Get("q"))
	var usuarios []structs.Usuario

	// Usa o LOWER() do SQL para ignorar se o e-mail no banco tem letra maiúscula
	db.DB.Where("LOWER(email) LIKE ? OR whatsapp LIKE ?", "%"+q+"%", "%"+q+"%").Limit(5).Find(&usuarios)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usuarios)
}

// AtualizarStatusCorrida muda o status da corrida quando o motorista chega ao local
func AtualizarStatusCorrida(w http.ResponseWriter, r *http.Request) {
    // 1. Trava de segurança para aceitar apenas POST
    if r.Method != http.MethodPost {
        http.Error(w, `{"erro": "Método não permitido"}`, http.StatusMethodNotAllowed)
        return
    }

    // 2. Estrutura para ler o JSON que o celular mandou
    var req struct {
        CorridaID uint   `json:"corrida_id"`
        Status    string `json:"status"`
    }

    // 3. Decodifica o JSON
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"erro": "Dados inválidos"}`, http.StatusBadRequest)
        return
    }

    // 4. Atualiza a corrida no banco de dados
    errDb := db.DB.Model(&structs.Corrida{}).Where("id = ?", req.CorridaID).Update("status", req.Status).Error
    if errDb != nil {
        http.Error(w, `{"erro": "Falha ao atualizar status no banco"}`, http.StatusInternalServerError)
        return
    }

    // 5. Devolve o sucesso para o aplicativo liberar a Etapa 2
	// 5. MÁGICA: Disparar Notificação Push para o Passageiro!
	var corridaAtualizada structs.Corrida
	if err := db.DB.Preload("Usuario").First(&corridaAtualizada, req.CorridaID).Error; err == nil {
		if corridaAtualizada.Usuario.FCMToken != "" {
			titulo := "Atualização da Viagem"
			corpo := "Status: " + req.Status
			
			if req.Status == "A Caminho" {
				titulo = "O motorista está a caminho! 🚗"
				corpo = "Prepare-se, seu motorista está indo até você."
			} else if req.Status == "Em Corrida" {
				titulo = "O motorista chegou! 📍"
				corpo = "O motorista já está no local de embarque aguardando."
			}

			// Dispara a notificação sem travar a resposta HTTP
			go services.EnviarPushNotification(corridaAtualizada.Usuario.FCMToken, titulo, corpo)
		}
	}

    // 6. Devolve o sucesso para o aplicativo liberar a Etapa 2
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]bool{"sucesso": true})
}
