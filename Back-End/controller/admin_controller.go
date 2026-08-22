package controller // ou o seu package atual

import (
	"Frota/db"
	"Frota/services"
	"Frota/structs"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func Ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

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
	db.DB.Preload("Usuario").Preload("Motorista").
		Where("status = ? OR status = ?", "Pendente", "Aguardando Confirmacao").
		Order("created_at DESC").
		Find(&pendentes)

	// 2. Busca TODAS as Aprovadas, Em Corrida, A Caminho
	var aprovadas []structs.Corrida
	db.DB.Preload("Usuario").Preload("Motorista").
		Where("status IN ?", []string{"Aprovada", "Confirmada", "Em Corrida", "A Caminho"}).
		Order("data_hora_agendada ASC").
		Find(&aprovadas)

	// 3. Busca TODO o Histórico de Concluídas
	var concluidas []structs.Corrida
	db.DB.Preload("Usuario").Preload("Motorista").
		Where("status = ? OR status = ?", "Concluida", "Realizada").
		Order("data_hora_agendada DESC").
		Find(&concluidas)

	// 4. Retorna tudo no JSON
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sucesso":    true,
		"usuario":    usuario, // Aqui é o perfil do admin para mostrar o nome dele no topo
		"pendentes":  pendentes,
		"aprovadas":  aprovadas,
		"concluidas": concluidas,
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

	// Busca o Motorista para notificar
	var motorista structs.Usuario
	if err := db.DB.First(&motorista, req.MotoristaID).Error; err == nil && motorista.FCMToken != "" {
		go services.EnviarPushNotification(motorista.FCMToken, "🚕 Nova Corrida Atribuída!", "A central enviou uma corrida para você. Abra o app.")
	}

	// Busca a Corrida e o Passageiro para notificar
	var corrida structs.Corrida
	if err := db.DB.Preload("Usuario").First(&corrida, req.CorridaID).Error; err == nil {
		if corrida.Usuario.FCMToken != "" {
			go services.EnviarPushNotification(corrida.Usuario.FCMToken, "🚗 Seu Motorista foi definido!", "A central aprovou a corrida e o motorista está a caminho.")
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sucesso": true,
	})
}

// BuscarLocalizacoesAdmin retorna a posição de todos os motoristas ativos
// BuscarLocalizacoesAdmin retorna a posição de todos os motoristas ativos
func BuscarLocalizacoesAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"erro": "Método não permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	var localizacoes []structs.LocalizacaoMotorista

	// O GORM agora já vai trazer o CorridaAtivaID automaticamente!
	if err := db.DB.Preload("Motorista").Find(&localizacoes).Error; err != nil {
		http.Error(w, `{"erro": "Falha ao buscar localizações"}`, http.StatusInternalServerError)
		return
	}

	// -------------------------------------------------------------------
	// GARANTIA DE NOME: Busca motoristas e força o nome no JSON
	// -------------------------------------------------------------------
	var motoristas []structs.Usuario
	db.DB.Where("papel = ?", "motorista").Find(&motoristas)

	mapaMotoristas := make(map[uint]string)
	for _, m := range motoristas {
		mapaMotoristas[m.ID] = m.Nome
	}

	var veiculosAtivos []structs.Veiculo
	db.DB.Where("ativo = ?", true).Find(&veiculosAtivos)
	mapaVeiculos := make(map[uint]structs.Veiculo)
	for _, v := range veiculosAtivos {
		mapaVeiculos[v.MotoristaID] = v
	}

	dadosJSON, _ := json.Marshal(localizacoes)
	var listaDinamica []map[string]interface{}
	json.Unmarshal(dadosJSON, &listaDinamica)

	for i, loc := range listaDinamica {
		var id float64
		if v, ok := loc["motorista_id"].(float64); ok {
			id = v
		} else if v, ok := loc["MotoristaID"].(float64); ok {
			id = v
		} else if v, ok := loc["usuario_id"].(float64); ok {
			id = v
		} else if v, ok := loc["id"].(float64); ok {
			id = v
		}

		if nome, existe := mapaMotoristas[uint(id)]; existe {
			loc["nome_garantido"] = nome
		} else {
			loc["nome_garantido"] = "Motorista #" + fmt.Sprint(id)
		}

		if veiculo, existe := mapaVeiculos[uint(id)]; existe {
			loc["veiculo_ativo"] = veiculo
		}

		listaDinamica[i] = loc
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(listaDinamica)
}

