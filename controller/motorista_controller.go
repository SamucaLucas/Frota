package controller

import (
	"log"
	"net/http"
	"strings"

	"Frota/db"
	"Frota/services"
	"Frota/structs"

	"gorm.io/gorm"
)

type DadosHomeMotorista struct {
	Usuario    structs.Usuario
	Atribuidas []structs.Corrida
	Concluidas []structs.Corrida
}

func HomeMotorista(w http.ResponseWriter, r *http.Request) {
	usuarioID, err := services.ExtrairUsuarioID(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var motorista structs.Usuario
	db.DB.First(&motorista, usuarioID)

	// Proteção: Apenas motoristas entram aqui
	if motorista.Papel != "motorista" {
		http.Redirect(w, r, "/passageiro/home", http.StatusSeeOther)
		return
	}

	// 1. Busca as corridas que o Dudu despachou para este motorista
	var atribuidas []structs.Corrida
	db.DB.Preload("Usuario").Where("status = ? AND motorista_id = ?", "Aprovada", usuarioID).Order("data_hora_agendada ASC").Find(&atribuidas)

	// 2. Busca o histórico de corridas que ele já finalizou hoje
	var concluidas []structs.Corrida
	db.DB.Preload("Usuario").Where("status = ? AND motorista_id = ?", "Concluida", usuarioID).Order("data_hora_agendada DESC").Limit(10).Find(&concluidas)

	dados := DadosHomeMotorista{
		Usuario:    motorista,
		Atribuidas: atribuidas,
		Concluidas: concluidas,
	}

	err = temp.ExecuteTemplate(w, "MotoristaHome", dados)
	if err != nil {
		log.Println("Erro na renderização do Motorista:", err)
	}
}

// ConcluirCorrida é acionada quando o motorista finaliza a viagem
func ConcluirCorrida(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/motorista/home", http.StatusSeeOther)
		return
	}

	corridaID := strings.TrimPrefix(r.URL.Path, "/motorista/concluir/")
	if corridaID == "" {
		http.Redirect(w, r, "/motorista/home", http.StatusSeeOther)
		return
	}

	// 1. Atualiza o status da corrida para "Concluida"
	errDb := db.DB.Model(&structs.Corrida{}).Where("id = ?", corridaID).Update("status", "Concluida").Error
	if errDb != nil {
		log.Println("Erro ao concluir corrida:", errDb)
	}

	// ==========================================
	// 🏆 MÁGICA DOS TOKENS: Recompensar o Passageiro
	// ==========================================

	// 2. Busca a corrida para saber quem foi o passageiro
	var corrida structs.Corrida
	db.DB.First(&corrida, corridaID)

	if corrida.UsuarioID != 0 {
		// 3. Adiciona +1 Token no saldo do Passageiro (Usando GORM Expr para somar de forma segura)
		db.DB.Model(&structs.Usuario{}).Where("id = ?", corrida.UsuarioID).UpdateColumn("tokens", gorm.Expr("tokens + ?", 1))

		// 4. Salva o recibo no Histórico de Tokens para o extrato do passageiro
		historico := structs.HistoricoToken{
			UsuarioID:     corrida.UsuarioID,
			Quantidade:    1,
			TipoTransacao: "corrida",
		}
		db.DB.Create(&historico)
	}

	// Redireciona de volta para a Home do Motorista
	http.Redirect(w, r, "/motorista/home", http.StatusSeeOther)
}
