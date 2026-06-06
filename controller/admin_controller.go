package controller

import (
	"log"
	"net/http"

	"Frota/db"
	"Frota/services"
	"Frota/structs"
)

// Pacote de dados que será enviado para a tela do Dudu
type DadosHomeAdmin struct {
	Usuario    structs.Usuario
	Pendentes  []structs.Corrida
	Aprovadas  []structs.Corrida
	Motoristas []structs.Usuario // Lista de todos os motoristas para o Dudu escolher
}

func HomeAdmin(w http.ResponseWriter, r *http.Request) {
	usuarioID, err := services.ExtrairUsuarioID(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var usuario structs.Usuario
	db.DB.First(&usuario, usuarioID)

	// Proteção: Apenas o ADMIN (Dudu) pode acessar essa tela
	if usuario.Papel != "admin" {
		http.Redirect(w, r, "/passageiro/home", http.StatusSeeOther)
		return
	}

	// 1. Busca todos os motoristas cadastrados na empresa
	var motoristas []structs.Usuario
	db.DB.Where("papel = ?", "motorista").Find(&motoristas)

	// 2. Busca solicitações aguardando atribuição
	var pendentes []structs.Corrida
	db.DB.Preload("Usuario").Where("status = ?", "Aguardando Confirmacao").Order("data_hora_agendada ASC").Find(&pendentes)

	// 3. Busca a agenda geral da frota (Corridas já atribuídas a algum motorista)
	var aprovadas []structs.Corrida
	db.DB.Preload("Usuario").Preload("Motorista").Where("status = ?", "Aprovada").Order("data_hora_agendada ASC").Find(&aprovadas)

	dados := DadosHomeAdmin{
		Usuario:    usuario,
		Pendentes:  pendentes,
		Aprovadas:  aprovadas,
		Motoristas: motoristas,
	}

	err = temp.ExecuteTemplate(w, "AdminHome", dados)
	if err != nil {
		log.Println("Erro na renderização do Admin:", err)
	}
}

// AtribuirCorrida recebe o POST quando o Dudu escolhe um motorista e clica em Atribuir
func AtribuirCorrida(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
        http.Redirect(w, r, "/admin/home", http.StatusSeeOther)
        return
    }
	
	if r.Method == "POST" {
		corridaID := r.FormValue("corrida_id")
		motoristaID := r.FormValue("motorista_id")

		// Atualiza a corrida no banco de dados
		err := db.DB.Model(&structs.Corrida{}).Where("id = ?", corridaID).Updates(map[string]interface{}{
			"motorista_id": motoristaID,
			"status":       "Aprovada",
		}).Error

		if err != nil {
			log.Println("Erro ao atribuir corrida:", err)
		}

		// Volta para o painel do Dudu
		http.Redirect(w, r, "/admin/home", http.StatusSeeOther)
	}
}