// Estrutura para receber o JSON do Front-End da Nova Chamada
type NovaChamadaRequest struct {
	NomePassageiro     string  `json:"nome_passageiro"`
	TelefonePassageiro string  `json:"telefone_passageiro"`
	OrigemTexto        string  `json:"origem_texto"`
	OrigemLat          float64 `json:"origem_lat"`
	OrigemLng          float64 `json:"origem_lng"`
	DestinoTexto       string  `json:"destino_texto"`
	DestinoLat         float64 `json:"destino_lat"`
	DestinoLng         float64 `json:"destino_lng"`
	DataHora           string  `json:"data_hora"`
	KmRodado           float64 `json:"km_rodado"`
	ValorEstimado      float64 `json:"valor_estimado"`
	MotoristaID        uint    `json:"motorista_id"`
	Tipo               string  `json:"tipo"`
}

// NovaChamada cria uma viagem manualmente via Central
func NovaChamada(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Verifica se o Admin está logado
	_, err := services.ExtrairUsuarioID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Sessão expirada."})
		return
	}

	// 2. Decodifica os dados enviados pelo JavaScript
	var req NovaChamadaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Dados inválidos."})
		return
	}

	// 3. Anexa os dados do cliente avulso na Origem para o motorista conseguir ler
	origemComNome := req.OrigemTexto
	if req.NomePassageiro != "" {
		contato := req.TelefonePassageiro
		if contato == "" {
			contato = "Sem telefone"
		}
		origemComNome = req.OrigemTexto + " (Cliente: " + req.NomePassageiro + " - " + contato + ")"
	}

	// 4. Tratamento da Data e Hora (Converte a string do HTML para o formato de tempo do Go)
	dataAgendada, err := time.Parse("2006-01-02T15:04", req.DataHora)
	if err != nil {
		dataAgendada = time.Now() // Fallback de segurança
	}

	// 5. Monta a nova corrida incluindo TODOS os dados de GPS, Distância e Tipo
	novaCorrida := structs.Corrida{
		UsuarioID:        1,
		OrigemTexto:      origemComNome,
		OrigemLat:        req.OrigemLat,
		OrigemLng:        req.OrigemLng,
		DestinoTexto:     req.DestinoTexto,
		DestinoLat:       req.DestinoLat,
		DestinoLng:       req.DestinoLng,
		KMRodado:         req.KmRodado,
		ValorEstimado:    req.ValorEstimado,
		DataHoraAgendada: dataAgendada,
		Tipo:             req.Tipo,
	}

	// 6. Define o status: Vai para a fila (Pendente) ou direto para um motorista (Aprovada)
	if req.MotoristaID == 0 {
		novaCorrida.Status = "Pendente"
	} else {
		novaCorrida.Status = "Aprovada"
		novaCorrida.MotoristaID = &req.MotoristaID
	}

	// 7. Salva no banco de dados
	if err := db.DB.Create(&novaCorrida).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"sucesso": false, "erro": "Erro ao salvar chamada no banco."})
		return
	}

	// 8. Se foi atribuída direto a um motorista, envia notificação
	if req.MotoristaID != 0 {
		var motorista structs.Usuario
		if err := db.DB.First(&motorista, req.MotoristaID).Error; err == nil && motorista.FCMToken != "" {
			go services.EnviarPushNotification(motorista.FCMToken, "🚕 Nova Corrida Atribuída!", "A central criou uma nova corrida diretamente para você.")
		}
	}

	// 8. Retorna sucesso e o ID gerado para o front-end poder redirecionar
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sucesso":    true,
		"corrida_id": novaCorrida.ID,
	})
}
