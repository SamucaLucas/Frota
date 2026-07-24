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

// ApiConfigPrecos devolve as taxas atuais do sistema para o App calcular viagens
func ApiConfigPrecos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Aqui no futuro você fará: db.DB.First(&configPrecos)
	// Por enquanto, enviamos os valores "dinâmicos" direto da API
	json.NewEncoder(w).Encode(map[string]interface{}{
		"base_urbano":   10.00,
		"km_urbano":     2.50,
		"km_inter":      4.00, // Ajustei para 4.00 baseado na sua struct DadosAgendamento
		"limite_urbano": 20.0,
	})
}

//golang
/*
type DadosHomePassageiro struct {
	Usuario         structs.Usuario
	ProximasViagens []structs.Corrida
	TemProximas     bool
	Historico       []structs.Corrida
	TemHistorico    bool
}

func HomePassageiro(w http.ResponseWriter, r *http.Request) {
	usuarioID, err := services.ExtrairUsuarioID(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var usuario structs.Usuario
	db.DB.First(&usuario, usuarioID)

	// 2. Mude de .First(&proxima) para .Find(&proximas)
	var proximas []structs.Corrida
	db.DB.Preload("Motorista").Where("usuario_id = ? AND status IN ?", usuarioID, []string{"Aguardando Confirmacao", "Aprovada"}).Order("data_hora_agendada ASC").Find(&proximas)

	var historico []structs.Corrida
	resHist := db.DB.Where("usuario_id = ? AND status = ?", usuarioID, "Concluida").Order("data_hora_agendada DESC").Limit(3).Find(&historico)

	// 3. Atualize os dados enviados
	dados := DadosHomePassageiro{
		Usuario:         usuario,
		TemProximas:     len(proximas) > 0, // <-- Verifica se a lista tem itens
		ProximasViagens: proximas,          // <-- Envia a lista toda
		TemHistorico:    resHist.RowsAffected > 0,
		Historico:       historico,
	}

	err = temp.ExecuteTemplate(w, "PassageiroHome", dados)
	if err != nil {
		log.Println("Erro na renderização:", err)
	}
}

type DadosAgendamento struct {
	PrecoBaseUrbano float64
	PrecoKmUrbano   float64
	PrecoKmInter    float64
	LimiteKmUrbano  float64
}

// AgendarViagem lida com a exibição do formulário e a gravação da nova corrida
func AgendarViagem(w http.ResponseWriter, r *http.Request) {
	usuarioID, err := services.ExtrairUsuarioID(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// 2. Se for método GET (Acessando a página)
	if r.Method == "GET" {
		// Simulando os valores que o Dudu configurou no Painel Administrativo
		// No futuro, faremos: db.DB.First(&config)
		configPrecos := DadosAgendamento{
			PrecoBaseUrbano: 10.00, // Inicia em R$ 10,00
			PrecoKmUrbano:   2.50,  // R$ 2,50 por KM após o 1º KM
			PrecoKmInter:    4.00,  // R$ 3,50 a 4,00 por KM em viagens longas
			LimiteKmUrbano:  20.0,
		}

		err := temp.ExecuteTemplate(w, "PassageiroAgendar", configPrecos)
		if err != nil {
			log.Println("Erro ao renderizar Agendar:", err)
		}
		return
	}

	// 3. Se for método POST (Clicou no botão "Solicitar Agendamento")
	if r.Method == "POST" {
		origem := r.FormValue("origem")
		destino := r.FormValue("destino")
		dataHoraStr := r.FormValue("data_hora")

		// 3.1 Captura das Coordenadas
		origemLat, _ := strconv.ParseFloat(r.FormValue("origem_lat"), 64)
		origemLng, _ := strconv.ParseFloat(r.FormValue("origem_lng"), 64)
		destinoLat, _ := strconv.ParseFloat(r.FormValue("destino_lat"), 64)
		destinoLng, _ := strconv.ParseFloat(r.FormValue("destino_lng"), 64)

		// 3.2 Captura da Distância e Valor (Novo!)
		kmRodado, _ := strconv.ParseFloat(r.FormValue("km_rodado"), 64)
		valorEstimado, _ := strconv.ParseFloat(r.FormValue("valor_estimado"), 64)

		// 1. Carrega o fuso horário oficial do Brasil
        loc, errLoc := time.LoadLocation("America/Sao_Paulo")
        if errLoc != nil {
            log.Println("Erro ao carregar Timezone:", errLoc)
            // Caso falhe por falta de pacotes no SO, usa o fuso local do sistema
            loc = time.Local
        }

        // 2. Substitua o time.Parse antigo por time.ParseInLocation
        dataHora, err := time.ParseInLocation("2006-01-02T15:04", dataHoraStr, loc)
        if err != nil {
            log.Println("Erro ao converter data e hora:", err)
            // trate o erro exibindo uma mensagem na tela...
            return
        }

		// 3.4 Cria a corrida na base de dados com as estimativas!
		novaCorrida := structs.Corrida{
			UsuarioID:        usuarioID,
			Tipo:             "padrao",
			OrigemTexto:      origem,
			OrigemLat:        origemLat,
			OrigemLng:        origemLng,
			DestinoTexto:     destino,
			DestinoLat:       destinoLat,
			DestinoLng:       destinoLng,
			KMRodado:         kmRodado,      // 👈 Gravando no banco
			ValorEstimado:    valorEstimado, // 👈 Gravando no banco
			DataHoraAgendada: dataHora,
			Status:           "Aguardando Confirmacao",
		}

		errDb := db.DB.Create(&novaCorrida).Error
		if errDb != nil {
			log.Println("Erro ao salvar corrida no banco:", errDb)
			http.Redirect(w, r, "/passageiro/agendar?erro=falha_banco", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/passageiro/home", http.StatusSeeOther)
	}
}
*/
