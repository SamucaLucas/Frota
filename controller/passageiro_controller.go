package controller

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"Frota/db"
	"Frota/services"
	"Frota/structs"
)

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

// AgendarViagem lida com a exibição do formulário e a gravação da nova corrida
func AgendarViagem(w http.ResponseWriter, r *http.Request) {
	// 1. Verifica se o usuário está logado
	usuarioID, err := services.ExtrairUsuarioID(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// 2. Se for método GET (Acessando a página), apenas exibe o HTML
	if r.Method == "GET" {
		err := temp.ExecuteTemplate(w, "PassageiroAgendar", nil)
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

		// 3.1 Captura das Coordenadas (Convertendo de string para Float64)
		origemLat, _ := strconv.ParseFloat(r.FormValue("origem_lat"), 64)
		origemLng, _ := strconv.ParseFloat(r.FormValue("origem_lng"), 64)
		destinoLat, _ := strconv.ParseFloat(r.FormValue("destino_lat"), 64)
		destinoLng, _ := strconv.ParseFloat(r.FormValue("destino_lng"), 64)

		// 3.2. Tratamento da Data/Hora
		dataHoraAgendada, err := time.Parse("2006-01-02T15:04", dataHoraStr)
		if err != nil {
			log.Println("Erro ao converter data:", err)
			http.Redirect(w, r, "/passageiro/agendar?erro=data_invalida", http.StatusSeeOther)
			return
		}

		// 3.3. Cria a corrida no banco de dados com as coordenadas exatas!
		novaCorrida := structs.Corrida{
			UsuarioID:        usuarioID,
			Tipo:             "padrao",
			OrigemTexto:      origem,
			OrigemLat:        origemLat,
			OrigemLng:        origemLng,
			DestinoTexto:     destino,
			DestinoLat:       destinoLat,
			DestinoLng:       destinoLng,
			DataHoraAgendada: dataHoraAgendada,
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
